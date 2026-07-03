package app

import (
	"path/filepath"
	"testing"
)

func TestLoadFromYAML(t *testing.T) {
	t.Parallel()

	path := filepath.Join("testdata", "config.test.yaml")
	cfg, err := loadFrom(path)
	if err != nil {
		t.Fatalf("loadFrom: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want 9090", cfg.Server.Port)
	}
	if cfg.Database.DSN != "user:pass@tcp(localhost:3306)/testdb?parseTime=true" {
		t.Errorf("Database.DSN = %q, unexpected value", cfg.Database.DSN)
	}
	if cfg.Redis.Addr != "redis:6379" {
		t.Errorf("Redis.Addr = %q, want redis:6379", cfg.Redis.Addr)
	}
	if cfg.JWT.Secret != "test-secret" {
		t.Errorf("JWT.Secret = %q, want test-secret", cfg.JWT.Secret)
	}
	if cfg.RateLimit.RequestsPerMinute != 50 {
		t.Errorf("RateLimit.RequestsPerMinute = %d, want 50", cfg.RateLimit.RequestsPerMinute)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %q, want debug", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("Logging.Format = %q, want json", cfg.Logging.Format)
	}
	if cfg.Addr() != "127.0.0.1:9090" {
		t.Errorf("Addr() = %q, want 127.0.0.1:9090", cfg.Addr())
	}
}

func TestLoadEnvOverridesYAML(t *testing.T) {
	path := filepath.Join("testdata", "config.test.yaml")

	t.Setenv("SERVER_PORT", "3000")
	t.Setenv("DATABASE_DSN", "env:pass@tcp(db:3306)/app?parseTime=true")
	t.Setenv("JWT_SECRET", "from-env")

	cfg, err := loadFrom(path)
	if err != nil {
		t.Fatalf("loadFrom: %v", err)
	}

	if cfg.Server.Port != 3000 {
		t.Errorf("Server.Port = %d, want 3000 (env override)", cfg.Server.Port)
	}
	if cfg.Database.DSN != "env:pass@tcp(db:3306)/app?parseTime=true" {
		t.Errorf("Database.DSN = %q, want env override", cfg.Database.DSN)
	}
	if cfg.JWT.Secret != "from-env" {
		t.Errorf("JWT.Secret = %q, want from-env", cfg.JWT.Secret)
	}
}

func TestLoadMissingFile(t *testing.T) {
	t.Parallel()

	_, err := loadFrom("testdata/does-not-exist.yaml")
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}
