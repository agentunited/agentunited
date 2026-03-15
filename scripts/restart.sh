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

echo "[restart] done."
