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

func main() {
	logger := slog.Default()

	mux := http.NewServeMux()

	// ヘルスチェック
	mux.HandleFunc("GET /healthz", healthz)

	// OAuth2エンドポイント
	mux.HandleFunc("GET /login", loginGetHandler)
	mux.HandleFunc("POST /login", loginPostHandler)
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
