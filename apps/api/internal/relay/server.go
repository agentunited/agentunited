package relay

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type clientConn struct {
	id        string
	subdomain string
	conn      *websocket.Conn
	writeMu   sync.Mutex
	pending   sync.Map // request_id -> chan ResponseMessage
}

func (c *clientConn) writeJSON(v any) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.WriteJSON(v)
}

type Server struct {
	domain   string
	redis    *redis.Client
	upgrader websocket.Upgrader

	mu      sync.RWMutex
	clients map[string]*clientConn // conn_id -> conn
}

func NewServer(redisClient *redis.Client, domain string) *Server {
	return &Server{
		domain: domain,
		redis:  redisClient,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients: make(map[string]*clientConn),
	}
}

func (s *Server) Routes(mux *http.ServeMux) {
	mux.HandleFunc("/health", s.health)
	mux.HandleFunc("/tunnel", s.handleTunnel)
	mux.HandleFunc("/", s.handlePublicHTTP)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok","service":"relay"}`))
}

func (s *Server) handleTunnel(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("relay upgrade failed")
		return
	}

	var reg RegisterMessage
	if err := ws.ReadJSON(&reg); err != nil || reg.Type != TypeRegister || reg.Token == "" {
		_ = ws.WriteJSON(ErrorMessage{Type: TypeError, Message: "invalid register payload"})
		_ = ws.Close()
		return
	}
	if !strings.HasPrefix(reg.Token, "rt_") {
		_ = ws.WriteJSON(ErrorMessage{Type: TypeError, Message: "invalid token"})
		_ = ws.Close()
		return
	}

	sub := deterministicSubdomain(reg.Token)
	connID := uuid.NewString()
	cc := &clientConn{id: connID, subdomain: sub, conn: ws}

	s.mu.Lock()
	s.clients[connID] = cc
	s.mu.Unlock()

	ctx := context.Background()
	_ = s.redis.Set(ctx, s.redisKeyForSub(sub), connID, 2*time.Minute).Err()

	if err := cc.writeJSON(RegisteredMessage{
		Type:      TypeRegistered,
		Subdomain: sub,
		URL:       fmt.Sprintf("https://%s.%s", sub, s.domain),
	}); err != nil {
		s.removeClient(connID)
		return
	}

	go s.heartbeat(ctx, cc)
	s.readLoop(ctx, cc)
}

func (s *Server) heartbeat(ctx context.Context, cc *clientConn) {
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	for range t.C {
		if err := cc.writeJSON(Envelope{Type: TypePing}); err != nil {
			s.removeClient(cc.id)
			return
		}
		_ = s.redis.Expire(ctx, s.redisKeyForSub(cc.subdomain), 2*time.Minute).Err()
	}
}

func (s *Server) readLoop(_ context.Context, cc *clientConn) {
	defer s.removeClient(cc.id)
	for {
		_, data, err := cc.conn.ReadMessage()
		if err != nil {
			return
		}

		var env Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			continue
		}

		switch env.Type {
		case TypePong:
			continue
		case TypeResponse:
			var resp ResponseMessage
			if err := json.Unmarshal(data, &resp); err != nil {
				continue
			}
			if chRaw, ok := cc.pending.Load(resp.ID); ok {
				ch := chRaw.(chan ResponseMessage)
				select {
				case ch <- resp:
				default:
				}
			}
		}
	}
}

func (s *Server) handlePublicHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/health" || r.URL.Path == "/tunnel" {
		http.NotFound(w, r)
		return
	}

	// Relay currently supports HTTP request/response tunneling only.
	// Explicitly reject WebSocket upgrades to avoid long 504 timeouts.
	if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUpgradeRequired)
		_, _ = w.Write([]byte(`{"error":"websocket upgrades are not supported over relay"}`))
		return
	}

	sub, ok := s.subdomainFromHost(r.Host)
	if !ok {
		http.Error(w, "invalid tunnel host", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	connID, err := s.redis.Get(ctx, s.redisKeyForSub(sub)).Result()
	if err != nil || connID == "" {
		http.Error(w, "workspace offline", http.StatusBadGateway)
		return
	}

	cc := s.getClient(connID)
	if cc == nil {
		http.Error(w, "workspace offline", http.StatusBadGateway)
		return
	}

	body, _ := io.ReadAll(r.Body)
	reqID := "req_" + uuid.NewString()
	msg := RequestMessage{
		Type:    TypeRequest,
		ID:      reqID,
		Method:  r.Method,
		Path:    r.URL.RequestURI(),
		Headers: r.Header,
		Body:    base64.StdEncoding.EncodeToString(body),
	}

	respCh := make(chan ResponseMessage, 1)
	cc.pending.Store(reqID, respCh)
	defer cc.pending.Delete(reqID)

	if err := cc.writeJSON(msg); err != nil {
		http.Error(w, "failed to route request", http.StatusBadGateway)
		return
	}

	select {
	case resp := <-respCh:
		for k, vals := range resp.Headers {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.Status)
		if resp.Body != "" {
			decoded, _ := base64.StdEncoding.DecodeString(resp.Body)
			_, _ = w.Write(decoded)
		}
	case <-time.After(30 * time.Second):
		http.Error(w, "upstream timeout", http.StatusGatewayTimeout)
	}
}

func (s *Server) removeClient(connID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c, ok := s.clients[connID]; ok {
		_ = c.conn.Close()
		delete(s.clients, connID)
	}
}

func (s *Server) getClient(connID string) *clientConn {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clients[connID]
}

func (s *Server) redisKeyForSub(sub string) string {
	return "relay:subdomain:" + sub
}

func (s *Server) subdomainFromHost(host string) (string, bool) {
	h := host
	if strings.Contains(h, ":") {
		h = strings.Split(h, ":")[0]
	}
	suffix := "." + s.domain
	if !strings.HasSuffix(h, suffix) {
		return "", false
	}
	sub := strings.TrimSuffix(h, suffix)
	if sub == "" || strings.Contains(sub, ".") {
		return "", false
	}
	return sub, true
}

func deterministicSubdomain(token string) string {
	h := sha1.Sum([]byte(token))
	return "w" + hex.EncodeToString(h[:])[:10]
}
