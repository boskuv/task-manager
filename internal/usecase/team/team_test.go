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
	svc := NewService(repo, newMockUserRepo(), nil)

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

	svc := NewService(newMockTeamRepo(), newMockUserRepo(), nil)

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
	svc := NewService(repo, newMockUserRepo(), nil)

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

	svc := NewService(newMockTeamRepo(), newMockUserRepo(), nil)

	_, err := svc.List(context.Background(), 0)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
}

func TestListEmpty(t *testing.T) {
	t.Parallel()

	teams, err := NewService(newMockTeamRepo(), newMockUserRepo(), nil).List(context.Background(), 42)
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

func TestInviteSendsEmail(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	users := newMockUserRepo()
	users.users["invitee@example.com"] = domain.User{ID: 2, Email: "invitee@example.com"}
	mailer := &stubInviteMailer{}
	svc := NewService(teams, users, mailer)

	team, err := svc.Create(context.Background(), 1, "Backend")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if _, err := svc.Invite(context.Background(), 1, team.ID, "invitee@example.com", "member"); err != nil {
		t.Fatalf("Invite: %v", err)
	}
	if mailer.calls != 1 {
		t.Fatalf("mailer calls = %d, want 1", mailer.calls)
	}
	if mailer.lastEmail != "invitee@example.com" || mailer.lastTeam != "Backend" {
		t.Fatalf("mailer payload = (%q, %q), want (invitee@example.com, Backend)", mailer.lastEmail, mailer.lastTeam)
	}
	if mailer.lastRole != domain.TeamRoleMember {
		t.Fatalf("mailer role = %q, want member", mailer.lastRole)
	}
}

func TestInviteSucceedsWhenEmailFails(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	users := newMockUserRepo()
	users.users["invitee@example.com"] = domain.User{ID: 2, Email: "invitee@example.com"}
	svc := NewService(teams, users, &stubInviteMailer{err: errors.New("smtp down")})

	team, err := svc.Create(context.Background(), 1, "Backend")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	member, err := svc.Invite(context.Background(), 1, team.ID, "invitee@example.com", "member")
	if err != nil {
		t.Fatalf("Invite: %v", err)
	}
	if member.UserID != 2 {
		t.Fatalf("member.UserID = %d, want 2", member.UserID)
	}
}

type stubInviteMailer struct {
	calls     int
	lastEmail string
	lastTeam  string
	lastRole  domain.TeamRole
	err       error
}

func (s *stubInviteMailer) SendTeamInvite(_ context.Context, toEmail, teamName string, role domain.TeamRole) error {
	if s.err != nil {
		return s.err
	}
	s.calls++
	s.lastEmail = toEmail
	s.lastTeam = teamName
	s.lastRole = role
	return nil
}

func TestInviteSuccessAsOwner(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	users := newMockUserRepo()
	users.users["invitee@example.com"] = domain.User{ID: 2, Email: "invitee@example.com"}
	svc := NewService(teams, users, nil)

	team, err := svc.Create(context.Background(), 1, "Backend")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	member, err := svc.Invite(context.Background(), 1, team.ID, "Invitee@Example.com", "member")
	if err != nil {
		t.Fatalf("Invite: %v", err)
	}
	if member.UserID != 2 || member.Role != domain.TeamRoleMember {
		t.Fatalf("member = %+v, want user_id=2 role=member", member)
	}
}

func TestInviteSuccessAsAdmin(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	users := newMockUserRepo()
	users.users["invitee@example.com"] = domain.User{ID: 2, Email: "invitee@example.com"}
	svc := NewService(teams, users, nil)

	team, err := svc.Create(context.Background(), 1, "Backend")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	teams.members[memberKey{teamID: team.ID, userID: 3}] = domain.TeamMember{
		TeamID: team.ID,
		UserID: 3,
		Role:   domain.TeamRoleAdmin,
	}

	member, err := svc.Invite(context.Background(), 3, team.ID, "invitee@example.com", "admin")
	if err != nil {
		t.Fatalf("Invite: %v", err)
	}
	if member.Role != domain.TeamRoleAdmin {
		t.Fatalf("member.Role = %q, want admin", member.Role)
	}
}

func TestInviteForbiddenForMember(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	users := newMockUserRepo()
	users.users["invitee@example.com"] = domain.User{ID: 2, Email: "invitee@example.com"}
	svc := NewService(teams, users, nil)

	team, err := svc.Create(context.Background(), 1, "Backend")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	teams.members[memberKey{teamID: team.ID, userID: 4}] = domain.TeamMember{
		TeamID: team.ID,
		UserID: 4,
		Role:   domain.TeamRoleMember,
	}

	_, err = svc.Invite(context.Background(), 4, team.ID, "invitee@example.com", "member")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("error = %v, want ErrForbidden", err)
	}
}

