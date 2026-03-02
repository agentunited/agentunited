package api

import (
	"net/http"
	"time"

	"github.com/agentunited/backend/internal/api/handlers"
	mw "github.com/agentunited/backend/internal/api/middleware"
	"github.com/agentunited/backend/internal/repository"
	"github.com/agentunited/backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog/log"
)

// NewRouter creates and configures the Chi router
func NewRouter(db *repository.DB, cache *repository.Cache, jwtSecret string) *chi.Mux {
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
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check handler
	healthHandler := NewHealthHandler(db, cache)
	r.Get("/health", healthHandler.Check)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	channelRepo := repository.NewChannelRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	agentRepo := repository.NewAgentRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	webhookRepo := repository.NewWebhookRepository(db)
	inviteRepo := repository.NewInviteRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, jwtSecret)
	channelService := service.NewChannelService(channelRepo)
	messageService := service.NewMessageService(messageRepo, channelRepo)
	agentService := service.NewAgentService(agentRepo)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, agentRepo)
	webhookService := service.NewWebhookService(webhookRepo, agentRepo)
	bootstrapService := service.NewBootstrapService(userRepo, agentRepo, apiKeyRepo, inviteRepo, channelRepo, jwtSecret, "http://localhost:8080")
	inviteService := service.NewInviteService(userRepo, inviteRepo, jwtSecret)

	// Initialize WebSocket hub first (needed by message handler)
	hub := handlers.NewHub()
	
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	channelHandler := handlers.NewChannelHandler(channelService)
	messageHandler := handlers.NewMessageHandler(messageService, webhookService, hub)
	agentHandler := handlers.NewAgentHandler(agentService)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyService)
	webhookHandler := handlers.NewWebhookHandler(webhookService)
	bootstrapHandler := handlers.NewBootstrapHandler(bootstrapService)
	inviteHandler := handlers.NewInviteHandler(inviteService)
	wsHandler := handlers.NewWebSocketHandlerV2(messageService, channelService, jwtSecret, hub)

	// WebSocket endpoint — no timeout middleware (long-lived connection)
	r.Handle("/ws", wsHandler)

	// Public routes (no authentication required)
	r.Post("/api/v1/bootstrap", bootstrapHandler.ServeHTTP)
	r.Get("/api/v1/invite", inviteHandler.ValidateInvite)
	r.Post("/api/v1/invite/accept", inviteHandler.AcceptInvite)
	
	// Public authentication routes
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})
	
	// Protected API v1 routes (require JWT or API key authentication)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(mw.Auth(jwtSecret, apiKeyRepo, agentRepo))
		
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
	})

	// Backward compatibility: Mount protected routes on /v1 as well
	r.Route("/v1", func(r chi.Router) {
		r.Use(mw.Auth(jwtSecret, apiKeyRepo, agentRepo))
		
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
