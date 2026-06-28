//go:build integration

package redis

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	_ "github.com/go-sql-driver/mysql"

	"github.com/boskuv/task-manager/internal/domain"
	mysqlrepo "github.com/boskuv/task-manager/internal/repository/mysql"
	"github.com/boskuv/task-manager/internal/repository"
)

func TestTaskCacheSetGetIntegration(t *testing.T) {
	client := newIntegrationRedisClient(t)
	ctx := context.Background()

	cache := NewTaskCache(client)
	filter := repository.TaskFilter{TeamID: 99, Page: 1, PageSize: 20}
	now := time.Now().UTC()
	want := repository.TaskListResult{
		Items: []domain.Task{{
			ID:        1,
			TeamID:    99,
			Title:     "cached",
			Status:    domain.TaskStatusTodo,
			CreatedBy: 1,
			CreatedAt: now,
			UpdatedAt: now,
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
	db, err := sql.Open("mysql", "taskmanager:taskmanager@tcp(localhost:3306)/taskmanager?parseTime=true&charset=utf8mb4")
	if err != nil {
		t.Fatalf("open mysql: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		t.Skipf("mysql not available: %v", err)
	}

	repo := mysqlrepo.NewTaskRepo(db)
	result, err := repo.List(ctx, repository.TaskFilter{TeamID: 1, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if _, err := json.Marshal(result); err != nil {
		t.Fatalf("marshal mysql result: %v", err)
	}

	client := newIntegrationRedisClient(t)
	cache := NewTaskCache(client)
	if err := cache.Set(ctx, repository.TaskFilter{TeamID: 1, Page: 1, PageSize: 20}, result); err != nil {
		t.Fatalf("Set: %v", err)
	}

	keys, err := client.Keys(ctx, "tasks:team:1:filter:*").Result()
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	if len(keys) == 0 {
		t.Fatal("expected cache key")
	}
}

func newIntegrationRedisClient(t *testing.T) *goredis.Client {
	t.Helper()

	client := goredis.NewClient(&goredis.Options{Addr: "localhost:6379"})
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("redis not available: %v", err)
	}

	t.Cleanup(func() {
		_ = client.FlushDB(ctx).Err()
		_ = client.Close()
	})

	return client
}
