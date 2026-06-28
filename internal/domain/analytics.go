package domain

// TeamMemberDoneStats holds per-member done-task counts for analytics.
type TeamMemberDoneStats struct {
	UserID      int64
	Email       string
	Role        TeamRole
	DoneTasks7d int
}

// TeamStats aggregates team membership and completed tasks in the last 7 days.
type TeamStats struct {
	TeamID      int64
	Name        string
	MemberCount int
	DoneTasks7d int
	Members     []TeamMemberDoneStats
}
