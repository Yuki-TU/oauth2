package main

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

// **クライアント**情報の構造体
type Client struct {
	ID          string
	Secret      string
	RedirectURI string
}

// **認可コード**に対応して保存する情報の構造体
type AuthCodeData struct {
	ClientID            string
	UserID              string
	RedirectURI         string
	CodeChallenge       string // PKCEコードチャレンジ
	CodeChallengeMethod string // "S256" or "plain"
	Nonce               string
	Expiry              time.Time
	Scope               string
}

// **グローバル設定**：デモ用クライアントとユーザー、ストレージ
var demoClient = Client{
	ID:          "client1",
	Secret:      "secret",
	RedirectURI: "http://localhost:8080/callback",
}
var demoUserID = "user1"
var demoPassword = "pass1"

// メモリ上のストレージ
var authCodes = make(map[string]AuthCodeData) // 認可コード -> 情報
var sessions = make(map[string]string)        // セッションID -> UserID（ログイン済みユーザー）

// 指定バイト数の暗号論的ランダム文字列を生成（URL安全なbase64エンコード）
func generateRandomString(byteLen int) string {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	// URLセーフなbase64でエンコード（余計な=は付かない）
	return base64.RawURLEncoding.EncodeToString(b)
}
