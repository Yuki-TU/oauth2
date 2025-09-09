package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// Repository はデータベース操作を提供する構造体
type Repository struct {
	db *Database
}

// NewRepository は新しいリポジトリインスタンスを作成します
func NewRepository(db *Database) *Repository {
	return &Repository{db: db}
}

// ユーザー関連のメソッド

// CreateUser は新しいユーザーを作成します
func (r *Repository) CreateUser(ctx context.Context, username, password, email string) (*User, error) {
	// パスワードをハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("パスワードのハッシュ化に失敗しました: %w", err)
	}

	query := `
		INSERT INTO users (username, password_hash, email)
		VALUES ($1, $2, $3)
		RETURNING id, username, password_hash, email, created_at, updated_at`

	var user User
	err = r.db.db.QueryRowContext(ctx, query, username, string(hashedPassword), email).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("ユーザーの作成に失敗しました: %w", err)
	}

	return &user, nil
}

// GetUserByUsername はユーザー名でユーザーを取得します
func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT id, username, password_hash, email, created_at, updated_at
		FROM users
		WHERE username = $1`

	var user User
	err := r.db.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ユーザーが見つかりません: %s", username)
		}
		return nil, fmt.Errorf("ユーザーの取得に失敗しました: %w", err)
	}

	return &user, nil
}

// ValidateUserPassword はユーザーのパスワードを検証します
func (r *Repository) ValidateUserPassword(ctx context.Context, username, password string) (*User, error) {
	user, err := r.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("パスワードが正しくありません")
	}

	return user, nil
}

// OAuth2クライアント関連のメソッド

// GetClientByID はクライアントIDでOAuth2クライアントを取得します
func (r *Repository) GetClientByID(ctx context.Context, clientID string) (*OAuthClient, error) {
	query := `
		SELECT id, client_id, client_secret, name, redirect_uris, scopes, created_at, updated_at
		FROM oauth_clients
		WHERE client_id = $1`

	var client OAuthClient
	err := r.db.db.QueryRowContext(ctx, query, clientID).Scan(
		&client.ID, &client.ClientID, &client.ClientSecret, &client.Name,
		&client.RedirectURIs, &client.Scopes, &client.CreatedAt, &client.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("クライアントが見つかりません: %s", clientID)
		}
		return nil, fmt.Errorf("クライアントの取得に失敗しました: %w", err)
	}

	return &client, nil
}

// ValidateClientCredentials はクライアントの認証情報を検証します
func (r *Repository) ValidateClientCredentials(ctx context.Context, clientID, clientSecret string) (*OAuthClient, error) {
	client, err := r.GetClientByID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	if client.ClientSecret != clientSecret {
		return nil, fmt.Errorf("クライアントシークレットが正しくありません")
	}

	return client, nil
}

// 認可コード関連のメソッド

// CreateAuthorizationCode は新しい認可コードを作成します
func (r *Repository) CreateAuthorizationCode(ctx context.Context, code string, clientID string, userID int, redirectURI string, scopes []string, codeChallenge, codeChallengeMethod, nonce, state *string, expiresAt time.Time) error {
	query := `
		INSERT INTO authorization_codes (code, client_id, user_id, redirect_uri, scopes, code_challenge, code_challenge_method, nonce, state, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.db.ExecContext(ctx, query, code, clientID, userID, redirectURI, pq.Array(scopes), codeChallenge, codeChallengeMethod, nonce, state, expiresAt)
	if err != nil {
		return fmt.Errorf("認可コードの作成に失敗しました: %w", err)
	}

	return nil
}

