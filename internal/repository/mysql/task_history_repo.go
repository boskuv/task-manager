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

var _ repository.TaskHistoryRepository = (*TaskHistoryRepo)(nil)

const taskHistorySelectColumns = `id, task_id, changed_by, field, old_value, new_value, created_at`

// TaskHistoryRepo implements repository.TaskHistoryRepository with MySQL.
type TaskHistoryRepo struct {
	db *sql.DB
}

// NewTaskHistoryRepo returns a MySQL-backed task history repository.
func NewTaskHistoryRepo(db *sql.DB) *TaskHistoryRepo {
	return &TaskHistoryRepo{db: db}
}

// Insert records a single field change on a task.
func (r *TaskHistoryRepo) Insert(ctx context.Context, entry domain.TaskHistory) (domain.TaskHistory, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO task_history (task_id, changed_by, field, old_value, new_value)
		 VALUES (?, ?, ?, ?, ?)`,
		entry.TaskID,
		entry.ChangedBy,
		entry.Field,
		entry.OldValue,
		entry.NewValue,
	)
	if err != nil {
		return domain.TaskHistory{}, fmt.Errorf("insert task history: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.TaskHistory{}, fmt.Errorf("last insert id: %w", err)
	}

	return r.getByID(ctx, id)
}

// ListByTaskID returns audit records for a task in chronological order.
func (r *TaskHistoryRepo) ListByTaskID(ctx context.Context, taskID int64) ([]domain.TaskHistory, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+taskHistorySelectColumns+`
		 FROM task_history
		 WHERE task_id = ?
		 ORDER BY created_at ASC, id ASC`,
		taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("list task history: %w", err)
	}
	defer rows.Close()

	entries := make([]domain.TaskHistory, 0)
	for rows.Next() {
		entry, err := scanTaskHistory(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate task history: %w", err)
	}

	return entries, nil
}

func (r *TaskHistoryRepo) getByID(ctx context.Context, id int64) (domain.TaskHistory, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+taskHistorySelectColumns+` FROM task_history WHERE id = ?`,
		id,
	)
	return scanTaskHistory(row)
}

func scanTaskHistory(row rowScanner) (domain.TaskHistory, error) {
	var m model.TaskHistory
	if err := row.Scan(
		&m.ID,
		&m.TaskID,
		&m.ChangedBy,
		&m.Field,
		&m.OldValue,
		&m.NewValue,
		&m.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.TaskHistory{}, domain.ErrNotFound
		}
		return domain.TaskHistory{}, fmt.Errorf("scan task history: %w", err)
	}
	return m.ToDomain(), nil
}
