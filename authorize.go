package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
)

// OAuth2認可エンドポイント
func authorizeHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	logger := slog.Default()

	// デバッグ用: リクエストURLをログ出力
	logger.Info("認可リクエストを受信しました",
		"url", r.URL.String(),
		"raw_query", r.URL.RawQuery)

	// クエリパラメータを取得
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	responseType := r.URL.Query().Get("response_type")
	scope := r.URL.Query().Get("scope")
	state := r.URL.Query().Get("state")
	codeChallenge := r.URL.Query().Get("code_challenge")
	codeChallengeMethod := r.URL.Query().Get("code_challenge_method")
	nonce := r.URL.Query().Get("nonce")

	logger.Info("パラメータ解析結果",
		"client_id", clientID,
		"redirect_uri", redirectURI,
		"response_type", responseType,
		"scope", scope,
		"state", state)

	// 基本的なバリデーション
	if clientID == "" {
		http.Error(w, "client_id is required", http.StatusBadRequest)
		return
	}
	if redirectURI == "" {
		http.Error(w, "redirect_uri is required", http.StatusBadRequest)
		return
	}
	if responseType != "code" {
		http.Error(w, "Unsupported response_type", http.StatusBadRequest)
		return
	}

	// データベースからクライアント情報を取得
	client, err := repository.GetClientByID(ctx, clientID)
	if err != nil {
		logger.Warn("無効なクライアントID", "client_id", clientID, "error", err.Error())
		http.Error(w, "Invalid client_id", http.StatusBadRequest)
		return
	}

	// リダイレクトURIの検証
	validRedirectURI := slices.Contains(client.RedirectURIs, redirectURI)
	if !validRedirectURI {
		logger.Warn("無効なリダイレクトURI",
			"client_id", clientID,
			"redirect_uri", redirectURI,
			"allowed_uris", client.RedirectURIs)
		http.Error(w, "Invalid redirect_uri", http.StatusBadRequest)
		return
	}

	// セッションからユーザー情報を取得
	sessionCookie, err := r.Cookie("session_id")
	if err != nil {
		// 未ログインの場合、ログインページにリダイレクト
		loginURL := "/login?redirect=" + url.QueryEscape(r.URL.RequestURI())
		http.Redirect(w, r, loginURL, http.StatusFound)
		return
	}

	session := getSessionUser(sessionCookie.Value)
	if session == nil {
		// セッションが無効または期限切れの場合
		loginURL := "/login?redirect=" + url.QueryEscape(r.URL.RequestURI())
		http.Redirect(w, r, loginURL, http.StatusFound)
		return
	}

	// スコープの処理
	requestedScopes := []string{}
	if scope != "" {
		requestedScopes = strings.Fields(scope)
	}

	// 認可コードを生成してデータベースに保存
	authCode := generateRandomString(32)

	var codeChallengePt, codeChallengeMethodPtr, noncePt, statePt *string
	if codeChallenge != "" {
		codeChallengePt = &codeChallenge
	}
	if codeChallengeMethod != "" {
		codeChallengeMethodPtr = &codeChallengeMethod
	}
	if nonce != "" {
		noncePt = &nonce
	}
	if state != "" {
		statePt = &state
	}

	expiresAt := time.Now().Add(10 * time.Minute)
	err = repository.CreateAuthorizationCode(ctx, authCode, clientID, session.UserID, redirectURI, requestedScopes, codeChallengePt, codeChallengeMethodPtr, noncePt, statePt, expiresAt)
	if err != nil {
		logger.Error("認可コードの作成に失敗しました", "error", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info("認可コードが生成されました",
		"code", authCode,
		"client_id", clientID,
		"user_id", session.UserID,
		"scopes", requestedScopes)

	// 認可コードをクライアントにリダイレクト
	redirectURL := fmt.Sprintf("%s?code=%s", redirectURI, authCode)
	if state != "" {
		redirectURL += "&state=" + url.QueryEscape(state)
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}
