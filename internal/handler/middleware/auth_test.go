package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtmanager "github.com/boskuv/task-manager/internal/pkg/jwt"
)

func TestAuthMiddlewareValidToken(t *testing.T) {
	t.Parallel()

	parser := jwtmanager.NewManager("secret", time.Hour)
	token, err := parser.GenerateAccessToken(99)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	var gotUserID int64
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Fatal("user id missing in context")
		}
		gotUserID = userID
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	Auth(parser)(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if gotUserID != 99 {
		t.Errorf("userID = %d, want 99", gotUserID)
	}
}

func TestAuthMiddlewareMissingToken(t *testing.T) {
	t.Parallel()

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	Auth(jwtmanager.NewManager("secret", 0))(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if called {
		t.Fatal("next handler should not be called")
	}
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	Auth(jwtmanager.NewManager("secret", 0))(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestUserIDFromContextMissing(t *testing.T) {
	t.Parallel()

	if _, ok := UserIDFromContext(context.Background()); ok {
		t.Fatal("expected missing user id")
	}
}
