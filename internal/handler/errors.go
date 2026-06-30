package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/dto"
)

// MappedError is the HTTP representation of an error.
type MappedError struct {
	Status  int
	Code    string
	Message string
}

// clientError represents a handler-level validation error.
type clientError struct {
	message string
}

func (e clientError) Error() string {
	return e.message
}

// ClientError builds a validation error mapped to HTTP 400.
func ClientError(message string) error {
	return clientError{message: message}
}

// MapError maps domain and handler errors to a unified API error response.
func MapError(err error) MappedError {
	if err == nil {
		return MappedError{Status: http.StatusOK}
	}

	var validationErr clientError
	if errors.As(err, &validationErr) {
		return MappedError{
			Status:  http.StatusBadRequest,
			Code:    "bad_request",
			Message: validationErr.message,
		}
	}

	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return MappedError{Status: http.StatusBadRequest, Code: "invalid_input", Message: "invalid input"}
	case errors.Is(err, domain.ErrConflict):
		return MappedError{Status: http.StatusConflict, Code: "conflict", Message: "conflict"}
	case errors.Is(err, domain.ErrUnauthorized):
		return MappedError{Status: http.StatusUnauthorized, Code: "unauthorized", Message: "unauthorized"}
	case errors.Is(err, domain.ErrForbidden):
		return MappedError{Status: http.StatusForbidden, Code: "forbidden", Message: "forbidden"}
	case errors.Is(err, domain.ErrNotFound):
		return MappedError{Status: http.StatusNotFound, Code: "not_found", Message: "not found"}
	default:
		return MappedError{
			Status:  http.StatusInternalServerError,
			Code:    "internal_error",
			Message: "internal server error",
		}
	}
}

// WriteErrorResponse writes a unified JSON error body.
func WriteErrorResponse(w http.ResponseWriter, status int, payload dto.ErrorResponse) {
	writeJSON(w, status, payload)
}

// WriteError writes a unified JSON error with an inferred code from status.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteErrorResponse(w, status, dto.ErrorResponse{
		Error: message,
		Code:  codeForStatus(status, message),
	})
}

// WriteServiceError maps a service error and writes the unified JSON response.
func WriteServiceError(w http.ResponseWriter, ctx context.Context, err error) {
	mapped := MapError(err)
	if mapped.Status >= http.StatusInternalServerError {
		slog.ErrorContext(ctx, "service error", "error", err, "code", mapped.Code)
	}
	WriteErrorResponse(w, mapped.Status, dto.ErrorResponse{
		Error: mapped.Message,
		Code:  mapped.Code,
	})
}

func writeError(w http.ResponseWriter, status int, message string) {
	WriteError(w, status, message)
}

func writeServiceError(w http.ResponseWriter, ctx context.Context, err error) {
	WriteServiceError(w, ctx, err)
}

func codeForStatus(status int, message string) string {
	switch status {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusConflict:
		return "conflict"
	case http.StatusTooManyRequests:
		return "rate_limit_exceeded"
	case http.StatusInternalServerError:
		return "internal_error"
	default:
		if message != "" {
			return "error"
		}
		return "error"
	}
}
