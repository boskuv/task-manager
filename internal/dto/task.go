package dto

// CreateTaskRequest is the body for POST /api/v1/tasks.
type CreateTaskRequest struct {
	TeamID      int64  `json:"team_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	AssigneeID  *int64 `json:"assignee_id,omitempty"`
}

// UpdateTaskRequest is the body for PUT /api/v1/tasks/{id}.
type UpdateTaskRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	AssigneeID  *int64  `json:"assignee_id,omitempty"`
}

// TaskResponse is a task returned by the API.
type TaskResponse struct {
	ID          int64  `json:"id"`
	TeamID      int64  `json:"team_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	AssigneeID  *int64 `json:"assignee_id,omitempty"`
	CreatedBy   int64  `json:"created_by"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// TaskHistoryResponse is a single audit record for a task.
type TaskHistoryResponse struct {
	ID        int64  `json:"id"`
	TaskID    int64  `json:"task_id"`
	ChangedBy int64  `json:"changed_by"`
	Field     string `json:"field"`
	OldValue  string `json:"old_value"`
	NewValue  string `json:"new_value"`
	CreatedAt string `json:"created_at"`
}

// TaskListQuery holds GET /api/v1/tasks filter parameters.
type TaskListQuery struct {
	TeamID     int64  `json:"team_id"`
	Status     string `json:"status,omitempty"`
	AssigneeID *int64 `json:"assignee_id,omitempty"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
}
