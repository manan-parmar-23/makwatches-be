# 🔧 Production Deployment Fix Guide

## Problem Identified

Your changes work locally and GitHub Actions shows successful builds, but changes don't appear in production.

## Root Causes

### 1. **Docker Image Caching**

- Server was pulling "latest" tag but Docker was serving cached images
- No force recreation of containers even when image was updated

### 2. **Missing Force Recreate Flag**

- `docker compose up -d` doesn't recreate containers by default if it thinks nothing changed
- Even with new image, same tag (`:latest`) makes Docker think nothing changed

## ✅ Solution Implemented

Updated `.github/workflows/docker-deploy.yml` with:

1. **Remove old image before pulling** - Forces Docker to fetch fresh image
2. **Use `--no-cache` flag** - Ensures no cached layers are used
3. **Add `--force-recreate`** - Forces container recreation even if config unchanged
4. **Add `--pull always`** - Always pulls image before starting
5. **Better logging** - Shows what's happening at each step
6. **Health check verification** - Confirms container is running and shows recent logs

## 🚀 How to Deploy Your Changes Now

### Method 1: Trigger Automatic Deployment (Recommended)

Simply push your changes to the main branch:

```bash
# Make sure all your changes are committed
git add .
git commit -m "fix: improve deployment process to prevent caching issues"
git push origin main
```

This will:

1. ✅ Trigger GitHub Actions
2. ✅ Build new Docker image
3. ✅ Push to Docker Hub
4. ✅ Deploy to server with force recreation
5. ✅ Verify deployment

### Method 2: Manual Force Deployment on Server

If you need to force redeploy the current latest image:

```bash
# SSH to your server
ssh root@139.59.71.95

# Navigate to project directory
cd /opt/makwatches-be

# Force pull and recreate
docker compose down
docker rmi $(docker images -q $(cat docker-compose.prod.yml | grep image: | awk '{print $2}')) || true
docker pull --no-cache <your-dockerhub-username>/makwatches-be:latest
docker compose up -d --force-recreate --pull always

# Verify
docker ps | grep makwatches-be-api
docker logs makwatches-be-api --tail=30
```

### Method 3: Test Locally First

Before pushing to production, test the Docker build locally:

```bash
# Build the image
docker build -t makwatches-be:test .

# Test it locally
docker run -p 8080:8080 --env-file .env makwatches-be:test

# Test the endpoint
curl http://localhost:8080/
# Should show: {"success":true,"message":"Welcome to Makwatches API's"}
```

## 🔍 Verify Your Changes Are Live

After deployment, verify your changes:

```bash
# Check the welcome endpoint
curl http://your-server-ip:8080/

# Expected output:
# {"success":true,"message":"Welcome to Makwatches API's"}
```

Or in your browser:

- Visit: `http://your-domain.com/` or `http://your-server-ip:8080/`

## 📊 Monitor Deployment

Watch the deployment in real-time:

1. **GitHub Actions**: https://github.com/manan-parmar-23/makwatches-be/actions
2. **Server Logs**:
   ```bash
   ssh root@139.59.71.95
   cd /opt/makwatches-be
   docker logs -f makwatches-be-api
   ```

## 🛠️ Troubleshooting

### Issue: Container still shows old version

**Solution:**

```bash
# On server
docker compose down
docker system prune -af  # WARNING: This removes ALL unused images
docker compose up -d --force-recreate
```

### Issue: GitHub Actions fails at deployment step

**Check:**

1. Server SSH credentials in GitHub Secrets
2. Server has enough disk space: `df -h`
3. Docker daemon is running: `systemctl status docker`

### Issue: Container starts but crashes immediately

**Check logs:**

```bash
docker logs makwatches-be-api --tail=100
```

Common causes:

- Missing environment variables
- Database connection issues
- Port already in use

## 🔐 Important Security Notes

### Check GitHub Secrets

Make sure these are set correctly:

- `SERVER_HOST` - Your server IP
- `SERVER_USER` - SSH username (usually `root`)
- `SERVER_SSH_KEY` - Private SSH key
- `SERVER_PORT` - SSH port (default: 22)
- `DOCKER_USERNAME` - Your Docker Hub username
- `DOCKER_PASSWORD` - Your Docker Hub password/token

### Docker Hub Username Issue

⚠️ **CRITICAL**: Based on `URGENT_FIX.md`, ensure your `DOCKER_USERNAME` secret matches your actual Docker Hub username.

Check: https://hub.docker.com/u/<your-username>

## 🎯 Best Practices Going Forward

1. **Always test locally** before pushing to main
2. **Use feature branches** for development
3. **Monitor deployment logs** after each push
4. **Keep environment variables in sync** between local and production
5. **Use semantic versioning** for release tags

## 📝 Quick Reference Commands

### Local Testing

```bash
# Build and test locally
docker build -t makwatches-be:test .
docker run -p 8080:8080 --env-file .env makwatches-be:test

# Test endpoint
curl http://localhost:8080/
```

### Production Deployment

```bash
# Push changes (automatic deployment)
git push origin main

# Manual force redeploy on server
ssh root@139.59.71.95 "cd /opt/makwatches-be && docker compose down && docker compose pull && docker compose up -d --force-recreate"
```

### Monitoring

```bash
# View logs
ssh root@139.59.71.95 "docker logs -f makwatches-be-api"

# Check container status
ssh root@139.59.71.95 "docker ps | grep makwatches"

# Check resource usage
ssh root@139.59.71.95 "docker stats makwatches-be-api --no-stream"
```

## ✨ Next Steps

1. Commit and push the updated workflow file
2. Monitor the GitHub Actions run
3. Verify the changes are live
4. Test all API endpoints
5. Check application logs for any errors

---

**Last Updated**: November 9, 2025
**Status**: ✅ Fixed - Force recreation and no-cache flags added
