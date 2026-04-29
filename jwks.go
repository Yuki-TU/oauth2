package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// JWKSエンドポイントハンドラー
func jwksHandler(w http.ResponseWriter, r *http.Request) {
	// JWKS JSONを生成
	jwksJSON, err := getJWKSJSON()
	if err != nil {
		slog.Error("JWKS生成エラー", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// レスポンスヘッダーを設定
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600") // 1時間キャッシュ
	w.Header().Set("Access-Control-Allow-Origin", "*")      // CORS対応

	// JWKSレスポンスを返す
	w.WriteHeader(http.StatusOK)
	w.Write(jwksJSON)

	slog.Info("JWKSエンドポイントにアクセスされました",
		"remoteAddr", r.RemoteAddr,
		"userAgent", r.UserAgent())
}

// /.well-known/jwksエンドポイントハンドラー（OpenID Connect Discovery用）
func wellKnownJwksHandler(w http.ResponseWriter, r *http.Request) {
	// JWKS JSONを生成
	jwksJSON, err := getJWKSJSON()
	if err != nil {
		slog.Error("Well-known JWKS生成エラー", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// レスポンスヘッダーを設定
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600") // 1時間キャッシュ
	w.Header().Set("Access-Control-Allow-Origin", "*")      // CORS対応

	// JWKSレスポンスを返す
	w.WriteHeader(http.StatusOK)
	w.Write(jwksJSON)

	slog.Info("Well-known JWKSエンドポイントにアクセスされました",
		"remoteAddr", r.RemoteAddr,
		"userAgent", r.UserAgent())
}

// OpenID Connect Discovery エンドポイント
func wellKnownOpenidConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	baseURL := issuerBaseURL()

	// OpenID Connect Discovery レスポンス（実装と整合する最小セット）
	discovery := map[string]any{
		"issuer":                 baseURL,
		"authorization_endpoint": baseURL + "/authorize",
		"token_endpoint":         baseURL + "/token",
		"jwks_uri":               baseURL + "/jwks",
		"userinfo_endpoint":      baseURL + "/userinfo",
		"response_types_supported": []string{
			"code",
		},
		"subject_types_supported": []string{
			"public",
		},
		"id_token_signing_alg_values_supported": []string{
			"RS256",
		},
		"scopes_supported": []string{
			"openid",
			"profile",
			"email",
			"read",
			"write",
		},
		"claims_supported": []string{
			"sub",
			"iss",
			"aud",
			"exp",
			"iat",
			"nonce",
			"username",
			"client_id",
			"scope",
		},
		"grant_types_supported": []string{
			"authorization_code",
			"refresh_token",
		},
		"code_challenge_methods_supported": []string{
			"S256",
			"plain",
		},
	}

	// レスポンスヘッダーを設定
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(discovery)

	slog.Info("OpenID Connect Discovery エンドポイントにアクセスされました",
		"remoteAddr", r.RemoteAddr,
		"userAgent", r.UserAgent())
}

// JWT トークン情報エンドポイント（デバッグ用）
func tokenInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authorizationヘッダーからトークンを取得
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusBadRequest)
		return
	}

	// "Bearer " プレフィックスを削除
	token := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	} else {
		http.Error(w, "Invalid authorization header format", http.StatusBadRequest)
		return
	}

	// JWTトークンを検証
	claims, err := validateJWTToken(token)
	if err != nil {
		slog.Warn("無効なトークンでtokenInfoにアクセス", "error", err, "remoteAddr", r.RemoteAddr)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// トークン情報をJSONで返す
	tokenInfo := fmt.Sprintf(`{
  "active": true,
  "sub": "%s",
  "username": "%s",
  "client_id": "%s",
  "scope": "%s",
  "exp": %d,
  "iat": %d,
  "iss": "%s"
}`,
		claims.Subject,
		claims.Username,
		claims.ClientID,
		claims.Scope,
		claims.ExpiresAt.Unix(),
		claims.IssuedAt.Unix(),
		claims.Issuer)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(tokenInfo))

	slog.Info("トークン情報が要求されました",
		"sub", claims.Subject,
		"clientID", claims.ClientID,
		"remoteAddr", r.RemoteAddr)
}
