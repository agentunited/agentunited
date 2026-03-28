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
	// Registered WebSocket clients by channel ID
	channels map[string]map[*websocket.Conn]string // channel_id -> conn -> user_id
	// Registered SSE subscribers by channel ID
	sseSubs map[string]map[chan []byte]string // channel_id -> ch -> user_id
	mu      sync.RWMutex
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		channels: make(map[string]map[*websocket.Conn]string),
		sseSubs:  make(map[string]map[chan []byte]string),
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
	sse := make([]chan []byte, 0, len(h.sseSubs[channelID]))
	for ch := range h.sseSubs[channelID] {
		sse = append(sse, ch)
	}
	h.mu.RUnlock()

	log.Info().Str("channel_id", channelID).Int("subscribers", len(conns)).Int("sse_subscribers", len(sse)).Msg("hub broadcast")

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
	for _, ch := range sse {
		select {
		case ch <- append([]byte(nil), message...):
		default:
		}
	}
}

// SubscribeSSE adds an SSE subscriber channel to one logical channel.
func (h *Hub) SubscribeSSE(channelID, userID string, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.sseSubs[channelID] == nil {
		h.sseSubs[channelID] = make(map[chan []byte]string)
	}
	h.sseSubs[channelID][ch] = userID
}

// UnsubscribeSSE removes an SSE subscriber channel from one logical channel.
func (h *Hub) UnsubscribeSSE(channelID string, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.sseSubs[channelID] == nil {
		return
	}
	delete(h.sseSubs[channelID], ch)
	if len(h.sseSubs[channelID]) == 0 {
		delete(h.sseSubs, channelID)
	}
}
