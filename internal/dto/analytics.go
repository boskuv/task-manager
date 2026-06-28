package dto

// TeamMemberDoneStatsResponse is per-member done-task stats for a team.
type TeamMemberDoneStatsResponse struct {
	UserID      int64  `json:"user_id"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	DoneTasks7d int    `json:"done_tasks_7d"`
}

// TeamStatsResponse aggregates team membership and done tasks in the last 7 days.
type TeamStatsResponse struct {
	TeamID      int64                         `json:"team_id"`
	Name        string                        `json:"name"`
	MemberCount int                           `json:"member_count"`
	DoneTasks7d int                           `json:"done_tasks_7d"`
	Members     []TeamMemberDoneStatsResponse `json:"members"`
}

// TeamTopCreatorResponse is a ranked task creator within a team for the last month.
type TeamTopCreatorResponse struct {
	TeamID       int64  `json:"team_id"`
	TeamName     string `json:"team_name"`
	UserID       int64  `json:"user_id"`
	Email        string `json:"email"`
	TasksCreated int    `json:"tasks_created"`
	Rank         int    `json:"rank"`
}
