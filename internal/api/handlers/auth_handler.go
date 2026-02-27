package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/service"
	"github.com/rs/zerolog/log"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User  *models.User `json:"user"`
	Token string       `json:"token"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("failed to decode register request")
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Email and password are required"})
		return
	}

	// Call service
	user, err := h.authService.Register(ctx, req.Email, req.Password)
	if err != nil {
		h.handleAuthError(w, err, "register")
		return
	}

	// Generate JWT token (call Login to get token)
	token, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate token after registration")
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
		return
	}

	// Return success response
	respondJSON(w, http.StatusCreated, AuthResponse{
		User:  user,
		Token: token,
	})
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("failed to decode login request")
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Email and password are required"})
		return
	}

	// Call service
	token, err := h.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		h.handleAuthError(w, err, "login")
		return
	}

	// Return success response
	respondJSON(w, http.StatusOK, map[string]string{
		"token": token,
	})
}

// handleAuthError maps service errors to HTTP status codes
func (h *AuthHandler) handleAuthError(w http.ResponseWriter, err error, operation string) {
	log.Error().Err(err).Str("operation", operation).Msg("authentication error")

	switch {
	case errors.Is(err, models.ErrInvalidEmail):
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid email format"})
	case errors.Is(err, models.ErrWeakPassword):
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Password must be at least 8 characters with 1 letter and 1 number"})
	case errors.Is(err, models.ErrEmailTaken):
		respondJSON(w, http.StatusConflict, ErrorResponse{Error: "Email already taken"})
	case errors.Is(err, models.ErrInvalidCredentials):
		respondJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
	default:
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("failed to encode JSON response")
	}
}
