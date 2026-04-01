#!/usr/bin/env bash
# ⛔ DO NOT USE THIS SCRIPT DIRECTLY — the local ASC API key is invalid.
# Use the CI workflow instead:
#   gh workflow run build-ios-testflight.yml --repo agentunited/agentunited --ref main -f build_number=N
# See: skills/testflight-release/SKILL.md
echo "❌ This script is disabled. The local ASC API key does not work."
echo "   Use: gh workflow run build-ios-testflight.yml --repo agentunited/agentunited --ref main -f build_number=<N>"
exit 1
