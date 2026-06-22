package domain

import "time"

// TeamRole defines a member's role inside a team.
type TeamRole string

const (
	TeamRoleOwner  TeamRole = "owner"
	TeamRoleAdmin  TeamRole = "admin"
	TeamRoleMember TeamRole = "member"
)

func (r TeamRole) IsValid() bool {
	switch r {
	case TeamRoleOwner, TeamRoleAdmin, TeamRoleMember:
		return true
	default:
		return false
	}
}

// CanInvite reports whether the role may invite new members.
func (r TeamRole) CanInvite() bool {
	return r == TeamRoleOwner || r == TeamRoleAdmin
}

// Team represents a collaborative workspace.
type Team struct {
	ID        int64
	Name      string
	CreatedBy int64
	CreatedAt time.Time
}

// TeamMember links a user to a team with a role.
type TeamMember struct {
	TeamID   int64
	UserID   int64
	Role     TeamRole
	JoinedAt time.Time
}
