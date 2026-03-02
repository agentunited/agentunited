package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/agentunited/backend/internal/api/middleware"
	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// MessageHandler handles message HTTP requests
type MessageHandler struct {
	messageService service.MessageService
	webhookService service.WebhookService
	hub            *Hub
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messageService service.MessageService, webhookService service.WebhookService, hub *Hub) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
		webhookService: webhookService,
		hub:            hub,
	}
}

// SendMessageRequest represents the send message request body
type SendMessageRequest struct {
	Text string `json:"text"`
}

// Send handles POST /api/v1/channels/:channel_id/messages
func (h *MessageHandler) Send(w http.ResponseWriter, r *http.Request) {
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
	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("failed to decode send message request")
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate text
	if req.Text == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Message text is required"})
		return
	}

	// Call service — use agent-aware method when authenticated as agent
	var message *models.Message
	var err error
	if agentID, ok := middleware.GetAgentID(ctx); ok {
		agentName := ""
		if name, ok := middleware.GetAgentName(ctx); ok {
			agentName = name
		}
		agentCtx := service.AgentContext{
			AgentID:     agentID,
			DisplayName: agentName,
		}
		message, err = h.messageService.SendAsAgent(ctx, channelID, userID, agentCtx, req.Text)
	} else {
		message, err = h.messageService.Send(ctx, channelID, userID, req.Text)
	}
	if err != nil {
		h.handleMessageError(w, err, "send message")
		return
	}

	// Dispatch webhooks for message.created event (async)
	webhookPayload := map[string]interface{}{
		"channel_id":  message.ChannelID,
		"message_id":  message.ID,
		"author_id":   message.AuthorID,
		"author_type": message.AuthorType,
		"text":        message.Text,
		"created_at":  message.CreatedAt,
	}
	h.webhookService.DispatchEvent(ctx, channelID, "message.created", webhookPayload)

	// Populate author display info for broadcast
	if message.AuthorEmail == "" {
		if agentName, ok := middleware.GetAgentName(ctx); ok && agentName != "" {
			message.AuthorEmail = agentName
		}
	}

	// Broadcast to WebSocket clients
	if h.hub != nil {
		wsMessage := map[string]interface{}{
			"type": "message.created",
			"data": message,
		}
		if msgBytes, err := json.Marshal(wsMessage); err == nil {
			h.hub.Broadcast(ctx, channelID, msgBytes)
		}
	}

	// Return success response
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": message,
	})
}

// GetMessages handles GET /api/v1/channels/:channel_id/messages
func (h *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
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

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // Default
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	before := r.URL.Query().Get("before")

	// Call service
	messageList, err := h.messageService.GetMessages(ctx, channelID, userID, limit, before)
	if err != nil {
		h.handleMessageError(w, err, "get messages")
		return
	}

	// Return success response
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"messages": messageList.Messages,
		"has_more": messageList.HasMore,
	})
}

// EditMessageRequest represents the edit message request body
type EditMessageRequest struct {
	Text string `json:"text"`
}

// EditMessage handles PATCH /api/v1/channels/{channel_id}/messages/{id}
func (h *MessageHandler) EditMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		respondJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	// Get message ID from URL param
	messageID := chi.URLParam(r, "message_id")
	if messageID == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Message ID is required"})
		return
	}

	// Parse request body
	var req EditMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("failed to decode edit message request")
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate text
	if req.Text == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Message text is required"})
		return
	}

	// Call service
	message, err := h.messageService.EditMessage(ctx, messageID, userID, req.Text)
	if err != nil {
		h.handleMessageError(w, err, "edit message")
		return
	}

	// Broadcast to WebSocket clients
	if h.hub != nil {
		wsMessage := map[string]interface{}{
			"type": "message.updated",
			"data": message,
		}
		if msgBytes, err := json.Marshal(wsMessage); err == nil {
			h.hub.Broadcast(ctx, message.ChannelID, msgBytes)
		}
	}

	// Return success response
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": message,
	})
}

// DeleteMessage handles DELETE /api/v1/channels/{channel_id}/messages/{id}
func (h *MessageHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		respondJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	// Get message ID from URL param
	messageID := chi.URLParam(r, "message_id")
	if messageID == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Message ID is required"})
		return
	}

	// Call service
	err := h.messageService.DeleteMessage(ctx, messageID, userID)
	if err != nil {
		h.handleMessageError(w, err, "delete message")
		return
	}

	// Note: WebSocket broadcast for message deletion is handled in the service layer
	// since we need the channel ID before deletion

	// Return success response (204 No Content)
	w.WriteHeader(http.StatusNoContent)
}

// SearchMessages handles GET /api/v1/messages/search
func (h *MessageHandler) SearchMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		respondJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	// Parse query parameters
	query := r.URL.Query().Get("q")
	if query == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Search query (q) is required"})
		return
	}

	channelID := r.URL.Query().Get("channel_id")

	limitStr := r.URL.Query().Get("limit")
	limit := 50 // Default
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Call service
	messages, err := h.messageService.SearchMessages(ctx, query, channelID, userID, limit)
	if err != nil {
		h.handleMessageError(w, err, "search messages")
		return
	}

	// Return success response
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"messages": messages,
		"query":    query,
		"count":    len(messages),
	})
}

// handleMessageError maps service errors to HTTP status codes
func (h *MessageHandler) handleMessageError(w http.ResponseWriter, err error, operation string) {
	log.Error().Err(err).Str("operation", operation).Msg("message error")

	switch {
	case errors.Is(err, models.ErrInvalidMessageText):
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Message text must be between 1 and 10,000 characters"})
	case errors.Is(err, models.ErrNotChannelMember):
		respondJSON(w, http.StatusForbidden, ErrorResponse{Error: "Not a member of this channel"})
	case errors.Is(err, models.ErrChannelNotFound):
		respondJSON(w, http.StatusNotFound, ErrorResponse{Error: "Channel not found"})
	case errors.Is(err, models.ErrMessageNotFound):
		respondJSON(w, http.StatusNotFound, ErrorResponse{Error: "Message not found"})
	case errors.Is(err, models.ErrUnauthorizedMessageEdit):
		respondJSON(w, http.StatusForbidden, ErrorResponse{Error: "You can only edit your own messages"})
	case errors.Is(err, models.ErrUnauthorizedMessageDelete):
		respondJSON(w, http.StatusForbidden, ErrorResponse{Error: "You can only delete your own messages"})
	default:
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}
}
