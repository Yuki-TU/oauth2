package main

import (
	"encoding/json"
	"net/http"
	"time"
)

func healthz(w http.ResponseWriter, r *http.Request) {
	noCacheJSON(w)
	resp := struct {
		Status    string        `json:"status"`
		Uptime    time.Duration `json:"uptime"`
		CheckedAt time.Time     `json:"checked_at"`
	}{
		Status:    "ok",
		Uptime:    time.Since(startedAt),
		CheckedAt: time.Now(),
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func noCacheJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
}
