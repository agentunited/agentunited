package handlers

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// Hub manages WebSocket connections and broadcasting.
type Hub struct {
	// Registered clients by channel ID
	channels map[string]map[*websocket.Conn]string // channel_id -> conn -> user_id
	mu       sync.RWMutex
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		channels: make(map[string]map[*websocket.Conn]string),
	}
}

// Subscribe adds a connection to a channel.
func (h *Hub) Subscribe(ctx context.Context, channelID string, conn *websocket.Conn, userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if h.channels[channelID] == nil {
		h.channels[channelID] = make(map[*websocket.Conn]string)
	}
	h.channels[channelID][conn] = userID
	log.Info().Str("channel_id", channelID).Str("user_id", userID).Int("total", len(h.channels[channelID])).Msg("hub subscribe")
}

// Unsubscribe removes a connection from all channels.
func (h *Hub) Unsubscribe(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	for channelID := range h.channels {
		delete(h.channels[channelID], conn)
		if len(h.channels[channelID]) == 0 {
			delete(h.channels, channelID)
		}
	}
}

// Broadcast sends a message to all connections in a channel.
func (h *Hub) Broadcast(ctx context.Context, channelID string, message []byte) {
	h.mu.RLock()
	// Copy connections to avoid holding lock during writes
	conns := make(map[*websocket.Conn]string, len(h.channels[channelID]))
	for c, uid := range h.channels[channelID] {
		conns[c] = uid
	}
	h.mu.RUnlock()

	log.Info().Str("channel_id", channelID).Int("subscribers", len(conns)).Msg("hub broadcast")
	
	for conn := range conns {
		go func(c *websocket.Conn) {
			c.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err := c.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Warn().Err(err).Msg("hub broadcast write failed, removing dead connection")
				h.Unsubscribe(c)
				c.Close()
			}
		}(conn)
	}
}
