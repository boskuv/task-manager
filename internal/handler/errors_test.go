package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/dto"
)

func TestMapErrorDomainSentinels(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		err     error
		status  int
		code    string
		message string
	}{
		{name: "invalid input", err: domain.ErrInvalidInput, status: http.StatusBadRequest, code: "invalid_input", message: "invalid input"},
		{name: "conflict", err: domain.ErrConflict, status: http.StatusConflict, code: "conflict", message: "conflict"},
		{name: "unauthorized", err: domain.ErrUnauthorized, status: http.StatusUnauthorized, code: "unauthorized", message: "unauthorized"},
		{name: "forbidden", err: domain.ErrForbidden, status: http.StatusForbidden, code: "forbidden", message: "forbidden"},
		{name: "not found", err: domain.ErrNotFound, status: http.StatusNotFound, code: "not_found", message: "not found"},
		{name: "wrapped", err: errors.Join(errors.New("db"), domain.ErrNotFound), status: http.StatusNotFound, code: "not_found", message: "not found"},
		{name: "client error", err: ClientError("invalid team_id"), status: http.StatusBadRequest, code: "bad_request", message: "invalid team_id"},
		{name: "unknown", err: errors.New("boom"), status: http.StatusInternalServerError, code: "internal_error", message: "internal server error"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mapped := MapError(tc.err)
			if mapped.Status != tc.status {
				t.Fatalf("status = %d, want %d", mapped.Status, tc.status)
			}
			if mapped.Code != tc.code {
				t.Fatalf("code = %q, want %q", mapped.Code, tc.code)
			}
			if mapped.Message != tc.message {
				t.Fatalf("message = %q, want %q", mapped.Message, tc.message)
			}
		})
	}
}

func TestWriteServiceErrorResponseBody(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	WriteServiceError(rec, context.Background(), domain.ErrForbidden)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	var body dto.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Error != "forbidden" {
		t.Fatalf("error = %q, want forbidden", body.Error)
	}
	if body.Code != "forbidden" {
		t.Fatalf("code = %q, want forbidden", body.Code)
	}
}

func TestWriteErrorIncludesCode(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	WriteError(rec, http.StatusTooManyRequests, "rate limit exceeded")

	var body dto.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != "rate_limit_exceeded" {
		t.Fatalf("code = %q, want rate_limit_exceeded", body.Code)
	}
}
