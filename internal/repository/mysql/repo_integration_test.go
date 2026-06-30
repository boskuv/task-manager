//go:build integration

package mysql_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/repository"
	mysqlrepo "github.com/boskuv/task-manager/internal/repository/mysql"
	"github.com/boskuv/task-manager/internal/testutil/integration"
)

func TestMain(m *testing.M) {
	code := m.Run()
	integration.Shutdown()
	os.Exit(code)
}

func TestUserRepoCreateAndGetIntegration(t *testing.T) {
	t.Parallel()

	db := integration.MySQL(t)
	repo := mysqlrepo.NewUserRepo(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, domain.User{
		Email:        "user@example.com",
		PasswordHash: "hash",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == 0 || created.Email != "user@example.com" {
		t.Fatalf("created = %+v", created)
	}
	if created.CreatedAt.IsZero() {
		t.Fatal("expected created_at to be set")
	}

	byID, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if byID.Email != created.Email {
		t.Fatalf("GetByID = %+v, want %+v", byID, created)
	}

	byEmail, err := repo.GetByEmail(ctx, "user@example.com")
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if byEmail.ID != created.ID {
		t.Fatalf("GetByEmail = %+v, want id %d", byEmail, created.ID)
	}
}

func TestUserRepoDuplicateEmailIntegration(t *testing.T) {
	t.Parallel()

	db := integration.MySQL(t)
	repo := mysqlrepo.NewUserRepo(db)
	ctx := context.Background()

	if _, err := repo.Create(ctx, domain.User{
		Email:        "dup@example.com",
		PasswordHash: "hash",
	}); err != nil {
		t.Fatalf("first Create: %v", err)
	}

	_, err := repo.Create(ctx, domain.User{
		Email:        "dup@example.com",
		PasswordHash: "other",
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("error = %v, want ErrConflict", err)
	}
}

func TestUserRepoNotFoundIntegration(t *testing.T) {
	t.Parallel()

	db := integration.MySQL(t)
	repo := mysqlrepo.NewUserRepo(db)
	ctx := context.Background()

	_, err := repo.GetByEmail(ctx, "missing@example.com")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetByEmail error = %v, want ErrNotFound", err)
	}

	_, err = repo.GetByID(ctx, 9999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetByID error = %v, want ErrNotFound", err)
	}
}

func TestTeamRepoCreateListMembersIntegration(t *testing.T) {
	t.Parallel()

	db := integration.MySQL(t)
	users := mysqlrepo.NewUserRepo(db)
	teams := mysqlrepo.NewTeamRepo(db)
	ctx := context.Background()

	owner, err := users.Create(ctx, domain.User{
		Email:        "owner@example.com",
		PasswordHash: "hash",
	})
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}
	member, err := users.Create(ctx, domain.User{
		Email:        "member@example.com",
		PasswordHash: "hash",
	})
	if err != nil {
		t.Fatalf("create member: %v", err)
	}

	team, err := teams.Create(ctx, domain.Team{
		Name:      "Backend",
		CreatedBy: owner.ID,
	})
	if err != nil {
		t.Fatalf("Create team: %v", err)
	}
	if team.Name != "Backend" || team.CreatedBy != owner.ID {
		t.Fatalf("team = %+v", team)
	}

	if err := teams.AddMember(ctx, domain.TeamMember{
		TeamID: team.ID,
		UserID: owner.ID,
		Role:   domain.TeamRoleOwner,
	}); err != nil {
		t.Fatalf("AddMember owner: %v", err)
	}
	if err := teams.AddMember(ctx, domain.TeamMember{
		TeamID: team.ID,
		UserID: member.ID,
		Role:   domain.TeamRoleMember,
	}); err != nil {
		t.Fatalf("AddMember member: %v", err)
	}

	role, err := teams.GetMemberRole(ctx, team.ID, member.ID)
	if err != nil {
		t.Fatalf("GetMemberRole: %v", err)
	}
	if role != domain.TeamRoleMember {
		t.Fatalf("role = %q, want member", role)
	}

	list, err := teams.ListByUserID(ctx, owner.ID)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(list) != 1 || list[0].ID != team.ID {
		t.Fatalf("teams = %+v, want team %d", list, team.ID)
	}

	got, err := teams.GetByID(ctx, team.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "Backend" {
		t.Fatalf("GetByID = %+v", got)
	}
}

func TestTeamRepoDuplicateMemberIntegration(t *testing.T) {
	t.Parallel()

	db := integration.MySQL(t)
	users := mysqlrepo.NewUserRepo(db)
	teams := mysqlrepo.NewTeamRepo(db)
	ctx := context.Background()

	owner := mustCreateUser(t, users, "owner-dup@example.com")
	team := mustCreateTeam(t, teams, owner.ID, "Team")

	member := domain.TeamMember{TeamID: team.ID, UserID: owner.ID, Role: domain.TeamRoleOwner}
	if err := teams.AddMember(ctx, member); err != nil {
		t.Fatalf("first AddMember: %v", err)
	}

	err := teams.AddMember(ctx, member)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("error = %v, want ErrConflict", err)
	}
}

func TestTaskRepoCRUDAndListIntegration(t *testing.T) {
	t.Parallel()

	db := integration.MySQL(t)
	users := mysqlrepo.NewUserRepo(db)
	teams := mysqlrepo.NewTeamRepo(db)
	tasks := mysqlrepo.NewTaskRepo(db)
	ctx := context.Background()

	owner := mustCreateUser(t, users, "task-owner@example.com")
	assignee := mustCreateUser(t, users, "task-assignee@example.com")
	team := mustCreateTeam(t, teams, owner.ID, "Tasks Team")
	mustAddMember(t, teams, team.ID, owner.ID, domain.TeamRoleOwner)
	mustAddMember(t, teams, team.ID, assignee.ID, domain.TeamRoleMember)

	assigneeID := assignee.ID
	created, err := tasks.Create(ctx, domain.Task{
		TeamID:      team.ID,
		Title:       "First task",
		Description: "details",
		Status:      domain.TaskStatusTodo,
		AssigneeID:  &assigneeID,
		CreatedBy:   owner.ID,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == 0 || created.Title != "First task" {
		t.Fatalf("created = %+v", created)
	}

	doneStatus := domain.TaskStatusDone
	if _, err := tasks.Create(ctx, domain.Task{
		TeamID:      team.ID,
		Title:       "Second task",
		Description: "",
		Status:      doneStatus,
		CreatedBy:   owner.ID,
	}); err != nil {
		t.Fatalf("create second task: %v", err)
	}

	got, err := tasks.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.AssigneeID == nil || *got.AssigneeID != assignee.ID {
		t.Fatalf("assignee = %v, want %d", got.AssigneeID, assignee.ID)
	}

	created.Title = "Updated task"
	created.Status = domain.TaskStatusInProgress
	updated, err := tasks.Update(ctx, created)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Title != "Updated task" || updated.Status != domain.TaskStatusInProgress {
		t.Fatalf("updated = %+v", updated)
	}

	status := domain.TaskStatusInProgress
	result, err := tasks.List(ctx, repository.TaskFilter{
		TeamID:   team.ID,
		Status:   &status,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("List by status: %v", err)
	}
	if result.Total != 1 || len(result.Items) != 1 || result.Items[0].ID != created.ID {
		t.Fatalf("result = %+v", result)
	}

	result, err = tasks.List(ctx, repository.TaskFilter{
		TeamID:     team.ID,
		AssigneeID: &assigneeID,
		Page:       1,
		PageSize:   10,
	})
	if err != nil {
		t.Fatalf("List by assignee: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("assignee filter total = %d, want 1", result.Total)
	}

	_, err = tasks.Update(ctx, domain.Task{ID: 9999, Title: "missing"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("update missing error = %v, want ErrNotFound", err)
	}
}

func TestTaskRepoOrphanAssigneeIntegration(t *testing.T) {
	t.Parallel()

	db := integration.MySQL(t)
	users := mysqlrepo.NewUserRepo(db)
	teams := mysqlrepo.NewTeamRepo(db)
	tasks := mysqlrepo.NewTaskRepo(db)
	ctx := context.Background()

	owner := mustCreateUser(t, users, "orphan-owner@example.com")
	outsider := mustCreateUser(t, users, "orphan-outsider@example.com")
	team := mustCreateTeam(t, teams, owner.ID, "Orphan Team")
	mustAddMember(t, teams, team.ID, owner.ID, domain.TeamRoleOwner)

	outsiderID := outsider.ID
	task, err := tasks.Create(ctx, domain.Task{
		TeamID:      team.ID,
		Title:       "Orphan task",
		Description: "",
		Status:      domain.TaskStatusTodo,
		AssigneeID:  &outsiderID,
		CreatedBy:   owner.ID,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	orphan, err := tasks.HasOrphanAssignee(ctx, task.ID)
	if err != nil {
		t.Fatalf("HasOrphanAssignee: %v", err)
	}
	if !orphan {
		t.Fatal("expected orphan assignee")
	}

	orphans, err := tasks.ListOrphanAssignees(ctx)
	if err != nil {
		t.Fatalf("ListOrphanAssignees: %v", err)
	}
	if len(orphans) != 1 || orphans[0].ID != task.ID {
		t.Fatalf("orphans = %+v", orphans)
	}

	mustAddMember(t, teams, team.ID, outsider.ID, domain.TeamRoleMember)
	orphan, err = tasks.HasOrphanAssignee(ctx, task.ID)
	if err != nil {
		t.Fatalf("HasOrphanAssignee after invite: %v", err)
	}
	if orphan {
		t.Fatal("expected assignee to be valid after joining team")
	}
}

func TestTaskHistoryRepoIntegration(t *testing.T) {
	t.Parallel()

	db := integration.MySQL(t)
	users := mysqlrepo.NewUserRepo(db)
	teams := mysqlrepo.NewTeamRepo(db)
	tasks := mysqlrepo.NewTaskRepo(db)
	history := mysqlrepo.NewTaskHistoryRepo(db)
	ctx := context.Background()

	owner := mustCreateUser(t, users, "history-owner@example.com")
	team := mustCreateTeam(t, teams, owner.ID, "History Team")
	mustAddMember(t, teams, team.ID, owner.ID, domain.TeamRoleOwner)

	task, err := tasks.Create(ctx, domain.Task{
		TeamID:      team.ID,
		Title:       "Tracked",
		Description: "",
		Status:      domain.TaskStatusTodo,
		CreatedBy:   owner.ID,
	})
	if err != nil {
		t.Fatalf("Create task: %v", err)
	}

	inserted, err := history.Insert(ctx, domain.TaskHistory{
		TaskID:    task.ID,
		ChangedBy: owner.ID,
		Field:     domain.HistoryFieldStatus,
		OldValue:  "todo",
		NewValue:  "in_progress",
	})
	if err != nil {
		t.Fatalf("Insert: %v", err)
	}
	if inserted.ID == 0 || inserted.CreatedAt.IsZero() {
		t.Fatalf("inserted = %+v", inserted)
	}

	entries, err := history.ListByTaskID(ctx, task.ID)
	if err != nil {
		t.Fatalf("ListByTaskID: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %+v", entries)
	}
	if entries[0].Field != domain.HistoryFieldStatus || entries[0].NewValue != "in_progress" {
		t.Fatalf("entry = %+v", entries[0])
	}
}

func mustCreateUser(t *testing.T, repo *mysqlrepo.UserRepo, email string) domain.User {
	t.Helper()
	user, err := repo.Create(context.Background(), domain.User{
		Email:        email,
		PasswordHash: "hash",
	})
	if err != nil {
		t.Fatalf("create user %s: %v", email, err)
	}
	return user
}

func mustCreateTeam(t *testing.T, repo *mysqlrepo.TeamRepo, ownerID int64, name string) domain.Team {
	t.Helper()
	team, err := repo.Create(context.Background(), domain.Team{
		Name:      name,
		CreatedBy: ownerID,
	})
	if err != nil {
		t.Fatalf("create team %s: %v", name, err)
	}
	return team
}

func mustAddMember(t *testing.T, repo *mysqlrepo.TeamRepo, teamID, userID int64, role domain.TeamRole) {
	t.Helper()
	if err := repo.AddMember(context.Background(), domain.TeamMember{
		TeamID: teamID,
		UserID: userID,
		Role:   role,
	}); err != nil {
		t.Fatalf("add member team=%d user=%d: %v", teamID, userID, err)
	}
}
