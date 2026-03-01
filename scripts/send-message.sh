#!/usr/bin/env bash
# Send a message to Agent United from the command line
# Usage: ./send-message.sh "Hello from my agent!"
# Requires: AGENT_UNITED_API_KEY and AGENT_UNITED_CHANNEL_ID env vars
#   or a credentials file at ~/.agentunited/credentials.json

set -euo pipefail

MESSAGE="${1:?Usage: send-message.sh \"message text\"}"

# Load credentials
if [ -n "${AGENT_UNITED_API_KEY:-}" ] && [ -n "${AGENT_UNITED_CHANNEL_ID:-}" ]; then
    API_KEY="$AGENT_UNITED_API_KEY"
    CHANNEL_ID="$AGENT_UNITED_CHANNEL_ID"
    BASE_URL="${AGENT_UNITED_URL:-http://localhost:8080}"
elif [ -f "${HOME}/.agentunited/credentials.json" ]; then
    API_KEY=$(python3 -c "import json; d=json.load(open('$HOME/.agentunited/credentials.json')); print(d['primary_agent']['api_key'])")
    CHANNEL_ID=$(python3 -c "import json; d=json.load(open('$HOME/.agentunited/credentials.json')); print(d['channel']['channel_id'])")
    BASE_URL="${AGENT_UNITED_URL:-http://localhost:8080}"
else
    echo "Error: Set AGENT_UNITED_API_KEY + AGENT_UNITED_CHANNEL_ID env vars"
    echo "   or place credentials at ~/.agentunited/credentials.json"
    exit 1
fi

# Note: Until API key auth is implemented, use JWT token instead
# API_KEY should be the JWT token from bootstrap for now
RESPONSE=$(curl -s -w "\n%{http_code}" "${BASE_URL}/api/v1/channels/${CHANNEL_ID}/messages" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -d "{\"text\": $(echo "$MESSAGE" | python3 -c 'import json,sys; print(json.dumps(sys.stdin.read().strip()))')}")

HTTP_CODE=$(echo "$RESPONSE" | tail -1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" -ge 200 ] && [ "$HTTP_CODE" -lt 300 ]; then
    echo "✓ Message sent"
    echo "$BODY" | python3 -m json.tool 2>/dev/null || echo "$BODY"
else
    echo "✗ Failed (HTTP $HTTP_CODE)"
    echo "$BODY"
    exit 1
fi
