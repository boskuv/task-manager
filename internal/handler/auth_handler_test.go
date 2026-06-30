package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/dto"
)

func TestAuthHandlerRegister(t *testing.T) {
	t.Parallel()

	handler := NewAuthHandler(&stubAuthService{
		registerToken: "jwt-token",
	})

	body := bytes.NewBufferString(`{"email":"user@example.com","password":"password1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register", body)
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var resp dto.AuthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Token != "jwt-token" {
		t.Errorf("token = %q, want jwt-token", resp.Token)
	}
}

func TestAuthHandlerLogin(t *testing.T) {
	t.Parallel()

	handler := NewAuthHandler(&stubAuthService{
		loginToken: "jwt-token",
	})

	body := bytes.NewBufferString(`{"email":"user@example.com","password":"password1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", body)
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp dto.AuthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Token != "jwt-token" {
		t.Errorf("token = %q, want jwt-token", resp.Token)
	}
}

func TestAuthHandlerLoginUnauthorized(t *testing.T) {
	t.Parallel()

	handler := NewAuthHandler(&stubAuthService{
		loginErr: domain.ErrUnauthorized,
	})

	body := bytes.NewBufferString(`{"email":"user@example.com","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", body)
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthHandlerInvalidJSON(t *testing.T) {
	t.Parallel()

	handler := NewAuthHandler(&stubAuthService{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewBufferString("{"))
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

type stubAuthService struct {
	registerToken string
	registerErr   error
	loginToken    string
	loginErr      error
}

func (s *stubAuthService) Register(_ context.Context, _, _ string) (string, error) {
	if s.registerErr != nil {
		return "", s.registerErr
	}
	if s.registerToken == "" {
		return "", errors.New("unexpected call")
	}
	return s.registerToken, nil
}

func (s *stubAuthService) Login(_ context.Context, _, _ string) (string, error) {
	if s.loginErr != nil {
		return "", s.loginErr
	}
	if s.loginToken == "" {
		return "", domain.ErrUnauthorized
	}
	return s.loginToken, nil
}
