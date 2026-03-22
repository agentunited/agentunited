package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	port := getenv("PORT", "8080")
	secret := os.Getenv("STRIPE_WEBHOOK_SECRET_PROD")
	dbURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")

	if secret == "" || dbURL == "" || redisURL == "" {
		log.Fatal().Msg("missing required env: STRIPE_WEBHOOK_SECRET_PROD, DATABASE_URL, REDIS_URL")
	}

	store, err := NewPostgresStore(dbURL)
	if err != nil {
		log.Fatal().Err(err).Msg("postgres init failed")
	}
	defer store.Close()

	cache, err := NewRedisCache(redisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("redis init failed")
	}
	defer cache.Close()

	email := NewSendgridEmailSender(os.Getenv("SENDGRID_API_KEY"), getenv("EMAIL_FROM", "hello@agentunited.ai"))
	h := NewHandler(secret, store, cache, email)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /webhook/stripe", h.StripeWebhook)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("billing-webhook listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
