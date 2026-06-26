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

const (
	defaultTaskPageSize = 20
	maxTaskPageSize     = 100
)

// List returns a paginated task list matching the given filters, ordered by creation time descending.
func (r *TaskRepo) List(ctx context.Context, filter repository.TaskFilter) (repository.TaskListResult, error) {
	where, args := taskListWhere(filter)

	var total int
	countQuery := `SELECT COUNT(*) FROM tasks ` + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return repository.TaskListResult{}, fmt.Errorf("count tasks: %w", err)
	}

	page, pageSize := normalizeTaskPagination(filter.Page, filter.PageSize)
	offset := (page - 1) * pageSize

	listArgs := append(append([]any{}, args...), pageSize, offset)
	query := `SELECT ` + taskSelectColumns + ` FROM tasks ` + where + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return repository.TaskListResult{}, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]domain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return repository.TaskListResult{}, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return repository.TaskListResult{}, fmt.Errorf("iterate tasks: %w", err)
	}

	return repository.TaskListResult{Items: tasks, Total: total}, nil
}

func taskListWhere(filter repository.TaskFilter) (string, []any) {
	where := `WHERE team_id = ?`
	args := []any{filter.TeamID}

	if filter.Status != nil {
		where += ` AND status = ?`
		args = append(args, string(*filter.Status))
	}
	if filter.AssigneeID != nil {
		where += ` AND assignee_id = ?`
		args = append(args, *filter.AssigneeID)
	}

	return where, args
}

func normalizeTaskPagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultTaskPageSize
	}
	if pageSize > maxTaskPageSize {
		pageSize = maxTaskPageSize
	}
	return page, pageSize
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
