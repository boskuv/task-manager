package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/model"
	"github.com/boskuv/task-manager/internal/repository"
)

var _ repository.TaskRepository = (*TaskRepo)(nil)

const taskSelectColumns = `id, team_id, title, description, status, assignee_id, created_by, created_at, updated_at`

// TaskRepo implements repository.TaskRepository with MySQL.
type TaskRepo struct {
	db *sql.DB
}

// NewTaskRepo returns a MySQL-backed task repository.
func NewTaskRepo(db *sql.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

// Create inserts a new task and returns the persisted entity.
func (r *TaskRepo) Create(ctx context.Context, task domain.Task) (domain.Task, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO tasks (team_id, title, description, status, assignee_id, created_by)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		task.TeamID,
		task.Title,
		task.Description,
		string(task.Status),
		task.AssigneeID,
		task.CreatedBy,
	)
	if err != nil {
		return domain.Task{}, fmt.Errorf("insert task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.Task{}, fmt.Errorf("last insert id: %w", err)
	}

	return r.GetByID(ctx, id)
}

// Update replaces mutable task fields and returns the persisted entity.
func (r *TaskRepo) Update(ctx context.Context, task domain.Task) (domain.Task, error) {
	result, err := r.db.ExecContext(ctx,
		`UPDATE tasks
		 SET title = ?, description = ?, status = ?, assignee_id = ?
		 WHERE id = ?`,
		task.Title,
		task.Description,
		string(task.Status),
		task.AssigneeID,
		task.ID,
	)
	if err != nil {
		return domain.Task{}, fmt.Errorf("update task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return domain.Task{}, fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return domain.Task{}, domain.ErrNotFound
	}

	return r.GetByID(ctx, task.ID)
}

// GetByID returns a task by id.
func (r *TaskRepo) GetByID(ctx context.Context, id int64) (domain.Task, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+taskSelectColumns+` FROM tasks WHERE id = ?`,
		id,
	)
	return scanTask(row)
}

// List returns tasks matching the given filters, ordered by creation time descending.
func (r *TaskRepo) List(ctx context.Context, filter repository.TaskFilter) ([]domain.Task, error) {
	query := `SELECT ` + taskSelectColumns + ` FROM tasks WHERE team_id = ?`
	args := []any{filter.TeamID}

	if filter.Status != nil {
		query += ` AND status = ?`
		args = append(args, string(*filter.Status))
	}
	if filter.AssigneeID != nil {
		query += ` AND assignee_id = ?`
		args = append(args, *filter.AssigneeID)
	}

	query += ` ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]domain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, nil
}

func scanTask(row rowScanner) (domain.Task, error) {
	var m model.Task
	if err := row.Scan(
		&m.ID,
		&m.TeamID,
		&m.Title,
		&m.Description,
		&m.Status,
		&m.AssigneeID,
		&m.CreatedBy,
		&m.CreatedAt,
		&m.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Task{}, domain.ErrNotFound
		}
		return domain.Task{}, fmt.Errorf("scan task: %w", err)
	}
	return m.ToDomain(), nil
}
