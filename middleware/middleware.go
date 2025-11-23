package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Logger logs requests around the provided handler.
type responseLogger struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *responseLogger) WriteHeader(statusCode int) {
	if r.status == 0 {
		r.status = statusCode
	}
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseLogger) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}

// Logger logs request timing, status, and response size.
func Logger(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger.Info("request.start",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote", r.RemoteAddr),
		)
		lrw := &responseLogger{ResponseWriter: w}
		next.ServeHTTP(lrw, r)
		status := lrw.status
		if status == 0 {
			status = http.StatusOK
		}
		logger.Info("request.done",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Duration("duration", time.Since(start)),
			zap.Int("status", status),
			zap.Int("bytes", lrw.size),
		)
	})
}

// Recover recovers from panics in the wrapped handler and responds with HTTP 500.
func Recover(logger *zap.Logger, next http.Handler) http.Handler {
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
