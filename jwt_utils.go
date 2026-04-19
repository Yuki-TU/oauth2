package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT関連の設定
var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	keyID      = "oauth2-server-key-1" // JWKSで使用するkey ID
)

// JWTクレーム構造体
type CustomClaims struct {
	jwt.RegisteredClaims
	Scope    string `json:"scope,omitempty"`
	ClientID string `json:"client_id,omitempty"`
	Username string `json:"username,omitempty"`
}

// JWKSレスポンス用の構造体
type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string `json:"kty"` // Key Type (RSA)
	Use string `json:"use"` // Public Key Use (sig)
	Kid string `json:"kid"` // Key ID
	Alg string `json:"alg"` // Algorithm (RS256)
	N   string `json:"n"`   // Modulus
	E   string `json:"e"`   // Exponent
}

// RSA鍵ペアの初期化
func initJWTKeys() error {
	privateKeyPath := filepath.Join("certificate", "jwt_private.pem")
	publicKeyPath := filepath.Join("certificate", "jwt_public.pem")

	// 鍵ファイルが存在するかチェック
	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		slog.Info("RSA鍵ペアが存在しません。新しく生成します...")
		if err := generateRSAKeys(privateKeyPath, publicKeyPath); err != nil {
			return fmt.Errorf("RSA鍵ペア生成エラー: %v", err)
		}
	}

	// 秘密鍵を読み込み
	privateKeyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("秘密鍵ファイル読み込みエラー: %v", err)
	}

	privateKeyBlock, _ := pem.Decode(privateKeyData)
	if privateKeyBlock == nil {
		return fmt.Errorf("無効な秘密鍵形式")
	}

	privateKey, err = x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("秘密鍵パースエラー: %v", err)
	}

	// 公開鍵を秘密鍵から取得
	publicKey = &privateKey.PublicKey

	slog.Info("RSA鍵ペアが正常に読み込まれました", "keyID", keyID)
	return nil
}

// RSA鍵ペアを生成
func generateRSAKeys(privateKeyPath, publicKeyPath string) error {
	// certificateディレクトリを作成
	if err := os.MkdirAll("certificate", 0755); err != nil {
		return fmt.Errorf("certificateディレクトリ作成エラー: %v", err)
	}

	// 2048ビットのRSA鍵ペアを生成
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("RSA鍵生成エラー: %v", err)
	}

	// 秘密鍵をPEM形式で保存
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(key)
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	privateKeyFile, err := os.Create(privateKeyPath)
	if err != nil {
		return fmt.Errorf("秘密鍵ファイル作成エラー: %v", err)
	}
	defer privateKeyFile.Close()

	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		return fmt.Errorf("秘密鍵PEMエンコードエラー: %v", err)
	}

	// 公開鍵をPEM形式で保存
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return fmt.Errorf("公開鍵マーシャルエラー: %v", err)
	}

	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	publicKeyFile, err := os.Create(publicKeyPath)
	if err != nil {
		return fmt.Errorf("公開鍵ファイル作成エラー: %v", err)
	}
	defer publicKeyFile.Close()

	if err := pem.Encode(publicKeyFile, publicKeyPEM); err != nil {
		return fmt.Errorf("公開鍵PEMエンコードエラー: %v", err)
	}

	slog.Info("RSA鍵ペアが生成されました",
		"privateKey", privateKeyPath,
		"publicKey", publicKeyPath)

	return nil
}

// JWTアクセストークンを生成
func generateJWTAccessToken(userID int, username, clientID, scope string, expiresIn time.Duration) (string, error) {
	if privateKey == nil {
		return "", fmt.Errorf("RSA秘密鍵が初期化されていません")
	}

	now := time.Now()
	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "oauth2-server",
			Subject:   fmt.Sprintf("%d", userID),
			Audience:  []string{clientID},
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        generateRandomString(16), // JTI (JWT ID)
		},
		Scope:    scope,
		ClientID: clientID,
		Username: username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = keyID

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("JWT署名エラー: %v", err)
	}

	slog.Info("JWTアクセストークンを生成しました",
		"userID", userID,
		"clientID", clientID,
		"scope", scope,
		"expiresIn", expiresIn)

	return tokenString, nil
}

// OpenID Connect ID Token を生成
func generateJWTIDToken(userID int, username, clientID, nonce string, expiresIn time.Duration) (string, error) {
	if privateKey == nil {
		return "", fmt.Errorf("RSA秘密鍵が初期化されていません")
	}

	now := time.Now()
	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "oauth2-server",
			Subject:   fmt.Sprintf("%d", userID),
			Audience:  []string{clientID},
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        generateRandomString(16), // JTI (JWT ID)
		},
		Username: username,
		ClientID: clientID,
	}

	// OpenID Connect用のクレームを追加
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = keyID

	// nonceクレームを追加
	if nonce != "" {
		// TODO: nonceクレームを適切に追加
		// 現在は簡略化のため省略
	}

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("ID Token JWT署名エラー: %v", err)
	}

	slog.Info("OpenID Connect ID Tokenを生成しました",
		"userID", userID,
		"clientID", clientID,
		"nonce", nonce,
		"expiresIn", expiresIn)

	return tokenString, nil
}

// JWTトークンを検証
func validateJWTToken(tokenString string) (*CustomClaims, error) {
	if publicKey == nil {
		return nil, fmt.Errorf("RSA公開鍵が初期化されていません")
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 署名方法がRS256であることを確認
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("予期しない署名方法: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("JWTトークン検証エラー: %v", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("無効なJWTトークン")
}

// JWKS形式の公開鍵を生成
func generateJWKS() (*JWKSResponse, error) {
	if publicKey == nil {
		return nil, fmt.Errorf("RSA公開鍵が初期化されていません")
	}

	// RSA公開鍵のN（modulus）とE（exponent）を取得
	nBytes := publicKey.N.Bytes()
	eBytes := big.NewInt(int64(publicKey.E)).Bytes()

	// Base64URL エンコード
	n := base64.RawURLEncoding.EncodeToString(nBytes)
	e := base64.RawURLEncoding.EncodeToString(eBytes)

	jwk := JWK{
		Kty: "RSA",
		Use: "sig",
		Kid: keyID,
		Alg: "RS256",
		N:   n,
		E:   e,
	}

	jwks := &JWKSResponse{
		Keys: []JWK{jwk},
	}

	return jwks, nil
}

// JWKS JSONレスポンスを生成
func getJWKSJSON() ([]byte, error) {
	jwks, err := generateJWKS()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(jwks, "", "  ")
}
