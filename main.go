package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/lugatuic/goberus/config"
	"github.com/lugatuic/goberus/handlers"
	"github.com/lugatuic/goberus/ldaps"
	"github.com/lugatuic/goberus/middleware"
)

// appHandler is an application handler that returns an error. Returned
// errors are considered server errors and are logged/translated to 500 by
// the adapter in main. Defined at package level so it can be used in tests.
type appHandler func(w http.ResponseWriter, r *http.Request) error

// handleGetMember handles the GET /v1/member logic. It returns an error for
// server-side failures; client errors are written directly to the ResponseWriter
// and return nil so the adapter doesn't treat them as 500s.
func handleGetMember(client *ldaps.Client, w http.ResponseWriter, r *http.Request) error {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "missing username parameter", http.StatusBadRequest)
		return nil
	}

	ctxTimeout, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	info, err := client.GetMemberInfo(ctxTimeout, username)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		return err
	}
	return nil
}

// handleCreateMember handles the POST /v1/member logic. Same error convention as above.
func handleCreateMember(client *ldaps.Client, w http.ResponseWriter, r *http.Request) error {
	defer func() {
		if err := r.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close request body: %v\n", err)
		}
	}()
	var u ldaps.UserInfo
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		// client error
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return nil
	}
	if err := handlers.SanitizeUser(&u); err != nil {
		http.Error(w, "invalid input: "+err.Error(), http.StatusBadRequest)
		return nil
	}

	ctxTimeout, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	if err := client.AddUser(ctxTimeout, &u); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "created"}); err != nil {
		return err
	}
	return nil
}

func main() {
	// initialize structured logger early so we can log config errors
	logger, lerr := zap.NewProduction()
	if lerr != nil {
		panic("failed to initialize logger: " + lerr.Error())
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "logger sync failed: %v\n", err)
		}
	}()

	cfg, err := config.LoadFromEnv()
	if err != nil {
		logger.Fatal("config load failed", zap.Error(err))
	}

	client, err := ldaps.NewClient(cfg)
	if err != nil {
		logger.Fatal("ldaps client init failed", zap.Error(err))
	}

	mux := http.NewServeMux()

	// makeAppHandler adapts an appHandler into an http.Handler, logging errors
	// using the provided logger and returning HTTP 500 for unexpected failures.
	makeAppHandler := func(logger *zap.Logger, fn appHandler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := fn(w, r); err != nil {
				logger.Error("handler error", zap.Error(err))
				http.Error(w, "internal server error: "+err.Error(), http.StatusInternalServerError)
			}
		})
	}

	userApp := appHandler(func(w http.ResponseWriter, r *http.Request) error {
		switch r.Method {
		case http.MethodGet:
			return handleGetMember(client, w, r)
		case http.MethodPost:
			return handleCreateMember(client, w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return nil
		}
	})

	// constructor-style middleware wrapping: makeAppHandler then Recover outermost
	userHandler := makeAppHandler(logger, userApp)
	wrappedUserHandler := middleware.Recover(logger, middleware.Logger(logger, userHandler))
	mux.Handle("/v1/member", wrappedUserHandler)

	logger.Info("listening", zap.String("addr", cfg.BindAddr))
	if err = http.ListenAndServe(cfg.BindAddr, mux); err != nil {
		logger.Fatal("http server failed", zap.Error(err))
	}
}
