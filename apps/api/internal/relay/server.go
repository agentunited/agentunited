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
	"net/textproto"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// safeWSForwardHeaders is the allowlist of headers the relay server includes in WSOpenMessage.
// Hop-by-hop and WS handshake headers (Connection, Upgrade, Sec-WebSocket-*) must be excluded
// because gorilla/websocket sets them automatically on the client side; duplicates cause 1013.
var safeWSForwardHeaders = map[string]bool{
	"Authorization":   true,
	"Cookie":          true,
	"X-Real-Ip":       true,
	"X-Forwarded-For": true,
}

func safeForwardHeaders(h http.Header) http.Header {
	out := http.Header{}
	for k, v := range h {
		if safeWSForwardHeaders[textproto.CanonicalMIMEHeaderKey(k)] {
			out[k] = v
		}
	}
	return out
}

type clientConn struct {
	id          string
	subdomain   string
	token       string
	workspaceID string
	conn        *websocket.Conn
	writeMu     sync.Mutex
	pending      sync.Map // request_id -> chan ResponseMessage
	pendingStart sync.Map // request_id -> chan ResponseStartMessage
	pendingChunk sync.Map // request_id -> chan ResponseChunkMessage
	pendingEnd   sync.Map // request_id -> chan struct{}
	wsOpen       sync.Map // ws_id -> chan WSOpenedMessage
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

	ctx := context.Background()
	plan := s.lookupPlan(ctx, reg.Token)

	// Enforce plan expiry.
	if plan.ExpiresAt != nil && time.Now().UTC().After(*plan.ExpiresAt) {
		_ = ws.WriteJSON(ErrorMessage{Type: TypeError, Message: "relay subscription expired — please renew"})
		_ = ws.Close()
		return
	}

	sub := plan.Subdomain
	if sub == "" {
		sub = deterministicSubdomain(reg.Token)
	}

	// Enforce connection limit (count existing connections for this subdomain).
	if active := s.activeConnCount(sub); active >= plan.ConnectionsMax {
		_ = ws.WriteJSON(ErrorMessage{Type: TypeError, Message: fmt.Sprintf("connection limit reached (%d/%d) — upgrade plan to connect more clients", active, plan.ConnectionsMax)})
		_ = ws.Close()
		return
	}

	connID := uuid.NewString()
	cc := &clientConn{id: connID, subdomain: sub, conn: ws, token: reg.Token, workspaceID: plan.WorkspaceID}

	s.mu.Lock()
	s.clients[connID] = cc
	s.mu.Unlock()

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
		case TypeResponseStart:
			var start ResponseStartMessage
			if err := json.Unmarshal(data, &start); err != nil {
				continue
			}
			if chRaw, ok := cc.pendingStart.Load(start.ID); ok {
				ch := chRaw.(chan ResponseStartMessage)
				select {
				case ch <- start:
				default:
				}
			}
		case TypeResponseChunk:
			var chunk ResponseChunkMessage
			if err := json.Unmarshal(data, &chunk); err != nil {
				continue
			}
			if chRaw, ok := cc.pendingChunk.Load(chunk.ID); ok {
				ch := chRaw.(chan ResponseChunkMessage)
				select {
				case ch <- chunk:
				default:
				}
			}
		case TypeResponseEnd:
			var end ResponseEndMessage
			if err := json.Unmarshal(data, &end); err != nil {
				continue
			}
			if chRaw, ok := cc.pendingEnd.Load(end.ID); ok {
				ch := chRaw.(chan struct{})
				select {
				case ch <- struct{}{}:
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

	plan := s.lookupPlan(ctx, cc.token)
	if plan.WorkspaceID == "" {
		plan.WorkspaceID = cc.workspaceID
	}
	// Enforce bandwidth cap before forwarding request.
	if plan.BandwidthLimitMB > 0 && plan.WorkspaceID != "" {
		if s.bandwidthUsedMB(ctx, plan.WorkspaceID) >= float64(plan.BandwidthLimitMB) {
			http.Error(w, "bandwidth limit exceeded. Upgrade to continue.", http.StatusTooManyRequests)
			return
		}
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
	startCh := make(chan ResponseStartMessage, 1)
	chunkCh := make(chan ResponseChunkMessage, 128)
	endCh := make(chan struct{}, 1)
	cc.pending.Store(reqID, respCh)
	cc.pendingStart.Store(reqID, startCh)
	cc.pendingChunk.Store(reqID, chunkCh)
	cc.pendingEnd.Store(reqID, endCh)
	defer cc.pending.Delete(reqID)
	defer cc.pendingStart.Delete(reqID)
	defer cc.pendingChunk.Delete(reqID)
	defer cc.pendingEnd.Delete(reqID)

	if err := cc.writeJSON(msg); err != nil {
		http.Error(w, "failed to route request", http.StatusBadGateway)
		return
	}

	var totalBytes int64
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
			totalBytes += int64(len(decoded))
		}
		if plan.WorkspaceID != "" && totalBytes > 0 {
			_ = s.trackBandwidth(ctx, plan.WorkspaceID, plan.BandwidthLimitMB, totalBytes)
		}
	case start := <-startCh:
		for k, vals := range start.Headers {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}
		w.Header().Set("X-Accel-Buffering", "no")
		w.WriteHeader(start.Status)
		flusher, ok := w.(http.Flusher)
		if !ok {
			return
		}
		flusher.Flush()
		for {
			select {
			case ch := <-chunkCh:
				if ch.Body == "" {
					continue
				}
				decoded, _ := base64.StdEncoding.DecodeString(ch.Body)
				if len(decoded) > 0 {
					_, _ = w.Write(decoded)
					totalBytes += int64(len(decoded))
					if plan.WorkspaceID != "" {
						_ = s.trackBandwidth(ctx, plan.WorkspaceID, plan.BandwidthLimitMB, int64(len(decoded)))
					}
					flusher.Flush()
				}
			case <-endCh:
				return
			case <-ctx.Done():
				return
			}
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

	openMsg := WSOpenMessage{Type: TypeWSOpen, ID: wsID, Path: r.URL.RequestURI(), Headers: safeForwardHeaders(r.Header)}
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

// planSnapshot is the decoded value of relay:token:{token} in Redis.
type planSnapshot struct {
	WorkspaceID      string     `json:"workspace_id"`
	Plan             string     `json:"plan"`
	RelayTier        string     `json:"relay_tier"`
	Subdomain        string     `json:"subdomain"`
	BandwidthLimitMB int        `json:"bandwidth_limit_mb"`
	ConnectionsMax   int        `json:"connections_max"`
	ExpiresAt        *time.Time `json:"expires_at"`
}

// freePlan returns safe fail-open defaults when no plan snapshot is found.
func freePlan(workspaceID string) planSnapshot {
	return planSnapshot{
		WorkspaceID:      workspaceID,
		Plan:             "free",
		RelayTier:        "free",
		BandwidthLimitMB: 1024,
		ConnectionsMax:   3,
	}
}

// lookupPlan reads plan snapshot from Redis by relay token. Fail-open: returns free plan on miss.
func (s *Server) lookupPlan(ctx context.Context, token string) planSnapshot {
	if token == "" {
		return freePlan("")
	}
	raw, err := s.redis.Get(ctx, "relay:token:"+token).Result()
	if err != nil {
		return freePlan("")
	}
	var snap planSnapshot
	if err := json.Unmarshal([]byte(raw), &snap); err != nil {
		return freePlan("")
	}
	if snap.BandwidthLimitMB == 0 {
		snap.BandwidthLimitMB = 1024
	}
	if snap.ConnectionsMax == 0 {
		snap.ConnectionsMax = 3
	}
	return snap
}

// activeConnCount returns the number of currently active connections for a subdomain.
func (s *Server) activeConnCount(sub string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, cc := range s.clients {
		if cc.subdomain == sub {
			count++
		}
	}
	return count
}

// bandwidthUsedMB returns MB used today for the workspace.
func (s *Server) bandwidthUsedMB(ctx context.Context, workspaceID string) float64 {
	key := fmt.Sprintf("relay:bw:%s:%s", workspaceID, time.Now().UTC().Format("2006-01-02"))
	v, _ := s.redis.Get(ctx, key).Int64()
	return float64(v) / (1024 * 1024)
}

// trackBandwidth increments the daily byte counter and returns false if limit is exceeded.
func (s *Server) trackBandwidth(ctx context.Context, workspaceID string, limitMB int, bytes int64) bool {
	if workspaceID == "" || limitMB <= 0 {
		return true
	}
	key := fmt.Sprintf("relay:bw:%s:%s", workspaceID, time.Now().UTC().Format("2006-01-02"))
	used, _ := s.redis.IncrBy(ctx, key, bytes).Result()
	// Set TTL on first write so key expires after 2 days.
	if used == bytes {
		_ = s.redis.Expire(ctx, key, 48*time.Hour).Err()
	}
	return used <= int64(limitMB)*1024*1024
}
