package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agentunited/backend/internal/config"
	"github.com/agentunited/backend/internal/relay"
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

	mgr := relay.NewManager("tunnel", cfg.Relay.ServerURL, cfg.Relay.LocalAPIURL, cfg.Relay.Token)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mgr.Start(ctx)

	log.Info().Str("relay_server", cfg.Relay.ServerURL).Msg("relay client started")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
