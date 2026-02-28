package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/agentunited/backend/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// WebSocketHandler handles WebSocket connections.
type WebSocketHandler struct {
	jwtSecret string
	upgrader  websocket.Upgrader
}

// NewWebSocketHandler creates a new WebSocket handler.
// messageService/channelService are reserved for later Part 3 steps.
func NewWebSocketHandler(_ service.MessageService, _ service.ChannelService, jwtSecret string) http.Handler {
	return &WebSocketHandler{
		jwtSecret: jwtSecret,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	defer conn.Close()

	ack, _ := json.Marshal(map[string]string{
		"type":    "connected",
		"user_id": claims.UserID,
	})
	if err := conn.WriteMessage(websocket.TextMessage, ack); err != nil {
		return
	}

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *WebSocketHandler) validateToken(tokenString string) (*service.JWTClaims, error) {
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
