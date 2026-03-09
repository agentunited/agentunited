package handlers

import (
	"io"
	"net/http"

	"github.com/agentunited/backend/internal/repository"
	"github.com/agentunited/backend/internal/service"
	"github.com/agentunited/backend/pkg/integrations"
)

// IntegrationInboundHandler handles inbound webhooks from external integrations
type IntegrationInboundHandler struct {
	messageService service.MessageService
	agentRepo      repository.AgentRepository
}

// NewIntegrationInboundHandler creates a new integration inbound handler
func NewIntegrationInboundHandler(messageService service.MessageService, agentRepo repository.AgentRepository) *IntegrationInboundHandler {
	return &IntegrationInboundHandler{
		messageService: messageService,
		agentRepo:      agentRepo,
	}
}

// ServeHTTP handles the inbound webhook request
func (h *IntegrationInboundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract platform from URL path
	// Expected path: /api/v1/webhooks/integration/{platform}
	path := r.URL.Path
	platformStart := len("/api/v1/webhooks/integration/")
	if len(path) <= platformStart {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "platform not specified"})
		return
	}
	platform := path[platformStart:]

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid body"})
		return
	}
	defer r.Body.Close()

	// Get adapter for platform
	adapter, ok := integrations.GetAdapter(integrations.Platform(platform))
	if !ok {
		respondJSON(w, http.StatusNotFound, ErrorResponse{Error: "unsupported platform: " + platform})
		return
	}

	// Parse inbound message
	ctx := r.Context()
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	inbound, err := adapter.HandleInbound(ctx, body, headers)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Look up agent to get owner ID for channel membership check
	agent, err := h.agentRepo.Get(ctx, inbound.AgentID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "agent not found: " + err.Error()})
		return
	}

	// Create the message via the message service
	// For agent-originated messages from integrations, use SendAsAgent
	agentCtx := service.AgentContext{
		AgentID:     inbound.AgentID,
		DisplayName: agent.DisplayName,
	}

	createdMessage, err := h.messageService.SendAsAgent(ctx, inbound.ChannelID, agent.OwnerID, agentCtx, inbound.Text)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to create message: " + err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "received",
		"platform":   platform,
		"channel_id": inbound.ChannelID,
		"agent_id":   inbound.AgentID,
		"message_id": createdMessage.ID,
		"text":       inbound.Text,
	})
}
