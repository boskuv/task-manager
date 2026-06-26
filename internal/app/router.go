package app

import (
	"net/http"

	"github.com/boskuv/task-manager/internal/handler"
	"github.com/boskuv/task-manager/internal/handler/middleware"
	jwtmanager "github.com/boskuv/task-manager/internal/pkg/jwt"
)

type routerDeps struct {
	auth       *handler.AuthHandler
	teams      *handler.TeamHandler
	tasks      *handler.TaskHandler
	jwtManager *jwtmanager.Manager
}

func newRouter(deps routerDeps) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	if deps.auth != nil {
		mux.HandleFunc("POST /api/v1/register", deps.auth.Register)
		mux.HandleFunc("POST /api/v1/login", deps.auth.Login)
	}

	if deps.jwtManager != nil {
		protected := middleware.Auth(deps.jwtManager)
		mux.Handle("GET /api/v1/me", protected(http.HandlerFunc(meHandler)))

		if deps.teams != nil {
			mux.Handle("POST /api/v1/teams", protected(http.HandlerFunc(deps.teams.Create)))
			mux.Handle("GET /api/v1/teams", protected(http.HandlerFunc(deps.teams.List)))
			mux.Handle("POST /api/v1/teams/{id}/invite", protected(http.HandlerFunc(deps.teams.Invite)))
		}

		if deps.tasks != nil {
			mux.Handle("POST /api/v1/tasks", protected(http.HandlerFunc(deps.tasks.Create)))
			mux.Handle("GET /api/v1/tasks", protected(http.HandlerFunc(deps.tasks.List)))
			mux.Handle("PUT /api/v1/tasks/{id}", protected(http.HandlerFunc(deps.tasks.Update)))
			mux.Handle("GET /api/v1/tasks/{id}/history", protected(http.HandlerFunc(deps.tasks.History)))
		}
	}

	return mux
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		handler.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	handler.WriteJSON(w, http.StatusOK, map[string]int64{"user_id": userID})
}
