#!/usr/bin/env bash
set -euo pipefail

# Canonical safe restart for dogfood/local stack.
# IMPORTANT: Always rebuild web with --no-cache from latest main so nginx.conf/api proxy changes are included.

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

echo "[restart] syncing latest main..."
git switch main >/dev/null 2>&1 || true
git pull --ff-only

echo "[restart] rebuilding web image without cache..."
docker compose build --no-cache web

echo "[restart] restarting stack..."
docker compose up -d

echo "[restart] waiting for API to be healthy..."
for i in $(seq 1 30); do
  if curl -sf http://localhost:8080/health >/dev/null 2>&1; then break; fi
  sleep 2
done

echo "[restart] checking database state..."
USER_COUNT=$(docker exec agentunited-postgres psql -U "${DB_USER:-agentunited}" -d "${DB_NAME:-agentunited}" \
  -t -c "SELECT count(*) FROM users;" 2>/dev/null | tr -d ' \n' || echo "0")

if [ "$USER_COUNT" = "0" ]; then
  echo "[restart] DB is empty — running bootstrap..."
  BOOT=$(curl -s -X POST "http://localhost:8080/api/v1/bootstrap" \
    -H "Content-Type: application/json" \
    -d '{"workspace_name":"Agent United Dogfood","primary_agent":{"email":"empire@agentunited.ai","password":"dogfood-empire-2026","agent_profile":{"name":"empire","display_name":"Empire"}}}')
  echo "[restart] bootstrap response: $(echo "$BOOT" | grep -o '"error":"[^"]*"' || echo 'ok')"
  # Restart API to pick up relay token written by bootstrap
  docker compose restart api >/dev/null 2>&1 || true
  sleep 5
  echo "[restart] bootstrap complete — workspace ready."
else
  echo "[restart] DB has $USER_COUNT user(s) — skipping bootstrap."
fi

echo "[restart] done."
