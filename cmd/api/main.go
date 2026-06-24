package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/boskuv/task-manager/internal/app"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	cfg, err := app.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}
	if err := application.Run(); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
