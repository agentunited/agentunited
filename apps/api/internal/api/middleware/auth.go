package middleware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"

	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/repository"
	"github.com/agentunited/backend/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

// contextKey is a type for context keys
type contextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey     contextKey = "user_id"
	AgentIDKey    contextKey = "agent_id"
	AgentNameKey  contextKey = "agent_name"
)

// Auth creates an authentication middleware that supports both JWT tokens and API keys
func Auth(jwtSecret string, apiKeyRepo repository.APIKeyRepository, agentRepo repository.AgentRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondUnauthorized(w, "Missing authorization header")
				return
			}

			// Check for Bearer token format
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondUnauthorized(w, "Invalid authorization header format")
				return
			}

			tokenString := parts[1]
			var userID string
			var err error

			// Check if this is an API key (starts with au_) or JWT token
			var agentID, agentName string
			if strings.HasPrefix(tokenString, "au_") {
				// API Key authentication
				result, authErr := authenticateAPIKey(r.Context(), tokenString, apiKeyRepo, agentRepo)
				if authErr != nil {
					log.Error().Err(authErr).Str("key_prefix", tokenString[:10]+"...").Msg("invalid API key")
					respondUnauthorized(w, "Invalid or expired API key")
					return
				}
				userID = result.UserID
				agentID = result.AgentID
				agentName = result.AgentName
			} else {
				// JWT token authentication
				userID, err = authenticateJWT(tokenString, jwtSecret)
				if err != nil {
					log.Error().Err(err).Msg("invalid JWT token")
					respondUnauthorized(w, "Invalid or expired token")
					return
				}
			}

			// Add user ID to context
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			if agentID != "" {
				ctx = context.WithValue(ctx, AgentIDKey, agentID)
				ctx = context.WithValue(ctx, AgentNameKey, agentName)
			}

			// Call next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// JWTAuth creates a JWT-only authentication middleware (backward compatibility)
func JWTAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondUnauthorized(w, "Missing authorization header")
				return
			}

			// Check for Bearer token format
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondUnauthorized(w, "Invalid authorization header format")
				return
			}

			tokenString := parts[1]

			// JWT token authentication only
			userID, err := authenticateJWT(tokenString, jwtSecret)
			if err != nil {
				log.Error().Err(err).Msg("invalid JWT token")
				respondUnauthorized(w, "Invalid or expired token")
				return
			}

			// Add user ID to context
			ctx := context.WithValue(r.Context(), UserIDKey, userID)

			// Call next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// authenticateJWT validates a JWT token and returns the user ID
func authenticateJWT(tokenString, jwtSecret string) (string, error) {
	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &service.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, models.ErrInvalidToken
		}
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return "", err
	}

	// Extract claims
	claims, ok := token.Claims.(*service.JWTClaims)
	if !ok {
		return "", models.ErrInvalidToken
	}

	return claims.UserID, nil
}

// authenticateAPIKey validates an API key and returns the user ID
type apiKeyResult struct {
	UserID      string
	AgentID     string
	AgentName   string
}

func authenticateAPIKey(ctx context.Context, apiKey string, apiKeyRepo repository.APIKeyRepository, agentRepo repository.AgentRepository) (*apiKeyResult, error) {
	// Hash the provided API key
	keyHash := hashAPIKey(apiKey)
	
	// Look up the API key in the database
	key, err := apiKeyRepo.GetByHash(ctx, keyHash)
	if err != nil {
		return nil, err
	}

	// Get the agent to find the owner (user)
	agent, err := agentRepo.Get(ctx, key.AgentID)
	if err != nil {
		return nil, err
	}

	return &apiKeyResult{
		UserID:    agent.OwnerID,
		AgentID:   agent.ID,
		AgentName: agent.DisplayName,
	}, nil
}

// hashAPIKey hashes an API key using SHA-256
func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", hash)
}

// respondUnauthorized sends a 401 response
func respondUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

// GetUserID extracts user ID from request context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// GetAgentID extracts agent ID from request context (set when API key auth is used)
func GetAgentID(ctx context.Context) (string, bool) {
	agentID, ok := ctx.Value(AgentIDKey).(string)
	return agentID, ok
}

// GetAgentName extracts agent display name from request context
func GetAgentName(ctx context.Context) (string, bool) {
	agentName, ok := ctx.Value(AgentNameKey).(string)
	return agentName, ok
}
