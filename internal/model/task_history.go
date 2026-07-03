package model

import (
	"time"

	"github.com/boskuv/task-manager/internal/domain"
)

// TaskHistory is the persistence representation of domain.TaskHistory.
type TaskHistory struct {
	ID        int64     `db:"id"`
	TaskID    int64     `db:"task_id"`
	ChangedBy int64     `db:"changed_by"`
	Field     string    `db:"field"`
	OldValue  string    `db:"old_value"`
	NewValue  string    `db:"new_value"`
	CreatedAt time.Time `db:"created_at"`
}

// ToDomain converts the persistence model to a domain entity.
func (h TaskHistory) ToDomain() domain.TaskHistory {
	return domain.TaskHistory{
		ID:        h.ID,
		TaskID:    h.TaskID,
		ChangedBy: h.ChangedBy,
		Field:     h.Field,
		OldValue:  h.OldValue,
		NewValue:  h.NewValue,
		CreatedAt: h.CreatedAt,
	}
}

// TaskHistoryToModel converts a domain entity to the persistence model.
func TaskHistoryToModel(h domain.TaskHistory) TaskHistory {
	return TaskHistory{
		ID:        h.ID,
		TaskID:    h.TaskID,
		ChangedBy: h.ChangedBy,
		Field:     h.Field,
		OldValue:  h.OldValue,
		NewValue:  h.NewValue,
		CreatedAt: h.CreatedAt,
	}
}
