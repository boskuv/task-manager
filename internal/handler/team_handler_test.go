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
)

func TestTeamHandlerCreate(t *testing.T) {
	t.Parallel()

	handler := NewTeamHandler(&stubTeamService{
		createTeam: domain.Team{
			ID:        1,
			Name:      "Backend",
			CreatedBy: 42,
			CreatedAt: time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC),
		},
	})

	body := bytes.NewBufferString(`{"name":"Backend"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", body)
	req = req.WithContext(ContextWithUserID(req.Context(), 42))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var resp dto.TeamResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID != 1 || resp.Name != "Backend" || resp.CreatedBy != 42 {
		t.Fatalf("response = %+v, want id=1 name=Backend created_by=42", resp)
	}
	if resp.CreatedAt != "2026-06-25T12:00:00Z" {
		t.Errorf("created_at = %q, want RFC3339 timestamp", resp.CreatedAt)
	}
}

func TestTeamHandlerCreateUnauthorized(t *testing.T) {
	t.Parallel()

	handler := NewTeamHandler(&stubTeamService{})

	body := bytes.NewBufferString(`{"name":"Backend"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", body)
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestTeamHandlerList(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	handler := NewTeamHandler(&stubTeamService{
		listTeams: []domain.Team{
			{ID: 1, Name: "Backend", CreatedBy: 42, CreatedAt: createdAt},
			{ID: 2, Name: "Frontend", CreatedBy: 42, CreatedAt: createdAt},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)
	req = req.WithContext(ContextWithUserID(req.Context(), 42))
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp []dto.TeamResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("teams count = %d, want 2", len(resp))
	}
}

func TestTeamHandlerInvalidJSON(t *testing.T) {
	t.Parallel()

	handler := NewTeamHandler(&stubTeamService{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", bytes.NewBufferString("{"))
	req = req.WithContext(ContextWithUserID(req.Context(), 1))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

type stubTeamService struct {
	createTeam domain.Team
	createErr  error
	listTeams  []domain.Team
	listErr    error
}

func (s *stubTeamService) Create(_ context.Context, userID int64, name string) (domain.Team, error) {
	if s.createErr != nil {
		return domain.Team{}, s.createErr
	}
	team := s.createTeam
	if team.Name == "" {
		team.Name = name
	}
	if team.CreatedBy == 0 {
		team.CreatedBy = userID
	}
	return team, nil
}

func (s *stubTeamService) List(_ context.Context, _ int64) ([]domain.Team, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.listTeams, nil
}
