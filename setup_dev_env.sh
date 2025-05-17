#!/bin/bash

# WHO: DevEnvSetup
# WHAT: Development environment configuration for GitHub MCP server
# WHEN: During development setup and environment initialization
# WHERE: System Layer 6 (Integration)
# WHY: To ensure consistent Go module resolution and tool configuration
# HOW: Using shell environment variables and workspace paths
# EXTENT: GitHub MCP server development operations

echo "Setting up Go development environment for GitHub MCP Server..."

# Set Go environment variables
GOROOT=$(go env GOROOT)
export GOROOT
export PATH="$GOROOT/bin:$PATH"
export GO111MODULE=on

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
GOBIN=$(go env GOBIN)
if [ -z "$GOBIN" ]; then
    GOBIN="$HOME/go/bin"
fi
export PATH="$GOBIN:$PATH"

# Create go/bin directory if it doesn't exist
mkdir -p "$GOPATH/bin"

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
    bash "$WORKSPACE_ROOT/scripts/shell/start_github_mcp_server.sh" --port=8889
    sleep 2
    
    # Start TNOS MCP Server on port 8083
    echo "Starting TNOS MCP Server on port 8083..."
    bash "$WORKSPACE_ROOT/scripts/shell/start_tnos_mcp_server.sh"
    sleep 2
    
    # Start Enhanced Visualization Server
    echo "Starting Enhanced Möbius Visualization Server on port 7779..."
    bash "$WORKSPACE_ROOT/scripts/shell/start_visualization_server.sh"
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
