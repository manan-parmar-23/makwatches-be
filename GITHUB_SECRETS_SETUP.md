# GitHub Secrets Setup Guide

## Required Secrets for CI/CD Pipeline

Go to your GitHub repository: `https://github.com/manan-parmar-23/makwatches-be`

Then navigate to: **Settings → Secrets and variables → Actions → New repository secret**

## Docker Hub Secrets

### DOCKER_USERNAME
- **Name:** `DOCKER_USERNAME`
- **Value:** `adityagarg646@gmail.com`
- Click "Add secret"

### DOCKER_PASSWORD
- **Name:** `DOCKER_PASSWORD`
- **Value:** `DeepAditya@10`
- Click "Add secret"

**⚠️ IMPORTANT SECURITY NOTE:**
For production, it's highly recommended to use a Docker Hub **Access Token** instead of your account password.

To create an access token:
1. Go to https://hub.docker.com/settings/security
2. Click "New Access Token"
3. Give it a name (e.g., "GitHub Actions")
4. Select permissions: "Read, Write, Delete"
5. Click "Generate"
6. Copy the token immediately (you won't see it again!)
7. Use this token as your `DOCKER_PASSWORD` instead

## Optional Secrets (for Server Deployment)

### SERVER_HOST
- **Name:** `SERVER_HOST`
- **Value:** Your server IP or domain (e.g., `123.45.67.89` or `api.yourdomain.com`)

### SERVER_USER
- **Name:** `SERVER_USER`
- **Value:** SSH username (e.g., `root` or `ubuntu`)

### SERVER_SSH_KEY
- **Name:** `SERVER_SSH_KEY`
- **Value:** Your private SSH key for server access
```
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAA...
-----END OPENSSH PRIVATE KEY-----
```

### SERVER_PORT (Optional)
- **Name:** `SERVER_PORT`
- **Value:** SSH port (default: `22`)

### APP_URL (Optional)
- **Name:** `APP_URL`
- **Value:** Your production API URL (e.g., `https://api.yourdomain.com`)

## Verification

After adding the secrets:

1. Go to your repository's **Actions** tab
2. Find the failed workflow run
3. Click "Re-run all jobs"
4. The pipeline should now build and push successfully

## Docker Hub Repository

Make sure you have created a repository on Docker Hub:
1. Go to https://hub.docker.com/
2. Log in with your credentials
3. Click "Create Repository"
4. Name it: `makwatches-be`
5. Set visibility (public or private)
6. Click "Create"

## Troubleshooting

### "invalid reference format" error
- Verify `DOCKER_USERNAME` is set correctly
- Check for any typos in the secret name
- Ensure the secret value doesn't have extra spaces

### "unauthorized: incorrect username or password" error
- Double-check your Docker Hub credentials
- Try using an access token instead of password
- Verify you can log in manually: `docker login -u adityagarg646@gmail.com`

### "denied: requested access to the resource is denied" error
- Make sure the Docker Hub repository exists
- Verify your account has push permissions
- Check if the repository name matches exactly

---

**Last Updated:** 2025-10-09
