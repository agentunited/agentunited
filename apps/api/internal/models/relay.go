package models

import "time"

type RelayStatusResponse struct {
	RelayURL         string     `json:"relay_url"`
	RelayTier        string     `json:"relay_tier"`
	BandwidthUsedMB  int        `json:"bandwidth_used_mb"`
	BandwidthLimitMB int        `json:"bandwidth_limit_mb"`
	ConnectionsMax   int        `json:"connections_max"`
	RelaySubdomain   string     `json:"relay_subdomain"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
}
