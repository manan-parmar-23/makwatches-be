#!/bin/bash

# Deployment Verification Script
# Usage: ./verify-deployment.sh [server-ip-or-domain]

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
SERVER="${1:-139.59.71.95}"
PORT="${2:-8080}"

echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  MakWatches Deployment Verification       ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
    fi
}

echo -e "${YELLOW}Testing server: ${SERVER}:${PORT}${NC}"
echo ""

# Test 1: Health endpoint
echo "🔍 Test 1: Health Check..."
HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" http://${SERVER}:${PORT}/health)
HTTP_CODE=$(echo "$HEALTH_RESPONSE" | tail -n1)
BODY=$(echo "$HEALTH_RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "200" ]; then
    print_status 0 "Health endpoint responding (HTTP $HTTP_CODE)"
    echo -e "   Response: ${BLUE}${BODY}${NC}"
else
    print_status 1 "Health endpoint failed (HTTP $HTTP_CODE)"
    echo -e "   Response: ${RED}${BODY}${NC}"
fi
echo ""

# Test 2: Welcome endpoint (the one you just changed)
echo "🔍 Test 2: Welcome Message (Your Recent Change)..."
WELCOME_RESPONSE=$(curl -s -w "\n%{http_code}" http://${SERVER}:${PORT}/)
HTTP_CODE=$(echo "$WELCOME_RESPONSE" | tail -n1)
BODY=$(echo "$WELCOME_RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "200" ]; then
    print_status 0 "Welcome endpoint responding (HTTP $HTTP_CODE)"
    echo -e "   Response: ${BLUE}${BODY}${NC}"
    
    # Check if the message contains the updated text
    if echo "$BODY" | grep -q "Makwatches API's"; then
        echo -e "   ${GREEN}✅ Your change IS LIVE! Message contains \"API's\"${NC}"
    else
        echo -e "   ${YELLOW}⚠️  Your change NOT LIVE yet. Old message detected.${NC}"
    fi
else
    print_status 1 "Welcome endpoint failed (HTTP $HTTP_CODE)"
    echo -e "   Response: ${RED}${BODY}${NC}"
fi
echo ""

# Test 3: Products endpoint
echo "🔍 Test 3: Products API..."
PRODUCTS_RESPONSE=$(curl -s -w "\n%{http_code}" http://${SERVER}:${PORT}/api/v1/products?limit=1)
HTTP_CODE=$(echo "$PRODUCTS_RESPONSE" | tail -n1)

if [ "$HTTP_CODE" = "200" ]; then
    print_status 0 "Products endpoint working (HTTP $HTTP_CODE)"
else
    print_status 1 "Products endpoint issue (HTTP $HTTP_CODE)"
fi
echo ""

# Test 4: Docker container status (requires SSH access)
echo "🔍 Test 4: Docker Container Status..."
if command -v ssh &> /dev/null; then
    echo "   Attempting SSH check..."
    # This will only work if you have SSH keys set up
    CONTAINER_STATUS=$(ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 root@${SERVER} "docker ps | grep makwatches-be-api" 2>/dev/null || echo "SSH_FAILED")
    
    if [ "$CONTAINER_STATUS" != "SSH_FAILED" ] && [ -n "$CONTAINER_STATUS" ]; then
        print_status 0 "Container is running"
        echo -e "   ${BLUE}${CONTAINER_STATUS}${NC}"
    else
        echo -e "   ${YELLOW}⚠️  Cannot verify via SSH (requires key authentication)${NC}"
    fi
else
    echo -e "   ${YELLOW}⚠️  SSH not available for container check${NC}"
fi
echo ""

# Summary
echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Verification Summary                      ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""
echo -e "Server: ${GREEN}${SERVER}:${PORT}${NC}"
echo -e "Timestamp: $(date)"
echo ""
echo -e "${YELLOW}💡 Tips:${NC}"
echo "  - If changes aren't live, wait 1-2 minutes for deployment"
echo "  - Check GitHub Actions: https://github.com/manan-parmar-23/makwatches-be/actions"
echo "  - View server logs: ssh root@${SERVER} 'docker logs makwatches-be-api --tail=50'"
echo ""
