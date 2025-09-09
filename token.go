package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// OAuth2トークンエンドポイント
func tokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	logger := slog.Default()

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

	if clientID == "" || clientSecret == "" {
		http.Error(w, "Client credentials required", http.StatusBadRequest)
		return
	}

	if code == "" {
		http.Error(w, "Authorization code required", http.StatusBadRequest)
		return
	}

	// クライアント認証
	_, err := repository.ValidateClientCredentials(ctx, clientID, clientSecret)
	if err != nil {
		logger.Warn("クライアント認証に失敗しました", "client_id", clientID, "error", err.Error())
		http.Error(w, "Invalid client credentials", http.StatusUnauthorized)
		return
	}

	// 認可コードを取得・検証（使用後に自動削除される）
	authCode, err := repository.GetAuthorizationCode(ctx, code)
	if err != nil {
		logger.Warn("認可コードの検証に失敗しました", "code", code, "error", err.Error())
		http.Error(w, "Invalid or expired authorization code", http.StatusBadRequest)
		return
	}

	// クライアントIDとリダイレクトURIの整合性チェック
	if authCode.ClientID != clientID {
		logger.Warn("認可コードのクライアントIDが一致しません",
			"auth_code_client_id", authCode.ClientID,
			"request_client_id", clientID)
		http.Error(w, "Invalid authorization code", http.StatusBadRequest)
		return
	}

	if authCode.RedirectURI != redirectURI {
		logger.Warn("認可コードのリダイレクトURIが一致しません",
			"auth_code_redirect_uri", authCode.RedirectURI,
			"request_redirect_uri", redirectURI)
		http.Error(w, "Invalid redirect_uri", http.StatusBadRequest)
		return
	}

	// PKCEコードチャレンジを検証（必須）
	if authCode.CodeChallenge != nil && *authCode.CodeChallenge != "" {
		if codeVerifier == "" {
			http.Error(w, "Code verifier required for PKCE", http.StatusBadRequest)
			return
		}

		var computedChallenge string
		if authCode.CodeChallengeMethod != nil && *authCode.CodeChallengeMethod == "S256" {
			// SHA256ハッシュ化
			hash := sha256.Sum256([]byte(codeVerifier))
			computedChallenge = base64.RawURLEncoding.EncodeToString(hash[:])
		} else {
			// plain method
			computedChallenge = codeVerifier
		}

		if computedChallenge != *authCode.CodeChallenge {
			logger.Warn("PKCEコードチャレンジの検証に失敗しました",
				"expected", *authCode.CodeChallenge,
				"computed", computedChallenge)
			http.Error(w, "Invalid code verifier", http.StatusBadRequest)
			return
		}

		logger.Info("PKCE検証が成功しました", "method", *authCode.CodeChallengeMethod)
	}

	// アクセストークンを生成してデータベースに保存
	accessToken := generateRandomString(32)
	scopes := []string{}
	for _, scope := range authCode.Scopes {
		if scope != "" {
			scopes = append(scopes, scope)
		}
	}

	expiresAt := time.Now().Add(1 * time.Hour) // 1時間有効
	createdToken, err := repository.CreateAccessToken(ctx, accessToken, clientID, &authCode.UserID, scopes, expiresAt)
	if err != nil {
		logger.Error("アクセストークンの作成に失敗しました", "error", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// リフレッシュトークンの生成（簡略化、本来はデータベースに保存）
	refreshToken := generateRandomString(32)

	// レスポンスを作成
	response := map[string]any{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    3600, // 秒単位
		"refresh_token": refreshToken,
	}

	// スコープが存在する場合は追加
	if len(scopes) > 0 {
		response["scope"] = strings.Join(scopes, " ")
	}

	// OpenID Connectのnonceがある場合はid_tokenを生成（簡略化）
	if authCode.Nonce != nil && *authCode.Nonce != "" {
		// 本来はJWTを生成するが、ここでは簡略化
		response["id_token"] = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwODAiLCJhdWQiOiJjbGllbnQxIiwic3ViIjoidXNlcjEiLCJub25jZSI6IiIrKmF1dGhDb2RlLk5vbmNlKyIsImV4cCI6MTYzNzIzNDU2N30.dummy_signature"
	}

	// HTTPヘッダーを設定
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	// レスポンスを送信
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("トークンレスポンスのエンコードに失敗しました", "error", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Info("アクセストークンを発行しました",
		"token_id", createdToken.ID,
		"client_id", clientID,
		"user_id", authCode.UserID,
		"scopes", scopes)
}
