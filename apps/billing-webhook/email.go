package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type EmailSender interface {
	SendUpgradeConfirmation(ctx context.Context, toEmail, plan string) error
}

type SendgridEmailSender struct {
	apiKey string
	from   string
}

func NewSendgridEmailSender(apiKey, from string) *SendgridEmailSender {
	return &SendgridEmailSender{apiKey: apiKey, from: from}
}

func (s *SendgridEmailSender) SendUpgradeConfirmation(ctx context.Context, toEmail, plan string) error {
	if s.apiKey == "" || toEmail == "" {
		return nil
	}
	subj := fmt.Sprintf("You're on %s — Agent United", stringsUpper(plan))
	body := fmt.Sprintf("Your workspace has been upgraded to %s. You now have relay access and expanded limits.", stringsUpper(plan))
	payload := map[string]any{
		"personalizations": []any{map[string]any{"to": []any{map[string]string{"email": toEmail}}}},
		"from":             map[string]string{"email": s.from},
		"subject":          subj,
		"content":          []any{map[string]string{"type": "text/plain", "value": body}},
	}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("sendgrid status %d", resp.StatusCode)
	}
	return nil
}

func stringsUpper(s string) string {
	if len(s) == 0 {
		return s
	}
	b := []byte(s)
	for i := range b {
		if b[i] >= 'a' && b[i] <= 'z' {
			b[i] -= 32
		}
	}
	return string(b)
}
