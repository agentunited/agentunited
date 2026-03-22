package service

import (
	"strings"
	"testing"
)

func TestUpgradeEmailKey(t *testing.T) {
	got := upgradeEmailKey("ws_123")
	if got != "email:upgrade-confirm:ws_123" {
		t.Fatalf("unexpected key: %s", got)
	}
}

func TestBuildUpgradeEmail(t *testing.T) {
	subject, textBody, htmlBody := buildUpgradeEmail("pro", "https://app.agentunited.ai")
	if !strings.Contains(subject, "PRO") {
		t.Fatalf("subject should mention PRO: %s", subject)
	}
	if !strings.Contains(textBody, "docs.agentunited.ai/docs/agent-connect") {
		t.Fatalf("text body missing docs link")
	}
	if !strings.Contains(htmlBody, "https://app.agentunited.ai") {
		t.Fatalf("html body missing workspace link")
	}
}
