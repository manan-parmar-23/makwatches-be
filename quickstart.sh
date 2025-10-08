#!/bin/bash

# MakWatches Backend - Quick Start Script
# This script helps you quickly set up and run the application

set -e

# Color codes
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}"
cat << "EOF"
╔═══════════════════════════════════════════════╗
║   MakWatches Backend - Quick Start Setup      ║
╚═══════════════════════════════════════════════╝
EOF
echo -e "${NC}"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: Docker is not installed!${NC}"
    echo "Please install Docker first: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}Error: Docker Compose is not installed!${NC}"
    echo "Please install Docker Compose first: https://docs.docker.com/compose/install/"
    exit 1
fi

echo -e "${GREEN}✓ Docker and Docker Compose are installed${NC}"
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}Creating .env file from example.env...${NC}"
    if [ -f example.env ]; then
        cp example.env .env
        echo -e "${GREEN}✓ .env file created${NC}"
        echo -e "${YELLOW}⚠ Please edit .env file with your configuration${NC}"
    else
        echo -e "${RED}Error: example.env not found!${NC}"
        exit 1
    fi
else
    echo -e "${GREEN}✓ .env file exists${NC}"
fi

echo ""

# Check if firebase-admin.json exists
if [ ! -f firebase-admin.json ]; then
    echo -e "${YELLOW}⚠ firebase-admin.json not found${NC}"
    echo "  Firebase Storage uploads will not work without this file."
    echo "  Please add your firebase-admin.json file to the root directory."
    echo "  See FIREBASE_SETUP.md for instructions."
    echo ""
    read -p "Do you want to continue without Firebase? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Setup cancelled. Please add firebase-admin.json and try again."
        exit 1
    fi
else
    echo -e "${GREEN}✓ firebase-admin.json exists${NC}"
fi

echo ""
echo -e "${BLUE}Starting MakWatches Backend...${NC}"
echo ""

# Ask user what they want to do
echo "What would you like to do?"
echo "1) Start in development mode (with local MongoDB and Redis)"
echo "2) Start in production mode (with cloud services)"
echo "3) Build and start"
echo "4) Stop all services"
echo "5) View logs"
echo "6) Clean up (remove containers and images)"
echo ""
read -p "Enter your choice (1-6): " choice

case $choice in
    1)
        echo -e "${BLUE}Starting in development mode...${NC}"
        docker-compose up -d
        ;;
    2)
        echo -e "${BLUE}Starting in production mode...${NC}"
        docker-compose -f docker-compose.prod.yml up -d
        ;;
    3)
        echo -e "${BLUE}Building and starting...${NC}"
        docker-compose up -d --build
        ;;
    4)
        echo -e "${BLUE}Stopping all services...${NC}"
        docker-compose down
        docker-compose -f docker-compose.prod.yml down 2>/dev/null || true
        echo -e "${GREEN}✓ All services stopped${NC}"
        exit 0
        ;;
    5)
        echo -e "${BLUE}Showing logs (Ctrl+C to exit)...${NC}"
        docker-compose logs -f
        exit 0
        ;;
    6)
        echo -e "${YELLOW}This will remove all containers, volumes, and images. Are you sure? (y/n)${NC}"
        read -p "" -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            docker-compose down -v
            docker-compose -f docker-compose.prod.yml down -v 2>/dev/null || true
            docker rmi makwatches-be:latest 2>/dev/null || true
            echo -e "${GREEN}✓ Cleanup completed${NC}"
        fi
        exit 0
        ;;
    *)
        echo -e "${RED}Invalid choice${NC}"
        exit 1
        ;;
esac

# Wait for services to start
echo ""
echo -e "${BLUE}Waiting for services to start...${NC}"
sleep 10

# Check if container is running
if docker ps | grep -q "makwatches-be"; then
    echo -e "${GREEN}✓ Application started successfully!${NC}"
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${GREEN}Application is running!${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "📍 API URL: http://localhost:8080"
    echo "🏥 Health Check: http://localhost:8080/health"
    echo "📖 API Docs: See API_DOCUMENTATION.md"
    echo ""
    echo "Useful commands:"
    echo "  View logs:     docker-compose logs -f"
    echo "  Stop:          docker-compose down"
    echo "  Restart:       docker-compose restart"
    echo "  Status:        docker-compose ps"
    echo ""
    
    # Test health endpoint
    echo -e "${BLUE}Testing health endpoint...${NC}"
    sleep 3
    if curl -f -s http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Health check passed!${NC}"
    else
        echo -e "${YELLOW}⚠ Health check failed. Check logs with: docker-compose logs -f${NC}"
    fi
else
    echo -e "${RED}✗ Failed to start application${NC}"
    echo "Check logs with: docker-compose logs"
    exit 1
fi

echo ""
echo -e "${GREEN}Setup complete! Happy coding! 🚀${NC}"
echo ""
