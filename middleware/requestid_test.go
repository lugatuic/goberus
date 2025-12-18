package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"

	"github.com/lugatuic/goberus/middleware"
)

func TestRequestID(t *testing.T) {
	t.Run("generates new request ID when missing", func(t *testing.T) {
		is := is.New(t)
		handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Request ID should be present in request header
			id := r.Header.Get("X-Request-ID")
			is.True(id != "")
			is.Equal(len(id), 32) // hex encoded 16 bytes = 32 chars
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		// Response should echo the request ID
		is.True(rr.Header().Get("X-Request-ID") != "")
		is.Equal(len(rr.Header().Get("X-Request-ID")), 32)
	})

	t.Run("preserves existing request ID", func(t *testing.T) {
		is := is.New(t)
		existingID := "existing-request-id-12345"
		handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-ID")
			is.Equal(id, existingID)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Request-ID", existingID)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		// Response should echo the existing request ID
		is.Equal(rr.Header().Get("X-Request-ID"), existingID)
	})
}
