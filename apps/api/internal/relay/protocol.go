package relay

import "net/http"

// Protocol message types.
const (
	TypeRegister   = "register"
	TypeRegistered = "registered"
	TypeRequest    = "request"
	TypeResponse      = "response"
	TypeResponseStart = "response_start"
	TypeResponseChunk = "response_chunk"
	TypeResponseEnd   = "response_end"
	TypePing          = "ping"
	TypePong          = "pong"
	TypeError         = "error"
	TypeWSOpen        = "ws_open"
	TypeWSOpened   = "ws_opened"
	TypeWSData     = "ws_data"
	TypeWSClose    = "ws_close"
)

type Envelope struct {
	Type string `json:"type"`
}

type RegisterMessage struct {
	Type         string   `json:"type"`
	Token        string   `json:"token"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
}

type RegisteredMessage struct {
	Type      string `json:"type"`
	Subdomain string `json:"subdomain"`
	URL       string `json:"url"`
}

type RequestMessage struct {
	Type    string      `json:"type"`
	ID      string      `json:"id"`
	Method  string      `json:"method"`
	Path    string      `json:"path"`
	Headers http.Header `json:"headers"`
	Body    string      `json:"body"` // base64
}

type ResponseMessage struct {
	Type    string      `json:"type"`
	ID      string      `json:"id"`
	Status  int         `json:"status"`
	Headers http.Header `json:"headers"`
	Body    string      `json:"body"` // base64
}

type ResponseStartMessage struct {
	Type    string      `json:"type"`
	ID      string      `json:"id"`
	Status  int         `json:"status"`
	Headers http.Header `json:"headers"`
}

type ResponseChunkMessage struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Body string `json:"body"` // base64
}

type ResponseEndMessage struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type ErrorMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type WSOpenMessage struct {
	Type    string      `json:"type"`
	ID      string      `json:"id"`
	Path    string      `json:"path"`
	Headers http.Header `json:"headers"`
}

type WSOpenedMessage struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type WSDataMessage struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	MessageType int    `json:"message_type"`
	Data        string `json:"data"` // base64
}

type WSCloseMessage struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}
