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
	"os"
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
	wsOpen    sync.Map // ws_id -> chan WSOpenedMessage
	wsData    sync.Map // ws_id -> chan WSDataMessage
	wsClosed  sync.Map // ws_id -> chan struct{}
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
	jwks     *JWKSVerifier

	mu      sync.RWMutex
	clients map[string]*clientConn // conn_id -> conn
}

// defaultJWKSURL returns the JWKS URL from env or falls back to the central service default.
func defaultJWKSURL() string {
	if u := os.Getenv("CENTRAL_JWKS_URL"); u != "" {
		return u
	}
	return "https://agent-united-central-961985674922.us-central1.run.app/.well-known/jwks.json"
}

func NewServer(redisClient *redis.Client, domain string) *Server {
	issuer := os.Getenv("CENTRAL_JWT_ISSUER")
	if issuer == "" {
		issuer = "https://agentunited.ai"
	}
	return &Server{
		domain: domain,
		redis:  redisClient,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients: make(map[string]*clientConn),
		jwks:    NewJWKSVerifier(defaultJWKSURL(), issuer),
	}
}

// Start fetches JWKS and begins the background refresh loop.
// Call this after NewServer, before serving traffic.
func (s *Server) Start(ctx context.Context) {
	_ = s.jwks.FetchOnce(ctx) // errors are logged + degrade-graceful; non-fatal
	s.jwks.StartRefreshLoop(ctx)
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

	// Central JWT verification — only for JWT-looking tokens.
	// Legacy relay tokens (rt_xxx) bypass JWT verification.
	if strings.Contains(reg.Token, ".") {
		if s.jwks.Enabled {
			if _, ok := s.jwks.VerifyJWT(reg.Token); !ok {
				_ = ws.WriteJSON(ErrorMessage{Type: TypeError, Message: "authorization required"})
				_ = ws.Close()
				return
			}
		} else {
			log.Warn().Msg("JWKS unavailable — allowing relay connection without central JWT verification")
		}
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
		case TypeWSOpened:
			var opened WSOpenedMessage
			if err := json.Unmarshal(data, &opened); err != nil {
				continue
			}
			if chRaw, ok := cc.wsOpen.Load(opened.ID); ok {
				ch := chRaw.(chan WSOpenedMessage)
				select {
				case ch <- opened:
				default:
				}
			}
		case TypeWSData:
			var wsData WSDataMessage
			if err := json.Unmarshal(data, &wsData); err != nil {
				continue
			}
			if chRaw, ok := cc.wsData.Load(wsData.ID); ok {
				ch := chRaw.(chan WSDataMessage)
				select {
				case ch <- wsData:
				default:
				}
			}
		case TypeWSClose:
			var wsClose WSCloseMessage
			if err := json.Unmarshal(data, &wsClose); err != nil {
				continue
			}
			if chRaw, ok := cc.wsClosed.Load(wsClose.ID); ok {
				ch := chRaw.(chan struct{})
				select {
				case ch <- struct{}{}:
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

	if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		s.handlePublicWebSocket(w, r, cc)
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

func (s *Server) handlePublicWebSocket(w http.ResponseWriter, r *http.Request, cc *clientConn) {
	pubWS, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("public websocket upgrade failed")
		return
	}
	defer pubWS.Close()

	wsID := "ws_" + uuid.NewString()
	openCh := make(chan WSOpenedMessage, 1)
	dataCh := make(chan WSDataMessage, 64)
	closeCh := make(chan struct{}, 1)
	cc.wsOpen.Store(wsID, openCh)
	cc.wsData.Store(wsID, dataCh)
	cc.wsClosed.Store(wsID, closeCh)
	defer cc.wsOpen.Delete(wsID)
	defer cc.wsData.Delete(wsID)
	defer cc.wsClosed.Delete(wsID)

	openMsg := WSOpenMessage{Type: TypeWSOpen, ID: wsID, Path: r.URL.RequestURI(), Headers: r.Header}
	if err := cc.writeJSON(openMsg); err != nil {
		return
	}

	select {
	case opened := <-openCh:
		if !opened.Success {
			_ = pubWS.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseTryAgainLater, opened.Error))
			return
		}
	case <-time.After(10 * time.Second):
		_ = pubWS.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseTryAgainLater, "upstream ws open timeout"))
		return
	}

	errCh := make(chan error, 2)
	go func() {
		for {
			mt, payload, err := pubWS.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}
			msg := WSDataMessage{Type: TypeWSData, ID: wsID, MessageType: mt, Data: base64.StdEncoding.EncodeToString(payload)}
			if err := cc.writeJSON(msg); err != nil {
				errCh <- err
				return
			}
		}
	}()

	for {
		select {
		case d := <-dataCh:
			payload, _ := base64.StdEncoding.DecodeString(d.Data)
			if err := pubWS.WriteMessage(d.MessageType, payload); err != nil {
				return
			}
		case <-closeCh:
			_ = pubWS.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		case <-errCh:
			_ = cc.writeJSON(WSCloseMessage{Type: TypeWSClose, ID: wsID})
			return
		}
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
