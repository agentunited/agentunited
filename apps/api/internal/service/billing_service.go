package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/repository"
	"github.com/agentunited/backend/pkg/billing"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

var ErrInvalidWebhookSignature = errors.New("invalid webhook signature")

type BillingService interface {
	GetCheckoutURL(ctx context.Context, workspaceID, email, name, plan, successURL, cancelURL string) (string, error)
	GetPortalURL(ctx context.Context, workspaceID, returnURL string) (string, error)
	GetStatus(ctx context.Context, workspaceID string) (*models.Subscription, int64, error)
	HandleWebhook(ctx context.Context, body []byte, signature string) error
}

type billingService struct {
	repo          repository.SubscriptionRepository
	userRepo      repository.UserRepository
	provider      billing.Service
	redisClient   *redis.Client
	webhookSecret string
	priceIDPro    string
	priceIDTeam   string
}

func NewBillingService(repo repository.SubscriptionRepository, userRepo repository.UserRepository, provider billing.Service, redisClient *redis.Client, webhookSecret, priceIDPro, priceIDTeam string) BillingService {
	return &billingService{repo: repo, userRepo: userRepo, provider: provider, redisClient: redisClient, webhookSecret: webhookSecret, priceIDPro: priceIDPro, priceIDTeam: priceIDTeam}
}

func (s *billingService) GetCheckoutURL(ctx context.Context, workspaceID, email, name, plan, successURL, cancelURL string) (string, error) {
	customerID := ""
	if existing, err := s.repo.GetByWorkspace(ctx, workspaceID); err == nil {
		customerID = existing.StripeCustomerID
	}
	var err error
	if plan == "" {
		plan = "pro"
	}
	if customerID == "" {
		customerID, err = s.provider.CreateCustomer(ctx, email, name)
		if err != nil {
			return "", err
		}
		if err := s.repo.UpsertByWorkspace(ctx, &models.Subscription{WorkspaceID: workspaceID, StripeCustomerID: customerID, Plan: "free", Status: "active"}); err != nil {
			return "", err
		}
	}
	priceID := s.priceIDForPlan(plan)
	return s.provider.CreateCheckoutSession(ctx, customerID, priceID, successURL, cancelURL)
}

func (s *billingService) GetStatus(ctx context.Context, workspaceID string) (*models.Subscription, int64, error) {
	sub, err := s.repo.GetByWorkspace(ctx, workspaceID)
	if err != nil {
		// Default free status if not yet present.
		sub = &models.Subscription{WorkspaceID: workspaceID, Plan: "free", Status: "free", RelayTier: "free", RelayBandwidthLimitMB: 1024, RelayConnectionsMax: 3}
	}
	users, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	return sub, users, nil
}

func (s *billingService) GetPortalURL(ctx context.Context, workspaceID, returnURL string) (string, error) {
	existing, err := s.repo.GetByWorkspace(ctx, workspaceID)
	if err != nil || existing.StripeCustomerID == "" {
		return "", billing.ErrNotConfigured
	}
	return s.provider.GetPortalURL(ctx, existing.StripeCustomerID, returnURL)
}

