#!/bin/bash
# WHO: ImportExampleRunner
# WHAT: Script to test Go module imports
# WHEN: During development and testing
# WHERE: System Layer 6 (Integration)
# WHY: To verify module resolution
# HOW: Using temporary module configuration
# EXTENT: Import example execution

# Set proper Go environment variables
export GO111MODULE=on
export GOWORK=off
export GOMOD="/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server/go.mod.fixed"

# Navigate to the MCP server directory
cd "/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server"

# Run the example with the fixed module configuration
echo "Running import example with fixed module configuration..."
env GO111MODULE=on GOMOD="${GOMOD}" GOWORK=off go run cmd/bridge/import_example.go

# Return status
exit $?
