package api

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/agentunited/backend/internal/api/handlers"
	mw "github.com/agentunited/backend/internal/api/middleware"
	"github.com/agentunited/backend/internal/config"
	"github.com/agentunited/backend/internal/realtime"
	"github.com/agentunited/backend/internal/repository"
	"github.com/agentunited/backend/internal/service"
	"github.com/agentunited/backend/pkg/billing"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog/log"
)

// NewRouter creates and configures the Chi router
func NewRouter(db *repository.DB, cache *repository.Cache, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(loggerMiddleware)
	r.Use(middleware.Recoverer)
	// Timeout for non-WebSocket routes (WS connections are long-lived)
	r.Use(func(next http.Handler) http.Handler {
		timeoutHandler := middleware.Timeout(60 * time.Second)(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/ws" {
				next.ServeHTTP(w, r)
				return
			}
			timeoutHandler.ServeHTTP(w, r)
		})
	})

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*", "https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check handler
	healthHandler := NewHealthHandler(db, cache)
	r.Get("/health", healthHandler.Check)

	// Static file serving for uploads
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "data", "uploads"))
	r.Handle("/uploads/*", http.StripPrefix("/uploads", http.FileServer(filesDir)))

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	channelRepo := repository.NewChannelRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	agentRepo := repository.NewAgentRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	webhookRepo := repository.NewWebhookRepository(db)
	integrationRepo := repository.NewIntegrationRepository(db)
	subscriptionRepo := repository.NewSubscriptionRepository(db)
	inviteRepo := repository.NewInviteRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret)
	channelService := service.NewChannelService(channelRepo)
	messageService := service.NewMessageService(messageRepo, channelRepo, agentRepo)
	realtimeEngine := realtime.New(cfg.Centrifugo)
	agentService := service.NewAgentService(agentRepo)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, agentRepo)
	webhookService := service.NewWebhookService(webhookRepo, agentRepo)
	integrationService := service.NewIntegrationService(integrationRepo, webhookService)
	billingProvider := billing.NewStub(cfg.Stripe.SecretKey != "" && cfg.Stripe.PriceIDPro != "")
	billingService := service.NewBillingService(subscriptionRepo, userRepo, billingProvider, cfg.Stripe.WebhookSecret, cfg.Stripe.PriceIDPro, cfg.Stripe.PriceIDTeam)
	bootstrapService := service.NewBootstrapService(userRepo, agentRepo, apiKeyRepo, inviteRepo, channelRepo, subscriptionRepo, cfg.JWT.Secret, "https://agentunited.ai", cfg.Relay.Domain, cache.Client)
	inviteService := service.NewInviteService(userRepo, inviteRepo, subscriptionRepo, cfg.JWT.Secret, "http://localhost:3001")
	relayService := service.NewRelayService(subscriptionRepo, cfg.Relay.Domain)

	// Initialize WebSocket hub first (needed by message handler)
	hub := handlers.NewHub()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	meHandler := handlers.NewMeHandler(authService)
	channelHandler := handlers.NewChannelHandler(channelService)
	messageHandler := handlers.NewMessageHandler(messageService, webhookService, hub, realtimeEngine, integrationService)
	integrationHandler := handlers.NewIntegrationHandler(integrationService)
	billingHandler := handlers.NewBillingHandler(billingService)
	agentHandler := handlers.NewAgentHandler(agentService)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyService)
	webhookHandler := handlers.NewWebhookHandler(webhookService)
	bootstrapHandler := handlers.NewBootstrapHandler(bootstrapService, cache.Client)
	inviteHandler := handlers.NewInviteHandler(inviteService)
	relayHandler := handlers.NewRelayHandler(relayService)
	usersHandler := handlers.NewUsersHandler(userRepo)
	pairingHandler := handlers.NewPairingHandler()
	centrifugoHandler := handlers.NewCentrifugoHandler(realtimeEngine, channelService)
	wsHandler := handlers.NewWebSocketHandlerV2(messageService, channelService, cfg.JWT.Secret, hub)

	// WebSocket endpoint — no timeout middleware (long-lived connection)
	r.Handle("/ws", wsHandler)

	// Public routes (no authentication required)
	r.Post("/api/v1/bootstrap", bootstrapHandler.ServeHTTP)
	r.Get("/api/v1/invite", inviteHandler.ValidateInvite)
	r.Post("/api/v1/invite/accept", inviteHandler.AcceptInvite)

	// Pairing routes for M5 Tunneling (Publicly accessible from local web UI)
	r.Get("/api/v1/pairing/code", pairingHandler.GetCode)
	r.Get("/api/v1/pairing/verify", pairingHandler.VerifyCode)

	// Public authentication routes
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	// Public billing webhook route
	r.Post("/api/v1/billing/webhook", billingHandler.Webhook)

	// Public integration inbound webhook route
	r.Post("/api/v1/webhooks/integration/*", handlers.IntegrationInboundHandler)

	// Protected API v1 routes (require JWT or API key authentication)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(mw.Auth(cfg.JWT.Secret, apiKeyRepo, agentRepo, userRepo))

		// Current user profile routes
		r.Get("/me", meHandler.GetMe)
		r.Put("/me", meHandler.UpdateMe)
		r.Post("/me/password", meHandler.ChangePassword)

		// Users/invites routes
		r.Get("/users", usersHandler.List)
		r.Post("/invites", inviteHandler.CreateInvite)

		// Integrations routes
		r.Get("/integrations", integrationHandler.List)
		r.Post("/integrations", integrationHandler.Create)
		r.Delete("/integrations/{id}", integrationHandler.Delete)

		// Billing routes
		r.Post("/billing/checkout", billingHandler.Checkout)
		r.Get("/billing/checkout", billingHandler.Checkout)
		r.Post("/billing/portal", billingHandler.Portal)
		r.Get("/billing/portal", billingHandler.Portal)
		r.Get("/billing/status", billingHandler.Status)

		// Relay routes
		r.Get("/relay/status", relayHandler.Status)

		// Admin routes
		r.Route("/admin", func(r chi.Router) {
			r.Post("/pairing", pairingHandler.AdminPairing)
			r.Get("/tunnel/status", pairingHandler.TunnelStatus)
			r.Post("/tunnel/subdomain/check", pairingHandler.SubdomainCheck)
			r.Post("/tunnel/subdomain/claim", pairingHandler.SubdomainClaim)
			r.Get("/pairing/status", pairingHandler.PairingStatus)
		})

		// Centrifugo auth/ops routes
		r.Route("/realtime/centrifugo", func(r chi.Router) {
			r.Post("/subscribe-token", centrifugoHandler.SubscribeToken)
			r.Post("/refresh-token", centrifugoHandler.RefreshToken)
			r.Get("/presence", centrifugoHandler.Presence)
			r.Get("/history", centrifugoHandler.History)
		})

		// Agent routes
		r.Route("/agents", func(r chi.Router) {
			r.Post("/", agentHandler.Create)
			r.Get("/", agentHandler.List)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", agentHandler.Get)
				r.Patch("/", agentHandler.Update)
				r.Delete("/", agentHandler.Delete)

				// API key routes
				r.Post("/keys", apiKeyHandler.Create)
				r.Get("/keys", apiKeyHandler.List)
				r.Delete("/keys/{key_id}", apiKeyHandler.Delete)

				// Webhook routes
				r.Post("/webhooks", webhookHandler.Create)
				r.Get("/webhooks", webhookHandler.List)
				r.Delete("/webhooks/{webhook_id}", webhookHandler.Delete)
				r.Get("/webhooks/{webhook_id}/deliveries", webhookHandler.ListDeliveries)
			})
		})

		// Channel routes
		r.Route("/channels", func(r chi.Router) {
			r.Post("/", channelHandler.Create)
			r.Get("/", channelHandler.List)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", channelHandler.Get)
				r.Patch("/", channelHandler.Update)
				r.Delete("/", channelHandler.Delete)
				r.Post("/read", channelHandler.MarkRead)

				// Channel member routes
				r.Get("/members", channelHandler.GetMembers)
				r.Post("/members", channelHandler.AddMember)
				r.Delete("/members/{user_id}", channelHandler.RemoveMember)

				// Message routes
				r.Post("/messages", messageHandler.Send)
				r.Get("/messages", messageHandler.GetMessages)
				r.Patch("/messages/{message_id}", messageHandler.EditMessage)
				r.Delete("/messages/{message_id}", messageHandler.DeleteMessage)
			})
		})

		// Message search route
		r.Get("/messages/search", messageHandler.SearchMessages)

		// DM routes
		r.Post("/dm", channelHandler.CreateDM)
		r.Get("/dm", channelHandler.ListDMs)
		r.Post("/dm/{id}/read", channelHandler.MarkDMRead)
	})

	// Backward compatibility: Mount protected routes on /v1 as well
	r.Route("/v1", func(r chi.Router) {
		r.Use(mw.Auth(cfg.JWT.Secret, apiKeyRepo, agentRepo, userRepo))

		// Current user profile routes
		r.Get("/me", meHandler.GetMe)
		r.Put("/me", meHandler.UpdateMe)
		r.Post("/me/password", meHandler.ChangePassword)

		// Users/invites routes
		r.Get("/users", usersHandler.List)
		r.Post("/invites", inviteHandler.CreateInvite)

		// Integrations routes
		r.Get("/integrations", integrationHandler.List)
		r.Post("/integrations", integrationHandler.Create)
		r.Delete("/integrations/{id}", integrationHandler.Delete)

		// Billing routes
		r.Post("/billing/checkout", billingHandler.Checkout)
		r.Get("/billing/checkout", billingHandler.Checkout)
		r.Post("/billing/portal", billingHandler.Portal)
		r.Get("/billing/portal", billingHandler.Portal)
		r.Get("/billing/status", billingHandler.Status)

		// Relay routes
		r.Get("/relay/status", relayHandler.Status)

		// Admin routes
		r.Route("/admin", func(r chi.Router) {
			r.Post("/pairing", pairingHandler.AdminPairing)
			r.Get("/tunnel/status", pairingHandler.TunnelStatus)
			r.Post("/tunnel/subdomain/check", pairingHandler.SubdomainCheck)
			r.Post("/tunnel/subdomain/claim", pairingHandler.SubdomainClaim)
			r.Get("/pairing/status", pairingHandler.PairingStatus)
		})

		// Centrifugo auth/ops routes
		r.Route("/realtime/centrifugo", func(r chi.Router) {
			r.Post("/subscribe-token", centrifugoHandler.SubscribeToken)
			r.Post("/refresh-token", centrifugoHandler.RefreshToken)
			r.Get("/presence", centrifugoHandler.Presence)
			r.Get("/history", centrifugoHandler.History)
		})

		// Agent routes
		r.Route("/agents", func(r chi.Router) {
			r.Post("/", agentHandler.Create)
			r.Get("/", agentHandler.List)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", agentHandler.Get)
				r.Patch("/", agentHandler.Update)
				r.Delete("/", agentHandler.Delete)

				// API key routes
				r.Post("/keys", apiKeyHandler.Create)
				r.Get("/keys", apiKeyHandler.List)
				r.Delete("/keys/{key_id}", apiKeyHandler.Delete)

				// Webhook routes
				r.Post("/webhooks", webhookHandler.Create)
				r.Get("/webhooks", webhookHandler.List)
				r.Delete("/webhooks/{webhook_id}", webhookHandler.Delete)
				r.Get("/webhooks/{webhook_id}/deliveries", webhookHandler.ListDeliveries)
			})
		})

		// Channel routes
		r.Route("/channels", func(r chi.Router) {
			r.Post("/", channelHandler.Create)
			r.Get("/", channelHandler.List)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", channelHandler.Get)
				r.Patch("/", channelHandler.Update)
				r.Delete("/", channelHandler.Delete)
				r.Post("/read", channelHandler.MarkRead)

				// Channel member routes
				r.Get("/members", channelHandler.GetMembers)
				r.Post("/members", channelHandler.AddMember)
				r.Delete("/members/{user_id}", channelHandler.RemoveMember)

				// Message routes
				r.Post("/messages", messageHandler.Send)
				r.Get("/messages", messageHandler.GetMessages)
				r.Patch("/messages/{message_id}", messageHandler.EditMessage)
				r.Delete("/messages/{message_id}", messageHandler.DeleteMessage)
			})
		})

		// Message search route
		r.Get("/messages/search", messageHandler.SearchMessages)

		// DM routes
		r.Post("/dm", channelHandler.CreateDM)
		r.Get("/dm", channelHandler.ListDMs)
		r.Post("/dm/{id}/read", channelHandler.MarkDMRead)
	})

	return r
}

// loggerMiddleware logs HTTP requests using zerolog
func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", ww.Status()).
			Int("bytes", ww.BytesWritten()).
			Dur("duration", time.Since(start)).
			Msg("request completed")
	})
}