// GetAuthorizationCode は認可コードを取得し、使用後に削除します
func (r *Repository) GetAuthorizationCode(ctx context.Context, code string) (*AuthorizationCode, error) {
	tx, err := r.db.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer tx.Rollback()

	// 認可コードを取得
	query := `
		SELECT id, code, client_id, user_id, redirect_uri, scopes, code_challenge, code_challenge_method, nonce, state, expires_at, created_at
		FROM authorization_codes
		WHERE code = $1`

	var authCode AuthorizationCode
	err = tx.QueryRowContext(ctx, query, code).Scan(
		&authCode.ID, &authCode.Code, &authCode.ClientID, &authCode.UserID, &authCode.RedirectURI,
		&authCode.Scopes, &authCode.CodeChallenge, &authCode.CodeChallengeMethod,
		&authCode.Nonce, &authCode.State, &authCode.ExpiresAt, &authCode.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("認可コードが見つかりません")
		}
		return nil, fmt.Errorf("認可コードの取得に失敗しました: %w", err)
	}

	// 期限切れチェック
	if time.Now().After(authCode.ExpiresAt) {
		// 期限切れの場合は削除
		_, err = tx.ExecContext(ctx, "DELETE FROM authorization_codes WHERE code = $1", code)
		if err != nil {
			return nil, fmt.Errorf("期限切れ認可コードの削除に失敗しました: %w", err)
		}
		return nil, fmt.Errorf("認可コードが期限切れです")
	}

	// 認可コードを削除（使用済みとして）
	_, err = tx.ExecContext(ctx, "DELETE FROM authorization_codes WHERE code = $1", code)
	if err != nil {
		return nil, fmt.Errorf("認可コードの削除に失敗しました: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return &authCode, nil
}

// アクセストークン関連のメソッド

// CreateAccessToken は新しいアクセストークンを作成します
func (r *Repository) CreateAccessToken(ctx context.Context, token, clientID string, userID *int, scopes []string, expiresAt time.Time) (*AccessToken, error) {
	query := `
		INSERT INTO access_tokens (token, client_id, user_id, scopes, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, token, client_id, user_id, scopes, expires_at, created_at`

	var accessToken AccessToken
	err := r.db.db.QueryRowContext(ctx, query, token, clientID, userID, pq.Array(scopes), expiresAt).Scan(
		&accessToken.ID, &accessToken.Token, &accessToken.ClientID, &accessToken.UserID,
		&accessToken.Scopes, &accessToken.ExpiresAt, &accessToken.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("アクセストークンの作成に失敗しました: %w", err)
	}

	return &accessToken, nil
}

// GetAccessTokenByToken はトークン文字列でアクセストークンを取得します
func (r *Repository) GetAccessTokenByToken(ctx context.Context, token string) (*AccessToken, error) {
	query := `
		SELECT id, token, client_id, user_id, scopes, expires_at, created_at
		FROM access_tokens
		WHERE token = $1`

	var accessToken AccessToken
	err := r.db.db.QueryRowContext(ctx, query, token).Scan(
		&accessToken.ID, &accessToken.Token, &accessToken.ClientID, &accessToken.UserID,
		&accessToken.Scopes, &accessToken.ExpiresAt, &accessToken.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("アクセストークンが見つかりません")
		}
		return nil, fmt.Errorf("アクセストークンの取得に失敗しました: %w", err)
	}

	// 期限切れチェック
	if time.Now().After(accessToken.ExpiresAt) {
		return nil, fmt.Errorf("アクセストークンが期限切れです")
	}

	return &accessToken, nil
}

// RevokeAccessToken はアクセストークンを無効化します
func (r *Repository) RevokeAccessToken(ctx context.Context, token string) error {
	_, err := r.db.db.ExecContext(ctx, "DELETE FROM access_tokens WHERE token = $1", token)
	if err != nil {
		return fmt.Errorf("アクセストークンの無効化に失敗しました: %w", err)
	}

	return nil
}

// セッション管理メソッド

// CreateSession は新しいセッションを作成します
func (r *Repository) CreateSession(ctx context.Context, sessionID string, userID int, expiresAt time.Time) error {
	query := `
		INSERT INTO sessions (id, user_id, expires_at)
		VALUES ($1, $2, $3)`

	_, err := r.db.db.ExecContext(ctx, query, sessionID, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("セッションの作成に失敗しました: %w", err)
	}

	return nil
}

