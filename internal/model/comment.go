package model

import (
	"time"

	"github.com/boskuv/task-manager/internal/domain"
)

// Comment is the persistence representation of domain.Comment.
type Comment struct {
	ID        int64     `db:"id"`
	TaskID    int64     `db:"task_id"`
	UserID    int64     `db:"user_id"`
	Body      string    `db:"body"`
	CreatedAt time.Time `db:"created_at"`
}

// ToDomain converts the persistence model to a domain entity.
func (c Comment) ToDomain() domain.Comment {
	return domain.Comment{
		ID:        c.ID,
		TaskID:    c.TaskID,
		UserID:    c.UserID,
		Body:      c.Body,
		CreatedAt: c.CreatedAt,
	}
}

// CommentToModel converts a domain entity to the persistence model.
func CommentToModel(c domain.Comment) Comment {
	return Comment{
		ID:        c.ID,
		TaskID:    c.TaskID,
		UserID:    c.UserID,
		Body:      c.Body,
		CreatedAt: c.CreatedAt,
	}
}
