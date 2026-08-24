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
	"github.com/preuni/app/internal/config"
	essayrepo "github.com/preuni/app/internal/essay/repository"
	gamificationrepo "github.com/preuni/app/internal/gamification/repository"
	"github.com/preuni/app/internal/router"
	"github.com/preuni/pkg/logger"
)

// reconcileInterval is how often the monolith polls correction_jobs for
// results (contracts/internal-bridge.md). The correction worker also wakes
// the monolith isn't notified directly — this poll is the reconciliation
// side of the async bridge, independent of the worker's own LISTEN/NOTIFY.
const reconcileInterval = 5 * time.Second

// runEssayReconciler periodically pulls completed/failed correction_jobs
// results back onto their originating essay submissions until ctx is
// cancelled.
func runEssayReconciler(ctx context.Context, repo *essayrepo.Repository, log *logger.Logger) {
	ticker := time.NewTicker(reconcileInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := repo.ReconcileOnce(ctx)
			if err != nil {
				log.Error("essay reconciler tick failed", logger.Err(err))
				continue
			}
			if n > 0 {
				log.Info("essay reconciler processed submissions", logger.Int("count", n))
			}
		}
	}
}

// weekCloseCheckInterval is how often the monolith checks whether the
// previous UTC week needs closing (ranking/tier promotion). WeekClose
// itself is idempotent (repository.WeekClose skips already-closed
// tier-weeks), so a generous, infrequent check is enough — this isn't a
// latency-sensitive path the way essay reconciliation is.
const weekCloseCheckInterval = 1 * time.Hour

// runWeekCloseJob periodically finalizes the most recently completed week
// (domain.TierMovement promotion/demotion, medal awards) until ctx is
// cancelled.
func runWeekCloseJob(ctx context.Context, repo *gamificationrepo.Repository, log *logger.Logger) {
	ticker := time.NewTicker(weekCloseCheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := repo.WeekClose(ctx, time.Now().UTC())
			if err != nil {
				log.Error("week-close tick failed", logger.Err(err))
				continue
			}
			if n > 0 {
				log.Info("week-close processed tiers", logger.Int("count", n))
			}
		}
	}
}

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

	r, essayRepo, gamificationRepo := router.New(cfg, pool, log)

	reconcileCtx, stopReconciler := context.WithCancel(context.Background())
	defer stopReconciler()
	go runEssayReconciler(reconcileCtx, essayRepo, log)
	go runWeekCloseJob(reconcileCtx, gamificationRepo, log)

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
