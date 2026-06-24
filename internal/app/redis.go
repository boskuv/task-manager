package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisPingTimeout = 5 * time.Second

func openRedis(cfg RedisConfig) (*redis.Client, error) {
	if cfg.Addr == "" {
		return nil, fmt.Errorf("redis addr is required")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), redisPingTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	slog.Info("redis connected", "addr", cfg.Addr, "db", cfg.DB)

	return client, nil
}

func closeRedis(client *redis.Client) {
	if client == nil {
		return
	}
	if err := client.Close(); err != nil {
		slog.Error("close redis", "error", err)
		return
	}
	slog.Info("redis connection closed")
}
