package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

// contextKey is a type for context keys
type contextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"
)

// JWTAuth creates a JWT authentication middleware
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

			// Parse and validate token
			token, err := jwt.ParseWithClaims(tokenString, &service.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
				// Verify signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, models.ErrInvalidToken
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				log.Error().Err(err).Msg("invalid JWT token")
				respondUnauthorized(w, "Invalid or expired token")
				return
			}

			// Extract claims
			claims, ok := token.Claims.(*service.JWTClaims)
			if !ok {
				respondUnauthorized(w, "Invalid token claims")
				return
			}

			// Add user ID to context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)

			// Call next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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
