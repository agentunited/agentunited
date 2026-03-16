package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/agentunited/backend/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// BootstrapService defines the bootstrap operations
type BootstrapService interface {
	Bootstrap(ctx context.Context, req *models.BootstrapRequest) (*models.BootstrapResponse, error)
}

// BootstrapHandler handles bootstrap requests
type bootstrapRateLimiter interface {
	Allow(ctx context.Context, ip string) (blocked bool, err error)
}

type redisBootstrapRateLimiter struct {
	redisClient *redis.Client
}

func (l *redisBootstrapRateLimiter) Allow(ctx context.Context, ip string) (bool, error) {
	if l == nil || l.redisClient == nil || ip == "" {
		return false, nil
	}
	now := time.Now().UTC()
	key := fmt.Sprintf("rate:bootstrap:%s:%s", ip, now.Format("2006-01-02"))
	count, err := l.redisClient.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		expires := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
		_ = l.redisClient.ExpireAt(ctx, key, expires).Err()
	}
	return count > 3, nil
}

// BootstrapHandler handles bootstrap requests
type BootstrapHandler struct {
	service   BootstrapService
	validator *validator.Validate
	limiter   bootstrapRateLimiter
}

// NewBootstrapHandler creates a new bootstrap handler
func NewBootstrapHandler(service BootstrapService, redisClient *redis.Client) *BootstrapHandler {
	return NewBootstrapHandlerWithLimiter(service, &redisBootstrapRateLimiter{redisClient: redisClient})
}

func NewBootstrapHandlerWithLimiter(service BootstrapService, limiter bootstrapRateLimiter) *BootstrapHandler {
	return &BootstrapHandler{
		service:   service,
		validator: validator.New(),
		limiter:   limiter,
	}
}

func (h *BootstrapHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if blocked, err := h.rateLimitBootstrap(r.Context(), clientIP(r)); err != nil {
		log.Error().Err(err).Msg("bootstrap rate limit check failed")
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	} else if blocked {
		respondError(w, http.StatusTooManyRequests, "bootstrap rate limit exceeded: max 3 requests per IP per day")
		return
	}

	// Parse request
	var req models.BootstrapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Debug().Err(err).Msg("failed to decode bootstrap request")
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		log.Debug().Err(err).Msg("bootstrap request validation failed")
		respondError(w, http.StatusBadRequest, "validation failed: "+err.Error())
		return
	}

	// Call service
	resp, err := h.service.Bootstrap(r.Context(), &req)
	if err != nil {
		if errors.Is(err, models.ErrInstanceAlreadyBootstrapped) {
			respondError(w, http.StatusConflict, "instance has already been bootstrapped")
			return
		}
		if errors.Is(err, models.ErrEntityLimitReached) {
			respondJSON(w, http.StatusPaymentRequired, map[string]any{
				"error":       "entity_limit_reached",
				"message":     "Your workspace has reached the plan entity limit.",
				"upgrade_url": "/pricing",
			})
			return
		}

		log.Error().Err(err).Msg("bootstrap failed")
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Return response
	respondJSON(w, http.StatusCreated, resp)
}

func (h *BootstrapHandler) rateLimitBootstrap(ctx context.Context, ip string) (bool, error) {
	if strings.EqualFold(os.Getenv("BOOTSTRAP_RATE_LIMIT_DISABLED"), "true") {
		return false, nil
	}
	if h.limiter == nil {
		return false, nil
	}
	return h.limiter.Allow(ctx, ip)
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, map[string]string{"error": message})
}
