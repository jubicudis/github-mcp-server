#!/bin/bash

# WHO: ModulePathFixer
# WHAT: Script to fix Go module path issues with the GitHub MCP server
# WHEN: When experiencing module path resolution errors
# WHERE: System Layer 6 (Integration)
# WHY: To ensure consistent module resolution across the codebase
# HOW: Using Go environment configuration and build flags
# EXTENT: GitHub MCP server Go modules

echo "Fixing module path issues for the GitHub MCP server..."

# Export environment variables to ensure modules are used
export GO111MODULE=on
export GOWORK=off  # Temporarily disable workspace mode to fix the single module

# Make sure we're in the correct directory
cd "/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server" || exit 1

# Create a replacement for the go.work file to fix this specific module
cat > ../go.work.fixed <<EOF
// WHO: DevelopmentEnvironmentIntegrator
// WHAT: Go workspace configuration for TNOS
// WHEN: During development setup (updated on $(date))
// WHERE: System Layer 6 (Integration)
// WHY: To enable proper module resolution
// HOW: Using standard Go workspace feature
// EXTENT: All GitHub MCP Server operations

go 1.24

use (
    ./github-mcp-server
)

replace tranquility-neuro-os/github-mcp-server => ./github-mcp-server
EOF

echo "Created fixed go.work file as go.work.fixed"

# Create a temporary copy of the current go.mod
cp go.mod go.mod.bak

# Add replace directive to the go.mod file
cat >> go.mod <<EOF

// Local development path mapping
replace tranquility-neuro-os/github-mcp-server => .
EOF

# Attempt to tidy up the module
go mod tidy

# Fix the GOPATH approach for local development
echo "Setting up a correctly mapped GOPATH for module path resolution..."

# Create the module directory in the current GOPATH if it doesn't exist
mkdir -p "$HOME/go/src/tranquility-neuro-os"

# Create a symbolic link to the actual module location
if [ ! -L "$HOME/go/src/tranquility-neuro-os/github-mcp-server" ]; then
    ln -sf "/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server" "$HOME/go/src/tranquility-neuro-os/github-mcp-server"
fi

echo "Module path fixed. You can now build and run the server."
echo
echo "To build: go build -o bin/github-mcp-server ./cmd/server"
echo "To run: ./bin/github-mcp-server"
echo
echo "If issues persist, try switching the go.work file:"
echo "cp ../go.work.fixed ../go.work"
