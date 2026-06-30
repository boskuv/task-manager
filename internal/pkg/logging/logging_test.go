package logging

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestNewTextLogger(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger, err := NewWithWriter(Config{Level: "info", Format: "text"}, &buf)
	if err != nil {
		t.Fatalf("NewWithWriter: %v", err)
	}

	logger.Info("boot")
	if !strings.Contains(buf.String(), "boot") {
		t.Fatalf("log = %q, want boot message", buf.String())
	}
}

func TestNewJSONLogger(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger, err := NewWithWriter(Config{Level: "info", Format: "json"}, &buf)
	if err != nil {
		t.Fatalf("NewWithWriter: %v", err)
	}

	logger.Info("boot")
	if !strings.Contains(buf.String(), `"msg":"boot"`) {
		t.Fatalf("log = %q, want json boot message", buf.String())
	}
}

func TestContextHandlerAddsRequestID(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	base := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(&contextHandler{Handler: base})

	ctx := ContextWithRequestID(context.Background(), "req-123")
	logger.InfoContext(ctx, "handled")

	if !strings.Contains(buf.String(), `"request_id":"req-123"`) {
		t.Fatalf("log = %q, want request_id field", buf.String())
	}
}

func TestNewRejectsUnknownFormat(t *testing.T) {
	t.Parallel()

	_, err := NewWithWriter(Config{Level: "info", Format: "yaml"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestNewRejectsUnknownLevel(t *testing.T) {
	t.Parallel()

	_, err := NewWithWriter(Config{Level: "trace", Format: "text"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for unsupported level")
	}
}
