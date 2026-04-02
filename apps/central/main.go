package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"
)

type Config struct {
	Port                  string
	Database              string
	Issuer                string
	JWTTTL                time.Duration
	KID                   string
	PrivKeyPEM            string
	GmailImpersonateEmail string
	GmailFromEmail        string
}

type App struct {
	cfg     Config
	db      *pgxpool.Pool
	privKey *rsa.PrivateKey
	pubKey  *rsa.PublicKey
}

type registerReq struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type claimGenerateReq struct {
	WorkspaceID string `json:"workspace_id"`
}

type claimKeyReq struct {
	ClaimKey string `json:"claim_key"`
}

type signupReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type forgotPasswordReq struct {
	Email string `json:"email"`
}

type resetPasswordReq struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type centralUser struct {
	ID           string
	Email        string
	DisplayName  string
	Plan         string
	PasswordHash string
}

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg := Config{
		Port:                  getenv("PORT", "8080"),
		Database:              os.Getenv("DATABASE_URL"),
		Issuer:                getenv("JWT_ISSUER", "https://agentunited.ai"),
		JWTTTL:                durationFromHours(getenv("JWT_TTL_HOURS", "24")),
		KID:                   getenv("JWT_KID", "2026-v1"),
		PrivKeyPEM:            os.Getenv("RSA_PRIVATE_KEY"),
		GmailImpersonateEmail: getenv("GMAIL_IMPERSONATE_EMAIL", "noreply@agentunited.ai"),
		GmailFromEmail:        getenv("GMAIL_FROM_EMAIL", "noreply@agentunited.ai"),
	}
	if cfg.Database == "" {
		log.Fatal().Msg("missing DATABASE_URL")
	}
	if cfg.PrivKeyPEM == "" {
		k, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to generate fallback rsa key")
		}
		cfg.PrivKeyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}))
		log.Warn().Msg("RSA_PRIVATE_KEY missing; generated ephemeral key for this process")
	}

	app, err := newApp(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("init failed")
	}
	defer app.db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", app.health)

	// Wrap mux with CORS middleware allowing agentunited.ai origins.
	handler := corsMiddleware(mux)
	mux.HandleFunc("GET /.well-known/jwks.json", app.jwks)
	mux.HandleFunc("POST /api/v1/users/register", app.register)
	mux.HandleFunc("POST /api/v1/users/login", app.login)
	mux.HandleFunc("POST /api/v1/users/forgot-password", app.forgotPassword)
	mux.HandleFunc("POST /api/v1/users/reset-password", app.resetPassword)
	mux.HandleFunc("POST /api/v1/auth/signup", app.signup)
	mux.HandleFunc("GET /api/v1/account", app.account)
	mux.HandleFunc("POST /api/v1/claim/generate", app.claimGenerate)
	mux.HandleFunc("POST /api/v1/claim/validate", app.claimValidate)
	mux.HandleFunc("POST /api/v1/claim/consume", app.claimConsume)
	mux.HandleFunc("GET /api/v1/workspaces", app.workspaces)

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Info().Str("addr", srv.Addr).Msg("central service listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	<-ctx.Done()
	stop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func newApp(cfg Config) (*App, error) {
	pool, err := pgxpool.New(context.Background(), cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}
	if err := ensureSchema(context.Background(), pool); err != nil {
		return nil, err
	}
	priv, err := parseRSAPrivateKey(cfg.PrivKeyPEM)
	if err != nil {
		return nil, err
	}
	return &App{cfg: cfg, db: pool, privKey: priv, pubKey: &priv.PublicKey}, nil
}

func (a *App) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.Email == "" || req.DisplayName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_input"})
		return
	}
	if err := validatePassword(req.Password); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "hash_failed"})
		return
	}

	var u centralUser
	err = a.db.QueryRow(r.Context(), `
		INSERT INTO central_users (email, display_name, password_hash)
		VALUES ($1,$2,$3)
		RETURNING id::text, email, display_name, plan
	`, req.Email, req.DisplayName, string(hash)).Scan(&u.ID, &u.Email, &u.DisplayName, &u.Plan)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "email_already_registered"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "register_failed"})
		return
	}
	tok, err := a.issueToken(u)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token_failed"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"user_id": u.ID, "email": u.Email, "display_name": u.DisplayName, "plan": u.Plan, "token": tok})
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_credentials"})
		return
	}

	var u centralUser
	err := a.db.QueryRow(r.Context(), `
		SELECT id::text, email, display_name, plan, password_hash
		FROM central_users WHERE email=$1
	`, req.Email).Scan(&u.ID, &u.Email, &u.DisplayName, &u.Plan, &u.PasswordHash)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_credentials"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_credentials"})
		return
	}
	tok, err := a.issueToken(u)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token_failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": tok, "user_id": u.ID, "plan": u.Plan})
}

