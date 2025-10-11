#!/bin/bash

# Complete Fix Script for MakWatches Backend
# This script will fix the Redis DNS issue and setup the domain

echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║                                                                    ║"
echo "║   🔧 COMPLETE FIX FOR MAKWATCHES BACKEND                          ║"
echo "║                                                                    ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo ""

SERVER_IP="139.59.71.95"
DOMAIN="api.makwatches.in"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "STEP 1: Push Updated docker-compose.yml with DNS Fix"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "✅ docker-compose.prod.yml already updated with Google DNS"
echo "   This will fix: 'no such host' error for Redis"
echo ""

# Create the complete server fix script
cat > /tmp/server-fix.sh << 'SERVER_FIX_EOF'
#!/bin/bash

echo "🔧 Fixing Redis DNS and Configuration on Server..."
echo ""

cd /opt/makwatches-be || exit 1

# Step 1: Update docker-compose.yml with DNS fix
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1. Updating docker-compose.yml with DNS servers..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

cat > docker-compose.yml << 'COMPOSE_EOF'
services:
  api:
    image: ${DOCKER_USERNAME}/makwatches-be:latest
    container_name: makwatches-be-api
    restart: always
    ports:
      - "${PORT:-8080}:8080"
    dns:
      - 8.8.8.8
      - 8.8.4.4
    environment:
      - ENVIRONMENT=production
      - PORT=8080
      - MONGO_URI=${MONGO_URI}
      - DATABASE_NAME=${DATABASE_NAME:-makwatches}
      - REDIS_URI=${REDIS_URI}
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - REDIS_DATABASE_NAME=${REDIS_DATABASE_NAME}
      - JWT_SECRET=${JWT_SECRET}
      - JWT_EXPIRY=${JWT_EXPIRY:-24h}
      - GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID}
      - GOOGLE_CLIENT_SECRET=${GOOGLE_CLIENT_SECRET}
      - GOOGLE_REDIRECT_URL=${GOOGLE_REDIRECT_URL}
      - RAZORPAY_MODE=${RAZORPAY_MODE:-live}
      - RAZORPAY_KEY_ID_TEST=${RAZORPAY_KEY_ID_TEST}
      - RAZORPAY_KEY_SECRET_TEST=${RAZORPAY_KEY_SECRET_TEST}
      - FIREBASE_PROJECT_ID=${FIREBASE_PROJECT_ID}
      - FIREBASE_BUCKET_NAME=${FIREBASE_BUCKET_NAME}
      - FIREBASE_CREDENTIALS_PATH=/app/firebase-admin.json
      - VERCEL_ORIGIN=${VERCEL_ORIGIN:-https://mak-watches.vercel.app}
      - DEV_ORIGIN=${DEV_ORIGIN:-http://localhost:4200}
    volumes:
      - ./uploads:/app/uploads
      - ./firebase-admin.json:/app/firebase-admin.json:ro
    networks:
      - makwatches-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

networks:
  makwatches-network:
    driver: bridge

volumes:
  uploads:
COMPOSE_EOF

echo "✅ docker-compose.yml updated with DNS servers (8.8.8.8, 8.8.4.4)"
echo ""

# Step 2: Update Redis configuration in .env
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2. Updating Redis configuration in .env..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Backup .env
cp .env .env.backup.$(date +%s)

# Update Redis configuration
sed -i '/^REDIS_URI=/d' .env
sed -i '/^REDIS_PASSWORD=/d' .env

cat >> .env << 'ENV_EOF'

# Redis Configuration (Updated)
REDIS_URI=redis://redis-14568.c301.ap-south-1-1.ec2.redns.redis-cloud.com:14568
REDIS_PASSWORD=A2kpg3t6swc401ilemy8y452qyoz6l6b3rdu6ebj0e69pvfouy3
ENV_EOF

echo "✅ Redis configuration updated in .env"
echo ""

# Step 3: Restart containers
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3. Restarting containers..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

docker compose down
echo "✅ Containers stopped"
echo ""

docker compose up -d
echo "✅ Containers started with new configuration"
echo ""

# Wait for container to start
echo "Waiting 10 seconds for container to start..."
sleep 10

# Step 4: Check status
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4. Checking container status..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

docker ps | grep makwatches-be-api

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "5. Checking logs..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

docker logs --tail 30 makwatches-be-api

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "6. Testing health endpoint..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

sleep 5
if curl -f http://localhost:8080/health 2>/dev/null; then
    echo ""
    echo "✅ API is responding successfully!"
else
    echo ""
    echo "⚠️  API not responding yet. Check logs above."
fi

echo ""
echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║                                                                    ║"
echo "║   ✅ REDIS FIX COMPLETE!                                          ║"
echo "║                                                                    ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo ""
echo "Access your API at:"
echo "  • http://139.59.71.95:8080/health"
echo "  • View in Portainer: http://139.59.71.95:9000"
echo ""
echo "Next: Setup domain with 'bash setup-domain.sh'"
echo ""
SERVER_FIX_EOF

echo "✅ Server fix script created"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "STEP 2: Upload and Run Fix on Server"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Uploading fix script to server..."

# Upload script
scp /tmp/server-fix.sh root@${SERVER_IP}:/tmp/server-fix.sh

echo "✅ Script uploaded"
echo ""
echo "Running fix on server..."
echo ""

# Run the fix
ssh root@${SERVER_IP} "chmod +x /tmp/server-fix.sh && /tmp/server-fix.sh"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "STEP 3: Setup Domain (if DNS is ready)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Checking if DNS is configured..."
echo ""

if nslookup ${DOMAIN} | grep -q "${SERVER_IP}"; then
    echo "✅ DNS is configured correctly!"
    echo ""
    read -p "Setup domain with SSL now? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Setting up domain..."
        ssh root@${SERVER_IP} "cd /opt/makwatches-be && bash setup-domain.sh"
    else
        echo "⏩ Skipped. You can run it later with:"
        echo "   ssh root@${SERVER_IP}"
        echo "   cd /opt/makwatches-be"
        echo "   bash setup-domain.sh"
    fi
else
    echo "⚠️  DNS not configured yet or not propagated"
    echo ""
    echo "Add this DNS record:"
    echo "  Type: A"
    echo "  Name: api"
    echo "  Value: ${SERVER_IP}"
    echo ""
    echo "Then run domain setup:"
    echo "  ssh root@${SERVER_IP}"
    echo "  cd /opt/makwatches-be"
    echo "  bash setup-domain.sh"
fi

echo ""
echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║                                                                    ║"
echo "║   🎉 SETUP COMPLETE!                                              ║"
echo "║                                                                    ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo ""
echo "Your API endpoints:"
echo "  📍 Direct IP: http://${SERVER_IP}:8080/health"
echo "  🌐 Domain: http://${DOMAIN}/health (after domain setup)"
echo "  🔒 HTTPS: https://${DOMAIN}/health (after domain setup)"
echo ""
echo "Management:"
echo "  🐳 Portainer: http://${SERVER_IP}:9000"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
