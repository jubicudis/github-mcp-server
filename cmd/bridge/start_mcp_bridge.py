#!/usr/bin/env python3
# -*- coding: utf-8 -*-

# WHO: MCP_Bridge_Starter
# WHAT: Entry point for the GitHub-TNOS MCP Bridge
# WHEN: During system startup or manual execution
# WHERE: MCP Bridge Command Layer
# WHY: To launch and configure the MCP bridge between GitHub and TNOS
# HOW: Using command-line configuration and WebSocket connections
# EXTENT: Bridge initialization and monitoring

"""
GitHub MCP Bridge Starter
-------------------------
This script starts the GitHub to TNOS MCP Bridge and configures
the connection between the GitHub MCP Server and the TNOS MCP Server.

Usage:
    python start_mcp_bridge.py [options]

Options:
    --github-port PORT    Port for the GitHub MCP server (default: 8080)
    --tnos-uri URI        URI for the TNOS MCP server (default: ws://localhost:8888)
    --verbose             Enable verbose logging
    --compression         Enable Möbius compression for message transfer
    --debug               Enable debug mode
    --config FILE         Path to custom configuration file
"""

import argparse
import asyncio
import importlib.util
import json
import logging
import os
import signal
import socket
import sys
import time
from pathlib import Path

# Setup argument parsing
parser = argparse.ArgumentParser(description="Start the GitHub-TNOS MCP Bridge")
parser.add_argument("--github-port", type=int, default=8080, help="Port for the GitHub MCP server")
parser.add_argument("--tnos-uri", type=str, default="ws://localhost:8888", help="URI for the TNOS MCP server")
parser.add_argument("--verbose", action="store_true", help="Enable verbose logging")
parser.add_argument("--compression", action="store_true", help="Enable Möbius compression")
parser.add_argument("--debug", action="store_true", help="Enable debug mode")
parser.add_argument("--config", type=str, help="Path to custom configuration file")

args = parser.parse_args()

# Get the project root directory
project_root = os.path.abspath(os.path.join(os.path.dirname(__file__), "../.."))

# Add project root to Python path
sys.path.insert(0, project_root)

# Configure logging
log_level = logging.DEBUG if args.debug else (logging.INFO if args.verbose else logging.WARNING)
logs_dir = os.path.join(project_root, "logs")
os.makedirs(logs_dir, exist_ok=True)

logging.basicConfig(
    level=log_level,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[
        logging.FileHandler(os.path.join(logs_dir, "github_mcp_bridge.log")),
        logging.StreamHandler(),
    ],
)
logger = logging.getLogger("github_mcp_bridge_starter")

# Check for required dependencies
try:
    import websockets
except ImportError:
    logger.error("Missing required package: websockets. Please install with 'pip install websockets'")
    sys.exit(1)

# Load configuration from file if specified
config = {}
if args.config:
    try:
        with open(args.config, "r") as f:
            config = json.load(f)
        logger.info(f"Loaded configuration from {args.config}")
    except Exception as e:
        logger.error(f"Failed to load configuration: {e}")
        sys.exit(1)

# Combine command line arguments with configuration
github_port = args.github_port or config.get("github_port", 8080)
tnos_uri = args.tnos_uri or config.get("tnos_uri", "ws://localhost:8888")
compression_enabled = args.compression or config.get("compression_enabled", False)

# Save PID file
pid_file = os.path.join(logs_dir, "github_bridge.pid")
with open(pid_file, "w") as f:
    f.write(str(os.getpid()))
logger.info(f"Wrote PID to {pid_file}")

# Import the bridge module
try:
    from internal.bridge.github_mcp_bridge import GitHubTNOSBridge
except ImportError as e:
    logger.error(f"Failed to import bridge module: {e}")
    logger.error("Make sure the bridge module is installed correctly")
    sys.exit(1)

# Define signal handlers
def signal_handler(sig, frame):
    logger.info("Received shutdown signal, exiting...")
    # Clean up PID file
    if os.path.exists(pid_file):
        os.unlink(pid_file)
    sys.exit(0)

signal.signal(signal.SIGINT, signal_handler)
signal.signal(signal.SIGTERM, signal_handler)

async def main():
    logger.info(f"Starting GitHub-TNOS MCP Bridge on port {github_port}")
    logger.info(f"Connecting to TNOS MCP at {tnos_uri}")
    
    if compression_enabled:
        logger.info("Möbius compression enabled")
    
    try:
        # Create the bridge
        bridge = GitHubTNOSBridge(tnos_server_uri=tnos_uri, github_mcp_port=github_port)
        
        # Start the bridge
        await bridge.start()
    except Exception as e:
        logger.error(f"Bridge failed: {e}")
        sys.exit(1)

if __name__ == "__main__":
    # Run the main async function
    asyncio.run(main())