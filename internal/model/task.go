package model

import (
	"time"

	"github.com/boskuv/task-manager/internal/domain"
)

// Task is the persistence representation of domain.Task.
type Task struct {
	ID          int64     `db:"id"`
	TeamID      int64     `db:"team_id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	Status      string    `db:"status"`
	AssigneeID  *int64    `db:"assignee_id"`
	CreatedBy   int64     `db:"created_by"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// ToDomain converts the persistence model to a domain entity.
func (t Task) ToDomain() domain.Task {
	return domain.Task{
		ID:          t.ID,
		TeamID:      t.TeamID,
		Title:       t.Title,
		Description: t.Description,
		Status:      domain.TaskStatus(t.Status),
		AssigneeID:  t.AssigneeID,
		CreatedBy:   t.CreatedBy,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// TaskToModel converts a domain entity to the persistence model.
func TaskToModel(t domain.Task) Task {
	return Task{
		ID:          t.ID,
		TeamID:      t.TeamID,
		Title:       t.Title,
		Description: t.Description,
		Status:      string(t.Status),
		AssigneeID:  t.AssigneeID,
		CreatedBy:   t.CreatedBy,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
