//go:build integration

package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

var (
	redisOnce     sync.Once
	redisClient   *goredis.Client
	redisCleanup  func()
	redisSetupErr error
	redisMu       sync.Mutex
)

// Redis returns a shared Redis client with a flushed database per test.
func Redis(t *testing.T) *goredis.Client {
	t.Helper()

	redisOnce.Do(func() {
		redisClient, redisCleanup, redisSetupErr = startRedis()
	})
	if redisSetupErr != nil {
		t.Fatalf("start redis: %v", redisSetupErr)
	}

	redisMu.Lock()
	flushRedis(t, redisClient)
	t.Cleanup(func() {
		flushRedis(t, redisClient)
		redisMu.Unlock()
	})

	return redisClient
}

func flushRedis(t *testing.T, client *goredis.Client) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("flush redis: %v", err)
	}
}

func startRedis() (*goredis.Client, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := tcredis.Run(ctx, "redis:7")
	if err != nil {
		return nil, nil, fmt.Errorf("run container: %w", err)
	}

	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, fmt.Errorf("endpoint: %w", err)
	}

	client := goredis.NewClient(&goredis.Options{Addr: endpoint})
	pingCtx, pingCancel := context.WithTimeout(ctx, 15*time.Second)
	defer pingCancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		_ = container.Terminate(ctx)
		return nil, nil, fmt.Errorf("ping redis: %w", err)
	}

	cleanup := func() {
		_ = client.Close()
		termCtx, termCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer termCancel()
		_ = container.Terminate(termCtx)
	}

	return client, cleanup, nil
}
