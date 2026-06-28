package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/boskuv/task-manager/internal/dto"
	"github.com/boskuv/task-manager/internal/handler"
)

type stubRateLimiter struct {
	allow bool
	err   error
	calls int
	last  string
}

func (s *stubRateLimiter) Allow(_ context.Context, key string) (bool, error) {
	s.calls++
	s.last = key
	return s.allow, s.err
}

func TestRateLimitAllowsRequest(t *testing.T) {
	t.Parallel()

	limiter := &stubRateLimiter{allow: true}
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req = req.WithContext(handler.ContextWithUserID(req.Context(), 42))
	rec := httptest.NewRecorder()

	RateLimit(limiter, 100)(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if !called {
		t.Fatal("next handler should be called")
	}
	if limiter.calls != 1 {
		t.Fatalf("limiter calls = %d, want 1", limiter.calls)
	}
	if limiter.last != "rate_limit:user:42" {
		t.Fatalf("limiter key = %q, want rate_limit:user:42", limiter.last)
	}
}

func TestRateLimitRejectsWhenExceeded(t *testing.T) {
	t.Parallel()

	limiter := &stubRateLimiter{allow: false}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req = req.WithContext(handler.ContextWithUserID(req.Context(), 7))
	rec := httptest.NewRecorder()

	RateLimit(limiter, 100)(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}

	var body dto.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Error != "rate limit exceeded" {
		t.Fatalf("error = %q, want rate limit exceeded", body.Error)
	}
}

func TestRateLimitMissingUserID(t *testing.T) {
	t.Parallel()

	limiter := &stubRateLimiter{allow: true}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	RateLimit(limiter, 100)(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if limiter.calls != 0 {
		t.Fatalf("limiter calls = %d, want 0", limiter.calls)
	}
}

func TestRateLimitDisabledWhenLimitZero(t *testing.T) {
	t.Parallel()

	limiter := &stubRateLimiter{allow: false}
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req = req.WithContext(handler.ContextWithUserID(req.Context(), 1))
	rec := httptest.NewRecorder()

	RateLimit(limiter, 0)(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if !called {
		t.Fatal("next handler should be called when limit is disabled")
	}
	if limiter.calls != 0 {
		t.Fatalf("limiter calls = %d, want 0", limiter.calls)
	}
}

func TestRateLimitFailsOpenOnStoreError(t *testing.T) {
	t.Parallel()

	limiter := &stubRateLimiter{err: context.DeadlineExceeded}
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req = req.WithContext(handler.ContextWithUserID(req.Context(), 1))
	rec := httptest.NewRecorder()

	RateLimit(limiter, 100)(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if !called {
		t.Fatal("next handler should be called when limiter errors")
	}
}

func TestChainRunsMiddlewareInOrder(t *testing.T) {
	t.Parallel()

	var order []string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	first := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "first")
			next.ServeHTTP(w, r)
		})
	}
	second := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "second")
			next.ServeHTTP(w, r)
		})
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	Chain(first, second)(next).ServeHTTP(rec, req)

	want := []string{"first", "second", "handler"}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order = %v, want %v", order, want)
		}
	}
}
