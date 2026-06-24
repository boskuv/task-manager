package app

import (
	"strings"
	"testing"
)

func TestOpenMySQLInvalidDSN(t *testing.T) {
	t.Parallel()

	_, err := openMySQL(DatabaseConfig{
		DSN:             "invalid-dsn",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 0,
	})
	if err == nil {
		t.Fatal("expected error for invalid dsn")
	}
	if !strings.Contains(err.Error(), "open database") && !strings.Contains(err.Error(), "ping database") {
		t.Errorf("error = %q, want open or ping failure", err.Error())
	}
}
