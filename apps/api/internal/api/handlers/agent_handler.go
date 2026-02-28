package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/service"
	"github.com/go-chi/chi/v5"
)

// AgentHandler handles agent HTTP requests
type AgentHandler struct {
	service service.AgentService
}

// NewAgentHandler creates a new agent handler
func NewAgentHandler(service service.AgentService) *AgentHandler {
	return &AgentHandler{service: service}
}

// Create handles agent creation
func (h *AgentHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	var req models.CreateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	agent, err := h.service.CreateAgent(r.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, service.ErrAgentNameTaken) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(agent)
}

// Get handles getting a single agent
func (h *AgentHandler) Get(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")

	agent, err := h.service.GetAgent(r.Context(), agentID)
	if err != nil {
		http.Error(w, "agent not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent)
}

// List handles listing user's agents
func (h *AgentHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	agents, err := h.service.ListAgents(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if agents == nil {
		agents = []*models.Agent{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"agents": agents})
}

// Update handles agent updates
func (h *AgentHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	agentID := chi.URLParam(r, "id")

	var req models.UpdateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	agent, err := h.service.UpdateAgent(r.Context(), agentID, userID, &req)
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
	json.NewEncoder(w).Encode(agent)
}

// Delete handles agent deletion
func (h *AgentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	agentID := chi.URLParam(r, "id")

	err := h.service.DeleteAgent(r.Context(), agentID, userID)
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
