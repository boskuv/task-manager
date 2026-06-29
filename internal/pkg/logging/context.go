package logging

import (
	"context"
	"log/slog"
)

type contextKey struct{}

var requestIDKey contextKey

// ContextWithRequestID stores the request id in context.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext returns the request id from context.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(requestIDKey).(string)
	return requestID, ok && requestID != ""
}

type contextHandler struct {
	slog.Handler
}

func (h *contextHandler) Handle(ctx context.Context, record slog.Record) error {
	if requestID, ok := RequestIDFromContext(ctx); ok {
		record.AddAttrs(slog.String("request_id", requestID))
	}
	return h.Handler.Handle(ctx, record)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{Handler: h.Handler.WithGroup(name)}
}
