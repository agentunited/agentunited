package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/agentunited/backend/internal/api/middleware"
	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/repository"
	"github.com/agentunited/backend/internal/service"
	"github.com/agentunited/backend/internal/utils"
	"github.com/agentunited/backend/pkg/integrations"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// MessageHandler handles message HTTP requests
type RealtimePublisher interface {
	Enabled() bool
	Publish(ctx context.Context, channelID string, payload any) error
}

type Broadcaster interface {
	Broadcast(ctx context.Context, channelID string, message []byte)
}

type IntegrationEventRouter interface {
	RouteEvent(ctx context.Context, event integrations.Event) error
}

type MessageHandler struct {
	messageService    service.MessageService
	webhookService    service.WebhookService
	hub               Broadcaster
	realtime          RealtimePublisher
	integrationRouter IntegrationEventRouter
	userRepo          repository.UserRepository
	redisClient       *redis.Client
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messageService service.MessageService, webhookService service.WebhookService, hub Broadcaster, rt RealtimePublisher, redisClient *redis.Client, routers ...IntegrationEventRouter) *MessageHandler {
	var router IntegrationEventRouter
	if len(routers) > 0 {
		router = routers[0]
	}
	return &MessageHandler{
		messageService:    messageService,
		webhookService:    webhookService,
		hub:               hub,
		realtime:          rt,
		integrationRouter: router,
		redisClient:       redisClient,
	}
}

// SetUserRepo injects the user repository (needed for SSE workspace owner resolution).
func (h *MessageHandler) SetUserRepo(repo repository.UserRepository) { h.userRepo = repo }

// workspaceStreamKey returns the Redis Stream key for workspace-level SSE events.
// We key by the earliest created user (owner), so all workspace members share the same stream.
func (h *MessageHandler) workspaceStreamKey(ctx context.Context) string {
	if h.userRepo != nil {
		if owner, err := h.userRepo.GetEarliestUser(ctx); err == nil {
			return fmt.Sprintf("workspace:%s:events", owner.ID)
		}
	}
	return "workspace:unknown:events"
}

// SendMessageRequest represents the send message request body
type SendMessageRequest struct {
	Text string `json:"text"`
}

// Send handles POST /api/v1/channels/:channel_id/messages
// Supports both JSON (application/json) and file upload (multipart/form-data)
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

	// Parse request based on content type
	contentType := r.Header.Get("Content-Type")
	var text, attachmentURL, attachmentName string

	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Handle multipart form data (with optional file)
		if err := r.ParseMultipartForm(utils.MaxFileSize); err != nil {
			log.Error().Err(err).Msg("failed to parse multipart form")
			respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid multipart form data"})
			return
		}

		// Get text from form
		text = r.FormValue("text")

		// Handle file upload if present
		if file, fileHeader, err := r.FormFile("file"); err == nil {
			defer file.Close()

			// Save the file
			url, name, saveErr := utils.SaveFile(fileHeader)
			if saveErr != nil {
				if fileErr, ok := saveErr.(*utils.FileUploadError); ok {
					switch fileErr.Code {
					case "FILE_TOO_LARGE":
						respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "File size exceeds 10MB limit"})
					case "INVALID_FILE_TYPE":
						respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "File type not allowed"})
					default:
						respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "File upload failed"})
					}
				} else {
					log.Error().Err(saveErr).Msg("file save error")
					respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "File upload failed"})
				}
				return
			}

			attachmentURL = url
			attachmentName = name
		}

		// At least one of text or file must be provided
		if text == "" && attachmentURL == "" {
			respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Message text or file is required"})
			return
		}

	} else {
		// Handle JSON request (backward compatibility)
		var req SendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error().Err(err).Msg("failed to decode send message request")
			respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
			return
		}

		text = req.Text
		if text == "" {
			respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Message text is required"})
			return
		}
	}

	// Create message with attachment info
	message := &models.Message{
		ChannelID:      channelID,
		AuthorID:       userID,
		AuthorType:     "user",
		Text:           text,
		AttachmentURL:  attachmentURL,
		AttachmentName: attachmentName,
	}

	// Call service — use agent-aware method when authenticated as agent
	var finalMessage *models.Message
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
		finalMessage, err = h.messageService.SendMessageWithAttachment(ctx, message, &agentCtx)
	} else {
		finalMessage, err = h.messageService.SendMessageWithAttachment(ctx, message, nil)
	}
	if err != nil {
		h.handleMessageError(w, err, "send message")
		return
	}

	// Dispatch webhooks for message.created event (async), except for agent-authored
	// messages to prevent agent->webhook->agent feedback loops.
	webhookPayload := map[string]interface{}{
		"channel_id":      finalMessage.ChannelID,
		"message_id":      finalMessage.ID,
		"author_id":       finalMessage.AuthorID,
		"author_type":     finalMessage.AuthorType,
		"text":            finalMessage.Text,
		"attachment_url":  finalMessage.AttachmentURL,
		"attachment_name": finalMessage.AttachmentName,
		"created_at":      finalMessage.CreatedAt,
	}
	if finalMessage.AuthorType != "agent" {
		h.webhookService.DispatchEvent(ctx, channelID, "message.created", webhookPayload)
		if h.redisClient != nil {
			eventData := map[string]any{
				"event":       "message.created",
				"channel_id":  finalMessage.ChannelID,
				"message_id":  finalMessage.ID,
				"author_id":   finalMessage.AuthorID,
				"author_type": finalMessage.AuthorType,
				"text":        finalMessage.Text,
				"created_at":  finalMessage.CreatedAt.Format(time.RFC3339Nano),
			}
			if b, err := json.Marshal(eventData); err == nil {
				streamKey := h.workspaceStreamKey(ctx)
				_, _ = h.redisClient.XAdd(ctx, &redis.XAddArgs{Stream: streamKey, MaxLen: 2000, Approx: true, Values: map[string]any{"event": string(b), "message_id": finalMessage.ID}}).Result()
			}
		}
	}

	if h.integrationRouter != nil {
		_ = h.integrationRouter.RouteEvent(ctx, integrations.Event{
			Type:        "message.created",
			WorkspaceID: userID,
			ChannelID:   channelID,
			Payload:     webhookPayload,
			OccurredAt:  finalMessage.CreatedAt,
		})
	}

	// Populate author display info for broadcast and response.
	// The message service doesn't resolve display names on write — do it here.
	if finalMessage.AuthorEmail == "" {
		if agentName, ok := middleware.GetAgentName(ctx); ok && agentName != "" {
			// Agent sender: use agent display name
			finalMessage.AuthorEmail = agentName
		} else if userEmail, ok := middleware.GetUserEmail(ctx); ok && userEmail != "" {
			// Human sender: resolve display name from user repo; fall back to email prefix
			if h.userRepo != nil {
				if u, err := h.userRepo.GetByID(ctx, finalMessage.AuthorID); err == nil && u != nil {
					if u.DisplayName != "" {
						finalMessage.AuthorEmail = u.DisplayName
					} else {
						finalMessage.AuthorEmail = strings.Split(u.Email, "@")[0]
					}
				}
			}
			if finalMessage.AuthorEmail == "" {
				finalMessage.AuthorEmail = strings.Split(userEmail, "@")[0]
			}
		}
	}

	// Dual fan-out: always broadcast to legacy WS hub and (best-effort) Centrifugo.
	if h.hub != nil {
		wsMessage := map[string]interface{}{
			"type": "message.created",
			"data": finalMessage,
		}
		if msgBytes, err := json.Marshal(wsMessage); err == nil {
			h.hub.Broadcast(ctx, channelID, msgBytes)
		}
	}
	if h.realtime != nil && h.realtime.Enabled() {
		_ = h.realtime.Publish(ctx, channelID, map[string]interface{}{
			"type": "message.created",
			"data": finalMessage,
		})
	}

	// Return success response
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": finalMessage,
	})
}

