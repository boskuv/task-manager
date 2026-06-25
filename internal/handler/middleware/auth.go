package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/boskuv/task-manager/internal/handler"
)

type contextKey string

const userIDKey contextKey = "userID"

var errMissingToken = errors.New("missing bearer token")

// TokenParser validates access tokens.
type TokenParser interface {
	ParseAccessToken(tokenString string) (int64, error)
}

// Auth validates the Bearer token and stores user id in request context.
func Auth(parser TokenParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := bearerToken(r.Header.Get("Authorization"))
			if err != nil {
				handler.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			userID, err := parser.ParseAccessToken(token)
			if err != nil {
				handler.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext returns the authenticated user id from context.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	return userID, ok
}

func bearerToken(header string) (string, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", errMissingToken
	}

	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return "", errMissingToken
	}

	return token, nil
}
