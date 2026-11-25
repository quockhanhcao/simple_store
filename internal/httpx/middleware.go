package httpx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"slices"
	"time"
)

type Middleware func(http.Handler) http.Handler

// Wraps h, mws[0] is the outer most layer
// => applied in reverse order
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for _, mw := range slices.Backward(mws) {
		h = mw(h)
	}
	return h
}

func Recover(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{ResponseWriter: w}
			defer func() {
				v := recover()
				if v == nil {
					return
				}
				if v == http.ErrAbortHandler {
					panic(v) // the "abort" signal: let net/http handle it
				}
				logger.Error("panic", "method", r.Method, "path", r.URL.Path,
					"panic", v, "stack", string(debug.Stack()), "request_id", RequestIDFrom(r.Context()))
				if rec.status != 0 {
					panic(http.ErrAbortHandler)
				}
				w.Header().Set("Connection", "close")
				WriteError(w, http.StatusInternalServerError, "internal", "An unexpected error occurred.")
			}()
			next.ServeHTTP(w, r)
		})
	}
}

type ctxKey int

const requestIDKey ctxKey = 0

func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := newRequestID()
		w.Header().Set("X-Request-Id", id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newRequestID() string {
	b := make([]byte, 8)
	rand.Read(b)
	hexStr := hex.EncodeToString(b)
	return hexStr
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(code int) {
	if r.status == 0 {
		r.status = code
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func AccessLog(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w}
			completed := false
			defer func() {
				status := rec.status
				if status == 0 {
					status = http.StatusOK
				}
				logger.Info("request", "method", r.Method, "path", r.URL.Path, "status", rec.status, "bytes", rec.bytes, "duration", time.Since(start), "aborted", !completed, "request_id", RequestIDFrom(r.Context()))
			}()
			next.ServeHTTP(rec, r)
			completed = true
		})
	}
}

// CrossOrigin rejects cross-site POST/PUT/DELETE from browsers
// apply it per route, only to the routes that had antiforgery
func CrossOrigin() Middleware {
	cop := http.NewCrossOriginProtection()
	return cop.Handler
}
