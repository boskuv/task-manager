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
	"github.com/boskuv/task-manager/internal/pkg/circuitbreaker"
	"github.com/boskuv/task-manager/internal/pkg/email"
	jwtmanager "github.com/boskuv/task-manager/internal/pkg/jwt"
	mysqlrepo "github.com/boskuv/task-manager/internal/repository/mysql"
	analyticsuc "github.com/boskuv/task-manager/internal/usecase/analytics"
	authuc "github.com/boskuv/task-manager/internal/usecase/auth"
	taskuc "github.com/boskuv/task-manager/internal/usecase/task"
	teamuc "github.com/boskuv/task-manager/internal/usecase/team"
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
	teamRepo := mysqlrepo.NewTeamRepo(db)
	taskRepo := mysqlrepo.NewTaskRepo(db)
	taskHistoryRepo := mysqlrepo.NewTaskHistoryRepo(db)
	analyticsRepo := mysqlrepo.NewAnalyticsRepo(db)
	jwtManager := jwtmanager.NewManager(cfg.JWT.Secret, cfg.JWT.AccessTTL)
	authService := authuc.NewService(userRepo, jwtManager)
	emailBreaker := circuitbreaker.New(circuitbreaker.Config{})
	inviteMailer := email.NewMockService(emailBreaker, slog.Default())
	teamService := teamuc.NewService(teamRepo, userRepo, inviteMailer)
	taskService := taskuc.NewService(taskRepo, teamRepo, taskHistoryRepo)
	analyticsService := analyticsuc.NewService(analyticsRepo, teamRepo)
	authHandler := handler.NewAuthHandler(authService)
	teamHandler := handler.NewTeamHandler(teamService)
	taskHandler := handler.NewTaskHandler(taskService)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService)

	return &App{
		cfg:   cfg,
		db:    db,
		redis: rdb,
		server: &http.Server{
			Addr:         cfg.Addr(),
			Handler: newRouter(			routerDeps{
				auth:       authHandler,
				teams:      teamHandler,
				tasks:      taskHandler,
				analytics:  analyticsHandler,
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
