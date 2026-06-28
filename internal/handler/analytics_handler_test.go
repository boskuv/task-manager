package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/dto"
)

func TestAnalyticsHandlerListTeamStats(t *testing.T) {
	t.Parallel()

	handler := NewAnalyticsHandler(&stubAnalyticsService{
		teamStats: []domain.TeamStats{
			{
				TeamID:      1,
				Name:        "Backend",
				MemberCount: 2,
				DoneTasks7d: 5,
				Members: []domain.TeamMemberDoneStats{
					{UserID: 10, Email: "alice@example.com", Role: domain.TeamRoleOwner, DoneTasks7d: 3},
					{UserID: 11, Email: "bob@example.com", Role: domain.TeamRoleMember, DoneTasks7d: 2},
				},
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/teams/stats", nil)
	req = req.WithContext(ContextWithUserID(req.Context(), 42))
	rec := httptest.NewRecorder()

	handler.ListTeamStats(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp []dto.TeamStatsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("len(resp) = %d, want 1", len(resp))
	}
	if resp[0].TeamID != 1 || resp[0].Name != "Backend" || resp[0].MemberCount != 2 || resp[0].DoneTasks7d != 5 {
		t.Fatalf("response = %+v, want team stats for Backend", resp[0])
	}
	if len(resp[0].Members) != 2 || resp[0].Members[0].Email != "alice@example.com" {
		t.Fatalf("members = %+v, want 2 members", resp[0].Members)
	}
}

func TestAnalyticsHandlerListTeamStatsUnauthorized(t *testing.T) {
	t.Parallel()

	handler := NewAnalyticsHandler(&stubAnalyticsService{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/teams/stats", nil)
	rec := httptest.NewRecorder()

	handler.ListTeamStats(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAnalyticsHandlerListTopCreators(t *testing.T) {
	t.Parallel()

	handler := NewAnalyticsHandler(&stubAnalyticsService{
		topCreators: []domain.TeamTopCreator{
			{
				TeamID:       1,
				TeamName:     "Backend",
				UserID:       10,
				Email:        "alice@example.com",
				TasksCreated: 12,
				Rank:         1,
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/top-creators", nil)
	req = req.WithContext(ContextWithUserID(req.Context(), 42))
	rec := httptest.NewRecorder()

	handler.ListTopCreators(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp []dto.TeamTopCreatorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("len(resp) = %d, want 1", len(resp))
	}
	if resp[0].TeamID != 1 || resp[0].Rank != 1 || resp[0].TasksCreated != 12 {
		t.Fatalf("response = %+v, want top creator for team 1", resp[0])
	}
}

type stubAnalyticsService struct {
	teamStats   []domain.TeamStats
	teamStatsErr error
	topCreators []domain.TeamTopCreator
	topCreatorsErr error
}

func (s *stubAnalyticsService) ListTeamStats(_ context.Context, _ int64) ([]domain.TeamStats, error) {
	if s.teamStatsErr != nil {
		return nil, s.teamStatsErr
	}
	return s.teamStats, nil
}

func (s *stubAnalyticsService) ListTopCreators(_ context.Context, _ int64) ([]domain.TeamTopCreator, error) {
	if s.topCreatorsErr != nil {
		return nil, s.topCreatorsErr
	}
	return s.topCreators, nil
}
