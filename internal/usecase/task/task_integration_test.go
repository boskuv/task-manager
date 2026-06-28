//go:build integration

package task

import (
	"context"
	"database/sql"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	_ "github.com/go-sql-driver/mysql"

	"github.com/boskuv/task-manager/internal/domain"
	mysqlrepo "github.com/boskuv/task-manager/internal/repository/mysql"
	"github.com/boskuv/task-manager/internal/repository"
	redisrepo "github.com/boskuv/task-manager/internal/repository/redis"
)

func TestListWritesToRedisIntegration(t *testing.T) {
	client := goredis.NewClient(&goredis.Options{Addr: "localhost:6379"})
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("redis not available: %v", err)
	}
	t.Cleanup(func() {
		_ = client.FlushDB(ctx).Err()
		_ = client.Close()
	})

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
	db, err := sql.Open("mysql", "taskmanager:taskmanager@tcp(localhost:3306)/taskmanager?parseTime=true&charset=utf8mb4")
	if err != nil {
		t.Fatalf("open mysql: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		t.Skipf("mysql not available: %v", err)
	}

	client := goredis.NewClient(&goredis.Options{Addr: "localhost:6379"})
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("redis not available: %v", err)
	}
	t.Cleanup(func() {
		_ = client.FlushDB(ctx).Err()
		_ = client.Close()
	})

	teams := mysqlrepo.NewTeamRepo(db)
	tasks := mysqlrepo.NewTaskRepo(db)
	cache := redisrepo.NewTaskCache(client)
	svc := NewService(tasks, teams, mysqlrepo.NewTaskHistoryRepo(db), cache)

	_, err = svc.List(ctx, 1, ListInput{TeamID: 1})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	keys, err := client.Keys(ctx, "tasks:team:1:filter:*").Result()
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	if len(keys) == 0 {
		t.Fatal("expected cache key after List with mysql repo")
	}
}
