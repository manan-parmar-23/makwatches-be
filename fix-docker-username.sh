#!/bin/bash

# Fix Docker Username Configuration
# This script updates the DOCKER_USERNAME on your production server

echo "🔧 Fixing Docker Username Configuration..."
echo ""

SERVER_HOST="139.59.71.95"
SERVER_USER="root"
CORRECT_USERNAME="aditya110"

echo "Target Server: $SERVER_HOST"
echo "Correct Docker Username: $CORRECT_USERNAME"
echo ""

# Create a temporary script to run on the server
cat > /tmp/fix_docker_env.sh << 'SCRIPT_EOF'
#!/bin/bash

cd /opt/makwatches-be

echo "Current DOCKER_USERNAME configuration:"
grep DOCKER_USERNAME .env || echo "DOCKER_USERNAME not found in .env"

echo ""
echo "Updating DOCKER_USERNAME to: aditya110"

# Backup current .env
cp .env .env.backup

# Remove old DOCKER_USERNAME line and add correct one
grep -v "^DOCKER_USERNAME=" .env > .env.tmp
echo "DOCKER_USERNAME=aditya110" >> .env.tmp
mv .env.tmp .env

echo ""
echo "✅ Updated configuration:"
grep DOCKER_USERNAME .env

echo ""
echo "Stopping current containers..."
docker compose down

echo ""
echo "Removing old incorrect images..."
docker rmi adityagarg646/makwatches-be:latest 2>/dev/null || echo "Image not found (OK)"

echo ""
echo "Pulling correct image from aditya110/makwatches-be:latest..."
docker pull aditya110/makwatches-be:latest

echo ""
echo "Starting services with correct image..."
docker compose up -d

echo ""
echo "Checking container status..."
sleep 5
docker ps | grep makwatches-be

echo ""
echo "✅ Docker username fix complete!"
echo ""
echo "Verify deployment:"
echo "  curl http://localhost:8080/health"

SCRIPT_EOF

echo "📤 Uploading fix script to server..."
scp /tmp/fix_docker_env.sh root@$SERVER_HOST:/tmp/fix_docker_env.sh

echo ""
echo "🚀 Running fix on server..."
ssh root@$SERVER_HOST "chmod +x /tmp/fix_docker_env.sh && /tmp/fix_docker_env.sh"

echo ""
echo "🎉 Done! Now update GitHub Secrets:"
echo ""
echo "1. Go to: https://github.com/manan-parmar-23/makwatches-be/settings/secrets/actions"
echo "2. Click on DOCKER_USERNAME"
echo "3. Update value to: aditya110"
echo "4. Save"
echo ""
echo "After updating GitHub Secrets, push a new commit to trigger deployment:"
echo "  git commit --allow-empty -m 'Trigger deployment with correct username'"
echo "  git push origin main"
