package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/model"
	"github.com/boskuv/task-manager/internal/repository"
)

var _ repository.UserRepository = (*UserRepo)(nil)

const userSelectColumns = `id, email, password_hash, created_at`

// UserRepo implements repository.UserRepository with MySQL.
type UserRepo struct {
	db *sql.DB
}

// NewUserRepo returns a MySQL-backed user repository.
func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create inserts a new user and returns the persisted entity.
func (r *UserRepo) Create(ctx context.Context, user domain.User) (domain.User, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO users (email, password_hash) VALUES (?, ?)`,
		user.Email,
		user.PasswordHash,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.User{}, domain.ErrConflict
		}
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.User{}, fmt.Errorf("last insert id: %w", err)
	}

	return r.GetByID(ctx, id)
}

// GetByEmail returns a user by email.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+userSelectColumns+` FROM users WHERE email = ?`,
		email,
	)
	return scanUser(row)
}

// GetByID returns a user by id.
func (r *UserRepo) GetByID(ctx context.Context, id int64) (domain.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+userSelectColumns+` FROM users WHERE id = ?`,
		id,
	)
	return scanUser(row)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner) (domain.User, error) {
	var m model.User
	if err := row.Scan(&m.ID, &m.Email, &m.PasswordHash, &m.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("scan user: %w", err)
	}
	return m.ToDomain(), nil
}

func isDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
