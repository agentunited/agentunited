package relay

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type Client struct {
	relayURL   string
	token      string
	localAPI   string
	httpClient *http.Client

	writeMu sync.Mutex
	wsMu    sync.Mutex
	wsConns map[string]*websocket.Conn
}

func NewClient(relayURL, token, localAPI string) *Client {
	return &Client{
		relayURL: relayURL,
		token:    token,
		localAPI: localAPI,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		wsConns: make(map[string]*websocket.Conn),
	}
}

func (c *Client) Start(ctx context.Context) {
	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err := c.runSession(ctx)
		if err != nil {
			log.Warn().Err(err).Dur("retry_in", backoff).Msg("relay client disconnected; reconnecting")
		}

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return
		}

		backoff *= 2
		if backoff > 60*time.Second {
			backoff = 60 * time.Second
		}
	}
}

func (c *Client) runSession(ctx context.Context) error {
	conn, _, err := websocket.DefaultDialer.Dial(c.relayURL, nil)
	if err != nil {
		return fmt.Errorf("dial relay: %w", err)
	}
	defer conn.Close()

	reg := RegisterMessage{
		Type:         TypeRegister,
		Token:        c.token,
		Version:      "1.0",
		Capabilities: []string{"http", "websocket"},
	}
	if err := conn.WriteJSON(reg); err != nil {
		return fmt.Errorf("send register: %w", err)
	}

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read relay message: %w", err)
		}

		var env Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			continue
		}

		switch env.Type {
		case TypeRegistered:
			var m RegisteredMessage
			_ = json.Unmarshal(data, &m)
			log.Info().Str("public_url", m.URL).Str("subdomain", m.Subdomain).Msg("relay connected")
		case TypePing:
			if err := c.writeJSON(conn, Envelope{Type: TypePong}); err != nil {
				return err
			}
		case TypeRequest:
			var req RequestMessage
			if err := json.Unmarshal(data, &req); err != nil {
				continue
			}
			go c.handleRequest(ctx, conn, req)
		case TypeWSOpen:
			var open WSOpenMessage
			if err := json.Unmarshal(data, &open); err != nil {
				continue
			}
			go c.handleWSOpen(ctx, conn, open)
		case TypeWSData:
			var d WSDataMessage
			if err := json.Unmarshal(data, &d); err != nil {
				continue
			}
			c.handleWSData(d)
		case TypeWSClose:
			var cl WSCloseMessage
			if err := json.Unmarshal(data, &cl); err != nil {
				continue
			}
			c.handleWSClose(cl.ID)
		case TypeError:
			var er ErrorMessage
			_ = json.Unmarshal(data, &er)
			return fmt.Errorf("relay error: %s", er.Message)
		}
	}
}

func (c *Client) handleRequest(ctx context.Context, ws *websocket.Conn, req RequestMessage) {
	decodedBody := []byte{}
	if req.Body != "" {
		b, err := base64.StdEncoding.DecodeString(req.Body)
		if err == nil {
			decodedBody = b
		}
	}

	localURL := c.localAPI + req.Path
	hreq, err := http.NewRequestWithContext(ctx, req.Method, localURL, bytes.NewReader(decodedBody))
	if err != nil {
		_ = c.writeJSON(ws, ResponseMessage{Type: TypeResponse, ID: req.ID, Status: 500})
		return
	}
	hreq.Header = req.Headers.Clone()
	hreq.Host = ""

	hresp, err := c.httpClient.Do(hreq)
	if err != nil {
		_ = c.writeJSON(ws, ResponseMessage{Type: TypeResponse, ID: req.ID, Status: 502})
		return
	}
	defer hresp.Body.Close()

	respBody, _ := io.ReadAll(hresp.Body)
	resp := ResponseMessage{
		Type:    TypeResponse,
		ID:      req.ID,
		Status:  hresp.StatusCode,
		Headers: hresp.Header,
		Body:    base64.StdEncoding.EncodeToString(respBody),
	}
	_ = c.writeJSON(ws, resp)
}

func (c *Client) handleWSOpen(ctx context.Context, ws *websocket.Conn, open WSOpenMessage) {
	dialer := websocket.Dialer{}
	localURL := strings.Replace(c.localAPI, "http://", "ws://", 1) + open.Path
	localConn, _, err := dialer.DialContext(ctx, localURL, open.Headers)
	if err != nil {
		_ = c.writeJSON(ws, WSOpenedMessage{Type: TypeWSOpened, ID: open.ID, Success: false, Error: err.Error()})
		return
	}

	c.wsMu.Lock()
	c.wsConns[open.ID] = localConn
	c.wsMu.Unlock()

	_ = c.writeJSON(ws, WSOpenedMessage{Type: TypeWSOpened, ID: open.ID, Success: true})

	go func() {
		defer c.handleWSClose(open.ID)
		for {
			mt, payload, err := localConn.ReadMessage()
			if err != nil {
				_ = c.writeJSON(ws, WSCloseMessage{Type: TypeWSClose, ID: open.ID})
				return
			}
			_ = c.writeJSON(ws, WSDataMessage{Type: TypeWSData, ID: open.ID, MessageType: mt, Data: base64.StdEncoding.EncodeToString(payload)})
		}
	}()
}

func (c *Client) handleWSData(d WSDataMessage) {
	c.wsMu.Lock()
	conn := c.wsConns[d.ID]
	c.wsMu.Unlock()
	if conn == nil {
		return
	}
	payload, err := base64.StdEncoding.DecodeString(d.Data)
	if err != nil {
		return
	}
	_ = conn.WriteMessage(d.MessageType, payload)
}

func (c *Client) handleWSClose(id string) {
	c.wsMu.Lock()
	conn := c.wsConns[id]
	delete(c.wsConns, id)
	c.wsMu.Unlock()
	if conn != nil {
		_ = conn.Close()
	}
}

func (c *Client) writeJSON(ws *websocket.Conn, v any) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return ws.WriteJSON(v)
}
