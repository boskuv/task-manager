//go:build integration

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/boskuv/task-manager/internal/testutil/integration"
)

func TestMain(m *testing.M) {
	code := m.Run()
	integration.Shutdown()
	os.Exit(code)
}

func TestHTTPTaskListWritesRedisIntegration(t *testing.T) {
	db := integration.MySQL(t)
	rdb := integration.Redis(t)

	cfg := &Config{
		Server: ServerConfig{
			Host:            "127.0.0.1",
			Port:            8080,
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 15 * time.Second,
		},
		JWT: JWTConfig{
			Secret:    "integration-test-secret",
			AccessTTL: time.Hour,
		},
		RateLimit: RateLimitConfig{RequestsPerMinute: 100},
		Logging:   LoggingConfig{Level: "error", Format: "text"},
	}

	application, err := newApp(cfg, db, rdb, nil)
	if err != nil {
		t.Fatalf("newApp: %v", err)
	}

	ctx := context.Background()
	email := fmt.Sprintf("http-redis-%d@example.com", time.Now().UnixNano())
	registerBody := fmt.Sprintf(`{"email":%q,"password":"secret123"}`, email)
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

	keys, err := rdb.Keys(ctx, "tasks:team:"+strconv.FormatInt(teamResp.ID, 10)+":filter:*").Result()
	if err != nil {
		t.Fatalf("Keys: %v", err)
	}
	if len(keys) == 0 {
		t.Fatal("expected cache key after HTTP list")
	}
}
