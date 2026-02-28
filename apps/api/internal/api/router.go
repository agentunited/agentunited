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
	r.Use(middleware.Timeout(60 * time.Second))

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

	// Initialize services
	authService := service.NewAuthService(userRepo, jwtSecret)
	channelService := service.NewChannelService(channelRepo)
	messageService := service.NewMessageService(messageRepo, channelRepo)
	agentService := service.NewAgentService(agentRepo)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, agentRepo)
	webhookService := service.NewWebhookService(webhookRepo, agentRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	channelHandler := handlers.NewChannelHandler(channelService)
	messageHandler := handlers.NewMessageHandler(messageService)
	agentHandler := handlers.NewAgentHandler(agentService)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyService)
	webhookHandler := handlers.NewWebhookHandler(webhookService)
	
	// Initialize WebSocket hub and handler
	hub := handlers.NewHub()
	wsHandler := handlers.NewWebSocketHandlerV2(messageService, channelService, jwtSecret, hub)

	// WebSocket endpoint (query param auth, not middleware)
	r.Handle("/ws", wsHandler)

	// Authentication routes (public)
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	// Protected routes (require JWT)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(mw.JWTAuth(jwtSecret))
		
		// Agent routes
		r.Route("/agents", func(r chi.Router) {
			r.Post("/", agentHandler.Create)
			r.Get("/", agentHandler.List)
			r.Get("/{id}", agentHandler.Get)
			r.Patch("/{id}", agentHandler.Update)
			r.Delete("/{id}", agentHandler.Delete)
			
			// API key routes (nested under agents)
			r.Post("/{agent_id}/keys", apiKeyHandler.Create)
			r.Get("/{agent_id}/keys", apiKeyHandler.List)
			r.Delete("/{agent_id}/keys/{key_id}", apiKeyHandler.Delete)
			
			// Webhook routes (nested under agents)
			r.Post("/{agent_id}/webhooks", webhookHandler.Create)
			r.Get("/{agent_id}/webhooks", webhookHandler.List)
			r.Delete("/{agent_id}/webhooks/{webhook_id}", webhookHandler.Delete)
			r.Get("/{agent_id}/webhooks/{webhook_id}/deliveries", webhookHandler.ListDeliveries)
		})
		
		// Channel routes
		r.Route("/channels", func(r chi.Router) {
			r.Post("/", channelHandler.Create)
			r.Get("/", channelHandler.List)
			r.Get("/{id}", channelHandler.Get)
			
			// Message routes (nested under channels)
			r.Post("/{channel_id}/messages", messageHandler.Send)
			r.Get("/{channel_id}/messages", messageHandler.GetMessages)
		})
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
