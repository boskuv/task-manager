//go:build integration

package redis

import (
	"context"
	"os"
	"strconv"
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

func TestTaskCacheSetGetIntegration(t *testing.T) {
	t.Parallel()

	client := integration.Redis(t)
	ctx := context.Background()

	cache := NewTaskCache(client)
	filter := repository.TaskFilter{TeamID: 99, Page: 1, PageSize: 20}
	want := repository.TaskListResult{
		Items: []domain.Task{{
			ID:     1,
			TeamID: 99,
			Title:  "cached",
			Status: domain.TaskStatusTodo,
		}},
		Total: 1,
	}

	if err := cache.Set(ctx, filter, want); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, ok, err := cache.Get(ctx, filter)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.Total != want.Total || len(got.Items) != 1 || got.Items[0].Title != "cached" {
		t.Fatalf("got = %+v, want %+v", got, want)
	}
}

func TestTaskCacheSetMySQLResultIntegration(t *testing.T) {
	t.Parallel()

	db := integration.MySQL(t)
	ctx := context.Background()

	users := mysqlrepo.NewUserRepo(db)
	teams := mysqlrepo.NewTeamRepo(db)
	tasks := mysqlrepo.NewTaskRepo(db)

	owner, err := users.Create(ctx, domain.User{Email: "cache-owner@example.com", PasswordHash: "hash"})
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}
	team, err := teams.Create(ctx, domain.Team{Name: "Cache Team", CreatedBy: owner.ID})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := teams.AddMember(ctx, domain.TeamMember{
		TeamID: team.ID,
		UserID: owner.ID,
		Role:   domain.TeamRoleOwner,
	}); err != nil {
		t.Fatalf("add owner: %v", err)
	}
	if _, err := tasks.Create(ctx, domain.Task{
		TeamID:      team.ID,
		Title:       "Listed task",
		Description: "",
		Status:      domain.TaskStatusTodo,
		CreatedBy:   owner.ID,
	}); err != nil {
		t.Fatalf("create task: %v", err)
	}

	repo := mysqlrepo.NewTaskRepo(db)
	result, err := repo.List(ctx, repository.TaskFilter{TeamID: team.ID, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("result.Total = %d, want 1", result.Total)
	}

	client := integration.Redis(t)
	cache := NewTaskCache(client)
	if err := cache.Set(ctx, repository.TaskFilter{TeamID: team.ID, Page: 1, PageSize: 20}, result); err != nil {
		t.Fatalf("Set: %v", err)
	}

	keys, err := client.Keys(ctx, "tasks:team:"+strconv.FormatInt(team.ID, 10)+":filter:*").Result()
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	if len(keys) == 0 {
		t.Fatal("expected cache key")
	}
}

func TestTaskCacheInvalidateTeamIntegration(t *testing.T) {
	t.Parallel()

	client := integration.Redis(t)
	ctx := context.Background()
	cache := NewTaskCache(client)

	filterTeam1 := repository.TaskFilter{TeamID: 1, Page: 1, PageSize: 20}
	filterTeam2 := repository.TaskFilter{TeamID: 2, Page: 1, PageSize: 20}
	result := repository.TaskListResult{Total: 1}

	if err := cache.Set(ctx, filterTeam1, result); err != nil {
		t.Fatalf("set team 1: %v", err)
	}
	if err := cache.Set(ctx, filterTeam2, result); err != nil {
		t.Fatalf("set team 2: %v", err)
	}

	if err := cache.InvalidateTeam(ctx, 1); err != nil {
		t.Fatalf("InvalidateTeam: %v", err)
	}

	_, ok, err := cache.Get(ctx, filterTeam1)
	if err != nil {
		t.Fatalf("get team 1: %v", err)
	}
	if ok {
		t.Fatal("expected team 1 cache to be invalidated")
	}

	_, ok, err = cache.Get(ctx, filterTeam2)
	if err != nil {
		t.Fatalf("get team 2: %v", err)
	}
	if !ok {
		t.Fatal("expected team 2 cache to remain")
	}
}
