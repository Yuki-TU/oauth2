# OAuth2 デモ（認可サーバー + Next クライアント + リソースサーバー）

PostgreSQL を使った **OAuth2 / OIDC 風の認可サーバー**（Go）、**Next.js のデモクライアント**、**JWT を JWKS で検証するリソースサーバー**（別モジュールの Go）を同じリポジトリで動かせるサンプルです。

| コンポーネント                 | 役割                                               | 既定 URL                | 詳細                                   |
| ------------------------------ | -------------------------------------------------- | ----------------------- | -------------------------------------- |
| 認可サーバー（ルートの Go）    | 認可コード（PKCE）、トークン、セッション、JWT 発行 | `http://localhost:8080` | 本文・`init.sql`                       |
| PostgreSQL                     | クライアント・ユーザー・コード・トークン等の永続化 | `localhost:5432`        | `compose.yaml`                         |
| デモクライアント（Next.js）    | PKCE 付きログイン、HttpOnly クッキー、リフレッシュ | `http://localhost:3000` | [client/README.md](client/README.md)   |
| リソースサーバー（`backend/`） | アクセストークン（JWT）の署名検証と保護 API        | `http://localhost:9090` | [backend/README.md](backend/README.md) |

リソースサーバーは認可サーバーの **`/jwks`** から公開鍵を取得し、発行されたアクセストークンを検証します。Next は HttpOnly のトークンをブラウザ JS に渡さず、[client/README.md](client/README.md) に書いたとおり **`/api/resource/me` でプロキシ**してリソースサーバーを呼びます。

## 前提条件

- **Go**（ルートと `backend/` は別モジュール）
- **Node.js 20+**（Next クライアント）
- **Docker**（PostgreSQL 用）
- 認可サーバー用の **RSA 鍵**（初回のみ `make create-key`。`certificate/` は `.gitignore` 対象）

## クイックスタート

1. **鍵を生成**（未作成のときのみ）

   ```bash
   make create-key
   ```

2. **DB を起動**（初回は `init.sql` が自動実行されます）

   ```bash
   make up
   ```

3. **ルートで依存取得**（認可サーバー）

   ```bash
   make deps
   ```

4. **認可サーバー用 `.env` を用意**  
   `make run` は `env $(cat .env | xargs) go run *.go` のため、リポジトリ直下に `.env` が必要です。未設定でも `database.go` の既定値（`localhost:5432` の `oauth2_db` 等）で動きますが、空ファイルではなく **少なくとも1行あるファイル**にしてください。例:

   ```bash
   # .env（例）
   DB_HOST=localhost
   DB_PORT=5432
   DB_NAME=oauth2_db
   DB_USER=oauth2_user
   DB_PASSWORD=oauth2_password
   DB_SSLMODE=disable
   ```

5. **認可サーバーを起動**

   ```bash
   make run
   ```

6. **（任意）Next クライアント** — 別ターミナルで [client/README.md](client/README.md) の手順（`client/env.example` を `.env.local` にコピーなど）。

7. **（任意）リソースサーバー** — 認可サーバー起動後、別ターミナルで `make backend-run`（[backend/README.md](backend/README.md)）。

## ディレクトリ構成（抜粋）

```
.
├── main.go, authorize.go, token.go, …   # 認可サーバー（:8080）
├── init.sql                            # DB スキーマ + シード（Docker 初回マウント）
├── compose.yaml
├── Makefile
├── client/                             # Next.js（:3000）
└── backend/                            # リソースサーバー（:9090、独立 go.mod）
```

## よく使う Make ターゲット

| ターゲット                                | 説明                                               |
| ----------------------------------------- | -------------------------------------------------- |
| `make help`                               | 一覧表示                                           |
| `make up` / `make down`                   | PostgreSQL の起動・停止                            |
| `make run`                                | 認可サーバー（要 `.env`）                          |
| `make build`                              | 認可サーバーを `oauth2-server` にビルド            |
| `make test`                               | ルートモジュールの `go test`                       |
| `make db`                                 | コンテナ内 `psql` 対話シェル                       |
| `make db-sync-demo-redirects`             | 起動済み DB に `init.sql` を再適用（開発用・冪等） |
| `make create-key`                         | JWT 用 RSA 鍵を `certificate/` に生成              |
| `make client-install` / `make client-dev` | Next の依存導入・開発サーバー                      |
| `make backend-run`                        | リソースサーバー                                   |

