package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/boskuv/task-manager/internal/handler"
	"github.com/boskuv/task-manager/internal/pkg/logging"
)

func TestRequestIDGeneratesHeader(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := logging.RequestIDFromContext(r.Context()); !ok {
			t.Fatal("request id missing in context")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	RequestID(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get(RequestIDHeader); got == "" {
		t.Fatal("expected generated request id header")
	}
}

func TestRequestIDPropagatesIncomingHeader(t *testing.T) {
	t.Parallel()

	const incoming = "client-request-id"

	var got string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		got, ok = logging.RequestIDFromContext(r.Context())
		if !ok {
			t.Fatal("request id missing in context")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set(RequestIDHeader, incoming)
	rec := httptest.NewRecorder()

	RequestID(next).ServeHTTP(rec, req)

	if got != incoming {
		t.Fatalf("context request id = %q, want %q", got, incoming)
	}
	if rec.Header().Get(RequestIDHeader) != incoming {
		t.Fatalf("response header = %q, want %q", rec.Header().Get(RequestIDHeader), incoming)
	}
}

func TestAccessLogWritesStructuredEntry(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger, err := logging.NewWithWriter(logging.Config{Level: "info", Format: "json"}, &buf)
	if err != nil {
		t.Fatalf("NewWithWriter: %v", err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", nil)
	req.Pattern = "POST /api/v1/tasks"
	req = req.WithContext(logging.ContextWithRequestID(req.Context(), "req-42"))
	rec := httptest.NewRecorder()

	AccessLog(logger)(next).ServeHTTP(rec, req)

	logLine := buf.String()
	for _, want := range []string{
		`"msg":"http request"`,
		`"request_id":"req-42"`,
		`"method":"POST"`,
		`"path":"POST /api/v1/tasks"`,
		`"status":201`,
	} {
		if !strings.Contains(logLine, want) {
			t.Fatalf("log = %q, want substring %q", logLine, want)
		}
	}
	if strings.Count(logLine, `"request_id":"req-42"`) != 1 {
		t.Fatalf("log = %q, want request_id exactly once", logLine)
	}
}

func TestAccessLogIncludesUserID(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger, err := logging.NewWithWriter(logging.Config{Level: "info", Format: "json"}, &buf)
	if err != nil {
		t.Fatalf("NewWithWriter: %v", err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Pattern = "GET /api/v1/me"
	req = req.WithContext(handler.ContextWithUserID(req.Context(), 99))
	rec := httptest.NewRecorder()

	AccessLog(logger)(next).ServeHTTP(rec, req)

	if !strings.Contains(buf.String(), `"user_id":99`) {
		t.Fatalf("log = %q, want user_id field", buf.String())
	}
}

func TestAccessLogSkipsMetricsEndpoint(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger, err := logging.NewWithWriter(logging.Config{Level: "info", Format: "json"}, &buf)
	if err != nil {
		t.Fatalf("NewWithWriter: %v", err)
	}

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	AccessLog(logger)(next).ServeHTTP(rec, req)

	if !called {
		t.Fatal("next handler should be called")
	}
	if buf.Len() != 0 {
		t.Fatalf("log = %q, want empty", buf.String())
	}
}
