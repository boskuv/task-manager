package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/dto"
	"github.com/boskuv/task-manager/internal/repository"
	taskuc "github.com/boskuv/task-manager/internal/usecase/task"
)

func TestTaskHandlerCreate(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC)
	handler := NewTaskHandler(&stubTaskService{
		createTask: domain.Task{
			ID:          1,
			TeamID:      3,
			Title:       "Fix bug",
			Description: "details",
			Status:      domain.TaskStatusTodo,
			CreatedBy:   42,
			CreatedAt:   createdAt,
			UpdatedAt:   createdAt,
		},
	})

	body := bytes.NewBufferString(`{"team_id":3,"title":"Fix bug","description":"details"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", body)
	req = req.WithContext(ContextWithUserID(req.Context(), 42))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var resp dto.TaskResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID != 1 || resp.TeamID != 3 || resp.Title != "Fix bug" || resp.Status != "todo" {
		t.Fatalf("response = %+v, want id=1 team_id=3 title=Fix bug status=todo", resp)
	}
}

func TestTaskHandlerCreateUnauthorized(t *testing.T) {
	t.Parallel()

	handler := NewTaskHandler(&stubTaskService{})

	body := bytes.NewBufferString(`{"team_id":3,"title":"Fix bug"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", body)
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestTaskHandlerList(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 26, 10, 0, 0, 0, time.UTC)
	handler := NewTaskHandler(&stubTaskService{
		listResult: repository.TaskListResult{
			Items: []domain.Task{
				{ID: 1, TeamID: 3, Title: "A", Status: domain.TaskStatusTodo, CreatedBy: 42, CreatedAt: createdAt, UpdatedAt: createdAt},
				{ID: 2, TeamID: 3, Title: "B", Status: domain.TaskStatusDone, CreatedBy: 42, CreatedAt: createdAt, UpdatedAt: createdAt},
			},
			Total: 2,
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?team_id=3&page=1&page_size=10", nil)
	req = req.WithContext(ContextWithUserID(req.Context(), 42))
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp dto.ListResponse[dto.TaskResponse]
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("items count = %d, want 2", len(resp.Items))
	}
	if resp.Pagination.Total != 2 || resp.Pagination.Page != 1 || resp.Pagination.PageSize != 10 {
		t.Fatalf("pagination = %+v, want page=1 page_size=10 total=2", resp.Pagination)
	}
}

func TestTaskHandlerListInvalidTeamID(t *testing.T) {
	t.Parallel()

	handler := NewTaskHandler(&stubTaskService{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?team_id=bad", nil)
	req = req.WithContext(ContextWithUserID(req.Context(), 42))
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestTaskHandlerUpdate(t *testing.T) {
	t.Parallel()

	updatedAt := time.Date(2026, 6, 26, 11, 0, 0, 0, time.UTC)
	handler := NewTaskHandler(&stubTaskService{
		updateTask: domain.Task{
			ID:          5,
			TeamID:      3,
			Title:       "Updated",
			Description: "desc",
			Status:      domain.TaskStatusInProgress,
			CreatedBy:   42,
			UpdatedAt:   updatedAt,
		},
	})

	body := bytes.NewBufferString(`{"title":"Updated","status":"in_progress"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/5", body)
	req.SetPathValue("id", "5")
	req = req.WithContext(ContextWithUserID(req.Context(), 42))
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp dto.TaskResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID != 5 || resp.Title != "Updated" || resp.Status != "in_progress" {
		t.Fatalf("response = %+v, want id=5 title=Updated status=in_progress", resp)
	}
}

func TestTaskHandlerUpdateForbidden(t *testing.T) {
	t.Parallel()

	handler := NewTaskHandler(&stubTaskService{
		updateErr: domain.ErrForbidden,
	})

	body := bytes.NewBufferString(`{"title":"Updated"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/5", body)
	req.SetPathValue("id", "5")
	req = req.WithContext(ContextWithUserID(req.Context(), 42))
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestTaskHandlerUpdateInvalidTaskID(t *testing.T) {
	t.Parallel()

	handler := NewTaskHandler(&stubTaskService{})

	body := bytes.NewBufferString(`{"title":"Updated"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/bad", body)
	req.SetPathValue("id", "bad")
	req = req.WithContext(ContextWithUserID(req.Context(), 42))
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

type stubTaskService struct {
	createTask domain.Task
	createErr  error
	updateTask domain.Task
	updateErr  error
	listResult repository.TaskListResult
	listErr    error
}

func (s *stubTaskService) Create(_ context.Context, userID int64, input taskuc.CreateInput) (domain.Task, error) {
	if s.createErr != nil {
		return domain.Task{}, s.createErr
	}
	task := s.createTask
	if task.Title == "" {
		task.Title = input.Title
	}
	if task.TeamID == 0 {
		task.TeamID = input.TeamID
	}
	if task.CreatedBy == 0 {
		task.CreatedBy = userID
	}
	return task, nil
}

func (s *stubTaskService) Update(_ context.Context, _, taskID int64, _ taskuc.UpdateInput) (domain.Task, error) {
	if s.updateErr != nil {
		return domain.Task{}, s.updateErr
	}
	task := s.updateTask
	if task.ID == 0 {
		task.ID = taskID
	}
	return task, nil
}

func (s *stubTaskService) List(_ context.Context, _ int64, _ taskuc.ListInput) (repository.TaskListResult, error) {
	if s.listErr != nil {
		return repository.TaskListResult{}, s.listErr
	}
	return s.listResult, nil
}
