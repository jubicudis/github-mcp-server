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
export GOFLAGS="-mod=mod"

# Set up workspace-specific GOPATH pointing to the project's location
export GOPATH="/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS"
export PATH="$GOPATH/bin:$PATH"

# Ensure the Gradle daemon is running for Java components
GRADLE_DAEMON_SCRIPT="${GOPATH}/scripts/shell/gradle_persistent_daemon.sh"
if [ -f "$GRADLE_DAEMON_SCRIPT" ]; then
    echo "Ensuring Gradle daemon is running..."
    bash "$GRADLE_DAEMON_SCRIPT"
fi

# Create go/bin directory if it doesn't exist
mkdir -p "$GOPATH/bin"

# Install Go tools if needed
echo "Ensuring necessary Go tools are installed..."
which gopls > /dev/null 2>&1 || { echo "Installing gopls..."; go install golang.org/x/tools/gopls@latest; }
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
    go build -o bin/github-mcp-server ./cmd/server
    ./bin/github-mcp-server
elif [ "$1" = "test" ]; then
    echo "Running tests..."
    go test ./...
fi
