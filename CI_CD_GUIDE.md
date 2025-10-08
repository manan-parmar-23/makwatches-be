# CI/CD Pipeline Configuration Guide

This guide explains the CI/CD pipeline setup for automated testing, building, and deployment of the MakWatches backend.

## Table of Contents

1. [GitHub Actions Pipeline](#github-actions-pipeline)
2. [Pipeline Stages](#pipeline-stages)
3. [Setup Instructions](#setup-instructions)
4. [Environment Variables](#environment-variables)
5. [Deployment Workflow](#deployment-workflow)
6. [Rollback Procedures](#rollback-procedures)
7. [Troubleshooting](#troubleshooting)

## GitHub Actions Pipeline

The project uses GitHub Actions for automated CI/CD. The workflow file is located at `.github/workflows/docker-deploy.yml`.

### Pipeline Triggers

The pipeline runs on:
- **Push to `main` branch**: Full pipeline (build → test → deploy)
- **Push to `develop` branch**: Build and test only
- **Pull requests to `main`**: Test only
- **Manual trigger**: Via GitHub Actions UI

## Pipeline Stages

### Stage 1: Build and Test

```yaml
jobs:
  build-and-test:
    - Checkout code
    - Set up Go 1.24
    - Install dependencies
    - Run tests
    - Run go vet (static analysis)
    - Build application
```

**Purpose**: Ensure code quality and that the application builds successfully.

**Duration**: ~2-3 minutes

### Stage 2: Docker Build and Push

```yaml
jobs:
  docker-build-push:
    - Checkout code
    - Set up Docker Buildx
    - Log in to Docker Hub
    - Extract metadata (tags, labels)
    - Build and push Docker image
```

**Purpose**: Build optimized Docker image and push to registry.

**Duration**: ~3-5 minutes

**Image Tags Generated**:
- `latest` (for main branch only)
- `<branch-name>-<commit-sha>` (e.g., `main-a1b2c3d`)
- `<branch-name>` (e.g., `main`, `develop`)

### Stage 3: Deploy to Server

```yaml
jobs:
  deploy-to-server:
    - SSH into production server
    - Pull latest Docker image
    - Stop old containers
    - Start new containers
    - Clean up old images
    - Verify deployment
```

**Purpose**: Deploy the application to production server.

**Duration**: ~1-2 minutes

## Setup Instructions

### 1. Docker Hub Setup

1. Create a Docker Hub account at https://hub.docker.com
2. Create a new repository named `makwatches-be`
3. Generate an access token:
   - Go to Account Settings → Security
   - Click "New Access Token"
   - Name it (e.g., "GitHub Actions")
   - Copy the token (you won't see it again!)

### 2. GitHub Secrets Configuration

Go to your GitHub repository → Settings → Secrets and variables → Actions

Add the following secrets:

#### Docker Registry Secrets

| Secret Name | Description | Example |
|------------|-------------|---------|
| `DOCKER_USERNAME` | Docker Hub username | `yourusername` |
| `DOCKER_PASSWORD` | Docker Hub access token | `dckr_pat_xxxxxxxxxxxxx` |

#### Server SSH Secrets

| Secret Name | Description | Example |
|------------|-------------|---------|
| `SERVER_HOST` | Production server IP or domain | `123.45.67.89` or `api.yourdomain.com` |
| `SERVER_USER` | SSH username | `root` or `ubuntu` |
| `SERVER_SSH_KEY` | Private SSH key for authentication | `-----BEGIN RSA PRIVATE KEY-----...` |
| `SERVER_PORT` | SSH port (optional, default: 22) | `22` |

#### Application Secrets (optional)

| Secret Name | Description |
|------------|-------------|
| `APP_URL` | Production application URL |

### 3. Generate SSH Key

On your local machine:

```bash
# Generate SSH key pair
ssh-keygen -t ed25519 -C "github-actions@makwatches" -f github-actions-key

# This creates:
# - github-actions-key (private key) → Add to GitHub Secrets
# - github-actions-key.pub (public key) → Add to server
```

### 4. Configure Production Server

On your production server:

```bash
# 1. Add the public key to authorized_keys
mkdir -p ~/.ssh
chmod 700 ~/.ssh
cat >> ~/.ssh/authorized_keys << 'EOF'
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx github-actions@makwatches
EOF
chmod 600 ~/.ssh/authorized_keys

# 2. Create application directory
sudo mkdir -p /opt/makwatches-be
cd /opt/makwatches-be

# 3. Create .env file
sudo nano .env
# Add your production environment variables

# 4. Add firebase-admin.json
sudo nano firebase-admin.json
# Paste your Firebase credentials

# 5. Create docker-compose.prod.yml
sudo nano docker-compose.prod.yml
# Copy from repository

# 6. Create uploads directory
sudo mkdir -p uploads

# 7. Set proper permissions
sudo chown -R $USER:$USER /opt/makwatches-be
```

### 5. Test SSH Connection

From your local machine:

```bash
# Test SSH connection with the private key
ssh -i github-actions-key username@your-server-ip

# If successful, test from GitHub Actions
# Push a commit to a test branch and check the Actions tab
```

## Environment Variables

### GitHub Actions Environment Variables

These are used within the CI/CD pipeline:

```yaml
env:
  REGISTRY: docker.io
  IMAGE_NAME: makwatches-be
```

### Server Environment Variables

These should be in `/opt/makwatches-be/.env` on your server:

```env
# See example.env for complete list
ENVIRONMENT=production
PORT=8080
MONGO_URI=mongodb+srv://...
JWT_SECRET=your_secret_key
# ... etc
```

## Deployment Workflow

### Normal Deployment Flow

1. **Developer pushes to `main` branch**
   ```bash
   git add .
   git commit -m "feat: new feature"
   git push origin main
   ```

2. **GitHub Actions pipeline starts automatically**
   - Tests run
   - Docker image builds
   - Image pushes to Docker Hub
   - Deploys to production server

3. **Monitor deployment**
   - Go to GitHub → Actions tab
   - Click on the running workflow
   - View logs for each step

4. **Verify deployment**
   ```bash
   # Check application health
   curl https://api.yourdomain.com/health
   
   # Or SSH into server
   ssh user@server
   docker ps
   docker logs makwatches-be-api
   ```

### Manual Deployment

Trigger deployment manually:

1. Go to GitHub repository
2. Click "Actions" tab
3. Select "Build and Deploy Docker Image" workflow
4. Click "Run workflow"
5. Select branch
6. Click "Run workflow" button

## Rollback Procedures

### Method 1: Rollback via GitHub

1. **Find the previous successful commit**
   ```bash
   git log --oneline
   ```

2. **Revert to previous commit**
   ```bash
   git revert <commit-hash>
   git push origin main
   ```

3. **Pipeline will automatically deploy the reverted version**

### Method 2: Manual Rollback on Server

1. **SSH into server**
   ```bash
   ssh user@server
   cd /opt/makwatches-be
   ```

2. **Pull previous image version**
   ```bash
   # Find previous image tag
   docker images | grep makwatches-be
   
   # Pull specific version
   docker pull yourusername/makwatches-be:main-<previous-commit-sha>
   ```

3. **Update docker-compose.prod.yml**
   ```yaml
   services:
     api:
       image: yourusername/makwatches-be:main-<previous-commit-sha>
   ```

4. **Restart services**
   ```bash
   docker-compose -f docker-compose.prod.yml down
   docker-compose -f docker-compose.prod.yml up -d
   ```

### Method 3: Use Deployment Script

```bash
# On server
cd /opt/makwatches-be

# Edit docker-compose.prod.yml to use specific tag
nano docker-compose.prod.yml

# Run deployment script
./deploy.sh deploy
```

## Monitoring Deployment

### View GitHub Actions Logs

1. Go to repository → Actions tab
2. Click on workflow run
3. Click on job name to see logs
4. Expand steps to see details

### View Server Logs

```bash
# SSH into server
ssh user@server

# View Docker logs
docker logs makwatches-be-api

# Follow logs
docker logs -f makwatches-be-api

# View last 100 lines
docker logs --tail=100 makwatches-be-api

# View with timestamps
docker logs -t makwatches-be-api
```

### Health Monitoring

```bash
# Check health endpoint
curl https://api.yourdomain.com/health

# Check container status
docker ps

# Check resource usage
docker stats makwatches-be-api

# Check Docker Compose status
docker-compose -f docker-compose.prod.yml ps
```

## Troubleshooting

### Pipeline Fails at Build Stage

**Symptoms**: Tests fail or build fails

**Solutions**:
```bash
# Run tests locally
go test ./... -v

# Check for syntax errors
go vet ./...

# Ensure all dependencies are in go.mod
go mod tidy
```

### Pipeline Fails at Docker Build

**Symptoms**: Docker build fails

**Solutions**:
```bash
# Test Docker build locally
docker build -t makwatches-be:test .

# Check Dockerfile syntax
docker build --no-cache -t makwatches-be:test .

# Check .dockerignore
cat .dockerignore
```

### Pipeline Fails at Docker Push

**Symptoms**: Cannot push to Docker Hub

**Solutions**:
- Verify `DOCKER_USERNAME` and `DOCKER_PASSWORD` secrets
- Check Docker Hub repository exists
- Verify access token is valid
- Check Docker Hub account status

### Pipeline Fails at Deployment

**Symptoms**: SSH connection fails or deployment script errors

**Solutions**:

1. **SSH Connection Issues**:
   ```bash
   # Test SSH from your machine
   ssh -i private-key user@server
   
   # Check server SSH config
   sudo nano /etc/ssh/sshd_config
   sudo systemctl restart sshd
   
   # Check firewall
   sudo ufw status
   sudo ufw allow 22/tcp
   ```

2. **Deployment Script Issues**:
   ```bash
   # SSH into server
   ssh user@server
   cd /opt/makwatches-be
   
   # Check if files exist
   ls -la
   
   # Check permissions
   chmod +x deploy.sh
   
   # Run manually to see errors
   ./deploy.sh deploy
   ```

3. **Docker Issues on Server**:
   ```bash
   # Check Docker status
   sudo systemctl status docker
   
   # Restart Docker
   sudo systemctl restart docker
   
   # Check Docker logs
   sudo journalctl -u docker -n 50
   ```

### Application Won't Start After Deployment

**Symptoms**: Container starts but health check fails

**Solutions**:
```bash
# Check container logs
docker logs makwatches-be-api

# Check environment variables
docker exec makwatches-be-api env

# Check if .env file is present
ls -la /opt/makwatches-be/.env

# Check if firebase-admin.json is present
ls -la /opt/makwatches-be/firebase-admin.json

# Test inside container
docker exec -it makwatches-be-api sh
curl http://localhost:8080/health
```

### Common Error Messages

#### "Permission denied (publickey)"
- SSH key not properly configured
- Public key not in server's authorized_keys
- Wrong username

#### "Cannot connect to Docker daemon"
- Docker not running on server
- User not in docker group
- Docker socket permissions

#### "Failed to pull image"
- Wrong image name
- Image doesn't exist
- Docker Hub authentication failed

#### "Container exited with code 1"
- Application crashed on startup
- Check logs: `docker logs makwatches-be-api`
- Usually environment variable or connection issues

## Best Practices

### 1. Use Feature Branches

```bash
# Create feature branch
git checkout -b feature/new-feature

# Make changes and test locally
make test
make docker-build

# Push to GitHub (won't deploy)
git push origin feature/new-feature

# Create Pull Request
# After review, merge to main (triggers deployment)
```

### 2. Tag Releases

```bash
# Tag important releases
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

### 3. Monitor Deployments

- Always watch the first deployment after changes
- Keep deployment logs for troubleshooting
- Set up alerts for deployment failures

### 4. Test in Staging First

- Use `develop` branch for staging
- Deploy to staging server first
- Test thoroughly before merging to `main`

### 5. Keep Secrets Secure

- Never commit secrets to repository
- Rotate secrets regularly
- Use different secrets for staging and production
- Audit secret access regularly

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Hub Documentation](https://docs.docker.com/docker-hub/)
- [SSH Key Authentication](https://www.ssh.com/academy/ssh/keygen)
- [Let's Encrypt for SSL](https://letsencrypt.org/getting-started/)

---

**Need Help?**

- Check GitHub Actions logs first
- Review server logs for runtime issues
- Consult [DEPLOYMENT.md](DEPLOYMENT.md) for server setup
- Check [DEPLOYMENT_CHECKLIST.md](DEPLOYMENT_CHECKLIST.md) for completeness

**Version**: 1.0.0  
**Last Updated**: 2025-10-08
