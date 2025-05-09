#!/bin/bash

# WHO: CleanupScript
# WHAT: Removes unnecessary files and symbolic links
# WHEN: During project cleanup
# WHERE: System Layer 6 (Integration)
# WHY: To clean up unnecessary files
# HOW: Using shell commands to remove files
# EXTENT: All GitHub MCP server files

echo "=== Cleaning up GitHub MCP Server ==="
echo "Removing unnecessary backup files and symbolic links..."

# Remove backup files
find . -name "*.bak" -type f -delete
find . -name "*.backup" -type f -delete

# Remove symbolic links
find . -type l -delete

# Remove any auto-generated fix scripts created earlier
find . -name "fix_*.sh" ! -name "simple_fix.sh" -type f -delete

echo "Cleaning completed successfully."
