package analytics

import (
	"context"
	"errors"
	"testing"

	"github.com/boskuv/task-manager/internal/domain"
)

func TestListTeamStatsFiltersByMembership(t *testing.T) {
	t.Parallel()

	svc := NewService(&stubAnalyticsRepo{
		stats: []domain.TeamStats{
			{TeamID: 1, Name: "Allowed", MemberCount: 1, DoneTasks7d: 2},
			{TeamID: 2, Name: "Hidden", MemberCount: 1, DoneTasks7d: 5},
		},
	}, &stubTeamRepo{
		teams: []domain.Team{{ID: 1, Name: "Allowed"}},
	})

	stats, err := svc.ListTeamStats(context.Background(), 42)
	if err != nil {
		t.Fatalf("ListTeamStats: %v", err)
	}
	if len(stats) != 1 || stats[0].TeamID != 1 {
		t.Fatalf("stats = %+v, want team 1 only", stats)
	}
}

func TestListTopCreatorsFiltersByMembership(t *testing.T) {
	t.Parallel()

	svc := NewService(&stubAnalyticsRepo{
		creators: []domain.TeamTopCreator{
			{TeamID: 1, TeamName: "Allowed", UserID: 10, Rank: 1, TasksCreated: 3},
			{TeamID: 2, TeamName: "Hidden", UserID: 11, Rank: 1, TasksCreated: 9},
		},
	}, &stubTeamRepo{
		teams: []domain.Team{{ID: 1, Name: "Allowed"}},
	})

	creators, err := svc.ListTopCreators(context.Background(), 42)
	if err != nil {
		t.Fatalf("ListTopCreators: %v", err)
	}
	if len(creators) != 1 || creators[0].TeamID != 1 {
		t.Fatalf("creators = %+v, want team 1 only", creators)
	}
}

func TestListTeamStatsInvalidUserID(t *testing.T) {
	t.Parallel()

	svc := NewService(&stubAnalyticsRepo{}, &stubTeamRepo{})

	_, err := svc.ListTeamStats(context.Background(), 0)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
}

func TestListTeamStatsRepositoryError(t *testing.T) {
	t.Parallel()

	svc := NewService(
		&stubAnalyticsRepo{statsErr: errors.New("analytics down")},
		&stubTeamRepo{teams: []domain.Team{{ID: 1}}},
	)

	_, err := svc.ListTeamStats(context.Background(), 1)
	if err == nil || err.Error() != "analytics down" {
		t.Fatalf("error = %v, want analytics down", err)
	}
}

type stubAnalyticsRepo struct {
	stats      []domain.TeamStats
	statsErr   error
	creators   []domain.TeamTopCreator
	creatorsErr error
}

func (s *stubAnalyticsRepo) ListTeamStats(context.Context) ([]domain.TeamStats, error) {
	if s.statsErr != nil {
		return nil, s.statsErr
	}
	return s.stats, nil
}

func (s *stubAnalyticsRepo) ListTopCreatorsPerTeam(context.Context) ([]domain.TeamTopCreator, error) {
	if s.creatorsErr != nil {
		return nil, s.creatorsErr
	}
	return s.creators, nil
}

type stubTeamRepo struct {
	teams   []domain.Team
	listErr error
}

func (s *stubTeamRepo) Create(context.Context, domain.Team) (domain.Team, error) {
	panic("not implemented")
}

func (s *stubTeamRepo) ListByUserID(context.Context, int64) ([]domain.Team, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.teams, nil
}

func (s *stubTeamRepo) GetByID(context.Context, int64) (domain.Team, error) {
	panic("not implemented")
}

func (s *stubTeamRepo) AddMember(context.Context, domain.TeamMember) error {
	panic("not implemented")
}

func (s *stubTeamRepo) GetMemberRole(context.Context, int64, int64) (domain.TeamRole, error) {
	panic("not implemented")
}
