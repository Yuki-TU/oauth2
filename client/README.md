# OAuth2 デモクライアント（Next.js）

認可サーバー（既定 `http://localhost:8080`）に対して **Authorization Code + PKCE** でログインし、トークンを **HttpOnly クッキー**に保存する最小デモです。ポートは **3000** 固定（`package.json` の `next dev -p 3000`）。

## 前提

- 認可サーバーと PostgreSQLが起動していること（リポジトリルートの `README.md` 参照）
- `init.sql` の `oauth2_demo_client` の `redirect_uris` に **`http://localhost:3000/callback`** が含まれていること

## セットアップ

```bash
cd client
cp env.example .env.local
npm install
npm run dev
```

ブラウザで `http://localhost:3000` を開きます。

## 環境変数（`.env.local`）

| 変数                  | 説明                                                                                                       |
| --------------------- | ---------------------------------------------------------------------------------------------------------- |
| `OAUTH_ISSUER`        | 認可サーバーのオリジン（例: `http://localhost:8080`）                                                      |
| `OAUTH_CLIENT_ID`     | 登録済みクライアント ID（デモは `oauth2_demo_client`）                                                     |
| `OAUTH_CLIENT_SECRET` | クライアントシークレット                                                                                   |
| `OAUTH_REDIRECT_URI`  | このアプリのコールバック URL。**DB の `redirect_uris` と完全一致**（例: `http://localhost:3000/callback`） |
| `RESOURCE_SERVER_URL` | リソースサーバーのベース URL（例: `http://localhost:9090`）。未使用なら既定でも可                          |

テンプレートは [env.example](env.example) です。

## 動作の要点

1. **`GET /api/oauth/start`** — PKCE 用の verifier/challenge と state をクッキーに保存し、認可サーバーの `/authorize` へリダイレクトします。
2. **`GET /callback`** — 認可コードを受け取り、サーバー側で `POST {issuer}/token` し、**`demo_access_token` / `demo_refresh_token` / `demo_id_token`**（HttpOnly）をセットして `/` へ戻します。
3. **`POST /api/oauth/refresh`** — リフレッシュトークンでアクセストークンを更新し、クッキーを書き換えます。
4. **`POST /api/oauth/logout`** — デモ用クッキーを削除します。
5. **`GET /api/resource/me`** — クッキーのアクセストークンを読み取り、**`RESOURCE_SERVER_URL`** の `GET /api/me` に `Authorization: Bearer` を付けてサーバー間プロキシします。  
   **HttpOnly のためブラウザ JS はトークン文字列を読めない**ため、別オリジンのリソースサーバーを直接 Bearer で叩く代わりにこのルートを使います。

ホームでは JWT ペイロードの表示（検証なし）と、リソースサーバー呼び出し用のデモ UI（`app/components/ResourceMePanel.tsx`）があります。期限が近いときは `TokenRefresher` がリフレッシュを試みます。

## npm スクリプト

| コマンド                          | 説明                        |
| --------------------------------- | --------------------------- |
| `npm run dev`                     | 開発サーバー（`:3000`）     |
| `npm run build` / `npm run start` | 本番ビルドと起動（`:3000`） |

リポジトリルートからは `make client-install` と `make client-dev` でも同じことができます。

## 関連

- ルートの概要: [../README.md](../README.md)
- リソースサーバー: [../backend/README.md](../backend/README.md)
