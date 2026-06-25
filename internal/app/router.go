package app

import (
	"net/http"

	"github.com/boskuv/task-manager/internal/handler"
	"github.com/boskuv/task-manager/internal/handler/middleware"
	jwtmanager "github.com/boskuv/task-manager/internal/pkg/jwt"
)

type routerDeps struct {
	auth       *handler.AuthHandler
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
