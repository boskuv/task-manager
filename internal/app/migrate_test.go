package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveMigrationsDirFromRepoRoot(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}
	t.Chdir(root)

	dir, err := resolveMigrationsDir("migrations")
	if err != nil {
		t.Fatalf("resolveMigrationsDir: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "00001_create_users.sql")); err != nil {
		t.Fatalf("expected users migration in %s: %v", dir, err)
	}
}

func TestResolveMigrationsDirMissing(t *testing.T) {
	t.Chdir(t.TempDir())

	_, err := resolveMigrationsDir("missing-migrations")
	if err == nil {
		t.Fatal("expected error for missing migrations dir")
	}
}

func TestApplyMigrationsDisabled(t *testing.T) {
	t.Parallel()

	err := applyMigrations(nil, DatabaseConfig{AutoMigrate: false})
	if err != nil {
		t.Fatalf("applyMigrations: %v", err)
	}
}
