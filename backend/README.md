# リソースサーバー（JWT 検証 API）

認可サーバーが発行した **JWT アクセストークン**を、認可サーバーの **JWKS**（RSA 公開鍵）で検証し、保護された JSON API を返す **別プロセス**の Go アプリです。独自の `go.mod` を持ち、**PostgreSQL やイントロスペクションには依存しません**。

## 前提

- **認可サーバー**（既定 `http://localhost:8080`）が起動し、`GET /jwks` から鍵セットが取得できること
- 検証する JWT の **`iss`** が `RESOURCE_EXPECTED_ISS`（既定 `oauth2-server`）と一致すること
- アルゴリズムは **RS256**、`kid` ヘッダで JWKS の鍵を引き当てます

## 起動

リポジトリルートから:

```bash
make backend-run
```

または:

```bash
cd backend
go run .
```

既定の待ち受けは **`http://localhost:9090`**（`RESOURCE_LISTEN_ADDR` で変更）。

## 環境変数

| 変数                         | 既定                         | 説明                                                                                    |
| ---------------------------- | ---------------------------- | --------------------------------------------------------------------------------------- |
| `RESOURCE_JWKS_URI`          | `http://localhost:8080/jwks` | JWKS JSON の URL                                                                        |
| `RESOURCE_LISTEN_ADDR`       | `:9090`                      | リッスンアドレス                                                                        |
| `RESOURCE_EXPECTED_ISS`      | `oauth2-server`              | JWT の `iss` 検証値                                                                     |
| `RESOURCE_ALLOWED_AUDIENCES` | （空）                       | 空のときは **`aud` を検証しない**（デモ向け）。指定時はカンマ区切りでいずれかと一致必須 |

起動時に JWKS を一度取得できない場合は **プロセス終了**します（鍵なしでは検証できないため）。

## HTTP エンドポイント

| メソッド・パス | 説明                                                                                                                               |
| -------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| `GET /healthz` | 常に 200 と `ok`（認証不要）                                                                                                       |
| `GET /api/me`  | **`Authorization: Bearer <JWT>`** 必須。検証成功時は `sub`, `username`, `client_id`, `scope`, `iss`, `aud`, `exp` 等を JSON で返す |

## アクセスログ

全リクエストを `slog` の **`access`**（INFO）で記録します。フィールド例: `method`, `path`, `query`, `status`, `duration_ms`, `remote_addr`, `user_agent`。

## Next クライアントから試す

ブラウザは HttpOnly トークンを読めないため、Next 側の **`GET /api/resource/me`** がクッキーを読み、本サーバーの `GET /api/me` へプロキシします。手順と環境変数は [../client/README.md](../client/README.md) を参照してください。

## 関連

- リポジトリ全体: [../README.md](../README.md)
- 認可サーバーの JWKS 実装: ルートの `jwks.go` 等
