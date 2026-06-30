//go:build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
)

const (
	mysqlDatabase = "taskmanager"
	mysqlUser     = "taskmanager"
	mysqlPassword = "taskmanager"
)

var (
	mysqlOnce     sync.Once
	mysqlDB       *sql.DB
	mysqlCleanup  func()
	mysqlSetupErr error
	mysqlMu       sync.Mutex
)

// MySQL returns a shared MySQL connection with migrations applied.
func MySQL(t *testing.T) *sql.DB {
	t.Helper()

	mysqlOnce.Do(func() {
		mysqlDB, mysqlCleanup, mysqlSetupErr = startMySQL()
	})
	if mysqlSetupErr != nil {
		t.Fatalf("start mysql: %v", mysqlSetupErr)
	}

	mysqlMu.Lock()
	CleanMySQLTables(t, mysqlDB)
	t.Cleanup(func() {
		CleanMySQLTables(t, mysqlDB)
		mysqlMu.Unlock()
	})

	return mysqlDB
}

func startMySQL() (*sql.DB, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := tcmysql.Run(ctx,
		"mysql:8.0",
		tcmysql.WithDatabase(mysqlDatabase),
		tcmysql.WithUsername(mysqlUser),
		tcmysql.WithPassword(mysqlPassword),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("run container: %w", err)
	}

	dsn, err := container.ConnectionString(ctx, "parseTime=true", "charset=utf8mb4")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, fmt.Errorf("connection string: %w", err)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	pingCtx, pingCancel := context.WithTimeout(ctx, 30*time.Second)
	defer pingCancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		return nil, nil, fmt.Errorf("ping database: %w", err)
	}

	if err := runMigrations(db); err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		return nil, nil, fmt.Errorf("run migrations: %w", err)
	}

	cleanup := func() {
		_ = db.Close()
		termCtx, termCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer termCancel()
		_ = container.Terminate(termCtx)
	}

	return db, cleanup, nil
}

func runMigrations(db *sql.DB) error {
	if err := goose.SetDialect("mysql"); err != nil {
		return err
	}
	return goose.Up(db, migrationsDir())
}

func migrationsDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "migrations")
}

// CleanMySQLTables removes rows from application tables between tests.
func CleanMySQLTables(t *testing.T, db *sql.DB) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, `SET FOREIGN_KEY_CHECKS = 0`); err != nil {
		t.Fatalf("disable foreign key checks: %v", err)
	}

	tables := []string{
		"task_history",
		"task_comments",
		"tasks",
		"team_members",
		"teams",
		"users",
	}
	for _, table := range tables {
		if _, err := db.ExecContext(ctx, "TRUNCATE TABLE "+table); err != nil {
			t.Fatalf("truncate %s: %v", table, err)
		}
	}

	if _, err := db.ExecContext(ctx, `SET FOREIGN_KEY_CHECKS = 1`); err != nil {
		t.Fatalf("enable foreign key checks: %v", err)
	}
}