func (s *billingService) HandleWebhook(ctx context.Context, body []byte, signature string) error {
	if s.webhookSecret == "" {
		log.Warn().Msg("billing webhook not configured")
		return nil // intentionally noop when not configured
	}
	if !verifyStripeSignature(body, signature, s.webhookSecret) {
		return ErrInvalidWebhookSignature
	}

	var evt struct {
		Type string `json:"type"`
		Data struct {
			Object map[string]interface{} `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("decode webhook: %w", err)
	}

	switch evt.Type {
	case "checkout.session.completed":
		wsID := strVal(evt.Data.Object["client_reference_id"])
		subID := strVal(evt.Data.Object["subscription"])
		custID := strVal(evt.Data.Object["customer"])
		plan := strVal(mapVal(evt.Data.Object, "metadata")["plan"])
		if plan == "" {
			plan = "pro"
		}
		status := "active"
		normalizedPlan := normalizePlan(plan)
		sub := &models.Subscription{WorkspaceID: wsID, StripeCustomerID: custID, StripeSubscriptionID: subID, Plan: normalizedPlan, Status: status}
		applyRelayTierDefaults(sub, normalizedPlan)
		if err := s.repo.UpsertByWorkspace(ctx, sub); err != nil {
			return fmt.Errorf("upsert subscription: %w", err)
		}
		s.updateRelayCache(ctx, wsID)
		s.sendUpgradeConfirmationEmail(ctx, wsID, normalizedPlan)
	case "customer.subscription.created", "customer.subscription.updated", "customer.subscription.deleted":
		obj := evt.Data.Object
		wsID := strVal(mapVal(obj, "metadata")["workspace_id"])
		subID := strVal(obj["id"])
		custID := strVal(obj["customer"])
		status := normalizeStatus(strVal(obj["status"]))
		plan := normalizePlan(strVal(mapVal(obj, "metadata")["plan"]))
		if evt.Type == "customer.subscription.deleted" {
			plan = "free"
			status = "canceled"
		}
		if plan == "free" && evt.Type != "customer.subscription.deleted" {
			plan = "pro"
		}
		var cpe *time.Time
		if unix := int64Val(obj["current_period_end"]); unix > 0 {
			t := time.Unix(unix, 0).UTC()
			cpe = &t
		}
		sub := &models.Subscription{WorkspaceID: wsID, StripeCustomerID: custID, StripeSubscriptionID: subID, Plan: plan, Status: status, CurrentPeriodEnd: cpe}
		applyRelayTierDefaults(sub, plan)
		if err := s.repo.UpsertByWorkspace(ctx, sub); err != nil {
			return fmt.Errorf("upsert subscription: %w", err)
		}
		s.updateRelayCache(ctx, wsID)
	case "invoice.payment_succeeded", "invoice.payment_failed":
		obj := evt.Data.Object
		subID := strVal(obj["subscription"])
		wsID := strVal(mapVal(obj, "metadata")["workspace_id"])
		if subID != "" {
			status := "active"
			if evt.Type == "invoice.payment_failed" {
				status = "past_due"
			}
			plan := normalizePlan(strVal(mapVal(obj, "metadata")["plan"]))
			if plan == "free" {
				plan = "pro"
			}
			var cpe *time.Time
			if unix := int64Val(obj["period_end"]); unix > 0 {
				t := time.Unix(unix, 0).UTC()
				cpe = &t
			}
			sub := &models.Subscription{WorkspaceID: wsID, StripeSubscriptionID: subID, Plan: plan, Status: status, CurrentPeriodEnd: cpe}
			applyRelayTierDefaults(sub, plan)
			if err := s.repo.UpsertByWorkspace(ctx, sub); err != nil {
				return fmt.Errorf("upsert subscription: %w", err)
			}
			s.updateRelayCache(ctx, wsID)
		}
	}
	return nil
}

func verifyStripeSignature(payload []byte, signatureHeader, secret string) bool {
	if signatureHeader == "" || secret == "" {
		return false
	}
	parts := strings.Split(signatureHeader, ",")
	var ts, sig string
	for _, p := range parts {
		kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
		if len(kv) != 2 {
			continue
		}
		if kv[0] == "t" {
			ts = kv[1]
		}
		if kv[0] == "v1" {
			sig = kv[1]
		}
	}
	if ts == "" || sig == "" {
		return false
	}
	signedPayload := ts + "." + string(payload)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(signedPayload))
	expected := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sig))
}

func strVal(v interface{}) string {
	s, _ := v.(string)
	return s
}
func int64Val(v interface{}) int64 {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case string:
		n, _ := strconv.ParseInt(x, 10, 64)
		return n
	default:
		return 0
	}
}
func mapVal(m map[string]interface{}, k string) map[string]interface{} {
	if m == nil {
		return map[string]interface{}{}
	}
	if v, ok := m[k].(map[string]interface{}); ok {
		return v
	}
	return map[string]interface{}{}
}
func normalizePlan(p string) string {
	s := strings.ToLower(strings.TrimSpace(p))
	switch s {
	case "pro", "team", "enterprise":
		return s
	default:
		return "free"
	}
}
func normalizeStatus(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "active", "past_due", "canceled", "trialing":
		return s
	default:
		return "active"
	}
}

func (s *billingService) updateRelayCache(ctx context.Context, workspaceID string) {
	if s.redisClient == nil || workspaceID == "" {
		return
	}
	token, err := s.redisClient.Get(ctx, "relay:workspace:"+workspaceID).Result()
	if err != nil || token == "" {
		return
	}
	sub, err := s.repo.GetByWorkspace(ctx, workspaceID)
	if err != nil || sub == nil {
		return
	}
	subdomain := sub.RelaySubdomain
	if subdomain == "" {
		h := sha1.Sum([]byte(token))
		subdomain = "w" + hex.EncodeToString(h[:])[:10]
	}
	payload := map[string]interface{}{
		"workspace_id":       workspaceID,
		"plan":               sub.Plan,
		"relay_tier":         sub.RelayTier,
		"subdomain":          subdomain,
		"bandwidth_limit_mb": sub.RelayBandwidthLimitMB,
		"connections_max":    sub.RelayConnectionsMax,
		"expires_at":         sub.RelayExpiresAt,
	}
	if b, err := json.Marshal(payload); err == nil {
		_ = s.redisClient.Set(ctx, "relay:token:"+token, string(b), 0).Err()
	}
}

func (s *billingService) sendUpgradeConfirmationEmail(ctx context.Context, workspaceID, plan string) {
	if workspaceID == "" {
		return
	}
	// Idempotency on webhook replay: only send once per 7 days.
	if s.redisClient != nil {
		key := upgradeEmailKey(workspaceID)
		ok, err := s.redisClient.SetNX(ctx, key, normalizePlan(plan), 7*24*time.Hour).Result()
		if err == nil && !ok {
			return
		}
	}

	owner, err := s.userRepo.GetByID(ctx, workspaceID)
	if err != nil || owner == nil || owner.Email == "" {
		return
	}

	workspaceURL := os.Getenv("FRONTEND_URL")
	if workspaceURL == "" {
		workspaceURL = "https://app.agentunited.ai"
	}
	subject, textBody, htmlBody := buildUpgradeEmail(plan, workspaceURL)
	nPlan := strings.ToUpper(normalizePlan(plan))

	sendgridKey := os.Getenv("SENDGRID_API_KEY")
	from := os.Getenv("FROM_EMAIL")
	if from == "" {
		from = "hello@agentunited.ai"
	}

	if sendgridKey == "" {
		log.Info().Str("to", owner.Email).Str("plan", nPlan).Msg("sending upgrade email (log fallback; provider not configured)")
		return
	}

	payload := map[string]interface{}{
		"personalizations": []map[string]interface{}{{
			"to": []map[string]string{{"email": owner.Email}},
		}},
		"from":    map[string]string{"email": from},
		"subject": subject,
		"content": []map[string]string{{"type": "text/plain", "value": textBody}, {"type": "text/html", "value": htmlBody}},
	}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+sendgridKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error().Err(err).Str("to", owner.Email).Msg("failed to send upgrade email")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Error().Int("status", resp.StatusCode).Str("to", owner.Email).Msg("sendgrid rejected upgrade email")
	}
}

func upgradeEmailKey(workspaceID string) string {
	return "email:upgrade-confirm:" + workspaceID
}

func buildUpgradeEmail(plan, workspaceURL string) (subject, textBody, htmlBody string) {
	nPlan := strings.ToUpper(normalizePlan(plan))
	subject = fmt.Sprintf("You're on %s — Agent United", nPlan)
	docsURL := "https://docs.agentunited.ai/docs/agent-connect"
	textBody = fmt.Sprintf("Hi there,\n\nYour Agent United workspace is now on the %s plan.\n\nNext step: connect your agent\n%s\n\nWorkspace: %s\n\n— The Agent United team\n", nPlan, docsURL, workspaceURL)
	htmlBody = fmt.Sprintf("<p>Hi there,</p><p>Your Agent United workspace is now on the <strong>%s</strong> plan.</p><p>Next step: <a href=\"%s\">connect your agent</a></p><p>Workspace: <a href=\"%s\">%s</a></p><p>— The Agent United team</p>", nPlan, docsURL, workspaceURL, workspaceURL)
	return
}

func (s *billingService) priceIDForPlan(plan string) string {
	switch strings.ToLower(strings.TrimSpace(plan)) {
	case "team":
		if s.priceIDTeam != "" {
			return s.priceIDTeam
		}
		return s.priceIDPro
	default:
		return s.priceIDPro
	}
}

func applyRelayTierDefaults(sub *models.Subscription, plan string) {
	now := time.Now().UTC()
	switch normalizePlan(plan) {
	case "pro", "team", "enterprise":
		sub.RelayTier = "pro"
		sub.RelayBandwidthLimitMB = 51200
		sub.RelayConnectionsMax = 20
		sub.RelayCustomSubdomain = true
		sub.RelayExpiresAt = nil
	default:
		expires := now.Add(30 * 24 * time.Hour)
		sub.RelayTier = "free"
		sub.RelayBandwidthLimitMB = 1024
		sub.RelayConnectionsMax = 3
		sub.RelayCustomSubdomain = false
		sub.RelayExpiresAt = &expires
	}
}
