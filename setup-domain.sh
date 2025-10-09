#!/bin/bash

# Automated Domain Setup Script for api.makwatches.in
# This script sets up Nginx reverse proxy with SSL

set -e  # Exit on error

DOMAIN="api.makwatches.in"
SERVER_IP="139.59.71.95"
BACKEND_PORT="8080"
EMAIL="adityagarg646@gmail.com"  # Change this to your email

echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║                                                                    ║"
echo "║   🌐 Domain Setup: $DOMAIN                         ║"
echo "║                                                                    ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "❌ Please run as root: sudo ./setup-domain.sh"
    exit 1
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1️⃣  Checking Prerequisites"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check if backend is running
if ! docker ps | grep -q makwatches-be-api; then
    echo "⚠️  WARNING: Backend container is not running!"
    echo "   Please fix the container first before setting up domain."
    read -p "   Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Check if backend responds
if curl -f -s http://localhost:$BACKEND_PORT/health > /dev/null 2>&1; then
    echo "✅ Backend is responding on port $BACKEND_PORT"
else
    echo "⚠️  WARNING: Backend not responding on port $BACKEND_PORT"
    read -p "   Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Check DNS
echo ""
echo "Checking DNS for $DOMAIN..."
if nslookup $DOMAIN | grep -q "$SERVER_IP"; then
    echo "✅ DNS is configured correctly"
else
    echo "⚠️  WARNING: DNS may not be configured yet"
    echo "   Make sure to add A record: $DOMAIN → $SERVER_IP"
    read -p "   Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2️⃣  Installing Nginx"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if command -v nginx &> /dev/null; then
    echo "✅ Nginx already installed"
else
    echo "Installing Nginx..."
    apt update -qq
    apt install -y nginx
    systemctl enable nginx
    echo "✅ Nginx installed"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3️⃣  Configuring Nginx"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Backup existing config if exists
if [ -f /etc/nginx/sites-available/$DOMAIN ]; then
    cp /etc/nginx/sites-available/$DOMAIN /etc/nginx/sites-available/$DOMAIN.backup.$(date +%s)
fi

# Create Nginx config
cat > /etc/nginx/sites-available/$DOMAIN << 'NGINX_EOF'
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
ln -sf /etc/nginx/sites-available/$DOMAIN /etc/nginx/sites-enabled/

# Remove default site if exists
if [ -f /etc/nginx/sites-enabled/default ]; then
    rm /etc/nginx/sites-enabled/default
fi

# Test configuration
if nginx -t; then
    echo "✅ Nginx configuration is valid"
    systemctl reload nginx
    echo "✅ Nginx reloaded"
else
    echo "❌ Nginx configuration error!"
    exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4️⃣  Installing SSL Certificate"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if command -v certbot &> /dev/null; then
    echo "✅ Certbot already installed"
else
    echo "Installing Certbot..."
    apt install -y certbot python3-certbot-nginx
    echo "✅ Certbot installed"
fi

echo ""
echo "Getting SSL certificate..."
echo "⚠️  Make sure DNS is propagated before continuing!"
read -p "Continue with SSL setup? (Y/n): " -n 1 -r
echo

if [[ $REPLY =~ ^[Nn]$ ]]; then
    echo "⏩ Skipping SSL setup. You can run it later:"
    echo "   certbot --nginx -d $DOMAIN"
else
    if certbot --nginx -d $DOMAIN --non-interactive --agree-tos --email $EMAIL --redirect; then
        echo "✅ SSL certificate installed successfully"
    else
        echo "⚠️  SSL certificate installation failed"
        echo "   You can try again manually: certbot --nginx -d $DOMAIN"
    fi
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "5️⃣  Configuring Firewall"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if command -v ufw &> /dev/null && ufw status | grep -q "Status: active"; then
    echo "Configuring UFW..."
    ufw allow 'Nginx Full'
    echo "✅ Firewall configured"
else
    echo "ℹ️  UFW not active, skipping"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "6️⃣  Testing Setup"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo ""
echo "Testing HTTP..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://$DOMAIN/health || echo "000")
if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "301" ]; then
    echo "✅ HTTP working (Status: $HTTP_CODE)"
else
    echo "⚠️  HTTP test returned: $HTTP_CODE"
fi

if [ -d "/etc/letsencrypt/live/$DOMAIN" ]; then
    echo ""
    echo "Testing HTTPS..."
    HTTPS_CODE=$(curl -s -o /dev/null -w "%{http_code}" https://$DOMAIN/health || echo "000")
    if [ "$HTTPS_CODE" = "200" ]; then
        echo "✅ HTTPS working (Status: $HTTPS_CODE)"
    else
        echo "⚠️  HTTPS test returned: $HTTPS_CODE"
    fi
fi

echo ""
echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║                                                                    ║"
echo "║   ✅ SETUP COMPLETE!                                              ║"
echo "║                                                                    ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo ""
echo "Your API is now accessible at:"
echo ""
if [ -d "/etc/letsencrypt/live/$DOMAIN" ]; then
    echo "  🔒 https://$DOMAIN"
    echo "  🔒 https://$DOMAIN/health"
else
    echo "  🌐 http://$DOMAIN"
    echo "  🌐 http://$DOMAIN/health"
    echo ""
    echo "  ⚠️  SSL not configured. Run: certbot --nginx -d $DOMAIN"
fi
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Useful commands:"
echo "  • View logs: tail -f /var/log/nginx/$DOMAIN.access.log"
echo "  • Reload Nginx: systemctl reload nginx"
echo "  • Test SSL: curl https://$DOMAIN/health"
echo "  • Renew SSL: certbot renew"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
