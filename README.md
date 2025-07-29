# Pehnaw Backend

This repository contains the backend API for the Pehnaw application.

## Tech Stack

- Go (1.21)
- Fiber v2 (Web Framework)
- MongoDB with Redis (for data persistence and caching)
- JWT Authentication

## Project Structure
```
├── cmd/               # Application entrypoints 
│   └── api/           # API server 
├── internal/          # Private application code
│   ├── config/        # Configuration management
│   ├── database/      # Database connections and operations
│   ├── handlers/      # Request handlers
│   ├── middleware/    # Middleware components
│   └── models/        # Data models
├── pkg/               # Public libraries
│   └── utils/         # Utility functions
├── .air.toml          # Hot reload configuration
├── .env               # Environment variables
├── example.env        # Example environment file
├── go.mod             # Go module dependencies
├── go.sum             # Go module checksum
└── dev.ps1            # PowerShell development script
```

## API Endpoints

### Authentication
- `POST /auth/register` - Register a new user (name, email, password)
- `POST /auth/login` - Login with email and password
- `GET /auth/google` - Initiate Google OAuth login
- `GET /auth/google/callback` - Handle Google OAuth callback
- `GET /me` - Get current authenticated user's profile

### Products
- `GET /products` - Get all products with optional category and price filters
- `GET /products/:id` - Get a single product by ID

### Cart (Protected Routes)
- `POST /cart` - Add product to cart (requires authentication)
- `GET /cart/:userID` - Get a user's cart (requires authentication)
- `DELETE /cart/:userID/:productID` - Remove item from cart (requires authentication)

### Orders (Protected Routes)
- `POST /checkout` - Place order (requires authentication)
- `GET /orders/:userID` - Get order history for a user (requires authentication)

### Recommendations (Protected Routes)
- `GET /recommendations/:userID` - Get AI-based product recommendations (requires authentication)

## Getting Started

### Prerequisites

- Go 1.21 or higher
- MongoDB (for data storage)
- Redis (for caching)
- Air (for hot reloading, similar to nodemon in Express)

### Installation

1. Clone the repository
   ```sh
   git clone https://github.com/the-devesta/pehnaw-be.git
   cd pehnaw-be
   ```

2. Install dependencies
   ```sh
   go mod download
   ```

3. Install Air (hot reloading tool)
   ```sh
   go install github.com/air-verse/air@latest
   ```

4. Set up environment variables
   ```sh
   # In PowerShell
   Copy-Item -Path example.env -Destination .env
   ```
   Edit the .env file with your configuration


## Development Commands

### Running the Application with Hot Reload (like nodemon)
```sh
# Start the development server with Air (hot reloading)
air

# Or use the PowerShell script
./dev.ps1
```

### Building and Running the Application
```sh
# Build the application
go build -o bin/pehnaw-be.exe ./cmd/api

# Run the built application
./bin/pehnaw-be.exe
```

### Testing
```sh
# Run all tests
go test ./... -v

# Run tests in a specific package
go test ./internal/handlers -v
```

### Maintenance
```sh
# Clean build artifacts
Remove-Item -Path ./bin -Recurse -Force -ErrorAction SilentlyContinue
go clean

# Check for issues with code
go vet ./...
```

## Setting up MongoDB and Redis

### Using Docker
The easiest way to set up MongoDB and Redis is with Docker:

```sh
# Start MongoDB and Redis with Docker Compose
docker-compose up -d
```

### Manual Installation
If you prefer manual installation:

1. Install MongoDB: [MongoDB Installation Guide](https://www.mongodb.com/docs/manual/installation/)
2. Install Redis: [Redis Installation Guide](https://redis.io/docs/getting-started/)

## Troubleshooting

### MongoDB Connection Issues

If you encounter errors like "No connection could be made because the target machine actively refused it" or "Connection refused":

1. Verify MongoDB is running:
   ```sh
   # Check if MongoDB is running (Windows)
   tasklist | findstr mongo

   # For Docker setup
   docker ps | findstr mongo
   ```

2. Check MongoDB connection string in `.env`:
   - Default is `mongodb://localhost:27017`
   - For Docker, it might be different depending on your Docker configuration

3. Try connecting with MongoDB Compass to verify connectivity.

### CORS Issues

If your frontend application is having CORS issues:

1. The API is configured to allow requests from:
   - `http://localhost:3000` (development frontend)
   - `https://pehnaw.com` (production frontend)

2. To allow requests from other origins, modify `handlers.go`:
   ```go
   app.Use(cors.New(cors.Config{
       AllowOrigins:     "http://localhost:3000,https://pehnaw.com,http://your-origin.com",
       AllowMethods:     "GET,POST,PUT,DELETE",
       AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
       AllowCredentials: true,
   }))
   ```

Note: You cannot use wildcard `*` for `AllowOrigins` when `AllowCredentials` is set to `true` as it's a security risk.

## Environment Configuration

Your `.env` file should include:

```
# Server Configuration
PORT=8080
ENVIRONMENT=development

# MongoDB Configuration
MONGO_URI=mongodb://localhost:27017
DATABASE_NAME=pehnaw

# Redis Configuration
REDIS_URI=localhost:6379

# JWT Configuration
JWT_SECRET=your_jwt_secret_key_here
JWT_EXPIRATION_HOURS=24

# Google OAuth Configuration
GOOGLE_CLIENT_ID=your_google_client_id_here
GOOGLE_CLIENT_SECRET=your_google_client_secret_here
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback

# Logging
LOG_LEVEL=debug
```

### Setting up Google OAuth

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Navigate to "APIs & Services" > "Credentials"
4. Click "Create Credentials" > "OAuth client ID"
5. Select "Web application" as the application type
6. Add your domain to "Authorized JavaScript origins" (e.g., `http://localhost:8080`)
7. Add your callback URL to "Authorized redirect URIs" (e.g., `http://localhost:8080/auth/google/callback`)
8. Copy the Client ID and Client Secret to your `.env` file
9. For production, update the redirect URL to your production domain (e.g., `https://api.pehnaw.com/auth/google/callback`)
