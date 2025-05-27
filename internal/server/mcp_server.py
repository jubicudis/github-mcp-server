# DEPRECATED: This file is no longer used. All MCP server logic has moved to mcp/bridge/tnos_mcp_server.py.
# Remove any references to this file in scripts, configs, or documentation.

# This file is intentionally left blank as a deprecation marker.

# Canonical QHP endpoints for all server/bridge connections
QHP_ENDPOINTS = {
    "tnos_mcp": "ws://localhost:9001",
    "github_mcp": "ws://localhost:10617",
    "mcp_bridge": "ws://localhost:10619",
    "copilot_llm": "ws://localhost:8083",
}

# All connection logic, fallback, and QHP handshake must use these endpoints (no /ws or /bridge)
