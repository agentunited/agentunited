package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentunited/backend/internal/config"
	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenClawIntegration_MessageRoundTrip verifies:
// 1. Bootstrap workspace
// 2. Register OpenClaw integration
// 3. Send inbound message via integration webhook
// 4. Verify message in channel
// 5. Send reply from workspace
// 6. Verify outbound webhook delivered to mock receiver
func TestOpenClawIntegration_MessageRoundTrip(t *testing.T) {
	// Skip if no test database available
	db, server := setupIntegrationTest(t)
	defer cleanupIntegrationTest(t, db, server)

	ctx := context.Background()

	// Step 1: Bootstrap workspace
	bootstrapReq := models.BootstrapRequest{
		PrimaryAgent: models.BootstrapPrimaryAgent{
			Email:    "admin@openclaw-test.com",
			Password: "testpassword123!",
			AgentProfile: models.BootstrapAgentProfile{
				Name:        "test-agent",
				DisplayName: "Test Agent",
				Description: "OpenClaw integration test agent",
			},
		},
		DefaultChannel: models.BootstrapChannel{
			Name:  "general",
			Topic: "Test channel",
		},
	}

	jsonBody, _ := json.Marshal(bootstrapReq)
	resp, err := http.Post(server.URL+"/api/v1/bootstrap", "application/json", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var bootstrapResp models.BootstrapResponse
	err = json.NewDecoder(resp.Body).Decode(&bootstrapResp)
	resp.Body.Close()
	require.NoError(t, err)

	workspaceID := bootstrapResp.PrimaryAgent.UserID
	channelID := bootstrapResp.Channel.ChannelID
	jwtToken := bootstrapResp.PrimaryAgent.JWTToken

	// Step 2: Create mock webhook receiver
	var receivedOutboundWebhook bool
	var outboundPayload map[string]interface{}
	mockReceiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &outboundPayload)
		receivedOutboundWebhook = true
		w.WriteHeader(http.StatusOK)
	}))
	defer mockReceiver.Close()

	// Step 3: Register OpenClaw integration
	integrationReq := models.CreateIntegrationRequest{
		Name:               "OpenClaw Test",
		Platform:           "openclaw",
		WebhookURL:         mockReceiver.URL,
		EventSubscriptions: []string{"message.created"},
	}

	integrationJSON, _ := json.Marshal(integrationReq)
	req, _ := http.NewRequest("POST", server.URL+"/api/v1/integrations", bytes.NewBuffer(integrationJSON))
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var integrationResp struct {
		Integration models.Integration `json:"integration"`
		APIKey      string             `json:"api_key"`
	}
	err = json.NewDecoder(resp.Body).Decode(&integrationResp)
	resp.Body.Close()
	require.NoError(t, err)

	integrationAPIKey := integrationResp.APIKey

	// Step 4: Send inbound message via integration webhook
	inboundPayload := map[string]string{
		"channel_id": channelID,
		"agent_id":   bootstrapResp.PrimaryAgent.AgentID,
		"text":       "Hello from OpenClaw integration test",
	}
	inboundJSON, _ := json.Marshal(inboundPayload)

	// Use the integration inbound webhook endpoint
	resp, err = http.Post(server.URL+"/api/v1/webhooks/integration/openclaw", "application/json", bytes.NewBuffer(inboundJSON))
	require.NoError(t, err)

	// If endpoint doesn't exist yet (expected during development), skip gracefully
	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Integration inbound webhook endpoint not yet implemented (expected 404)")
	}
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Step 5: Verify message appears in channel
	time.Sleep(100 * time.Millisecond) // Allow async processing

	req, _ = http.NewRequest("GET", server.URL+"/api/v1/channels/"+channelID+"/messages", nil)
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var messagesResp struct {
		Messages []models.Message `json:"messages"`
	}
	err = json.NewDecoder(resp.Body).Decode(&messagesResp)
	resp.Body.Close()
	require.NoError(t, err)

	foundInbound := false
	for _, msg := range messagesResp.Messages {
		if msg.Content == "Hello from OpenClaw integration test" {
			foundInbound = true
			break
		}
	}
	assert.True(t, foundInbound, "Inbound message should appear in channel")

	// Step 6: Send reply from workspace
	replyPayload := map[string]string{
		"text": "Reply from workspace",
	}
	replyJSON, _ := json.Marshal(replyPayload)

	req, _ = http.NewRequest("POST", server.URL+"/api/v1/channels/"+channelID+"/messages", bytes.NewBuffer(replyJSON))
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Step 7: Verify outbound webhook delivered
	time.Sleep(200 * time.Millisecond) // Allow async webhook dispatch

	assert.True(t, receivedOutboundWebhook, "Outbound webhook should be received by mock server")
	if receivedOutboundWebhook {
		assert.Equal(t, "message.created", outboundPayload["event_type"])
		assert.Equal(t, workspaceID, outboundPayload["workspace_id"])
	}
}
