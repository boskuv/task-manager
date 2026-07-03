package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/dto"
	"github.com/boskuv/task-manager/internal/repository"
	taskuc "github.com/boskuv/task-manager/internal/usecase/task"
)

const (
	defaultTaskPage     = 1
	defaultTaskPageSize = 20
	maxTaskPageSize     = 100
)

type taskService interface {
	Create(ctx context.Context, userID int64, input taskuc.CreateInput) (domain.Task, error)
	Update(ctx context.Context, userID, taskID int64, input taskuc.UpdateInput) (domain.Task, error)
	List(ctx context.Context, userID int64, input taskuc.ListInput) (repository.TaskListResult, error)
	ListHistory(ctx context.Context, userID, taskID int64) ([]domain.TaskHistory, error)
}

// TaskHandler serves task endpoints.
type TaskHandler struct {
	tasks taskService
}

// NewTaskHandler creates a task HTTP handler.
func NewTaskHandler(tasks taskService) *TaskHandler {
	return &TaskHandler{tasks: tasks}
}

// Create handles POST /api/v1/tasks.
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	task, err := h.tasks.Create(r.Context(), userID, taskuc.CreateInput{
		TeamID:      req.TeamID,
		Title:       req.Title,
		Description: req.Description,
		AssigneeID:  req.AssigneeID,
	})
	if err != nil {
		writeServiceError(w, r.Context(), err)
		return
	}

	writeJSON(w, http.StatusCreated, taskToResponse(task))
}

// List handles GET /api/v1/tasks.
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	input, err := parseTaskListQuery(r)
	if err != nil {
		writeServiceError(w, r.Context(), err)
		return
	}

	result, err := h.tasks.List(r.Context(), userID, input)
	if err != nil {
		writeServiceError(w, r.Context(), err)
		return
	}

	page, pageSize := normalizeTaskPagination(input.Page, input.PageSize)

	items := make([]dto.TaskResponse, 0, len(result.Items))
	for _, task := range result.Items {
		items = append(items, taskToResponse(task))
	}

	writeJSON(w, http.StatusOK, dto.ListResponse[dto.TaskResponse]{
		Items: items,
		Pagination: dto.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    result.Total,
		},
	})
}

// Update handles PUT /api/v1/tasks/{id}.
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	taskID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || taskID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var req dto.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	task, err := h.tasks.Update(r.Context(), userID, taskID, taskuc.UpdateInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		AssigneeID:  req.AssigneeID,
	})
	if err != nil {
		writeServiceError(w, r.Context(), err)
		return
	}

	writeJSON(w, http.StatusOK, taskToResponse(task))
}

// History handles GET /api/v1/tasks/{id}/history.
func (h *TaskHandler) History(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	taskID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || taskID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	entries, err := h.tasks.ListHistory(r.Context(), userID, taskID)
	if err != nil {
		writeServiceError(w, r.Context(), err)
		return
	}

	resp := make([]dto.TaskHistoryResponse, 0, len(entries))
	for _, entry := range entries {
		resp = append(resp, taskHistoryToResponse(entry))
	}

	writeJSON(w, http.StatusOK, resp)
}

func parseTaskListQuery(r *http.Request) (taskuc.ListInput, error) {
	query := r.URL.Query()

	teamID, err := strconv.ParseInt(query.Get("team_id"), 10, 64)
	if err != nil || teamID <= 0 {
		return taskuc.ListInput{}, ClientError("invalid team_id")
	}

	input := taskuc.ListInput{
		TeamID: teamID,
		Status: query.Get("status"),
	}

	if raw := query.Get("assignee_id"); raw != "" {
		assigneeID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || assigneeID <= 0 {
			return taskuc.ListInput{}, ClientError("invalid assignee_id")
		}
		input.AssigneeID = &assigneeID
	}

	if raw := query.Get("page"); raw != "" {
		page, err := strconv.Atoi(raw)
		if err != nil || page < 0 {
			return taskuc.ListInput{}, ClientError("invalid page")
		}
		input.Page = page
	}

	if raw := query.Get("page_size"); raw != "" {
		pageSize, err := strconv.Atoi(raw)
		if err != nil || pageSize < 0 {
			return taskuc.ListInput{}, ClientError("invalid page_size")
		}
		input.PageSize = pageSize
	}

	return input, nil
}

func normalizeTaskPagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = defaultTaskPage
	}
	if pageSize < 1 {
		pageSize = defaultTaskPageSize
	}
	if pageSize > maxTaskPageSize {
		pageSize = maxTaskPageSize
	}
	return page, pageSize
}

func taskToResponse(task domain.Task) dto.TaskResponse {
	resp := dto.TaskResponse{
		ID:          task.ID,
		TeamID:      task.TeamID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		AssigneeID:  task.AssigneeID,
		CreatedBy:   task.CreatedBy,
		CreatedAt:   task.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   task.UpdatedAt.UTC().Format(time.RFC3339),
	}
	return resp
}

func taskHistoryToResponse(entry domain.TaskHistory) dto.TaskHistoryResponse {
	return dto.TaskHistoryResponse{
		ID:        entry.ID,
		TaskID:    entry.TaskID,
		ChangedBy: entry.ChangedBy,
		Field:     entry.Field,
		OldValue:  entry.OldValue,
		NewValue:  entry.NewValue,
		CreatedAt: entry.CreatedAt.UTC().Format(time.RFC3339),
	}
}
