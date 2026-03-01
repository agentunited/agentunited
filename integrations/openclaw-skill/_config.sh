#!/usr/bin/env bash
# Shared config loader for Agent United OpenClaw skill
# Sources credentials from env vars or ~/.agentunited/credentials.json
#
# Per-agent support:
#   Set AGENT_UNITED_AGENT_NAME to match a name in the credentials file
#   (e.g. "naomi", "golden") to use that agent's API key.
#   Falls back to primary_agent if not set.

CREDS_FILE="${HOME}/.agentunited/credentials.json"

if [ -n "${AGENT_UNITED_API_KEY:-}" ]; then
  # Use env vars directly
  AGENT_UNITED_URL="${AGENT_UNITED_URL:-http://localhost:8080}"
  AGENT_UNITED_CHANNEL_ID="${AGENT_UNITED_CHANNEL_ID:-}"
elif [ -f "$CREDS_FILE" ]; then
  AGENT_UNITED_URL="${AGENT_UNITED_URL:-http://localhost:8080}"
  AGENT_NAME="${AGENT_UNITED_AGENT_NAME:-}"

  if [ -n "$AGENT_NAME" ]; then
    # Look up this agent's key from the agents array
    AGENT_UNITED_API_KEY=$(python3 -c "
import json, sys
creds = json.load(open('$CREDS_FILE'))
name = '$AGENT_NAME'.lower()
# Check agents array
for a in creds.get('agents', []):
    if a.get('name', '').lower() == name:
        print(a['api_key']); sys.exit(0)
# Check if it's the primary agent
if creds.get('primary_agent', {}).get('email', '').lower().startswith(name + '@'):
    print(creds['primary_agent']['api_key']); sys.exit(0)
print('ERROR: Agent not found: ' + name, file=sys.stderr); sys.exit(1)
")
  else
    # Default to primary agent
    AGENT_UNITED_API_KEY=$(python3 -c "import json; print(json.load(open('$CREDS_FILE'))['primary_agent']['api_key'])")
  fi

  if [ -z "${AGENT_UNITED_CHANNEL_ID:-}" ]; then
    AGENT_UNITED_CHANNEL_ID=$(python3 -c "import json; print(json.load(open('$CREDS_FILE'))['channel']['channel_id'])")
  fi
else
  echo "Error: No Agent United credentials found." >&2
  echo "Set AGENT_UNITED_API_KEY env var or run setup.sh + provision.py first." >&2
  exit 1
fi

if [ -z "${AGENT_UNITED_CHANNEL_ID:-}" ]; then
  echo "Error: AGENT_UNITED_CHANNEL_ID not set and not found in credentials." >&2
  exit 1
fi

export AGENT_UNITED_URL AGENT_UNITED_API_KEY AGENT_UNITED_CHANNEL_ID
