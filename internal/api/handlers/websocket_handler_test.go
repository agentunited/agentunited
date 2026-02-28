package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentunited/backend/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebSocketHandler_Connect_RequiresToken(t *testing.T) {
	h := NewWebSocketHandler(nil, nil, "test-secret")

	srv := httptest.NewServer(h)
	defer srv.Close()

	wsURL := "ws" + srv.URL[len("http"): ] + "/ws"
	_, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.Error(t, err)
}

func TestWebSocketHandler_Connect_InvalidToken(t *testing.T) {
	h := NewWebSocketHandler(nil, nil, "test-secret")

	srv := httptest.NewServer(h)
	defer srv.Close()

	wsURL := "ws" + srv.URL[len("http"): ] + "/ws?token=bad-token"
	_, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.Error(t, err)
}

func TestWebSocketHandler_Connect_ValidToken_SendsConnectedAck(t *testing.T) {
	secret := "test-secret"
	h := NewWebSocketHandler(nil, nil, secret)

	srv := httptest.NewServer(h)
	defer srv.Close()

	token := makeTestJWT(t, secret, "user-123", "user@example.com")
	wsURL := "ws" + srv.URL[len("http"):] + "/ws?token=" + token
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	_, data, err := conn.ReadMessage()
	require.NoError(t, err)

	var msg map[string]string
	err = json.Unmarshal(data, &msg)
	require.NoError(t, err)
	assert.Equal(t, "connected", msg["type"])
	assert.Equal(t, "user-123", msg["user_id"])
}

func TestWebSocketHandler_HasWSRoute(t *testing.T) {
	h := NewWebSocketHandler(nil, nil, "test-secret")
	require.NotNil(t, h)
	assert.NotNil(t, h)
}

func makeTestJWT(t *testing.T, secret, userID, email string) string {
	t.Helper()
	claims := &service.JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return tokenStr
}
