package main

import (
	"fmt"
	"net/http"
	"strings"
)

// デバッグ用のログハンドラー - すべてのリクエストを記録
func debugRequestHandler(w http.ResponseWriter, r *http.Request) {
	// ルートパスでない場合は404を返す
	if r.URL.Path != "/debug" {
		http.NotFound(w, r)
		return
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>デバッグ情報 - OAuth2 Server</title>
    <style>
        body {
            font-family: monospace;
            margin: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 { color: #333; }
        h2 { color: #666; border-bottom: 1px solid #eee; padding-bottom: 5px; }
        .info { margin: 10px 0; }
        .key { font-weight: bold; color: #007bff; }
        .value { color: #333; }
        pre { background: #f8f9fa; padding: 10px; border-left: 4px solid #007bff; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="container">
        <h1>リクエストデバッグ情報</h1>
        
        <h2>基本情報</h2>
        <div class="info"><span class="key">Method:</span> <span class="value">%s</span></div>
        <div class="info"><span class="key">URL:</span> <span class="value">%s</span></div>
        <div class="info"><span class="key">Path:</span> <span class="value">%s</span></div>
        <div class="info"><span class="key">Raw Query:</span> <span class="value">%s</span></div>
        <div class="info"><span class="key">Host:</span> <span class="value">%s</span></div>
        <div class="info"><span class="key">Remote Address:</span> <span class="value">%s</span></div>
        
        <h2>クエリパラメータ</h2>
        <pre>%s</pre>
        
        <h2>ヘッダー</h2>
        <pre>%s</pre>
        
        <h2>利用可能なエンドポイント</h2>
        <ul>
            <li><a href="/healthz">GET /healthz</a> - ヘルスチェック</li>
            <li><a href="/login">GET /login</a> - ログインページ</li>
            <li><a href="/signup">GET /signup</a> - サインアップページ</li>
            <li>/authorize - OAuth2認可エンドポイント</li>
            <li>/token - OAuth2トークンエンドポイント</li>
            <li>/callback - OAuth2コールバック</li>
        </ul>
        
        <h2>テスト用OAuth2 URL</h2>
        <p>以下のURLでOAuth2フローをテストできます：</p>
        <pre>http://localhost:8080/authorize?client_id=fdaaf3fdafd3fs&redirect_uri=http://localhost:3000/&response_type=code&scope=read</pre>
    </div>
</body>
</html>`,
		escapeHTML(r.Method),
		escapeHTML(r.URL.String()),
		escapeHTML(r.URL.Path),
		escapeHTML(r.URL.RawQuery),
		escapeHTML(r.Host),
		escapeHTML(r.RemoteAddr),
		escapeHTML(formatQueryParams(r.URL.Query())),
		escapeHTML(formatHeaders(r.Header)),
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// クエリパラメータを見やすい形式でフォーマット
func formatQueryParams(params map[string][]string) string {
	if len(params) == 0 {
		return "(パラメータなし)"
	}

	var lines []string
	for key, values := range params {
		for _, value := range values {
			lines = append(lines, fmt.Sprintf("%s = %s", key, value))
		}
	}
	return strings.Join(lines, "\n")
}

// ヘッダーを見やすい形式でフォーマット
func formatHeaders(headers map[string][]string) string {
	var lines []string
	for key, values := range headers {
		lines = append(lines, fmt.Sprintf("%s: %s", key, strings.Join(values, ", ")))
	}
	return strings.Join(lines, "\n")
}
