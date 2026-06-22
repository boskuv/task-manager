package domain

import "time"

// Task history field names used when recording audit diffs.
const (
	HistoryFieldStatus      = "status"
	HistoryFieldAssignee    = "assignee"
	HistoryFieldTitle       = "title"
	HistoryFieldDescription = "description"
)

// TaskHistory records a single field change on a task.
type TaskHistory struct {
	ID        int64
	TaskID    int64
	ChangedBy int64
	Field     string
	OldValue  string
	NewValue  string
	CreatedAt time.Time
}
