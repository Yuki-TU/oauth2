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

const (
	// アクセストークン（JWT）の有効期間。expires_in レスポンスと DB の expires_at に一致させる。
	accessTokenLifetime = time.Hour
	// リフレッシュトークン（不透明文字列）の有効期間。refresh_tokens.expires_at に保存する。
	refreshTokenLifetime = 30 * 24 * time.Hour
)

// OAuth2トークンエンドポイント（POST /token）。
// grant_type ごとに処理を分岐する。クライアント認証（client_id + client_secret）は全グラント共通。
func tokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	logger := slog.Default()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	grantType := r.FormValue("grant_type")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")

	if clientID == "" || clientSecret == "" {
		http.Error(w, "Client credentials required", http.StatusBadRequest)
		return
	}

	_, err := repository.ValidateClientCredentials(ctx, clientID, clientSecret)
	if err != nil {
		logger.Warn("クライアント認証に失敗しました", "client_id", clientID, "error", err.Error())
		http.Error(w, "Invalid client credentials", http.StatusUnauthorized)
		return
	}

	switch grantType {
	case "authorization_code":
		// RFC 6749 4.1.3: 認可コードをアクセストークン（＋任意でリフレッシュ）に交換
		handleAuthorizationCodeGrant(ctx, w, r, logger, clientID)
	case "refresh_token":
		// RFC 6749 6: リフレッシュトークンでアクセストークンを再発行（本実装ではローテーション）
		handleRefreshTokenGrant(ctx, w, r, logger, clientID)
	default:
		http.Error(w, "Unsupported grant_type", http.StatusBadRequest)
	}
}

