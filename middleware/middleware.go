package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Middleware helpers are provided as constructor-style wrappers.
// Use `Logger(logger, next)` and `Recover(logger, next)` to wrap handlers.

// _logger returns a func http.Handler  that logs basic request information and duration.
func _logger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			logger.Info("request.start",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote", r.RemoteAddr),
			)
			next.ServeHTTP(w, r)
			logger.Info("request.done",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}

// _recover returns a func http.Handler that recovers from panics in handlers and returns HTTP 500.
func _recover(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered", zap.Any("panic", rec))
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Logger is a constructor-style wrapper equivalent to Logger(logger)(next).
// This lets callers write: Logger(logger, RecoverWrap(logger, mux))
func Logger(logger *zap.Logger, next http.Handler) http.Handler {
	return _logger(logger)(next)
}

// Recover is a constructor-style wrapper equivalent to Recover(logger)(next).
func Recover(logger *zap.Logger, next http.Handler) http.Handler {
	return _recover(logger)(next)
}
