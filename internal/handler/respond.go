package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode response", "err", err)
	}
}

func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"error": msg})
}

func Errors(w http.ResponseWriter, status int, msgs []string) {
	JSON(w, status, map[string][]string{"errors": msgs})
}

func NotFound(w http.ResponseWriter) {
	Error(w, http.StatusNotFound, "Not found")
}

func InternalError(w http.ResponseWriter, err error) {
	slog.Error("internal error", "err", err)
	Error(w, http.StatusInternalServerError, "Internal server error")
}

func ValidateRequired(w http.ResponseWriter, field, value string) bool {
	if strings.TrimSpace(value) == "" {
		Errors(w, http.StatusBadRequest, []string{field + ": required"})
		return false
	}
	return true
}
