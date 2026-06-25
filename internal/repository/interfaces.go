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
