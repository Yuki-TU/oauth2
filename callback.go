package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// コールバックエンドポイント（デモ用）
func callbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	// クライアント情報を取得（セッション/クエリパラメータから）
	clientID := r.URL.Query().Get("client_id")
	clientSecret := r.URL.Query().Get("client_secret")
	redirectURI := r.URL.Query().Get("redirect_uri")
	codeVerifier := strings.TrimSpace(r.URL.Query().Get("code_verifier"))

	// デフォルト値を設定
	if clientID == "" {
		clientID = "oauth2_demo_client"
	}
	if clientSecret == "" {
		clientSecret = "demo_client_secret_12345"
	}
	if redirectURI == "" {
		redirectURI = "http://localhost:8080/callback"
	}

	pkceVals := url.Values{}
	pkceVals.Set("grant_type", "authorization_code")
	pkceVals.Set("code", code)
	pkceVals.Set("redirect_uri", redirectURI)
	pkceVals.Set("client_id", clientID)
	pkceVals.Set("client_secret", clientSecret)
	if codeVerifier != "" {
		pkceVals.Set("code_verifier", codeVerifier)
	} else {
		pkceVals.Set("code_verifier", "YOUR_CODE_VERIFIER")
	}
	curlPKCE := fmt.Sprintf(`curl -X POST http://localhost:8080/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d %s`, strconv.Quote(pkceVals.Encode()))

	noPKCEVals := url.Values{}
	noPKCEVals.Set("grant_type", "authorization_code")
	noPKCEVals.Set("code", code)
	noPKCEVals.Set("redirect_uri", redirectURI)
	noPKCEVals.Set("client_id", clientID)
	noPKCEVals.Set("client_secret", clientSecret)
	curlNoPKCE := fmt.Sprintf(`curl -X POST http://localhost:8080/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d %s`, strconv.Quote(noPKCEVals.Encode()))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if errorParam != "" {
		fmt.Fprintf(w, `<h1>認可エラー</h1><p>エラー: %s</p>`, errorParam)
		return
	}

	if code == "" {
		fmt.Fprintf(w, `<h1>認可エラー</h1><p>認可コードが取得できませんでした</p>`)
		return
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OAuth2認可完了 - OAuth2 Server</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 900px;
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
        h1 { color: #28a745; text-align: center; }
        h2 { color: #007bff; border-bottom: 2px solid #007bff; padding-bottom: 10px; }
        .success { 
            background: #d4edda; 
            border: 1px solid #c3e6cb; 
            padding: 15px; 
            border-radius: 4px; 
            margin: 20px 0;
        }
        .info { 
            background: #e7f3ff; 
            border-left: 4px solid #007bff; 
            padding: 15px; 
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
        pre {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 4px;
            overflow-x: auto;
            font-size: 14px;
            border: 1px solid #e9ecef;
        }
        .form-section {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            margin: 20px 0;
        }
        .form-group {
            margin: 15px 0;
        }
        label {
            display: block;
            font-weight: bold;
            margin-bottom: 5px;
            color: #495057;
        }
        input[type="text"] {
            width: 100%%;
            padding: 8px;
            border: 1px solid #ced4da;
            border-radius: 4px;
            font-family: monospace;
            font-size: 14px;
        }
        .btn {
            background-color: #007bff;
            color: white;
            padding: 10px 20px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
            margin: 5px;
        }
        .btn:hover { background-color: #0056b3; }
        .btn-copy { background-color: #28a745; }
        .btn-copy:hover { background-color: #1e7e34; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎉 OAuth2認可完了</h1>
        
        <div class="success">
            <p><strong>✅ 認可が正常に完了しました！</strong></p>
            <p><strong>認可コード:</strong> <code>%s</code></p>
            <p><strong>State:</strong> <code>%s</code></p>
        </div>

        <h2>📝 クライアント情報の設定</h2>
        
        <div class="form-section">
            <p>以下のフォームでクライアント情報を変更できます：</p>
            <form method="GET" action="/callback">
                <input type="hidden" name="code" value="%s">
                <input type="hidden" name="state" value="%s">
                
                <div class="form-group">
                    <label for="client_id">Client ID:</label>
                    <input type="text" id="client_id" name="client_id" value="%s">
                </div>
                
                <div class="form-group">
                    <label for="client_secret">Client Secret:</label>
                    <input type="text" id="client_secret" name="client_secret" value="%s">
                </div>
                
                <div class="form-group">
                    <label for="redirect_uri">Redirect URI:</label>
                    <input type="text" id="redirect_uri" name="redirect_uri" value="%s">
                </div>

                <div class="form-group">
                    <label for="code_verifier">Code Verifier（PKCE）:</label>
                    <input type="text" id="code_verifier" name="code_verifier" value="%s" placeholder="認可リクエスト時の code_verifier（S256 なら必須）" autocomplete="off" spellcheck="false">
                </div>
                
                <button type="submit" class="btn">curlコマンドを更新</button>
            </form>
        </div>

        <h2>🚀 トークン取得コマンド</h2>
        
        <div class="info">
            <strong>注意:</strong> PKCE（S256）の場合は認可時に使った <code>code_verifier</code> を上のフォームに入れて「curlコマンドを更新」すると、下の PKCE 用 curl に反映されます。未入力のときはプレースホルダ <code>YOUR_CODE_VERIFIER</code> のままです。
        </div>

        <h3>PKCEありの場合:</h3>
        <pre id="curlCommand">%s</pre>
        
        <button onclick="copyToClipboard('curlCommand')" class="btn btn-copy">📋 コピー</button>
        
        <h3>PKCEなしの場合:</h3>
        <pre id="curlCommandNoPKCE">%s</pre>
        
        <button onclick="copyToClipboard('curlCommandNoPKCE')" class="btn btn-copy">📋 コピー</button>

        <h2>📚 利用可能なクライアント</h2>
        
        <div class="info">
            <h4>メインデモクライアント:</h4>
            <p>Client ID: <code>oauth2_demo_client</code><br>
            Client Secret: <code>demo_client_secret_12345</code></p>
            
            <h4>SPAクライアント:</h4>
            <p>Client ID: <code>spa_client_example</code><br>
            Client Secret: <code>spa_secret_abcdef67890</code></p>
        </div>

        <hr style="margin: 30px 0;">
        <p><a href="/">← ホームに戻る</a> | <a href="/pkce">PKCEデモ</a></p>
    </div>

    <script>
        function copyToClipboard(elementId) {
            const element = document.getElementById(elementId);
            const text = element.textContent;
            navigator.clipboard.writeText(text).then(function() {
                alert('クリップボードにコピーしました！');
            });
        }
    </script>
</body>
</html>`,
		escapeHTML(code),
		escapeHTML(state),
		escapeHTML(code),
		escapeHTML(state),
		escapeHTML(clientID),
		escapeHTML(clientSecret),
		escapeHTML(redirectURI),
		escapeHTML(codeVerifier),
		escapeHTML(curlPKCE),
		escapeHTML(curlNoPKCE))

	w.Write([]byte(html))
}
