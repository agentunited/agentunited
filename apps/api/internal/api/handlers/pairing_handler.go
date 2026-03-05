package handlers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/agentunited/backend/internal/relay"
	"github.com/rs/zerolog/log"
)

var relayTokenRegex = regexp.MustCompile(`^rt_[A-Za-z0-9_-]{8,}$`)

// PairingCode represents a 6-digit code linked to a workspace instance
type PairingCode struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}

// PairingHandler handles instance pairing and admin relay token updates.
type PairingHandler struct {
	codes     map[string]PairingCode
	codesLock sync.RWMutex
}

// NewPairingHandler creates a new pairing handler.
func NewPairingHandler() *PairingHandler {
	return &PairingHandler{
		codes:     make(map[string]PairingCode),
		codesLock: sync.RWMutex{},
	}
}

// GetCode generates or returns a valid 6-digit pairing code.
func (h *PairingHandler) GetCode(w http.ResponseWriter, r *http.Request) {
	h.codesLock.Lock()
	defer h.codesLock.Unlock()

	now := time.Now()
	for k, v := range h.codes {
		if now.After(v.ExpiresAt) {
			delete(h.codes, k)
		}
	}

	rand.Seed(time.Now().UnixNano())
	codeChars := "0123456789ABCDEF"
	code := ""
	for i := 0; i < 6; i++ {
		code += string(codeChars[rand.Intn(len(codeChars))])
	}

	pairingCode := PairingCode{
		Code:      code,
		ExpiresAt: now.Add(15 * time.Minute),
	}

	h.codes[code] = pairingCode

	log.Info().Str("code", code).Msg("generated new pairing code")

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pairingCode)
}

// VerifyCode checks if a code is valid.
func (h *PairingHandler) VerifyCode(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	h.codesLock.RLock()
	v, ok := h.codes[code]
	h.codesLock.RUnlock()

	if !ok || time.Now().After(v.ExpiresAt) {
		http.Error(w, "invalid or expired code", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "verified",
		"message": "instance paired successfully",
	})
}

type AdminPairingRequest struct {
	RelayToken string `json:"relay_token"`
}

// AdminPairing updates relay token, persists it, and triggers embedded relay reconnect.
// POST /api/v1/admin/pairing
func (h *PairingHandler) AdminPairing(w http.ResponseWriter, r *http.Request) {
	var req AdminPairingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	if !relayTokenRegex.MatchString(req.RelayToken) {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid relay token format"})
		return
	}

	// Update in-process state for current runtime.
	_ = os.Setenv("DEPLOYMENT_MODE", "tunnel")
	_ = os.Setenv("RELAY_TOKEN", req.RelayToken)

	// Persist token so it survives restart.
	cfgPath := os.Getenv("RELAY_CONFIG_FILE")
	if cfgPath == "" {
		cfgPath = "data/relay_config.json"
	}
	if err := persistRelayConfig(cfgPath, req.RelayToken); err != nil {
		log.Error().Err(err).Str("path", cfgPath).Msg("failed to persist relay config")
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Failed to persist pairing config"})
		return
	}

	// Trigger (re)connect with new token.
	mgr := relay.GetGlobalManager()
	if mgr == nil {
		respondJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "Relay manager unavailable"})
		return
	}
	mgr.UpdateToken(req.RelayToken)

	respondJSON(w, http.StatusOK, map[string]any{
		"status":      "paired",
		"mode":        "tunnel",
		"config_path": cfgPath,
	})
}

type relayConfigFile struct {
	DeploymentMode string `json:"deployment_mode"`
	RelayToken     string `json:"relay_token"`
	UpdatedAt      string `json:"updated_at"`
}

func persistRelayConfig(path, token string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	payload := relayConfigFile{
		DeploymentMode: "tunnel",
		RelayToken:     token,
		UpdatedAt:      time.Now().UTC().Format(time.RFC3339),
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}
