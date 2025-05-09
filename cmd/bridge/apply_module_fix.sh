#!/bin/bash
# WHO: ModuleFixApplier
# WHAT: Script to apply Go module fixes
# WHEN: During development setup
# WHERE: System Layer 6 (Integration)
# WHY: To permanently fix module resolution
# HOW: Using module configuration replacement
# EXTENT: Repository-wide fix

# Set proper Go environment variables
export GO111MODULE=on

# Navigate to the MCP server directory
cd "/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server"

# Backup the original go.mod file
cp go.mod go.mod.bak

# Replace the go.mod file with the fixed version
cp go.mod.fixed go.mod

# Run go mod tidy to update dependencies
go mod tidy

echo "Module fix applied successfully. Original go.mod saved as go.mod.bak"
echo "You can now import packages from this module using: \"tranquility-neuro-os/github-mcp-server/pkg/...\""
