#!/bin/bash

# Deploy Portainer for Docker Management
# Portainer is a web-based Docker management UI

echo "🚀 Deploying Portainer..."
echo ""

# Create volume for Portainer data
echo "Creating Portainer data volume..."
docker volume create portainer_data

# Deploy Portainer
echo "Starting Portainer container..."
docker run -d \
  --name portainer \
  --restart=always \
  -p 9000:9000 \
  -p 9443:9443 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v portainer_data:/data \
  portainer/portainer-ce:latest

echo ""
echo "✅ Portainer deployed successfully!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🌐 Access Portainer at:"
echo ""
echo "   HTTP:  http://139.59.71.95:9000"
echo "   HTTPS: https://139.59.71.95:9443"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📋 First Time Setup:"
echo "1. Open http://139.59.71.95:9000 in your browser"
echo "2. Create admin username and password (min 12 characters)"
echo "3. Click 'Local' environment to manage your Docker containers"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🔍 Portainer Features:"
echo "  • View all containers, images, volumes, networks"
echo "  • Start/stop/restart containers"
echo "  • View container logs in real-time"
echo "  • Access container shell"
echo "  • Monitor resource usage"
echo "  • Deploy new containers"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🐛 Debug Your MakWatches Container:"
echo "1. Go to Containers → makwatches-be-api"
echo "2. Click 'Logs' to see error messages"
echo "3. Click 'Inspect' to see configuration"
echo "4. Click 'Console' to access container shell"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
