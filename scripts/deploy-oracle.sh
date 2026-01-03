#!/bin/bash
# Oracle Cloud VM Deployment Script for xiaohongshu-mcp
# Run this script ON the Oracle Cloud VM after SSH-ing in

set -e

echo "=== Xiaohongshu MCP - Oracle Cloud Deployment ==="

# Update system
echo "[1/6] Updating system..."
sudo apt-get update && sudo apt-get upgrade -y

# Install Docker
echo "[2/6] Installing Docker..."
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
rm get-docker.sh

# Install dependencies for Chrome
echo "[3/6] Installing Chrome dependencies..."
sudo apt-get install -y \
    fonts-liberation \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libatspi2.0-0 \
    libcups2 \
    libdbus-1-3 \
    libdrm2 \
    libgbm1 \
    libgtk-3-0 \
    libnspr4 \
    libnss3 \
    libxcomposite1 \
    libxdamage1 \
    libxfixes3 \
    libxkbcommon0 \
    libxrandr2 \
    xdg-utils

# Create data directory
echo "[4/6] Creating data directory..."
mkdir -p ~/.xiaohongshu-mcp

# Pull and run container
echo "[5/6] Pulling Docker image..."
sudo docker pull ghcr.io/xingyezhiqiu/xiaohongshu-mcp:latest || {
    echo "GitHub Container Registry failed, building from source..."
    cd /tmp
    git clone https://github.com/xingyezhiqiu/xiaohongshu-mcp.git
    cd xiaohongshu-mcp
    sudo docker build -t xiaohongshu-mcp:latest .
}

echo "[6/6] Starting container..."
sudo docker run -d \
    --name xiaohongshu-mcp \
    --restart unless-stopped \
    -p 18060:18060 \
    -v ~/.xiaohongshu-mcp:/root/.xiaohongshu-mcp \
    ghcr.io/xingyezhiqiu/xiaohongshu-mcp:latest || \
sudo docker run -d \
    --name xiaohongshu-mcp \
    --restart unless-stopped \
    -p 18060:18060 \
    -v ~/.xiaohongshu-mcp:/root/.xiaohongshu-mcp \
    xiaohongshu-mcp:latest

# Wait for startup
echo "Waiting for service to start..."
sleep 10

# Check health
echo ""
echo "=== Deployment Complete ==="
curl -s http://localhost:18060/health | jq . || echo "Service starting up..."

# Get public IP
PUBLIC_IP=$(curl -s ifconfig.me)
echo ""
echo "Service URL: http://${PUBLIC_IP}:18060"
echo ""
echo "Next steps:"
echo "1. Open firewall port 18060 in Oracle Cloud Console"
echo "2. Get QR code: curl http://${PUBLIC_IP}:18060/api/v1/login/qrcode"
echo "3. Update social-agent wrangler.toml with: XHS_MCP_BASE = \"http://${PUBLIC_IP}:18060\""
