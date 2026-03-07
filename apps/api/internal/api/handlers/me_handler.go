package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	mw "github.com/agentunited/backend/internal/api/middleware"
	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/service"
)

// MeHandler manages current-user profile endpoints.
type MeHandler struct {
	authService service.AuthService
}

func NewMeHandler(authService service.AuthService) *MeHandler {
	return &MeHandler{authService: authService}
}

type UpdateMeRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// GetMe handles GET /api/v1/me.
func (h *MeHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.GetUserID(r.Context())
	if !ok || userID == "" {
		respondJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	user, err := h.authService.GetCurrentUser(r.Context(), userID)
	if err != nil {
		h.handleMeError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// UpdateMe handles PUT /api/v1/me.
func (h *MeHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.GetUserID(r.Context())
	if !ok || userID == "" {
		respondJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	var req UpdateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	displayName := ""
	avatarURL := ""
	if req.DisplayName != nil {
		displayName = *req.DisplayName
	}
	if req.AvatarURL != nil {
		avatarURL = *req.AvatarURL
	}

	user, err := h.authService.UpdateProfile(r.Context(), userID, displayName, avatarURL)
	if err != nil {
		h.handleMeError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// ChangePassword handles POST /api/v1/me/password.
func (h *MeHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.GetUserID(r.Context())
	if !ok || userID == "" {
		respondJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}
	if req.CurrentPassword == "" || req.NewPassword == "" {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "current_password and new_password are required"})
		return
	}

	if err := h.authService.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		h.handleMeError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "password updated"})
}

func (h *MeHandler) handleMeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, models.ErrUserNotFound):
		respondJSON(w, http.StatusNotFound, ErrorResponse{Error: "User not found"})
	case errors.Is(err, models.ErrWeakPassword):
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Password must be at least 8 characters with 1 letter and 1 number"})
	case errors.Is(err, models.ErrInvalidCredentials):
		respondJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
	default:
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}
}
