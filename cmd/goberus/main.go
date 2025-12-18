package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/lugatuic/goberus/config"
	"github.com/lugatuic/goberus/internal/httpserver"
	"github.com/lugatuic/goberus/ldaps"
)

func main() {
	// Initialize structured logger early so we can log config errors.
	logger, lerr := zap.NewProduction()
	if lerr != nil {
		panic("failed to initialize logger: " + lerr.Error())
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "logger sync failed: %v\n", err)
		}
	}()

	// Load configuration (consider flags + env overrides as a next step).
	cfg, err := config.LoadFromEnv()
	if err != nil {
		logger.Fatal("config load failed", zap.Error(err))
	}

	// Initialize dependency clients.
	client, err := ldaps.NewClient(cfg, logger)
	if err != nil {
		logger.Fatal("ldaps client init failed", zap.Error(err))
	}

	// Build HTTP handler using Mat Ryer–style server composition.
	s := httpserver.New(cfg, logger, client)
	handler := s.Handler()

	// Harden the HTTP server with sensible timeouts.
	srv := &http.Server{
		Addr:              cfg.BindAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start server in background.
	errCh := make(chan error, 1)
	go func() {
		logger.Info("http.listen", zap.String("addr", cfg.BindAddr))
		errCh <- srv.ListenAndServe()
	}()

	// Graceful shutdown on signals.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Info("shutdown.signal", zap.String("signal", sig.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("shutdown.error", zap.Error(err))
		} else {
			logger.Info("shutdown.complete")
		}
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			logger.Fatal("http.server.failed", zap.Error(err))
		}
	}
}
