package main

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/quockhanhcao/simple_store/internal/config"
)

func TestRoutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		env        string
		method     string
		path       string
		wantStatus int
		wantBody   string // checked only when non-empty
	}{
		{"alive in Development", "Development", http.MethodGet, "/alive", http.StatusOK, "Healthy"},
		{"HEAD alive is allowed", "Development", http.MethodHead, "/alive", http.StatusOK, ""},
		{"wrong method on alive", "Development", http.MethodPost, "/alive", http.StatusMethodNotAllowed, ""},
		{"alive is not mapped in Production", "Production", http.MethodGet, "/alive", http.StatusNotFound, ""},
		{"unknown path", "Development", http.MethodGet, "/nope", http.StatusNotFound, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// DiscardHandler: tests don't need log output.
			srv := NewHandler(slog.New(slog.DiscardHandler), &config.Config{Env: tt.env})
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)

			srv.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("%s %s: status = %d, want %d", tt.method, tt.path, rec.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
			if tt.wantStatus == http.StatusOK {
				if ct := rec.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
					t.Errorf("Content-Type = %q, want %q", ct, "text/plain; charset=utf-8")
				}
			}
		})
	}
}
