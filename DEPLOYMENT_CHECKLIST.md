# Production Deployment Checklist

Use this checklist to ensure a smooth production deployment of MakWatches Backend.

## Pre-Deployment

### Infrastructure Setup

- [ ] **Server provisioned** (VPS/Cloud instance with at least 2GB RAM, 2 CPU cores)
- [ ] **Domain name configured** (DNS pointing to server IP)
- [ ] **SSL certificate ready** (Let's Encrypt or commercial)
- [ ] **Firewall configured**
  - [ ] Allow SSH (port 22 or custom)
  - [ ] Allow HTTP (port 80)
  - [ ] Allow HTTPS (port 443)
  - [ ] Allow application port (8080 or custom)

### Database & Services

- [ ] **MongoDB Atlas setup**
  - [ ] Cluster created
  - [ ] Database user created with proper permissions
  - [ ] Network access configured (whitelist server IP or 0.0.0.0/0)
  - [ ] Connection string tested
  - [ ] Indexes created for performance

- [ ] **Redis Cloud setup**
  - [ ] Instance created
  - [ ] Connection details obtained
  - [ ] Connection tested from server

- [ ] **Firebase setup**
  - [ ] Firebase project created
  - [ ] Firebase Storage enabled
  - [ ] Service account created
  - [ ] firebase-admin.json downloaded
  - [ ] Storage bucket configured
  - [ ] Security rules set

### Third-Party Services

- [ ] **Google OAuth configured**
  - [ ] Google Cloud Console project created
  - [ ] OAuth credentials created
  - [ ] Authorized redirect URIs added
  - [ ] Client ID and Secret obtained

- [ ] **Razorpay configured**
  - [ ] Account created and verified
  - [ ] API keys obtained (test and live)
  - [ ] Webhook configured (if needed)

### Docker Setup

- [ ] **Docker Hub account created**
- [ ] **Repository created** (makwatches-be)
- [ ] **Access tokens generated**

### Version Control

- [ ] **GitHub repository setup**
- [ ] **All sensitive files in .gitignore**
  - [ ] .env
  - [ ] firebase-admin.json
  - [ ] Any credential files
- [ ] **README.md updated** with deployment info

## Environment Configuration

### Required Environment Variables

- [ ] `PORT` - Application port (default: 8080)
- [ ] `ENVIRONMENT` - Set to "production"
- [ ] `MONGO_URI` - MongoDB connection string
- [ ] `DATABASE_NAME` - Database name
- [ ] `REDIS_URI` - Redis connection URL
- [ ] `REDIS_PASSWORD` - Redis password
- [ ] `REDIS_DATABASE_NAME` - Redis database name
- [ ] `JWT_SECRET` - Strong random secret for JWT
- [ ] `JWT_EXPIRY` - Token expiration (e.g., "24h")
- [ ] `GOOGLE_CLIENT_ID` - Google OAuth client ID
- [ ] `GOOGLE_CLIENT_SECRET` - Google OAuth client secret
- [ ] `GOOGLE_REDIRECT_URL` - Production callback URL
- [ ] `RAZORPAY_MODE` - Set to "live" for production
- [ ] `RAZORPAY_KEY_ID_TEST` - Razorpay test key
- [ ] `RAZORPAY_KEY_SECRET_TEST` - Razorpay test secret
- [ ] `FIREBASE_PROJECT_ID` - Firebase project ID
- [ ] `FIREBASE_BUCKET_NAME` - Firebase storage bucket
- [ ] `FIREBASE_CREDENTIALS_PATH` - Path to firebase-admin.json
- [ ] `VERCEL_ORIGIN` - Frontend production URL
- [ ] `DEV_ORIGIN` - Frontend development URL (optional)

### Security Checks

- [ ] **Strong JWT secret** (minimum 32 characters, random)
- [ ] **Database credentials secure** (not default passwords)
- [ ] **API keys not exposed** in client-side code
- [ ] **CORS properly configured** (only allow trusted origins)
- [ ] **Rate limiting enabled** (if applicable)
- [ ] **Input validation** implemented

## CI/CD Setup

### GitHub Actions

- [ ] **Secrets configured in GitHub**
  - [ ] DOCKER_USERNAME
  - [ ] DOCKER_PASSWORD
  - [ ] SERVER_HOST
  - [ ] SERVER_USER
  - [ ] SERVER_SSH_KEY
  - [ ] SERVER_PORT (if custom)
  - [ ] APP_URL

- [ ] **Workflow tested**
  - [ ] Build succeeds
  - [ ] Tests pass
  - [ ] Docker image builds
  - [ ] Image pushes to registry

### Server SSH Setup

- [ ] **SSH key generated** (ed25519 or RSA 4096)
- [ ] **Public key added to server** (~/.ssh/authorized_keys)
- [ ] **Private key added to GitHub Secrets**
- [ ] **SSH connection tested** from GitHub Actions
- [ ] **Server directory structure created** (/opt/makwatches-be)

## Deployment

### Initial Setup on Server

- [ ] **Docker installed** on server
- [ ] **Docker Compose installed**
- [ ] **User added to docker group** (optional, for non-root)
- [ ] **Application directory created** (/opt/makwatches-be)
- [ ] **.env file created** with production values
- [ ] **firebase-admin.json placed** in application directory
- [ ] **docker-compose.prod.yml configured**
- [ ] **uploads directory created** with proper permissions

### First Deployment

- [ ] **Build Docker image locally** (test build)
- [ ] **Push image to Docker Hub**
- [ ] **Pull image on server**
- [ ] **Start containers** with docker-compose
- [ ] **Verify containers running**
- [ ] **Check health endpoint** (http://your-server:8080/health)
- [ ] **Test API endpoints**
- [ ] **Check logs** for errors

### Reverse Proxy Setup (Recommended)

- [ ] **Nginx/Traefik installed**
- [ ] **Reverse proxy configured**
  ```nginx
  server {
      listen 80;
      server_name api.yourdomain.com;
      
      location / {
          proxy_pass http://localhost:8080;
          proxy_set_header Host $host;
          proxy_set_header X-Real-IP $remote_addr;
          proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
          proxy_set_header X-Forwarded-Proto $scheme;
      }
  }
  ```
- [ ] **SSL certificate installed** (Let's Encrypt)
- [ ] **HTTPS configured**
- [ ] **HTTP to HTTPS redirect** enabled

## Testing

### Functional Testing

- [ ] **Health check endpoint** working
- [ ] **Authentication** working
  - [ ] Registration
  - [ ] Login
  - [ ] Google OAuth
  - [ ] JWT validation
- [ ] **Products API** working
  - [ ] List products
  - [ ] Get product details
  - [ ] Create product (admin)
  - [ ] Update product (admin)
  - [ ] Delete product (admin)
  - [ ] Image uploads to Firebase
- [ ] **Cart API** working
- [ ] **Orders API** working
- [ ] **Payment integration** working
- [ ] **CORS** properly configured for frontend

### Performance Testing

- [ ] **Load testing** completed
- [ ] **Response times** acceptable (< 200ms for most endpoints)
- [ ] **Database queries** optimized
- [ ] **Caching** working (Redis)
- [ ] **Image uploads** working (Firebase)

### Security Testing

- [ ] **SQL injection** protected (N/A for MongoDB, but verify)
- [ ] **XSS prevention** implemented
- [ ] **CSRF protection** (if needed)
- [ ] **Authentication** properly secured
- [ ] **Authorization** working (role-based access)
- [ ] **Rate limiting** tested
- [ ] **Security headers** configured

## Monitoring & Logging

- [ ] **Application logs** accessible
- [ ] **Error logging** configured
- [ ] **Log rotation** setup
- [ ] **Monitoring solution** chosen
  - [ ] Basic: Docker logs
  - [ ] Advanced: ELK, Prometheus, Grafana
- [ ] **Alerting** configured
  - [ ] Server down alerts
  - [ ] High error rate alerts
  - [ ] Disk space alerts
  - [ ] Memory/CPU alerts
- [ ] **Uptime monitoring** (UptimeRobot, Pingdom, etc.)

## Backup Strategy

- [ ] **MongoDB backup** configured
  - [ ] Automated daily backups
  - [ ] Backup retention policy defined
  - [ ] Restore procedure tested
- [ ] **Uploads backup** configured
  - [ ] Firebase handles storage (verify backup policy)
  - [ ] Local uploads directory backed up
- [ ] **Configuration backup**
  - [ ] .env file backed up securely
  - [ ] docker-compose files backed up
  - [ ] firebase-admin.json backed up securely

## Documentation

- [ ] **API documentation** updated
- [ ] **Deployment documentation** created
- [ ] **Environment variables** documented
- [ ] **Troubleshooting guide** created
- [ ] **Rollback procedure** documented
- [ ] **Team trained** on deployment process

## Post-Deployment

### Immediate Checks (First Hour)

- [ ] **Application accessible** from internet
- [ ] **All API endpoints** responding
- [ ] **Frontend can connect** to backend
- [ ] **Authentication flows** working end-to-end
- [ ] **File uploads** working
- [ ] **Payment processing** working
- [ ] **No critical errors** in logs
- [ ] **SSL certificate** valid and working

### First Day Monitoring

- [ ] **Monitor error rates**
- [ ] **Monitor response times**
- [ ] **Check resource usage** (CPU, memory, disk)
- [ ] **Verify backups** are running
- [ ] **Test all critical user flows**
- [ ] **Monitor database performance**
- [ ] **Check cache hit rates**

### First Week Tasks

- [ ] **Performance optimization** based on real traffic
- [ ] **Scale resources** if needed
- [ ] **Fine-tune monitoring** and alerts
- [ ] **Update documentation** with learnings
- [ ] **Create runbook** for common issues

## Rollback Plan

- [ ] **Previous version** tagged in Docker Hub
- [ ] **Rollback procedure** documented:
  ```bash
  # On server
  cd /opt/makwatches-be
  
  # Change image tag to previous version
  docker-compose pull yourusername/makwatches-be:previous-tag
  
  # Restart with previous version
  docker-compose down
  docker-compose up -d
  ```
- [ ] **Database rollback** strategy (if schema changed)
- [ ] **Rollback tested** in staging environment

## Maintenance

### Regular Tasks

- [ ] **Weekly**: Review logs for errors
- [ ] **Weekly**: Check disk space
- [ ] **Weekly**: Review monitoring dashboards
- [ ] **Monthly**: Update dependencies
- [ ] **Monthly**: Review and rotate credentials
- [ ] **Monthly**: Test backup restoration
- [ ] **Quarterly**: Security audit
- [ ] **Quarterly**: Performance review
- [ ] **Quarterly**: Disaster recovery drill

### Update Procedure

- [ ] **Test updates** in staging first
- [ ] **Create backup** before update
- [ ] **Follow deployment checklist**
- [ ] **Monitor closely** after update
- [ ] **Have rollback plan** ready

## Sign-off

- [ ] **Technical lead** reviewed and approved
- [ ] **Security review** completed
- [ ] **Performance baseline** established
- [ ] **Monitoring** confirmed working
- [ ] **Documentation** complete
- [ ] **Team notified** of deployment
- [ ] **Stakeholders informed**

---

## Quick Commands Reference

```bash
# Deploy
./deploy.sh deploy

# Check status
./deploy.sh status

# View logs
./deploy.sh logs

# Restart
./deploy.sh restart

# Stop
./deploy.sh stop

# Health check
curl https://api.yourdomain.com/health

# View running containers
docker ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f --tail=100
```

---

**Date Completed:** _______________  
**Deployed By:** _______________  
**Version:** _______________
