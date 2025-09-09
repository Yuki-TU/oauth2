package main

import (
	"time"

	"github.com/lib/pq"
)

// User はユーザー情報を表す構造体
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"password_hash"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// OAuthClient はOAuth2クライアント情報を表す構造体
type OAuthClient struct {
	ID           int            `json:"id"`
	ClientID     string         `json:"client_id"`
	ClientSecret string         `json:"client_secret"`
	Name         string         `json:"name"`
	RedirectURIs pq.StringArray `json:"redirect_uris"`
	Scopes       pq.StringArray `json:"scopes"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// AuthorizationCode は認可コード情報を表す構造体
type AuthorizationCode struct {
	ID                  int            `json:"id"`
	Code                string         `json:"code"`
	ClientID            string         `json:"client_id"`
	UserID              int            `json:"user_id"`
	RedirectURI         string         `json:"redirect_uri"`
	Scopes              pq.StringArray `json:"scopes"`
	CodeChallenge       *string        `json:"code_challenge"`
	CodeChallengeMethod *string        `json:"code_challenge_method"`
	Nonce               *string        `json:"nonce"`
	State               *string        `json:"state"`
	ExpiresAt           time.Time      `json:"expires_at"`
	CreatedAt           time.Time      `json:"created_at"`
}

// AccessToken はアクセストークン情報を表す構造体
type AccessToken struct {
	ID        int            `json:"id"`
	Token     string         `json:"token"`
	ClientID  string         `json:"client_id"`
	UserID    *int           `json:"user_id"`
	Scopes    pq.StringArray `json:"scopes"`
	ExpiresAt time.Time      `json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
}

// RefreshToken はリフレッシュトークン情報を表す構造体
type RefreshToken struct {
	ID            int       `json:"id"`
	Token         string    `json:"token"`
	AccessTokenID int       `json:"access_token_id"`
	ExpiresAt     time.Time `json:"expires_at"`
	CreatedAt     time.Time `json:"created_at"`
}

// Session はセッション情報を表す構造体
type Session struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}
