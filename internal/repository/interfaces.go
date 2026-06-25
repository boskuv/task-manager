package repository

import (
	"context"

	"github.com/boskuv/task-manager/internal/domain"
)

// UserRepository persists and loads users.
type UserRepository interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByID(ctx context.Context, id int64) (domain.User, error)
}

// TeamRepository persists and loads teams and memberships.
type TeamRepository interface {
	Create(ctx context.Context, team domain.Team) (domain.Team, error)
	ListByUserID(ctx context.Context, userID int64) ([]domain.Team, error)
	GetByID(ctx context.Context, id int64) (domain.Team, error)
	AddMember(ctx context.Context, member domain.TeamMember) error
	GetMemberRole(ctx context.Context, teamID, userID int64) (domain.TeamRole, error)
}
