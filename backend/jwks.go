package main

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"
)

// jwkKey は RFC 7517 JWK のうち、RSA 署名検証に必要なフィールドだけを受け取る。
type jwkKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"` // RSA modulus（Base64URL）
	E   string `json:"e"` // RSA exponent（Base64URL、通常は AQAB = 65537）
}

type jwksDoc struct {
	Keys []jwkKey `json:"keys"`
}

// jwksCache は認可サーバー GET した JWKS を kid → *rsa.PublicKey に変換して保持する。
// 毎リクエストで JWKS を取りに行かないよう TTL 経過後だけ refresh する（getKey 内で判定）。
type jwksCache struct {
	mu      sync.RWMutex
	keys    map[string]*rsa.PublicKey
	fetched time.Time
	ttl     time.Duration
	uri     string
}

func newJWKSCache(uri string, ttl time.Duration) *jwksCache {
	return &jwksCache{
		keys: make(map[string]*rsa.PublicKey),
		ttl:  ttl,
		uri:  uri,
	}
}

// rsaPublicKeyFromJWK は JWK の n, e（いずれも Base64URL エンコード）から *rsa.PublicKey を復元する。
func rsaPublicKeyFromJWK(nB64, eB64 string) (*rsa.PublicKey, error) {
	nb, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, fmt.Errorf("JWK n のデコード: %w", err)
	}
	eb, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, fmt.Errorf("JWK e のデコード: %w", err)
	}
	n := new(big.Int).SetBytes(nb)
	e := new(big.Int).SetBytes(eb)
	if !e.IsInt64() {
		return nil, fmt.Errorf("JWK e が大きすぎます")
	}
	return &rsa.PublicKey{N: n, E: int(e.Int64())}, nil
}

// refresh は JWKS を HTTP で取り直し、キャッシュを置き換える（失敗時は旧キャッシュのままにしない）。
func (c *jwksCache) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.uri, nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("JWKS の取得: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return fmt.Errorf("JWKS HTTP %d: %s", res.StatusCode, string(b))
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var doc jwksDoc
	if err := json.Unmarshal(body, &doc); err != nil {
		return fmt.Errorf("JWKS JSON: %w", err)
	}
	next := make(map[string]*rsa.PublicKey)
	for _, k := range doc.Keys {
		if k.Kty != "RSA" {
			continue
		}
		// Use が空なら署名鍵として扱う。enc など別用途の鍵は除外
		if k.Use != "" && k.Use != "sig" {
			continue
		}
		pub, err := rsaPublicKeyFromJWK(k.N, k.E)
		if err != nil {
			return fmt.Errorf("kid %q: %w", k.Kid, err)
		}
		if k.Kid != "" {
			next[k.Kid] = pub
		}
	}
	if len(next) == 0 {
		return fmt.Errorf("JWKS に利用可能な RSA 鍵がありません")
	}
	c.mu.Lock()
	c.keys = next
	c.fetched = time.Now()
	c.mu.Unlock()
	return nil
}

// getKey は JWT ヘッダの kid に対応する RSA 公開鍵を返す。
// キャッシュが空または TTL 切れなら refresh する。kid が見つからない場合は JWKS を 1 回だけ再取得してから再検索する
// （認可サーバー側で鍵を差し替えた直後など）。
func (c *jwksCache) getKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	c.mu.Lock()
	stale := len(c.keys) == 0 || time.Since(c.fetched) > c.ttl
	c.mu.Unlock()
	if stale {
		if err := c.refresh(ctx); err != nil {
			return nil, err
		}
	}
	c.mu.RLock()
	pub, ok := c.keys[kid]
	c.mu.RUnlock()
	if ok {
		return pub, nil
	}
	if err := c.refresh(ctx); err != nil {
		return nil, err
	}
	c.mu.RLock()
	pub, ok = c.keys[kid]
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("kid %q に対応する鍵がありません", kid)
	}
	return pub, nil
}
