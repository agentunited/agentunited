package handlers

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	mw "github.com/agentunited/backend/internal/api/middleware"
	"github.com/agentunited/backend/internal/service"
	"github.com/agentunited/backend/pkg/billing"
)

type BillingHandler struct {
	svc service.BillingService
}

type checkoutRequest struct {
	Email      string `json:"email"`
	Name       string `json:"name"`
	Plan       string `json:"plan"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

type portalRequest struct {
	ReturnURL string `json:"return_url"`
}

func NewBillingHandler(svc service.BillingService) *BillingHandler { return &BillingHandler{svc: svc} }

func (h *BillingHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	workspaceID, ok := mw.GetUserID(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	req := checkoutRequest{
		Email:      r.URL.Query().Get("email"),
		Name:       r.URL.Query().Get("name"),
		Plan:       r.URL.Query().Get("plan"),
		SuccessURL: r.URL.Query().Get("success_url"),
		CancelURL:  r.URL.Query().Get("cancel_url"),
	}
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON"})
			return
		}
	}

	if req.SuccessURL == "" {
		req.SuccessURL = "https://agentunited.ai/billing/success"
	}
	if req.CancelURL == "" {
		req.CancelURL = "https://agentunited.ai/billing/cancel"
	}

	url, err := h.svc.GetCheckoutURL(r.Context(), workspaceID, req.Email, req.Name, req.Plan, req.SuccessURL, req.CancelURL)
	if err != nil {
		if err == billing.ErrNotConfigured {
			respondJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "billing not configured"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"checkout_url": url})
}

func (h *BillingHandler) Portal(w http.ResponseWriter, r *http.Request) {
	workspaceID, ok := mw.GetUserID(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	req := portalRequest{ReturnURL: r.URL.Query().Get("return_url")}
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON"})
			return
		}
	}
	if req.ReturnURL == "" {
		req.ReturnURL = "https://agentunited.ai/settings/billing"
	}

	url, err := h.svc.GetPortalURL(r.Context(), workspaceID, req.ReturnURL)
	if err != nil {
		if err == billing.ErrNotConfigured {
			respondJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "billing not configured"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"portal_url": url})
}

func (h *BillingHandler) Status(w http.ResponseWriter, r *http.Request) {
	workspaceID, ok := mw.GetUserID(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	sub, entityCount, err := h.svc.GetStatus(r.Context(), workspaceID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	relaySubdomain := sub.RelaySubdomain
	if relaySubdomain == "" && sub.RelayToken != "" {
		h := sha1.Sum([]byte(sub.RelayToken))
		relaySubdomain = "w" + hex.EncodeToString(h[:])[:10]
	}
	relayEnabled := relaySubdomain != "" || sub.RelayTier != "free"

	respondJSON(w, http.StatusOK, map[string]any{
		"plan":                    sub.Plan,
		"subscription_status":     sub.Status,
		"entity_count":            entityCount,
		"entity_limit":            planEntityLimit(sub.Plan),
		"relay_enabled":           relayEnabled,
		"relay_hostname":          relaySubdomain,
		"relay_tier":              sub.RelayTier,
		"relay_subdomain":         relaySubdomain,
		"bandwidth_limit_mb":      sub.RelayBandwidthLimitMB,
		"subscription_period_end": sub.CurrentPeriodEnd,
		"stripe_customer_id":      sub.StripeCustomerID,
	})
}

func planEntityLimit(plan string) int {
	switch plan {
	case "pro":
		return 10
	case "team", "enterprise":
		return -1
	default:
		return 3
	}
}

func (h *BillingHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid body"})
		return
	}
	sig := r.Header.Get("Stripe-Signature")
	err = h.svc.HandleWebhook(r.Context(), body, sig)
	if err != nil {
		if err == service.ErrInvalidWebhookSignature {
			respondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid signature"})
			return
		}
		respondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
