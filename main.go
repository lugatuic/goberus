package main

import (
	"fmt"
	"net/http"
	"os"

	"go.uber.org/zap"

	"github.com/lugatuic/goberus/config"
	"github.com/lugatuic/goberus/ldaps"
	"github.com/lugatuic/goberus/middleware"
	"github.com/lugatuic/goberus/server"
)

// appHandler is an application handler that returns an error. Returned
// errors are considered server errors and are logged/translated to 500 by
// the adapter in main. Defined at package level so it can be used in tests.
type appHandler func(w http.ResponseWriter, r *http.Request) error

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

	client, err := ldaps.NewClient(cfg, logger)
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
			return server.HandleGetMember(client, w, r)
		case http.MethodPost:
			return server.HandleCreateMember(client, w, r)
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
