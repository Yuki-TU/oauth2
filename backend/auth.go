package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// accessClaims は認可サーバー側 jwt_utils.CustomClaims と同じ JSON 形を想定した検証用構造体。
// リソースサーバーは秘密鍵を持たず、JWKS の公開鍵だけで署名検証する。
type accessClaims struct {
	jwt.RegisteredClaims
	Scope    string `json:"scope,omitempty"`
	ClientID string `json:"client_id,omitempty"`
	Username string `json:"username,omitempty"`
}

// bearerToken は Authorization ヘッダから Bearer トークン文字列を取り出す。
func bearerToken(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", false
	}
	const p = "Bearer "
	if len(h) < len(p) || !strings.EqualFold(h[:len(p)], p) {
		return "", false
	}
	t := strings.TrimSpace(h[len(p):])
	if t == "" {
		return "", false
	}
	return t, true
}

// expectedIssuer は jwt.WithIssuer に渡す値。認可サーバーが付与している iss と一致させる。
func expectedIssuer() string {
	if v := os.Getenv("RESOURCE_EXPECTED_ISS"); v != "" {
		return v
	}
	return "oauth2-server"
}

// allowedAudiences は jwt.WithAudience に渡す許可リスト。
// 空のときは aud を検証しない（デモ向け。本番では自サービスを表す audience を指定することが多い）。
func allowedAudiences() []string {
	raw := strings.TrimSpace(os.Getenv("RESOURCE_ALLOWED_AUDIENCES"))
	if raw == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// parseAndValidateAccessToken は JWT をパースし、署名（RS256）・iss・aud（任意）・exp を確認する。
// 署名検証用の鍵は JWT ヘッダの kid に対応する公開鍵を cache から取得する。
func parseAndValidateAccessToken(ctx context.Context, cache *jwksCache, tokenString string) (*accessClaims, error) {
	issuer := expectedIssuer()
	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
		jwt.WithIssuer(issuer),
	}
	if aud := allowedAudiences(); len(aud) > 0 {
		opts = append(opts, jwt.WithAudience(aud...))
	}
	parser := jwt.NewParser(opts...)

	claims := &accessClaims{}
	// 署名検証用の鍵を取得
	// 鍵は JWT ヘッダの kid に対応する公開鍵を cache から取得する。
	_, err := parser.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, fmt.Errorf("JWT ヘッダに kid がありません")
		}
		return cache.getKey(ctx, kid)
	})
	if err != nil {
		return nil, err
	}
	// ライブラリの検証に加え、期限を明示チェック（時計のずれ対策は未実装の簡易版）
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("トークンが期限切れです")
	}
	return claims, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
