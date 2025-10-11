#!/bin/bash

# Fix Redis DNS Resolution in Docker

echo "🔧 Fixing Redis DNS Resolution Issue..."
echo ""

cat << 'EOF'

ISSUE IDENTIFIED:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Error: "no such host" for redis-14568.c301.ap-south-1-1.ec2.redns.redis-cloud.com

This is a Docker DNS resolution issue. The container can't resolve the Redis hostname.

SOLUTION:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Option 1: Use Google DNS (8.8.8.8) in docker-compose.yml
Option 2: Update Redis URI format

Let me apply the fix...

EOF

# Create updated docker-compose.yml with DNS fix
cat > /tmp/docker-compose.prod.yml.new << 'COMPOSE_EOF'
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

echo "✅ Created updated docker-compose.prod.yml with DNS fix"
echo ""
echo "📋 Changes made:"
echo "  • Added Google DNS servers (8.8.8.8, 8.8.4.4)"
echo "  • This will help resolve Redis Cloud hostname"
echo ""
echo "Now updating Redis configuration..."
echo ""

cat << 'REDIS_EOF'

REDIS CONFIGURATION UPDATE:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Your Redis details:
  Host: redis-14568.c301.ap-south-1-1.ec2.redns.redis-cloud.com
  Port: 14568
  Password: A2kpg3t6swc401ilemy8y452qyoz6l6b3rdu6ebj0e69pvfouy3

Update your .env file with:

REDIS_URI=redis://redis-14568.c301.ap-south-1-1.ec2.redns.redis-cloud.com:14568
REDIS_PASSWORD=A2kpg3t6swc401ilemy8y452qyoz6l6b3rdu6ebj0e69pvfouy3

REDIS_EOF

echo ""
echo "Files ready to upload to server!"
