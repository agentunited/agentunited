package relay

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTestRSAKey creates a fresh RSA key pair for testing.
func generateTestRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return priv
}

// makeJWKSServer starts an httptest server serving a JWKS with the given key.
func makeJWKSServer(t *testing.T, kid string, pub *rsa.PublicKey) *httptest.Server {
	t.Helper()
	eBytes := big.NewInt(int64(pub.E)).Bytes()
	nBytes := pub.N.Bytes()
	body := map[string]interface{}{
		"keys": []map[string]interface{}{
			{
				"kty": "RSA",
				"kid": kid,
				"alg": "RS256",
				"n":   base64.RawURLEncoding.EncodeToString(nBytes),
				"e":   base64.RawURLEncoding.EncodeToString(eBytes),
			},
		},
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}))
}

// signJWT creates a signed RS256 JWT for testing.
func signJWT(t *testing.T, priv *rsa.PrivateKey, kid, issuer, sub, email string, exp time.Time) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, &CentralClaims{
		Sub:   sub,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})
	tok.Header["kid"] = kid
	signed, err := tok.SignedString(priv)
	require.NoError(t, err)
	return signed
}

func TestJWKSVerifier_ValidJWT(t *testing.T) {
	priv := generateTestRSAKey(t)
	srv := makeJWKSServer(t, "test-kid", &priv.PublicKey)
	defer srv.Close()

	v := NewJWKSVerifier(srv.URL, "https://agentunited.ai")
	err := v.FetchOnce(context.Background())
	require.NoError(t, err)
	assert.True(t, v.Enabled)

	tok := signJWT(t, priv, "test-kid", "https://agentunited.ai", "user-1", "user@example.com", time.Now().Add(time.Hour))
	claims, ok := v.VerifyJWT(tok)
	require.True(t, ok)
	assert.Equal(t, "user-1", claims.Sub)
	assert.Equal(t, "user@example.com", claims.Email)
}

func TestJWKSVerifier_ExpiredJWT(t *testing.T) {
	priv := generateTestRSAKey(t)
	srv := makeJWKSServer(t, "test-kid", &priv.PublicKey)
	defer srv.Close()

	v := NewJWKSVerifier(srv.URL, "https://agentunited.ai")
	require.NoError(t, v.FetchOnce(context.Background()))

	tok := signJWT(t, priv, "test-kid", "https://agentunited.ai", "user-1", "user@example.com", time.Now().Add(-time.Minute))
	_, ok := v.VerifyJWT(tok)
	assert.False(t, ok)
}

func TestJWKSVerifier_WrongIssuer(t *testing.T) {
	priv := generateTestRSAKey(t)
	srv := makeJWKSServer(t, "test-kid", &priv.PublicKey)
	defer srv.Close()

	v := NewJWKSVerifier(srv.URL, "https://agentunited.ai")
	require.NoError(t, v.FetchOnce(context.Background()))

	tok := signJWT(t, priv, "test-kid", "https://evil.example.com", "user-1", "user@example.com", time.Now().Add(time.Hour))
	_, ok := v.VerifyJWT(tok)
	assert.False(t, ok)
}

func TestJWKSVerifier_UnknownKID(t *testing.T) {
	priv := generateTestRSAKey(t)
	srv := makeJWKSServer(t, "test-kid", &priv.PublicKey)
	defer srv.Close()

	v := NewJWKSVerifier(srv.URL, "https://agentunited.ai")
	require.NoError(t, v.FetchOnce(context.Background()))

	tok := signJWT(t, priv, "unknown-kid", "https://agentunited.ai", "user-1", "user@example.com", time.Now().Add(time.Hour))
	_, ok := v.VerifyJWT(tok)
	assert.False(t, ok)
}

func TestJWKSVerifier_DegradedMode_UnreachableJWKS(t *testing.T) {
	v := NewJWKSVerifier("http://127.0.0.1:19999/jwks.json", "https://agentunited.ai")
	err := v.FetchOnce(context.Background())
	assert.Error(t, err)
	assert.False(t, v.Enabled)

	// In degraded mode VerifyJWT should return false (caller allows through)
	_, ok := v.VerifyJWT("any.token.here")
	assert.False(t, ok)
}

func TestExtractBearerToken(t *testing.T) {
	tok, ok := ExtractBearerToken("Bearer mytoken123")
	assert.True(t, ok)
	assert.Equal(t, "mytoken123", tok)

	_, ok = ExtractBearerToken("Basic user:pass")
	assert.False(t, ok)

	_, ok = ExtractBearerToken("")
	assert.False(t, ok)

	_, ok = ExtractBearerToken("Bearer ")
	assert.False(t, ok)
}
