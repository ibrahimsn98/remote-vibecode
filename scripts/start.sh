#!/bin/bash
# Start script for Claude Web Dashboard
# This script starts all components in the correct order
#
# NOTE: The plugin system is deprecated. Use the shell wrapper instead.
# To use: export CLAUDE_SERVICE_URL=http://localhost:8080
#         ./scripts/remote-claude.sh "your prompt here"
# See README.md for more details.

# Get project directory relative to script location
PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

echo "==================================="
echo "Starting Claude Web Dashboard"
echo "==================================="
echo ""

# Kill any existing processes
echo "Stopping any existing processes..."
pkill -f "cmd/server/main.go" 2>/dev/null
sleep 1

# Start Service
echo "1. Starting Service API..."
cd "$PROJECT_DIR/service"
go run cmd/server/main.go > /tmp/telegram-bridge-service.log 2>&1 &
SERVICE_PID=$!
echo "   Service started (PID: $SERVICE_PID)"
sleep 2

# Verify service is running
if ! curl -s http://localhost:8080/api/v1/health > /dev/null; then
    echo "   ERROR: Service failed to start!"
    exit 1
fi
echo "   ✓ Service is healthy"

# Start Plugin (DEPRECATED - commented out)
# The plugin is deprecated in favor of the shell wrapper approach.
# See plugin/DEPRECATED.md for details.
# echo ""
# echo "2. Starting Plugin..."
# echo "   NOTE: Plugin is deprecated. Use shell wrapper instead."
# CLAUDE_PLUGIN_ROOT=/Users/ibrahim/.claude/plugins/telegram-bridge /Users/ibrahim/.claude/plugins/telegram-bridge/telegram-bridge > /tmp/telegram-bridge-plugin.log 2>&1 &
# PLUGIN_PID=$!
# echo "   Plugin started (PID: $PLUGIN_PID)"
# sleep 3

echo ""
echo "==================================="
echo "Service Started!"
echo "==================================="
echo ""
echo "PID:"
echo "  Service: $SERVICE_PID"
echo ""
echo "To use the shell wrapper:"
echo "  export CLAUDE_SERVICE_URL=http://localhost:8080"
echo "  ./scripts/remote-claude.sh \"your prompt here\""
echo ""
echo "To stop the service, run: pkill -f 'cmd/server/main.go' or Ctrl+C"
echo ""
echo "Access the dashboard at:"
echo "  http://localhost:8080/"
echo ""
echo "Logs:"
echo "  Service: tail -f /tmp/telegram-bridge-service.log"
echo ""

# Save PIDs for stop script
echo "$SERVICE_PID" > /tmp/telegram-bridge-service.pid
