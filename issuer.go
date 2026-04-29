package main

import (
	"os"
	"strings"
)

// issuerBaseURL は OpenID Provider の issuer（オリジン）を返す。
// OIDC の issuer は URL であることが期待されるため、JWT の iss と Discovery の issuer を揃える。
//
// 例: http://localhost:8080
func issuerBaseURL() string {
	if v := strings.TrimSpace(os.Getenv("OAUTH_ISSUER")); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "http://localhost:8080"
}
