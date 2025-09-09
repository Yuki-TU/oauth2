package main

import (
	"log/slog"
	"net/http"
)

// ログアウトハンドラー
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	logger := slog.Default()

	// セッションクッキーを取得
	sessionCookie, err := r.Cookie("session_id")
	if err == nil && sessionCookie.Value != "" {
		// データベースからセッションを削除
		deleteSession(sessionCookie.Value)
		logger.Info("ユーザーがログアウトしました", "session_id", sessionCookie.Value)
	}

	// セッションクッキーを削除
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1, // 即座に削除
	})

	// リダイレクト先を取得（デフォルトはホーム）
	redirectTo := r.URL.Query().Get("redirect")
	if redirectTo == "" {
		redirectTo = "/"
	}

	// ログアウト完了ページまたはリダイレクト
	http.Redirect(w, r, redirectTo, http.StatusFound)
}
