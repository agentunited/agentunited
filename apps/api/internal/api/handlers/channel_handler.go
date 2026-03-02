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

// UpdateChannelRequest represents the update channel request body
type UpdateChannelRequest struct {
	Name        *string `json:"name,omitempty"`
	Topic       *string `json:"topic,omitempty"`
	Description *string `json:"description,omitempty"`
}

// Update handles PATCH /api/v1/channels/{id}
func (h *ChannelHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	// Parse request body
	var req UpdateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("failed to decode update channel request")
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Get current channel to merge fields
	current, err := h.channelService.Get(ctx, channelID, userID)
	if err != nil {
		h.handleChannelError(w, err, "get channel for update")
		return
	}
	
	name := current.Name
	topic := current.Topic
	if req.Name != nil {
		name = *req.Name
	}
	if req.Topic != nil {
		topic = *req.Topic
	}
	if req.Description != nil {
		topic = *req.Description
	}

	// Call service
	channel, err := h.channelService.Update(ctx, channelID, userID, name, topic)
	if err != nil {
		h.handleChannelError(w, err, "update channel")
		return
	}

	// Return success response
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"channel": channel,
	})
}

// Delete handles DELETE /api/v1/channels/{id}
func (h *ChannelHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
	err := h.channelService.Delete(ctx, channelID, userID)
	if err != nil {
		h.handleChannelError(w, err, "delete channel")
		return
	}

	// Return success response (204 No Content)
	w.WriteHeader(http.StatusNoContent)
}

// GetMembers handles GET /api/v1/channels/{id}/members
func (h *ChannelHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
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
	members, err := h.channelService.GetMembers(ctx, channelID, userID)
	if err != nil {
		h.handleChannelError(w, err, "get members")
		return
	}

	// Return success response
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"members": members,
	})
}

// AddMemberRequest represents the add member request body
type AddMemberRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role,omitempty"`
}

// AddMember handles POST /api/v1/channels/{id}/members
func (h *ChannelHandler) AddMember(w http.ResponseWriter, r *http.Request) {
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

	// Parse request body
	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("failed to decode add member request")
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate required fields
	if req.UserID == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "User ID is required"})
		return
	}

	// Default role to member if not specified
	role := req.Role
	if role == "" {
		role = "member"
	}

	// Call service
	err := h.channelService.AddMember(ctx, channelID, req.UserID, userID, role)
	if err != nil {
		h.handleChannelError(w, err, "add member")
		return
	}

	// Return success response (201 Created)
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Member added successfully",
	})
}

// RemoveMember handles DELETE /api/v1/channels/{id}/members/{user_id}
func (h *ChannelHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		respondJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	// Get channel ID and target user ID from URL params
	channelID := chi.URLParam(r, "id")
	targetUserID := chi.URLParam(r, "user_id")
	if channelID == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
		return
	}
	if targetUserID == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "User ID is required"})
		return
	}

	// Call service
	err := h.channelService.RemoveMember(ctx, channelID, targetUserID, userID)
	if err != nil {
		h.handleChannelError(w, err, "remove member")
		return
	}

	// Return success response (204 No Content)
	w.WriteHeader(http.StatusNoContent)
}

// CreateDMRequest represents the create DM request body
type CreateDMRequest struct {
	UserID string `json:"user_id"`
}

// CreateDM handles POST /api/v1/dm
func (h *ChannelHandler) CreateDM(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		respondJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	// Parse request body
	var req CreateDMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("failed to decode create dm request")
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate required fields
	if req.UserID == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "User ID is required"})
		return
	}

	// Call service
	channel, err := h.channelService.CreateOrGetDMChannel(ctx, userID, req.UserID)
	if err != nil {
		h.handleChannelError(w, err, "create dm")
		return
	}

	// Return success response
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"channel": channel,
	})
}

// ListDMs handles GET /api/v1/dm
func (h *ChannelHandler) ListDMs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		respondJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	// Call service
	channels, err := h.channelService.ListDMChannels(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("failed to list dm channels")
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
		return
	}

	// Return success response
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"channels": channels,
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
	case errors.Is(err, models.ErrInsufficientPermissions):
		respondJSON(w, http.StatusForbidden, ErrorResponse{Error: "Insufficient permissions"})
	case errors.Is(err, models.ErrAlreadyChannelMember):
		respondJSON(w, http.StatusConflict, ErrorResponse{Error: "User is already a member of this channel"})
	case errors.Is(err, models.ErrInvalidDMTarget):
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Cannot create DM with yourself"})
	default:
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}
}
