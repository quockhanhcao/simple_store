package httpx

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

// newTestLogger returns a JSON logger that writes into a buffer, so a test can
// read back exactly what was logged.
func newTestLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(slog.NewJSONHandler(&buf, nil)), &buf
}

// logLines decodes every JSON log line in buf.
func logLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var lines []map[string]any
	dec := json.NewDecoder(buf)
	for {
		var line map[string]any
		err := dec.Decode(&line)
		if err == io.EOF {
			return lines
		}
		if err != nil {
			t.Fatalf("log output is not JSON lines: %v", err)
		}
		lines = append(lines, line)
	}
}

// assertErrorBody checks the {"error":{"code","message"}} envelope.
func assertErrorBody(t *testing.T, rec *httptest.ResponseRecorder, wantCode string) {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want JSON", ct)
	}
	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not the error envelope: %v; body = %s", err, rec.Body)
	}
	if body.Error.Code != wantCode {
		t.Errorf("error.code = %q, want %q", body.Error.Code, wantCode)
	}
	if body.Error.Message == "" {
		t.Error("error.message is empty")
	}
}

// ---------- Chain ----------

func TestChainOrder(t *testing.T) {
	t.Parallel()

	var order []string
	mark := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+" in")
				next.ServeHTTP(w, r)
				order = append(order, name+" out")
			})
		}
	}
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	h := Chain(final, mark("a"), mark("b"))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	got := strings.Join(order, ", ")
	want := "a in, b in, handler, b out, a out"
	if got != want {
		t.Errorf("order = %q, want %q", got, want)
	}
}

// ---------- Recover ----------

func TestRecoverPanicGives500(t *testing.T) {
	t.Parallel()

	logger, buf := newTestLogger()
	h := Recover(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if got := rec.Header().Get("Connection"); got != "close" {
		t.Errorf("Connection = %q, want close", got)
	}
	assertErrorBody(t, rec, "internal")
	if strings.Contains(rec.Body.String(), "boom") {
		t.Errorf("body leaks the panic value: %s", rec.Body)
	}

	lines := logLines(t, buf)
	if len(lines) != 1 {
		t.Fatalf("got %d log lines, want 1", len(lines))
	}
	if lines[0]["level"] != "ERROR" {
		t.Errorf("level = %v, want ERROR", lines[0]["level"])
	}
	if lines[0]["panic"] != "boom" {
		t.Errorf("panic = %v, want boom", lines[0]["panic"])
	}
	if stack, _ := lines[0]["stack"].(string); stack == "" {
		t.Error("log line has no stack")
	}
}

func TestRecoverKeepsServing(t *testing.T) {
	t.Parallel()

	h := Recover(slog.New(slog.DiscardHandler))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panic" {
			panic("boom")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	for _, tc := range []struct {
		path string
		want int
	}{
		{"/panic", http.StatusInternalServerError},
		{"/ok", http.StatusNoContent},
		{"/panic", http.StatusInternalServerError},
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if rec.Code != tc.want {
			t.Errorf("%s: status = %d, want %d", tc.path, rec.Code, tc.want)
		}
	}
}

func TestRecoverRepanicsOnErrAbortHandler(t *testing.T) {
	t.Parallel()

	logger, buf := newTestLogger()
	h := Recover(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(http.ErrAbortHandler)
	}))

	defer func() {
		if v := recover(); v != http.ErrAbortHandler {
			t.Errorf("recovered %v, want http.ErrAbortHandler to be re-panicked", v)
		}
		if lines := logLines(t, buf); len(lines) != 0 {
			t.Errorf("got %d log lines, want 0 (an abort is not an error)", len(lines))
		}
	}()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	t.Error("ServeHTTP returned normally, want a re-panic")
}

// ---------- RequestID ----------

var hex16 = regexp.MustCompile(`^[0-9a-f]{16}$`)

func TestRequestID(t *testing.T) {
	t.Parallel()

	var seen string
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = RequestIDFrom(r.Context())
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	id := rec.Header().Get("X-Request-Id")
	if !hex16.MatchString(id) {
		t.Fatalf("X-Request-Id = %q, want 16 lowercase hex chars", id)
	}
	if seen != id {
		t.Errorf("handler saw ID %q, header has %q", seen, id)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/", nil))
	if id2 := rec2.Header().Get("X-Request-Id"); id2 == id {
		t.Errorf("two requests got the same ID %q", id)
	}
}

func TestRequestIDIgnoresClientHeader(t *testing.T) {
	t.Parallel()

	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Id", "client-chosen")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-Id"); got == "client-chosen" {
		t.Error("the client's X-Request-Id was trusted")
	}
}

func TestRequestIDFromOutsideRequest(t *testing.T) {
	t.Parallel()

	if got := RequestIDFrom(t.Context()); got != "" {
		t.Errorf("RequestIDFrom(empty context) = %q, want \"\"", got)
	}
}

// ---------- AccessLog ----------

