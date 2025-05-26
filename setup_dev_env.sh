#!/bin/bash

# WHO: DevEnvSetup
# WHAT: Development environment configuration for GitHub MCP server
# WHEN: During development setup and environment initialization
# WHERE: System Layer 6 (Integration)
# WHY: To ensure consistent Go module resolution and tool configuration
# HOW: Using shell environment variables and workspace paths
# EXTENT: GitHub MCP server development operations

echo "Setting up Go development environment for GitHub MCP Server..."
# HOW: Using shell environment variables and workspace paths
# EXTENT: GitHub MCP server development operations

echo "Setting up Go development environment for GitHub MCP Server..."

# Remove any legacy tnos_venv if it exists (safety, idempotent)
if [ -d "$WORKSPACE_ROOT/tnos_venv" ]; then
    echo "Removing legacy tnos_venv virtual environment..."
    rm -rf "$WORKSPACE_ROOT/tnos_venv"
fi

# Check if Go is installed and in the PATH
if ! command -v go &> /dev/null; then
    echo "Go is not found in your PATH. Looking for alternative locations..."
    
    # Check common Go installation locations
    if [ -d "/usr/local/go/bin" ]; then
        export PATH="/usr/local/go/bin:$PATH"
        echo "Found Go in /usr/local/go/bin - added to PATH"
    elif [ -d "$HOME/go/bin" ]; then
        export PATH="$HOME/go/bin:$PATH"
        echo "Found Go tools in $HOME/go/bin - added to PATH"
    elif [ -d "/opt/homebrew/bin" ] && [ -f "/opt/homebrew/bin/go" ]; then
        export PATH="/opt/homebrew/bin:$PATH"
        echo "Found Go in Homebrew - added to PATH"
    else
        echo "WARNING: Go was not found. Some features may not work properly."
        # Continue anyway, as we'll use what tools we can find
    fi
fi

# Set Go environment variables if Go is now available
if command -v go &> /dev/null; then
    GOROOT=$(go env GOROOT)
    export GOROOT
    export PATH="$GOROOT/bin:$PATH"
    export GO111MODULE=on
    echo "Go version: $(go version)"
else
    echo "Go version: Not available"
fi

# Robustly detect the workspace root (directory containing this script's parent)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
export WORKSPACE_ROOT

# Ensure all root path references use the canonical workspace root
# Set go.work file location (absolute path)
export GOWORK="$WORKSPACE_ROOT/go.work"

if [ ! -f "$GOWORK" ]; then
    echo "ERROR: go.work file not found at $GOWORK. Please ensure it exists at the workspace root."
    exit 1
fi

# Do NOT set GOPATH to the project directory as it causes issues
# Instead, use workspace mode with GOWORK

# MCP server doesn't directly need Gradle
# Skip Gradle daemon initialization to avoid terminal pollution

# Detect Go bin directory
if command -v go &> /dev/null; then
    GOBIN=$(go env GOBIN)
    if [ -z "$GOBIN" ]; then
        GOBIN="$HOME/go/bin"
    fi
else
    # If go command is not available, use standard location
    GOBIN="$HOME/go/bin"
fi
export PATH="$GOBIN:$PATH"

# Create go/bin directory if it doesn't exist
mkdir -p "$GOBIN"

# Install Go tools if needed
echo "Ensuring necessary Go tools are installed..."
GOPLS_PATH=$(command -v gopls 2>/dev/null)
if [ -z "$GOPLS_PATH" ]; then
    echo "Installing gopls..."
    go install golang.org/x/tools/gopls@latest
    GOPLS_PATH=$(command -v gopls 2>/dev/null)
fi
if [ -n "$GOPLS_PATH" ]; then
    echo "gopls is installed at: $GOPLS_PATH"
else
    echo "ERROR: gopls installation failed or not found in PATH!"
fi
which golint > /dev/null 2>&1 || { echo "Installing golint..."; go install golang.org/x/lint/golint@latest; }
which errcheck > /dev/null 2>&1 || { echo "Installing errcheck..."; go install github.com/kisielk/errcheck@latest; }

# Move to the github-mcp-server directory
cd "$WORKSPACE_ROOT/github-mcp-server" || exit 1

