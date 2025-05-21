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
cd "/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server" || exit 1

# Verify module setup
echo "Verifying Go module setup..."
echo "Current directory: $(pwd)"
echo "Current module: $(go list -m)"

# Clean Go module cache for this project if there are issues
# go clean -modcache

echo "Go version: $(go version)"
echo "Gopls version: $(gopls version 2>/dev/null || echo "gopls not installed")"

# Set environment variables for the GitHub MCP Server
export MCP_SERVER_PORT=8889
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
        SERVER_DIR="/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server"
        cd "$SERVER_DIR" && \
        GO111MODULE=on go build -o bin/github-mcp-server ./cmd/server && \
        ./bin/github-mcp-server
    fi
elif [ "$1" = "test" ]; then
    echo "Running tests..."
    go test ./...
# Ensure the script exits after starting all components
elif [ "$1" = "start-all" ]; then
    echo "Starting all MCP components..."

    # First kill any existing instances
    echo "Stopping any existing MCP processes..."
    pkill -f "github-mcp-server"
    pkill -f "tnos_mcp_server.py"
    pkill -f "tnos_mcp_bridge.py"
    pkill -f "enhanced_mobius_visualization_server.py"
    sleep 2
    
    # Build the GitHub MCP Server
    echo "Building GitHub MCP Server..."
    
    # Check if Go is available
    if ! command -v go &> /dev/null; then
        echo "ERROR: Go compiler is not available in your PATH."
        echo "Please install Go from https://golang.org/dl/"
        echo "For macOS users, you can install Go using Homebrew: brew install go"
        exit 1
    fi
    
    go build -o ./bin/github-mcp-server ./cmd/server
    if [ $? -ne 0 ]; then
        echo "Build failed. Please check the Go source code for errors."
        exit 1
    fi
    
    # Copy binary to workspace bin directory
    cp -f ./bin/github-mcp-server "$WORKSPACE_ROOT/bin/github-mcp-server" 2>/dev/null
    
    # Start all MCP components in sequence using the shell scripts
    echo "Starting all MCP components using shell scripts..."
    
    # Start GitHub MCP Server on custom port 8889
    echo "Starting GitHub MCP Server on port 8889..."
    "$WORKSPACE_ROOT/bin/github-mcp-server" -config "$WORKSPACE_ROOT/github-mcp-server/config/tools.json" &
    sleep 2
    
    # Start TNOS MCP Server on port 8083
    echo "Starting TNOS MCP Server on port 8083..."
    if [ -f "$WORKSPACE_ROOT/scripts/shell/start_tnos_mcp_server.sh" ]; then
        bash "$WORKSPACE_ROOT/scripts/shell/start_tnos_mcp_server.sh"
    else
        # Fallback: launch directly using venv
        MCP_SERVER_SCRIPT="$WORKSPACE_ROOT/mcp/bridge/tnos_mcp_server.py"
        VENV_DIR="$WORKSPACE_ROOT/venv"
        if [ -f "$MCP_SERVER_SCRIPT" ]; then
            echo "[FALLBACK] Launching TNOS MCP Server directly with venv..."
            (cd "$WORKSPACE_ROOT" && nohup "$VENV_DIR/bin/python3.12" -u "$MCP_SERVER_SCRIPT" > "$WORKSPACE_ROOT/logs/tnos_mcp_server.log" 2>&1 &)
            echo $! > "$WORKSPACE_ROOT/logs/tnos_mcp_server.pid"
        else
            echo "ERROR: TNOS MCP server script not found!"
        fi
    fi
    sleep 2
    
    # Start Enhanced Visualization Server
    echo "Starting Enhanced Möbius Visualization Server on port 7779..."
    if [ -f "$WORKSPACE_ROOT/scripts/shell/start_visualization_server.sh" ]; then
        bash "$WORKSPACE_ROOT/scripts/shell/start_visualization_server.sh"
    else
        # Fallback: launch directly using venv
        VISUALIZATION_SERVER_SCRIPT="$WORKSPACE_ROOT/mcp/bridge/visualization/enhanced_mobius_visualization_server.py"
        VENV_DIR="$WORKSPACE_ROOT/venv"
        if [ -f "$VISUALIZATION_SERVER_SCRIPT" ]; then
            echo "[FALLBACK] Launching Visualization Server directly with venv..."
            (cd "$WORKSPACE_ROOT" && nohup "$VENV_DIR/bin/python3.12" "$VISUALIZATION_SERVER_SCRIPT" > "$WORKSPACE_ROOT/logs/enhanced_visualization_server.log" 2>&1 &)
            echo $! > "$WORKSPACE_ROOT/logs/enhanced_visualization_server.pid"
        else
            echo "ERROR: Enhanced Visualization server script not found!"
        fi
    fi
    sleep 2
    
    # Start MCP Bridge connecting GitHub and TNOS MCP servers
    echo "Starting MCP Bridge between GitHub port 8889 and TNOS port 8083..."
    # Set as environment variables instead of flags
    export GITHUB_PORT=8889
    export TNOS_PORT=8083
    bash "$WORKSPACE_ROOT/scripts/shell/start_mcp_bridge.sh"
    
    echo "All MCP components started successfully."
    echo "Visualization server available at: http://localhost:7779/"
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
    local timeout="${3:-20}"
    local waited=0
    echo "[MCP][PortWaiter] Waiting for $host:$port to be open (timeout ${timeout}s)..."
    while ! nc -z "$host" "$port" 2>/dev/null; do
        sleep 1
        waited=$((waited+1))
        if [ "$waited" -ge "$timeout" ]; then
            echo "[MCP][PortWaiter][ERROR] Timeout waiting for $host:$port to be open after $timeout seconds."
            return 1
        fi
    done
    echo "[MCP][PortWaiter] $host:$port is now open."
    return 0
}

