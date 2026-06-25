package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/dto"
)

type teamService interface {
	Create(ctx context.Context, userID int64, name string) (domain.Team, error)
	List(ctx context.Context, userID int64) ([]domain.Team, error)
}

// TeamHandler serves team endpoints.
type TeamHandler struct {
	teams teamService
}

// NewTeamHandler creates a team HTTP handler.
func NewTeamHandler(teams teamService) *TeamHandler {
	return &TeamHandler{teams: teams}
}

// Create handles POST /api/v1/teams.
func (h *TeamHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	team, err := h.teams.Create(r.Context(), userID, req.Name)
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusCreated, teamToResponse(team))
}

// List handles GET /api/v1/teams.
func (h *TeamHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	teams, err := h.teams.List(r.Context(), userID)
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	resp := make([]dto.TeamResponse, 0, len(teams))
	for _, team := range teams {
		resp = append(resp, teamToResponse(team))
	}

	writeJSON(w, http.StatusOK, resp)
}

func teamToResponse(team domain.Team) dto.TeamResponse {
	return dto.TeamResponse{
		ID:        team.ID,
		Name:      team.Name,
		CreatedBy: team.CreatedBy,
		CreatedAt: team.CreatedAt.UTC().Format(time.RFC3339),
	}
}
