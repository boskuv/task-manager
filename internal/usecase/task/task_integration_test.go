//go:build integration

package task

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/boskuv/task-manager/internal/domain"
	"github.com/boskuv/task-manager/internal/repository"
	mysqlrepo "github.com/boskuv/task-manager/internal/repository/mysql"
	redisrepo "github.com/boskuv/task-manager/internal/repository/redis"
	"github.com/boskuv/task-manager/internal/testutil/integration"
)

func TestMain(m *testing.M) {
	code := m.Run()
	integration.Shutdown()
	os.Exit(code)
}

func TestListWritesToRedisIntegration(t *testing.T) {
	t.Parallel()

	client := integration.Redis(t)
	ctx := context.Background()

	teams := newMockTeamRepo()
	teams.members[memberKey{teamID: 1, userID: 1}] = domain.TeamMember{
		TeamID: 1,
		UserID: 1,
		Role:   domain.TeamRoleMember,
	}

	tasks := newMockTaskRepo()
	tasks.tasks[1] = domain.Task{
		ID:        1,
		TeamID:    1,
		Title:     "DB task",
		Status:    domain.TaskStatusTodo,
		CreatedBy: 1,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	cache := redisrepo.NewTaskCache(client)
	svc := NewService(tasks, teams, newMockTaskHistoryRepo(), cache)

	_, err := svc.List(ctx, 1, ListInput{TeamID: 1})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	keys, err := client.Keys(ctx, "tasks:team:1:filter:*").Result()
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	if len(keys) == 0 {
		t.Fatal("expected cache key after List")
	}

	cached, ok, err := cache.Get(ctx, repository.TaskFilter{TeamID: 1})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Fatal("expected cache hit")
	}
	if cached.Total != 1 || cached.Items[0].Title != "DB task" {
		t.Fatalf("cached = %+v", cached)
	}
}

func TestListWritesToRedisWithMySQLIntegration(t *testing.T) {
	t.Parallel()

	db := integration.MySQL(t)
	ctx := context.Background()

	users := mysqlrepo.NewUserRepo(db)
	teamsRepo := mysqlrepo.NewTeamRepo(db)
	tasksRepo := mysqlrepo.NewTaskRepo(db)

	owner, err := users.Create(ctx, domain.User{Email: "usecase-owner@example.com", PasswordHash: "hash"})
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}
	team, err := teamsRepo.Create(ctx, domain.Team{Name: "Usecase Team", CreatedBy: owner.ID})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := teamsRepo.AddMember(ctx, domain.TeamMember{
		TeamID: team.ID,
		UserID: owner.ID,
		Role:   domain.TeamRoleOwner,
	}); err != nil {
		t.Fatalf("add owner: %v", err)
	}
	if _, err := tasksRepo.Create(ctx, domain.Task{
		TeamID:      team.ID,
		Title:       "MySQL task",
		Description: "",
		Status:      domain.TaskStatusTodo,
		CreatedBy:   owner.ID,
	}); err != nil {
		t.Fatalf("create task: %v", err)
	}

	client := integration.Redis(t)
	cache := redisrepo.NewTaskCache(client)
	svc := NewService(tasksRepo, teamsRepo, mysqlrepo.NewTaskHistoryRepo(db), cache)

	_, err = svc.List(ctx, owner.ID, ListInput{TeamID: team.ID})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	keys, err := client.Keys(ctx, "tasks:team:"+strconv.FormatInt(team.ID, 10)+":filter:*").Result()
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	if len(keys) == 0 {
		t.Fatal("expected cache key after List with mysql repo")
	}
}
