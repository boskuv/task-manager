package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// AccessLog writes a structured log line for each HTTP request.
func AccessLog(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL != nil && r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(recorder, r)

			attrs := []any{
				"method", r.Method,
				"path", routePattern(r),
				"status", recorder.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"remote_addr", r.RemoteAddr,
			}
			if userID, ok := UserIDFromContext(r.Context()); ok {
				attrs = append(attrs, "user_id", userID)
			}

			level := slog.LevelInfo
			switch {
			case recorder.status >= http.StatusInternalServerError:
				level = slog.LevelError
			case recorder.status >= http.StatusBadRequest:
				level = slog.LevelWarn
			}

			logger.Log(r.Context(), level, "http request", attrs...)
		})
	}
}
