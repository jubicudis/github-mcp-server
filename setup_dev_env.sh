#!/bin/bash

# WHO: DevEnvSetup
# WHAT: Unified development environment setup for TNOS and GitHub MCP servers with QHP bridge
# WHEN: During development setup and environment initialization
# WHERE: System Layer 6 (Integration)
# WHY: To ensure both MCP servers are started in canonical, symmetric fashion
# HOW: Using shell environment variables, canonical port assignments, and server startup routines
# EXTENT: All MCP development operations

# Canonical QHP port assignments
QHP_TNOS_PORT=9001
QHP_GITHUB_PORT=10617
QHP_COPILOT_PORT=8083

# TNOS policy: Always use absolute paths for all critical directories and binaries
# Set absolute paths for all major components
WORKSPACE_ROOT="/Users/Jubicudis/Tranquility-Neuro-OS"
MCP_VENV_DIR="/Users/Jubicudis/Tranquility-Neuro-OS/mcp/venv"
MCP_PYTHON311="/Users/Jubicudis/Tranquility-Neuro-OS/mcp/venv/bin/python3.11"
TNOS_VENV_DIR="/Users/Jubicudis/Tranquility-Neuro-OS/systems/python/nervous/venv311"
TNOS_PYTHON311="/Users/Jubicudis/Tranquility-Neuro-OS/systems/python/nervous/venv311/bin/python3.11"
GITHUB_MCP_BIN="/Users/Jubicudis/Tranquility-Neuro-OS/github-mcp-server/bin/github-mcp-server"
GITHUB_MCP_SRC="/Users/Jubicudis/Tranquility-Neuro-OS/github-mcp-server/cmd/server"
TNOS_MCP_SERVER_PY="/Users/Jubicudis/Tranquility-Neuro-OS/mcp/bridge/tnos_mcp_server.py"
CIRCULATORY_BIN="/Users/Jubicudis/Tranquility-Neuro-OS/systems/cpp/circulatory/bin/circulatory"
FORMULA_REGISTRY_DYLIB="/Users/Jubicudis/Tranquility-Neuro-OS/systems/cpp/circulatory/libformula_registry.dylib"
LOGS_DIR="/Users/Jubicudis/Tranquility-Neuro-OS/logs"

# Debug: Print workspace and venv paths
echo "[DEBUG] WORKSPACE_ROOT: $WORKSPACE_ROOT"
echo "[DEBUG] MCP_VENV_DIR: $MCP_VENV_DIR"
echo "[DEBUG] MCP_PYTHON311: $MCP_PYTHON311"

# Canonical logs
mkdir -p "$LOGS_DIR"

# QHP-compliant environment variables
export QHP_TNOS_PORT
export QHP_GITHUB_PORT
export QHP_COPILOT_PORT
export TNOS_CIRCULATORY_BIN="$CIRCULATORY_BIN"
export TNOS_FORMULA_REGISTRY="$FORMULA_REGISTRY_DYLIB"

# Utility: Wait for port
wait_for_port() {
    local host="$1"; local port="$2"; local timeout="${3:-60}"; local waited=0
    echo "[PortWaiter] Waiting for $host:$port (timeout ${timeout}s)..."
    while ! nc -z "$host" "$port" 2>/dev/null; do
        sleep 1; waited=$((waited+1))
        if [ "$waited" -ge "$timeout" ]; then
            echo "[PortWaiter][ERROR] Timeout waiting for $host:$port after $timeout seconds."
            return 1
        fi
        if (( $waited % 10 == 0 )); then
            echo "[PortWaiter] Still waiting for $host:$port... ($waited seconds elapsed)"
        fi
    done
    echo "[PortWaiter] $host:$port is now open after $waited seconds."
    return 0
}

# Utility: Wait for /health endpoint
wait_for_health() {
    local url="$1"; local timeout="${2:-60}"; local waited=0
    while true; do
        if command -v curl >/dev/null 2>&1; then
            local status=$(curl -s -o /dev/null -w "%{http_code}" "$url")
            if [ "$status" = "200" ]; then
                echo "[HealthWaiter] $url is ready."
                return 0
            fi
        fi
        sleep 1; waited=$((waited+1))
        if [ "$waited" -ge "$timeout" ]; then
            echo "[HealthWaiter][ERROR] Timeout waiting for $url after $timeout seconds."
            return 1
        fi
        if (( $waited % 10 == 0 )); then
            echo "[HealthWaiter] Still waiting for $url... ($waited seconds elapsed)"
        fi
    done
}

# Check Python version for TNOS MCP server venv
check_python_version() {
    local version_output
    version_output=$($MCP_PYTHON311 --version 2>&1)
    if [[ "$version_output" != *"3.11"* ]]; then
        echo "[ERROR] Python 3.11 is required in $MCP_VENV_DIR, but got: $version_output"
        exit 1
    fi
    echo "[INFO] Python version OK: $version_output"
}

# Ensure Flask and websockets are installed in TNOS MCP server venv
ensure_python_deps() {
    source "$MCP_VENV_DIR/bin/activate"
    "$MCP_PYTHON311" -m pip install --quiet flask websockets || {
        echo "[ERROR] Failed to install Flask and websockets in MCP venv."; exit 1;
    }
    deactivate
}

# Check for formula registry dylib
check_formula_registry() {
    if [ ! -f "$FORMULA_REGISTRY_DYLIB" ]; then
        echo "[ERROR] Formula registry dylib not found: $FORMULA_REGISTRY_DYLIB"
        exit 1
    fi
    echo "[INFO] Formula registry dylib found: $FORMULA_REGISTRY_DYLIB"
}

