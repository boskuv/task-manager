package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// App wires dependencies and runs the HTTP server.
type App struct {
	cfg    *Config
	db     *sql.DB
	server *http.Server
}

// New builds the application with an HTTP server and MySQL connection pool.
func New(cfg *Config) (*App, error) {
	db, err := openMySQL(cfg.Database)
	if err != nil {
		return nil, err
	}

	return &App{
		cfg: cfg,
		db:  db,
		server: &http.Server{
			Addr:         cfg.Addr(),
			Handler:      newRouter(),
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
		},
	}, nil
}

// Run starts the HTTP server and shuts it down gracefully on SIGINT or SIGTERM.
func (a *App) Run() error {
	errCh := make(chan error, 1)

	go func() {
		slog.Info("server starting", "addr", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("listen and serve: %w", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		closeMySQL(a.db)
		return err
	case sig := <-quit:
		slog.Info("shutdown signal received", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		closeMySQL(a.db)
		return fmt.Errorf("server shutdown: %w", err)
	}

	closeMySQL(a.db)
	slog.Info("server stopped")
	return nil
}

func newRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	return mux
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
