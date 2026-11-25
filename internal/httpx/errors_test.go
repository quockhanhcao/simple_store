package httpx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/quockhanhcao/simple_store/internal/database"
)

func TestWriteError(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	WriteError(rec, http.StatusConflict, "conflict", "Already exists.")

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", rec.Code)
	}
	want := `{"error":{"code":"conflict","message":"Already exists."}}` + "\n"
	if got := rec.Body.String(); got != want {
		t.Errorf("body = %q, want %q (no \"fields\" key when empty)", got, want)
	}
}

func TestError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantCode    string
		wantMessage string // "" means: don't check the exact text
		wantLogged  bool
	}{
		{
			name:        "bad request",
			err:         BadRequest("x"),
			wantStatus:  http.StatusBadRequest,
			wantCode:    "bad_request",
			wantMessage: "x",
		},
		{
			name:        "wrapped bad request (errors.As unwraps)",
			err:         fmt.Errorf("wrap: %w", BadRequest("y")),
			wantStatus:  http.StatusBadRequest,
			wantCode:    "bad_request",
			wantMessage: "y",
		},
		{
			name:        "not found",
			err:         database.ErrNotFound,
			wantStatus:  http.StatusNotFound,
			wantCode:    "not_found",
			wantMessage: NotFoundMsg,
		},
		{
			name:        "wrapped not found (errors.Is unwraps)",
			err:         fmt.Errorf("get: %w", database.ErrNotFound),
			wantStatus:  http.StatusNotFound,
			wantCode:    "not_found",
			wantMessage: NotFoundMsg,
		},
		{
			name:       "anything else is a 500",
			err:        errors.New("boom: password=hunter2"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal",
			wantLogged: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger, buf := newTestLogger()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/products/7", nil)
			req = req.WithContext(context.WithValue(req.Context(), requestIDKey, "abc123"))
			rec := httptest.NewRecorder()

			Error(rec, req, logger, tc.err)

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			assertErrorBody(t, rec, tc.wantCode)
			if tc.wantMessage != "" && !strings.Contains(rec.Body.String(), `"message":"`+tc.wantMessage+`"`) {
				t.Errorf("body = %s, want message %q", rec.Body, tc.wantMessage)
			}
			if tc.wantStatus == http.StatusInternalServerError && strings.Contains(rec.Body.String(), "boom") {
				t.Errorf("500 body leaks the internal error: %s", rec.Body)
			}

			lines := logLines(t, buf)
			if !tc.wantLogged {
				if len(lines) != 0 {
					t.Errorf("got %d log lines, want 0 (4xx is not a server error): %v", len(lines), lines)
				}
				return
			}
			if len(lines) != 1 {
				t.Fatalf("got %d log lines, want 1", len(lines))
			}
			line := lines[0]
			if line["level"] != "ERROR" {
				t.Errorf("level = %v, want ERROR", line["level"])
			}
			if err, _ := line["err"].(string); !strings.Contains(err, "boom") {
				t.Errorf("err = %v, want the real error in the log", line["err"])
			}
			if line["path"] != "/api/v1/products/7" {
				t.Errorf("path = %v", line["path"])
			}
			if line["request_id"] != "abc123" {
				t.Errorf("request_id = %v, want abc123", line["request_id"])
			}
		})
	}
}
