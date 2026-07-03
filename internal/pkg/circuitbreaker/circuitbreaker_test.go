package circuitbreaker

import (
	"errors"
	"testing"
	"time"
)

func TestBreakerOpensAfterFailures(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	breaker := New(Config{FailureThreshold: 2, OpenTimeout: time.Minute})
	breaker.now = func() time.Time { return now }

	callErr := errors.New("send failed")
	for range 2 {
		if err := breaker.Execute(func() error { return callErr }); !errors.Is(err, callErr) {
			t.Fatalf("expected send error, got %v", err)
		}
	}

	if breaker.State() != StateOpen {
		t.Fatalf("state = %v, want open", breaker.State())
	}

	if err := breaker.Execute(func() error { return nil }); !errors.Is(err, ErrOpen) {
		t.Fatalf("error = %v, want ErrOpen", err)
	}
}

func TestBreakerRecoversAfterTimeout(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	current := start
	breaker := New(Config{FailureThreshold: 1, OpenTimeout: time.Second, SuccessThreshold: 1})
	breaker.now = func() time.Time { return current }

	_ = breaker.Execute(func() error { return errors.New("fail") })
	if breaker.State() != StateOpen {
		t.Fatalf("state = %v, want open", breaker.State())
	}

	current = start.Add(2 * time.Second)
	if err := breaker.Execute(func() error { return nil }); err != nil {
		t.Fatalf("Execute after timeout: %v", err)
	}
	if breaker.State() != StateClosed {
		t.Fatalf("state = %v, want closed", breaker.State())
	}
}

func TestBreakerSuccessResetsFailures(t *testing.T) {
	t.Parallel()

	breaker := New(Config{FailureThreshold: 2})
	breaker.now = time.Now

	_ = breaker.Execute(func() error { return errors.New("fail") })
	if err := breaker.Execute(func() error { return nil }); err != nil {
		t.Fatalf("Execute success: %v", err)
	}
	if err := breaker.Execute(func() error { return errors.New("fail") }); err == nil {
		t.Fatal("expected failure")
	}
	if breaker.State() != StateClosed {
		t.Fatalf("state = %v, want closed after single failure post-reset", breaker.State())
	}
}
