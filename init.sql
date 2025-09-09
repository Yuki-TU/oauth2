-- OAuth2 データベース初期化スクリプト

-- クライアントテーブル
CREATE TABLE IF NOT EXISTS oauth_clients (
    id SERIAL PRIMARY KEY,
    client_id VARCHAR(255) UNIQUE NOT NULL,
    client_secret VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    redirect_uris TEXT[],
    scopes TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ユーザーテーブル
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 認可コードテーブル
CREATE TABLE IF NOT EXISTS authorization_codes (
    id SERIAL PRIMARY KEY,
    code VARCHAR(255) UNIQUE NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    user_id INTEGER NOT NULL,
    redirect_uri VARCHAR(255) NOT NULL,
    scopes TEXT[],
    -- PKCE (Proof Key for Code Exchange) サポート
    code_challenge VARCHAR(255),           -- Base64URL-encoded SHA256 hash
    code_challenge_method VARCHAR(10),     -- 'S256' or 'plain'
    -- OpenID Connect サポート
    nonce VARCHAR(255),                    -- OIDC nonce parameter
    state VARCHAR(255),                    -- OAuth2 state parameter
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (client_id) REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- アクセストークンテーブル
CREATE TABLE IF NOT EXISTS access_tokens (
    id SERIAL PRIMARY KEY,
    token VARCHAR(255) UNIQUE NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    user_id INTEGER,
    scopes TEXT[],
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (client_id) REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- リフレッシュトークンテーブル
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    token VARCHAR(255) UNIQUE NOT NULL,
    access_token_id INTEGER NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (access_token_id) REFERENCES access_tokens(id) ON DELETE CASCADE
);

-- セッションテーブル
CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id INTEGER NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- セッションテーブルのインデックス
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- サンプルデータの挿入

-- テストユーザーの挿入（パスワード: password123）
INSERT INTO users (username, password_hash, email) VALUES 
('testuser', '$2a$10$ei3h8UKPrJxoLzIRBUeE/uk.tLTqJXDlprnrnt./WceLuKxVZs7Yq', 'test@example.com'),
('demo', '$2a$10$ei3h8UKPrJxoLzIRBUeE/uk.tLTqJXDlprnrnt./WceLuKxVZs7Yq', 'demo@example.com'),
('admin', '$2a$10$ei3h8UKPrJxoLzIRBUeE/uk.tLTqJXDlprnrnt./WceLuKxVZs7Yq', 'admin@example.com')
ON CONFLICT (username) DO NOTHING;

-- テストクライアントの挿入
INSERT INTO oauth_clients (client_id, client_secret, name, redirect_uris, scopes) VALUES 
-- Webアプリケーション用クライアント
('oauth2_demo_client', 'demo_client_secret_12345', 'OAuth2 Demo Application', 
 '{"http://localhost:3000/callback", "http://localhost:3000/auth/callback", "https://oauthdebugger.com/debug"}',
 '{"read", "write", "openid", "profile", "email"}'),

-- SPAアプリケーション用クライアント（PKCE必須）
('spa_client_example', 'spa_secret_abcdef67890', 'Single Page Application', 
 '{"http://localhost:8080/callback", "http://127.0.0.1:8080/callback"}',
 '{"read", "profile"}'),

-- モバイルアプリ用クライアント
('mobile_app_client', 'mobile_secret_xyz789012', 'Mobile Application', 
 '{"com.example.oauth://callback", "https://app.example.com/auth/callback"}',
 '{"read", "write", "push_notifications"}'),

-- 管理者用クライアント
('admin_console', 'admin_secret_super_secure_456', 'Admin Console', 
 '{"http://localhost:8081/admin/callback"}',
 '{"read", "write", "admin", "user_management"}')
ON CONFLICT (client_id) DO NOTHING;
