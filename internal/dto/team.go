package dto

// CreateTeamRequest is the body for POST /api/v1/teams.
type CreateTeamRequest struct {
	Name string `json:"name"`
}

// InviteRequest is the body for POST /api/v1/teams/{id}/invite.
type InviteRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// TeamResponse is a team returned by the API.
type TeamResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedBy int64  `json:"created_by"`
	CreatedAt string `json:"created_at"`
}

// TeamMemberResponse describes a user's membership in a team.
type TeamMemberResponse struct {
	TeamID   int64  `json:"team_id"`
	UserID   int64  `json:"user_id"`
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at"`
}
