package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	mw "github.com/agentunited/backend/internal/api/middleware"
	"github.com/agentunited/backend/internal/repository"
)

// EventsHandler serves SSE stream of message events.
type EventsHandler struct {
	channelRepo repository.ChannelRepository
	hub         *Hub
}

func NewEventsHandler(channelRepo repository.ChannelRepository, hub *Hub) *EventsHandler {
	return &EventsHandler{channelRepo: channelRepo, hub: hub}
}

// Stream handles GET /api/v1/events/stream
// Streams message.created events for channels the authenticated user belongs to.
func (h *EventsHandler) Stream(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.GetUserID(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "streaming unsupported"})
		return
	}

	channels, err := h.channelRepo.ListByUser(r.Context(), userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	eventCh := make(chan []byte, 128)
	for _, c := range channels {
		h.hub.SubscribeSSE(c.ID, userID, eventCh)
	}
	defer func() {
		for _, c := range channels {
			h.hub.UnsubscribeSSE(c.ID, eventCh)
		}
		close(eventCh)
	}()

	// Initial connect frame
	_, _ = fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"ok\"}\n\n")
	flusher.Flush()

	ping := time.NewTicker(25 * time.Second)
	defer ping.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ping.C:
			_, _ = fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		case msg := <-eventCh:
			// msg payload is JSON from hub broadcast: {"type":"message.created","data":...}
			var payload map[string]interface{}
			eventName := "message"
			if err := json.Unmarshal(msg, &payload); err == nil {
				if t, ok := payload["type"].(string); ok && t != "" {
					eventName = t
				}
			}
			_, _ = fmt.Fprintf(w, "event: %s\n", eventName)
			_, _ = fmt.Fprintf(w, "data: %s\n\n", string(msg))
			flusher.Flush()
		}
	}
}
