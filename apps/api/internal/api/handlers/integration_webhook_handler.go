package handlers

import (
	"io"
	"net/http"

	"github.com/agentunited/backend/pkg/integrations"
)

// IntegrationInboundHandler handles inbound webhooks from external integrations
func IntegrationInboundHandler(w http.ResponseWriter, r *http.Request) {
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

	// For now, just acknowledge receipt
	// Full implementation would call message service to create the message
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "received",
		"platform":   platform,
		"channel_id": inbound.ChannelID,
		"agent_id":   inbound.AgentID,
		"text":       inbound.Text,
	})
}
