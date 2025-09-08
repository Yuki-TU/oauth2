package main

import (
	"log"
	"net/http"
)

// ログイン認証処理用のPOSTハンドラー
func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	// フォームから送信されたユーザー認証
	user := r.FormValue("user")
	pass := r.FormValue("pass")
	redirectTo := r.FormValue("redirect")
	// 認証失敗
	if user != demoUserID || pass != demoPassword {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 認証成功: 新しいセッションIDを発行
	sessionID := generateRandomString(16) // 16バイトランダム→base64文字列
	sessions[sessionID] = demoUserID
	// SecureやSameSiteの属性は省略（必要に応じ設定）
	http.SetCookie(w, &http.Cookie{
		Name: "session_id", Value: sessionID, Path: "/", HttpOnly: true,
	})
	log.Printf("User '%s' logged in, set session %s", user, sessionID)
	// 保存されていた元のリクエスト先へリダイレクト
	http.Redirect(w, r, redirectTo, http.StatusFound)
}
