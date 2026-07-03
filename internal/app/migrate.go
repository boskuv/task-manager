package app

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
)

func applyMigrations(db *sql.DB, cfg DatabaseConfig) error {
	if !cfg.AutoMigrate {
		return nil
	}

	dir, err := resolveMigrationsDir(cfg.MigrationsDir)
	if err != nil {
		return err
	}

	if err := goose.SetDialect("mysql"); err != nil {
		return fmt.Errorf("set migration dialect: %w", err)
	}

	slog.Info("applying database migrations", "dir", dir)
	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func resolveMigrationsDir(dir string) (string, error) {
	if dir == "" {
		dir = "migrations"
	}

	candidates := []string{dir}
	if !filepath.IsAbs(dir) {
		if cwd, err := os.Getwd(); err == nil {
			candidates = append(candidates, filepath.Join(cwd, dir))
		}
		if exe, err := os.Executable(); err == nil {
			exeDir := filepath.Dir(exe)
			candidates = append(candidates,
				filepath.Join(exeDir, dir),
				filepath.Join(exeDir, "..", dir),
			)
		}
	}

	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		candidate = filepath.Clean(candidate)
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}

		info, err := os.Stat(candidate)
		if err != nil || !info.IsDir() {
			continue
		}
		return candidate, nil
	}

	return "", fmt.Errorf("migrations directory not found: %s", dir)
}
