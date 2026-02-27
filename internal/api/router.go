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

	// Initialize services
	authService := service.NewAuthService(userRepo, jwtSecret)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Authentication routes (public)
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	// Protected routes (require JWT)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(mw.JWTAuth(jwtSecret))
		// Add protected endpoints here (e.g., /me, /users, etc.)
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
