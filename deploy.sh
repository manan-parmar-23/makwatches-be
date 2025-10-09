#!/bin/bash

# MakWatches Backend Deployment Script
# This script automates the deployment process

set -e  # Exit on any error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="makwatches-be"
COMPOSE_FILE="docker-compose.prod.yml"
ENV_FILE=".env"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_requirements() {
    log_info "Checking requirements..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    if [ ! -f "$ENV_FILE" ]; then
        log_error ".env file not found. Please create it from example.env"
        exit 1
    fi
    
    if [ ! -f "firebase-admin.json" ]; then
        log_warning "firebase-admin.json not found. Make sure to add it if using Firebase."
    fi
    
    log_success "All requirements met!"
}

build_image() {
    log_info "Building Docker image..."
    docker build -t ${PROJECT_NAME}:latest .
    log_success "Docker image built successfully!"
}

deploy() {
    log_info "Deploying application..."
    
    # Pull latest image if using remote registry
    if [ ! -z "$DOCKER_USERNAME" ]; then
        log_info "Pulling latest image from registry..."
        docker pull ${DOCKER_USERNAME}/${PROJECT_NAME}:latest || log_warning "Could not pull image, will use local build"
    fi
    
    # Stop existing containers
    log_info "Stopping existing containers..."
    docker compose -f $COMPOSE_FILE down || true
    
    # Start new containers
    log_info "Starting new containers..."
    docker compose -f $COMPOSE_FILE up -d
    
    # Wait for health check
    log_info "Waiting for application to be healthy..."
    sleep 10
    
    # Check if container is running
    if docker ps | grep -q "${PROJECT_NAME}-api"; then
        log_success "Application deployed successfully!"
    else
        log_error "Application failed to start!"
        docker compose -f $COMPOSE_FILE logs --tail=50
        exit 1
    fi
}

cleanup() {
    log_info "Cleaning up old images..."
    docker image prune -af
    log_success "Cleanup completed!"
}

show_status() {
    log_info "Application Status:"
    docker compose -f $COMPOSE_FILE ps
    echo ""
    log_info "Recent Logs:"
    docker compose -f $COMPOSE_FILE logs --tail=20
}

show_logs() {
    docker compose -f $COMPOSE_FILE logs -f
}

stop_app() {
    log_info "Stopping application..."
    docker compose -f $COMPOSE_FILE down
    log_success "Application stopped!"
}

restart_app() {
    log_info "Restarting application..."
    docker compose -f $COMPOSE_FILE restart
    log_success "Application restarted!"
}

# Main script
case "${1:-deploy}" in
    deploy)
        check_requirements
        deploy
        show_status
        ;;
    build)
        check_requirements
        build_image
        ;;
    build-deploy)
        check_requirements
        build_image
        deploy
        show_status
        ;;
    stop)
        stop_app
        ;;
    restart)
        restart_app
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs
        ;;
    cleanup)
        cleanup
        ;;
    *)
        echo "Usage: $0 {deploy|build|build-deploy|stop|restart|status|logs|cleanup}"
        echo ""
        echo "Commands:"
        echo "  deploy        - Deploy the application (pull and start)"
        echo "  build         - Build Docker image locally"
        echo "  build-deploy  - Build and deploy the application"
        echo "  stop          - Stop the application"
        echo "  restart       - Restart the application"
        echo "  status        - Show application status"
        echo "  logs          - Show and follow application logs"
        echo "  cleanup       - Remove old Docker images"
        exit 1
        ;;
esac

exit 0
