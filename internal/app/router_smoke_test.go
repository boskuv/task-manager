package app

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/dto"
	"github.com/boskuv/task-manager/internal/handler"
	"github.com/boskuv/task-manager/internal/handler/middleware"
	jwtmanager "github.com/boskuv/task-manager/internal/pkg/jwt"
	"github.com/boskuv/task-manager/internal/repository"
	taskuc "github.com/boskuv/task-manager/internal/usecase/task"
)

func TestRouterSmokeMainEndpoints(t *testing.T) {
	router := newSmokeRouter(t)

	assertStatus(t, router, "GET", "/health", "", "", http.StatusOK)
	if body := smokeRequest(t, router, "GET", "/health", "", ""); body != "ok" {
		t.Fatalf("health body = %q, want ok", body)
	}

	registerBody := `{"email":"smoke@example.com","password":"password1"}`
	registerResp := smokeRequestJSON[dto.AuthResponse](t, router, "POST", "/api/v1/register", registerBody, "")
	if registerResp.Token == "" {
		t.Fatal("expected register token")
	}

	loginBody := `{"email":"smoke@example.com","password":"password1"}`
	loginResp := smokeRequestJSON[dto.AuthResponse](t, router, "POST", "/api/v1/login", loginBody, "")
	if loginResp.Token == "" {
		t.Fatal("expected login token")
	}

	token := registerResp.Token
	me := smokeRequestJSON[map[string]int64](t, router, "GET", "/api/v1/me", "", token)
	if me["user_id"] != 42 {
		t.Fatalf("me user_id = %d, want 42", me["user_id"])
	}

	team := smokeRequestJSON[dto.TeamResponse](t, router, "POST", "/api/v1/teams", `{"name":"Smoke Team"}`, token)
	if team.ID != 1 || team.Name != "Smoke Team" {
		t.Fatalf("team = %+v", team)
	}

	teams := smokeRequestJSON[[]dto.TeamResponse](t, router, "GET", "/api/v1/teams", "", token)
	if len(teams) != 1 || teams[0].ID != 1 {
		t.Fatalf("teams = %+v", teams)
	}

	member := smokeRequestJSON[dto.TeamMemberResponse](t, router, "POST", "/api/v1/teams/1/invite",
		`{"email":"invitee@example.com","role":"member"}`, token)
	if member.TeamID != 1 || member.UserID != 7 {
		t.Fatalf("member = %+v", member)
	}

	task := smokeRequestJSON[dto.TaskResponse](t, router, "POST", "/api/v1/tasks",
		`{"team_id":1,"title":"Smoke task","description":"details"}`, token)
	if task.ID != 10 || task.Title != "Smoke task" {
		t.Fatalf("task = %+v", task)
	}

	taskList := smokeRequestJSON[dto.ListResponse[dto.TaskResponse]](t, router, "GET",
		"/api/v1/tasks?team_id=1&page=1&page_size=10", "", token)
	if taskList.Pagination.Total != 1 || len(taskList.Items) != 1 {
		t.Fatalf("task list = %+v", taskList)
	}

	updated := smokeRequestJSON[dto.TaskResponse](t, router, "PUT", "/api/v1/tasks/10",
		`{"title":"Updated task","status":"in_progress"}`, token)
	if updated.Title != "Updated task" || updated.Status != "in_progress" {
		t.Fatalf("updated task = %+v", updated)
	}

	history := smokeRequestJSON[[]dto.TaskHistoryResponse](t, router, "GET", "/api/v1/tasks/10/history", "", token)
	if len(history) != 1 || history[0].Field != "status" {
		t.Fatalf("history = %+v", history)
	}

	stats := smokeRequestJSON[[]dto.TeamStatsResponse](t, router, "GET", "/api/v1/analytics/teams/stats", "", token)
	if len(stats) != 1 || stats[0].TeamID != 1 {
		t.Fatalf("team stats = %+v", stats)
	}

	topCreators := smokeRequestJSON[[]dto.TeamTopCreatorResponse](t, router, "GET", "/api/v1/analytics/top-creators", "", token)
	if len(topCreators) != 1 || topCreators[0].Rank != 1 {
		t.Fatalf("top creators = %+v", topCreators)
	}

	assertStatus(t, router, "GET", "/api/v1/teams", "", "", http.StatusUnauthorized)
}

func newSmokeRouter(t *testing.T) http.Handler {
	t.Helper()

	jwtMgr := jwtmanager.NewManager("smoke-test-secret", time.Hour)
	now := time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC)

	return newRouter(routerDeps{
		auth:               handler.NewAuthHandler(&smokeAuthService{jwt: jwtMgr, userID: 42}),
		teams:              handler.NewTeamHandler(&smokeTeamService{now: now}),
		tasks:              handler.NewTaskHandler(&smokeTaskService{now: now}),
		analytics:          handler.NewAnalyticsHandler(&smokeAnalyticsService{}),
		jwtManager:         jwtMgr,
		rateLimiter:        &smokeRateLimiter{},
		rateLimitPerMinute: 100,
		metrics:            middleware.NewHTTPMetrics(),
	})
}

