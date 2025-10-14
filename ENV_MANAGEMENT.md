# Environment Variables Management Guide

## Problem
When pushing code changes, the CI/CD pipeline rebuilds the Docker image and redeploys to `/opt/makwatches-be`, but the environment variables in that directory don't automatically sync with your development environment in `/root/makwatches/makwatches-be`.

## Solution
Use the provided script to sync your environment variables from development to production before deployment.

## Directory Structure
- **Development**: `/root/makwatches/makwatches-be/` (your working directory)
- **Production**: `/opt/makwatches-be/` (where CI/CD deploys)

## How to Update Environment Variables

### Method 1: Using the Script (Recommended)
```bash
# 1. Edit your .env file in development directory
cd /root/makwatches/makwatches-be
nano .env

# 2. Run the update script
./update-production-env.sh

# 3. Redeploy the application
make deploy-prod
```

### Method 2: Using Make Commands
```bash
# Update production environment and deploy in one go
make update-prod-env deploy-prod
```

### Method 3: Manual Update
```bash
# Copy .env to production directory
sudo cp /root/makwatches/makwatches-be/.env /opt/makwatches-be/.env

# Copy docker-compose file
sudo cp /root/makwatches/makwatches-be/docker-compose.prod.yml /opt/makwatches-be/docker-compose.yml

# Redeploy
cd /opt/makwatches-be
docker compose pull
docker compose down
docker compose up -d
```

## Environment Variables Structure

### Database Configuration
```env
MONGO_URI=your_mongodb_connection_string
DATABASE_NAME=makwatches
```

### Redis Configuration
```env
REDIS_URI=redis://your-redis-host:port
REDIS_PASSWORD=your_redis_password
REDIS_DATABASE_NAME=your_redis_database_name
```

### Authentication
```env
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRY=24h
```

### Google OAuth
```env
GOOGLE_CLIENT_ID=your_google_client_id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URL=https://yourdomain.com/auth/google/callback
```

### Payment Gateway (Razorpay)
```env
RAZORPAY_MODE=live
RAZORPAY_KEY_ID=your_razorpay_key_id
RAZORPAY_KEY_SECRET=your_razorpay_key_secret
```

### Firebase
```env
FIREBASE_PROJECT_ID=your_firebase_project_id
FIREBASE_BUCKET_NAME=your_firebase_bucket_name
FIREBASE_CREDENTIALS_PATH=firebase-admin.json
```

### Docker
```env
DOCKER_USERNAME=your_docker_username
```

> **Note**: Replace all placeholder values with your actual credentials. Never commit real credentials to Git.

## Important Notes

1. **Never commit `.env` to Git**: The `.env` file contains sensitive credentials and should never be committed to version control.

2. **Backup**: The update script automatically creates backups of your production `.env` file with timestamps.

3. **Production vs Development**: 
   - Development directory: `/root/makwatches/makwatches-be/`
   - Production directory: `/opt/makwatches-be/`
   - CI/CD deploys to production directory

4. **After CI/CD Deployment**: 
   - CI/CD builds and pushes Docker image
   - CI/CD pulls the image and restarts containers in `/opt/makwatches-be`
   - Environment variables are read from `/opt/makwatches-be/.env`
   - Make sure to sync your `.env` changes before or after CI/CD runs

## Verification

### Check Production Environment
```bash
# View production .env
sudo cat /opt/makwatches-be/.env

# Check running container
docker ps | grep makwatches

# View container logs
docker logs makwatches-be-api --tail=50

# Check container environment
docker exec makwatches-be-api env | grep -E "REDIS|MONGO|JWT"
```

### Test Application
```bash
# Health check
curl http://localhost:8080/health

# Check Redis connection (look for Redis-related logs)
docker logs makwatches-be-api 2>&1 | grep -i redis
```

## Troubleshooting

### Environment variables not updating
```bash
# Ensure production .env is updated
sudo cat /opt/makwatches-be/.env | head -5

# Restart containers to pick up changes
cd /opt/makwatches-be
docker compose restart

# Or do a full redeploy
docker compose down && docker compose up -d
```

### Redis connection issues
```bash
# Check Redis connectivity
docker exec makwatches-be-api nc -zv redis-14568.c301.ap-south-1-1.ec2.redns.redis-cloud.com 14568

# View Redis-related logs
docker logs makwatches-be-api 2>&1 | grep -i redis
```

### Container not starting
```bash
# Check container status
docker ps -a | grep makwatches

# View full logs
docker logs makwatches-be-api

# Check for configuration errors
docker inspect makwatches-be-api | grep -A 10 "Env"
```

## Best Practices

1. **Always test in development first**: Make changes in `/root/makwatches/makwatches-be/.env` and test locally before syncing to production.

2. **Use the update script**: The `update-production-env.sh` script ensures proper backup and sync.

3. **Verify after deployment**: Always check logs and test the application after updating environment variables.

4. **Document changes**: Keep track of what environment variables were changed and why.

5. **Secure credentials**: Never share `.env` files or commit them to version control.

## Quick Reference

```bash
# Update production environment variables
make update-prod-env

# Deploy to production
make deploy-prod

# Both in one command
make update-prod-env deploy-prod

# Check production status
docker ps | grep makwatches
docker logs makwatches-be-api --tail=30
```
