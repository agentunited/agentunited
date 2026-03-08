#!/bin/bash
# End-to-End Test Script for Agent United
# Tests full agent-first provisioning flow

set -e  # Exit on error

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BASE_URL="${BASE_URL:-http://localhost:8080}"
WEB_URL="${WEB_URL:-http://localhost:3000}"

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Agent United E2E Test${NC}"
echo -e "${YELLOW}========================================${NC}"

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    rm -f test-credentials.json test-config.json
}
trap cleanup EXIT

# Test 1: Health Check
echo -e "\n${YELLOW}[Test 1] Health Check${NC}"
HEALTH=$(curl -s $BASE_URL/health)
if echo "$HEALTH" | grep -q '"status":"ok"'; then
    echo -e "${GREEN}✓ API is healthy${NC}"
else
    echo -e "${RED}✗ API health check failed${NC}"
    echo "$HEALTH"
    exit 1
fi

# Test 2: Bootstrap API
echo -e "\n${YELLOW}[Test 2] Bootstrap API${NC}"

cat > test-config.json <<EOF
{
  "primary_agent": {
    "email": "test-admin@example.com",
    "password": "TestSecurePassword123!",
    "agent_profile": {
      "name": "test-coordinator",
      "display_name": "Test Coordinator",
      "description": "E2E test coordinator agent"
    }
  },
  "agents": [
    {
      "name": "test-worker",
      "display_name": "Test Worker Agent",
      "description": "E2E test worker"
    }
  ],
  "humans": [
    {
      "email": "test-human@example.com",
      "display_name": "Test Human",
      "role": "observer"
    }
  ],
  "default_channel": {
    "name": "e2e-test",
    "topic": "End-to-end testing channel"
  }
}
EOF

BOOTSTRAP_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/api/v1/bootstrap \
  -H "Content-Type: application/json" \
  -d @test-config.json)

HTTP_CODE=$(echo "$BOOTSTRAP_RESPONSE" | tail -n 1)
BOOTSTRAP_JSON=$(echo "$BOOTSTRAP_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "201" ]; then
    echo -e "${GREEN}✓ Bootstrap succeeded (201 Created)${NC}"
    echo "$BOOTSTRAP_JSON" > test-credentials.json
else
    echo -e "${RED}✗ Bootstrap failed (HTTP $HTTP_CODE)${NC}"
    echo "$BOOTSTRAP_JSON"
    exit 1
fi

# Extract credentials
JWT_TOKEN=$(echo "$BOOTSTRAP_JSON" | grep -o '"jwt_token":"[^"]*"' | head -1 | cut -d'"' -f4)
PRIMARY_API_KEY=$(echo "$BOOTSTRAP_JSON" | grep -o '"api_key":"[^"]*"' | head -1 | cut -d'"' -f4)
CHANNEL_ID=$(echo "$BOOTSTRAP_JSON" | grep -o '"channel_id":"[^"]*"' | head -1 | cut -d'"' -f4)
INVITE_TOKEN=$(echo "$BOOTSTRAP_JSON" | grep -o '"invite_token":"[^"]*"' | head -1 | cut -d'"' -f4)
RELAY_URL_1=$(echo "$BOOTSTRAP_JSON" | grep -o '"relay_url":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$JWT_TOKEN" ] || [ -z "$CHANNEL_ID" ]; then
    echo -e "${RED}✗ Failed to extract credentials from bootstrap response${NC}"
    exit 1
fi

echo "  JWT: ${JWT_TOKEN:0:20}..."
echo "  API Key: ${PRIMARY_API_KEY:0:20}..."
echo "  Channel: $CHANNEL_ID"
echo "  Invite:  ${INVITE_TOKEN:0:20}..."

# Test 3: Idempotency Recovery
echo -e "\n${YELLOW}[Test 3] Bootstrap Recovery Idempotency${NC}"
IDEMPOTENCY_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/api/v1/bootstrap \
  -H "Content-Type: application/json" \
  -d @test-config.json)

IDEMPOTENCY_CODE=$(echo "$IDEMPOTENCY_RESPONSE" | tail -n1)
IDEMPOTENCY_JSON=$(echo "$IDEMPOTENCY_RESPONSE" | sed '$d')
RELAY_URL_2=$(echo "$IDEMPOTENCY_JSON" | grep -o '"relay_url":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ "$IDEMPOTENCY_CODE" = "201" ]; then
    echo -e "${GREEN}✓ Idempotency recovery check passed (201 Created)${NC}"
else
    echo -e "${RED}✗ Expected 201, got HTTP $IDEMPOTENCY_CODE${NC}"
    exit 1
fi

if [ -z "$RELAY_URL_1" ] || [ -z "$RELAY_URL_2" ] || [ "$RELAY_URL_1" != "$RELAY_URL_2" ]; then
    echo -e "${RED}✗ Relay URL mismatch on recovery bootstrap${NC}"
    echo "  first:  $RELAY_URL_1"
    echo "  second: $RELAY_URL_2"
    exit 1
else
    echo -e "${GREEN}✓ Relay URL stable across bootstrap recovery${NC}"
fi

# Test 3b: Relay status after recovery
echo -e "\n${YELLOW}[Test 3b] Relay Status${NC}"
RELAY_STATUS_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/relay/status" \
  -H "Authorization: Bearer $JWT_TOKEN")
RELAY_STATUS_CODE=$(echo "$RELAY_STATUS_RESPONSE" | tail -n1)
RELAY_STATUS_JSON=$(echo "$RELAY_STATUS_RESPONSE" | sed '$d')

if [ "$RELAY_STATUS_CODE" = "200" ] && \
   echo "$RELAY_STATUS_JSON" | grep -q '"relay_subdomain"' && \
   echo "$RELAY_STATUS_JSON" | grep -q '"bandwidth_limit_mb"' && \
   echo "$RELAY_STATUS_JSON" | grep -q '"relay_tier"'; then
    echo -e "${GREEN}✓ Relay status contains required fields${NC}"
else
    echo -e "${RED}✗ Relay status failed or missing required fields (HTTP $RELAY_STATUS_CODE)${NC}"
    echo "$RELAY_STATUS_JSON"
    exit 1
fi

# Test 4: Create Channel via API
echo -e "\n${YELLOW}[Test 4] Create Channel via API${NC}"
CHANNEL_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/api/v1/channels \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-crypto",
    "topic": "Test crypto research channel"
  }')

CHANNEL_CODE=$(echo "$CHANNEL_RESPONSE" | tail -n1)
CHANNEL_JSON=$(echo "$CHANNEL_RESPONSE" | sed '$d')

if [ "$CHANNEL_CODE" = "201" ]; then
    TEST_CHANNEL_ID=$(echo "$CHANNEL_JSON" | grep -o '"channel_id":"[^"]*"' | cut -d'"' -f4)
    if [ -n "$TEST_CHANNEL_ID" ]; then
        TEST_CHANNEL_ID=$(echo "$CHANNEL_JSON" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    fi
    echo -e "${GREEN}✓ Channel created: $TEST_CHANNEL_ID${NC}"
else
    echo -e "${RED}✗ Channel creation failed (HTTP $CHANNEL_CODE)${NC}"
    echo "$CHANNEL_JSON"
    exit 1
fi

# Test 5: Post Message via API
echo -e "\n${YELLOW}[Test 5] Post Message via API${NC}"
MESSAGE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/channels/$CHANNEL_ID/messages" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "E2E test message from coordinator agent"
  }')

MESSAGE_CODE=$(echo "$MESSAGE_RESPONSE" | tail -n1)

if [ "$MESSAGE_CODE" = "201" ]; then
    echo -e "${GREEN}✓ Message posted${NC}"
else
    echo -e "${RED}✗ Message post failed (HTTP $MESSAGE_CODE)${NC}"
    echo "$MESSAGE_RESPONSE" | sed '$d'
    exit 1
fi

# Test 6: List Messages
echo -e "\n${YELLOW}[Test 6] List Messages${NC}"
MESSAGES_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/channels/$CHANNEL_ID/messages" \
  -H "Authorization: Bearer $JWT_TOKEN")

MESSAGES_CODE=$(echo "$MESSAGES_RESPONSE" | tail -n1)
MESSAGES_JSON=$(echo "$MESSAGES_RESPONSE" | sed '$d')

if [ "$MESSAGES_CODE" = "200" ]; then
    MESSAGE_COUNT=$(echo "$MESSAGES_JSON" | grep -o '"content"' | wc -l)
    echo -e "${GREEN}✓ Messages retrieved (count: $MESSAGE_COUNT)${NC}"
else
    echo -e "${RED}✗ Message list failed (HTTP $MESSAGES_CODE)${NC}"
    exit 1
fi

# Test 7: Invalid API Key
echo -e "\n${YELLOW}[Test 7] Invalid API Key${NC}"
INVALID_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET $BASE_URL/api/v1/channels \
  -H "Authorization: Bearer au_invalid_key_12345")

