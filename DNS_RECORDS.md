# 🌐 DNS Records for api.makwatches.in (Vercel Nameservers)

Since you're using Vercel nameservers for `makwatches.in`, you need to add DNS records in your domain registrar or Vercel DNS settings.

---

## DNS Record to Add

Add this **A Record** to point `api.makwatches.in` to your server:

```
Type:    A
Name:    api
Value:   139.59.71.95
TTL:     Auto (or 3600)
Proxy:   Off (if using Cloudflare)
```

---

## Where to Add These Records

### Option 1: If Domain Registered on Vercel

1. Go to: https://vercel.com/dashboard
2. Click on your project or domains section
3. Select `makwatches.in`
4. Go to DNS/Settings
5. Add the A record above

### Option 2: If Using External Domain Registrar (GoDaddy, Namecheap, etc.)

Since you're using Vercel nameservers, you have two options:

#### Option A: Add in Domain Registrar's DNS (Before Vercel Nameservers)

1. Log in to your domain registrar
2. Go to DNS Management
3. Add the A record
4. Wait 5-10 minutes for propagation

#### Option B: Configure via Vercel DNS

1. Go to Vercel dashboard
2. Add domain to project
3. Configure DNS records there

---

## Exact DNS Configuration

### For api.makwatches.in → Your Backend Server

| Type | Name | Value | TTL |
|------|------|-------|-----|
| A | api | 139.59.71.95 | 3600 |

This will make `api.makwatches.in` point to `139.59.71.95`

---

## Verify DNS Propagation

After adding the record, wait 5-10 minutes and verify:

### Online Tools:
- https://dnschecker.org/#A/api.makwatches.in
- https://www.whatsmydns.net/#A/api.makwatches.in

### Command Line:
```bash
# On your local machine
nslookup api.makwatches.in

# Should return:
# Name: api.makwatches.in
# Address: 139.59.71.95
```

Or use dig:
```bash
dig api.makwatches.in

# Should show A record pointing to 139.59.71.95
```

---

## Common DNS Configurations

### If You Want Multiple Subdomains:

| Type | Name | Value | Purpose |
|------|------|-------|---------|
| A | api | 139.59.71.95 | Backend API |
| CNAME | www | cname.vercel-dns.com | Frontend (Vercel) |
| CNAME | admin | cname.vercel-dns.com | Admin Panel (Vercel) |

---

## After DNS is Set Up

Once DNS propagation is complete (you can verify with nslookup), run on your server:

```bash
cd /opt/makwatches-be
chmod +x setup-domain.sh
./setup-domain.sh
```

This will:
1. Install Nginx
2. Configure reverse proxy
3. Install SSL certificate (Let's Encrypt)
4. Enable HTTPS

---

## Testing

### Before SSL (HTTP):
```bash
curl http://api.makwatches.in/health
```

### After SSL (HTTPS):
```bash
curl https://api.makwatches.in/health
```

Should return: `{"status":"healthy"}`

---

## Troubleshooting

### DNS Not Resolving

**Check nameservers:**
```bash
nslookup -type=ns makwatches.in
```

If using Vercel nameservers, they should be something like:
- ns1.vercel-dns.com
- ns2.vercel-dns.com

**Solution:** Make sure you added the A record in the correct place (where nameservers are managed)

### Still Can't Access

1. **Check DNS:** `nslookup api.makwatches.in` should return `139.59.71.95`
2. **Check server firewall:** 
   ```bash
   ufw status
   ufw allow 80/tcp
   ufw allow 443/tcp
   ```
3. **Check Nginx:** `systemctl status nginx`

---

## Complete Setup Flow

```
Step 1: Add DNS A Record (api → 139.59.71.95)
   ↓
Step 2: Wait 5-10 minutes for DNS propagation
   ↓
Step 3: Verify DNS resolves (nslookup api.makwatches.in)
   ↓
Step 4: Run setup-domain.sh on server
   ↓
Step 5: Access https://api.makwatches.in/health
```

---

## Screenshot Guide

If you need visual help:

1. **GoDaddy:** DNS Management → Add → Type: A, Name: api, Value: 139.59.71.95
2. **Namecheap:** Advanced DNS → Add New Record → A Record
3. **Cloudflare:** DNS → Add record → Type: A, Name: api, IPv4: 139.59.71.95, Proxy: Off
4. **Vercel:** Domains → makwatches.in → DNS Records → Add

---

## Summary

**Single DNS Record Needed:**
```
Type: A
Name: api
Value: 139.59.71.95
```

That's it! Once this propagates, run the domain setup script and you'll have:
- ✅ https://api.makwatches.in
- ✅ SSL certificate
- ✅ Professional setup

---

## Questions?

Share:
1. Screenshot of your DNS settings
2. Output of: `nslookup api.makwatches.in`
3. Where your domain is registered
