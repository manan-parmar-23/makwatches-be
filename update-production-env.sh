#!/bin/bash

# Update Production Environment Variables
# This script updates the .env file in the production directory

set -e

PROD_DIR="/opt/makwatches-be"
CURRENT_DIR="$(pwd)"

echo "================================================"
echo "  Update Production Environment Variables"
echo "================================================"
echo ""

# Check if production directory exists
if [ ! -d "$PROD_DIR" ]; then
    echo "[ERROR] Production directory not found: $PROD_DIR"
    echo "[INFO] Creating production directory..."
    sudo mkdir -p "$PROD_DIR"
fi

# Check if .env file exists in current directory
if [ ! -f "$CURRENT_DIR/.env" ]; then
    echo "[ERROR] .env file not found in current directory"
    exit 1
fi

echo "[INFO] Backing up current production .env..."
if [ -f "$PROD_DIR/.env" ]; then
    sudo cp "$PROD_DIR/.env" "$PROD_DIR/.env.backup.$(date +%Y%m%d_%H%M%S)"
    echo "[SUCCESS] Backup created"
fi

echo "[INFO] Copying .env to production directory..."
sudo cp "$CURRENT_DIR/.env" "$PROD_DIR/.env"
echo "[SUCCESS] .env file updated in production directory"

echo "[INFO] Copying docker-compose.prod.yml to production directory..."
sudo cp "$CURRENT_DIR/docker-compose.prod.yml" "$PROD_DIR/docker-compose.yml"
echo "[SUCCESS] docker-compose.yml updated"

# Copy firebase-admin.json if it exists
if [ -f "$CURRENT_DIR/firebase-admin.json" ]; then
    echo "[INFO] Copying firebase-admin.json..."
    sudo cp "$CURRENT_DIR/firebase-admin.json" "$PROD_DIR/firebase-admin.json"
    echo "[SUCCESS] firebase-admin.json updated"
fi

echo ""
echo "[INFO] Files in production directory:"
sudo ls -la "$PROD_DIR" | grep -E "\\.env|\\.yml|firebase"

echo ""
echo "================================================"
echo "  Environment Update Complete!"
echo "================================================"
echo ""
echo "Next steps:"
echo "  1. Verify the .env file: sudo cat $PROD_DIR/.env"
echo "  2. Redeploy the application: cd $PROD_DIR && sudo docker compose pull && sudo docker compose up -d"
echo ""
