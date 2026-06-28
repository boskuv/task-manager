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

// TaskFilter holds list criteria for tasks.
type TaskFilter struct {
	TeamID     int64
	Status     *domain.TaskStatus
	AssigneeID *int64
	Page       int
	PageSize   int
}

// TaskListResult is a paginated task list.
type TaskListResult struct {
	Items []domain.Task
	Total int
}

// TaskRepository persists and loads tasks.
type TaskRepository interface {
	Create(ctx context.Context, task domain.Task) (domain.Task, error)
	Update(ctx context.Context, task domain.Task) (domain.Task, error)
	GetByID(ctx context.Context, id int64) (domain.Task, error)
	List(ctx context.Context, filter TaskFilter) (TaskListResult, error)
	// HasOrphanAssignee reports whether the task assignee is set but not a team member.
	HasOrphanAssignee(ctx context.Context, taskID int64) (bool, error)
	// ListOrphanAssignees returns tasks whose assignee is outside the task team.
	ListOrphanAssignees(ctx context.Context) ([]domain.Task, error)
}

// TaskHistoryRepository persists task change audit records.
type TaskHistoryRepository interface {
	Insert(ctx context.Context, entry domain.TaskHistory) (domain.TaskHistory, error)
	ListByTaskID(ctx context.Context, taskID int64) ([]domain.TaskHistory, error)
}

// AnalyticsRepository runs aggregate reporting queries.
type AnalyticsRepository interface {
	// ListTeamStats returns each team with member count, done tasks in the last 7 days, and per-member breakdown.
	ListTeamStats(ctx context.Context) ([]domain.TeamStats, error)
}