func TestInviteForbiddenForNonMember(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	users := newMockUserRepo()
	users.users["invitee@example.com"] = domain.User{ID: 2, Email: "invitee@example.com"}
	svc := NewService(teams, users, nil)

	team, err := svc.Create(context.Background(), 1, "Backend")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = svc.Invite(context.Background(), 99, team.ID, "invitee@example.com", "member")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("error = %v, want ErrForbidden", err)
	}
}

func TestInviteUserNotFound(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	svc := NewService(teams, newMockUserRepo(), nil)

	team, err := svc.Create(context.Background(), 1, "Backend")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = svc.Invite(context.Background(), 1, team.ID, "missing@example.com", "member")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestInviteTeamNotFound(t *testing.T) {
	t.Parallel()

	svc := NewService(newMockTeamRepo(), newMockUserRepo(), nil)

	_, err := svc.Invite(context.Background(), 1, 999, "user@example.com", "member")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestInviteAlreadyMember(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	users := newMockUserRepo()
	users.users["invitee@example.com"] = domain.User{ID: 2, Email: "invitee@example.com"}
	svc := NewService(teams, users, nil)

	team, err := svc.Create(context.Background(), 1, "Backend")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if _, err := svc.Invite(context.Background(), 1, team.ID, "invitee@example.com", "member"); err != nil {
		t.Fatalf("first Invite: %v", err)
	}

	_, err = svc.Invite(context.Background(), 1, team.ID, "invitee@example.com", "admin")
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("error = %v, want ErrConflict", err)
	}
}

func TestInviteInvalidInput(t *testing.T) {
	t.Parallel()

	teams := newMockTeamRepo()
	users := newMockUserRepo()
	users.users["invitee@example.com"] = domain.User{ID: 2, Email: "invitee@example.com"}
	svc := NewService(teams, users, nil)

	team, err := svc.Create(context.Background(), 1, "Backend")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	tests := []struct {
		name  string
		email string
		role  string
	}{
		{name: "empty email", email: "", role: "member"},
		{name: "invalid email", email: "not-email", role: "member"},
		{name: "empty role", email: "invitee@example.com", role: ""},
		{name: "owner role", email: "invitee@example.com", role: "owner"},
		{name: "unknown role", email: "invitee@example.com", role: "guest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := svc.Invite(context.Background(), 1, team.ID, tt.email, tt.role)
			if !errors.Is(err, domain.ErrInvalidInput) {
				t.Fatalf("error = %v, want ErrInvalidInput", err)
			}
		})
	}
}

type mockUserRepo struct {
	users map[string]domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]domain.User)}
}

func (m *mockUserRepo) Create(_ context.Context, user domain.User) (domain.User, error) {
	if _, exists := m.users[user.Email]; exists {
		return domain.User{}, domain.ErrConflict
	}
	m.users[user.Email] = user
	return user, nil
}

func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (domain.User, error) {
	user, ok := m.users[email]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id int64) (domain.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return domain.User{}, domain.ErrNotFound
}
