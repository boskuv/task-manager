package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/boskuv/task-manager/internal/pkg/logging"
)

const RequestIDHeader = "X-Request-ID"

// RequestID propagates or generates a request id and stores it in context.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := strings.TrimSpace(r.Header.Get(RequestIDHeader))
		if requestID == "" {
			var err error
			requestID, err = newRequestID()
			if err != nil {
				requestID = "unknown"
			}
		}

		w.Header().Set(RequestIDHeader, requestID)
		ctx := logging.ContextWithRequestID(r.Context(), requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext returns the request id attached to the request context.
func RequestIDFromContext(r *http.Request) (string, bool) {
	return logging.RequestIDFromContext(r.Context())
}

func newRequestID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