# Verify module setup
echo "Verifying Go module setup..."
echo "Current directory: $(pwd)"
echo "Current module: $(go list -m)"

# Clean Go module cache for this project if there are issues
# go clean -modcache

echo "Go version: $(go version)"
echo "Gopls version: $(gopls version 2>/dev/null || echo "gopls not installed")"

# Set environment variables for the GitHub MCP Server
export MCP_SERVER_PORT=10617
export MCP_LOG_FILE="$WORKSPACE_ROOT/logs/github_mcp_server.log"

echo
echo "Development environment set up successfully!"
echo
echo "You can now run the following commands:"
echo "  - Build the server: go build -o bin/github-mcp-server ./cmd/server"
echo "  - Run the server: ./bin/github-mcp-server"
echo "  - Run tests: go test ./..."
echo
echo "OR use these shortcuts:"
echo "  - Build and run: ./setup_dev_env.sh run"
echo "  - Run tests: ./setup_dev_env.sh test"
echo

# Handle command line arguments
if [ "$1" = "run" ]; then
    echo "Building and running GitHub MCP server..."
    # Fix mod dependencies first
    echo "Running go mod tidy..."
    go mod tidy
    
    # Build with explicit workspace mode
    echo "Building with workspace mode..."
    go build -o bin/github-mcp-server ./cmd/server
    
    if [ $? -eq 0 ]; then
        echo "Build successful. Running server..."
        ./bin/github-mcp-server
    else
        echo "Build failed. Trying alternative approach..."
        # Try building without workspace mode, using direct module
        SERVER_DIR="$WORKSPACE_ROOT/github-mcp-server"
        cd "$SERVER_DIR" && \
        GO111MODULE=on go build -o bin/github-mcp-server ./cmd/server && \
        ./bin/github-mcp-server
    fi
elif [ "$1" = "test" ]; then
    echo "Running tests..."
    go test ./...
# Ensure the script exits after starting all components
elif [ "$1" = "start-all" ]; then
    echo "Starting all MCP components in one script..."
    # Stop any existing processes
    pkill -f "github-mcp-server" 2>/dev/null
    pkill -f "tnos_mcp_server.py" 2>/dev/null
    pkill -f "github_mcp_bridge.py" 2>/dev/null
    sleep 2

    # Start GitHub MCP Server (Go binary)
    echo "Starting GitHub MCP Server..."
    "$WORKSPACE_ROOT/bin/github-mcp-server" &
    GITHUB_MCP_PID=$!
    sleep 2

    # Start TNOS MCP Server (Python)
    echo "Starting TNOS MCP Server..."
    PYTHON311="$WORKSPACE_ROOT/systems/python/venv311/bin/python3.11"
    TNOS_MCP_SERVER_PY="$WORKSPACE_ROOT/mcp/bridge/tnos_mcp_server.py"
    "$PYTHON311" "$TNOS_MCP_SERVER_PY" &
    TNOS_MCP_PID=$!
    sleep 2

    # Start MCP Bridge (Python)
    echo "Starting MCP Bridge..."
    MCP_BRIDGE_PY="$WORKSPACE_ROOT/github-mcp-server/internal/bridge/github_mcp_bridge.py"
    "$PYTHON311" "$MCP_BRIDGE_PY" --log-level info &
    MCP_BRIDGE_PID=$!
    sleep 2

    echo "All MCP components started. PIDs:"
    echo "  GitHub MCP Server: $GITHUB_MCP_PID"
    echo "  TNOS MCP Server: $TNOS_MCP_PID"
    echo "  MCP Bridge: $MCP_BRIDGE_PID"
    echo "To stop all, use: pkill -f github-mcp-server; pkill -f tnos_mcp_server.py; pkill -f github_mcp_bridge.py"
    exit 0
fi

# Ensure the GitHub MCP Server binary is built
echo "Building GitHub MCP Server binary..."
go build -o "$WORKSPACE_ROOT/bin/github-mcp-server" "$WORKSPACE_ROOT/github-mcp-server/cmd/server"
# Also copy to the main bin directory for the task to find it
cp "$WORKSPACE_ROOT/github-mcp-server/bin/github-mcp-server" "$WORKSPACE_ROOT/bin/github-mcp-server" 2>/dev/null || true

