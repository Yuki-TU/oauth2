package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// userInfoHandler は OpenID Connect UserInfo Endpoint（GET /userinfo）。
// Authorization: Bearer <access_token> の JWT を検証し、ユーザー属性を返す。
func userInfoHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	logger := slog.Default()

	authHeader := r.Header.Get("Authorization")
	const p = "Bearer "
	if authHeader == "" || len(authHeader) < len(p) || !strings.EqualFold(authHeader[:len(p)], p) {
		w.Header().Set("WWW-Authenticate", `Bearer realm="userinfo"`)
		http.Error(w, "Bearer token required", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimSpace(authHeader[len(p):])
	if tokenString == "" {
		w.Header().Set("WWW-Authenticate", `Bearer error="invalid_request"`)
		http.Error(w, "Bearer token required", http.StatusUnauthorized)
		return
	}

	claims, err := validateJWTToken(tokenString)
	if err != nil {
		w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// sub は user_id を文字列化したもの（jwt_utils.go を参照）
	var (
		userID   int
		username = claims.Username
		email    string
	)
	if claims.Subject != "" {
		if n, convErr := strconv.Atoi(claims.Subject); convErr == nil {
			userID = n
		}
	}
	if userID != 0 {
		u, getErr := repository.GetUserByID(ctx, userID)
		if getErr != nil {
			logger.Warn("userinfo: ユーザー取得に失敗", "user_id", userID, "error", getErr.Error())
		} else {
			if u.Username != "" {
				username = u.Username
			}
			email = u.Email
		}
	}

	resp := map[string]any{
		"sub":      claims.Subject,
		"username": username,
	}
	if email != "" {
		resp["email"] = email
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
