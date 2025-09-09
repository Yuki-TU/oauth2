package main

import (
	"context"
	"log/slog"
	"time"
)

// セッションを作成
func createSession(userID int) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sessionID := generateRandomString(32)
	expiresAt := time.Now().Add(24 * time.Hour) // 24時間有効

	err := repository.CreateSession(ctx, sessionID, userID, expiresAt)
	if err != nil {
		// ログにエラーを記録し、空文字列を返す
		slog.Default().Error("セッション作成に失敗しました", "error", err.Error(), "user_id", userID)
		return ""
	}

	return sessionID
}

// セッションユーザーを取得
func getSessionUser(sessionID string) *Session {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	session, err := repository.GetSession(ctx, sessionID)
	if err != nil {
		// セッションが見つからない場合や期限切れの場合はnilを返す
		return nil
	}

	return session
}

// セッションを削除
func deleteSession(sessionID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := repository.DeleteSession(ctx, sessionID)
	if err != nil {
		slog.Default().Error("セッション削除に失敗しました", "error", err.Error(), "session_id", sessionID)
	}
}
