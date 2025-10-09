# 🌐 Domain Setup: api.makwatches.in

This guide will help you setup your domain `api.makwatches.in` with SSL certificate.

---

## Prerequisites

✅ Container must be running successfully on port 8080
✅ Domain DNS must point to your server IP (139.59.71.95)

---

## Part 1: DNS Configuration

### Option 1: If using Cloudflare

1. Go to Cloudflare Dashboard
2. Select your domain: `makwatches.in`
3. Go to DNS → Records
4. Add A record:
   - **Type:** A
   - **Name:** api
   - **IPv4 address:** 139.59.71.95
   - **Proxy status:** DNS only (Grey cloud) ⚠️ Important for SSL setup
   - **TTL:** Auto
5. Save

### Option 2: If using other DNS provider

Add an A record:
```
Type: A
Host: api
Value: 139.59.71.95
TTL: 3600
```

### Verify DNS

Wait 5-10 minutes, then test:
```bash
# On your local machine
nslookup api.makwatches.in

# Should return: 139.59.71.95
```

Or use: https://dnschecker.org/#A/api.makwatches.in

---

## Part 2: Install Nginx (Reverse Proxy)

Run on your server (139.59.71.95):

```bash
# Update package list
apt update

# Install Nginx
apt install -y nginx

# Check status
systemctl status nginx

# Enable on boot
systemctl enable nginx
```

---

## Part 3: Configure Nginx for Your API

### Create Nginx configuration

```bash
# Create config file
nano /etc/nginx/sites-available/api.makwatches.in
```

Paste this configuration:

```nginx
# Redirect HTTP to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name api.makwatches.in;
    
    # For Let's Encrypt validation
    location /.well-known/acme-challenge/ {
        root /var/www/html;
    }
    
    # Redirect all other traffic to HTTPS
    location / {
        return 301 https://$server_name$request_uri;
    }
}

# HTTPS Server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name api.makwatches.in;
    
    # SSL certificates (will be added by Certbot)
    # ssl_certificate /etc/letsencrypt/live/api.makwatches.in/fullchain.pem;
    # ssl_certificate_key /etc/letsencrypt/live/api.makwatches.in/privkey.pem;
    
    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    
    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    
    # Proxy settings
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
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    # Access and error logs
    access_log /var/log/nginx/api.makwatches.in.access.log;
    error_log /var/log/nginx/api.makwatches.in.error.log;
}
```

Save (Ctrl+X, Y, Enter)

### Enable the site

```bash
# Create symbolic link
ln -s /etc/nginx/sites-available/api.makwatches.in /etc/nginx/sites-enabled/

# Remove default site
rm /etc/nginx/sites-enabled/default

# Test configuration
nginx -t

# Reload Nginx
systemctl reload nginx
```

---

## Part 4: Install SSL Certificate (Let's Encrypt)

### Install Certbot

```bash
# Install Certbot
apt install -y certbot python3-certbot-nginx

# Get SSL certificate
certbot --nginx -d api.makwatches.in
```

### During Certbot setup:

1. Enter your email address
2. Agree to Terms of Service (Y)
3. Share email with EFF (your choice)
4. Certbot will automatically configure SSL

### Auto-renewal

Certbot sets up auto-renewal automatically. Test it:

```bash
# Test renewal
certbot renew --dry-run
```

---

## Part 5: Update Firewall (if UFW is enabled)

```bash
# Allow HTTP and HTTPS
ufw allow 'Nginx Full'

# Check status
ufw status
```

---

## Part 6: Verify Setup

### Test HTTP to HTTPS redirect

```bash
curl -I http://api.makwatches.in
# Should return: 301 Moved Permanently
# Location: https://api.makwatches.in/
```

### Test HTTPS

```bash
curl https://api.makwatches.in/health
# Should return: {"status":"healthy"}
```

### Test from browser

Open: https://api.makwatches.in/health

---

## Part 7: Update Application Configuration

Update your frontend and any services to use:

**Old:** `http://139.59.71.95:8080`
**New:** `https://api.makwatches.in`

### Update environment variables

If you have CORS origins in your `.env`:

```bash
nano /opt/makwatches-be/.env
```

Ensure VERCEL_ORIGIN and other origins are set:
```env
VERCEL_ORIGIN=https://mak-watches.vercel.app
DEV_ORIGIN=http://localhost:4200
```

Restart container:
```bash
cd /opt/makwatches-be
docker compose restart
```

---

## Troubleshooting

### 1. 502 Bad Gateway

**Cause:** Backend not running
**Fix:**
```bash
docker ps | grep makwatches-be
# If not running, check logs:
docker logs makwatches-be-api
```

### 2. SSL Certificate Error

**Cause:** DNS not propagated or Certbot failed
**Fix:**
```bash
# Check DNS
nslookup api.makwatches.in

# Re-run Certbot
certbot --nginx -d api.makwatches.in --force-renewal
```

### 3. Connection Refused

**Cause:** Nginx not running or port blocked
**Fix:**
```bash
# Check Nginx
systemctl status nginx

# Check if port 80/443 are open
netstat -tuln | grep -E ':80|:443'

# Restart Nginx
systemctl restart nginx
```

### 4. CORS Errors

**Cause:** Domain not in allowed origins
**Fix:** Update `.env` with correct VERCEL_ORIGIN and restart container

---

## Nginx Useful Commands

```bash
# Test configuration
nginx -t

# Reload (no downtime)
systemctl reload nginx

# Restart
systemctl restart nginx

# Check status
systemctl status nginx

# View logs
tail -f /var/log/nginx/api.makwatches.in.access.log
tail -f /var/log/nginx/api.makwatches.in.error.log

# View error log
tail -f /var/log/nginx/error.log
```

---

## Security Best Practices

### 1. Enable Fail2Ban (Optional)

Protect against brute force attacks:
```bash
apt install -y fail2ban
systemctl enable fail2ban
systemctl start fail2ban
```

### 2. Regular Updates

```bash
# Update system
apt update && apt upgrade -y

# Update SSL certificates (auto, but can force)
certbot renew
```

### 3. Monitor Logs

```bash
# Check access patterns
tail -f /var/log/nginx/api.makwatches.in.access.log

# Check for errors
tail -f /var/log/nginx/api.makwatches.in.error.log
```

---

## Expected Result

After setup:

✅ HTTP redirects to HTTPS automatically
✅ SSL certificate (A+ rating)
✅ Clean URL: https://api.makwatches.in
✅ No need to specify port
✅ Professional setup

### API Endpoints

- Health: https://api.makwatches.in/health
- Auth: https://api.makwatches.in/api/v1/auth/login
- Products: https://api.makwatches.in/api/v1/products
- etc.

---

## Automated Setup Script

Want to automate this? Run:

```bash
cd /opt/makwatches-be
chmod +x setup-domain.sh
./setup-domain.sh
```

(I'll create this script if you want!)

---

## Need Help?

Share:
1. Output of: `nginx -t`
2. Output of: `curl -I https://api.makwatches.in`
3. Nginx error logs: `tail -50 /var/log/nginx/error.log`
