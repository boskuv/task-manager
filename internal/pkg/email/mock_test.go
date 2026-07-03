package email

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/pkg/circuitbreaker"
)

func TestMockServiceSendTeamInvite(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	breaker := circuitbreaker.New(circuitbreaker.Config{FailureThreshold: 1})
	svc := NewMockService(breaker, logger)

	err := svc.SendTeamInvite(context.Background(), "user@example.com", "Backend", domain.TeamRoleMember)
	if err != nil {
		t.Fatalf("SendTeamInvite: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("mock email: team invite sent")) {
		t.Fatalf("log = %q, want invite log entry", buf.String())
	}
}

func TestMockServiceUsesCircuitBreaker(t *testing.T) {
	t.Parallel()

	breaker := circuitbreaker.New(circuitbreaker.Config{FailureThreshold: 1})
	svc := NewMockService(breaker, slog.Default())
	svc.send = func(context.Context, string, string, domain.TeamRole) error {
		return errors.New("smtp unavailable")
	}

	if err := svc.SendTeamInvite(context.Background(), "user@example.com", "Backend", domain.TeamRoleMember); err == nil {
		t.Fatal("expected send error")
	}

	if err := svc.SendTeamInvite(context.Background(), "user@example.com", "Backend", domain.TeamRoleMember); !errors.Is(err, circuitbreaker.ErrOpen) {
		t.Fatalf("error = %v, want ErrOpen", err)
	}
}
