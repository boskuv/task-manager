package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/boskuv/task-manager/internal/handler"
	"github.com/boskuv/task-manager/internal/repository"
)

// RateLimit enforces a per-user request cap using a sliding window store.
// It must run after Auth so the user id is present in context.
func RateLimit(limiter repository.RateLimiter, requestsPerMinute int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if requestsPerMinute <= 0 || limiter == nil {
				next.ServeHTTP(w, r)
				return
			}

			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				handler.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			allowed, err := limiter.Allow(r.Context(), rateLimitUserKey(userID))
			if err != nil {
				slog.Warn("rate limit check failed", "user_id", userID, "error", err)
				next.ServeHTTP(w, r)
				return
			}
			if !allowed {
				handler.WriteError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func rateLimitUserKey(userID int64) string {
	return fmt.Sprintf("rate_limit:user:%d", userID)
}

// Chain applies middleware in order: the first entry runs first on the request path.
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}
