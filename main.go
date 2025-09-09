// main.go
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var startedAt = time.Now()

// グローバル変数
var (
	db         *Database
	repository *Repository
)

func main() {
	logger := slog.Default()

	// データベース接続を初期化
	var err error
	db, err = NewDatabase()
	if err != nil {
		logger.Error("データベース接続の初期化に失敗しました", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// リポジトリを初期化
	repository = NewRepository(db)

	// クリーンアップタスクを定期実行
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := repository.CleanupExpiredTokens(ctx); err != nil {
				logger.Error("期限切れトークンのクリーンアップに失敗しました", "error", err)
			}
			cancel()
		}
	}()

	mux := http.NewServeMux()

	// ホームページ
	mux.HandleFunc("GET /", homeHandler)

	// ヘルスチェックとデバッグ
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("GET /debug", debugRequestHandler)
	mux.HandleFunc("GET /pkce", pkceDemo)

	// 認証エンドポイント
	mux.HandleFunc("GET /login", loginGetHandler)
	mux.HandleFunc("POST /login", loginPostHandler)
	mux.HandleFunc("GET /signup", signupGetHandler)
	mux.HandleFunc("POST /signup", signupPostHandler)
	mux.HandleFunc("GET /logout", logoutHandler)
	mux.HandleFunc("POST /logout", logoutHandler)

	// OAuth2エンドポイント
	mux.HandleFunc("GET /authorize", authorizeHandler)
	mux.HandleFunc("POST /token", tokenHandler)
	mux.HandleFunc("GET /callback", callbackHandler)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("http server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server error", "err", err)
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
		return
	}
	logger.Info("server stopped gracefully")
}
