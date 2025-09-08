package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// OAuth2トークンエンドポイント
func tokenHandler(w http.ResponseWriter, r *http.Request) {
	// フォームデータを解析
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	grantType := r.FormValue("grant_type")
	code := r.FormValue("code")
	redirectURI := r.FormValue("redirect_uri")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	codeVerifier := r.FormValue("code_verifier")

	// 基本的なバリデーション
	if grantType != "authorization_code" {
		http.Error(w, "Unsupported grant_type", http.StatusBadRequest)
		return
	}
	if clientID != demoClient.ID || clientSecret != demoClient.Secret {
		http.Error(w, "Invalid client credentials", http.StatusUnauthorized)
		return
	}

	// 認可コードを検証
	authData, exists := authCodes[code]
	if !exists {
		http.Error(w, "Invalid authorization code", http.StatusBadRequest)
		return
	}
	if time.Now().After(authData.Expiry) {
		delete(authCodes, code)
		http.Error(w, "Authorization code expired", http.StatusBadRequest)
		return
	}
	if authData.RedirectURI != redirectURI {
		http.Error(w, "Invalid redirect_uri", http.StatusBadRequest)
		return
	}

	// PKCEコードチャレンジを検証（省略可能）
	if authData.CodeChallenge != "" {
		// 実際の実装では、codeVerifierをハッシュ化してcodeChallengeと比較
		log.Printf("PKCE verification: challenge=%s, verifier=%s", authData.CodeChallenge, codeVerifier)
	}

	// 認可コードを削除（一度だけ使用可能）
	delete(authCodes, code)

	// アクセストークンを生成
	accessToken := generateRandomString(32)
	refreshToken := generateRandomString(32)

	// レスポンスを作成
	response := map[string]any{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": refreshToken,
		"scope":         authData.Scope,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding token response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	log.Printf("Issued access token for user %s", authData.UserID)
}
