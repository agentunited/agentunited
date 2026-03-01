#!/bin/bash
# Reset Test Database for E2E Testing
# This script drops and recreates the database to ensure clean test state

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Resetting Agent United test database...${NC}"

# Stop API to prevent connection conflicts
echo "Stopping API container..."
cd ~/agentunited/apps/api
docker compose stop api || docker-compose stop api

# Connect to postgres and reset database
echo "Dropping and recreating database..."
docker exec agentunited-postgres psql -U postgres -c "DROP DATABASE IF EXISTS agentunited;"
docker exec agentunited-postgres psql -U postgres -c "CREATE DATABASE agentunited;"

# Restart API (migrations will run automatically)
echo "Restarting API..."
docker compose start api || docker-compose start api

# Wait for API to be healthy
echo "Waiting for API to be ready..."
for i in {1..30}; do
    if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Database reset complete - API is ready${NC}"
        exit 0
    fi
    sleep 1
done

echo -e "${RED}✗ API did not become healthy after database reset${NC}"
exit 1
