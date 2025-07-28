# dev.ps1 - PowerShell script for development commands

param(
    [Parameter(Position=0)]
    [string]$Command = "help"
)

# Application name
$AppName = "pehnaw-be"

# Main application entrypoint
$MainPath = "./cmd/api"

# Function to show help
function Show-Help {
    Write-Host "Available commands:"
    Write-Host "  .\dev.ps1 build    - Build the application"
    Write-Host "  .\dev.ps1 run      - Build and run the application"
    Write-Host "  .\dev.ps1 dev      - Run with hot reload using Air"
    Write-Host "  .\dev.ps1 test     - Run tests"
    Write-Host "  .\dev.ps1 clean    - Clean build artifacts"
    Write-Host "  .\dev.ps1 vet      - Run go vet"
    Write-Host "  .\dev.ps1 docs     - Generate API documentation (requires swag)"
}

# Function to build the application
function Build-App {
    Write-Host "Building $AppName..."
    if (-not (Test-Path -Path "bin")) {
        New-Item -ItemType Directory -Path "bin" | Out-Null
    }
    go build -o "bin/$AppName.exe" $MainPath
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Build successful!" -ForegroundColor Green
    } else {
        Write-Host "Build failed!" -ForegroundColor Red
    }
}

# Function to run the application
function Run-App {
    Build-App
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Running $AppName..."
        & "./bin/$AppName.exe"
    }
}

# Function to run with hot reload
function Dev-App {
    Write-Host "Starting development server with hot reload..."
    try {
        air
    }
    catch {
        Write-Host "Error: Air is not installed or not in your PATH" -ForegroundColor Red
        Write-Host "To install Air, run: go install github.com/air-verse/air@latest" -ForegroundColor Yellow
        Write-Host "Make sure your Go bin directory is in your PATH: $env:USERPROFILE\go\bin" -ForegroundColor Yellow
    }
}

# Function to run tests
function Test-App {
    Write-Host "Running tests..."
    go test ./... -v
}

# Function to clean build artifacts
function Clean-App {
    Write-Host "Cleaning..."
    if (Test-Path -Path "bin") {
        Remove-Item -Path "bin" -Recurse -Force
    }
    if (Test-Path -Path "tmp") {
        Remove-Item -Path "tmp" -Recurse -Force
    }
    go clean
    Write-Host "Clean complete!" -ForegroundColor Green
}

# Function to run go vet
function Vet-App {
    Write-Host "Running go vet..."
    go vet ./...
}

# Function to generate API documentation
function Gen-Docs {
    Write-Host "Generating API documentation..."
    try {
        swag init -g cmd/api/main.go -o ./docs
    }
    catch {
        Write-Host "Error: swag is not installed" -ForegroundColor Red
        Write-Host "To install swag, run: go install github.com/swaggo/swag/cmd/swag@latest" -ForegroundColor Yellow
    }
}

# Execute the requested command
switch ($Command.ToLower()) {
    "build" { Build-App }
    "run" { Run-App }
    "dev" { Dev-App }
    "test" { Test-App }
    "clean" { Clean-App }
    "vet" { Vet-App }
    "docs" { Gen-Docs }
    default { Show-Help }
}
