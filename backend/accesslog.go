package main

import (
	"log/slog"
	"net/http"
	"time"
)

// accessResponseWriter は WriteHeader 前に status を記録する。
type accessResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *accessResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *accessResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

// accessLog はリクエストごとに method / path / status / 所要時間 / クライアント情報を INFO で出す。
func accessLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		aw := &accessResponseWriter{ResponseWriter: w, status: 0}
		next.ServeHTTP(aw, r)
		status := aw.status
		if status == 0 {
			status = http.StatusOK
		}
		logger.Info("access",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"status", status,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}
