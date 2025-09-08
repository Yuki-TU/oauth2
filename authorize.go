package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

// OAuth2認可エンドポイント
func authorizeHandler(w http.ResponseWriter, r *http.Request) {
	// クエリパラメータを取得
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	responseType := r.URL.Query().Get("response_type")
	scope := r.URL.Query().Get("scope")
	state := r.URL.Query().Get("state")
	codeChallenge := r.URL.Query().Get("code_challenge")
	codeChallengeMethod := r.URL.Query().Get("code_challenge_method")
	nonce := r.URL.Query().Get("nonce")

	// 基本的なバリデーション
	if clientID != demoClient.ID {
		http.Error(w, "Invalid client_id", http.StatusBadRequest)
		return
	}
	if redirectURI != demoClient.RedirectURI {
		http.Error(w, "Invalid redirect_uri", http.StatusBadRequest)
		return
	}
	if responseType != "code" {
		http.Error(w, "Unsupported response_type", http.StatusBadRequest)
		return
	}

	// セッションからユーザーIDを取得
	sessionCookie, err := r.Cookie("session_id")
	if err != nil || sessions[sessionCookie.Value] == "" {
		// 未ログインの場合、ログインページにリダイレクト
		loginURL := "/login?redirect=" + url.QueryEscape(r.URL.RequestURI())
		http.Redirect(w, r, loginURL, http.StatusFound)
		return
	}

	userID := sessions[sessionCookie.Value]

	// 認可コードを生成
	authCode := generateRandomString(32)
	authCodes[authCode] = AuthCodeData{
		ClientID:            clientID,
		UserID:              userID,
		RedirectURI:         redirectURI,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		Nonce:               nonce,
		Expiry:              time.Now().Add(10 * time.Minute),
		Scope:               scope,
	}

	log.Printf("Generated auth code %s for user %s", authCode, userID)

	// 認可コードをクライアントにリダイレクト
	redirectURL := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, authCode, state)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
