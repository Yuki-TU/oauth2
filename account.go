package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// accountHandler はログイン済みユーザーのみ閲覧できるマイアカウントページです。
func accountHandler(w http.ResponseWriter, r *http.Request) {
	session := requireBrowserSession(w, r)
	if session == nil {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	user, err := repository.GetUserByID(ctx, session.UserID)
	if err != nil {
		slog.Default().Error("account: user lookup failed", "error", err, "user_id", session.UserID)
		http.Error(w, "ユーザー情報の取得に失敗しました", http.StatusInternalServerError)
		return
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>マイアカウント - OAuth2 Server</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 520px;
            margin: 50px auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 40px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 { color: #333; margin-top: 0; font-size: 1.5rem; }
        .badge {
            display: inline-block;
            background: #e7f3ff;
            color: #0066cc;
            font-size: 0.85rem;
            padding: 4px 10px;
            border-radius: 4px;
            margin-bottom: 24px;
        }
        dl { margin: 0; }
        dt { color: #666; font-size: 0.85rem; margin-top: 16px; }
        dt:first-child { margin-top: 0; }
        dd { margin: 4px 0 0 0; font-size: 1.05rem; color: #222; }
        .actions { margin-top: 32px; display: flex; gap: 12px; flex-wrap: wrap; }
        .actions a {
            display: inline-block;
            padding: 10px 18px;
            border-radius: 6px;
            text-decoration: none;
            font-size: 0.95rem;
        }
        .primary { background: #007bff; color: white; }
        .primary:hover { background: #0069d9; color: white; }
        .secondary { background: #f0f0f0; color: #333; border: 1px solid #ddd; }
        .secondary:hover { background: #e8e8e8; color: #333; }
    </style>
</head>
<body>
    <div class="container">
        <div class="badge">ログイン中のみ表示</div>
        <h1>マイアカウント</h1>
        <dl>
            <dt>ユーザー名</dt>
            <dd>%s</dd>
            <dt>メールアドレス</dt>
            <dd>%s</dd>
            <dt>ユーザー ID</dt>
            <dd>%d</dd>
            <dt>セッション有効期限</dt>
            <dd>%s</dd>
        </dl>
        <div class="actions">
            <a class="primary" href="/">トップへ</a>
            <a class="secondary" href="/logout?redirect=/account">ログアウト</a>
        </div>
    </div>
</body>
</html>`,
		escapeHTML(user.Username),
		escapeHTML(user.Email),
		user.ID,
		escapeHTML(session.ExpiresAt.Format(time.RFC3339)),
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}
