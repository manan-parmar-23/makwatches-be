#!/bin/bash

# One-Command Deployment Script
# Updates env, syncs to compose, commits, pushes, and triggers deployment

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "================================================"
echo "  🚀 One-Command Deployment Pipeline"
echo "================================================"
echo ""

# Step 1: Sync environment variables to docker-compose
echo "📋 Step 1: Syncing environment variables to docker-compose.prod.yml..."
./sync-env-to-compose.sh
echo ""

# Step 2: Update production environment
echo "📦 Step 2: Updating production environment..."
./update-production-env.sh
echo ""

# Step 3: Git operations
read -p "📝 Enter commit message (or press Enter for default): " COMMIT_MSG
if [ -z "$COMMIT_MSG" ]; then
    COMMIT_MSG="Update environment variables and configuration - $(date +%Y-%m-%d)"
fi

echo "🔄 Step 3: Committing changes..."
git add .env docker-compose.prod.yml sync-env-to-compose.sh
git commit -m "$COMMIT_MSG" || echo "No changes to commit"
echo ""

# Step 4: Push to trigger CI/CD
echo "⬆️  Step 4: Pushing to GitHub (triggers CI/CD)..."
git push
echo ""

# Step 5: Wait for CI/CD and redeploy
echo "⏳ Step 5: Waiting for CI/CD pipeline (30 seconds)..."
sleep 30
echo ""

echo "🔄 Step 6: Redeploying application..."
cd /opt/makwatches-be && docker compose pull && docker compose down && docker compose up -d
echo ""

echo "✅ Step 7: Verifying deployment..."
sleep 10
docker ps | grep makwatches
echo ""

echo "================================================"
echo "  ✨ Deployment Complete!"
echo "================================================"
echo ""
echo "📊 Quick Health Check:"
curl -s http://localhost:8080/health | head -20 || echo "Health check endpoint not responding yet..."
echo ""
echo ""
echo "📝 Next Steps:"
echo "  - View logs: docker logs makwatches-be-api -f"
echo "  - Check status: docker ps | grep makwatches"
echo "  - Test API: curl http://localhost:8080/health"
echo ""
