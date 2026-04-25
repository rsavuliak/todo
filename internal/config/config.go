package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port        string
	DatabaseURL string
	// JWTSecret is a standard Base64-encoded HMAC key (not URL-safe Base64).
	// Must match the encoding used by the auth service when it signs tokens.
	JWTSecret   string
	CookieName  string
	CORSOrigins []string
}

func Load() (Config, error) {
	cfg := Config{
		Port:       getEnv("PORT", "8080"),
		CookieName: getEnv("COOKIE_NAME", "token"),
	}

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	if err := validateJWTSecret(cfg.JWTSecret); err != nil {
		return Config{}, err
	}

	originsRaw := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")
	for _, o := range strings.Split(originsRaw, ",") {
		if trimmed := strings.TrimSpace(o); trimmed != "" {
			cfg.CORSOrigins = append(cfg.CORSOrigins, trimmed)
		}
	}

	return cfg, nil
}

func validateJWTSecret(secret string) error {
	decoded, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return fmt.Errorf("JWT_SECRET must be valid standard Base64: %w", err)
	}
	if len(decoded) < 32 {
		return fmt.Errorf("JWT_SECRET must decode to at least 32 bytes (got %d)", len(decoded))
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
