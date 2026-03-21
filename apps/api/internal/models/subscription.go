package models

import "time"

type Subscription struct {
	ID                    string     `json:"id"`
	WorkspaceID           string     `json:"workspace_id"`
	StripeCustomerID      string     `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID  string     `json:"stripe_subscription_id,omitempty"`
	Plan                  string     `json:"plan"`
	Status                string     `json:"status"`
	CurrentPeriodEnd      *time.Time `json:"current_period_end,omitempty"`
	RelayTier             string     `json:"relay_tier"`
	RelayBandwidthUsedMB  int        `json:"relay_bandwidth_used_mb"`
	RelayBandwidthLimitMB int        `json:"relay_bandwidth_limit_mb"`
	RelayConnectionsMax   int        `json:"relay_connections_max"`
	RelayCustomSubdomain  bool       `json:"relay_custom_subdomain"`
	RelaySubdomain        string     `json:"relay_subdomain,omitempty"`
	RelayToken            string     `json:"relay_token,omitempty"`
	RelayExpiresAt        *time.Time `json:"relay_expires_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}