## 認可サーバー（ルート）の概要

### 主な機能

- OAuth2 **Authorization Code**（**PKCE** 対応）
- セッション、リフレッシュトークン（DB 永続化）
- **JWT** アクセストークン（RS256）と **JWKS**（`/jwks` 等）
- **OpenID Connect Discovery**: `GET /.well-known/openid_configuration`（メタデータ）、`GET /.well-known/jwks.json`（JWKS）
- 期限切れトークンの定期クリーンアップ

### 主な HTTP エンドポイント（抜粋）

| メソッド・パス                            | 説明                                                  |
| ----------------------------------------- | ----------------------------------------------------- |
| `GET /`                                   | ホーム                                                |
| `GET /healthz`                            | ヘルスチェック                                        |
| `GET /login`, `POST /login`               | ログイン                                              |
| `GET /signup`, `POST /signup`             | 登録                                                  |
| `GET /authorize`                          | 認可エンドポイント                                    |
| `POST /token`                             | トークン（`authorization_code` / `refresh_token` 等） |
| `GET /jwks`, `GET /.well-known/jwks.json` | 公開鍵セット                                          |
| `GET /pkce`                               | PKCE デモ用 UI                                        |

### データベース環境変数

| 変数          | 既定              |
| ------------- | ----------------- |
| `DB_HOST`     | `localhost`       |
| `DB_PORT`     | `5432`            |
| `DB_NAME`     | `oauth2_db`       |
| `DB_USER`     | `oauth2_user`     |
| `DB_PASSWORD` | `oauth2_password` |
| `DB_SSLMODE`  | `disable`         |

## テストデータ（`init.sql`）

### ユーザー（パスワードはいずれも `password123`）

- `testuser` / `test@example.com`
- `demo` / `demo@example.com`
- `admin` / `admin@example.com`

### クライアント（抜粋）

- **Next デモ用**: `oauth2_demo_client` / `demo_client_secret_12345`  
  Redirect URI の例: `http://localhost:3000/callback`（`init.sql` の配列と `client/.env.local` を一致させること）
- **SPA 例**: `spa_client_example`（PKCE 前提の URI）
- その他: `mobile_app_client`, `admin_console`

`oauth_clients` は `INSERT ... ON CONFLICT DO UPDATE` により、シードを流し直すと **redirect_uris 等も更新**されます。

## 手動での認可 URL例（PKCE あり）

ブラウザで開き、ログイン後に `redirect_uri` へコードが付きます。

```
http://localhost:8080/authorize?client_id=oauth2_demo_client&redirect_uri=http://localhost:3000/callback&response_type=code&scope=read%20write%20openid&state=xyz123&nonce=n-0S6_WzA2Mj&code_challenge=E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM&code_challenge_method=S256
```

トークン交換の例（`code` と `code_verifier` を差し替え）:

```bash
curl -sS -X POST http://localhost:8080/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&code=認可コード&client_id=oauth2_demo_client&client_secret=demo_client_secret_12345&redirect_uri=http://localhost:3000/callback&code_verifier=code_verifier の値"
```

## トラブルシュート

- **`Invalid redirect_uri`**  
  `init.sql` の `oauth2_demo_client` の `redirect_uris` と、クライアントが送る `redirect_uri` を完全一致させる。既存ボリュームなら `make db-sync-demo-redirects` で `init.sql` を再適用するか、DB ボリュームを削除して `make up` し直す。
- **リソースサーバーが JWKS を取れない**  
  先に認可サーバー（:8080）を起動する。`RESOURCE_JWKS_URI` を [backend/README.md](backend/README.md) で確認。
- **`make run` が `.env` で失敗する**  
  リポジトリ直下に `.env` を置き、上記の例のように変数を定義する。

## 関連ドキュメント

- [client/README.md](client/README.md) — Next クライアントの環境変数・API ルート・デモ UI
- [backend/README.md](backend/README.md) — リソースサーバーの環境変数とエンドポイント
