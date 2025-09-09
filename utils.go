package main

import (
	"crypto/rand"
	"encoding/base64"
	"html"
)

// 指定バイト数の暗号論的ランダム文字列を生成（URL安全なbase64エンコード）
func generateRandomString(byteLen int) string {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	// URLセーフなbase64でエンコード（余計な=は付かない）
	return base64.RawURLEncoding.EncodeToString(b)
}

// HTMLエスケープ用のヘルパー関数
func escapeHTML(s string) string {
	return html.EscapeString(s)
}
