#!/bin/bash

# WHO: PackageFixerScript
# WHAT: Script to fix package imports in GitHub MCP Server
# WHEN: During setup or after code changes
# WHERE: System Layer 6 (Integration)
# WHY: To ensure proper module resolution
# HOW: Using Go module commands
# EXTENT: All source files in the GitHub MCP Server

# Set strict error handling
set -e

echo "==== TNOS GitHub MCP Server Fix Script ===="
echo "WHO: PackageFixerScript"
echo "WHAT: Fixing package imports"
echo "WHEN: $(date)"
echo "WHERE: $(pwd)"
echo "WHY: To ensure proper Go module resolution"
echo "HOW: Running Go commands to fix imports"
echo "EXTENT: All source files in GitHub MCP Server"
echo "=========================================="

# Navigate to the GitHub MCP server directory
cd "$(dirname "$0")"
echo "Working in: $(pwd)"

# Ensure Go modules are enabled
export GO111MODULE=on
export GOFLAGS="-mod=mod"

echo "Tidying go.mod..."
go mod tidy

echo "Verifying package imports..."
go mod verify

echo "Downloading dependencies..."
go mod download

echo "Updating Go imports in all .go files..."
find . -name "*.go" -type f -print0 | xargs -0 grep -l "github.com/google/go-github/v69" | tr '\n' '\0' | xargs -0 sed -i '' 's|github.com/google/go-github/v69|github.com/google/go-github/v49|g'

echo "Building project to verify changes..."
go build -v ./...

echo "Fix script completed successfully!"
