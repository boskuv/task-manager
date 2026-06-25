package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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

func TestNewRouterRegistersHealthRoute(t *testing.T) {
	t.Parallel()

	mux := newRouter(routerDeps{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
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
