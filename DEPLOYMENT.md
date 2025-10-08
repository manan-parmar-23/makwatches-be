# MakWatches Backend - Docker Deployment Guide

This guide provides comprehensive instructions for deploying the MakWatches backend using Docker and setting up CI/CD pipelines.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Local Development with Docker](#local-development-with-docker)
3. [Production Deployment](#production-deployment)
4. [CI/CD Setup](#cicd-setup)
5. [Environment Variables](#environment-variables)
6. [Monitoring and Logs](#monitoring-and-logs)
7. [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Software

- Docker (20.10 or higher)
- Docker Compose (2.0 or higher)
- Git

### Installation

#### Ubuntu/Debian

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo apt-get update
sudo apt-get install docker-compose-plugin

# Add your user to docker group (optional, to run without sudo)
sudo usermod -aG docker $USER
```

#### Verify Installation

```bash
docker --version
docker-compose --version
```

## Local Development with Docker

### 1. Clone the Repository

```bash
git clone https://github.com/manan-parmar-23/makwatches-be.git
cd makwatches-be
```

### 2. Setup Environment Variables

```bash
# Copy example.env to .env
cp example.env .env

# Edit .env with your configuration
nano .env
```

### 3. Add Firebase Credentials

```bash
# Place your firebase-admin.json file in the root directory
# Make sure it's in .gitignore (already configured)
cp /path/to/your/firebase-admin.json .
```

### 4. Build and Run with Docker Compose

```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Check status
docker-compose ps
```

### 5. Access the Application

- API: http://localhost:8080
- Health Check: http://localhost:8080/health

### 6. Development Commands

```bash
# Stop services
docker-compose down

# Rebuild after code changes
docker-compose up -d --build

# View logs for specific service
docker-compose logs -f api

# Execute commands in container
docker-compose exec api sh

# Restart services
docker-compose restart
```

## Production Deployment

### Option 1: Using Deployment Script (Recommended)

The project includes a deployment script that automates the entire process.

```bash
# Make script executable (if not already)
chmod +x deploy.sh

# Deploy application
./deploy.sh deploy

# Other commands
./deploy.sh build          # Build Docker image
./deploy.sh build-deploy   # Build and deploy
./deploy.sh stop           # Stop application
./deploy.sh restart        # Restart application
./deploy.sh status         # Show status
./deploy.sh logs           # Show logs
./deploy.sh cleanup        # Clean up old images
```

### Option 2: Manual Deployment

#### Step 1: Prepare Server

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker and Docker Compose (see Prerequisites)

# Create application directory
sudo mkdir -p /opt/makwatches-be
cd /opt/makwatches-be
```

#### Step 2: Clone Repository

```bash
git clone https://github.com/manan-parmar-23/makwatches-be.git .
```

#### Step 3: Configure Environment

```bash
# Create .env file
cp example.env .env
nano .env  # Edit with production values

# Add Firebase credentials
nano firebase-admin.json  # Paste your Firebase admin JSON
```

#### Step 4: Deploy

```bash
# Using production docker-compose file
docker-compose -f docker-compose.prod.yml up -d

# Monitor deployment
docker-compose -f docker-compose.prod.yml logs -f
```

### Option 3: Using Pre-built Docker Image

If you're using CI/CD to build images:

```bash
# Pull latest image
docker pull yourusername/makwatches-be:latest

# Run with docker-compose
docker-compose -f docker-compose.prod.yml up -d
```

## CI/CD Setup

### GitHub Actions

The project includes a GitHub Actions workflow for automated builds and deployments.

#### 1. Setup Docker Hub Repository

1. Create account on [Docker Hub](https://hub.docker.com/)
2. Create a repository named `makwatches-be`

#### 2. Configure GitHub Secrets

Go to your GitHub repository → Settings → Secrets → Actions, and add:

| Secret Name | Description | Example |
|------------|-------------|---------|
| `DOCKER_USERNAME` | Docker Hub username | `yourusername` |
| `DOCKER_PASSWORD` | Docker Hub password/token | `dckr_pat_xxx...` |
| `SERVER_HOST` | Production server IP/domain | `123.45.67.89` |
| `SERVER_USER` | SSH username | `root` or `ubuntu` |
| `SERVER_SSH_KEY` | Private SSH key | `-----BEGIN RSA PRIVATE KEY-----...` |
| `SERVER_PORT` | SSH port (optional) | `22` |
| `APP_URL` | Application URL | `https://api.makwatches.com` |

#### 3. Setup Server for Deployment

On your production server:

```bash
# Create deployment directory
sudo mkdir -p /opt/makwatches-be
cd /opt/makwatches-be

# Create .env file with production values
sudo nano .env

# Add firebase-admin.json
sudo nano firebase-admin.json

# Create docker-compose.prod.yml
sudo nano docker-compose.prod.yml
# Paste the contents from the repository
```

#### 4. Workflow Triggers

The CI/CD pipeline automatically runs on:
- Push to `main` branch (builds, tests, deploys)
- Push to `develop` branch (builds, tests)
- Pull requests to `main` (tests only)
- Manual trigger (workflow_dispatch)

### Alternative: GitLab CI/CD

Create `.gitlab-ci.yml`:

```yaml
stages:
  - test
  - build
  - deploy

variables:
  IMAGE_NAME: $CI_REGISTRY_IMAGE:$CI_COMMIT_SHORT_SHA
  LATEST_IMAGE: $CI_REGISTRY_IMAGE:latest

test:
  stage: test
  image: golang:1.24-alpine
  script:
    - go mod download
    - go test -v ./...
    - go vet ./...

build:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  before_script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
  script:
    - docker build -t $IMAGE_NAME -t $LATEST_IMAGE .
    - docker push $IMAGE_NAME
    - docker push $LATEST_IMAGE
  only:
    - main
    - develop

deploy:
  stage: deploy
  image: alpine:latest
  before_script:
    - apk add --no-cache openssh-client
    - eval $(ssh-agent -s)
    - echo "$SSH_PRIVATE_KEY" | tr -d '\r' | ssh-add -
    - mkdir -p ~/.ssh
    - chmod 700 ~/.ssh
  script:
    - ssh -o StrictHostKeyChecking=no $SERVER_USER@$SERVER_HOST "
        cd /opt/makwatches-be &&
        docker pull $LATEST_IMAGE &&
        docker-compose -f docker-compose.prod.yml down &&
        docker-compose -f docker-compose.prod.yml up -d &&
        docker image prune -af"
  only:
    - main
```

## Environment Variables

### Required Variables

```env
# Server
PORT=8080
ENVIRONMENT=production

# Database
MONGO_URI=mongodb+srv://username:password@cluster.mongodb.net/makwatches
DATABASE_NAME=makwatches

# Redis Cache
REDIS_URI=redis-host:port
REDIS_PASSWORD=your_redis_password
REDIS_DATABASE_NAME=database_name

# Authentication
JWT_SECRET=your_super_secret_jwt_key_change_this
JWT_EXPIRY=24h

# Google OAuth
GOOGLE_CLIENT_ID=your_client_id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your_client_secret
GOOGLE_REDIRECT_URL=https://yourapp.com/auth/google/callback

# Payment Gateway
RAZORPAY_MODE=live  # or 'test' for testing
RAZORPAY_KEY_ID_TEST=rzp_test_xxxxx
RAZORPAY_KEY_SECRET_TEST=your_secret

# Firebase Storage
FIREBASE_PROJECT_ID=your-project-id
FIREBASE_BUCKET_NAME=your-project.appspot.com
FIREBASE_CREDENTIALS_PATH=firebase-admin.json

# CORS Origins
VERCEL_ORIGIN=https://yourapp.vercel.app
DEV_ORIGIN=http://localhost:4200
```

### Docker Compose Variables

For `docker-compose.prod.yml`, also set:

```env
DOCKER_USERNAME=your_dockerhub_username
```

## Monitoring and Logs

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f api

# Last 100 lines
docker-compose logs --tail=100 api

# With timestamps
docker-compose logs -f -t api
```

### Check Container Status

```bash
# List running containers
docker-compose ps

# Detailed container info
docker inspect makwatches-be-api

# Resource usage
docker stats makwatches-be-api
```

### Health Checks

```bash
# Manual health check
curl http://localhost:8080/health

# Check from within container
docker exec makwatches-be-api curl -f http://localhost:8080/health
```

### Log Management

Logs are automatically rotated with the following settings:
- Maximum size: 10MB per file
- Maximum files: 3
- Total log size: ~30MB

Configure in `docker-compose.prod.yml`:

```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

## Troubleshooting

### Common Issues

#### 1. Container Won't Start

```bash
# Check logs
docker-compose logs api

# Common causes:
# - Missing .env file
# - Invalid environment variables
# - Firebase credentials not found
# - Port already in use
```

#### 2. Database Connection Error

```bash
# Verify MongoDB connection
docker-compose exec api sh
# Inside container:
ping mongodb  # If using local MongoDB
# or test the connection string

# Check MongoDB Atlas network access
# - Add server IP to whitelist
# - Verify credentials
```

#### 3. Redis Connection Error

```bash
# Check Redis connectivity
docker-compose exec api sh
# Inside container:
redis-cli -h redis -p 6379 ping  # Local Redis
# or
redis-cli -h your-redis-host -p port -a password ping  # Cloud Redis
```

#### 4. Firebase Storage Upload Fails

```bash
# Verify firebase-admin.json exists
ls -la firebase-admin.json

# Check permissions
docker-compose exec api cat /app/firebase-admin.json

# Verify Firebase Storage is enabled in Firebase Console
# Check service account has Storage Admin role
```

#### 5. Port Already in Use

```bash
# Find process using port 8080
sudo lsof -i :8080

# Kill the process
sudo kill -9 <PID>

# Or change port in .env
PORT=8081
```

#### 6. Build Fails

```bash
# Clean Docker build cache
docker builder prune -af

# Rebuild without cache
docker-compose build --no-cache

# Check Go version compatibility
docker run golang:1.24-alpine go version
```

### Debugging Commands

```bash
# Enter container shell
docker-compose exec api sh

# Check environment variables
docker-compose exec api env

# Test API endpoints
docker-compose exec api wget -O- http://localhost:8080/health

# View container processes
docker-compose top

# Inspect container
docker inspect makwatches-be-api

# Check network connectivity
docker network ls
docker network inspect makwatches-network
```

### Recovery Procedures

#### Complete Reset

```bash
# Stop all containers
docker-compose down

# Remove volumes (WARNING: deletes data)
docker-compose down -v

# Remove images
docker rmi makwatches-be:latest

# Rebuild and restart
docker-compose up -d --build
```

#### Rollback Deployment

```bash
# Pull specific version
docker pull yourusername/makwatches-be:previous-tag

# Update docker-compose to use specific tag
# Edit docker-compose.prod.yml image tag

# Restart
docker-compose -f docker-compose.prod.yml up -d
```

## Production Best Practices

### Security

1. **Use secrets management**
   - Never commit `.env` or `firebase-admin.json`
   - Use Docker secrets or vault services
   - Rotate credentials regularly

2. **Network security**
   - Use firewall rules
   - Limit exposed ports
   - Use reverse proxy (nginx/traefik)

3. **Container security**
   - Run as non-root user (already configured)
   - Keep images updated
   - Scan for vulnerabilities

### Performance

1. **Resource limits**
   
   Add to `docker-compose.prod.yml`:
   ```yaml
   deploy:
     resources:
       limits:
         cpus: '2'
         memory: 2G
       reservations:
         cpus: '1'
         memory: 1G
   ```

2. **Enable caching**
   - Ensure Redis is properly configured
   - Monitor cache hit rates

### Backup

1. **MongoDB backups**
   ```bash
   # Backup MongoDB (if using cloud, use their backup tools)
   docker exec makwatches-mongodb mongodump --out /backup
   ```

2. **Uploads backup**
   ```bash
   # Backup uploads directory
   tar -czf uploads-backup-$(date +%Y%m%d).tar.gz uploads/
   ```

### Monitoring

Consider integrating:
- **Prometheus** for metrics
- **Grafana** for visualization
- **ELK Stack** for log aggregation
- **Sentry** for error tracking

## Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Go Docker Best Practices](https://docs.docker.com/language/golang/)
- [Firebase Admin SDK](https://firebase.google.com/docs/admin/setup)

## Support

For issues and questions:
- GitHub Issues: [Repository Issues](https://github.com/manan-parmar-23/makwatches-be/issues)
- Documentation: Check README.md and API_DOCUMENTATION.md

---

**Version:** 1.0.0  
**Last Updated:** 2025-10-08
