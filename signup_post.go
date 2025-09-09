package main

import (
	"context"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// サインアップ処理用のPOSTハンドラー
func signupPostHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	logger := slog.Default()

	// フォームから送信されたデータを取得
	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")
	redirectTo := r.FormValue("redirect")

	// バリデーション
	if err := validateSignupInput(username, email, password, confirmPassword); err != nil {
		logger.Warn("サインアップ入力検証エラー", "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// データベースに新しいユーザーを作成
	user, err := repository.CreateUser(ctx, username, password, email)
	if err != nil {
		// ユニーク制約違反の場合
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			if strings.Contains(err.Error(), "username") {
				logger.Warn("ユーザー名が既に使用されています", "username", username)
				http.Error(w, "このユーザー名は既に使用されています", http.StatusConflict)
			} else if strings.Contains(err.Error(), "email") {
				logger.Warn("メールアドレスが既に使用されています", "email", email)
				http.Error(w, "このメールアドレスは既に使用されています", http.StatusConflict)
			} else {
				logger.Warn("重複エラー", "error", err.Error())
				http.Error(w, "このユーザー名またはメールアドレスは既に使用されています", http.StatusConflict)
			}
			return
		}

		logger.Error("ユーザー作成に失敗しました", "error", err.Error())
		http.Error(w, "アカウントの作成に失敗しました", http.StatusInternalServerError)
		return
	}

	// アカウント作成成功: 自動的にログインセッションを作成
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

	logger.Info("新しいユーザーが作成されました",
		"username", user.Username,
		"user_id", user.ID,
		"email", user.Email,
		"session_id", sessionID)

	// 元のリクエスト先またはデフォルトページへリダイレクト
	if redirectTo == "" {
		redirectTo = "/"
	}

	logger.Info("サインアップ後のリダイレクト",
		"redirect_to", redirectTo,
		"username", user.Username)

	http.Redirect(w, r, redirectTo, http.StatusFound)
}

// サインアップ入力のバリデーション
func validateSignupInput(username, email, password, confirmPassword string) error {
	// ユーザー名の検証
	if username == "" {
		return &ValidationError{"ユーザー名は必須です"}
	}
	if len(username) < 3 || len(username) > 50 {
		return &ValidationError{"ユーザー名は3-50文字で入力してください"}
	}
	// ユーザー名は英数字とアンダースコアのみ
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(username) {
		return &ValidationError{"ユーザー名は英数字とアンダースコアのみ使用できます"}
	}

	// メールアドレスの検証
	if email == "" {
		return &ValidationError{"メールアドレスは必須です"}
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return &ValidationError{"有効なメールアドレスを入力してください"}
	}

	// パスワードの検証
	if password == "" {
		return &ValidationError{"パスワードは必須です"}
	}
	if len(password) < 8 {
		return &ValidationError{"パスワードは8文字以上で入力してください"}
	}
	if len(password) > 128 {
		return &ValidationError{"パスワードは128文字以下で入力してください"}
	}

	// パスワード確認の検証
	if password != confirmPassword {
		return &ValidationError{"パスワードが一致しません"}
	}

	return nil
}

// バリデーションエラー用のカスタムエラー型
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