// handleAuthorizationCodeGrant は認可コードグラントを処理する。
// 認可コードは GetAuthorizationCode 内で検証後に DB から削除される（ワンタイム）。
func handleAuthorizationCodeGrant(ctx context.Context, w http.ResponseWriter, r *http.Request, logger *slog.Logger, clientID string) {
	code := r.FormValue("code")
	redirectURI := r.FormValue("redirect_uri")
	codeVerifier := r.FormValue("code_verifier")

	if code == "" {
		http.Error(w, "Authorization code required", http.StatusBadRequest)
		return
	}

	authCode, err := repository.GetAuthorizationCode(ctx, code)
	if err != nil {
		logger.Warn("認可コードの検証に失敗しました", "code", code, "error", err.Error())
		http.Error(w, "Invalid or expired authorization code", http.StatusBadRequest)
		return
	}

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

	// PKCE: 認可リクエスト時に保存した code_challenge と code_verifier の整合を取る
	if authCode.CodeChallenge != nil && *authCode.CodeChallenge != "" {
		if codeVerifier == "" {
			http.Error(w, "Code verifier required for PKCE", http.StatusBadRequest)
			return
		}

		var computedChallenge string
		if authCode.CodeChallengeMethod != nil && *authCode.CodeChallengeMethod == "S256" {
			// S256: SHA256(code_verifier) を BASE64URL（パディングなし）
			hash := sha256.Sum256([]byte(codeVerifier))
			computedChallenge = base64.RawURLEncoding.EncodeToString(hash[:])
		} else {
			// plain: code_challenge と code_verifier が同一（非推奨だが互換のため）
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

	scopes := []string{}
	for _, scope := range authCode.Scopes {
		if scope != "" {
			scopes = append(scopes, scope)
		}
	}
	scopeString := strings.Join(scopes, " ")

	user, err := repository.GetUserByID(ctx, authCode.UserID)
	if err != nil {
		logger.Error("ユーザー情報の取得に失敗しました", "error", err.Error(), "userID", authCode.UserID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	accessToken, err := generateJWTAccessToken(authCode.UserID, user.Username, clientID, scopeString, accessTokenLifetime)
	if err != nil {
		logger.Error("JWTアクセストークンの生成に失敗しました", "error", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// access_tokens にメタデータ保存（token カラムには JWT 文字列そのものを格納）
	expiresAt := time.Now().Add(accessTokenLifetime)
	createdToken, err := repository.CreateAccessToken(ctx, accessToken, clientID, &authCode.UserID, scopes, expiresAt)
	if err != nil {
		logger.Error("アクセストークンの作成に失敗しました", "error", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// refresh_tokens は access_tokens.id に外部キーで紐づく（スキーマ上必須）
	refreshPlain := generateRandomString(32)
	refreshExpires := time.Now().Add(refreshTokenLifetime)
	if _, err := repository.CreateRefreshToken(ctx, refreshPlain, createdToken.ID, refreshExpires); err != nil {
		logger.Error("リフレッシュトークンの作成に失敗しました", "error", err.Error())
		// アクセスだけ先に INSERT 済みのため、孤立行を残さないよう失効させる
		if revErr := repository.RevokeAccessToken(ctx, accessToken); revErr != nil {
			logger.Error("ロールバック用アクセストークン失効に失敗", "error", revErr.Error())
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    int(accessTokenLifetime.Seconds()),
		"refresh_token": refreshPlain,
	}
	if len(scopes) > 0 {
		response["scope"] = strings.Join(scopes, " ")
	}
	// OpenID Connect: 認可時に nonce が付いていれば ID Token を同梱
	if authCode.Nonce != nil && *authCode.Nonce != "" {
		idToken, err := generateJWTIDToken(authCode.UserID, user.Username, clientID, *authCode.Nonce, accessTokenLifetime)
		if err != nil {
			logger.Warn("ID Token生成に失敗しました", "error", err.Error())
		} else {
			response["id_token"] = idToken
		}
	}

	writeTokenJSON(w, logger, response, "アクセストークンを発行しました",
		"token_id", createdToken.ID,
		"client_id", clientID,
		"user_id", authCode.UserID,
		"scopes", scopes,
	)
}

// handleRefreshTokenGrant はリフレッシュトークングラントを処理する。
// JWT 署名にユーザー名が必要なため、コミット前に bundle で user_id / scopes を解決する。
// 真正な排他は CommitRefreshRotation 内の FOR UPDATE + トランザクションで行う（二重使用を防ぐ）。
func handleRefreshTokenGrant(ctx context.Context, w http.ResponseWriter, r *http.Request, logger *slog.Logger, clientID string) {
	refreshPlain := r.FormValue("refresh_token")
	if refreshPlain == "" {
		http.Error(w, "refresh_token required", http.StatusBadRequest)
		return
	}

	// ロックなしの読み取り（失効・競合の最終判定は CommitRefreshRotation 側）
	bundle, err := repository.GetRefreshTokenBundle(ctx, refreshPlain, clientID)
	if err != nil {
		logger.Warn("リフレッシュトークンが無効です", "client_id", clientID, "error", err.Error())
		http.Error(w, "Invalid refresh token", http.StatusBadRequest)
		return
	}

	user, err := repository.GetUserByID(ctx, bundle.UserID)
	if err != nil {
		logger.Error("ユーザー情報の取得に失敗しました", "error", err.Error(), "userID", bundle.UserID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	scopeString := strings.Join(bundle.Scopes, " ")
	newAccessJWT, err := generateJWTAccessToken(bundle.UserID, user.Username, clientID, scopeString, accessTokenLifetime)
	if err != nil {
		logger.Error("JWTアクセストークンの生成に失敗しました", "error", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	newRefreshPlain := generateRandomString(32)
	accessExpires := time.Now().Add(accessTokenLifetime)
	refreshExpires := time.Now().Add(refreshTokenLifetime)

	// 新しい access / refresh を追加したうえで旧 access 行を DELETE（CASCADE で旧 refresh も削除）＝ローテーション
	if err := repository.CommitRefreshRotation(ctx, refreshPlain, clientID, newAccessJWT, newRefreshPlain, accessExpires, refreshExpires); err != nil {
		logger.Warn("リフレッシュトークンのローテーションに失敗しました", "error", err.Error())
		http.Error(w, "Invalid refresh token", http.StatusBadRequest)
		return
	}

	// リフレッシュ応答では id_token は付与しない（OIDC の推奨挙動は未実装）
	response := map[string]any{
		"access_token":  newAccessJWT,
		"token_type":    "Bearer",
		"expires_in":    int(accessTokenLifetime.Seconds()),
		"refresh_token": newRefreshPlain,
	}
	if len(bundle.Scopes) > 0 {
		response["scope"] = strings.Join(bundle.Scopes, " ")
	}

	writeTokenJSON(w, logger, response, "リフレッシュによりアクセストークンを再発行しました",
		"client_id", clientID,
		"user_id", bundle.UserID,
		"scopes", bundle.Scopes,
	)
}

// writeTokenJSON は OAuth 2.0 のトークンレスポンス用ヘッダを付与して JSON を書き出す。
func writeTokenJSON(w http.ResponseWriter, logger *slog.Logger, response map[string]any, logMsg string, logAttrs ...any) {
	w.Header().Set("Content-Type", "application/json")
	// RFC 6749 5.1: トークンエンドポイントのレスポンスはキャッシュさせない
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("トークンレスポンスのエンコードに失敗しました", "error", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Info(logMsg, logAttrs...)
}
