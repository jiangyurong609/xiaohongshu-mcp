#!/bin/bash
# Comprehensive deployment script for xiaohongshu-mcp
# Usage: ./scripts/deploy.sh
#
# Prerequisites:
# - SSH key access to opc@165.1.72.115
# - Go 1.21+ installed locally

set -e

# Configuration
REMOTE_HOST="opc@165.1.72.115"
REMOTE_DIR="/home/opc"
BINARY_NAME="xiaohongshu-mcp-linux-amd64"
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "=== Xiaohongshu MCP Deployment Script ==="
echo "Project directory: $PROJECT_DIR"
echo ""

# Step 1: Build
echo "[1/4] Building for Linux AMD64..."
cd "$PROJECT_DIR"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o "$BINARY_NAME" .
echo "       Built: $BINARY_NAME ($(ls -lh "$BINARY_NAME" | awk '{print $5}'))"

# Step 2: Upload
echo "[2/4] Uploading to $REMOTE_HOST..."
scp "$BINARY_NAME" "$REMOTE_HOST:$REMOTE_DIR/${BINARY_NAME}.new"
echo "       Uploaded successfully"

# Step 3: Deploy
echo "[3/4] Deploying on remote server..."
ssh "$REMOTE_HOST" << 'ENDSSH'
set -e
cd /home/opc

# Stop existing process
if pgrep -f xiaohongshu-mcp-linux-amd64 > /dev/null; then
    echo "       Stopping existing process..."
    pkill -f xiaohongshu-mcp-linux-amd64 || true
    sleep 2
fi

# Backup old binary
if [ -f xiaohongshu-mcp-linux-amd64 ]; then
    mv xiaohongshu-mcp-linux-amd64 xiaohongshu-mcp-linux-amd64.bak
fi

# Replace with new binary
mv xiaohongshu-mcp-linux-amd64.new xiaohongshu-mcp-linux-amd64
chmod +x xiaohongshu-mcp-linux-amd64

# Start new process
echo "       Starting new process..."
nohup ./xiaohongshu-mcp-linux-amd64 > xiaohongshu-mcp.log 2>&1 &

# Wait for startup
sleep 3
ENDSSH

# Step 4: Verify
echo "[4/4] Verifying deployment..."
HEALTH=$(ssh "$REMOTE_HOST" "curl -s http://localhost:18060/health")
if echo "$HEALTH" | grep -q '"status":"healthy"'; then
    echo "       Health check: OK"
else
    echo "       Health check: FAILED"
    echo "       Response: $HEALTH"
    exit 1
fi

echo ""
echo "=== Deployment Complete ==="
echo "Service URL: http://165.1.72.115.nip.io:18060"
echo ""
echo "API Endpoints:"
echo "  - GET  /health                - Health check"
echo "  - GET  /api/v1/login/status   - Check login status"
echo "  - GET  /api/v1/login/qrcode   - Get login QR code"
echo "  - POST /api/v1/publish        - Publish image post"
echo "  - POST /api/v1/publish_video  - Publish video (supports URL)"
echo "  - POST /api/v1/upload/video   - Upload video file"
echo "  - POST /api/v1/feeds/search   - Search feeds"
echo "  - POST /api/v1/feeds/comment  - Post comment"
echo ""

# Cleanup local binary
rm -f "$PROJECT_DIR/$BINARY_NAME"
echo "Local build cleaned up."
