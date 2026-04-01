// Package email provides a thin helper for sending transactional emails.
// Primary path: Gmail API via service-account DWD (same pattern as the central service).
// Fallback: SendGrid if SENDGRID_API_KEY is set.
// Last resort: log the message body so resets are never silently swallowed.
package email

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"
)

const (
	defaultImpersonateEmail = "noreply@agentunited.ai"
	defaultFromEmail        = "noreply@agentunited.ai"
	// Cloud Run service account used for DWD impersonation.
	deployServiceAccount = "empire-deploy@agentunited-prod.iam.gserviceaccount.com"
)

// Send dispatches a plain-text email.
// It tries Gmail DWD first, then SendGrid, then logs as last resort.
func Send(ctx context.Context, toEmail, subject, body string) {
	impersonateEmail := getenv("GMAIL_IMPERSONATE_EMAIL", defaultImpersonateEmail)
	fromEmail := getenv("GMAIL_FROM_EMAIL", defaultFromEmail)

	// Try Gmail DWD first.
	if err := sendViaGmailAPI(ctx, toEmail, impersonateEmail, fromEmail, subject, body); err == nil {
		log.Info().Str("to", toEmail).Str("from", impersonateEmail).Msg("email sent via gmail api")
		return
	}

	// Fallback: SendGrid.
	if apiKey := os.Getenv("SENDGRID_API_KEY"); apiKey != "" {
		from := getenv("FROM_EMAIL", defaultFromEmail)
		if sendViaSendGrid(ctx, apiKey, from, toEmail, subject, body) {
			return
		}
	}

	// Last resort: log so ops can manually handle.
	log.Warn().Str("to", toEmail).Str("subject", subject).Msg("all email transports failed; logging body as fallback")
	log.Info().Str("to", toEmail).Str("body", body).Msg("email body fallback")
}

// sendViaGmailAPI sends via Google Gmail API using DWD.
func sendViaGmailAPI(ctx context.Context, toEmail, impersonateEmail, fromEmail, subject, body string) error {
	if strings.TrimSpace(impersonateEmail) == "" {
		impersonateEmail = defaultImpersonateEmail
	}
	ts, err := gmailTokenSource(ctx, impersonateEmail)
	if err != nil {
		return fmt.Errorf("gmail token source: %w", err)
	}
	svc, err := gmail.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return fmt.Errorf("create gmail service: %w", err)
	}
	if strings.TrimSpace(fromEmail) == "" {
		fromEmail = defaultFromEmail
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

// gmailTokenSource returns an OAuth2 token source for Gmail DWD.
func gmailTokenSource(ctx context.Context, subject string) (oauth2.TokenSource, error) {
	// If a local service-account key file is available, use it directly.
	if credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); credPath != "" {
		if b, err := os.ReadFile(credPath); err == nil {
			if cfg, err := google.JWTConfigFromJSON(b, gmail.GmailSendScope); err == nil {
				cfg.Subject = subject
				return cfg.TokenSource(ctx), nil
			}
		}
	}
	// Cloud Run: use impersonate package for ambient credentials + DWD.
	ts, err := impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
		TargetPrincipal: deployServiceAccount,
		Scopes:          []string{gmail.GmailSendScope},
		Subject:         subject,
	})
	if err != nil {
		return nil, fmt.Errorf("impersonate token source: %w", err)
	}
	return ts, nil
}

// sendViaSendGrid dispatches via SendGrid v3 API. Returns true on success.
func sendViaSendGrid(ctx context.Context, apiKey, from, toEmail, subject, body string) bool {
	payload := map[string]interface{}{
		"personalizations": []map[string]interface{}{{"to": []map[string]string{{"email": toEmail}}}},
		"from":             map[string]string{"email": from},
		"subject":          subject,
		"content":          []map[string]string{{"type": "text/plain", "value": body}},
	}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(b))
	if err != nil {
		log.Error().Err(err).Msg("build sendgrid request failed")
		return false
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error().Err(err).Str("to", toEmail).Msg("sendgrid send failed")
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Error().Int("status", resp.StatusCode).Str("to", toEmail).Msg("sendgrid rejected email")
		return false
	}
	log.Info().Str("to", toEmail).Msg("email sent via sendgrid")
	return true
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
