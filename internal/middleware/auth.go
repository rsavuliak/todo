package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rsavuliak/todo/internal/config"
)

type contextKey int

const userIDKey contextKey = iota

func Auth(cfg config.Config) func(http.Handler) http.Handler {
	// JWT_SECRET uses standard Base64 (base64.StdEncoding), not URL-safe Base64.
	// This must match the encoding the auth service uses when it creates the key.
	keyBytes, err := base64.StdEncoding.DecodeString(cfg.JWTSecret)
	if err != nil {
		// Config validation ensures this never fires in production.
		panic(fmt.Sprintf("middleware: invalid JWT secret: %v", err))
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cfg.CookieName)
			if err != nil {
				unauthorized(w)
				return
			}

			token, err := jwt.ParseWithClaims(
				cookie.Value,
				&jwt.RegisteredClaims{},
				func(t *jwt.Token) (interface{}, error) {
					if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
					}
					return keyBytes, nil
				},
			)
			if err != nil || !token.Valid {
				unauthorized(w)
				return
			}

			claims, ok := token.Claims.(*jwt.RegisteredClaims)
			if !ok || claims.Subject == "" {
				unauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok && id != ""
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
}
