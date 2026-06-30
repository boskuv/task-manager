//go:build integration

package redis

import (
	"context"
	"fmt"
	"testing"
)

func TestRateLimiterSlidingWindowIntegration(t *testing.T) {
	client := newIntegrationRedisClient(t)
	ctx := context.Background()

	const limit = 3
	limiter := NewRateLimiter(client, limit)
	key := "rate_limit:user:integration"

	for i := 0; i < limit; i++ {
		allowed, err := limiter.Allow(ctx, key)
		if err != nil {
			t.Fatalf("Allow #%d: %v", i+1, err)
		}
		if !allowed {
			t.Fatalf("request #%d should be allowed", i+1)
		}
	}

	allowed, err := limiter.Allow(ctx, key)
	if err != nil {
		t.Fatalf("Allow over limit: %v", err)
	}
	if allowed {
		t.Fatal("expected request over limit to be denied")
	}
}

func TestRateLimiterIsolatesKeysIntegration(t *testing.T) {
	client := newIntegrationRedisClient(t)
	ctx := context.Background()

	limiter := NewRateLimiter(client, 1)

	allowed, err := limiter.Allow(ctx, "rate_limit:user:1")
	if err != nil {
		t.Fatalf("Allow user 1: %v", err)
	}
	if !allowed {
		t.Fatal("first request for user 1 should be allowed")
	}

	allowed, err = limiter.Allow(ctx, "rate_limit:user:2")
	if err != nil {
		t.Fatalf("Allow user 2: %v", err)
	}
	if !allowed {
		t.Fatal("first request for user 2 should be allowed")
	}

	allowed, err = limiter.Allow(ctx, fmt.Sprintf("rate_limit:user:%d", 1))
	if err != nil {
		t.Fatalf("Allow user 1 again: %v", err)
	}
	if allowed {
		t.Fatal("second request for user 1 should be denied")
	}
}