# If not using start-all, use the dedicated script
echo "To start the GitHub MCP Server, run: bash $WORKSPACE_ROOT/scripts/shell/start_github_mcp_server.sh"

# === MCP Integration Management (Quantum Symmetry Aligned) ===
# WHO: DevEnvSetup
# WHAT: Start/stop/status for all MCP integration components, enforcing quantum symmetry
# WHEN: During dev setup, run, or stop
# WHERE: System Layer 6 (Integration)
# WHY: To ensure all MCP components are managed from a single canonical script, with mirrored and entangled logic
# HOW: In-place shell logic, symmetric PID/port checks, no new files
# EXTENT: All MCP integration processes, with explicit symmetry and entanglement

# Quantum Symmetry Principle: The following functions are mirrored (start/stop/status) and operate on all MCP components in a symmetric, entangled fashion. Each component's state is checked and managed using both PID and port (input/output) to ensure reversible, loss-aware operations. Any change in one module (start/stop) is reflected in its paired status logic.

# === Utility: Wait for Port to be Open ===
wait_for_port() {
    # WHO: PortWaiter
    # WHAT: Waits for a TCP port to be open and listening
    # WHEN: Before starting dependent MCP components
    # WHERE: System Layer 6 (Integration)
    # WHY: To ensure reliable MCP startup order and avoid race conditions
    # HOW: Uses netcat (nc) in a loop with timeout
    # EXTENT: Used for all MCP component port dependencies
    local host="$1"
    local port="$2"
    local timeout="${3:-60}"
    local waited=0
    echo "[MCP][PortWaiter] Waiting for $host:$port to be open (timeout ${timeout}s)..."
    while ! nc -z "$host" "$port" 2>/dev/null; do
        sleep 1
        waited=$((waited+1))
        if [ "$waited" -ge "$timeout" ]; then
            echo "[MCP][PortWaiter][ERROR] Timeout waiting for $host:$port to be open after $timeout seconds."
            return 1
        fi
        if (( $waited % 10 == 0 )); then
            echo "[MCP][PortWaiter] Still waiting for $host:$port... ($waited seconds elapsed)"
        fi
    done
    echo "[MCP][PortWaiter] $host:$port is now open after $waited seconds."
    return 0
}

# === SOLO DEVELOPER MCP STARTUP LOGIC ===
# Only start the TNOS MCP server (Python) directly, no other shell scripts or Go servers
start_tnos_mcp_server_only() {
    echo "[SOLO] Stopping any existing TNOS MCP server processes..."
    pkill -f "tnos_mcp_server.py"
    sleep 1
    PROJECT_ROOT="$WORKSPACE_ROOT"
    LOGS_DIR="$PROJECT_ROOT/logs"
    MCP_SERVER_SCRIPT="$PROJECT_ROOT/mcp/bridge/tnos_mcp_server.py"
    VENV_DIR="/Users/Jubicudis/Tranquility-Neuro-OS/systems/python/venv311"
    CANONICAL_MCP_PATH="$PROJECT_ROOT/mcp/bridge/python"
    export PYTHONPATH="$CANONICAL_MCP_PATH:$PYTHONPATH"
    mkdir -p "$LOGS_DIR"
    if [[ ! -f "$MCP_SERVER_SCRIPT" ]]; then
      echo "[SOLO][ERROR] TNOS MCP server script not found!"
      exit 1
    fi
    if lsof -i :8083 | grep LISTEN; then
      echo "[SOLO][ERROR] Port 8083 (TCP) is already in use. TNOS MCP server will not be started."
      exit 1
    fi
    if [[ -f "$VENV_DIR/bin/activate" ]]; then
      source "$VENV_DIR/bin/activate"
      "$VENV_DIR/bin/python3.11" -m pip install flask --quiet
    fi
    (cd "$PROJECT_ROOT" && nohup "$VENV_DIR/bin/python3.11" -u "$MCP_SERVER_SCRIPT" > "$LOGS_DIR/tnos_mcp_server.log" 2>&1 & echo $! > "$LOGS_DIR/tnos_mcp_server.pid")
    sleep 2
    TNOS_PID=$(cat "$LOGS_DIR/tnos_mcp_server.pid")
    if ! kill -0 $TNOS_PID 2>/dev/null; then
      echo "[SOLO][ERROR] TNOS MCP server process failed to start (PID $TNOS_PID not running). Check $LOGS_DIR/tnos_mcp_server.log for details."
      exit 1
    fi
    if ! wait_for_port_and_health "127.0.0.1" 8083 60; then
      echo "[SOLO][ERROR] TNOS MCP Server did not open port 8083 or /health endpoint in time. Checking process..."
      if ! kill -0 $TNOS_PID 2>/dev/null; then
        echo "[SOLO][ERROR] TNOS MCP server process died before port 8083 opened. Check $LOGS_DIR/tnos_mcp_server.log for errors."
      else
        echo "[SOLO][ERROR] TNOS MCP server process is running but port 8083 is not open or /health is not ready. Possible binding or Flask startup issue."
      fi
      exit 1
    fi
    if ! wait_for_port "127.0.0.1" 9001 20; then
      echo "[SOLO][ERROR] TNOS MCP Server did not open port 9001 (WebSocket) in time. Checking process..."
      if ! kill -0 $TNOS_PID 2>/dev/null; then
        echo "[SOLO][ERROR] TNOS MCP server process died before port 9001 opened. Check $LOGS_DIR/tnos_mcp_server.log for errors."
      else
        echo "[SOLO][ERROR] TNOS MCP server process is running but port 9001 is not open. Possible binding issue."
      fi
      exit 1
    fi
    echo "[SOLO] TNOS MCP server started with PID $TNOS_PID (log: $LOGS_DIR/tnos_mcp_server.log)"
}

