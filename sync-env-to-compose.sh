#!/bin/bash

# Dynamic Environment Variable Sync Script
# Automatically updates docker-compose.prod.yml with ALL variables from .env

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="$SCRIPT_DIR/.env"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.prod.yml"

echo "================================================"
echo "  Dynamic Environment Variable Sync"
echo "================================================"
echo ""

# Check if .env file exists
if [ ! -f "$ENV_FILE" ]; then
    echo "[ERROR] .env file not found: $ENV_FILE"
    exit 1
fi

# Backup current compose file
cp "$COMPOSE_FILE" "$COMPOSE_FILE.backup.$(date +%Y%m%d_%H%M%S)"
echo "[INFO] Backed up docker-compose.prod.yml"

# Read all environment variables from .env (excluding comments and empty lines)
echo "[INFO] Reading environment variables from .env..."
ENV_VARS=$(grep -v '^#' "$ENV_FILE" | grep -v '^$' | cut -d'=' -f1 | sort)

# Generate environment section for docker-compose
ENV_SECTION=""
while IFS= read -r var; do
    if [ -n "$var" ]; then
        ENV_SECTION="${ENV_SECTION}      - ${var}=\${${var}}\n"
    fi
done <<< "$ENV_VARS"

# Create new docker-compose.prod.yml with dynamic environment variables
cat > "$COMPOSE_FILE" << 'EOF'
version: '3.8'

services:
  api:
    image: ${DOCKER_USERNAME}/makwatches-be:latest
    container_name: makwatches-be-api
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./uploads:/app/uploads
      - ./firebase-admin.json:/app/firebase-admin.json:ro
    environment:
EOF

# Append the dynamic environment variables
echo -e "$ENV_SECTION" >> "$COMPOSE_FILE"

# Add the rest of the compose file
cat >> "$COMPOSE_FILE" << 'EOF'
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    networks:
      - makwatches-network

networks:
  makwatches-network:
    driver: bridge
EOF

echo "[SUCCESS] docker-compose.prod.yml updated with all environment variables"
echo ""

# Count variables
VAR_COUNT=$(echo "$ENV_VARS" | wc -l)
echo "[INFO] Synced $VAR_COUNT environment variables:"
echo "$ENV_VARS" | sed 's/^/  - /'
echo ""
echo "================================================"
echo "  Sync Complete!"
echo "================================================"
