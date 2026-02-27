package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/agentunited/backend/internal/api/middleware"
	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// ChannelHandler handles channel HTTP requests
type ChannelHandler struct {
	channelService service.ChannelService
}

// NewChannelHandler creates a new channel handler
func NewChannelHandler(channelService service.ChannelService) *ChannelHandler {
	return &ChannelHandler{
		channelService: channelService,
	}
}

// CreateChannelRequest represents the create channel request body
type CreateChannelRequest struct {
	Name  string `json:"name"`
	Topic string `json:"topic"`
}

// Create handles POST /api/v1/channels
func (h *ChannelHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		respondJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	// Parse request body
	var req CreateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("failed to decode create channel request")
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate required fields
	if req.Name == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Channel name is required"})
		return
	}

	// Call service
	channel, err := h.channelService.Create(ctx, userID, req.Name, req.Topic)
	if err != nil {
		h.handleChannelError(w, err, "create channel")
		return
	}

	// Return success response
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"channel": channel,
	})
}

// List handles GET /api/v1/channels
func (h *ChannelHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		respondJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	// Call service
	channels, err := h.channelService.List(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("failed to list channels")
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
		return
	}

	// Return success response
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"channels": channels,
	})
}

// Get handles GET /api/v1/channels/:id
func (h *ChannelHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		respondJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	// Get channel ID from URL param
	channelID := chi.URLParam(r, "id")
	if channelID == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
		return
	}

	// Call service
	channel, err := h.channelService.Get(ctx, channelID, userID)
	if err != nil {
		h.handleChannelError(w, err, "get channel")
		return
	}

	// Return success response
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"channel": channel,
	})
}

// handleChannelError maps service errors to HTTP status codes
func (h *ChannelHandler) handleChannelError(w http.ResponseWriter, err error, operation string) {
	log.Error().Err(err).Str("operation", operation).Msg("channel error")

	switch {
	case errors.Is(err, models.ErrInvalidChannelName):
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Channel name must be lowercase alphanumeric with hyphens only"})
	case errors.Is(err, models.ErrChannelNameTaken):
		respondJSON(w, http.StatusConflict, ErrorResponse{Error: "Channel name already taken"})
	case errors.Is(err, models.ErrChannelNotFound):
		respondJSON(w, http.StatusNotFound, ErrorResponse{Error: "Channel not found"})
	case errors.Is(err, models.ErrNotChannelMember):
		respondJSON(w, http.StatusForbidden, ErrorResponse{Error: "Not a member of this channel"})
	default:
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}
}
