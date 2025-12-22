package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/lugatuic/goberus/config"
	"github.com/lugatuic/goberus/ldaps"
	"github.com/lugatuic/goberus/middleware"
	"github.com/lugatuic/goberus/server"
)

// UserClient defines the interface for LDAP operations.
type UserClient interface {
	Ping(ctx context.Context) error
	GetMemberInfo(ctx context.Context, username string) (*ldaps.MemberInfo, error)
	AddUser(ctx context.Context, u *ldaps.UserInfo) error
}

// Server composes dependencies and constructs the HTTP handler graph.
type Server struct {
	cfg    *config.Config
	logger *zap.Logger
	client UserClient
	mux    *http.ServeMux
}

// New creates a Server.
func New(cfg *config.Config, logger *zap.Logger, client UserClient) *Server {
	return &Server{
		cfg:    cfg,
		logger: logger,
		client: client,
		mux:    http.NewServeMux(),
	}
}

// Handler wires routes and middleware, returning the root handler.
func (s *Server) Handler() http.Handler {
	// Health endpoints
	s.mux.HandleFunc("/livez", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(s.logger, w, http.StatusOK, map[string]string{"status": "ok"})
	})

	s.mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := s.client.Ping(ctx); err != nil {
			s.logger.Warn("readyz.ping_failed", zap.Error(err))
			respondJSON(s.logger, w, http.StatusServiceUnavailable, map[string]string{"status": "degraded"})
			return
		}
		respondJSON(s.logger, w, http.StatusOK, map[string]string{"status": "ready"})
	})

	// Business routes
	userApp := appHandler(func(w http.ResponseWriter, r *http.Request) error {
		switch r.Method {
		case http.MethodGet:
			return server.HandleGetMember(s.client, w, r)
		case http.MethodPost:
			return server.HandleCreateMember(s.client, w, r)
		default:
			respondJSON(s.logger, w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return nil
		}
	})

	// Wrap business handler with error handling
	s.mux.Handle("/v1/member", s.makeAppHandler(userApp))

	// Mat-style middleware stack: Recover (outer), RequestID, Logger.
	// Apply to entire mux so all routes get middleware
	var handler http.Handler = s.mux
	handler = middleware.Logger(s.logger, handler)
	handler = middleware.RequestID(handler) // adds X-Request-ID if missing
	handler = middleware.Recover(s.logger, handler)

	return handler
}

// appHandler is an application handler that returns an error.
// Errors are logged and translated to HTTP responses by the adapter.
type appHandler func(http.ResponseWriter, *http.Request) error

// makeAppHandler adapts appHandler to http.Handler with sanitized error responses.
func (s *Server) makeAppHandler(fn appHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			// Log internal error with context
			s.logger.Error("handler.error", zap.Error(err), zap.String("path", r.URL.Path), zap.String("method", r.Method))
			// Return a generic 500 with JSON (do not leak internal details).
			respondJSON(s.logger, w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
	})
}

// respondJSON writes a JSON response with proper headers.
func respondJSON(logger *zap.Logger, w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		if logger != nil {
			logger.Error("respond_json.encode_error", zap.Error(err))
		}
		_, _ = w.Write([]byte("\n"))
	}
}
