package team

import (
	"context"
	"strings"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/repository"
)

const maxTeamNameLength = 255

// Service handles team creation and listing.
type Service struct {
	teams repository.TeamRepository
}

// NewService creates a team use case service.
func NewService(teams repository.TeamRepository) *Service {
	return &Service{teams: teams}
}

// Create creates a team and adds the creator as owner.
func (s *Service) Create(ctx context.Context, userID int64, name string) (domain.Team, error) {
	name, err := validateTeamName(name)
	if err != nil {
		return domain.Team{}, err
	}
	if userID <= 0 {
		return domain.Team{}, domain.ErrInvalidInput
	}

	team, err := s.teams.Create(ctx, domain.Team{
		Name:      name,
		CreatedBy: userID,
	})
	if err != nil {
		return domain.Team{}, err
	}

	if err := s.teams.AddMember(ctx, domain.TeamMember{
		TeamID: team.ID,
		UserID: userID,
		Role:   domain.TeamRoleOwner,
	}); err != nil {
		return domain.Team{}, err
	}

	return team, nil
}

// List returns teams where the user is a member.
func (s *Service) List(ctx context.Context, userID int64) ([]domain.Team, error) {
	if userID <= 0 {
		return nil, domain.ErrInvalidInput
	}
	return s.teams.ListByUserID(ctx, userID)
}

func validateTeamName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", domain.ErrInvalidInput
	}
	if len(name) > maxTeamNameLength {
		return "", domain.ErrInvalidInput
	}
	return name, nil
}
