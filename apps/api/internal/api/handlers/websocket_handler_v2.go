package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/agentunited/backend/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

// WebSocketHandlerV2 handles WebSocket connections with broadcasting.
type WebSocketHandlerV2 struct {
	jwtSecret      string
	upgrader       websocket.Upgrader
	hub            *Hub
	messageService service.MessageService
	channelService service.ChannelService
}

// NewWebSocketHandlerV2 creates a new WebSocket handler with hub.
func NewWebSocketHandlerV2(messageService service.MessageService, channelService service.ChannelService, jwtSecret string, hub *Hub) *WebSocketHandlerV2 {
	return &WebSocketHandlerV2{
		jwtSecret:      jwtSecret,
		messageService: messageService,
		channelService: channelService,
		hub:            hub,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

type wsMessage struct {
	Type        string `json:"type"`
	ChannelID   string `json:"channel_id,omitempty"`
	Text        string `json:"text,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	MessageID   string `json:"message_id,omitempty"`
	AuthorID    string `json:"author_id,omitempty"`
	AuthorType  string `json:"author_type,omitempty"`
	AuthorEmail string `json:"author_email,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

func (h *WebSocketHandlerV2) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/ws" {
		http.NotFound(w, r)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	claims, err := h.validateToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("websocket upgrade failed")
		return
	}
	defer func() {
		h.hub.Unsubscribe(conn)
		conn.Close()
	}()

	// Set up ping/pong
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Send connected ack
	h.send(conn, wsMessage{Type: "connected", UserID: claims.UserID})

	// Start ping ticker
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	// Read loop
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Msg("websocket error")
			}
			return
		}

		var msg wsMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "subscribe":
			if msg.ChannelID == "" {
				continue
			}
			h.hub.Subscribe(r.Context(), msg.ChannelID, conn, claims.UserID)
			h.send(conn, wsMessage{Type: "subscribed", ChannelID: msg.ChannelID})

		case "send_message":
			if msg.ChannelID == "" || msg.Text == "" {
				continue
			}
			
			// Persist message
			message, err := h.messageService.Send(r.Context(), msg.ChannelID, claims.UserID, msg.Text)
			if err != nil {
				h.send(conn, wsMessage{Type: "error", Text: err.Error()})
				continue
			}

			// Use email from JWT claims since message.AuthorEmail isn't populated on create
			authorEmail := claims.Email
			authorType := message.AuthorType
			if authorType == "" {
				authorType = "user"
			}

			// Broadcast to all subscribers
			broadcast := wsMessage{
				Type:        "message",
				ChannelID:   msg.ChannelID,
				MessageID:   message.ID,
				Text:        message.Text,
				UserID:      message.AuthorID,
				AuthorID:    message.AuthorID,
				AuthorType:  authorType,
				AuthorEmail: authorEmail,
				CreatedAt:   message.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
			broadcastData, _ := json.Marshal(broadcast)
			h.hub.Broadcast(r.Context(), msg.ChannelID, broadcastData)
		}
	}
}

func (h *WebSocketHandlerV2) send(conn *websocket.Conn, msg wsMessage) {
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	data, _ := json.Marshal(msg)
	conn.WriteMessage(websocket.TextMessage, data)
}

func (h *WebSocketHandlerV2) validateToken(tokenString string) (*service.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &service.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	claims, ok := token.Claims.(*service.JWTClaims)
	if !ok {
		return nil, jwt.ErrTokenMalformed
	}
	return claims, nil
}
