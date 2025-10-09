# 🐳 Portainer Setup & Container Debugging

## Your Container is Restarting!

Your container status shows:
```
STATUS: Restarting (1) 7 seconds ago
```

This means the application is crashing. Let's fix it and set up Portainer for easier debugging.

---

## Part 1: Deploy Portainer (Docker Management UI)

### Quick Deploy

**On your server (139.59.71.95):**

```bash
# Create Portainer data volume
docker volume create portainer_data

# Deploy Portainer
docker run -d \
  --name portainer \
  --restart=always \
  -p 9000:9000 \
  -p 9443:9443 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v portainer_data:/data \
  portainer/portainer-ce:latest
```

### Access Portainer

1. **Open in browser:** http://139.59.71.95:9000
2. **First time setup:**
   - Create admin username
   - Create password (minimum 12 characters)
   - Click "Create user"
3. **Connect to local Docker:**
   - Select "Local" environment
   - Click "Connect"

### What You Can Do in Portainer

- 📊 View all containers, images, volumes, networks
- ▶️ Start/stop/restart containers
- 📝 View real-time logs
- 💻 Access container shell
- 📈 Monitor resource usage
- 🔧 Edit container configuration

---

## Part 2: Debug Your Restarting Container

### Method 1: Using Portainer (EASIEST)

1. Go to: http://139.59.71.95:9000
2. Click on "Local" environment
3. Go to "Containers"
4. Click on `makwatches-be-api`
5. Click "Logs" tab → See the error!
6. Click "Stats" → See resource usage
7. Click "Console" → Access shell if needed

### Method 2: Using Command Line

**Run this on your server:**

```bash
# See detailed logs
docker logs makwatches-be-api

# Follow logs in real-time
docker logs -f makwatches-be-api

# See last 50 lines
docker logs --tail 50 makwatches-be-api
```

### Method 3: Use Diagnostic Script

```bash
cd /opt/makwatches-be
chmod +x diagnose-container.sh
./diagnose-container.sh
```

---

## Common Issues & Solutions

### 1. Missing Environment Variables

**Check:**
```bash
cd /opt/makwatches-be
cat .env
```

**Required variables:**
- `MONGO_URI`
- `REDIS_URI`
- `JWT_SECRET`
- `FIREBASE_PROJECT_ID`
- `FIREBASE_BUCKET_NAME`
- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET`

**Fix:**
```bash
nano .env
# Add missing variables
# Save (Ctrl+X, Y, Enter)

# Restart container
docker compose restart
```

### 2. Firebase Credentials Invalid

**Check:**
```bash
cat /opt/makwatches-be/firebase-admin.json | jq .
```

If you get error, the JSON is invalid. Re-download from Firebase Console.

**Fix:**
```bash
# Re-upload firebase-admin.json
# Then restart
docker compose restart
```

### 3. MongoDB/Redis Connection Failed

**Test connections:**
```bash
# Test MongoDB
docker run --rm mongo:6 mongosh "YOUR_MONGO_URI" --eval "db.adminCommand('ping')"

# Test Redis
docker run --rm redis:7 redis-cli -u "YOUR_REDIS_URI" PING
```

### 4. Port Already in Use

**Check:**
```bash
netstat -tuln | grep 8080
```

**Fix:**
```bash
# Stop any service using port 8080
# Or change port in docker-compose.yml
nano docker-compose.yml
# Change: - "8080:8080" to - "8081:8080"
```

### 5. Application Code Error

**Check logs for specific error:**
```bash
docker logs makwatches-be-api 2>&1 | grep -i error
```

---

## Quick Fix Commands

```bash
# Stop everything
docker compose down

# Remove container
docker rm -f makwatches-be-api

# Pull latest image
docker pull aditya110/makwatches-be:latest

# Start fresh
docker compose up -d

# Check status
docker ps

# View logs
docker compose logs -f
```

---

## Debugging Workflow

1. **Deploy Portainer** (easier debugging)
   ```bash
   docker run -d --name portainer --restart=always \
     -p 9000:9000 -p 9443:9443 \
     -v /var/run/docker.sock:/var/run/docker.sock \
     -v portainer_data:/data \
     portainer/portainer-ce:latest
   ```

2. **Access Portainer:** http://139.59.71.95:9000

3. **View Container Logs** in Portainer UI

4. **Identify the Error**

5. **Fix the Issue**:
   - Missing env vars → Edit `.env`
   - Invalid Firebase → Re-upload `firebase-admin.json`
   - Connection error → Check MongoDB/Redis URIs
   - Code error → Check application logs

6. **Restart Container:**
   ```bash
   docker compose restart
   ```

7. **Verify:**
   ```bash
   curl http://139.59.71.95/health
   ```

---

## Next Steps

1. ✅ Deploy Portainer
2. 🔍 Check container logs (via Portainer or CLI)
3. 🐛 Identify specific error
4. 🔧 Fix the issue
5. ♻️ Restart container
6. ✅ Verify with health check

**Share the error logs and I'll help you fix it!**
