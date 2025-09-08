package main

import (
	"fmt"
	"net/http"
)

// ログイン画面表示用のGETハンドラー
func loginGetHandler(w http.ResponseWriter, r *http.Request) {
	// ログインフォームを表示（簡易的なHTML）
	redirectTo := r.URL.Query().Get("redirect") // 認証後に戻る先

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprintf(w, `<form method="POST" action="/login">
            <input type="hidden" name="redirect" value="%s"/>
            <label>User: <input name="user"></label><br/>
            <label>Password: <input type="password" name="pass"></label><br/>
            <button type="submit">Login</button>
        </form>`, redirectTo)
}
