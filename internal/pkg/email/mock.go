package email

import (
	"context"
	"log/slog"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/pkg/circuitbreaker"
)

// MockService logs invite emails and protects sends with a circuit breaker.
type MockService struct {
	breaker *circuitbreaker.Breaker
	logger  *slog.Logger
	send    func(ctx context.Context, toEmail, teamName string, role domain.TeamRole) error
}

// NewMockService creates a mock email sender wrapped by a circuit breaker.
func NewMockService(breaker *circuitbreaker.Breaker, logger *slog.Logger) *MockService {
	if logger == nil {
		logger = slog.Default()
	}

	svc := &MockService{
		breaker: breaker,
		logger:  logger,
	}
	svc.send = svc.logInvite
	return svc
}

// SendTeamInvite sends a mock team invite email.
func (s *MockService) SendTeamInvite(ctx context.Context, toEmail, teamName string, role domain.TeamRole) error {
	return s.breaker.Execute(func() error {
		return s.send(ctx, toEmail, teamName, role)
	})
}

func (s *MockService) logInvite(ctx context.Context, toEmail, teamName string, role domain.TeamRole) error {
	s.logger.InfoContext(ctx, "mock email: team invite sent",
		"to", toEmail,
		"team", teamName,
		"role", string(role),
	)
	return nil
}
