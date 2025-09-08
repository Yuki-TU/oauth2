package main

import (
	"fmt"
	"net/http"
)

// コールバックエンドポイント（デモ用）
func callbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if errorParam != "" {
		fmt.Fprintf(w, `<h1>認可エラー</h1><p>エラー: %s</p>`, errorParam)
		return
	}

	if code == "" {
		fmt.Fprintf(w, `<h1>認可エラー</h1><p>認可コードが取得できませんでした</p>`)
		return
	}

	fmt.Fprintf(w, `
		<h1>OAuth2認可完了</h1>
		<p><strong>認可コード:</strong> %s</p>
		<p><strong>State:</strong> %s</p>
		<p>認可が正常に完了しました。このコードを使用してアクセストークンを取得できます。</p>
		<hr>
		<h2>トークン取得例</h2>
		<pre>
curl -X POST http://localhost:8080/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&code=%s&redirect_uri=http://localhost:8080/callback&client_id=client1&client_secret=secret"
		</pre>
	`, code, state, code)
}
