package main

import (
	"fmt"
	"net/http"
	"net/url"
)

// ログイン画面表示用のGETハンドラー
func loginGetHandler(w http.ResponseWriter, r *http.Request) {
	redirectTo := r.URL.Query().Get("redirect") // 認証後に戻る先
	if redirectTo == "" {
		redirectTo = "/"
	}

	// サインアップページと統一されたデザインのHTMLを生成
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ログイン - OAuth2 Server</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 400px;
            margin: 100px auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 40px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            text-align: center;
            color: #333;
            margin-bottom: 30px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            color: #555;
            font-weight: 500;
        }
        input[type="text"], input[type="password"] {
            width: 100%%;
            padding: 12px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 16px;
            box-sizing: border-box;
        }
        input[type="text"]:focus, input[type="password"]:focus {
            outline: none;
            border-color: #007bff;
            box-shadow: 0 0 0 2px rgba(0,123,255,0.25);
        }
        .btn {
            width: 100%%;
            padding: 12px;
            background-color: #007bff;
            color: white;
            border: none;
            border-radius: 4px;
            font-size: 16px;
            cursor: pointer;
            margin-top: 10px;
        }
        .btn:hover {
            background-color: #0056b3;
        }
        .link {
            text-align: center;
            margin-top: 20px;
        }
        .link a {
            color: #007bff;
            text-decoration: none;
        }
        .link a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>ログイン</h1>
        <form method="post" action="/login">
            <input type="hidden" name="redirect" value="%s">
            
            <div class="form-group">
                <label for="user">ユーザー名</label>
                <input type="text" id="user" name="user" required>
            </div>
            
            <div class="form-group">
                <label for="pass">パスワード</label>
                <input type="password" id="pass" name="pass" required>
            </div>
            
            <button type="submit" class="btn">ログイン</button>
        </form>
        
        <div class="link">
            <p>アカウントをお持ちでない方は <a href="/signup?redirect=%s">サインアップ</a></p>
        </div>
    </div>
</body>
</html>`, escapeHTML(redirectTo), url.QueryEscape(redirectTo))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
