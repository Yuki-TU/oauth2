package main

import (
	"fmt"
	"net/http"
	"net/url"
)

// サインアップページ表示用のGETハンドラー
func signupGetHandler(w http.ResponseWriter, r *http.Request) {
	redirectTo := r.URL.Query().Get("redirect")
	if redirectTo == "" {
		redirectTo = "/"
	}

	// 簡単なHTMLフォームを生成
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>サインアップ - OAuth2 Server</title>
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
        input[type="text"], input[type="email"], input[type="password"] {
            width: 100%%;
            padding: 12px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 16px;
            box-sizing: border-box;
        }
        input[type="text"]:focus, input[type="email"]:focus, input[type="password"]:focus {
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
        .help-text {
            font-size: 14px;
            color: #666;
            margin-top: 5px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>サインアップ</h1>
        <form method="post" action="/signup">
            <input type="hidden" name="redirect" value="%s">
            
            <div class="form-group">
                <label for="username">ユーザー名</label>
                <input type="text" id="username" name="username" required>
                <div class="help-text">3-50文字の英数字とアンダースコア</div>
            </div>
            
            <div class="form-group">
                <label for="email">メールアドレス</label>
                <input type="email" id="email" name="email" required>
                <div class="help-text">有効なメールアドレスを入力してください</div>
            </div>
            
            <div class="form-group">
                <label for="password">パスワード</label>
                <input type="password" id="password" name="password" required minlength="8">
                <div class="help-text">8文字以上のパスワード</div>
            </div>
            
            <div class="form-group">
                <label for="confirm_password">パスワード（確認）</label>
                <input type="password" id="confirm_password" name="confirm_password" required minlength="8">
                <div class="help-text">上記と同じパスワードを入力してください</div>
            </div>
            
            <button type="submit" class="btn">アカウント作成</button>
        </form>
        
        <div class="link">
            <p>すでにアカウントをお持ちですか？ <a href="/login?redirect=%s">ログイン</a></p>
        </div>
    </div>

    <script>
        // パスワード確認のクライアントサイド検証
        document.getElementById('confirm_password').addEventListener('input', function() {
            const password = document.getElementById('password').value;
            const confirmPassword = this.value;
            
            if (password !== confirmPassword) {
                this.setCustomValidity('パスワードが一致しません');
            } else {
                this.setCustomValidity('');
            }
        });
    </script>
</body>
</html>`, escapeHTML(redirectTo), url.QueryEscape(redirectTo))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