// GetSession はセッションIDでセッション情報を取得します
func (r *Repository) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	query := `
		SELECT id, user_id, created_at, expires_at
		FROM sessions
		WHERE id = $1`

	var session Session
	err := r.db.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID, &session.UserID, &session.CreatedAt, &session.ExpiresAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("セッションが見つかりません")
		}
		return nil, fmt.Errorf("セッションの取得に失敗しました: %w", err)
	}

	// 期限切れチェック
	if time.Now().After(session.ExpiresAt) {
		// 期限切れセッションを削除
		r.DeleteSession(ctx, sessionID)
		return nil, fmt.Errorf("セッションが期限切れです")
	}

	return &session, nil
}

// DeleteSession はセッションを削除します
func (r *Repository) DeleteSession(ctx context.Context, sessionID string) error {
	_, err := r.db.db.ExecContext(ctx, "DELETE FROM sessions WHERE id = $1", sessionID)
	if err != nil {
		return fmt.Errorf("セッションの削除に失敗しました: %w", err)
	}

	return nil
}

// UpdateSessionExpiry はセッションの有効期限を更新します
func (r *Repository) UpdateSessionExpiry(ctx context.Context, sessionID string, expiresAt time.Time) error {
	_, err := r.db.db.ExecContext(ctx, "UPDATE sessions SET expires_at = $1 WHERE id = $2", expiresAt, sessionID)
	if err != nil {
		return fmt.Errorf("セッション有効期限の更新に失敗しました: %w", err)
	}

	return nil
}

// GetUserSessions は特定ユーザーのすべてのアクティブセッションを取得します
func (r *Repository) GetUserSessions(ctx context.Context, userID int) ([]*Session, error) {
	query := `
		SELECT id, user_id, created_at, expires_at
		FROM sessions
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY created_at DESC`

	rows, err := r.db.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("ユーザーセッションの取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var session Session
		err := rows.Scan(&session.ID, &session.UserID, &session.CreatedAt, &session.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("セッション情報の読み取りに失敗しました: %w", err)
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// DeleteUserSessions は特定ユーザーのすべてのセッションを削除します（強制ログアウト）
func (r *Repository) DeleteUserSessions(ctx context.Context, userID int) error {
	_, err := r.db.db.ExecContext(ctx, "DELETE FROM sessions WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("ユーザーセッションの削除に失敗しました: %w", err)
	}

	return nil
}

// クリーンアップメソッド

// CleanupExpiredTokens は期限切れのトークンとコードを削除します
func (r *Repository) CleanupExpiredTokens(ctx context.Context) error {
	now := time.Now()

	// 期限切れの認可コードを削除
	_, err := r.db.db.ExecContext(ctx, "DELETE FROM authorization_codes WHERE expires_at < $1", now)
	if err != nil {
		return fmt.Errorf("期限切れ認可コードの削除に失敗しました: %w", err)
	}

	// 期限切れのアクセストークンを削除
	_, err = r.db.db.ExecContext(ctx, "DELETE FROM access_tokens WHERE expires_at < $1", now)
	if err != nil {
		return fmt.Errorf("期限切れアクセストークンの削除に失敗しました: %w", err)
	}

	// 期限切れのリフレッシュトークンを削除
	_, err = r.db.db.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE expires_at < $1", now)
	if err != nil {
		return fmt.Errorf("期限切れリフレッシュトークンの削除に失敗しました: %w", err)
	}

	// 期限切れのセッションを削除
	_, err = r.db.db.ExecContext(ctx, "DELETE FROM sessions WHERE expires_at < $1", now)
	if err != nil {
		return fmt.Errorf("期限切れセッションの削除に失敗しました: %w", err)
	}

	return nil
}
