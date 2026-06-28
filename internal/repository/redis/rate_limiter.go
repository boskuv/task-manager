package redis

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/boskuv/task-manager/internal/repository"
)

const rateLimitWindow = time.Minute

var _ repository.RateLimiter = (*RateLimiter)(nil)

var rateLimitScript = goredis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window_ms = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local member = ARGV[4]

redis.call('ZREMRANGEBYSCORE', key, 0, now - window_ms)
local count = redis.call('ZCARD', key)
if count < limit then
  redis.call('ZADD', key, now, member)
  redis.call('PEXPIRE', key, window_ms)
  return 1
end
return 0
`)

// RateLimiter implements a Redis-backed sliding window rate limiter.
type RateLimiter struct {
	client *goredis.Client
	limit  int
}

// NewRateLimiter returns a rate limiter with the given requests-per-minute cap.
func NewRateLimiter(client *goredis.Client, limit int) *RateLimiter {
	return &RateLimiter{client: client, limit: limit}
}

// Allow reports whether the request for key is within the sliding window limit.
func (r *RateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	if r.limit <= 0 {
		return true, nil
	}

	now := time.Now().UnixMilli()
	member, err := rateLimitMember(now)
	if err != nil {
		return false, fmt.Errorf("rate limit member: %w", err)
	}

	allowed, err := rateLimitScript.Run(
		ctx,
		r.client,
		[]string{key},
		now,
		rateLimitWindow.Milliseconds(),
		r.limit,
		member,
	).Int()
	if err != nil {
		return false, fmt.Errorf("rate limit script: %w", err)
	}

	return allowed == 1, nil
}

func rateLimitMember(nowMillis int64) (string, error) {
	var suffix [8]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d:%d", nowMillis, binary.BigEndian.Uint64(suffix[:])), nil
}