func smokeRequest(t *testing.T, router http.Handler, method, path, body, token string) string {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	router.ServeHTTP(rec, req)
	if rec.Code < 200 || rec.Code >= 300 {
		t.Fatalf("%s %s: status = %d body = %s", method, path, rec.Code, rec.Body.String())
	}
	data, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(data)
}

func smokeRequestJSON[T any](t *testing.T, router http.Handler, method, path, body, token string) T {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	router.ServeHTTP(rec, req)
	if rec.Code < 200 || rec.Code >= 300 {
		t.Fatalf("%s %s: status = %d body = %s", method, path, rec.Code, rec.Body.String())
	}
	var out T
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode %s %s: %v", method, path, err)
	}
	return out
}

func assertStatus(t *testing.T, router http.Handler, method, path, body, token string, want int) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	router.ServeHTTP(rec, req)
	if rec.Code != want {
		t.Fatalf("%s %s: status = %d, want %d body = %s", method, path, rec.Code, want, rec.Body.String())
	}
}

type smokeRateLimiter struct{}

func (s *smokeRateLimiter) Allow(context.Context, string) (bool, error) {
	return true, nil
}

type smokeAuthService struct {
	jwt    *jwtmanager.Manager
	userID int64
}

func (s *smokeAuthService) Register(_ context.Context, _, _ string) (string, error) {
	return s.jwt.GenerateAccessToken(s.userID)
}

func (s *smokeAuthService) Login(_ context.Context, _, _ string) (string, error) {
	return s.jwt.GenerateAccessToken(s.userID)
}

type smokeTeamService struct {
	now time.Time
}

func (s *smokeTeamService) Create(_ context.Context, userID int64, name string) (domain.Team, error) {
	return domain.Team{
		ID:        1,
		Name:      name,
		CreatedBy: userID,
		CreatedAt: s.now,
	}, nil
}

func (s *smokeTeamService) List(_ context.Context, _ int64) ([]domain.Team, error) {
	return []domain.Team{{
		ID:        1,
		Name:      "Smoke Team",
		CreatedBy: 42,
		CreatedAt: s.now,
	}}, nil
}

func (s *smokeTeamService) Invite(_ context.Context, _, teamID int64, _, role string) (domain.TeamMember, error) {
	return domain.TeamMember{
		TeamID:   teamID,
		UserID:   7,
		Role:     domain.TeamRole(role),
		JoinedAt: s.now,
	}, nil
}

type smokeTaskService struct {
	now time.Time
}

func (s *smokeTaskService) Create(_ context.Context, userID int64, input taskuc.CreateInput) (domain.Task, error) {
	return domain.Task{
		ID:          10,
		TeamID:      input.TeamID,
		Title:       input.Title,
		Description: input.Description,
		Status:      domain.TaskStatusTodo,
		CreatedBy:   userID,
		CreatedAt:   s.now,
		UpdatedAt:   s.now,
	}, nil
}

func (s *smokeTaskService) Update(_ context.Context, _, taskID int64, input taskuc.UpdateInput) (domain.Task, error) {
	task := domain.Task{
		ID:        taskID,
		TeamID:    1,
		Title:     "Smoke task",
		Status:    domain.TaskStatusTodo,
		CreatedBy: 42,
		UpdatedAt: s.now,
	}
	if input.Title != nil {
		task.Title = *input.Title
	}
	if input.Status != nil {
		task.Status = domain.TaskStatus(*input.Status)
	}
	return task, nil
}

func (s *smokeTaskService) List(_ context.Context, _ int64, _ taskuc.ListInput) (repository.TaskListResult, error) {
	return repository.TaskListResult{
		Items: []domain.Task{{
			ID:        10,
			TeamID:    1,
			Title:     "Smoke task",
			Status:    domain.TaskStatusTodo,
			CreatedBy: 42,
			CreatedAt: s.now,
			UpdatedAt: s.now,
		}},
		Total: 1,
	}, nil
}

func (s *smokeTaskService) ListHistory(_ context.Context, _, taskID int64) ([]domain.TaskHistory, error) {
	return []domain.TaskHistory{{
		ID:        1,
		TaskID:    taskID,
		ChangedBy: 42,
		Field:     domain.HistoryFieldStatus,
		OldValue:  "todo",
		NewValue:  "in_progress",
		CreatedAt: s.now,
	}}, nil
}

type smokeAnalyticsService struct{}

func (s *smokeAnalyticsService) ListTeamStats(_ context.Context, _ int64) ([]domain.TeamStats, error) {
	return []domain.TeamStats{{
		TeamID:      1,
		Name:        "Smoke Team",
		MemberCount: 2,
		DoneTasks7d: 3,
	}}, nil
}

func (s *smokeAnalyticsService) ListTopCreators(_ context.Context, _ int64) ([]domain.TeamTopCreator, error) {
	return []domain.TeamTopCreator{{
		TeamID:       1,
		TeamName:     "Smoke Team",
		UserID:       42,
		Email:        "smoke@example.com",
		TasksCreated: 5,
		Rank:         1,
	}}, nil
}