start_all_mcp() {
    echo "[MCP] Stopping any existing MCP processes..."
    pkill -f "github-mcp-server"
    pkill -f "tnos_mcp_server.py"
    pkill -f "tnos_mcp_bridge.py"
    pkill -f "enhanced_mobius_visualization_server.py"
    sleep 2

    echo "[MCP] Building GitHub MCP Server..."
    # Canonical Go build location
    cd "$WORKSPACE_ROOT/github-mcp-server" || exit 1
    go build -o "$WORKSPACE_ROOT/bin/github-mcp-server" ./cmd/server || { echo "[MCP] Build failed."; exit 1; }
    SERVER_BIN="$WORKSPACE_ROOT/bin/github-mcp-server"
    if [[ ! -f "$SERVER_BIN" ]]; then
      echo "ERROR: GitHub MCP server binary not found at $SERVER_BIN."
      exit 1
    fi
    LOGS_DIR="$WORKSPACE_ROOT/logs"
    mkdir -p "$LOGS_DIR"
    if lsof -i :10617 | grep LISTEN; then
      echo "ERROR: Port 10617 is already in use. GitHub MCP server will not be started."
    else
      nohup "$SERVER_BIN" > "$LOGS_DIR/github_mcp_server.log" 2>&1 &
      echo $! > "$LOGS_DIR/github_mcp_server.pid"
      echo "GitHub MCP server started with PID $(cat "$LOGS_DIR/github_mcp_server.pid") on port 10617 (log: $LOGS_DIR/github_mcp_server.log)"
      wait_for_port "127.0.0.1" 10617 20 || { echo "[MCP][ERROR] GitHub MCP Server did not open port 10617 in time."; exit 1; }
    fi
    sleep 2

    # Start TNOS MCP Server (canonical location)
    echo "[MCP] Starting TNOS MCP Server on port 8083..."
    MCP_SERVER_SCRIPT="$WORKSPACE_ROOT/mcp/bridge/tnos_mcp_server.py"
    VENV_DIR="$WORKSPACE_ROOT/systems/python/venv311"
    CANONICAL_MCP_PATH="$WORKSPACE_ROOT/mcp/bridge/python"
    export PYTHONPATH="$CANONICAL_MCP_PATH:$PYTHONPATH"
    if [[ ! -f "$MCP_SERVER_SCRIPT" ]]; then
      echo "ERROR: TNOS MCP server script not found at $MCP_SERVER_SCRIPT!"
    else
      if lsof -i :8083 | grep LISTEN; then
        echo "ERROR: Port 8083 (TCP) is already in use. TNOS MCP server will not be started."
      else
        if [[ -f "$VENV_DIR/bin/activate" ]]; then
          source "$VENV_DIR/bin/activate"
          "$VENV_DIR/bin/python3.11" -m pip install flask --quiet
        fi
        (cd "$WORKSPACE_ROOT" && nohup "$VENV_DIR/bin/python3.11" -u "$MCP_SERVER_SCRIPT" > "$LOGS_DIR/tnos_mcp_server.log" 2>&1 & echo $! > "$LOGS_DIR/tnos_mcp_server.pid")
        sleep 2
        TNOS_PID=$(cat "$LOGS_DIR/tnos_mcp_server.pid")
        if ! kill -0 $TNOS_PID 2>/dev/null; then
          echo "[MCP][ERROR] TNOS MCP server process failed to start (PID $TNOS_PID not running). Check $LOGS_DIR/tnos_mcp_server.log for details."
          exit 1
        fi
        if ! wait_for_port_and_health "127.0.0.1" 8083 60; then
          echo "[MCP][ERROR] TNOS MCP Server did not open port 8083 or /health endpoint in time. Checking process..."
          if ! kill -0 $TNOS_PID 2>/dev/null; then
            echo "[MCP][ERROR] TNOS MCP server process died before port 8083 opened. Check $LOGS_DIR/tnos_mcp_server.log for errors."
          else
            echo "[MCP][ERROR] TNOS MCP server process is running but port 8083 is not open or /health is not ready. Possible binding or Flask startup issue."
          fi
          exit 1
        fi
        if ! wait_for_port "127.0.0.1" 9001 20; then
          echo "[MCP][ERROR] TNOS MCP Server did not open port 9001 (WebSocket) in time. Checking process..."
          if ! kill -0 $TNOS_PID 2>/dev/null; then
            echo "[MCP][ERROR] TNOS MCP server process died before port 9001 opened. Check $LOGS_DIR/tnos_mcp_server.log for errors."
          else
            echo "[MCP][ERROR] TNOS MCP server process is running but port 9001 is not open. Possible binding issue."
          fi
          exit 1
        fi
        echo "TNOS MCP server started with PID $TNOS_PID (log: $LOGS_DIR/tnos_mcp_server.log)"
      fi
    fi
    sleep 2

    # Start MCP Bridge (canonical location)
    echo "[MCP] Starting MCP Bridge between GitHub port 10617 and TNOS port 9001..."
    PYTHON_BRIDGE="$WORKSPACE_ROOT/github-mcp-server/internal/bridge/github_mcp_bridge.py"
    if [[ ! -f "$PYTHON_BRIDGE" ]]; then
      echo "[ERROR] Python bridge script not found: $PYTHON_BRIDGE"
      exit 1
    fi
    nohup "$VENV_DIR/bin/python3.11" "$PYTHON_BRIDGE" --port 10619 > "$LOGS_DIR/mcp_bridge.log" 2>&1 &
    echo $! > "$LOGS_DIR/mcp_bridge.pid"
    echo "MCP Bridge started with PID $(cat "$LOGS_DIR/mcp_bridge.pid") (log: $LOGS_DIR/mcp_bridge.log)"
    # No port to wait for, but log connection attempt

    echo "[MCP] All MCP components started."
    echo "[MCP] Visualization is now provided by the TNOS MCP server as an internal component."
}

