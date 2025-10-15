#!/bin/bash

# Quick Environment Update Script for MakWatches Backend
# This script quickly updates the running container with new environment variables

set -e

# Configuration
CONTAINER_NAME="makwatches-be-api"
IMAGE_NAME="aditya110/makwatches-be:latest"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}[INFO]${NC} Quick environment update for MakWatches Backend"
echo ""

# Check if container is running
if ! docker ps | grep -q "$CONTAINER_NAME"; then
    echo -e "${YELLOW}[WARNING]${NC} Container $CONTAINER_NAME is not running."
    echo "Use the full update script: ./update-env-and-redeploy.sh"
    exit 1
fi

echo -e "${BLUE}[INFO]${NC} Pulling latest image..."
docker pull "$IMAGE_NAME"

echo -e "${BLUE}[INFO]${NC} Stopping current container..."
docker stop "$CONTAINER_NAME" && docker rm "$CONTAINER_NAME"

echo -e "${BLUE}[INFO]${NC} Starting container with updated environment..."
docker compose -f docker-compose.prod.yml up -d

echo -e "${BLUE}[INFO]${NC} Waiting for container to be ready..."
sleep 10

# Check if container is healthy
if docker ps | grep -q "$CONTAINER_NAME"; then
    echo -e "${GREEN}[SUCCESS]${NC} Container updated successfully!"
    echo ""
    echo "Container status:"
    docker ps --filter "name=$CONTAINER_NAME" --format "table {{.ID}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}"
else
    echo -e "${YELLOW}[WARNING]${NC} Container may not be running properly. Check logs:"
    docker logs "$CONTAINER_NAME"
fi
