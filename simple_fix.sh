#!/bin/bash

# WHO: MCP-FixerScript
# WHAT: Resolves module conflicts in GitHub MCP server
# WHEN: Run during development environment setup
# WHERE: System Layer 6 (Integration)
# WHY: Fix package version conflicts and import issues
# HOW: Using shell commands to manipulate files and Go modules
# EXTENT: All Go files in the GitHub MCP server component

# Set error handling
set -e

cd /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server

echo "Cleaning up any problematic package structures..."

# Remove any existing go.mod files in subdirectories
find . -path "*/pkg/*" -name "go.mod" -delete
find . -path "*/models/*" -name "go.mod" -delete 
find . -path "*/utils/*" -name "go.mod" -delete

# Fix the main go.mod file
echo '// WHO: ModuleDefinition
// WHAT: Go module definition with proper dependencies
// WHEN: Build and module resolution time
// WHERE: System Layer 6 (Integration)
// WHY: Proper dependency resolution with compatible versions
// HOW: Using Go module system
// EXTENT: All package imports in GitHub MCP server
module tranquility-neuro-os/github-mcp-server

go 1.23.7

toolchain go1.24.2

require (
	github.com/docker/docker v28.0.4+incompatible
	github.com/google/go-cmp v0.7.0
	github.com/google/go-github/v49 v49.1.0 // Use v49 instead of v69 which has compatibility issues
	github.com/gorilla/websocket v1.5.3
	github.com/mark3labs/mcp-go v0.26.0
	github.com/migueleliasweb/go-github-mock v1.3.0
	github.com/shurcooL/graphql v0.0.0-20230722043721-ed46e5a46466
	github.com/spf13/cobra v1.9.1
	github.com/spf13/viper v1.20.1
	github.com/stretchr/testify v1.10.0
	golang.org/x/oauth2 v0.30.0
)

// WHO: Dependency Path Resolver
// WHAT: Map MCP dependency to correct path
// WHEN: Build time
// WHERE: System Layer 6 (Integration)
// WHY: Resolve import path mismatch
// HOW: Using Go module replace directive
// EXTENT: All MCP functionality
replace github.com/tranquility-dev/mcp-go => github.com/mark3labs/mcp-go v0.26.0
' > go.mod

echo "Fixed go.mod with proper version constraints"

# Fix import statements in Go files to use v49 instead of v69
echo "Updating import statements in Go files..."
find . -name "*.go" -type f -exec grep -l "github.com/google/go-github/v69" {} \; | while read file; do
    echo "Updating $file"
    sed -i '' 's|github.com/google/go-github/v69|github.com/google/go-github/v49|g' "$file"
done

# Check for problem in pkg/translations/context.go
if grep -q "package translations.*package translations" "pkg/translations/context.go"; then
    echo "Fixing duplicate package declaration in pkg/translations/context.go"
    sed -i '' '/package translations/,/^import/d' "pkg/translations/context.go"
fi

# Clean up Go caches
echo "Cleaning Go module cache..."
go clean -modcache
go mod tidy

# Create bin directory if it doesn't exist
mkdir -p bin

# Try to build
echo "Building the server..."
GO111MODULE=on GOWORK=off GOFLAGS="-mod=mod" go build -v -o bin/github-mcp-server ./cmd/server

# Check if build was successful
if [ -f bin/github-mcp-server ]; then
    echo "Build successful! Server binary is at: bin/github-mcp-server"
    chmod +x bin/github-mcp-server
    ls -la bin/github-mcp-server
    
    echo "==================================================="
    echo "WHO: GitHub-MCP-Server"
    echo "WHAT: GitHub MCP server built successfully"
    echo "WHEN: $(date)"
    echo "WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server"
    echo "WHY: To provide GitHub integration for TNOS"
    echo "HOW: Using Go build system with proper module resolution"
    echo "EXTENT: Ready for GitHub API interactions"
    echo "==================================================="
    
    # Create a setup script for development environment
    echo "#!/bin/bash

# WHO: Development-Environment-Setup
# WHAT: Set up development environment for GitHub MCP server
# WHEN: During development setup
# WHERE: System Layer 6 (Integration)
# WHY: To provide consistent environment for all developers
# HOW: Using environment variables and path configuration
# EXTENT: All development operations

export GO111MODULE=on
export GOFLAGS=\"-mod=mod\"
export GOWORK=off
export TNOS_CONTEXT_MODE=\"debug\"
export MCP_LOG_LEVEL=\"debug\"

echo \"GitHub MCP development environment configured\"
echo \"Run './bin/github-mcp-server' to start the server\"
" > setup_dev_env.sh
    chmod +x setup_dev_env.sh
    
    echo "Created setup_dev_env.sh for configuring development environment"
else
    echo "Build failed. Please check the error messages above."
    echo "==================================================="
    echo "WHO: GitHub-MCP-Server"
    echo "WHAT: GitHub MCP server build failed"
    echo "WHEN: $(date)"
    echo "WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server"
    echo "WHY: There may still be module or import issues"
    echo "HOW: Analyze the error messages and fix accordingly"
    echo "EXTENT: Server is not operational"
    echo "==================================================="
    exit 1
fi
