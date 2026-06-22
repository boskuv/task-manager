package model

import (
	"time"

	"github.com/boskuv/task-manager/internal/domain"
)

// Team is the persistence representation of domain.Team.
type Team struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedBy int64     `db:"created_by"`
	CreatedAt time.Time `db:"created_at"`
}

// ToDomain converts the persistence model to a domain entity.
func (t Team) ToDomain() domain.Team {
	return domain.Team{
		ID:        t.ID,
		Name:      t.Name,
		CreatedBy: t.CreatedBy,
		CreatedAt: t.CreatedAt,
	}
}

// TeamToModel converts a domain entity to the persistence model.
func TeamToModel(t domain.Team) Team {
	return Team{
		ID:        t.ID,
		Name:      t.Name,
		CreatedBy: t.CreatedBy,
		CreatedAt: t.CreatedAt,
	}
}

// TeamMember is the persistence representation of domain.TeamMember.
type TeamMember struct {
	TeamID   int64     `db:"team_id"`
	UserID   int64     `db:"user_id"`
	Role     string    `db:"role"`
	JoinedAt time.Time `db:"joined_at"`
}

// ToDomain converts the persistence model to a domain entity.
func (m TeamMember) ToDomain() domain.TeamMember {
	return domain.TeamMember{
		TeamID:   m.TeamID,
		UserID:   m.UserID,
		Role:     domain.TeamRole(m.Role),
		JoinedAt: m.JoinedAt,
	}
}

// TeamMemberToModel converts a domain entity to the persistence model.
func TeamMemberToModel(m domain.TeamMember) TeamMember {
	return TeamMember{
		TeamID:   m.TeamID,
		UserID:   m.UserID,
		Role:     string(m.Role),
		JoinedAt: m.JoinedAt,
	}
}
