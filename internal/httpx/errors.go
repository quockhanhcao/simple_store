package httpx

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/quockhanhcao/simple_store/internal/database"
)

type badRequestError struct {
	message string
}

func (e *badRequestError) Error() string {
	return e.message
}

func BadRequest(msg string) error {
	return &badRequestError{message: msg}
}

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
	if badReq, ok := errors.AsType[*badRequestError](err); ok {
		WriteError(w, http.StatusBadRequest, "bad_request", badReq.message)
		return
	}
	if errors.Is(err, database.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "not_found", "The requested resource not found")
		return
	}
	WriteError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred.")
}
