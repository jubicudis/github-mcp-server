#!/bin/bash

# WHO: GoPathFixer
# WHAT: Script to create a proper GOPATH structure for local development
# WHEN: When experiencing module path resolution errors
# WHERE: System Layer 6 (Integration)
# WHY: To ensure proper module path resolution
# HOW: Using Go environment configuration and symlinks
# EXTENT: GitHub MCP server Go modules

echo "Setting up symbolic links for proper Go module resolution..."

# Set up environment variables for Go
export GO111MODULE=on
export GOWORK=off  # Disable workspace mode for direct package resolution

# Create the directory structure in the user's GOPATH
mkdir -p ~/go/src/tranquility-neuro-os

# Create a symbolic link from the actual project location to the GOPATH location
if [ ! -L ~/go/src/tranquility-neuro-os/github-mcp-server ]; then
    ln -sf /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server ~/go/src/tranquility-neuro-os/
    echo "Created symbolic link in GOPATH"
else
    echo "Symbolic link already exists"
fi

# Now set GOPATH to include both the standard location and the project location
export GOPATH=~/go:/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS

echo "GOPATH is now set to: $GOPATH"

# Create a temp directory to build in
mkdir -p /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server/bin

# Try to build the server
echo "Attempting to build the server..."
cd /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server || exit 1
go build -o bin/github-mcp-server ./cmd/server

# Check if the build was successful
if [ -f bin/github-mcp-server ]; then
    echo "Build successful! Server binary created at: bin/github-mcp-server"
    echo ""
    echo "You can now run the server with: ./bin/github-mcp-server"
else
    echo "Build failed. Please check the error messages above."
fi
