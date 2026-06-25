package team

import (
	"context"
	"net/mail"
	"strings"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/repository"
)

const maxTeamNameLength = 255

// Service handles team creation, listing, and invites.
type Service struct {
	teams repository.TeamRepository
	users repository.UserRepository
}

// NewService creates a team use case service.
func NewService(teams repository.TeamRepository, users repository.UserRepository) *Service {
	return &Service{
		teams: teams,
		users: users,
	}
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

// Invite adds an existing user to a team. Only owner or admin may invite.
func (s *Service) Invite(ctx context.Context, actorUserID, teamID int64, email, role string) (domain.TeamMember, error) {
	if actorUserID <= 0 || teamID <= 0 {
		return domain.TeamMember{}, domain.ErrInvalidInput
	}

	email, inviteRole, err := validateInviteInput(email, role)
	if err != nil {
		return domain.TeamMember{}, err
	}

	if _, err := s.teams.GetByID(ctx, teamID); err != nil {
		return domain.TeamMember{}, err
	}

	actorRole, err := s.teams.GetMemberRole(ctx, teamID, actorUserID)
	if err != nil {
		if err == domain.ErrNotFound {
			return domain.TeamMember{}, domain.ErrForbidden
		}
		return domain.TeamMember{}, err
	}
	if !actorRole.CanInvite() {
		return domain.TeamMember{}, domain.ErrForbidden
	}

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return domain.TeamMember{}, err
	}

	member := domain.TeamMember{
		TeamID: teamID,
		UserID: user.ID,
		Role:   inviteRole,
	}
	if err := s.teams.AddMember(ctx, member); err != nil {
		return domain.TeamMember{}, err
	}

	return member, nil
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

func validateInviteInput(email, role string) (string, domain.TeamRole, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	role = strings.TrimSpace(strings.ToLower(role))

	if email == "" || role == "" {
		return "", "", domain.ErrInvalidInput
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return "", "", domain.ErrInvalidInput
	}

	inviteRole := domain.TeamRole(role)
	if !inviteRole.IsValid() || inviteRole == domain.TeamRoleOwner {
		return "", "", domain.ErrInvalidInput
	}

	return email, inviteRole, nil
}
