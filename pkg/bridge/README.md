# Canonical Go MCP Bridge

This directory (`pkg/bridge`) contains the canonical Go implementation of the MCP bridge for the GitHub MCP Server. All bridge logic, protocol, and shared types should reside here. 

- Do not duplicate bridge logic in `pkg/github` or other subfolders.
- If bridge functionality is needed elsewhere, import from this package.
- For GitHub-specific API logic, use `pkg/github` and call into the bridge as needed.

This ensures a single source of truth for bridge operations and avoids redundancy.
