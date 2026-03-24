package relay

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

// JWKSVerifier fetches RS256 public keys from a JWKS endpoint and verifies JWTs.
// It degrades gracefully if JWKS is unreachable: the Enabled flag is false and
// VerifyJWT always returns (nil, false) — callers should allow the connection
// through and log a warning rather than hard-reject.
type JWKSVerifier struct {
	url        string
	httpClient *http.Client
	issuer     string

	mu          sync.RWMutex
	keys        map[string]*rsa.PublicKey
	fetchedAt   time.Time
	Enabled     bool // false when JWKS could not be fetched
}

// NewJWKSVerifier creates a verifier for the given JWKS URL and expected issuer.
func NewJWKSVerifier(jwksURL, issuer string) *JWKSVerifier {
	return &JWKSVerifier{
		url:        jwksURL,
		issuer:     issuer,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		keys:       map[string]*rsa.PublicKey{},
	}
}

type jwkSet struct {
	Keys []struct {
		Kty string `json:"kty"`
		Kid string `json:"kid"`
		Alg string `json:"alg"`
		N   string `json:"n"`
		E   string `json:"e"`
	} `json:"keys"`
}

// FetchOnce loads JWKS synchronously. Call at startup.
// Returns nil on success; on error logs a warning and sets Enabled=false.
func (v *JWKSVerifier) FetchOnce(ctx context.Context) error {
	keys, err := v.fetch(ctx)
	if err != nil {
		log.Warn().Err(err).Str("url", v.url).Msg("JWKS fetch failed at startup — relay will allow unauthenticated connections")
		v.mu.Lock()
		v.Enabled = false
		v.mu.Unlock()
		return err
	}
	v.mu.Lock()
	v.keys = keys
	v.fetchedAt = time.Now()
	v.Enabled = true
	v.mu.Unlock()
	log.Info().Str("url", v.url).Int("keys", len(keys)).Msg("JWKS loaded")
	return nil
}

// StartRefreshLoop refreshes JWKS every hour in the background.
func (v *JWKSVerifier) StartRefreshLoop(ctx context.Context) {
	go func() {
		t := time.NewTicker(time.Hour)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				keys, err := v.fetch(ctx)
				if err != nil {
					log.Warn().Err(err).Msg("JWKS refresh failed — keeping cached keys")
					continue
				}
				v.mu.Lock()
				v.keys = keys
				v.fetchedAt = time.Now()
				v.Enabled = true
				v.mu.Unlock()
				log.Debug().Int("keys", len(keys)).Msg("JWKS refreshed")
			}
		}
	}()
}

// VerifyJWT verifies a raw JWT string. Returns the parsed claims and true on success.
// Returns (nil, false) if JWKS is not enabled (degrade gracefully) or if verification fails.
func (v *JWKSVerifier) VerifyJWT(tokenStr string) (*CentralClaims, bool) {
	v.mu.RLock()
	enabled := v.Enabled
	v.mu.RUnlock()
	if !enabled {
		return nil, false
	}

	claims := &CentralClaims{}
	parsed, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected alg: %v", t.Header["alg"])
		}
		kid, _ := t.Header["kid"].(string)
		v.mu.RLock()
		key := v.keys[kid]
		v.mu.RUnlock()
		if key == nil {
			return nil, fmt.Errorf("kid %q not found in JWKS", kid)
		}
		return key, nil
	})
	if err != nil || !parsed.Valid {
		return nil, false
	}
	if v.issuer != "" && claims.Issuer != v.issuer {
		return nil, false
	}
	return claims, true
}

// CentralClaims represents the JWT payload issued by the central service.
type CentralClaims struct {
	Sub         string `json:"sub"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Plan        string `json:"plan"`
	jwt.RegisteredClaims
}

// ExtractBearerToken extracts the token from "Authorization: Bearer <token>".
func ExtractBearerToken(authHeader string) (string, bool) {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", false
	}
	tok := strings.TrimPrefix(authHeader, "Bearer ")
	tok = strings.TrimSpace(tok)
	if tok == "" {
		return "", false
	}
	return tok, true
}

func (v *JWKSVerifier) fetch(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned %d", resp.StatusCode)
	}

	var doc jwkSet
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode JWKS: %w", err)
	}

	out := map[string]*rsa.PublicKey{}
	for _, k := range doc.Keys {
		if k.Kty != "RSA" || k.Kid == "" || k.N == "" || k.E == "" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		e := 0
		for _, b := range eBytes {
			e = e<<8 + int(b)
		}
		if e == 0 {
			continue
		}
		out[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("JWKS response contained no valid RSA keys")
	}
	return out, nil
}
