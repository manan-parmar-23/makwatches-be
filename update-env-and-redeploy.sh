#!/bin/bash

# MakWatches Backend Environment Update and Redeployment Script
# This script updates environment variables and redeploys the container

set -e  # Exit on any error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CONTAINER_NAME="makwatches-be-api"
IMAGE_NAME="aditya110/makwatches-be:latest"
ENV_FILE=".env"
COMPOSE_FILE="docker-compose.prod.yml"

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

show_help() {
    echo "MakWatches Backend Environment Update and Redeployment Script"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --commit-first    Commit changes before redeployment"
    echo "  --no-backup      Skip creating backup of current container"
    echo "  --force          Force update without confirmation"
    echo "  --help           Show this help message"
    echo ""
    echo "This script will:"
    echo "  1. Check current container status"
    echo "  2. Backup current container (optional)"
    echo "  3. Commit changes to git (optional)"
    echo "  4. Pull latest image from Docker Hub"
    echo "  5. Stop current container"
    echo "  6. Start new container with updated environment"
    echo "  7. Verify deployment"
}

check_requirements() {
    log_info "Checking requirements..."
    
    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check if docker-compose is installed
    if ! docker compose version &> /dev/null; then
        log_error "Docker Compose is not available. Please install Docker Compose first."
        exit 1
    fi
    
    # Check if .env file exists
    if [ ! -f "$ENV_FILE" ]; then
        log_error "Environment file $ENV_FILE not found."
        exit 1
    fi
    
    # Check if git is available (for commit option)
    if ! command -v git &> /dev/null; then
        log_warning "Git is not installed. Commit functionality will be disabled."
    fi
    
    log_success "All requirements met."
}

check_container_status() {
    log_info "Checking current container status..."
    
    if docker ps | grep -q "$CONTAINER_NAME"; then
        log_success "Container $CONTAINER_NAME is currently running."
        
        # Show container info
        echo ""
        echo "Current container information:"
        docker ps --filter "name=$CONTAINER_NAME" --format "table {{.ID}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}"
        echo ""
        
        return 0
    else
        log_warning "Container $CONTAINER_NAME is not running."
        
        # Check if container exists but stopped
        if docker ps -a | grep -q "$CONTAINER_NAME"; then
            log_info "Container exists but is stopped."
            return 1
        else
            log_info "Container does not exist. Will create new container."
            return 2
        fi
    fi
}

backup_container() {
    if [ "$NO_BACKUP" = true ]; then
        log_info "Skipping container backup as requested."
        return 0
    fi
    
    log_info "Creating backup of current container..."
    
    # Create backup image with timestamp
    BACKUP_TAG="backup-$(date +%Y%m%d-%H%M%S)"
    
    if docker ps | grep -q "$CONTAINER_NAME"; then
        docker commit "$CONTAINER_NAME" "${IMAGE_NAME%:*}:$BACKUP_TAG"
        log_success "Container backed up as ${IMAGE_NAME%:*}:$BACKUP_TAG"
    else
        log_warning "Container not running, skipping backup."
    fi
}

commit_changes() {
    if [ "$COMMIT_FIRST" != true ]; then
        return 0
    fi
    
    if ! command -v git &> /dev/null; then
        log_warning "Git not available, skipping commit."
        return 0
    fi
    
    log_info "Committing changes to git..."
    
    # Check if there are any changes
    if git diff --quiet && git diff --staged --quiet; then
        log_info "No changes to commit."
        return 0
    fi
    
    # Show current status
    echo ""
    echo "Current git status:"
    git status --short
    echo ""
    
    # Add all changes
    git add .
    
    # Commit with timestamp
    COMMIT_MSG="Update environment configuration - $(date '+%Y-%m-%d %H:%M:%S')"
    git commit -m "$COMMIT_MSG"
    
    log_success "Changes committed: $COMMIT_MSG"
    
    # Ask if user wants to push
    if [ "$FORCE" != true ]; then
        read -p "Do you want to push changes to remote repository? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            git push
            log_success "Changes pushed to remote repository."
        fi
    fi
}

pull_latest_image() {
    log_info "Pulling latest image from Docker Hub..."
    
    docker pull "$IMAGE_NAME"
    log_success "Latest image pulled: $IMAGE_NAME"
}

stop_current_container() {
    log_info "Stopping current container..."
    
    if docker ps | grep -q "$CONTAINER_NAME"; then
        docker stop "$CONTAINER_NAME"
        log_success "Container $CONTAINER_NAME stopped."
    else
        log_info "Container $CONTAINER_NAME is not running."
    fi
    
    # Remove the container if it exists
    if docker ps -a | grep -q "$CONTAINER_NAME"; then
        docker rm "$CONTAINER_NAME"
        log_success "Container $CONTAINER_NAME removed."
    fi
}

