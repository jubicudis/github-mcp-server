# TNOS GitHub MCP Server - Troubleshooting Guide

## Overview

This document provides troubleshooting steps for the GitHub MCP server component of the Tranquility Neuro-OS (TNOS).

## Common Issues

### 1. Module Path Resolution

**Symptoms:**

- Import errors in Go files
- "cannot find module providing package" errors
- Version conflicts in go.mod

**Solution:**
Run the `simple_fix.sh` script to resolve module path issues:

```bash
cd /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server
./simple_fix.sh
```

### 2. Go-GitHub Version Compatibility

**Symptoms:**

- Errors related to GitHub API calls
- Missing methods or fields in github.* packages

**Solution:**
The script fixes the dependency by using v49 instead of v69:

- Removes conflicting version requirements
- Updates import statements in all Go files
- Rebuilds with the correct version

### 3. Duplicate Package Declarations

**Symptoms:**

- Syntax errors in Go files
- "package XXX already declared" errors

**Solution:**
The fix script removes duplicate package declarations and import statements.

## Development Setup

For proper development environment configuration:

1. Run the setup script:

```bash
cd /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server
source ./setup_dev_env.sh
```

2. Use VS Code tasks:
   - Press `Ctrl+Shift+P` (or `Cmd+Shift+P` on macOS)
   - Select "Tasks: Run Task"
   - Choose "run-github-mcp-server"

## Building Manually

If you prefer to build manually:

```bash
cd /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server
export GO111MODULE=on
export GOFLAGS="-mod=mod"
export GOWORK=off
go build -v -o bin/github-mcp-server ./cmd/server
```

## 7D Context for Development

When developing for the GitHub MCP server, ensure adherence to the 7D Context Framework:

1. **WHO**: Component name and interaction parties
2. **WHAT**: Function or operation being performed
3. **WHEN**: Execution context and timing
4. **WHERE**: System location (Layer 6 - Integration)
5. **WHY**: Purpose of the component or function
6. **HOW**: Implementation approach with proper Module Context Protocol
7. **EXTENT**: Scope and limitations

## Reference

For more information, see:

- [VS Code Go Configuration](https://code.visualstudio.com/docs/languages/go)
- [Go Modules Documentation](https://go.dev/ref/mod)
- [Model Context Protocol](https://modelcontextprotocol.io/)
