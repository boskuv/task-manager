package app

import (
	"strings"
	"testing"
)

func TestOpenRedisEmptyAddr(t *testing.T) {
	t.Parallel()

	_, err := openRedis(RedisConfig{})
	if err == nil {
		t.Fatal("expected error for empty addr")
	}
	if !strings.Contains(err.Error(), "addr is required") {
		t.Errorf("error = %q, want addr required message", err.Error())
	}
}

func TestOpenRedisInvalidAddr(t *testing.T) {
	t.Parallel()

	_, err := openRedis(RedisConfig{
		Addr: "127.0.0.1:1",
	})
	if err == nil {
		t.Fatal("expected error for unreachable redis")
	}
	if !strings.Contains(err.Error(), "ping redis") {
		t.Errorf("error = %q, want ping failure", err.Error())
	}
}