func TestAccessLog(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantStatus int
		wantBytes  int
	}{
		{
			name: "WriteHeader then Write",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("created"))
			},
			wantStatus: http.StatusCreated,
			wantBytes:  len("created"),
		},
		{
			name: "only Write means 200",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("hello"))
			},
			wantStatus: http.StatusOK,
			wantBytes:  len("hello"),
		},
		{
			name:       "writes nothing means 200",
			handler:    func(w http.ResponseWriter, r *http.Request) {},
			wantStatus: http.StatusOK,
			wantBytes:  0,
		},
		{
			name: "second WriteHeader is ignored",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusAccepted)
				w.WriteHeader(http.StatusTeapot)
			},
			wantStatus: http.StatusAccepted,
			wantBytes:  0,
		},
		{
			name: "error envelope from a handler",
			handler: func(w http.ResponseWriter, r *http.Request) {
				WriteError(w, http.StatusBadRequest, "bad_request", "id must be an integer.")
			},
			wantStatus: http.StatusBadRequest,
			wantBytes:  -1, // body size depends on the envelope; just check it's > 0
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger, buf := newTestLogger()
			h := AccessLog(logger)(tc.handler)

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/things", nil))

			lines := logLines(t, buf)
			if len(lines) != 1 {
				t.Fatalf("got %d log lines, want exactly 1: %v", len(lines), lines)
			}
			line := lines[0]

			if line["msg"] != "request" {
				t.Errorf("msg = %v, want request", line["msg"])
			}
			if line["method"] != http.MethodPost {
				t.Errorf("method = %v, want POST", line["method"])
			}
			if line["path"] != "/things" {
				t.Errorf("path = %v, want /things", line["path"])
			}
			// encoding/json decodes every number as float64.
			if got := line["status"]; got != float64(tc.wantStatus) {
				t.Errorf("status = %v, want %d", got, tc.wantStatus)
			}
			bytesLogged, _ := line["bytes"].(float64)
			if tc.wantBytes >= 0 && bytesLogged != float64(tc.wantBytes) {
				t.Errorf("bytes = %v, want %d", line["bytes"], tc.wantBytes)
			}
			if tc.wantBytes < 0 && int(bytesLogged) != rec.Body.Len() {
				t.Errorf("bytes = %v, but the body is %d bytes", line["bytes"], rec.Body.Len())
			}
			if _, ok := line["duration"]; !ok {
				t.Error("log line has no duration")
			}
		})
	}
}

// TestFullStack runs the real order from server.go. It proves AccessLog sits
// outside Recover (the panic is logged as 500) and inside RequestID (both log
// lines carry the same ID as the response header).
func TestFullStack(t *testing.T) {
	t.Parallel()

	logger, buf := newTestLogger()
	panicky := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	h := Chain(panicky, RequestID, AccessLog(logger), Recover(logger))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	id := rec.Header().Get("X-Request-Id")
	if id == "" {
		t.Fatal("no X-Request-Id header")
	}

	lines := logLines(t, buf)
	if len(lines) != 2 {
		t.Fatalf("got %d log lines, want 2 (panic + access): %v", len(lines), lines)
	}
	panicLine, accessLine := lines[0], lines[1]

	if panicLine["msg"] != "panic" {
		t.Errorf("first line msg = %v, want panic", panicLine["msg"])
	}
	if accessLine["msg"] != "request" {
		t.Errorf("second line msg = %v, want request", accessLine["msg"])
	}
	if accessLine["status"] != float64(http.StatusInternalServerError) {
		t.Errorf("access log status = %v, want 500", accessLine["status"])
	}
	for _, line := range lines {
		if line["request_id"] != id {
			t.Errorf("%v line request_id = %v, want %q", line["msg"], line["request_id"], id)
		}
	}
}

// ---------- statusRecorder ----------

// TestStatusRecorderUnwrap checks that http.ResponseController can reach the
// real writer's Flush through the recorder.
func TestStatusRecorderUnwrap(t *testing.T) {
	t.Parallel()

	inner := httptest.NewRecorder()
	rec := &statusRecorder{ResponseWriter: inner}

	if err := http.NewResponseController(rec).Flush(); err != nil {
		t.Fatalf("Flush through the recorder: %v", err)
	}
	if !inner.Flushed {
		t.Error("the inner writer was not flushed")
	}
}

// ---------- CrossOrigin ----------

func TestCrossOrigin(t *testing.T) {
	t.Parallel()

	h := CrossOrigin()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// httptest.NewRequest uses Host "example.com", so that is "our" origin.
	tests := []struct {
		name    string
		method  string
		headers map[string]string
		want    int
	}{
		{"cross-site POST", http.MethodPost, map[string]string{"Sec-Fetch-Site": "cross-site"}, http.StatusForbidden},
		{"cross-site PUT", http.MethodPut, map[string]string{"Sec-Fetch-Site": "cross-site"}, http.StatusForbidden},
		{"cross-site DELETE", http.MethodDelete, map[string]string{"Sec-Fetch-Site": "cross-site"}, http.StatusForbidden},
		{"same-site POST (other subdomain)", http.MethodPost, map[string]string{"Sec-Fetch-Site": "same-site"}, http.StatusForbidden},
		{"same-origin POST", http.MethodPost, map[string]string{"Sec-Fetch-Site": "same-origin"}, http.StatusOK},
		{"user typed the URL (none)", http.MethodPost, map[string]string{"Sec-Fetch-Site": "none"}, http.StatusOK},
		{"POST with no headers (curl)", http.MethodPost, nil, http.StatusOK},
		{"cross-site GET", http.MethodGet, map[string]string{"Sec-Fetch-Site": "cross-site"}, http.StatusOK},
		{"cross-site HEAD", http.MethodHead, map[string]string{"Sec-Fetch-Site": "cross-site"}, http.StatusOK},
		{"old browser: foreign Origin", http.MethodPost, map[string]string{"Origin": "https://evil.example"}, http.StatusForbidden},
		{"old browser: our own Origin", http.MethodPost, map[string]string{"Origin": "http://example.com"}, http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tc.method, "/orders", nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}
