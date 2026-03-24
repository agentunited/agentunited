package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agentunited/backend/internal/config"
	"github.com/agentunited/backend/internal/relay"
	"github.com/agentunited/backend/internal/repository"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}

	ctx := context.Background()
	cache, err := repository.NewCache(ctx, &cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("connect redis")
	}
	defer cache.Close()

	r := relay.NewServer(cache.Client, cfg.Relay.Domain)
	r.Start(ctx) // fetch JWKS + start refresh loop
	mux := http.NewServeMux()
	r.Routes(mux)

	srv := &http.Server{Addr: cfg.Relay.ListenAddr, Handler: mux}

	go func() {
		log.Info().Str("addr", cfg.Relay.ListenAddr).Msg("relay server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("relay server failed")
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
