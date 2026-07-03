package domain

import "time"

// Comment is a user note attached to a task.
type Comment struct {
	ID        int64
	TaskID    int64
	UserID    int64
	Body      string
	CreatedAt time.Time
}
