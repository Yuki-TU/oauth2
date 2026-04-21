// リソースサーバー（別プロセス・既定 :9090）。
//
// 認可サーバー（:8080）が発行した JWT アクセストークンを、認可サーバーの JWKS から
// 取得した RSA 公開鍵で検証し、保護 API を返す。DB やトークンイントロスペクションには依存しない。
//
// 環境変数（省略時は括弧内が既定）:
//   RESOURCE_JWKS_URI          … JWKS の URL（http://localhost:8080/jwks）
//   RESOURCE_LISTEN_ADDR       … 待ち受けアドレス（:9090）
//   RESOURCE_EXPECTED_ISS      … JWT の iss と一致させる値（oauth2-server）
//   RESOURCE_ALLOWED_AUDIENCES … 空なら aud 検証なし。指定時はカンマ区切りでいずれかと一致必須
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

func main() {
	logger := slog.Default()

	// 検証に使う公開鍵の出所（認可サーバーと同一ホスト想定）
	jwksURI := os.Getenv("RESOURCE_JWKS_URI")
	if jwksURI == "" {
		jwksURI = "http://localhost:8080/jwks"
	}
	cache := newJWKSCache(jwksURI, 5*time.Minute)

	// 起動時に 1 回必ず JWKS を取れないと、鍵なしで動けないため失敗時は終了
	initCtx, initCancel := context.WithTimeout(context.Background(), 10*time.Second)
	err := cache.refresh(initCtx)
	initCancel()
	if err != nil {
		logger.Error("起動時の JWKS 取得に失敗しました（認可サーバーが起動しているか確認）", "uri", jwksURI, "err", err)
		os.Exit(1)
	}
	logger.Info("JWKS を取得しました", "uri", jwksURI)

	addr := os.Getenv("RESOURCE_LISTEN_ADDR")
	if addr == "" {
		addr = ":9090"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	// 保護 API の例: Bearer の JWT を JWKS で検証し、クレームをそのまま JSON で返す
	mux.HandleFunc("GET /api/me", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		raw, ok := bearerToken(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Authorization: Bearer が必要です",
			})
			return
		}
		claims, err := parseAndValidateAccessToken(ctx, cache, raw)
		if err != nil {
			// 詳細はクライアントに返さずログのみ（トークン内容の推測を避ける）
			logger.Debug("JWT 検証失敗", "err", err.Error())
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "無効なアクセストークン",
			})
			return
		}

		resp := map[string]any{
			"sub":       claims.Subject,
			"username":  claims.Username,
			"client_id": claims.ClientID,
			"scope":     claims.Scope,
			"iss":       claims.Issuer,
			"aud":       claims.Audience,
		}
		if claims.ExpiresAt != nil {
			resp["exp"] = claims.ExpiresAt.Unix()
		}
		writeJSON(w, http.StatusOK, resp)
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           accessLog(logger, mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctxSig, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("resource server starting", "addr", addr, "jwks", jwksURI)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctxSig.Done()
	logger.Info("shutdown signal received")

	shCtx, c2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer c2()
	_ = srv.Shutdown(shCtx)
	logger.Info("resource server stopped")
}
