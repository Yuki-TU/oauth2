package main

import (
	"encoding/json"
	"net/http"
	"time"
)

func healthz(w http.ResponseWriter, r *http.Request) {
	noCacheJSON(w)

	// データベース接続をチェック
	status := "ok"
	dbStatus := "ok"
	if db != nil {
		if err := db.Health(); err != nil {
			status = "degraded"
			dbStatus = "error"
		}
	} else {
		status = "degraded"
		dbStatus = "not_initialized"
	}

	resp := struct {
		Status    string        `json:"status"`
		Database  string        `json:"database"`
		Uptime    time.Duration `json:"uptime"`
		CheckedAt time.Time     `json:"checked_at"`
	}{
		Status:    status,
		Database:  dbStatus,
		Uptime:    time.Since(startedAt),
		CheckedAt: time.Now(),
	}

	if status != "ok" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func noCacheJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
}
