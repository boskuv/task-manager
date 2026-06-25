package app

import (
	"net/http"

	"github.com/boskuv/task-manager/internal/handler"
)

func newRouter(auth *handler.AuthHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	if auth != nil {
		mux.HandleFunc("POST /api/v1/register", auth.Register)
		mux.HandleFunc("POST /api/v1/login", auth.Login)
	}

	return mux
}