func (a *App) claimGenerate(w http.ResponseWriter, r *http.Request) {
	userID, err := a.authUserID(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	var req claimGenerateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	req.WorkspaceID = strings.TrimSpace(req.WorkspaceID)
	if req.WorkspaceID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "workspace_id_required"})
		return
	}
	claimKey, err := generateClaimKey()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "generate_failed"})
		return
	}
	expiresAt := time.Now().UTC().Add(30 * 24 * time.Hour)
	_, err = a.db.Exec(r.Context(), `
		INSERT INTO claim_keys (user_id, key, expires_at, workspace_id)
		VALUES ($1, $2, $3, $4)
	`, userID, claimKey, expiresAt, req.WorkspaceID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "generate_failed"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"claim_key": claimKey, "expires_at": expiresAt.Format(time.RFC3339)})
}

func (a *App) claimValidate(w http.ResponseWriter, r *http.Request) {
	var req claimKeyReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	req.ClaimKey = strings.TrimSpace(req.ClaimKey)
	if req.ClaimKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "claim_key_required"})
		return
	}
	var workspaceID string
	err := a.db.QueryRow(r.Context(), `
		SELECT workspace_id
		FROM claim_keys
		WHERE key=$1 AND consumed_at IS NULL AND expires_at > NOW()
	`, req.ClaimKey).Scan(&workspaceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "validate_failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"valid": true, "workspace_id": workspaceID})
}

func (a *App) claimConsume(w http.ResponseWriter, r *http.Request) {
	_, err := a.authUserID(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	var req claimKeyReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	req.ClaimKey = strings.TrimSpace(req.ClaimKey)
	if req.ClaimKey == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "claim_key_required"})
		return
	}
	_, err = a.db.Exec(r.Context(), `
		UPDATE claim_keys
		SET consumed_at = COALESCE(consumed_at, NOW())
		WHERE key=$1
	`, req.ClaimKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "consume_failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (a *App) workspaces(w http.ResponseWriter, r *http.Request) {
	userID, err := a.authUserID(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	type workspace struct {
		WorkspaceID string    `json:"workspace_id"`
		CreatedAt   time.Time `json:"created_at"`
	}
	rows, err := a.db.Query(r.Context(), `
		SELECT workspace_id, joined_at
		FROM workspace_members
		WHERE central_user_id=$1
		ORDER BY joined_at DESC
	`, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "workspaces_failed"})
		return
	}
	defer rows.Close()

	out := make([]workspace, 0)
	for rows.Next() {
		var ws workspace
		if err := rows.Scan(&ws.WorkspaceID, &ws.CreatedAt); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "workspaces_failed"})
			return
		}
		out = append(out, ws)
	}
	if rows.Err() != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "workspaces_failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"workspaces": out})
}

