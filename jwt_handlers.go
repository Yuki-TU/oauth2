package main

import (
	"encoding/json"
	"net/http"
	"time"
)

func testJWTHandler(w http.ResponseWriter, r *http.Request) {
	token, err := generateJWTAccessToken(1, "testuser", "test-client", "openid profile", time.Hour)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"access_token": token})
}

func verifyJWTHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if body.Token == "" {
		http.Error(w, "token required", http.StatusBadRequest)
		return
	}
	claims, err := validateJWTToken(body.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(claims)
}
