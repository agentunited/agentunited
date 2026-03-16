package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	mw "github.com/agentunited/backend/internal/api/middleware"
	"github.com/agentunited/backend/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

// InviteService defines invite operations
type InviteService interface {
	ValidateInvite(ctx context.Context, token string) (*models.Invite, *models.User, error)
	AcceptInvite(ctx context.Context, token, password, displayName string) (string, string, error)
	CreateInvite(ctx context.Context, workspaceID, email, displayName string) (string, string, error)
}

// InviteHandler handles invite requests
type InviteHandler struct {
	service   InviteService
	validator *validator.Validate
}

// NewInviteHandler creates a new invite handler
func NewInviteHandler(service InviteService) *InviteHandler {
	return &InviteHandler{
		service:   service,
		validator: validator.New(),
	}
}

// ValidateInvite handles GET /api/v1/invite?token=...
func (h *InviteHandler) ValidateInvite(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		respondError(w, http.StatusBadRequest, "token is required")
		return
	}

	invite, user, err := h.service.ValidateInvite(r.Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrInviteNotFound):
			respondError(w, http.StatusNotFound, "invite not found")
		case errors.Is(err, models.ErrInviteExpired):
			respondError(w, http.StatusNotFound, "invite has expired")
		case errors.Is(err, models.ErrInviteAlreadyConsumed):
			respondError(w, http.StatusNotFound, "invite has already been used")
		default:
			log.Error().Err(err).Msg("validate invite failed")
			respondError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	// Return invite info
	resp := map[string]interface{}{
		"email":      user.Email,
		"status":     string(invite.Status),
		"expires_at": invite.ExpiresAt,
	}

	respondJSON(w, http.StatusOK, resp)
}

// AcceptInvite handles POST /api/v1/invite/accept
func (h *InviteHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	workspaceID, ok := mw.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.InviteCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Role == "" {
		req.Role = "member"
	}
	if req.Role != "member" {
		respondError(w, http.StatusBadRequest, "role must be member")
		return
	}
	if err := h.validator.Struct(&req); err != nil {
		respondError(w, http.StatusBadRequest, "validation failed: "+err.Error())
		return
	}

	token, inviteURL, err := h.service.CreateInvite(r.Context(), workspaceID, req.Email, req.DisplayName)
	if err != nil {
		if errors.Is(err, models.ErrEntityLimitReached) {
			respondJSON(w, http.StatusPaymentRequired, ErrorResponse{Error: "entity_limit_reached"})
			return
		}
		log.Error().Err(err).Msg("create invite failed")
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{
		"invite_token": token,
		"invite_url":   inviteURL,
	})
}

func (h *InviteHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	var req models.InviteAcceptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Debug().Err(err).Msg("failed to decode invite accept request")
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		log.Debug().Err(err).Msg("invite accept request validation failed")
		respondError(w, http.StatusBadRequest, "validation failed: "+err.Error())
		return
	}

	// Accept invite
	jwtToken, dmChannelID, err := h.service.AcceptInvite(r.Context(), req.Token, req.Password, req.DisplayName)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrInviteNotFound):
			respondError(w, http.StatusNotFound, "invite not found")
		case errors.Is(err, models.ErrInviteExpired):
			respondError(w, http.StatusNotFound, "invite has expired")
		case errors.Is(err, models.ErrInviteAlreadyConsumed):
			respondError(w, http.StatusConflict, "invite has already been used")
		case errors.Is(err, models.ErrWeakPassword):
			respondError(w, http.StatusBadRequest, "password must be at least 12 characters")
		default:
			log.Error().Err(err).Msg("accept invite failed")
			respondError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	// Return JWT token
	resp := map[string]string{
		"jwt_token": jwtToken,
		"message":   "invite accepted successfully",
	}
	if dmChannelID != "" {
		resp["dm_channel_id"] = dmChannelID
	}

	respondJSON(w, http.StatusOK, resp)
}