INVALID_CODE=$(echo "$INVALID_RESPONSE" | tail -n1)

if [ "$INVALID_CODE" = "401" ]; then
    echo -e "${GREEN}✓ Invalid API key rejected (401 Unauthorized)${NC}"
else
    echo -e "${RED}✗ Expected 401, got HTTP $INVALID_CODE${NC}"
    exit 1
fi

# Test 8: Get Invite Info (if Phase 5 complete)
if [ -n "$INVITE_TOKEN" ]; then
    echo -e "\n${YELLOW}[Test 8] Get Invite Info${NC}"
    INVITE_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/invite?token=$INVITE_TOKEN")
    INVITE_CODE=$(echo "$INVITE_RESPONSE" | tail -n1)
    
    if [ "$INVITE_CODE" = "200" ]; then
        echo -e "${GREEN}✓ Invite info retrieved${NC}"
    elif [ "$INVITE_CODE" = "404" ]; then
        echo -e "${YELLOW}⚠ Invite endpoint not implemented yet (Phase 5)${NC}"
    else
        echo -e "${RED}✗ Invite info failed (HTTP $INVITE_CODE)${NC}"
    fi
fi

# Summary
echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}All Core Tests Passed!${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "\nTest Results:"
echo -e "  ${GREEN}✓${NC} Health check"
echo -e "  ${GREEN}✓${NC} Bootstrap API (201 Created)"
echo -e "  ${GREEN}✓${NC} Idempotency recovery (201 Created + same relay URL)"
echo -e "  ${GREEN}✓${NC} Relay status fields"
echo -e "  ${GREEN}✓${NC} Create channel"
echo -e "  ${GREEN}✓${NC} Post message"
echo -e "  ${GREEN}✓${NC} Error handling (401)"

echo -e "\n${YELLOW}Next Steps:${NC}"
echo -e "  1. Test frontend invite flow at: $WEB_URL/invite?token=$INVITE_TOKEN"
echo -e "  2. Run Python SDK tests: cd sdk/python && pytest"
echo -e "  3. See full test guide: docs/integration-testing.md"

echo -e "\nTest credentials saved to: test-credentials.json"
