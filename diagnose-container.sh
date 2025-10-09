#!/bin/bash

# Script to diagnose why makwatches-be container keeps restarting

echo "🔍 Diagnosing MakWatches Container Issue..."
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check container status
echo "1️⃣  Container Status:"
docker ps -a --filter name=makwatches-be-api

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Get last 50 lines of logs
echo "2️⃣  Container Logs (Last 50 lines):"
docker logs --tail 50 makwatches-be-api 2>&1

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check if .env file exists and has required variables
echo "3️⃣  Environment Configuration Check:"
if [ -f /opt/makwatches-be/.env ]; then
    echo "✅ .env file exists"
    echo ""
    echo "Required variables present:"
    echo "MONGO_URI: $(grep -q 'MONGO_URI=' /opt/makwatches-be/.env && echo '✅ Present' || echo '❌ Missing')"
    echo "REDIS_URI: $(grep -q 'REDIS_URI=' /opt/makwatches-be/.env && echo '✅ Present' || echo '❌ Missing')"
    echo "JWT_SECRET: $(grep -q 'JWT_SECRET=' /opt/makwatches-be/.env && echo '✅ Present' || echo '❌ Missing')"
    echo "FIREBASE_PROJECT_ID: $(grep -q 'FIREBASE_PROJECT_ID=' /opt/makwatches-be/.env && echo '✅ Present' || echo '❌ Missing')"
else
    echo "❌ .env file NOT found at /opt/makwatches-be/.env"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check if firebase-admin.json exists
echo "4️⃣  Firebase Credentials Check:"
if [ -f /opt/makwatches-be/firebase-admin.json ]; then
    echo "✅ firebase-admin.json exists"
    echo "File size: $(stat -f%z /opt/makwatches-be/firebase-admin.json 2>/dev/null || stat -c%s /opt/makwatches-be/firebase-admin.json) bytes"
else
    echo "❌ firebase-admin.json NOT found at /opt/makwatches-be/firebase-admin.json"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check docker-compose.yml
echo "5️⃣  Docker Compose Configuration:"
if [ -f /opt/makwatches-be/docker-compose.yml ]; then
    echo "✅ docker-compose.yml exists"
    echo ""
    echo "Image configured:"
    grep "image:" /opt/makwatches-be/docker-compose.yml | head -1
else
    echo "❌ docker-compose.yml NOT found"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Get restart count
echo "6️⃣  Restart Count:"
docker inspect makwatches-be-api --format='{{.RestartCount}}' 2>/dev/null || echo "Could not get restart count"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "💡 Common Issues & Solutions:"
echo ""
echo "1. Missing environment variables"
echo "   → Check .env file has all required variables"
echo ""
echo "2. Firebase credentials invalid"
echo "   → Verify firebase-admin.json is valid JSON"
echo ""
echo "3. MongoDB/Redis connection failed"
echo "   → Check connection strings are correct"
echo "   → Verify database credentials"
echo ""
echo "4. Port already in use"
echo "   → Check if port 8080 is available: netstat -tuln | grep 8080"
echo ""
echo "5. Application crash on startup"
echo "   → Check container logs above for specific error"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
