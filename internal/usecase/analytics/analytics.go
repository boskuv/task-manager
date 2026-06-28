package analytics

import (
	"context"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/repository"
)

// Service exposes analytics reports for teams the user belongs to.
type Service struct {
	analytics repository.AnalyticsRepository
	teams     repository.TeamRepository
}

// NewService creates an analytics use case service.
func NewService(analytics repository.AnalyticsRepository, teams repository.TeamRepository) *Service {
	return &Service{
		analytics: analytics,
		teams:     teams,
	}
}

// ListTeamStats returns team stats for teams where the user is a member.
func (s *Service) ListTeamStats(ctx context.Context, userID int64) ([]domain.TeamStats, error) {
	allowed, err := s.allowedTeamIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	stats, err := s.analytics.ListTeamStats(ctx)
	if err != nil {
		return nil, err
	}

	return filterTeamStats(stats, allowed), nil
}

// ListTopCreators returns top task creators for teams where the user is a member.
func (s *Service) ListTopCreators(ctx context.Context, userID int64) ([]domain.TeamTopCreator, error) {
	allowed, err := s.allowedTeamIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	creators, err := s.analytics.ListTopCreatorsPerTeam(ctx)
	if err != nil {
		return nil, err
	}

	return filterTopCreators(creators, allowed), nil
}

func (s *Service) allowedTeamIDs(ctx context.Context, userID int64) (map[int64]struct{}, error) {
	if userID <= 0 {
		return nil, domain.ErrInvalidInput
	}

	teams, err := s.teams.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	allowed := make(map[int64]struct{}, len(teams))
	for _, team := range teams {
		allowed[team.ID] = struct{}{}
	}
	return allowed, nil
}

func filterTeamStats(stats []domain.TeamStats, allowed map[int64]struct{}) []domain.TeamStats {
	filtered := make([]domain.TeamStats, 0, len(allowed))
	for _, stat := range stats {
		if _, ok := allowed[stat.TeamID]; !ok {
			continue
		}
		filtered = append(filtered, stat)
	}
	return filtered
}

func filterTopCreators(creators []domain.TeamTopCreator, allowed map[int64]struct{}) []domain.TeamTopCreator {
	filtered := make([]domain.TeamTopCreator, 0, len(creators))
	for _, creator := range creators {
		if _, ok := allowed[creator.TeamID]; !ok {
			continue
		}
		filtered = append(filtered, creator)
	}
	return filtered
}
