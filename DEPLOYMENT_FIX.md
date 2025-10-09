# 🔧 Deployment Fix - Docker Username Mismatch

## Problem
The deployment is failing because of WRONG Docker Hub username:
```
❌ Trying to pull: adityagarg646/makwatches-be (WRONG - doesn't exist)
✅ Should pull: aditya110/makwatches-be (CORRECT - exists at https://hub.docker.com/repository/docker/aditya110/makwatches-be)
```

Your Docker Hub repository exists at `aditya110/makwatches-be`, but the configuration is using `adityagarg646`.

## Quick Fix

### Automated Fix (RECOMMENDED)

Run the automated fix script:

```bash
cd /root/makwatches/makwatches-be
chmod +x fix-docker-username.sh
./fix-docker-username.sh
```

This script will:
1. Update `.env` on your server to use `aditya110`
2. Stop current containers
3. Pull the correct image from `aditya110/makwatches-be`
4. Restart services

### Manual Fix

**Step 1: Update GitHub Secret**
1. Go to: https://github.com/manan-parmar-23/makwatches-be/settings/secrets/actions
2. Click on `DOCKER_USERNAME`
3. Update value from `adityagarg646` to: **`aditya110`**
4. Save

**Step 2: Update Server Configuration**
SSH to your server and update `.env`:

```bash
ssh root@139.59.71.95
cd /opt/makwatches-be
nano .env
```

Change this line:
```env
# FROM:
DOCKER_USERNAME=adityagarg646

# TO:
DOCKER_USERNAME=aditya110
```

Save and exit (Ctrl+X, Y, Enter)

**Step 3: Deploy with Correct Image**

```bash
# Stop current containers
docker compose down

# Remove wrong image
docker rmi adityagarg646/makwatches-be:latest

# Pull correct image
docker pull aditya110/makwatches-be:latest

# Start services
docker compose up -d

# Check status
docker ps
```

**Step 5: Remove version from docker-compose.yml**
The warning says "version is obsolete" - let's remove it.

**Step 6: Trigger New Deployment**
```bash
git commit --allow-empty -m "Trigger deployment after Docker Hub setup"
git push origin main
```

---

### Option 2: Use Different Docker Registry

If you don't want to use Docker Hub, you can use GitHub Container Registry:

1. Update `.github/workflows/docker-deploy.yml` to use `ghcr.io`
2. Update `docker-compose.prod.yml` to pull from `ghcr.io`

---

## Quick Fix Command

**On your server**, run this to update the configuration:

```bash
cd /opt/makwatches-be

# Update .env with correct DOCKER_USERNAME
echo "DOCKER_USERNAME=YOUR_ACTUAL_DOCKERHUB_USERNAME" >> .env.tmp
grep -v "DOCKER_USERNAME" .env >> .env.tmp
mv .env.tmp .env

# Remove old images
docker image prune -af

# Try pulling manually to verify
docker pull YOUR_DOCKERHUB_USERNAME/makwatches-be:latest
```

---

## Verification

After creating the repository, verify:

1. **Docker Hub**: Check repository exists at `https://hub.docker.com/r/YOUR_USERNAME/makwatches-be`
2. **GitHub Secrets**: Verify DOCKER_USERNAME at https://github.com/manan-parmar-23/makwatches-be/settings/secrets/actions
3. **Server .env**: SSH to server and check `cat /opt/makwatches-be/.env | grep DOCKER_USERNAME`

---

## What Happens Next

Once you create the Docker Hub repository:
1. GitHub Actions will build your Docker image
2. Push it to Docker Hub under your username
3. Server will pull the image successfully
4. Application will start running

---

## Need Help?

If you're still stuck, provide:
1. Your Docker Hub username
2. Screenshot of Docker Hub repositories page
3. GitHub Actions logs from latest run