stop_all_mcp() {
    echo "[MCP] Stopping all MCP components..."
    pkill -f "github-mcp-server"
    pkill -f "tnos_mcp_server.py"
    pkill -f "tnos_mcp_bridge.py"
    pkill -f "enhanced_mobius_visualization_server.py"
    echo "[MCP] All MCP components stopped."
}

status_all_mcp() {
    echo "[MCP] MCP Component Status (Quantum Symmetry):"
    LOGS_DIR="$WORKSPACE_ROOT/logs"
    # Mirrored status checks: PID and port for each component
    # GitHub MCP Server
    if [[ -f "$LOGS_DIR/github_mcp_server.pid" ]]; then
      G_PID=$(cat "$LOGS_DIR/github_mcp_server.pid")
      if kill -0 $G_PID 2>/dev/null && lsof -i :10617 | grep LISTEN >/dev/null; then
        echo "  [RUNNING] GitHub MCP Server (PID $G_PID, port 10617)"
      else
        echo "  [STOPPED] GitHub MCP Server"
      fi
    else
      echo "  [STOPPED] GitHub MCP Server"
    fi
    # TNOS MCP Server (WebSocket)
    if [[ -f "$LOGS_DIR/tnos_mcp_server.pid" ]]; then
      T_PID=$(cat "$LOGS_DIR/tnos_mcp_server.pid")
      if kill -0 $T_PID 2>/dev/null && lsof -i :9001 | grep LISTEN >/dev/null; then
        echo "  [RUNNING] TNOS MCP Server (PID $T_PID, WebSocket port 9001)"
      else
        echo "  [STOPPED] TNOS MCP Server"
      fi
    else
      echo "  [STOPPED] TNOS MCP Server"
    fi
    # MCP Bridge
    if [[ -f "$LOGS_DIR/mcp_bridge.pid" ]]; then
      B_PID=$(cat "$LOGS_DIR/mcp_bridge.pid")
      if kill -0 $B_PID 2>/dev/null && lsof -i :10619 | grep LISTEN >/dev/null; then
        echo "  [RUNNING] MCP Bridge (PID $B_PID, port 10619)"
      else
        echo "  [STOPPED] MCP Bridge"
      fi
    else
      echo "  [STOPPED] MCP Bridge"
    fi
    # Visualization (internal)
    echo "  [INFO] Visualization is now provided by the TNOS MCP server as an internal component on port 8083."
}

