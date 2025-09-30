.PHONY: build run dev test clean lint vet

# Application name
APP_NAME=makwatches-be

# Main application entrypoint
MAIN_PATH=./cmd/api

# Build the application
build:
    @echo "Building $(APP_NAME)..."
    go build -o bin/$(APP_NAME) $(MAIN_PATH)

# Run the application
run: build
    @echo "Running $(APP_NAME)..."
    ./bin/$(APP_NAME)

# Run with hot reload using air (install with: go install github.com/cosmtrek/air@latest)
dev:
    @echo "Starting development server with hot reload..."
    air

# Run tests
test:
    @echo "Running tests..."
    go test ./... -v

# Clean build artifacts
clean:
    @echo "Cleaning..."
    rm -rf ./bin
    go clean

# Run linter
lint:
    @echo "Running linter..."
    golangci-lint run ./...

# Run go vet
vet:
    @echo "Running go vet..."
    go vet ./...

# Generate API documentation
docs:
    @echo "Generating API documentation..."
    swag init -g cmd/api/main.go -o ./docs

# Show help
help:
    @echo "Available commands:"
    @echo "  make build    - Build the application"
    @echo "  make run      - Build and run the application"
    @echo "  make dev      - Run with hot reload (requires air)"
    @echo "  make test     - Run tests"
    @echo "  make clean    - Clean build artifacts"
    @echo "  make lint     - Run linter (requires golangci-lint)"
    @echo "  make vet      - Run go vet"
    @echo "  make docs     - Generate API documentation (requires swag)"
