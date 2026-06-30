package app

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/boskuv/task-manager/internal/handler/middleware"
	"github.com/boskuv/task-manager/internal/pkg/logging"
)

func TestHealthHandler(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	healthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(body) != "ok" {
		t.Errorf("body = %q, want ok", body)
	}
}

func TestNewRouterRegistersOpenAPIRoutes(t *testing.T) {
	t.Parallel()

	mux := newRouter(routerDeps{metrics: middleware.NewHTTPMetrics()})

	specReq := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	specRec := httptest.NewRecorder()
	mux.ServeHTTP(specRec, specReq)
	if specRec.Code != http.StatusOK {
		t.Fatalf("openapi status = %d, want %d", specRec.Code, http.StatusOK)
	}
	if !strings.Contains(specRec.Header().Get("Content-Type"), "yaml") {
		t.Fatalf("content-type = %q, want yaml", specRec.Header().Get("Content-Type"))
	}
	if !strings.Contains(specRec.Body.String(), "openapi: 3.0.3") {
		t.Fatal("expected openapi spec body")
	}

	swaggerReq := httptest.NewRequest(http.MethodGet, "/swagger/", nil)
	swaggerRec := httptest.NewRecorder()
	mux.ServeHTTP(swaggerRec, swaggerReq)
	if swaggerRec.Code != http.StatusOK {
		t.Fatalf("swagger status = %d, want %d", swaggerRec.Code, http.StatusOK)
	}
	if !strings.Contains(swaggerRec.Body.String(), "swagger-ui") {
		t.Fatal("expected swagger ui html")
	}
}

func TestNewRouterRegistersHealthRoute(t *testing.T) {
	t.Parallel()

	mux := newRouter(routerDeps{metrics: middleware.NewHTTPMetrics()})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestNewRouterRegistersMetricsRoute(t *testing.T) {
	t.Parallel()

	mux := newRouter(routerDeps{metrics: middleware.NewHTTPMetrics()})

	healthReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthRec := httptest.NewRecorder()
	mux.ServeHTTP(healthRec, healthReq)
	if healthRec.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", healthRec.Code, http.StatusOK)
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "task_manager_http_requests_total") {
		t.Fatalf("metrics body = %q, want requests counter", rec.Body.String())
	}
}

func TestNewRouterSetsRequestIDHeader(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger, err := logging.NewWithWriter(logging.Config{Level: "info", Format: "json"}, &buf)
	if err != nil {
		t.Fatalf("NewWithWriter: %v", err)
	}

	mux := newRouter(routerDeps{
		metrics: middleware.NewHTTPMetrics(),
		logger:  logger,
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get(middleware.RequestIDHeader); got == "" {
		t.Fatal("expected request id response header")
	}
	if !strings.Contains(buf.String(), `"msg":"http request"`) {
		t.Fatalf("log = %q, want access log entry", buf.String())
	}
}

func TestOpenMySQLEmptyDSN(t *testing.T) {
	t.Parallel()

	_, err := openMySQL(DatabaseConfig{})
	if err == nil {
		t.Fatal("expected error for empty dsn")
	}
	if !strings.Contains(err.Error(), "dsn is required") {
		t.Errorf("error = %q, want dsn required message", err.Error())
	}
}
