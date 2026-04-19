-- Next デモクライアント (localhost:3000) 用。
-- 既存 DB で init.sql が一度だけ流れ、古い redirect_uris のまま残っている場合に実行する。

UPDATE oauth_clients
SET
  redirect_uris = ARRAY[
    'http://localhost:3000/callback',
    'http://localhost:3000/auth/callback',
    'https://oauthdebugger.com/debug'
  ]::text[],
  updated_at = CURRENT_TIMESTAMP
WHERE client_id = 'oauth2_demo_client';