// StreamEvents handles GET /api/v1/events/stream
func (h *MessageHandler) StreamEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, ok := middleware.GetUserID(ctx)
	if !ok {
		respondJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}
	if h.redisClient == nil {
		respondJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "stream unavailable"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher, ok := w.(http.Flusher)
	if !ok { return }
	_, _ = fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	streamKey := h.workspaceStreamKey(ctx)
	lastID := r.Header.Get("Last-Event-ID")
	if lastID == "" {
		// No resume cursor: start from new events only.
		lastID = "$"
	} else {
		// Resume strictly after the last seen event id.
		lastID = "(" + lastID
	}

	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			_, _ = fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		default:
			res, err := h.redisClient.XRead(ctx, &redis.XReadArgs{Streams: []string{streamKey, lastID}, Count: 20, Block: 5 * time.Second}).Result()
			if err != nil {
				if err == redis.Nil { continue }
				if ctx.Err() != nil { return }
				continue
			}
			for _, s := range res {
				for _, m := range s.Messages {
					eventJSON, _ := m.Values["event"].(string)
					if eventJSON == "" { continue }
					_, _ = fmt.Fprintf(w, "id: %s\n", m.ID)
					_, _ = fmt.Fprintf(w, "event: message.created\n")
					_, _ = fmt.Fprintf(w, "data: %s\n\n", eventJSON)
					flusher.Flush()
					lastID = m.ID
				}
			}
		}
	}
}

// PollMessages handles GET /api/v1/messages?channel_id=&since=
func (h *MessageHandler) PollMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		respondJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}
	channelID := r.URL.Query().Get("channel_id")
	if channelID == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "channel_id is required"})
		return
	}
	sinceRaw := r.URL.Query().Get("since")
	var since time.Time
	var err error
	if sinceRaw != "" {
		since, err = time.Parse(time.RFC3339, sinceRaw)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "since must be ISO8601"})
			return
		}
	}
	list, err := h.messageService.GetMessages(ctx, channelID, userID, 100, "")
	if err != nil {
		h.handleMessageError(w, err, "poll messages")
		return
	}
	out := make([]*models.Message, 0)
	for _, m := range list.Messages {
		if sinceRaw == "" || m.CreatedAt.After(since) {
			out = append(out, m)
		}
	}
	respondJSON(w, http.StatusOK, map[string]any{"messages": out})
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

	// Dual fan-out: always broadcast to legacy WS hub and (best-effort) Centrifugo.
	if h.hub != nil {
		wsMessage := map[string]interface{}{
			"type": "message.updated",
			"data": message,
		}
		if msgBytes, err := json.Marshal(wsMessage); err == nil {
			h.hub.Broadcast(ctx, message.ChannelID, msgBytes)
		}
	}
	if h.realtime != nil && h.realtime.Enabled() {
		_ = h.realtime.Publish(ctx, message.ChannelID, map[string]interface{}{
			"type": "message.updated",
			"data": message,
		})
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