start_all_mcp() {
    echo "[MCP] Stopping any existing MCP processes..."
    pkill -f "github-mcp-server"
    pkill -f "tnos_mcp_server.py"
    pkill -f "tnos_mcp_bridge.py"
    pkill -f "enhanced_mobius_visualization_server.py"
    sleep 2

    echo "[MCP] Building GitHub MCP Server..."
    go build -o ./bin/github-mcp-server ./cmd/server || { echo "[MCP] Build failed."; exit 1; }
    cp -f ./bin/github-mcp-server "$WORKSPACE_ROOT/bin/github-mcp-server" 2>/dev/null

    # Start GitHub MCP Server (mirrored logic)
    echo "[MCP] Starting GitHub MCP Server on port 8889..."
    PROJECT_ROOT="$WORKSPACE_ROOT"
    LOGS_DIR="$PROJECT_ROOT/logs"
    SERVER_LOG="$LOGS_DIR/github_mcp_server.log"
    SERVER_BIN="$PROJECT_ROOT/github-mcp-server/bin/github-mcp-server"
    SERVER_BIN_ALT="$PROJECT_ROOT/bin/github-mcp-server"
    if [[ -f "$SERVER_BIN" ]]; then
      GITHUB_MCP_SERVER="$SERVER_BIN"
    elif [[ -f "$SERVER_BIN_ALT" ]]; then
      GITHUB_MCP_SERVER="$SERVER_BIN_ALT"
    else
      echo "ERROR: GitHub MCP server binary not found."
      exit 1
    fi
    mkdir -p "$LOGS_DIR"
    if lsof -i :8889 | grep LISTEN; then
      echo "ERROR: Port 8889 is already in use. GitHub MCP server will not be started."
    else
      nohup "$GITHUB_MCP_SERVER" > "$SERVER_LOG" 2>&1 &
      echo $! > "$LOGS_DIR/github_mcp_server.pid"
      echo "GitHub MCP server started with PID $(cat "$LOGS_DIR/github_mcp_server.pid") on port 8889 (log: $SERVER_LOG)"
      wait_for_port "127.0.0.1" 8889 20 || { echo "[MCP][ERROR] GitHub MCP Server did not open port 8889 in time."; exit 1; }
    fi
    sleep 2

    # Start TNOS MCP Server (mirrored logic)
    echo "[MCP] Starting TNOS MCP Server on port 8083..."
    MCP_SERVER_SCRIPT="$PROJECT_ROOT/mcp/bridge/tnos_mcp_server.py"
    VENV_DIR="$PROJECT_ROOT/venv"
    if [[ ! -f "$MCP_SERVER_SCRIPT" ]]; then
      echo "ERROR: TNOS MCP server script not found!"
    else
      if lsof -i :8083 | grep LISTEN; then
        echo "ERROR: Port 8083 (TCP) is already in use. TNOS MCP server will not be started."
      else
        if [[ -f "$VENV_DIR/bin/activate" ]]; then
          source "$VENV_DIR/bin/activate"
          "$VENV_DIR/bin/pip" install flask --quiet
        fi
        export PYTHONPATH="$PROJECT_ROOT/python:$PROJECT_ROOT:$PROJECT_ROOT/core:$PROJECT_ROOT/github-mcp-server:$PYTHONPATH"
        (cd "$PROJECT_ROOT" && nohup "$VENV_DIR/bin/python3.12" -u "$MCP_SERVER_SCRIPT" > "$LOGS_DIR/tnos_mcp_server.log" 2>&1 &)
        echo $! > "$LOGS_DIR/tnos_mcp_server.pid"
        echo "TNOS MCP server started with PID $(cat "$LOGS_DIR/tnos_mcp_server.pid") (log: $LOGS_DIR/tnos_mcp_server.log)"
        wait_for_port "127.0.0.1" 8083 20 || { echo "[MCP][ERROR] TNOS MCP Server did not open port 8083 in time."; exit 1; }
      fi
    fi
    sleep 2

    # Start Enhanced Möbius Visualization Server (mirrored logic)
    echo "[MCP] Starting Enhanced Möbius Visualization Server on port 7779..."
    VISUALIZATION_SERVER_SCRIPT="$PROJECT_ROOT/mcp/bridge/visualization/enhanced_mobius_visualization_server.py"
    if [[ ! -f "$VISUALIZATION_SERVER_SCRIPT" ]]; then
      echo "ERROR: Enhanced Visualization server script not found!"
    else
      if lsof -i :7779 | grep LISTEN; then
        echo "ERROR: Port 7779 is already in use. Enhanced Visualization server will not be started."
      else
        if [[ -f "$VENV_DIR/bin/python" ]]; then
          PYTHON_EXEC="$VENV_DIR/bin/python"
          "$VENV_DIR/bin/pip" install flask --quiet
        else
          PYTHON_EXEC=$(which python3)
          pip3 install flask --quiet
        fi
        nohup $PYTHON_EXEC "$VISUALIZATION_SERVER_SCRIPT" > "$LOGS_DIR/enhanced_visualization_server.log" 2>&1 &
        echo $! > "$LOGS_DIR/enhanced_visualization_server.pid"
        echo "Enhanced Möbius Visualization Server started on port 7779 (log: $LOGS_DIR/enhanced_visualization_server.log)"
        wait_for_port "127.0.0.1" 7779 20 || { echo "[MCP][ERROR] Visualization Server did not open port 7779 in time."; exit 1; }
      fi
    fi
    sleep 2

    # Start MCP Bridge (mirrored logic)
    echo "[MCP] Starting MCP Bridge between GitHub port 8889 and TNOS port 8083..."
    PYTHON_BRIDGE="$PROJECT_ROOT/mcp/bridge/tnos_mcp_bridge.py"
    if [[ ! -f "$PYTHON_BRIDGE" ]]; then
      echo "ERROR: MCP Bridge script not found!"
    else
      echo "[MCP][Bridge] Attempting to connect MCP Bridge to GitHub MCP (127.0.0.1:8889) and TNOS MCP (127.0.0.1:8083)..."
      nohup "$VENV_DIR/bin/python3.12" "$PYTHON_BRIDGE" > "$LOGS_DIR/mcp_bridge.log" 2>&1 &
      echo $! > "$LOGS_DIR/mcp_bridge.pid"
      echo "MCP Bridge started with PID $(cat "$LOGS_DIR/mcp_bridge.pid") (log: $LOGS_DIR/mcp_bridge.log)"
      # No port to wait for, but log connection attempt
    fi
    sleep 2

    echo "[MCP] All MCP components started."
    echo "[MCP] Visualization server: http://localhost:7779/"
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
      if kill -0 $G_PID 2>/dev/null && lsof -i :8889 | grep LISTEN >/dev/null; then
        echo "  [RUNNING] GitHub MCP Server (PID $G_PID, port 8889)"
      else
        echo "  [STOPPED] GitHub MCP Server"
      fi
    else
      echo "  [STOPPED] GitHub MCP Server"
    fi
    # TNOS MCP Server
    if [[ -f "$LOGS_DIR/tnos_mcp_server.pid" ]]; then
      T_PID=$(cat "$LOGS_DIR/tnos_mcp_server.pid")
      if kill -0 $T_PID 2>/dev/null && lsof -i :8083 | grep LISTEN >/dev/null; then
        echo "  [RUNNING] TNOS MCP Server (PID $T_PID, port 8083)"
      else
        echo "  [STOPPED] TNOS MCP Server"
      fi
    else
      echo "  [STOPPED] TNOS MCP Server"
    fi
    # MCP Bridge
    if [[ -f "$LOGS_DIR/mcp_bridge.pid" ]]; then
      B_PID=$(cat "$LOGS_DIR/mcp_bridge.pid")
      if kill -0 $B_PID 2>/dev/null; then
        echo "  [RUNNING] MCP Bridge (PID $B_PID)"
      else
        echo "  [STOPPED] MCP Bridge"
      fi
    else
      echo "  [STOPPED] MCP Bridge"
    fi
    # Möbius Visualization Server
    if [[ -f "$LOGS_DIR/enhanced_visualization_server.pid" ]]; then
      V_PID=$(cat "$LOGS_DIR/enhanced_visualization_server.pid")
      if kill -0 $V_PID 2>/dev/null && lsof -i :7779 | grep LISTEN >/dev/null; then
        echo "  [RUNNING] Möbius Visualization Server (PID $V_PID, port 7779)"
      else
        echo "  [STOPPED] Möbius Visualization Server"
      fi
    else
      echo "  [STOPPED] Möbius Visualization Server"
    fi
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

# If no arguments are provided, start all MCP components by default (symmetric default)
if [ $# -eq 0 ]; then
    start_all_mcp
    status_all_mcp
    exit 0
fi
