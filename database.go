package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// Database は PostgreSQL データベース接続を管理する構造体
type Database struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewDatabase は新しいデータベース接続を作成します
func NewDatabase() (*Database, error) {
	logger := slog.Default()

	// 環境変数からデータベース接続情報を取得
	host := getEnvWithDefault("DB_HOST", "localhost")
	port := getEnvWithDefault("DB_PORT", "5432")
	user := getEnvWithDefault("DB_USER", "oauth2_user")
	password := getEnvWithDefault("DB_PASSWORD", "oauth2_password")
	dbname := getEnvWithDefault("DB_NAME", "oauth2_db")
	sslmode := getEnvWithDefault("DB_SSLMODE", "disable")

	// PostgreSQL接続文字列を構築
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// データベースに接続
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("データベース接続の初期化に失敗しました: %w", err)
	}

	// 接続プールの設定
	db.SetMaxOpenConns(25)                 // 最大接続数
	db.SetMaxIdleConns(5)                  // アイドル接続数
	db.SetConnMaxLifetime(5 * time.Minute) // 接続の最大生存時間

	// 接続テスト
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("データベースへの接続に失敗しました: %w", err)
	}

	logger.Info("データベースに正常に接続されました",
		"host", host,
		"port", port,
		"database", dbname)

	return &Database{
		db:     db,
		logger: logger,
	}, nil
}

// Close はデータベース接続を閉じます
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// Health はデータベースの健全性をチェックします
func (d *Database) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return d.db.PingContext(ctx)
}

// getEnvWithDefault は環境変数を取得し、存在しない場合はデフォルト値を返します
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
