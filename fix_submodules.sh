#!/bin/bash

# WHO: SubmoduleFixer
# WHAT: Script to fix inconsistent module paths in subpackages
# WHEN: When experiencing module path resolution errors
# WHERE: System Layer 6 (Integration)
# WHY: To ensure consistent module paths across packages
# HOW: Using text replacement in go.mod files
# EXTENT: GitHub MCP server Go modules

echo "Fixing inconsistent module paths in subpackages..."

# Set the consistent module path we want to use
MAIN_MODULE_PATH="tranquility-neuro-os/github-mcp-server"
OLD_MODULE_PATH="github.com/jubicudis/github-mcp-server"

# Function to update a go.mod file
update_gomod() {
    local file="$1"
    echo "Processing $file"
    
    # Check if the file contains the old module path
    if grep -q "$OLD_MODULE_PATH" "$file"; then
        # Create a backup
        cp "$file" "$file.bak"
        
        # Replace the module path
        sed -i '' "s|module $OLD_MODULE_PATH|module $MAIN_MODULE_PATH|g" "$file"
        echo "  - Updated module path in $file"
    else
        echo "  - No changes needed in $file"
    fi
}

# Find all go.mod files in subdirectories
echo "Looking for go.mod files in subdirectories..."
find_output=$(find "/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server" -name "go.mod" -not -path "*/\.*" -not -path "*/vendor/*")

# Process each file
for file in $find_output; do
    update_gomod "$file"
done

echo "Submodule paths have been fixed. Now running go mod tidy..."

# Move to the project root
cd "/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server" || exit

# Run go mod tidy to clean up dependencies
go mod tidy

echo "Process complete. Try building the project now."
echo
echo "To build: go build -o bin/github-mcp-server ./cmd/server"
echo "To run: ./bin/github-mcp-server"
