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

	"github.com/redis/go-redis/v9"

	"github.com/boskuv/task-manager/internal/handler"
	jwtmanager "github.com/boskuv/task-manager/internal/pkg/jwt"
	mysqlrepo "github.com/boskuv/task-manager/internal/repository/mysql"
	authuc "github.com/boskuv/task-manager/internal/usecase/auth"
)

// App wires dependencies and runs the HTTP server.
type App struct {
	cfg    *Config
	db     *sql.DB
	redis  *redis.Client
	server *http.Server
}

// New builds the application with an HTTP server, MySQL pool, and Redis client.
func New(cfg *Config) (*App, error) {
	db, err := openMySQL(cfg.Database)
	if err != nil {
		return nil, err
	}

	rdb, err := openRedis(cfg.Redis)
	if err != nil {
		closeMySQL(db)
		return nil, err
	}

	userRepo := mysqlrepo.NewUserRepo(db)
	jwtManager := jwtmanager.NewManager(cfg.JWT.Secret, cfg.JWT.AccessTTL)
	authService := authuc.NewService(userRepo, jwtManager)
	authHandler := handler.NewAuthHandler(authService)

	return &App{
		cfg:   cfg,
		db:    db,
		redis: rdb,
		server: &http.Server{
			Addr:         cfg.Addr(),
			Handler: newRouter(routerDeps{
				auth:       authHandler,
				jwtManager: jwtManager,
			}),
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
		a.close()
		return err
	case sig := <-quit:
		slog.Info("shutdown signal received", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.close()
		return fmt.Errorf("server shutdown: %w", err)
	}

	a.close()
	slog.Info("server stopped")
	return nil
}

func (a *App) close() {
	closeRedis(a.redis)
	closeMySQL(a.db)
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
