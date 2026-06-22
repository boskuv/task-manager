package model

import (
	"time"

	"github.com/boskuv/task-manager/internal/domain"
)

// User is the persistence representation of domain.User.
type User struct {
	ID           int64     `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

// ToDomain converts the persistence model to a domain entity.
func (u User) ToDomain() domain.User {
	return domain.User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
	}
}

// UserToModel converts a domain entity to the persistence model.
func UserToModel(u domain.User) User {
	return User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
	}
}
