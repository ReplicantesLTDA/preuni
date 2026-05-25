// Command server is the unified preuni backend monolith.
// It mounts the auth, user, and mail routers under a single chi.Router and
// serves the gateway-facing /v1/* surface plus the internal endpoints.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/preuni/pkg/logger"
	"github.com/preuni/app/internal/config"
	"github.com/preuni/app/internal/router"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)
	defer log.Sync()

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", logger.Err(err))
		os.Exit(1)
	}
	defer pool.Close()

	r := router.New(cfg, pool, log)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Info("starting monolith", logger.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server exited", logger.Err(err))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Info("shutting down monolith")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("graceful shutdown failed", logger.Err(err))
	}
}
