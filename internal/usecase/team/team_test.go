package team

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/boskuv/task-manager/internal/domain"
)

func TestCreateSuccess(t *testing.T) {
	t.Parallel()

	repo := newMockTeamRepo()
	svc := NewService(repo)

	team, err := svc.Create(context.Background(), 1, "  Backend  ")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if team.Name != "Backend" {
		t.Errorf("team.Name = %q, want Backend", team.Name)
	}
	if team.CreatedBy != 1 {
		t.Errorf("team.CreatedBy = %d, want 1", team.CreatedBy)
	}
	if len(repo.teams) != 1 {
		t.Fatalf("teams count = %d, want 1", len(repo.teams))
	}
	if len(repo.members) != 1 {
		t.Fatalf("members count = %d, want 1", len(repo.members))
	}

	member := repo.members[memberKey{teamID: team.ID, userID: 1}]
	if member.Role != domain.TeamRoleOwner {
		t.Errorf("member.Role = %q, want owner", member.Role)
	}
}

func TestCreateInvalidInput(t *testing.T) {
	t.Parallel()

	svc := NewService(newMockTeamRepo())

	tests := []struct {
		name   string
		userID int64
		team   string
	}{
		{name: "empty name", userID: 1, team: ""},
		{name: "whitespace name", userID: 1, team: "   "},
		{name: "name too long", userID: 1, team: strings.Repeat("a", maxTeamNameLength+1)},
		{name: "invalid user id", userID: 0, team: "Backend"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := svc.Create(context.Background(), tt.userID, tt.team)
			if !errors.Is(err, domain.ErrInvalidInput) {
				t.Fatalf("error = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestListSuccess(t *testing.T) {
	t.Parallel()

	repo := newMockTeamRepo()
	svc := NewService(repo)

	first, err := svc.Create(context.Background(), 1, "Team A")
	if err != nil {
		t.Fatalf("first Create: %v", err)
	}
	second, err := svc.Create(context.Background(), 1, "Team B")
	if err != nil {
		t.Fatalf("second Create: %v", err)
	}

	teams, err := svc.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(teams) != 2 {
		t.Fatalf("teams count = %d, want 2", len(teams))
	}
	ids := map[int64]struct{}{teams[0].ID: {}, teams[1].ID: {}}
	if _, ok := ids[first.ID]; !ok {
		t.Fatalf("first team %d missing from list", first.ID)
	}
	if _, ok := ids[second.ID]; !ok {
		t.Fatalf("second team %d missing from list", second.ID)
	}
}

func TestListInvalidUserID(t *testing.T) {
	t.Parallel()

	svc := NewService(newMockTeamRepo())

	_, err := svc.List(context.Background(), 0)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
}

func TestListEmpty(t *testing.T) {
	t.Parallel()

	teams, err := NewService(newMockTeamRepo()).List(context.Background(), 42)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(teams) != 0 {
		t.Fatalf("teams count = %d, want 0", len(teams))
	}
}

type memberKey struct {
	teamID int64
	userID int64
}

type mockTeamRepo struct {
	teams   map[int64]domain.Team
	members map[memberKey]domain.TeamMember
	nextID  int64
}

func newMockTeamRepo() *mockTeamRepo {
	return &mockTeamRepo{
		teams:   make(map[int64]domain.Team),
		members: make(map[memberKey]domain.TeamMember),
		nextID:  1,
	}
}

func (m *mockTeamRepo) Create(_ context.Context, team domain.Team) (domain.Team, error) {
	team.ID = m.nextID
	m.nextID++
	team.CreatedAt = time.Now().UTC()
	m.teams[team.ID] = team
	return team, nil
}

func (m *mockTeamRepo) ListByUserID(_ context.Context, userID int64) ([]domain.Team, error) {
	teams := make([]domain.Team, 0)
	for _, member := range m.members {
		if member.UserID != userID {
			continue
		}
		team, ok := m.teams[member.TeamID]
		if !ok {
			continue
		}
		teams = append(teams, team)
	}
	for i := 0; i < len(teams); i++ {
		for j := i + 1; j < len(teams); j++ {
			if teams[j].CreatedAt.After(teams[i].CreatedAt) {
				teams[i], teams[j] = teams[j], teams[i]
			}
		}
	}
	return teams, nil
}

func (m *mockTeamRepo) GetByID(_ context.Context, id int64) (domain.Team, error) {
	team, ok := m.teams[id]
	if !ok {
		return domain.Team{}, domain.ErrNotFound
	}
	return team, nil
}

func (m *mockTeamRepo) AddMember(_ context.Context, member domain.TeamMember) error {
	key := memberKey{teamID: member.TeamID, userID: member.UserID}
	if _, exists := m.members[key]; exists {
		return domain.ErrConflict
	}
	member.JoinedAt = time.Now().UTC()
	m.members[key] = member
	return nil
}

func (m *mockTeamRepo) GetMemberRole(_ context.Context, teamID, userID int64) (domain.TeamRole, error) {
	member, ok := m.members[memberKey{teamID: teamID, userID: userID}]
	if !ok {
		return "", domain.ErrNotFound
	}
	return member.Role, nil
}
