package middleware_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rsavuliak/todo/internal/config"
	"github.com/rsavuliak/todo/internal/middleware"
)

var testSecret = base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32)))

func cfg() config.Config {
	return config.Config{
		JWTSecret:  testSecret,
		CookieName: "token",
	}
}

func mintToken(subject string, secret []byte, exp time.Time) string {
	claims := jwt.RegisteredClaims{
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(exp),
	}
	t, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	return t
}

func secretBytes() []byte {
	b, _ := base64.StdEncoding.DecodeString(testSecret)
	return b
}

func TestAuth_NoCookie(t *testing.T) {
	mw := middleware.Auth(cfg())
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	mw := middleware.Auth(cfg())
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })

	token := mintToken("user-1", secretBytes(), time.Now().Add(-time.Hour))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	mw(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_WrongSignature(t *testing.T) {
	mw := middleware.Auth(cfg())
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })

	wrongKey := []byte(strings.Repeat("w", 32))
	token := mintToken("user-1", wrongKey, time.Now().Add(time.Hour))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	mw(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_ValidToken_UserIDInContext(t *testing.T) {
	mw := middleware.Auth(cfg())

	var gotUserID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			t.Error("expected userID in context")
		}
		gotUserID = id
		w.WriteHeader(200)
	})

	token := mintToken("user-abc", secretBytes(), time.Now().Add(time.Hour))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	mw(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if gotUserID != "user-abc" {
		t.Errorf("expected userID 'user-abc', got %q", gotUserID)
	}
}

func TestAuth_OPTIONS_PassThrough(t *testing.T) {
	mw := middleware.Auth(cfg())
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	mw(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("OPTIONS should pass through, got %d", rec.Code)
	}
}
