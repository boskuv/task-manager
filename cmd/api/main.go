package main

import (
	"log"
	"log/slog"

	"github.com/boskuv/task-manager/internal/app"
	"github.com/boskuv/task-manager/internal/pkg/logging"
)

func main() {
	cfg, err := app.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger, err := logging.New(logging.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
	})
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	slog.SetDefault(logger)

	application, err := app.New(cfg, logger)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}
	if err := application.Run(); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
