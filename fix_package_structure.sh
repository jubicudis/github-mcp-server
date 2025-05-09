#!/bin/bash

# WHO: PackageStructureFixer
# WHAT: Script to fix package structure and build issues
# WHEN: When experiencing package resolution errors
# WHERE: System Layer 6 (Integration)
# WHY: To ensure proper package resolution across the module
# HOW: Using Go environment configuration and module inspection
# EXTENT: GitHub MCP server Go module structure

echo "Fixing package structure for GitHub MCP server..."

# Set the working directory
cd "/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server" || exit 1

# Step 1: Save the current directory structure for debugging
echo "Current package structure:"
find . -type d -not -path "*/\.*" -not -path "*/vendor/*" -not -path "*/node_modules/*" | sort

# Step 2: Check the go.mod file
echo "Current go.mod content:"
cat go.mod

# Step 3: Create a fixed go.mod file
cat > go.mod.fixed <<EOF
// WHO: DependencyManager
// WHAT: Module definition for GitHub MCP server
// WHEN: Build time and dependency resolution
// WHERE: System Layer 6 (Integration)
// WHY: To support proper Go module structure
// HOW: Using Go module system
// EXTENT: All GitHub MCP functionality
module tranquility-neuro-os/github-mcp-server

go 1.23.7

toolchain go1.24.2

// WHO: DependencyManager
// WHAT: Direct project dependencies
// WHEN: Build and runtime
// WHERE: System Layer 6 (Integration)
// WHY: Core functionality requirements
// HOW: Using semantic versioning
// EXTENT: All required external libraries
require (
	github.com/docker/docker v28.0.4+incompatible
	github.com/google/go-cmp v0.7.0
	github.com/google/go-github/v69 v69.2.0
	github.com/gorilla/websocket v1.5.3
	github.com/mark3labs/mcp-go v0.26.0
	github.com/migueleliasweb/go-github-mock v1.3.0
	github.com/spf13/cobra v1.9.1
	github.com/spf13/viper v1.20.1
	github.com/stretchr/testify v1.10.0
	github.com/shurcooL/graphql v0.0.0-20230722043721-ed46e5a46466
	golang.org/x/oauth2 v0.30.0
)

// WHO: DependencyManager
// WHAT: MCP dependency replacement
// WHEN: Build time
// WHERE: System Layer 6 (Integration)
// WHY: To resolve import path mismatch between code and dependency
// HOW: Using Go module replace directive with verified version tag
// EXTENT: All MCP functionality
replace github.com/tranquility-dev/mcp-go => github.com/mark3labs/mcp-go v0.26.0

// Local development path mapping - we only need this one, all others are derived from it
replace tranquility-neuro-os/github-mcp-server => .
EOF

echo "Created fixed go.mod file"

# Step 4: Remove any submodule go.mod files that might be causing conflicts
echo "Removing any conflicting submodule go.mod files..."
find ./pkg -name "go.mod" -type f -print0 | xargs -0 rm -f 
find ./models -name "go.mod" -type f -print0 | xargs -0 rm -f
find ./utils -name "go.mod" -type f -print0 | xargs -0 rm -f

# Step 5: Use the fixed go.mod file
cp go.mod go.mod.original
cp go.mod.fixed go.mod

# Step 6: Clean the Go cache to remove any cached module info
echo "Cleaning Go cache..."
go clean -modcache

# Step 7: Tidy up the module and verify packages
echo "Running go mod tidy..."
GO111MODULE=on GOWORK=off go mod tidy

# Step 8: Try to build with the correct configuration
echo "Attempting to build..."
GO111MODULE=on GOWORK=off go build -v -o bin/github-mcp-server ./cmd/server

# Step 9: Check if build was successful
if [ -f bin/github-mcp-server ]; then
    echo "Build SUCCESSFUL! Server binary created at: bin/github-mcp-server"
    chmod +x bin/github-mcp-server
    ls -la bin/github-mcp-server
else
    echo "Build FAILED. Here's some debugging information:"
    
    # Print the import paths in one of the problem files
    echo "Imports in context_translator.go:"
    grep -A 10 "import" pkg/github/context_translator.go
    
    # Try to list the package
    echo "Package listing for pkg/log:"
    GO111MODULE=on GOWORK=off go list -f '{{.Dir}}' tranquility-neuro-os/github-mcp-server/pkg/log
    
    # Get more details on the module
    echo "Module details:"
    GO111MODULE=on GOWORK=off go list -m -json tranquility-neuro-os/github-mcp-server
fi

echo "Script completed."
