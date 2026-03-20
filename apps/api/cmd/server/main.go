package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agentunited/backend/internal/api"
	"github.com/agentunited/backend/internal/config"
	"github.com/agentunited/backend/internal/relay"
	"github.com/agentunited/backend/internal/repository"
	"github.com/agentunited/backend/internal/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Setup structured logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: zerolog.SyncWriter(os.Stdout), TimeFormat: time.RFC3339})

	log.Info().Msg("starting Agent United backend")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	ctx := context.Background()

	// Connect to database
	db, err := repository.NewDB(ctx, &cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(ctx, "migrations"); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}

	// Connect to Redis
	cache, err := repository.NewCache(ctx, &cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}
	defer cache.Close()

	// Initialize file storage
	if err := utils.InitializeStorage(); err != nil {
		log.Fatal().Err(err).Msg("failed to initialize file storage")
	}
	log.Info().Msg("file storage initialized")

	// Initialize embedded relay manager (runtime-updatable)
	relayManager := relay.NewManager(cfg.Relay.DeploymentMode, cfg.Relay.ServerURL, cfg.Relay.LocalAPIURL, cfg.Relay.Token)
	relay.SetGlobalManager(relayManager)
	relayCtx, relayCancel := context.WithCancel(context.Background())
	defer relayCancel()
	relayManager.Start(relayCtx)
	if cfg.Relay.DeploymentMode == "tunnel" {
		log.Info().Str("relay_server", cfg.Relay.ServerURL).Msg("tunnel mode configured")
	}

	// Setup router
	router := api.NewRouter(db, cache, cfg)

	// HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: router,
		// Do not set a global write timeout; SSE streams are long-lived.
		ReadTimeout: 15 * time.Second,
		IdleTimeout: 60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info().Str("port", cfg.Server.Port).Msg("HTTP server listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server exited")
}