# Robust health check for circulatory system using HelicalMemory and round-trip SUCCESS
check_circulatory_health() {
    local output
    for args in "--health" "health" ""; do
        if [ -z "$args" ]; then
            output=$("$CIRCULATORY_BIN")
        else
            output=$("$CIRCULATORY_BIN" $args)
        fi
        helical_ok=$(echo "$output" | grep -qE "HelicalMemory|SUCCESS" && echo 1 || echo 0)
        roundtrip_ok=$(echo "$output" | grep -q "SUCCESS" && echo 1 || echo 0)
        if [ "$helical_ok" -eq 1 ] && [ "$roundtrip_ok" -eq 1 ]; then
            return 0
        fi
    done
    return 1
}

# Validate C++ Circulatory System (HemoFlux/Blood) with robust health check
validate_circulatory_system() {
    echo "[INFO] Validating C++ Circulatory System (HemoFlux/Blood) health..."
    if check_circulatory_health; then
        echo "[INFO] Circulatory system health: OK."
    else
        echo "[ERROR] Circulatory system health check failed."
        exit 1
    fi
}

# Build GitHub MCP server if needed (output to github-mcp-server/bin/)
build_github_mcp() {
    if [ ! -f "$GITHUB_MCP_BIN" ]; then
        echo "Building GitHub MCP server..."
        mkdir -p "$WORKSPACE_ROOT/github-mcp-server/bin"
        cd "$WORKSPACE_ROOT/github-mcp-server" || exit 1
        go build -o "$GITHUB_MCP_BIN" "$GITHUB_MCP_SRC" || { echo "Build failed."; exit 1; }
    fi
}

# Start GitHub MCP server (Go, port 10617)
start_github_mcp() {
    if lsof -i :$QHP_GITHUB_PORT | grep LISTEN; then
        echo "[INFO] GitHub MCP server already running on port $QHP_GITHUB_PORT."
    else
        echo "Starting GitHub MCP server on port $QHP_GITHUB_PORT..."
        nohup "$GITHUB_MCP_BIN" > "$LOGS_DIR/github_mcp_server.log" 2>&1 &
        echo $! > "$LOGS_DIR/github_mcp_server.pid"
        wait_for_port "127.0.0.1" $QHP_GITHUB_PORT 30 || { echo "[ERROR] GitHub MCP server did not open port $QHP_GITHUB_PORT."; exit 1; }
    fi
}

# Start TNOS MCP server (Python, port 9001/8083, with visualization internal)
start_tnos_mcp() {
    if lsof -i :$QHP_TNOS_PORT | grep LISTEN; then
        echo "[INFO] TNOS MCP server already running on port $QHP_TNOS_PORT."
    else
        echo "Starting TNOS MCP server on port $QHP_TNOS_PORT..."
        source "$MCP_VENV_DIR/bin/activate"
        nohup "$MCP_PYTHON311" -u "$TNOS_MCP_SERVER_PY" > "$LOGS_DIR/tnos_mcp_server.log" 2>&1 &
        echo $! > "$LOGS_DIR/tnos_mcp_server.pid"
        deactivate
        wait_for_port "127.0.0.1" $QHP_TNOS_PORT 30 || { echo "[ERROR] TNOS MCP server did not open port $QHP_TNOS_PORT."; exit 1; }
        wait_for_port "127.0.0.1" $QHP_COPILOT_PORT 30 || echo "[WARN] Copilot LLM port $QHP_COPILOT_PORT not open (optional)."
    fi
}

# Stop all components
stop_all() {
    echo "Stopping all MCP components..."
    pkill -f "$CIRCULATORY_BIN"
    pkill -f "github-mcp-server"
    pkill -f "tnos_mcp_server.py"
    sleep 2
}

# Status for all components
status_all() {
    echo "MCP Component Status:"
    for port in $QHP_GITHUB_PORT $QHP_TNOS_PORT $QHP_COPILOT_PORT; do
        if lsof -i :$port | grep LISTEN >/dev/null; then
            echo "  [RUNNING] Port $port"
        else
            echo "  [STOPPED] Port $port"
        fi
    done
    # Circulatory system: report [HEALTHY]/[UNHEALTHY] using robust health check
    if check_circulatory_health; then
        echo "  [HEALTHY] Circulatory System (C++)"
    else
        echo "  [UNHEALTHY] Circulatory System (C++)"
    fi
    echo "[INFO] Visualization is provided internally by the TNOS MCP server (port $QHP_COPILOT_PORT)."
}

# Main CLI logic
case "$1" in
    start-all|"")
        stop_all
        check_python_version  # Checks MCP venv only
        ensure_python_deps    # Installs Flask/websockets in MCP venv only
        check_formula_registry
        validate_circulatory_system
        build_github_mcp
        start_tnos_mcp        # Uses MCP venv only
        start_github_mcp
        status_all
        ;;
    stop-all)
        stop_all
        status_all
        ;;
    status-all)
        status_all
        ;;
    *)
        echo "Usage: $0 [start-all|stop-all|status-all]"
        exit 1
        ;;
esac

# NOTE: TNOS system venv ($TNOS_VENV_DIR) is NOT used for MCP server operations.
#       If you need to start TNOS system Python components, use $TNOS_VENV_DIR and $TNOS_PYTHON311 explicitly.
