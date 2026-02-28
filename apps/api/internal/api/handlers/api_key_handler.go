package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/service"
	mw "github.com/agentunited/backend/internal/api/middleware"
	"github.com/go-chi/chi/v5"
)

// APIKeyHandler handles API key HTTP requests
type APIKeyHandler struct {
	service service.APIKeyService
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(service service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{service: service}
}

// Create handles API key creation
func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	agentID := chi.URLParam(r, "agent_id")

	var req models.CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	key, err := h.service.CreateKey(r.Context(), agentID, userID, req.Name)
	if err != nil {
		if errors.Is(err, service.ErrAgentNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrNotAgentOwner) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(key)
}

// List handles listing agent's API keys
func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	agentID := chi.URLParam(r, "agent_id")

	keys, err := h.service.ListKeys(r.Context(), agentID, userID)
	if err != nil {
		if errors.Is(err, service.ErrAgentNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrNotAgentOwner) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if keys == nil {
		keys = []*models.APIKey{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"keys": keys})
}

// Delete handles API key deletion (revoke)
func (h *APIKeyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	agentID := chi.URLParam(r, "agent_id")
	keyID := chi.URLParam(r, "key_id")

	err := h.service.DeleteKey(r.Context(), keyID, agentID, userID)
	if err != nil {
		if errors.Is(err, service.ErrAgentNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrNotAgentOwner) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
