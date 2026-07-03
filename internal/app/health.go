package app

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/boskuv/task-manager/internal/handler"
)

const healthCheckTimeout = 2 * time.Second

type readinessResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

func livenessHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func newReadinessHandler(db *sql.DB, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), healthCheckTimeout)
		defer cancel()

		checks := map[string]string{
			"mysql": checkMySQL(ctx, db),
			"redis": checkRedis(ctx, rdb),
		}

		status := "ready"
		code := http.StatusOK
		for _, result := range checks {
			if result != "ok" {
				status = "not_ready"
				code = http.StatusServiceUnavailable
				break
			}
		}

		handler.WriteJSON(w, code, readinessResponse{
			Status: status,
			Checks: checks,
		})
	}
}

func checkMySQL(ctx context.Context, db *sql.DB) string {
	if db == nil {
		return "unavailable"
	}
	if err := db.PingContext(ctx); err != nil {
		return "error"
	}
	return "ok"
}

func checkRedis(ctx context.Context, rdb *redis.Client) string {
	if rdb == nil {
		return "unavailable"
	}
	if err := rdb.Ping(ctx).Err(); err != nil {
		return "error"
	}
	return "ok"
}
