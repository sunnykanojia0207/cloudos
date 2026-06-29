package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudos/cloudos/packages/logging"
)

// contextKey is an unexported type used for context keys to avoid collisions.
type contextKey string

// RequestIDKey is the context key for the request ID.
const RequestIDKey contextKey = "request_id"

// GetRequestID extracts the request ID from the context, returning an empty
// string if none is present.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// generateRequestID produces a 16-byte hex string suitable for tracing requests.
func generateRequestID() string {
	b := make([]byte, 16)
	rand.Read(b) //nolint:errcheck
	return hex.EncodeToString(b)
}

// ---------------------------------------------------------------------------
// Middleware
// ---------------------------------------------------------------------------

// Middleware is a function that wraps an http.Handler.
type Middleware func(next http.Handler) http.Handler

// RequestIDMiddleware injects a unique request ID into every request's context
// and sets the X-Request-Id response header.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = generateRequestID()
		}
		ctx := context.WithValue(r.Context(), RequestIDKey, id)
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RecoveryMiddleware catches panics in downstream handlers, logs them, and
// returns a 500 Internal Server Error instead of crashing the process.
func RecoveryMiddleware(log *logging.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered",
						"path", r.URL.Path,
						"method", r.Method,
						"panic", fmt.Sprintf("%v", rec),
					)
					InternalError(w, "INTERNAL_ERROR",
						"An unexpected error occurred. Please try again later.")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware logs every request with its method, path, status code,
// duration, and request ID.
func LoggingMiddleware(log *logging.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lrw := newLoggingResponseWriter(w)
			next.ServeHTTP(lrw, r)

			args := []interface{}{
				"method", r.Method,
				"path", r.URL.Path,
				"remote", r.RemoteAddr,
				"status", lrw.statusCode,
				"duration", time.Since(start).String(),
				"request_id", GetRequestID(r.Context()),
			}

			// Log at warn for 5xx, info otherwise.
			if lrw.statusCode >= 500 {
				log.Warn("request completed", args...)
			} else {
				log.Info("request completed", args...)
			}
		})
	}
}

// loggingResponseWriter wraps http.ResponseWriter to capture the status code.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

// WriteHeader captures the status code before delegating to the wrapped writer.
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
