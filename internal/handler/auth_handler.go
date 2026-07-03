package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/boskuv/task-manager/internal/dto"
)

type authService interface {
	Register(ctx context.Context, email, password string) (string, error)
	Login(ctx context.Context, email, password string) (string, error)
}

// AuthHandler serves registration and login endpoints.
type AuthHandler struct {
	auth authService
}

// NewAuthHandler creates an auth HTTP handler.
func NewAuthHandler(auth authService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Register handles POST /api/v1/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	token, err := h.auth.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		writeServiceError(w, r.Context(), err)
		return
	}

	writeJSON(w, http.StatusCreated, dto.AuthResponse{Token: token})
}

// Login handles POST /api/v1/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	token, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeServiceError(w, r.Context(), err)
		return
	}

	writeJSON(w, http.StatusOK, dto.AuthResponse{Token: token})
}
