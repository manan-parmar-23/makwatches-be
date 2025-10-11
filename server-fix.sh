#!/bin/bash

# RUN THIS SCRIPT ON YOUR SERVER (139.59.71.95)
# ssh root@139.59.71.95
# Then copy and paste this entire script

echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║                                                                    ║"
echo "║   🔧 FIX REDIS DNS + SETUP DOMAIN                                 ║"
echo "║                                                                    ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
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

# Remove old Redis config
sed -i '/^REDIS_URI=/d' .env
sed -i '/^REDIS_PASSWORD=/d' .env

# Add new Redis config
cat >> .env << 'ENV_EOF'

# Redis Configuration (Updated)
REDIS_URI=redis-14568.c301.ap-south-1-1.ec2.redns.redis-cloud.com:14568
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
echo "Waiting 15 seconds for container to start..."
sleep 15

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
if curl -f -s http://localhost:8080/health; then
    echo ""
    echo "✅ ✅ ✅ API IS WORKING! ✅ ✅ ✅"
    echo ""
    
    # Now setup domain
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "7. Setting up domain with Nginx + SSL..."
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    # Install Nginx if not present
    if ! command -v nginx &> /dev/null; then
        echo "Installing Nginx..."
        apt update -qq
        apt install -y nginx
        systemctl enable nginx
    fi
    
    # Create Nginx config
    cat > /etc/nginx/sites-available/api.makwatches.in << 'NGINX_EOF'
server {
    listen 80;
    listen [::]:80;
    server_name api.makwatches.in;
    
    location /.well-known/acme-challenge/ {
        root /var/www/html;
    }
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    access_log /var/log/nginx/api.makwatches.in.access.log;
    error_log /var/log/nginx/api.makwatches.in.error.log;
}
NGINX_EOF
    
    # Enable site
    ln -sf /etc/nginx/sites-available/api.makwatches.in /etc/nginx/sites-enabled/
    rm -f /etc/nginx/sites-enabled/default
    
    # Test and reload Nginx
    if nginx -t; then
        systemctl reload nginx
        echo "✅ Nginx configured"
    fi
    
    # Install Certbot and get SSL
    if ! command -v certbot &> /dev/null; then
        echo "Installing Certbot..."
        apt install -y certbot python3-certbot-nginx
    fi
    
    echo ""
    echo "Getting SSL certificate..."
    if certbot --nginx -d api.makwatches.in --non-interactive --agree-tos --email adityagarg646@gmail.com --redirect; then
        echo "✅ SSL certificate installed!"
    else
        echo "⚠️  SSL setup skipped (run manually if needed)"
    fi
    
    # Allow firewall
    if command -v ufw &> /dev/null && ufw status | grep -q "Status: active"; then
        ufw allow 'Nginx Full'
    fi
    
else
    echo ""
    echo "⚠️  API not responding yet. Check logs above for errors."
    echo ""
    echo "Common fixes:"
    echo "  • Wait 30 more seconds and test: curl http://localhost:8080/health"
    echo "  • Check logs: docker logs -f makwatches-be-api"
    echo "  • Check Portainer: http://139.59.71.95:9000"
fi

echo ""
echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║                                                                    ║"
echo "║   ✅ SETUP COMPLETE!                                              ║"
echo "║                                                                    ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo ""
echo "🎯 Your API is accessible at:"
echo ""
echo "  📍 Direct IP:  http://139.59.71.95:8080/health"
echo "  🌐 HTTP:       http://api.makwatches.in/health"
echo "  🔒 HTTPS:      https://api.makwatches.in/health"
echo ""
echo "🐳 Management:"
echo "  Portainer: http://139.59.71.95:9000"
echo ""
echo "📝 Test commands:"
echo "  curl http://localhost:8080/health"
echo "  curl http://api.makwatches.in/health"
echo "  curl https://api.makwatches.in/health"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
