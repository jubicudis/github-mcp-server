#!/bin/bash

# WHO: DevEnvSetup
# WHAT: Unified development environment setup for TNOS and GitHub MCP servers with QHP bridge
# WHEN: During development setup and environment initialization
# WHERE: System Layer 6 (Integration)
# WHY: To ensure both MCP servers are started in canonical, symmetric fashion
# HOW: Using shell environment variables, canonical port assignments, and server startup routines
# EXTENT: All MCP development operations

# Canonical QHP port assignments
QHP_TNOS_PORT=9001
QHP_GITHUB_PORT=10617
QHP_COPILOT_PORT=8083

# Detect workspace root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
export WORKSPACE_ROOT

# Set up Go environment
if ! command -v go &> /dev/null; then
    echo "Go is not found in your PATH. Attempting to locate..."
    if [ -d "/usr/local/go/bin" ]; then
        export PATH="/usr/local/go/bin:$PATH"
    elif [ -d "$HOME/go/bin" ]; then
        export PATH="$HOME/go/bin:$PATH"
    elif [ -d "/opt/homebrew/bin" ] && [ -f "/opt/homebrew/bin/go" ]; then
        export PATH="/opt/homebrew/bin:$PATH"
    else
        echo "WARNING: Go not found. Some features may not work."
    fi
fi

if command -v go &> /dev/null; then
    GOROOT=$(go env GOROOT)
    export GOROOT
    export PATH="$GOROOT/bin:$PATH"
    export GO111MODULE=on
    echo "Go version: $(go version)"
else
    echo "Go version: Not available"
fi

# Set up Python venv for TNOS MCP server
VENV_DIR="$WORKSPACE_ROOT/systems/python/venv311"
PYTHON311="$VENV_DIR/bin/python3.11"

# Canonical server scripts
GITHUB_MCP_BIN="$WORKSPACE_ROOT/bin/github-mcp-server"
GITHUB_MCP_SRC="$WORKSPACE_ROOT/github-mcp-server/cmd/server"
TNOS_MCP_SERVER_PY="$WORKSPACE_ROOT/mcp/bridge/tnos_mcp_server.py"

# Canonical logs
LOGS_DIR="$WORKSPACE_ROOT/logs"
mkdir -p "$LOGS_DIR"

# Utility: Wait for port
wait_for_port() {
    local host="$1"; local port="$2"; local timeout="${3:-60}"; local waited=0
    echo "[PortWaiter] Waiting for $host:$port (timeout ${timeout}s)..."
    while ! nc -z "$host" "$port" 2>/dev/null; do
        sleep 1; waited=$((waited+1))
        if [ "$waited" -ge "$timeout" ]; then
            echo "[PortWaiter][ERROR] Timeout waiting for $host:$port after $timeout seconds."
            return 1
        fi
        if (( $waited % 10 == 0 )); then
            echo "[PortWaiter] Still waiting for $host:$port... ($waited seconds elapsed)"
        fi
    done
    echo "[PortWaiter] $host:$port is now open after $waited seconds."
    return 0
}

# Utility: Wait for /health endpoint
wait_for_health() {
    local url="$1"; local timeout="${2:-60}"; local waited=0
    while true; do
        if command -v curl >/dev/null 2>&1; then
            local status=$(curl -s -o /dev/null -w "%{http_code}" "$url")
            if [ "$status" = "200" ]; then
                echo "[HealthWaiter] $url is ready."
                return 0
            fi
        fi
        sleep 1; waited=$((waited+1))
        if [ "$waited" -ge "$timeout" ]; then
            echo "[HealthWaiter][ERROR] Timeout waiting for $url after $timeout seconds."
            return 1
        fi
        if (( $waited % 10 == 0 )); then
            echo "[HealthWaiter] Still waiting for $url... ($waited seconds elapsed)"
        fi
    done
}

# Build GitHub MCP server if needed
if [ ! -f "$GITHUB_MCP_BIN" ]; then
    echo "Building GitHub MCP server..."
    cd "$WORKSPACE_ROOT/github-mcp-server" || exit 1
    go build -o "$GITHUB_MCP_BIN" "$GITHUB_MCP_SRC" || { echo "Build failed."; exit 1; }
fi

# Start GitHub MCP server (Go, port 10617)
start_github_mcp() {
    if lsof -i :$QHP_GITHUB_PORT | grep LISTEN; then
        echo "[INFO] GitHub MCP server already running on port $QHP_GITHUB_PORT."
    else
        echo "Starting GitHub MCP server on port $QHP_GITHUB_PORT..."
        nohup "$GITHUB_MCP_BIN" > "$LOGS_DIR/github_mcp_server.log" 2>&1 &
        echo $! > "$LOGS_DIR/github_mcp_server.pid"
        wait_for_port "127.0.0.1" $QHP_GITHUB_PORT 30 || { echo "[ERROR] GitHub MCP server did not open port $QHP_GITHUB_PORT."; exit 1; }
    fi
}

# Start TNOS MCP server (Python, port 9001/8083, with visualization internal)
start_tnos_mcp() {
    if lsof -i :$QHP_TNOS_PORT | grep LISTEN; then
        echo "[INFO] TNOS MCP server already running on port $QHP_TNOS_PORT."
    else
        echo "Starting TNOS MCP server on port $QHP_TNOS_PORT..."
        if [[ -f "$VENV_DIR/bin/activate" ]]; then
            source "$VENV_DIR/bin/activate"
            "$PYTHON311" -m pip install flask --quiet
        fi
        nohup "$PYTHON311" -u "$TNOS_MCP_SERVER_PY" > "$LOGS_DIR/tnos_mcp_server.log" 2>&1 &
        echo $! > "$LOGS_DIR/tnos_mcp_server.pid"
        wait_for_port "127.0.0.1" $QHP_TNOS_PORT 30 || { echo "[ERROR] TNOS MCP server did not open port $QHP_TNOS_PORT."; exit 1; }
        wait_for_port "127.0.0.1" $QHP_COPILOT_PORT 30 || echo "[WARN] Copilot LLM port $QHP_COPILOT_PORT not open (optional)."
    fi
}

# Stop all components
stop_all() {
    echo "Stopping all MCP components..."
    pkill -f "github-mcp-server"
    pkill -f "tnos_mcp_server.py"
    sleep 2
}

# Status for all components
status_all() {
    echo "MCP Component Status:"
    for port in $QHP_GITHUB_PORT $QHP_TNOS_PORT $QHP_COPILOT_PORT; do
        if lsof -i :$port | grep LISTEN >/dev/null; then
            echo "  [RUNNING] Port $port"
        else
            echo "  [STOPPED] Port $port"
        fi
    done
    echo "[INFO] Visualization is provided internally by the TNOS MCP server (port $QHP_COPILOT_PORT)."
}

# Main CLI logic
case "$1" in
    start-all|"")
        stop_all
        start_github_mcp
        start_tnos_mcp
        status_all
        ;;
    stop-all)
        stop_all
        status_all
        ;;
    status-all)
        status_all
        ;;
    *)
        echo "Usage: $0 [start-all|stop-all|status-all]"
        exit 1
        ;;
esac
