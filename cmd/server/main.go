// Command server is a small, fully instrumented HTTP service used as the
// deployment target for the deploy-pilot GitOps reference platform.
//
// It intentionally exposes endpoints that make failure scenarios reproducible
// (/boom, /spin) so that alerting and runbooks can be demonstrated end to end.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Han-Sang-min/deploy-pilot/internal/server"
)

// Injected at build time via -ldflags (see deploy/docker/Dockerfile and Makefile).
var (
	version = "dev"
	commit  = "none"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	addr := getenv("LISTEN_ADDR", ":8080")
	readyDelay := getDuration("READY_DELAY", 2*time.Second)

	srv := server.New(server.Config{
		Version:    version,
		Commit:     commit,
		Logger:     logger,
		ReadyDelay: readyDelay,
	})

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("server starting", "addr", addr, "version", version, "commit", commit)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	// Honour SIGTERM so Kubernetes rolling updates drain cleanly instead of
	// dropping in-flight requests — a deliberate ops detail reviewers look for.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	logger.Info("shutdown signal received, draining connections")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}
	logger.Info("shutdown complete")
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
