package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

type Handler struct {
	secret string
	store  SubscriptionStore
	cache  Cache
	email  EmailSender
}

func NewHandler(secret string, store SubscriptionStore, cache Cache, email EmailSender) *Handler {
	return &Handler{secret: secret, store: store, cache: cache, email: email}
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondJSON(w, http.StatusMethodNotAllowed, `{"error":"method not allowed"}`)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 512*1024))
	if err != nil {
		h.respondJSON(w, http.StatusBadRequest, `{"error":"invalid body"}`)
		return
	}
	event, err := webhook.ConstructEventWithOptions(body, r.Header.Get("Stripe-Signature"), h.secret, webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true})
	if err != nil {
		h.respondJSON(w, http.StatusBadRequest, `{"error":"invalid signature"}`)
		return
	}
	processed, err := h.cache.AlreadyProcessed(r.Context(), event.ID)
	if err == nil && processed {
		h.respondJSON(w, http.StatusOK, `{"status":"ok"}`)
		return
	}
	if err != nil {
		log.Warn().Err(err).Str("event_id", event.ID).Msg("idempotency check skipped — Redis unavailable")
	}

	sub, ok := h.extract(event)
	if !ok {
		log.Warn().Str("event_id", event.ID).Str("type", string(event.Type)).Msg("no workspace_id — skipping")
		h.respondJSON(w, http.StatusOK, `{"status":"ok"}`)
		return
	}

	if err := h.store.UpsertSubscription(r.Context(), sub); err != nil {
		log.Error().Err(err).Str("event_id", event.ID).Msg("DB upsert failed")
		h.respondJSON(w, http.StatusInternalServerError, `{"error":"internal error"}`)
		return
	}
	if err := h.cache.UpdateRelayPlan(r.Context(), sub.WorkspaceID, sub); err != nil {
		log.Error().Err(err).Str("workspace_id", sub.WorkspaceID).Msg("relay cache update failed")
	}
	if err := h.cache.MarkProcessed(r.Context(), event.ID, 24*time.Hour); err != nil {
		log.Warn().Err(err).Str("event_id", event.ID).Msg("failed to write idempotency key")
	}
	if event.Type == stripe.EventTypeCheckoutSessionCompleted {
		if email, err := h.store.GetWorkspaceEmail(r.Context(), sub.WorkspaceID); err == nil {
			_ = h.email.SendUpgradeConfirmation(r.Context(), email, sub.Plan)
		}
	}
	log.Info().Str("event_id", event.ID).Str("type", string(event.Type)).Str("workspace_id", sub.WorkspaceID).Str("plan", sub.Plan).Str("status", sub.Status).Msg("webhook processed")
	h.respondJSON(w, http.StatusOK, `{"status":"ok"}`)
}

func (h *Handler) respondJSON(w http.ResponseWriter, code int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(body))
}

func (h *Handler) extract(event stripe.Event) (SubscriptionRecord, bool) {
	obj := event.Data.Object
	metadata := mapVal(obj, "metadata")
	wsID := ""
	sub := SubscriptionRecord{Plan: "free", Status: "active", RelayTier: "free", BandwidthLimitMB: 1024, EntityLimit: 3, RelayEnabled: false}

	switch event.Type {
	case stripe.EventTypeCheckoutSessionCompleted:
		wsID = strVal(obj["client_reference_id"])
		sub.Plan = normalizePlan(strVal(metadata["plan"]))
		if sub.Plan == "free" {
			sub.Plan = "pro"
		}
		sub.Status = "active"
		sub.StripeCustomerID = strVal(obj["customer"])
		sub.StripeSubscriptionID = strVal(obj["subscription"])
	case stripe.EventTypeCustomerSubscriptionCreated, stripe.EventTypeCustomerSubscriptionUpdated, stripe.EventTypeCustomerSubscriptionDeleted:
		wsID = strVal(metadata["workspace_id"])
		sub.Plan = normalizePlan(strVal(metadata["plan"]))
		sub.Status = normalizeStatus(strVal(obj["status"]))
		sub.StripeCustomerID = strVal(obj["customer"])
		sub.StripeSubscriptionID = strVal(obj["id"])
		if event.Type == stripe.EventTypeCustomerSubscriptionDeleted {
			sub.Plan = "free"
			sub.Status = "canceled"
		}
		if ts, ok := obj["current_period_end"]; ok {
			if v, ok := toInt64(ts); ok && v > 0 {
				t := time.Unix(v, 0).UTC()
				sub.CurrentPeriodEnd = &t
			}
		}
	case stripe.EventTypeInvoicePaymentSucceeded, stripe.EventTypeInvoicePaymentFailed:
		wsID = strVal(metadata["workspace_id"])
		sub.StripeSubscriptionID = strVal(obj["subscription"])
		if event.Type == stripe.EventTypeInvoicePaymentSucceeded {
			sub.Status = "active"
		} else {
			sub.Status = "past_due"
		}
	default:
		return SubscriptionRecord{}, false
	}
	if wsID == "" {
		return SubscriptionRecord{}, false
	}
	sub.WorkspaceID = wsID
	applyPlanLimits(&sub)
	return sub, true
}

func applyPlanLimits(s *SubscriptionRecord) {
	switch s.Plan {
	case "team", "enterprise":
		s.RelayTier = "team"
		s.BandwidthLimitMB = 50 * 1024
		s.EntityLimit = -1
		s.RelayEnabled = true
	case "pro":
		s.RelayTier = "pro"
		s.BandwidthLimitMB = 10 * 1024
		s.EntityLimit = 10
		s.RelayEnabled = true
	default:
		s.Plan = "free"
		s.RelayTier = "free"
		s.BandwidthLimitMB = 1024
		s.EntityLimit = 3
		s.RelayEnabled = false
	}
}

func normalizePlan(p string) string {
	switch p {
	case "pro", "team", "enterprise":
		return p
	default:
		return "free"
	}
}

func normalizeStatus(s string) string {
	switch s {
	case "active", "past_due", "canceled", "trialing":
		return s
	default:
		return "active"
	}
}

func mapVal(m map[string]any, k string) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	if v, ok := m[k].(map[string]any); ok {
		return v
	}
	return map[string]any{}
}

func strVal(v any) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		return ""
	}
}

func toInt64(v any) (int64, bool) {
	switch t := v.(type) {
	case float64:
		return int64(t), true
	case int64:
		return t, true
	case int:
		return int64(t), true
	case string:
		x, err := strconv.ParseInt(t, 10, 64)
		return x, err == nil
	case json.Number:
		x, err := t.Int64()
		return x, err == nil
	default:
		return 0, false
	}
}
