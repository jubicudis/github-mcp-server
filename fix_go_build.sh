#!/bin/bash

# WHO: GoPathSetup
# WHAT: Ensure Go is properly in the PATH and build the GitHub MCP server
# WHEN: During development environment setup
# WHERE: System Layer 6 (Integration)
# WHY: To fix path issues with Go and build the server
# HOW: Using environment variables and build commands
# EXTENT: One-time setup for GitHub MCP server

echo "Setting up Go environment for GitHub MCP server..."

# Add common Go installation paths to PATH if not already there
export PATH="/usr/local/bin:$HOME/go/bin:/opt/homebrew/bin:$PATH"

# Verify Go is available
if ! command -v go &> /dev/null; then
    echo "ERROR: Go is not found in your PATH even after including common locations."
    echo "Please install Go from https://golang.org/dl/"
    echo "For macOS users, you can install Go using Homebrew: brew install go"
    exit 1
fi

# Verify Go version
GO_VERSION=$(go version)
echo "Found Go: $GO_VERSION"

# Set up Go environment variables
export GOROOT=$(go env GOROOT)
export GO111MODULE=on
export GOWORK="/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/go.work"

# Navigate to the GitHub MCP server directory
cd "/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server" || exit 1
echo "Current directory: $(pwd)"

# Create bin directory if it doesn't exist
mkdir -p bin

# Fix dependencies
echo "Running go mod tidy..."
go mod tidy

# Build the server
echo "Building GitHub MCP server..."
go build -o bin/github-mcp-server ./cmd/server

if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo "You can now run the server with: ./bin/github-mcp-server"
    
    # Copy to workspace bin directory for tasks to find it
    cp -f ./bin/github-mcp-server "/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/bin/github-mcp-server" 2>/dev/null || true
    echo "Binary also copied to workspace bin directory"
else
    echo "Build failed. Check the error messages above for details."
    exit 1
fi

echo "To start all MCP components, run: ./setup_dev_env.sh start-all"
