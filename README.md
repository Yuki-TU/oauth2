# OAuth2 認可サーバー

PostgreSQLデータベースを使用したOAuth2認可サーバーの実装です。

## 機能

- OAuth2 Authorization Code フロー
- PKCE（Proof Key for Code Exchange）サポート
- PostgreSQLデータベースによる永続化
- セッション管理
- ヘルスチェックエンドポイント
- 期限切れトークンの自動クリーンアップ

## セットアップ

### 1. 依存関係のインストール

```bash
make deps
```

### 2. データベースの起動

```bash
make up
```

### 3. アプリケーションの実行

```bash
make run
```

## エンドポイント

- `GET /` - ホームページ
- `GET /healthz` - ヘルスチェック
- `GET /debug` - デバッグ情報
- `GET /pkce` - PKCEデモページ
- `GET /login` - ログインページ
- `POST /login` - ログイン処理
- `GET /signup` - サインアップページ
- `POST /signup` - ユーザー登録処理
- `GET|POST /logout` - ログアウト処理
- `GET /authorize` - OAuth2認可エンドポイント
- `POST /token` - OAuth2トークンエンドポイント
- `GET /callback` - OAuth2コールバック

## データベース接続

デフォルトの接続設定：
- Host: localhost
- Port: 5432
- Database: oauth2_db
- User: oauth2_user
- Password: oauth2_password

環境変数で設定を変更できます：
- `DB_HOST`
- `DB_PORT`
- `DB_NAME`
- `DB_USER`
- `DB_PASSWORD`
- `DB_SSLMODE`

## Makeコマンド

- `make up` - Dockerサービスを起動
- `make down` - Dockerサービスを停止
- `make run` - アプリケーションを実行
- `make build` - アプリケーションをビルド
- `make deps` - 依存関係をインストール
- `make db` - データベースに接続
- `make logs` - ログを表示
- `make test` - テストを実行

## テストデータ

`init.sql` には以下のテストデータが含まれています：

### テストユーザー
- **testuser** / password123 (test@example.com)
- **demo** / password123 (demo@example.com)  
- **admin** / password123 (admin@example.com)

### テストクライアント

#### メインデモクライアント
- Client ID: oauth2_demo_client
- Client Secret: demo_client_secret_12345
- Name: OAuth2 Demo Application
- Redirect URIs: 
  - http://localhost:3000/callback
  - http://localhost:3000/auth/callback
  - https://oauthdebugger.com/debug
- Scopes: read, write, openid, profile, email

#### SPAクライアント（PKCE推奨）
- Client ID: spa_client_example
- Client Secret: spa_secret_abcdef67890
- Name: Single Page Application
- Redirect URIs:
  - http://localhost:8080/callback
  - http://127.0.0.1:8080/callback
- Scopes: read, profile

#### その他のクライアント
- **モバイルアプリ**: mobile_app_client
- **管理コンソール**: admin_console

## 使用例

### 簡単なテスト方法

1. **ホームページにアクセス**: `http://localhost:8080/`
2. **PKCEデモを試す**: `http://localhost:8080/pkce` でcode_verifierとcode_challengeを自動生成
3. 認可フローを実行して、表示されるcurlコマンドでトークンを取得

### 新規ユーザー登録
1. ブラウザで `http://localhost:8080/signup` にアクセス
2. ユーザー名、メールアドレス、パスワードを入力
3. アカウント作成後、自動的にログイン状態になる

### 手動でのOAuth2フロー

#### メインデモクライアント（PKCEありの場合・推奨）
1. ブラウザで認可エンドポイントにアクセス：
```
http://localhost:8080/authorize?client_id=oauth2_demo_client&redirect_uri=http://localhost:3000/callback&response_type=code&scope=read%20write%20openid&state=xyz123&nonce=n-0S6_WzA2Mj&code_challenge=E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM&code_challenge_method=S256
```

2. ログイン（既存ユーザー: testuser/password123 または新規作成したユーザー）

3. 認可コードが redirect_uri に送信される

4. トークンエンドポイントで認可コードをアクセストークンに交換：
```bash
curl -X POST http://localhost:8080/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&code=<認可コード>&client_id=oauth2_demo_client&client_secret=demo_client_secret_12345&redirect_uri=http://localhost:3000/callback&code_verifier=<code_verifier>"
```

#### SPAクライアント（PKCE必須）
```
http://localhost:8080/authorize?client_id=spa_client_example&redirect_uri=http://localhost:8080/callback&response_type=code&scope=read%20profile&code_challenge=E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM&code_challenge_method=S256
```

#### PKCEなしの場合（非推奨）
```
http://localhost:8080/authorize?client_id=oauth2_demo_client&redirect_uri=http://localhost:3000/callback&response_type=code&scope=read
```

トークン交換時はcode_verifierパラメータなしでOK。
