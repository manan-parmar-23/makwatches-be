.PHONY: build run dev test clean lint vet docker-build docker-run docker-stop deploy help

# Application name
APP_NAME=makwatches-be

# Main application entrypoint
MAIN_PATH=./cmd/api

# Docker configuration
DOCKER_IMAGE=$(APP_NAME):latest
DOCKER_COMPOSE=docker-compose
DOCKER_COMPOSE_PROD=docker-compose -f docker-compose.prod.yml

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	go build -o bin/$(APP_NAME) $(MAIN_PATH)

# Run the application
run: build
	@echo "Running $(APP_NAME)..."
	./bin/$(APP_NAME)

# Run with hot reload using air (install with: go install github.com/air-verse/air@latest)
dev:
	@echo "Starting development server with hot reload..."
	air

# Run tests
test:
	@echo "Running tests..."
	go test ./... -v

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf ./bin
	rm -f coverage.out coverage.html
	go clean

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod verify

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Docker: Build image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

# Docker: Run with docker-compose (development)
docker-run:
	@echo "Starting services with docker-compose..."
	$(DOCKER_COMPOSE) up -d

# Docker: Stop services
docker-stop:
	@echo "Stopping services..."
	$(DOCKER_COMPOSE) down

# Docker: View logs
docker-logs:
	@echo "Showing logs..."
	$(DOCKER_COMPOSE) logs -f

# Docker: Rebuild and restart
docker-restart:
	@echo "Rebuilding and restarting services..."
	$(DOCKER_COMPOSE) up -d --build

# Docker: Production deployment
docker-prod:
	@echo "Starting production services..."
	$(DOCKER_COMPOSE_PROD) up -d

# Docker: Stop production services
docker-prod-stop:
	@echo "Stopping production services..."
	$(DOCKER_COMPOSE_PROD) down

# Deploy using deployment script
deploy:
	@echo "Deploying application..."
	./deploy.sh deploy

# Quick start
quickstart:
	@echo "Running quick start..."
	./quickstart.sh

# Check if required files exist
check:
	@echo "Checking required files..."
	@test -f .env && echo "✓ .env exists" || echo "✗ .env missing (copy from example.env)"
	@test -f firebase-admin.json && echo "✓ firebase-admin.json exists" || echo "⚠ firebase-admin.json missing (optional)"
	@test -f go.mod && echo "✓ go.mod exists" || echo "✗ go.mod missing"
	@which docker > /dev/null && echo "✓ Docker installed" || echo "✗ Docker not installed"
	@which docker-compose > /dev/null && echo "✓ Docker Compose installed" || echo "✗ Docker Compose not installed"

# Initialize project (first time setup)
init:
	@echo "Initializing project..."
	@test -f .env || cp example.env .env && echo "✓ Created .env from example.env"
	@test -d uploads || mkdir -p uploads && echo "✓ Created uploads directory"
	@echo "✓ Project initialized"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Edit .env with your configuration"
	@echo "  2. Add firebase-admin.json (if using Firebase)"
	@echo "  3. Run 'make docker-run' or 'make quickstart'"

# Show help
help:
	@echo "MakWatches Backend - Available Commands"
	@echo ""
	@echo "Development:"
	@echo "  make build           - Build the application"
	@echo "  make run             - Build and run the application"
	@echo "  make dev             - Run with hot reload (requires air)"
	@echo "  make test            - Run tests"
	@echo "  make test-coverage   - Run tests with coverage report"
	@echo "  make fmt             - Format code"
	@echo "  make lint            - Run linter (requires golangci-lint)"
	@echo "  make vet             - Run go vet"
	@echo ""
	@echo "Dependencies:"
	@echo "  make deps            - Download dependencies"
	@echo "  make tidy            - Tidy dependencies"
	@echo ""
	@echo "Docker (Development):"
	@echo "  make docker-build    - Build Docker image"
	@echo "  make docker-run      - Start with docker-compose"
	@echo "  make docker-stop     - Stop services"
	@echo "  make docker-logs     - View logs"
	@echo "  make docker-restart  - Rebuild and restart"
	@echo ""
	@echo "Docker (Production):"
	@echo "  make docker-prod     - Start production services"
	@echo "  make docker-prod-stop - Stop production services"
	@echo "  make deploy          - Deploy using deployment script"
	@echo ""
	@echo "Utilities:"
	@echo "  make init            - Initialize project (first time setup)"
	@echo "  make check           - Check if required files exist"
	@echo "  make quickstart      - Run quick start script"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make help            - Show this help message"
	@echo ""
