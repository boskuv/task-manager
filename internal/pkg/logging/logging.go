package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Config controls slog output format and minimum level.
type Config struct {
	Level  string
	Format string
}

// New builds a structured slog logger writing to stdout.
func New(cfg Config) (*slog.Logger, error) {
	return NewWithWriter(cfg, os.Stdout)
}

// NewWithWriter builds a structured slog logger writing to w.
func NewWithWriter(cfg Config, w io.Writer) (*slog.Logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler

	switch strings.ToLower(strings.TrimSpace(cfg.Format)) {
	case "json":
		handler = slog.NewJSONHandler(w, opts)
	case "text", "":
		handler = slog.NewTextHandler(w, opts)
	default:
		return nil, fmt.Errorf("unsupported log format %q", cfg.Format)
	}

	return slog.New(&contextHandler{Handler: handler}), nil
}

func parseLevel(raw string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info", "":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unsupported log level %q", raw)
	}
}
