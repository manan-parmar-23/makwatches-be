# 🚀 Quick Start Guide - MakWatches Backend

Get your MakWatches backend up and running in minutes!

## Prerequisites

- Docker (20.10+)
- Docker Compose (2.0+)
- Git

## 🏃 Quick Start (Development)

```bash
# 1. Clone the repository
git clone https://github.com/manan-parmar-23/makwatches-be.git
cd makwatches-be

# 2. Run the quick start script
./quickstart.sh

# OR manually:
# Copy environment file
cp example.env .env

# Edit .env with your settings
nano .env

# Start with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f
```

Your API will be available at: **http://localhost:8080**

## 📋 What You Need

### Required Files

1. **`.env`** - Environment configuration
   ```bash
   cp example.env .env
   # Edit with your values
   ```

2. **`FIREBASE_CREDENTIALS_JSON`** - Firebase credentials (for image uploads)
   - Download from Firebase Console
   - See [FIREBASE_SETUP.md](FIREBASE_SETUP.md) for instructions

### Environment Variables

Minimum required variables in `.env`:

```env
# Database
MONGO_URI=your_mongodb_connection_string

# Redis (optional but recommended)
REDIS_URI=your_redis_host:port
REDIS_PASSWORD=your_redis_password

# JWT Secret (generate a strong random string)
JWT_SECRET=your_super_secret_jwt_key

# Firebase
FIREBASE_PROJECT_ID=your-project-id
FIREBASE_BUCKET_NAME=your-project.appspot.com
```

See `example.env` for all available options.

## 🐳 Docker Commands

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f

# Restart services
docker-compose restart

# Rebuild and start
docker-compose up -d --build

# Check status
docker-compose ps

# Execute command in container
docker-compose exec api sh
```

## 🧪 Testing

```bash
# Health check
curl http://localhost:8080/health

# Test API endpoint
curl http://localhost:8080/welcome

# Run Go tests
go test ./...
```

## 📚 Documentation

- [API Documentation](API_DOCUMENTATION.md) - Complete API reference
- [Deployment Guide](DEPLOYMENT.md) - Production deployment
- [Deployment Checklist](DEPLOYMENT_CHECKLIST.md) - Pre-flight checklist
- [Firebase Setup](FIREBASE_SETUP.md) - Firebase configuration
- [Implementation Guide](IMPLEMENTATION_GUIDE.md) - Technical details
- [Usage Guide](USAGE_GUIDE.md) - Code examples

## 🚀 Production Deployment

For production deployment with CI/CD, see [DEPLOYMENT.md](DEPLOYMENT.md).

Quick production deployment:

```bash
# Use the deployment script
./deploy.sh deploy

# OR manually with production compose file
docker-compose -f docker-compose.prod.yml up -d
```

## 🔧 Troubleshooting

### Common Issues

**Container won't start:**
```bash
docker-compose logs api
```

**Port already in use:**
```bash
# Change PORT in .env
PORT=8081
```

**Database connection error:**
- Check MongoDB URI in .env
- Verify MongoDB is accessible
- Check firewall/network settings

**Firebase uploads fail:**
- Ensure FIREBASE_CREDENTIALS_JSON is set
- Verify Firebase Storage is enabled
- Check service account permissions

See [DEPLOYMENT.md](DEPLOYMENT.md#troubleshooting) for more.

## 📦 Project Structure

```
makwatches-be/
├── cmd/api/              # Application entry point
├── internal/             # Private application code
│   ├── config/          # Configuration
│   ├── database/        # Database layer
│   ├── firebase/        # Firebase integration
│   ├── handlers/        # HTTP handlers
│   ├── middleware/      # Middleware
│   └── models/          # Data models
├── uploads/             # Upload directory (mounted as volume)
├── .github/workflows/   # CI/CD pipelines
├── Dockerfile           # Production Docker image
├── docker-compose.yml   # Development setup
├── docker-compose.prod.yml  # Production setup
├── deploy.sh           # Deployment script
├── quickstart.sh       # Quick start script
└── *.md                # Documentation
```

## 🔐 Security Notes

- Never commit `.env` or service-account JSON
- Use strong JWT secrets (32+ characters)
- Enable HTTPS in production
- Keep dependencies updated
- Review CORS settings for production

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📝 License

See [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Issues**: [GitHub Issues](https://github.com/manan-parmar-23/makwatches-be/issues)
- **Documentation**: Check the `*.md` files in this repository

---

**Happy Coding! 🎉**
