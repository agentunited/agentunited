package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentunited/backend/internal/config"
	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupIntegrationTest creates a test database and router
func setupIntegrationTest(t *testing.T) (*repository.DB, *httptest.Server) {
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		Database: "agentunited_test",
		SSLMode:  "disable",
	}

	ctx := context.Background()
	db, err := repository.NewDB(ctx, cfg)
	require.NoError(t, err, "Failed to connect to test database")

	// Clean database
	_, err = db.Pool.Exec(ctx, "DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;")
	require.NoError(t, err, "Failed to reset schema")

	// Run migrations
	err = db.RunMigrations(ctx, "../../migrations")
	require.NoError(t, err, "Failed to run migrations")

	// Create router
	cache := &repository.Cache{} // Mock cache for tests
	router := NewRouter(db, cache, "test-jwt-secret")

	server := httptest.NewServer(router)
	return db, server
}

func cleanupIntegrationTest(t *testing.T, db *repository.DB, server *httptest.Server) {
	server.Close()
	db.Close()
}

func TestBootstrapIntegration_HappyPath(t *testing.T) {
	db, server := setupIntegrationTest(t)
	defer cleanupIntegrationTest(t, db, server)

	// 1. Bootstrap the instance
	bootstrapReq := models.BootstrapRequest{
		PrimaryAgent: models.BootstrapPrimaryAgent{
			Email:    "admin@example.com",
			Password: "supersecurepassword123",
			AgentProfile: models.BootstrapAgentProfile{
				Name:        "coordinator",
				DisplayName: "Coordination Agent",
				Description: "Main coordination agent",
			},
		},
		Agents: []models.BootstrapAgent{
			{
				Name:        "worker",
				DisplayName: "Worker Agent",
				Description: "Background task handler",
			},
		},
		Humans: []models.BootstrapHuman{
			{
				Email:       "human@example.com",
				DisplayName: "Test Human",
				Role:        "member",
			},
		},
		DefaultChannel: models.BootstrapChannel{
			Name:  "general",
			Topic: "Team coordination",
		},
	}

	jsonBody, _ := json.Marshal(bootstrapReq)
	resp, err := http.Post(server.URL+"/api/v1/bootstrap", "application/json", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify bootstrap response
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var bootstrapResp models.BootstrapResponse
	err = json.NewDecoder(resp.Body).Decode(&bootstrapResp)
	require.NoError(t, err)

	// Verify primary agent
	assert.Equal(t, "admin@example.com", bootstrapResp.PrimaryAgent.Email)
	assert.NotEmpty(t, bootstrapResp.PrimaryAgent.UserID)
	assert.NotEmpty(t, bootstrapResp.PrimaryAgent.AgentID)
	assert.NotEmpty(t, bootstrapResp.PrimaryAgent.JWTToken)
	assert.NotEmpty(t, bootstrapResp.PrimaryAgent.APIKey)
	assert.True(t, len(bootstrapResp.PrimaryAgent.APIKey) > 40) // au_ + base64
	assert.Contains(t, bootstrapResp.PrimaryAgent.APIKey, "au_")

	// Verify additional agents
	require.Len(t, bootstrapResp.Agents, 1)
	assert.Equal(t, "worker", bootstrapResp.Agents[0].Name)
	assert.Equal(t, "Worker Agent", bootstrapResp.Agents[0].DisplayName)
	assert.NotEmpty(t, bootstrapResp.Agents[0].APIKey)
	assert.Contains(t, bootstrapResp.Agents[0].APIKey, "au_")

	// Verify humans
	require.Len(t, bootstrapResp.Humans, 1)
	assert.Equal(t, "human@example.com", bootstrapResp.Humans[0].Email)
	assert.NotEmpty(t, bootstrapResp.Humans[0].InviteToken)
	assert.NotEmpty(t, bootstrapResp.Humans[0].InviteURL)
	assert.Contains(t, bootstrapResp.Humans[0].InviteToken, "inv_")
	assert.Contains(t, bootstrapResp.Humans[0].InviteURL, "token=")

	// Verify channel
	assert.Equal(t, "general", bootstrapResp.Channel.Name)
	assert.Equal(t, "Team coordination", bootstrapResp.Channel.Topic)
	assert.Len(t, bootstrapResp.Channel.Members, 2) // primary + human

	// 2. Test invite validation
	inviteToken := bootstrapResp.Humans[0].InviteToken
	validateResp, err := http.Get(server.URL + "/api/v1/invite?token=" + inviteToken)
	require.NoError(t, err)
	defer validateResp.Body.Close()

	assert.Equal(t, http.StatusOK, validateResp.StatusCode)

	var validateBody map[string]interface{}
	err = json.NewDecoder(validateResp.Body).Decode(&validateBody)
	require.NoError(t, err)

	assert.Equal(t, "human@example.com", validateBody["email"])
	assert.Equal(t, "pending", validateBody["status"])
	assert.NotEmpty(t, validateBody["expires_at"])

	// 3. Test invite acceptance
	acceptReq := models.InviteAcceptRequest{
		Token:    inviteToken,
		Password: "humanpassword123",
	}

	acceptJSON, _ := json.Marshal(acceptReq)
	acceptResp, err := http.Post(server.URL+"/api/v1/invite/accept", "application/json", bytes.NewBuffer(acceptJSON))
	require.NoError(t, err)
	defer acceptResp.Body.Close()

	assert.Equal(t, http.StatusOK, acceptResp.StatusCode)

	var acceptBody map[string]string
	err = json.NewDecoder(acceptResp.Body).Decode(&acceptBody)
	require.NoError(t, err)

	assert.NotEmpty(t, acceptBody["jwt_token"])
	assert.Contains(t, acceptBody["message"], "accepted successfully")

	// 4. Test that invite token is consumed (can't be used again)
	secondAcceptResp, err := http.Post(server.URL+"/api/v1/invite/accept", "application/json", bytes.NewBuffer(acceptJSON))
	require.NoError(t, err)
	defer secondAcceptResp.Body.Close()

	assert.Equal(t, http.StatusConflict, secondAcceptResp.StatusCode) // 409 for already consumed
}

func TestBootstrapIntegration_Idempotency(t *testing.T) {
	db, server := setupIntegrationTest(t)
	defer cleanupIntegrationTest(t, db, server)

	bootstrapReq := models.BootstrapRequest{
		PrimaryAgent: models.BootstrapPrimaryAgent{
			Email:    "admin@example.com",
			Password: "supersecurepassword123",
			AgentProfile: models.BootstrapAgentProfile{
				Name:        "coordinator",
				DisplayName: "Coordination Agent",
			},
		},
	}

	// First call - should succeed
	jsonBody, _ := json.Marshal(bootstrapReq)
	resp1, err := http.Post(server.URL+"/api/v1/bootstrap", "application/json", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	defer resp1.Body.Close()

	assert.Equal(t, http.StatusCreated, resp1.StatusCode)

	// Second call - should return 409 Conflict
	jsonBody2, _ := json.Marshal(bootstrapReq)
	resp2, err := http.Post(server.URL+"/api/v1/bootstrap", "application/json", bytes.NewBuffer(jsonBody2))
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusConflict, resp2.StatusCode)

	var errorResp map[string]string
	err = json.NewDecoder(resp2.Body).Decode(&errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp["error"], "already been bootstrapped")
}

func TestInviteIntegration_ErrorCases(t *testing.T) {
	db, server := setupIntegrationTest(t)
	defer cleanupIntegrationTest(t, db, server)

	// Test invalid token
	resp, err := http.Get(server.URL + "/api/v1/invite?token=invalid-token")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	// Test missing token
	resp2, err := http.Get(server.URL + "/api/v1/invite")
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp2.StatusCode)

	// Test weak password on accept
	acceptReq := models.InviteAcceptRequest{
		Token:    "any-token",
		Password: "weak", // Too short
	}

	acceptJSON, _ := json.Marshal(acceptReq)
	acceptResp, err := http.Post(server.URL+"/api/v1/invite/accept", "application/json", bytes.NewBuffer(acceptJSON))
	require.NoError(t, err)
	defer acceptResp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, acceptResp.StatusCode)

	var errorResp map[string]interface{}
	err = json.NewDecoder(acceptResp.Body).Decode(&errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp["error"], "validation")
}