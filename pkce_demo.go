package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
)

// PKCEデモページ - code_verifierとcode_challengeを生成
func pkceDemo(w http.ResponseWriter, r *http.Request) {
	// PKCE用のcode_verifierを生成（43-128文字のランダム文字列）
	codeVerifier := generateRandomString(32) // 32バイト → 43文字のbase64

	// code_challengeを生成（SHA256ハッシュ）
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	// デフォルト値
	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		clientID = "oauth2_demo_client"
	}

	clientSecret := r.URL.Query().Get("client_secret")
	if clientSecret == "" {
		clientSecret = "demo_client_secret_12345"
	}

	redirectURI := r.URL.Query().Get("redirect_uri")
	if redirectURI == "" {
		redirectURI = "http://localhost:3000/callback"
	}

	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = "read write openid profile"
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		state = "demo_state"
	}

	nonce := r.URL.Query().Get("nonce")
	if nonce == "" {
		nonce = "demo_nonce"
	}

	// OAuth2認可URLを生成
	authURL := fmt.Sprintf(
		"http://localhost:8080/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s&nonce=%s&code_challenge=%s&code_challenge_method=S256",
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape(scope),
		url.QueryEscape(state),
		url.QueryEscape(nonce),
		codeChallenge,
	)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PKCE デモ - OAuth2 Server</title>
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
        h1 { color: #333; text-align: center; }
        h2 { color: #666; border-bottom: 2px solid #007bff; padding-bottom: 10px; }
        .info { 
            background: #e7f3ff; 
            padding: 15px; 
            border-left: 4px solid #007bff; 
            margin: 20px 0; 
        }
        .code { 
            background: #f8f9fa; 
            padding: 15px; 
            border-radius: 4px; 
            font-family: monospace; 
            word-break: break-all;
            margin: 10px 0;
        }
        .btn {
            display: inline-block;
            padding: 12px 24px;
            background-color: #007bff;
            color: white;
            text-decoration: none;
            border-radius: 4px;
            margin: 10px 0;
        }
        .btn:hover { background-color: #0056b3; }
        pre {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 4px;
            overflow-x: auto;
            font-size: 14px;
        }
        .warning {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            border-radius: 4px;
            padding: 15px;
            margin: 15px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>PKCE デモページ</h1>
        
        <div class="info">
            <p><strong>PKCE (Proof Key for Code Exchange)</strong> は、OAuth2のセキュリティを向上させる拡張仕様です。</p>
        </div>

        <h2>1. OAuth2パラメータ設定</h2>
        
        <form method="GET" action="/pkce" style="background: #f8f9fa; padding: 20px; border-radius: 8px; margin: 20px 0;">
            <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 15px;">
                <div>
                    <label for="client_id" style="display: block; font-weight: bold; margin-bottom: 5px;">Client ID:</label>
                    <input type="text" id="client_id" name="client_id" value="%s" 
                           style="width: 100%%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;">
                </div>
                <div>
                    <label for="client_secret" style="display: block; font-weight: bold; margin-bottom: 5px;">Client Secret:</label>
                    <input type="text" id="client_secret" name="client_secret" value="%s" 
                           style="width: 100%%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;">
                </div>
                <div>
                    <label for="redirect_uri" style="display: block; font-weight: bold; margin-bottom: 5px;">Redirect URI:</label>
                    <input type="text" id="redirect_uri" name="redirect_uri" value="%s" 
                           style="width: 100%%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;">
                </div>
                <div>
                    <label for="scope" style="display: block; font-weight: bold; margin-bottom: 5px;">Scope:</label>
                    <input type="text" id="scope" name="scope" value="%s" 
                           style="width: 100%%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;">
                </div>
                <div>
                    <label for="state" style="display: block; font-weight: bold; margin-bottom: 5px;">State:</label>
                    <input type="text" id="state" name="state" value="%s" 
                           style="width: 100%%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;">
                </div>
                <div>
                    <label for="nonce" style="display: block; font-weight: bold; margin-bottom: 5px;">Nonce:</label>
                    <input type="text" id="nonce" name="nonce" value="%s" 
                           style="width: 100%%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;">
                </div>
            </div>
            <button type="submit" class="btn" style="margin-top: 15px;">パラメータを更新してPKCE再生成</button>
        </form>

        <h2>2. 生成されたPKCEパラメータ</h2>
        
        <h3>Code Verifier:</h3>
        <div class="code" id="codeVerifier">%s</div>
        <button onclick="copyToClipboard('codeVerifier')">コピー</button>
        
        <h3>Code Challenge:</h3>
        <div class="code">%s</div>
        
        <h3>Code Challenge Method:</h3>
        <div class="code">S256</div>

        <h2>3. OAuth2認可フロー開始</h2>
        
        <p>設定されたパラメータで認可フローを開始：</p>
        <a href="%s" class="btn" target="_blank">認可フローを開始</a>
        
        <div style="margin-top: 15px; padding: 15px; background: #e7f3ff; border-left: 4px solid #007bff; border-radius: 4px;">
            <strong>生成される認可URL:</strong><br>
            <code style="word-break: break-all; font-size: 12px;">%s</code>
        </div>
        
        <div class="warning">
            <strong>重要:</strong> 上記のCode Verifierを保存しておいてください。トークン交換時に必要です。
        </div>

        <h2>4. 認可完了後のトークン交換</h2>
        
        <p>認可が完了したら、以下のcurlコマンドでアクセストークンを取得してください：</p>
        
        <pre>
curl -X POST http://localhost:8080/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&code=&lt;認可コード&gt;&redirect_uri=%s&client_id=%s&client_secret=%s&code_verifier=%s"
        </pre>
        
        <p><small>&lt;認可コード&gt; を実際に取得した認可コードに置き換えてください</small></p>

        <h2>5. 新しいPKCEパラメータを生成</h2>
        <a href="/pkce" class="btn">新しいパラメータを生成</a>
        
        <hr style="margin: 30px 0;">
        <p><a href="/">← ホームに戻る</a> | <a href="/debug">デバッグ情報</a></p>
    </div>

    <script>
        function copyToClipboard(elementId) {
            const element = document.getElementById(elementId);
            const text = element.textContent;
            navigator.clipboard.writeText(text).then(function() {
                alert('クリップボードにコピーしました');
            });
        }
    </script>
</body>
</html>`,
		escapeHTML(clientID),
		escapeHTML(clientSecret),
		escapeHTML(redirectURI),
		escapeHTML(scope),
		escapeHTML(state),
		escapeHTML(nonce),
		escapeHTML(codeVerifier),
		escapeHTML(codeChallenge),
		authURL,
		escapeHTML(authURL),
		url.QueryEscape(redirectURI),
		clientID,
		clientSecret,
		codeVerifier)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
