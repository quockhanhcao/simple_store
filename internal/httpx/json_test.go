package httpx

import (
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   int
		value    any
		wantBody string
	}{
		{"object", http.StatusOK, map[string]int{"id": 1}, `{"id":1}` + "\n"},
		{"slice with a non-200 status", http.StatusCreated, []string{"a", "b"}, `["a","b"]` + "\n"},
		{"empty slice is [] not null", http.StatusOK, []int{}, "[]\n"},
		{"HTML is escaped", http.StatusOK, map[string]string{"x": "<b>"}, `{"x":"` + "\\" + `u003cb` + "\\" + `u003e"}` + "\n"}, // < and > become < and >
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			if err := WriteJSON(rec, tc.status, tc.value); err != nil {
				t.Fatalf("WriteJSON: %v", err)
			}
			if rec.Code != tc.status {
				t.Errorf("status = %d, want %d", rec.Code, tc.status)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
				t.Errorf("Content-Type = %q", ct)
			}
			if got := rec.Body.String(); got != tc.wantBody {
				t.Errorf("body = %q, want %q", got, tc.wantBody)
			}
		})
	}
}

// TestWriteJSONEncodeError proves the buffer: when encoding fails, nothing at
// all reaches the client, so the caller can still send a clean 500.
func TestWriteJSONEncodeError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
	}{
		{"channel", map[string]any{"x": make(chan int)}},
		{"infinity", map[string]float64{"x": math.Inf(1)}},
		{"NaN", math.NaN()},
		{"function", func() {}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			if err := WriteJSON(rec, http.StatusOK, tc.value); err == nil {
				t.Fatal("WriteJSON returned nil, want an encoding error")
			}
			if rec.Body.Len() != 0 {
				t.Errorf("body = %q, want empty", rec.Body)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "" {
				t.Errorf("Content-Type = %q, want unset", ct)
			}
			if rec.Code != http.StatusOK || rec.Flushed {
				// httptest.ResponseRecorder starts at 200; if WriteHeader had
				// been called with our status we couldn't tell, so also make
				// sure the caller can still write a different one.
				t.Errorf("recorder state changed: code=%d flushed=%v", rec.Code, rec.Flushed)
			}
			WriteError(rec, http.StatusInternalServerError, "internal", "x")
			if rec.Code != http.StatusInternalServerError {
				t.Errorf("caller's 500 was ignored: status = %d (headers were already sent)", rec.Code)
			}
		})
	}
}