# Enhanced solo developer port/health check for TNOS MCP server
wait_for_port_and_health() {
    local host="$1"
    local port="$2"
    local timeout="${3:-60}"
    local waited=0
    local health_url="http://$host:$port/health"
    echo "[SOLO][PortWaiter] Waiting for $host:$port to be open (timeout ${timeout}s)..."
    while ! nc -z "$host" "$port" 2>/dev/null; do
        sleep 1
        waited=$((waited+1))
        if [ "$waited" -ge "$timeout" ]; then
            echo "[SOLO][PortWaiter][ERROR] Timeout waiting for $host:$port to be open after $timeout seconds."
            return 1
        fi
        if (( $waited % 10 == 0 )); then
            echo "[SOLO][PortWaiter] Still waiting for $host:$port... ($waited seconds elapsed)"
        fi
    done
    echo "[SOLO][PortWaiter] $host:$port is now open after $waited seconds. Checking /health endpoint..."
    # Now check /health endpoint for Flask readiness
    local health_waited=0
    while true; do
        if command -v curl >/dev/null 2>&1; then
            local status=$(curl -s -o /dev/null -w "%{http_code}" "$health_url")
            if [ "$status" = "200" ]; then
                echo "[SOLO][PortWaiter] /health endpoint is ready."
                return 0
            fi
        fi
        sleep 1
        health_waited=$((health_waited+1))
        if [ "$health_waited" -ge "$timeout" ]; then
            echo "[SOLO][PortWaiter][ERROR] Timeout waiting for $health_url to return 200 after $timeout seconds."
            return 1
        fi
        if (( $health_waited % 10 == 0 )); then
            echo "[SOLO][PortWaiter] Still waiting for $health_url... ($health_waited seconds elapsed)"
        fi
    done
}

# Add CLI argument handling for MCP integration (mirrored logic)
if [ "$1" = "start-all" ]; then
    start_all_mcp
    status_all_mcp
    exit 0
elif [ "$1" = "stop-all" ]; then
    stop_all_mcp
    status_all_mcp
    exit 0
elif [ "$1" = "status-all" ]; then
    status_all_mcp
    exit 0
fi

# CLI argument for solo developer mode
if [ "$1" = "solo" ]; then
    start_tnos_mcp_server_only
    exit 0
fi

# If no arguments are provided, start all MCP components by default (symmetric default)
if [ $# -eq 0 ]; then
    start_all_mcp
    status_all_mcp
    exit 0
fi

# Ensure PYTHONPATH is set for this shell session
if [ "$0" = "$BASH_SOURCE" ]; then
    echo "[WARNING] setup_dev_env.sh was executed, not sourced. PYTHONPATH and other environment variables will NOT persist in your current shell."
    echo "To ensure TNOS Python modules are always available, run:"
    echo "    source $0"
fi
