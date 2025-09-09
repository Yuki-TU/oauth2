package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// ログイン認証処理用のPOSTハンドラー
func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	logger := slog.Default()

	// フォームから送信されたユーザー認証
	username := r.FormValue("user")
	password := r.FormValue("pass")
	redirectTo := r.FormValue("redirect")

	if username == "" || password == "" {
		http.Error(w, "ユーザー名とパスワードが必要です", http.StatusBadRequest)
		return
	}

	// データベースでユーザー認証
	user, err := repository.ValidateUserPassword(ctx, username, password)
	if err != nil {
		logger.Warn("ログイン失敗", "username", username, "error", err.Error())
		http.Error(w, "認証に失敗しました", http.StatusUnauthorized)
		return
	}

	// 認証成功: 新しいセッションIDを発行
	sessionID := createSession(user.ID)

	// セッションクッキーを設定
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // 本番環境ではtrueに設定
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 24時間
	})

	logger.Info("ユーザーがログインしました",
		"username", user.Username,
		"user_id", user.ID,
		"session_id", sessionID)

	// 元のリクエスト先またはデフォルトページへリダイレクト
	if redirectTo == "" {
		redirectTo = "/"
	}

	logger.Info("ログイン後のリダイレクト",
		"redirect_to", redirectTo,
		"username", user.Username)

	http.Redirect(w, r, redirectTo, http.StatusFound)
}
