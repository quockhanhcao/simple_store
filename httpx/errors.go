package httpx

import (
	"log/slog"
	"net/http"
)

type errorResponse struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// WriteError send error envelope with a status you choose
func WriteError(w http.ResponseWriter, status int, code, message string) {
	_ = WriteJSON(w, status, errorResponse{Error: errorDetail{Code: code, Message: message}})
}

func Error(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	logger.Error("request failed", "method", r.Method, "path", r.URL.Path, "err", err)
	WriteError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred.")
}
