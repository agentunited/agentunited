package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/agentunited/backend/internal/models"
	mw "github.com/agentunited/backend/internal/api/middleware"
	"github.com/agentunited/backend/internal/service"
	"github.com/go-chi/chi/v5"
)

// WebhookHandler handles webhook HTTP requests
type WebhookHandler struct {
	service service.WebhookService
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(service service.WebhookService) *WebhookHandler {
	return &WebhookHandler{service: service}
}

// Create handles webhook creation
func (h *WebhookHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	agentID := chi.URLParam(r, "id")

	var req models.CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	webhook, err := h.service.CreateWebhook(r.Context(), agentID, userID, &req)
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
	json.NewEncoder(w).Encode(webhook)
}

// List handles listing agent's webhooks
func (h *WebhookHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	agentID := chi.URLParam(r, "id")

	webhooks, err := h.service.ListWebhooks(r.Context(), agentID, userID)
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

	if webhooks == nil {
		webhooks = []*models.Webhook{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"webhooks": webhooks})
}

// Delete handles webhook deletion
func (h *WebhookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	agentID := chi.URLParam(r, "id")
	webhookID := chi.URLParam(r, "webhook_id")

	err := h.service.DeleteWebhook(r.Context(), webhookID, agentID, userID)
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

// ListDeliveries handles listing webhook deliveries
func (h *WebhookHandler) ListDeliveries(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	agentID := chi.URLParam(r, "id")
	webhookID := chi.URLParam(r, "webhook_id")

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	deliveries, err := h.service.ListDeliveries(r.Context(), webhookID, agentID, userID, limit)
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

	if deliveries == nil {
		deliveries = []*models.WebhookDelivery{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"deliveries": deliveries})
}
