package main

import (
	"fmt"
	"net/http"
)

// ホームページハンドラー
func homeHandler(w http.ResponseWriter, r *http.Request) {
	// ルートパス以外は404
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	html := `
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OAuth2 Server</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 800px;
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
        h1 { 
            color: #333; 
            text-align: center; 
            margin-bottom: 40px;
        }
        .section {
            margin: 30px 0;
            padding: 20px;
            background: #f8f9fa;
            border-radius: 8px;
        }
        .section h2 {
            color: #007bff;
            margin-top: 0;
        }
        .links {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin: 20px 0;
        }
        .link-card {
            background: white;
            padding: 20px;
            border-radius: 8px;
            border: 1px solid #ddd;
            text-decoration: none;
            color: #333;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        .link-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
            text-decoration: none;
            color: #333;
        }
        .link-card h3 {
            color: #007bff;
            margin: 0 0 10px 0;
        }
        .link-card p {
            margin: 0;
            font-size: 14px;
            color: #666;
        }
        .feature-list {
            list-style: none;
            padding: 0;
        }
        .feature-list li {
            padding: 8px 0;
            border-bottom: 1px solid #eee;
        }
        .feature-list li:before {
            content: "✅ ";
            margin-right: 10px;
        }
        .footer {
            text-align: center;
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #eee;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔐 OAuth2 認可サーバー</h1>
        
        <div class="section">
            <h2>🚀 クイックスタート</h2>
            <div class="links">
                <a href="/pkce" class="link-card">
                    <h3>PKCE デモ</h3>
                    <p>PKCEを使用したOAuth2フローを体験</p>
                </a>
                <a href="/login" class="link-card">
                    <h3>ログイン</h3>
                    <p>既存のアカウントでログイン</p>
                </a>
                <a href="/signup" class="link-card">
                    <h3>サインアップ</h3>
                    <p>新しいアカウントを作成</p>
                </a>
            </div>
        </div>

        <div class="section">
            <h2>🛠️ 開発者向けツール</h2>
            <div class="links">
                <a href="/debug" class="link-card">
                    <h3>デバッグ情報</h3>
                    <p>リクエスト情報とエンドポイント一覧</p>
                </a>
                <a href="/healthz" class="link-card">
                    <h3>ヘルスチェック</h3>
                    <p>サーバーとデータベースの状態確認</p>
                </a>
            </div>
        </div>

        <div class="section">
            <h2>📋 機能一覧</h2>
            <ul class="feature-list">
                <li>OAuth2 Authorization Code フロー</li>
                <li>PKCE (Proof Key for Code Exchange) サポート</li>
                <li>PostgreSQL データベース連携</li>
                <li>ユーザー登録・ログイン機能</li>
                <li>セッション管理</li>
                <li>セキュアなパスワードハッシュ化</li>
                <li>期限切れトークンの自動クリーンアップ</li>
                <li>構造化ログ</li>
            </ul>
        </div>

        <div class="section">
            <h2>🔗 API エンドポイント</h2>
            <ul class="feature-list">
                <li><strong>GET /authorize</strong> - OAuth2認可エンドポイント</li>
                <li><strong>POST /token</strong> - OAuth2トークンエンドポイント</li>
                <li><strong>GET /callback</strong> - OAuth2コールバック</li>
                <li><strong>GET|POST /login</strong> - ログイン</li>
                <li><strong>GET|POST /signup</strong> - ユーザー登録</li>
            </ul>
        </div>

        <div class="footer">
            <p>OAuth2 Server v1.0 - Powered by Go & PostgreSQL</p>
        </div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}