func (a *App) authUserID(r *http.Request) (string, error) {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if auth == "" || !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return "", errors.New("missing bearer")
	}
	tokenStr := strings.TrimSpace(auth[len("Bearer "):])
	tok, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodRS256 {
			return nil, errors.New("invalid signing method")
		}
		return a.pubKey, nil
	})
	if err != nil || !tok.Valid {
		return "", errors.New("invalid token")
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}
	if iss, _ := claims["iss"].(string); iss != a.cfg.Issuer {
		return "", errors.New("invalid issuer")
	}
	sub, _ := claims["sub"].(string)
	if strings.TrimSpace(sub) == "" {
		return "", errors.New("missing sub")
	}
	return sub, nil
}

func (a *App) issueToken(u centralUser) (string, error) {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"iss":          a.cfg.Issuer,
		"aud":          []string{"agentunited:workspace"},
		"sub":          u.ID,
		"email":        u.Email,
		"plan":         u.Plan,
		"display_name": u.DisplayName,
		"iat":          now.Unix(),
		"exp":          now.Add(a.cfg.JWTTTL).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	t.Header["kid"] = a.cfg.KID
	return t.SignedString(a.privKey)
}

func (a *App) jwks(w http.ResponseWriter, _ *http.Request) {
	n := base64.RawURLEncoding.EncodeToString(a.pubKey.N.Bytes())
	e := big.NewInt(int64(a.pubKey.E)).Bytes()
	jwks := map[string]any{"keys": []map[string]string{{
		"kty": "RSA", "use": "sig", "kid": a.cfg.KID, "alg": "RS256", "n": n, "e": base64.RawURLEncoding.EncodeToString(e),
	}}}
	writeJSON(w, http.StatusOK, jwks)
}

