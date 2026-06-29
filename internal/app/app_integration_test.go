//go:build integration

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	goredis "github.com/redis/go-redis/v9"
)

func TestHTTPTaskListWritesRedisIntegration(t *testing.T) {
	t.Chdir("../..")

	cfgPath := filepath.Join("configs", "config.yaml")
	if _, err := os.Stat(cfgPath); err != nil {
		t.Fatalf("stat config: %v", err)
	}
	t.Setenv("CONFIG_PATH", cfgPath)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	application, err := New(cfg, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() {
		application.close()
	})

	client := goredis.NewClient(&goredis.Options{Addr: cfg.Redis.Addr})
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("redis not available: %v", err)
	}
	t.Cleanup(func() {
		_ = client.FlushDB(ctx).Err()
		_ = client.Close()
	})

	email := "http-redis@example.com"
	registerBody := `{"email":"` + email + `","password":"secret123","name":"HTTP Redis"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/register", strings.NewReader(registerBody))
	application.server.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register status = %d body = %s", rec.Code, rec.Body.String())
	}

	var registerResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&registerResp); err != nil {
		t.Fatalf("decode register: %v", err)
	}

	teamReq := httptest.NewRequest(http.MethodPost, "/api/v1/teams", strings.NewReader(`{"name":"HTTP Redis Team"}`))
	teamReq.Header.Set("Authorization", "Bearer "+registerResp.Token)
	teamReq.Header.Set("Content-Type", "application/json")
	teamRec := httptest.NewRecorder()
	application.server.Handler.ServeHTTP(teamRec, teamReq)
	if teamRec.Code != http.StatusCreated {
		t.Fatalf("create team status = %d body = %s", teamRec.Code, teamRec.Body.String())
	}

	var teamResp struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(teamRec.Body).Decode(&teamResp); err != nil {
		t.Fatalf("decode team: %v", err)
	}

	taskBody := `{"team_id":` + strconv.FormatInt(teamResp.ID, 10) + `,"title":"Task"}`
	taskReq := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", strings.NewReader(taskBody))
	taskReq.Header.Set("Authorization", "Bearer "+registerResp.Token)
	taskReq.Header.Set("Content-Type", "application/json")
	taskRec := httptest.NewRecorder()
	application.server.Handler.ServeHTTP(taskRec, taskReq)
	if taskRec.Code != http.StatusCreated {
		t.Fatalf("create task status = %d body = %s", taskRec.Code, taskRec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?team_id="+strconv.FormatInt(teamResp.ID, 10), nil)
	listReq.Header.Set("Authorization", "Bearer "+registerResp.Token)
	listRec := httptest.NewRecorder()
	application.server.Handler.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d body = %s", listRec.Code, listRec.Body.String())
	}

	keys, err := client.Keys(ctx, "tasks:team:"+strconv.FormatInt(teamResp.ID, 10)+":filter:*").Result()
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	if len(keys) == 0 {
		t.Fatal("expected cache key after HTTP list")
	}
}
