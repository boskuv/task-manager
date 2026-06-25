package handler

import "context"

type contextKey string

const userIDKey contextKey = "userID"

// ContextWithUserID stores the authenticated user id in context.
func ContextWithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext returns the authenticated user id from context.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	return userID, ok
}