func ensureSchema(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
CREATE TABLE IF NOT EXISTS central_users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT UNIQUE NOT NULL,
  display_name TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  plan TEXT NOT NULL DEFAULT 'free',
  stripe_customer_id TEXT,
  email_verified BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_central_users_email ON central_users(email);

CREATE TABLE IF NOT EXISTS claim_keys (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES central_users(id) ON DELETE CASCADE,
  key TEXT UNIQUE NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  consumed_at TIMESTAMPTZ,
  workspace_id TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_claim_keys_key ON claim_keys(key);
CREATE INDEX IF NOT EXISTS idx_claim_keys_user_id ON claim_keys(user_id);

CREATE TABLE IF NOT EXISTS workspace_members (
  workspace_id TEXT NOT NULL,
  central_user_id UUID NOT NULL REFERENCES central_users(id),
  relay_url TEXT NOT NULL,
  workspace_name TEXT,
  joined_at TIMESTAMPTZ DEFAULT NOW(),
  last_seen_at TIMESTAMPTZ,
  PRIMARY KEY (workspace_id, central_user_id)
);
CREATE INDEX IF NOT EXISTS idx_workspace_members_user ON workspace_members(central_user_id);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
  token      TEXT PRIMARY KEY,
  user_id    TEXT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '1 hour'
);
`)
	if err != nil {
		return fmt.Errorf("ensure schema: %w", err)
	}

	// M9: add new columns to central_users if not already present.
	migrations := []string{
		`ALTER TABLE central_users ADD COLUMN IF NOT EXISTS relay_token TEXT UNIQUE`,
		`ALTER TABLE central_users ADD COLUMN IF NOT EXISTS relay_token_created_at TIMESTAMPTZ`,
		`ALTER TABLE central_users ADD COLUMN IF NOT EXISTS entity_count_snapshot INT DEFAULT 0`,
		`ALTER TABLE central_users ADD COLUMN IF NOT EXISTS entity_limit INT DEFAULT 3`,
		`ALTER TABLE central_users ADD COLUMN IF NOT EXISTS github_oauth_id TEXT`,
	}
	for _, m := range migrations {
		if _, err := db.Exec(ctx, m); err != nil {
			return fmt.Errorf("migration %q: %w", m, err)
		}
	}
	return nil
}

func parseRSAPrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("invalid RSA_PRIVATE_KEY pem")
	}
	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return k, nil
	}
	kAny, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	k, ok := kAny.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("RSA_PRIVATE_KEY is not RSA")
	}
	return k, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func durationFromHours(h string) time.Duration {
	if h == "" {
		return 24 * time.Hour
	}
	var n int
	_, err := fmt.Sscanf(h, "%d", &n)
	if err != nil || n <= 0 {
		return 24 * time.Hour
	}
	return time.Duration(n) * time.Hour
}

func generateClaimKey() (string, error) {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "au_claim_" + base64.RawURLEncoding.EncodeToString(b), nil
}

// validatePassword enforces a minimum length of 8 characters.
// No complexity requirements — Apple Keychain and similar tools generate
// formats like "Gm3s-9kPx-vR4t-hWj2" that must not be rejected.
func validatePassword(pw string) error {
	if len(pw) < 8 {
		return errors.New("weak_password")
	}
	return nil
}

// generateResetToken generates a 32-byte hex reset token.
func generateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// sendResetEmail sends a password-reset email via SendGrid, or logs if key is absent.
func sendResetEmail(ctx context.Context, toEmail, token, impersonateEmail, fromEmail string) {
	resetLink := "agentunited://reset-password?token=" + token
	subject := "Reset your Agent United password"
	body := fmt.Sprintf(
		"Click to reset your password: %s\n\nThis link expires in 1 hour. If you didn't request a reset, ignore this email.",
		resetLink,
	)

	// Local/dev fallback: if SendGrid is configured, use it.
	if apiKey := os.Getenv("SENDGRID_API_KEY"); apiKey != "" {
		from := getenv("FROM_EMAIL", "noreply@agentunited.ai")
		payload := map[string]any{
			"personalizations": []map[string]any{{"to": []map[string]string{{"email": toEmail}}}},
			"from":             map[string]string{"email": from},
			"subject":          subject,
			"content":          []map[string]string{{"type": "text/plain", "value": body}},
		}
		b, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(b))
		if err != nil {
			log.Error().Err(err).Msg("build sendgrid request failed")
			return
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Error().Err(err).Str("to", toEmail).Msg("failed to send reset email via sendgrid")
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			log.Error().Int("status", resp.StatusCode).Str("to", toEmail).Msg("sendgrid rejected reset email")
		}
		return
	}

	// Default path: Gmail API via service account creds + domain-wide delegation.
	if err := sendViaGmailAPI(ctx, toEmail, impersonateEmail, fromEmail, subject, body); err != nil {
		log.Error().Err(err).Str("to", toEmail).Msg("gmail send failed; logging reset link fallback")
		log.Info().Str("to", toEmail).Str("reset_link", resetLink).Msg("password reset email fallback")
		return
	}
	log.Info().Str("to", toEmail).Str("from", impersonateEmail).Msg("password reset email sent via gmail api")
}

func sendViaGmailAPI(ctx context.Context, toEmail, impersonateEmail, fromEmail, subject, body string) error {
	if strings.TrimSpace(impersonateEmail) == "" {
		impersonateEmail = "noreply@agentunited.ai"
	}

	ts, err := gmailTokenSource(ctx, impersonateEmail)
	if err != nil {
		return err
	}
	svc, err := gmail.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return fmt.Errorf("create gmail service: %w", err)
	}

	if strings.TrimSpace(fromEmail) == "" {
		fromEmail = "noreply@agentunited.ai"
	}

	rfc2822 := strings.Join([]string{
		"From: " + fromEmail,
		"To: " + toEmail,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	raw := base64.RawURLEncoding.EncodeToString([]byte(rfc2822))
	_, err = svc.Users.Messages.Send("me", &gmail.Message{Raw: raw}).Do()
	if err != nil {
		return fmt.Errorf("gmail users.messages.send: %w", err)
	}
	return nil
}

func gmailTokenSource(ctx context.Context, subject string) (oauth2.TokenSource, error) {
	// If a service-account key is available locally, use JWTConfigFromJSON with Subject.
	if credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); credPath != "" {
		if b, err := os.ReadFile(credPath); err == nil {
			if cfg, err := google.JWTConfigFromJSON(b, gmail.GmailSendScope); err == nil {
				cfg.Subject = subject
				return cfg.TokenSource(ctx), nil
			}
		}
	}

	// Cloud Run ambient credentials — use impersonate package for DWD.
	ts, err := impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
		TargetPrincipal: "empire-deploy@agentunited-prod.iam.gserviceaccount.com",
		Scopes:          []string{gmail.GmailSendScope},
		Subject:         subject,
	})
	if err != nil {
		return nil, fmt.Errorf("impersonate token source: %w", err)
	}
	return ts, nil
}

// forgotPassword handles POST /api/v1/users/forgot-password
func (a *App) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	okMsg := map[string]string{"message": "If that email is registered, you'll receive a reset link shortly."}

	if req.Email == "" {
		writeJSON(w, http.StatusOK, okMsg)
		return
	}

	var userID, userEmail string
	err := a.db.QueryRow(r.Context(),
		`SELECT id::text, email FROM central_users WHERE email=$1`, req.Email,
	).Scan(&userID, &userEmail)
	if err != nil {
		writeJSON(w, http.StatusOK, okMsg) // don't leak existence
		return
	}

	token, err := generateResetToken()
	if err != nil {
		log.Error().Err(err).Msg("generate reset token failed")
		writeJSON(w, http.StatusOK, okMsg)
		return
	}

	_, err = a.db.Exec(r.Context(), `
		INSERT INTO password_reset_tokens (token, user_id)
		VALUES ($1, $2)
		ON CONFLICT (token) DO NOTHING
	`, token, userID)
	if err != nil {
		log.Error().Err(err).Msg("insert reset token failed")
		writeJSON(w, http.StatusOK, okMsg)
		return
	}

	sendResetEmail(r.Context(), userEmail, token, a.cfg.GmailImpersonateEmail, a.cfg.GmailFromEmail)

	writeJSON(w, http.StatusOK, okMsg)
}

// resetPassword handles POST /api/v1/users/reset-password
func (a *App) resetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
		return
	}
	req.Token = strings.TrimSpace(req.Token)

	if req.Token == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_or_expired_token"})
		return
	}
	if err := validatePassword(req.NewPassword); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "weak_password"})
		return
	}

	var userID string
	err := a.db.QueryRow(r.Context(), `
		SELECT user_id FROM password_reset_tokens
		WHERE token=$1 AND expires_at > NOW()
	`, req.Token).Scan(&userID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_or_expired_token"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "hash_failed"})
		return
	}

	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "tx_failed"})
		return
	}
	defer tx.Rollback(r.Context()) //nolint:errcheck

	if _, err := tx.Exec(r.Context(),
		`UPDATE central_users SET password_hash=$1, updated_at=NOW() WHERE id::text=$2`,
		string(hash), userID,
	); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "update_failed"})
		return
	}
	if _, err := tx.Exec(r.Context(),
		`DELETE FROM password_reset_tokens WHERE token=$1`, req.Token,
	); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "cleanup_failed"})
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "commit_failed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Password updated successfully."})
}

// base62Alphabet contains URL-safe alphanumeric characters for relay token generation.
const base62Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// generateRelayToken returns "rt_" + 32 cryptographically random base62 characters.
func generateRelayToken() (string, error) {
	const length = 32
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand.Read: %w", err)
	}
	out := make([]byte, length)
	for i, v := range b {
		out[i] = base62Alphabet[int(v)%len(base62Alphabet)]
	}
	return "rt_" + string(out), nil
}

// entityLimitForPlan returns the entity limit for a given plan.
func entityLimitForPlan(plan string) int {
	switch strings.ToLower(strings.TrimSpace(plan)) {
	case "pro":
		return 15
	case "team", "enterprise":
		return 50
	default:
		return 3
	}
}

// signup handles POST /api/v1/auth/signup
// Creates a new central account with relay_token; returns JWT + account data.
func (a *App) signup(w http.ResponseWriter, r *http.Request) {
	var req signupReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid request body.", "code": "invalid_json"})
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "Email is required.", "code": "validation_error"})
		return
	}
	if err := validatePassword(req.Password); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "Password must be at least 8 characters.", "code": "validation_error"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "Internal error.", "code": "hash_failed"})
		return
	}

	relayToken, err := generateRelayToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "Internal error.", "code": "token_gen_failed"})
		return
	}

	plan := "free"
	entityLimit := entityLimitForPlan(plan)

	var u centralUser
	err = a.db.QueryRow(r.Context(), `
		INSERT INTO central_users (email, display_name, password_hash, relay_token, relay_token_created_at, entity_limit)
		VALUES ($1, $1, $2, $3, NOW(), $4)
		RETURNING id::text, email, display_name, plan
	`, req.Email, string(hash), relayToken, entityLimit).Scan(&u.ID, &u.Email, &u.DisplayName, &u.Plan)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			writeJSON(w, http.StatusConflict, map[string]any{"error": "An account with this email already exists.", "code": "email_taken"})
			return
		}
		log.Error().Err(err).Str("email", req.Email).Msg("signup insert failed")
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "Internal error.", "code": "register_failed"})
		return
	}

	tok, err := a.issueToken(u)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "Internal error.", "code": "token_failed"})
		return
	}

	// Set JWT in httpOnly cookie for browser clients (SameSite=Lax, Secure).
	http.SetCookie(w, &http.Cookie{
		Name:     "au_session",
		Value:    tok,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(a.cfg.JWTTTL.Seconds()),
	})

	writeJSON(w, http.StatusCreated, map[string]any{
		"user_id":      u.ID,
		"email":        u.Email,
		"plan":         u.Plan,
		"relay_token":  relayToken,
		"entity_limit": entityLimit,
		"token":        tok,
	})
}

// account handles GET /api/v1/account (JWT required).
// Returns dashboard state: relay_token, plan, entity usage, email.
func (a *App) account(w http.ResponseWriter, r *http.Request) {
	userID, err := a.authUserID(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "Unauthorized.", "code": "unauthorized"})
		return
	}

	var (
		email              string
		displayName        string
		plan               string
		relayToken         *string
		entityCountSnapshot int
		entityLimit        int
	)
	err = a.db.QueryRow(r.Context(), `
		SELECT email, display_name, plan,
		       relay_token,
		       COALESCE(entity_count_snapshot, 0),
		       COALESCE(entity_limit, 3)
		FROM central_users
		WHERE id::text = $1
	`, userID).Scan(&email, &displayName, &plan, &relayToken, &entityCountSnapshot, &entityLimit)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("account lookup failed")
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "Internal error.", "code": "lookup_failed"})
		return
	}

	resp := map[string]any{
		"user_id":            userID,
		"email":              email,
		"display_name":       displayName,
		"plan":               plan,
		"relay_token":        relayToken,
		"entity_count":       entityCountSnapshot,
		"entity_limit":       entityLimit,
		"stripe_portal_url":  nil,
	}
	writeJSON(w, http.StatusOK, resp)
}

// corsMiddleware adds CORS headers allowing agentunited.ai origins.
func corsMiddleware(next http.Handler) http.Handler {
	allowed := map[string]bool{
		"https://agentunited.ai":     true,
		"https://app.agentunited.ai": true,
		"http://localhost:3000":      true,
		"http://localhost:3001":      true,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Origin")
			w.Header().Set("Access-Control-Max-Age", "300")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