start_new_container() {
    log_info "Starting new container with updated environment..."
    
    # Use docker compose to start with the updated .env file
    if [ -f "$COMPOSE_FILE" ]; then
        docker compose -f "$COMPOSE_FILE" up -d
        log_success "Container started using docker compose."
    else
        log_warning "Docker compose file not found, using docker run..."
        
        # Fallback to docker run with environment variables
        docker run -d \
            --name "$CONTAINER_NAME" \
            --env-file "$ENV_FILE" \
            -p 8080:8080 \
            -v "$(pwd)/uploads:/app/uploads" \
            -v "$(pwd)/firebase-admin.json:/app/firebase-admin.json:ro" \
            --restart unless-stopped \
            "$IMAGE_NAME"
        
        log_success "Container started using docker run."
    fi
}

verify_deployment() {
    log_info "Verifying deployment..."
    
    # Wait a moment for container to start
    sleep 5
    
    # Check if container is running
    if docker ps | grep -q "$CONTAINER_NAME"; then
        log_success "Container is running."
        
        # Check health endpoint
        log_info "Checking health endpoint..."
        
        for i in {1..12}; do  # Try for 60 seconds (5 * 12)
            if curl -f -s http://localhost:8080/health > /dev/null 2>&1; then
                log_success "Health check passed. Application is running correctly."
                break
            else
                if [ $i -eq 12 ]; then
                    log_error "Health check failed after 60 seconds."
                    log_info "Container logs:"
                    docker logs --tail 20 "$CONTAINER_NAME"
                    exit 1
                else
                    log_info "Health check attempt $i/12 failed, retrying in 5 seconds..."
                    sleep 5
                fi
            fi
        done
        
        # Show final container status
        echo ""
        echo "Final container status:"
        docker ps --filter "name=$CONTAINER_NAME" --format "table {{.ID}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}"
        echo ""
        
    else
        log_error "Container failed to start."
        
        # Show logs for debugging
        log_info "Container logs:"
        docker logs "$CONTAINER_NAME" 2>/dev/null || echo "No logs available"
        
        exit 1
    fi
}

show_summary() {
    echo ""
    echo "=============================="
    echo "DEPLOYMENT SUMMARY"
    echo "=============================="
    echo "✓ Environment file: $ENV_FILE"
    echo "✓ Container name: $CONTAINER_NAME"
    echo "✓ Image: $IMAGE_NAME"
    echo "✓ Application URL: http://localhost:8080"
    echo "✓ Health check: http://localhost:8080/health"
    echo ""
    
    if [ "$COMMIT_FIRST" = true ]; then
        echo "✓ Changes committed to git"
    fi
    
    echo "✓ Deployment completed successfully!"
}

# Parse command line arguments
COMMIT_FIRST=false
NO_BACKUP=false
FORCE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --commit-first)
            COMMIT_FIRST=true
            shift
            ;;
        --no-backup)
            NO_BACKUP=true
            shift
            ;;
        --force)
            FORCE=true
            shift
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Main execution
main() {
    echo "=============================="
    echo "MakWatches Backend Update & Redeploy"
    echo "=============================="
    echo ""
    
    # Check requirements
    check_requirements
    
    # Show current environment info
    log_info "Current environment configuration:"
    echo "Environment file: $ENV_FILE"
    echo "Container name: $CONTAINER_NAME"
    echo "Image: $IMAGE_NAME"
    echo ""
    
    # Check current container status
    set +e  # Disable exit on error temporarily
    check_container_status
    CONTAINER_STATUS=$?
    set -e  # Re-enable exit on error
    
    # Confirmation (unless force flag is used)
    if [ "$FORCE" != true ]; then
        echo "This will:"
        echo "  1. $([ "$COMMIT_FIRST" = true ] && echo "Commit changes to git" || echo "Skip git commit")"
        if [ $CONTAINER_STATUS -eq 0 ] || [ $CONTAINER_STATUS -eq 1 ]; then
            echo "  2. $([ "$NO_BACKUP" = true ] && echo "Skip backup creation" || echo "Create backup of current container")"
            echo "  3. Pull latest image from Docker Hub"
            echo "  4. Stop and remove current container"
            echo "  5. Start new container with updated environment"
        else
            echo "  2. Pull latest image from Docker Hub"
            echo "  3. Start new container with updated environment"
        fi
        echo "  $([ $CONTAINER_STATUS -eq 2 ] && echo "4" || echo "6"). Verify deployment"
        echo ""
        
        read -p "Do you want to continue? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Deployment cancelled."
            exit 0
        fi
    fi
    
    # Execute deployment steps
    if [ $CONTAINER_STATUS -eq 0 ] || [ $CONTAINER_STATUS -eq 1 ]; then
        # Container exists, do full backup and stop process
        backup_container
        stop_current_container
    fi
    
    commit_changes
    pull_latest_image
    start_new_container
    verify_deployment
    show_summary
    
    log_success "Environment update and redeployment completed successfully!"
}

# Run main function
main "$@"
