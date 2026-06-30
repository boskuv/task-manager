package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/boskuv/task-manager/internal/dto"
	"github.com/boskuv/task-manager/internal/pkg/logging"
)

func TestRecoverHandlesPanic(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger, err := logging.NewWithWriter(logging.Config{Level: "error", Format: "json"}, &buf)
	if err != nil {
		t.Fatalf("NewWithWriter: %v", err)
	}

	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	Recover(logger)(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var body dto.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Error != "internal server error" {
		t.Fatalf("error = %q, want internal server error", body.Error)
	}
	if body.Code != "internal_error" {
		t.Fatalf("code = %q, want internal_error", body.Code)
	}
}
