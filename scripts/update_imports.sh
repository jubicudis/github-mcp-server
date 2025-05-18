#!/bin/bash
# WHO: ImportUpdater
# WHAT: Update import paths in Go files
# WHEN: After module path change
# WHERE: System Layer 6 (Integration)
# WHY: To ensure compatibility with new module path
# HOW: Using grep and sed to find and replace
# EXTENT: All Go files in the github-mcp-server directory

# Script to update all import statements from github.com/tranquility-dev/github-mcp-server to github.com/jubicudis/github-mcp-server

echo "Updating import statements..."

# Find all Go files in the github-mcp-server directory
GO_FILES=$(find /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server -name "*.go")

# For each file, replace the import statements
for file in $GO_FILES; do
    echo "Processing $file"
    sed -i '' 's|github.com/tranquility-dev/github-mcp-server|github.com/jubicudis/github-mcp-server|g' "$file"
done

echo "All import statements updated successfully."
