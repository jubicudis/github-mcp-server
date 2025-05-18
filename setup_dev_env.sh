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
