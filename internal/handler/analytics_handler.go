package handler

import (
	"context"
	"net/http"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/dto"
)

type analyticsService interface {
	ListTeamStats(ctx context.Context, userID int64) ([]domain.TeamStats, error)
	ListTopCreators(ctx context.Context, userID int64) ([]domain.TeamTopCreator, error)
}

// AnalyticsHandler serves analytics endpoints.
type AnalyticsHandler struct {
	analytics analyticsService
}

// NewAnalyticsHandler creates an analytics HTTP handler.
func NewAnalyticsHandler(analytics analyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analytics: analytics}
}

// ListTeamStats handles GET /api/v1/analytics/teams/stats.
func (h *AnalyticsHandler) ListTeamStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	stats, err := h.analytics.ListTeamStats(r.Context(), userID)
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	resp := make([]dto.TeamStatsResponse, 0, len(stats))
	for _, stat := range stats {
		resp = append(resp, teamStatsToResponse(stat))
	}

	writeJSON(w, http.StatusOK, resp)
}

// ListTopCreators handles GET /api/v1/analytics/top-creators.
func (h *AnalyticsHandler) ListTopCreators(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	creators, err := h.analytics.ListTopCreators(r.Context(), userID)
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	resp := make([]dto.TeamTopCreatorResponse, 0, len(creators))
	for _, creator := range creators {
		resp = append(resp, topCreatorToResponse(creator))
	}

	writeJSON(w, http.StatusOK, resp)
}

func teamStatsToResponse(stat domain.TeamStats) dto.TeamStatsResponse {
	members := make([]dto.TeamMemberDoneStatsResponse, 0, len(stat.Members))
	for _, member := range stat.Members {
		members = append(members, dto.TeamMemberDoneStatsResponse{
			UserID:      member.UserID,
			Email:       member.Email,
			Role:        string(member.Role),
			DoneTasks7d: member.DoneTasks7d,
		})
	}

	return dto.TeamStatsResponse{
		TeamID:      stat.TeamID,
		Name:        stat.Name,
		MemberCount: stat.MemberCount,
		DoneTasks7d: stat.DoneTasks7d,
		Members:     members,
	}
}

func topCreatorToResponse(creator domain.TeamTopCreator) dto.TeamTopCreatorResponse {
	return dto.TeamTopCreatorResponse{
		TeamID:       creator.TeamID,
		TeamName:     creator.TeamName,
		UserID:       creator.UserID,
		Email:        creator.Email,
		TasksCreated: creator.TasksCreated,
		Rank:         creator.Rank,
	}
}
