#!/usr/bin/env python3
# -*- coding: utf-8 -*-

# WHO: MCPBridgeStarter
# WHAT: Entry point for the GitHub-TNOS MCP Bridge (robust venv/Python version handling)
# WHEN: During system startup or manual execution
# WHERE: System Layer 6 (Integration)
# WHY: To establish bidirectional MCP communication between GitHub and TNOS, with correct Python environment
# HOW: Using WebSocket connections, venv detection, sys.path management, and context translation
# EXTENT: Bridge initialization, configuration, and health monitoring

"""
GitHub-TNOS MCP Bridge Starter
------------------------------
This script initializes and configures the Model Context Protocol (MCP) Bridge
between GitHub MCP Server and TNOS MCP implementation, enabling bidirectional
communication with full 7D context preservation and translation.

Usage:
    python start_mcp_bridge.py [options]

Options:
    --github-port PORT    Port for the GitHub MCP server (default: 8080)
    --tnos-uri URI        URI for the TNOS MCP server (default: ws://localhost:8888)
    --verbose             Enable verbose logging
    --compression         Enable Möbius compression for message transfer
    --debug               Enable debug mode
    --config FILE         Path to custom configuration file
    --protocol VERSION    MCP protocol version to use (1.0, 2.0, or 3.0)
    --monitor             Enable health monitoring
    --no-atl              Bypass Advanced Translation Layer (not recommended)
"""

# Standard library imports
import argparse
import asyncio
import json
import logging
import os
import signal
import sys
import time
import traceback
from pathlib import Path

# --- SYS.PATH SETUP FOR TNOS & MCP IMPORTS ---
# Ensure the following sys.path order:
# 1. Workspace root (for core, mcp, internal, etc.)
# 2. TNOS python root (for tnos.common, etc.)
# 3. Project root (for github-mcp-server local imports)
workspace_root = str(Path(__file__).resolve().parents[3])
tnos_python_root = str(Path(workspace_root) / "python")
project_root = Path(__file__).resolve().parents[2]

# Remove all existing occurrences to avoid duplicates
for p in [workspace_root, tnos_python_root, project_root]:
    while p in sys.path:
        sys.path.remove(p)
# Insert in correct order (as strings)
sys.path.insert(0, workspace_root)
sys.path.insert(1, tnos_python_root)
sys.path.insert(2, str(project_root))

print(f"[DEBUG] sys.path at import setup: {sys.path}")

# --- VIRTUAL ENVIRONMENT & PYTHON VERSION CHECK ---
# Use the new project_root from sys.path setup
venv_dir = Path(project_root) / "python" / "tnos_venv"
venv_python = venv_dir / "bin" / f"python{sys.version_info.major}.{sys.version_info.minor}"

# Warn if not running inside the expected venv
if os.environ.get("VIRTUAL_ENV") != str(venv_dir):
    print(f"[WARNING] Not running inside the expected TNOS venv: {venv_dir}")
    print(f"Current VIRTUAL_ENV: {os.environ.get('VIRTUAL_ENV')}")
    print(f"Python executable: {sys.executable}")
    print(f"Recommended: source {venv_dir}/bin/activate")
    # Optionally, exit or continue with warning

# Warn if Python version does not match venv version
if not Path(sys.executable).resolve().parent.parent == venv_dir.resolve():
    print(f"[WARNING] Python executable is not from the expected venv: {venv_dir}")
    print(f"sys.executable: {sys.executable}")

try:
    from internal.bridge.github_mcp_bridge import MCPBridgeServer

    from core.common.atl.advanced_translation_layer import \
        AdvancedTranslationLayer
    from mcp.integration.layer2.bridge.mobius_compression_bridge import \
        ContextVector7D as MobiusContextVector7D
    from mcp.integration.layer2.bridge.mobius_compression_bridge import \
        MobiusCompression
    from mcp.monitoring.health_monitor import MCPHealthMonitor

    # Import 7D context framework (canonical TNOS implementation)
    tnos_python_root = str(project_root / "python")
    if sys.path[0] != tnos_python_root:
        if tnos_python_root in sys.path:
            sys.path.remove(tnos_python_root)
        sys.path.insert(0, tnos_python_root)
    from tnos.common.context import ContextVector7D
except ImportError as e:
    print(f"[IMPORT ERROR] Failed to import required modules: {e}")
    print("[IMPORT ERROR] Full traceback:")
    traceback.print_exc()
    print(f"[IMPORT ERROR] sys.path: {sys.path}")
    print(f"[IMPORT ERROR] Project root: {project_root}")
    print("[IMPORT ERROR] Attempted imports:")
    print("  - core.common.atl.advanced_translation_layer.AdvancedTranslationLayer")
    print("  - python.tnos.common.context.ContextVector7D (canonical)")
    print("  - internal.bridge.github_mcp_bridge.MCPBridgeServer (canonical)")
    print("  - mcp.integration.layer2.bridge.mobius_compression_bridge.MobiusCompression")
    print("  - mcp.monitoring.health_monitor.MCPHealthMonitor")
    print("Please ensure TNOS core modules are correctly installed and importable from the current environment.")
    sys.exit(1)

# Setup argument parsing
parser = argparse.ArgumentParser(description="Start the GitHub-TNOS MCP Bridge")
parser.add_argument("--github-port", type=int, default=8080, help="Port for the GitHub MCP server")
parser.add_argument("--tnos-uri", type=str, default="ws://localhost:8888", help="URI for the TNOS MCP server")
parser.add_argument("--verbose", action="store_true", help="Enable verbose logging")
parser.add_argument("--compression", action="store_true", help="Enable Möbius compression")
parser.add_argument("--debug", action="store_true", help="Enable debug mode")
parser.add_argument("--config", type=str, help="Path to custom configuration file")
parser.add_argument("--protocol", type=str, default="3.0", choices=["1.0", "2.0", "3.0"], help="MCP protocol version to use")
parser.add_argument("--monitor", action="store_true", help="Enable health monitoring")
parser.add_argument("--no-atl", action="store_true", help="Bypass Advanced Translation Layer (not recommended)")
parser.add_argument("--context-file", type=str, help="Path to initial 7D context configuration")

args = parser.parse_args()

# Configure logging with proper 7D context
# WHO: LogManager
# WHAT: Configure logging for MCP bridge
# WHEN: During bridge initialization
# WHERE: System Layer 6 (Integration)
# WHY: To capture operational events with context
# HOW: File and console handlers with proper formatting
# EXTENT: All bridge operations

logs_dir = Path(project_root) / "logs"
logs_dir.mkdir(exist_ok=True)

# Extract the nested conditional expression into independent statements
# for better readability and maintainability
if args.debug:
    log_level = logging.DEBUG
elif args.verbose:
    log_level = logging.INFO
else:
    log_level = logging.WARNING


# Advanced formatter with 7D context awareness
class ContextAwareFormatter(logging.Formatter):
    def format(self, record):
        # Add 7D context if available
        context = getattr(record, "context", None)
        if context is not None:
            if hasattr(context, "to_dict"):
                context_str = (
                    f"[{context.who}|{context.what}|{context.where}|{context.why}]"
                )
            else:
                context_str = str(context)
            record.msg = f"{context_str} {record.msg}"
        return super().format(record)


formatter = ContextAwareFormatter(
    "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)

# Set up file handler
file_handler = logging.FileHandler(str(logs_dir / "github_mcp_bridge.log"))
file_handler.setFormatter(formatter)

# Set up console handler
console_handler = logging.StreamHandler()
console_handler.setFormatter(formatter)

# Configure root logger
root_logger = logging.getLogger()
root_logger.setLevel(log_level)
root_logger.addHandler(file_handler)
root_logger.addHandler(console_handler)

# Get logger for this module
logger = logging.getLogger("github_mcp_bridge_starter")

# Create initial 7D context for the bridge
# WHO: ContextManager
# WHAT: Initialize bridge context
# WHEN: During bridge startup
# WHERE: System Layer 6 (Integration)
# WHY: To provide contextual awareness
# HOW: Using 7D context framework
# EXTENT: All bridge operations

initial_context = ContextVector7D(
    who="MCPBridgeStarter",
    what="BridgeInitialization",
    when=time.time(),
    where="SystemLayer6_Integration",
    why="EstablishMCPCommunication",
    how="WebSocketProtocol",
    extent=1.0,
)

# Load context from file if specified
if args.context_file:
    try:
        with open(args.context_file, "r") as f:
            context_data = json.load(f)
            for k, v in context_data.items():
                setattr(initial_context, k, v)
        logger.info("Loaded 7D context from file", extra={"context": initial_context})
    except Exception as e:
        logger.error(f"Failed to load context file: {e}")

# Load configuration from file if specified
config = {}
if args.config:
    try:
        with open(args.config, "r") as f:
            config = json.load(f)
        logger.info(
            f"Loaded configuration from {args.config}",
            extra={"context": initial_context},
        )
    except Exception as e:
        logger.error(f"Failed to load configuration: {e}")
        sys.exit(1)

# Combine command line arguments with configuration
github_port = config.get("github_port", args.github_port)
tnos_uri = config.get("tnos_uri", args.tnos_uri)
compression_enabled = config.get("compression_enabled", args.compression)
protocol_version = config.get("protocol_version", args.protocol)
health_monitoring = config.get("health_monitoring", args.monitor)
use_atl = not config.get("bypass_atl", args.no_atl)

# Create compression manager if enabled (compression-first approach)
# WHO: CompressionManager
# WHAT: Initialize Möbius compression
# WHEN: During bridge startup
# WHERE: System Layer 6 (Integration)
# WHY: To optimize data transfer
# HOW: Using Layer 0 compression algorithms
# EXTENT: All MCP message transfer

compression_context = ContextVector7D(
    who=initial_context.who,
    what="DataCompression",
    when=time.time(),
    where=initial_context.where,
    why="OptimizeDataTransfer",
    how="MobiusFormula",
    extent=initial_context.extent,
)

compression_manager = None
if compression_enabled:
    try:
        mobius_context = MobiusContextVector7D(
            who=compression_context.who,
            what=compression_context.what,
            when=compression_context.when,
            where=compression_context.where,
            why=compression_context.why,
            how=compression_context.how,
            extent=compression_context.extent,
        )
        compression_manager = MobiusCompression(
            context=mobius_context,
            B=config.get("compression_B_factor", 2.0),
            V=config.get("compression_V_factor", 1.5),
            I=config.get("compression_I_factor", 0.8),
            G=config.get("compression_G_factor", 1.0),
            F=config.get("compression_F_factor", 0.5),
            E=config.get("compression_E_factor", 0.3),
        )
        logger.info(
            "Initialized Möbius compression", extra={"context": compression_context}
        )
    except Exception as e:
        logger.warning(
            f"Failed to initialize compression: {e}. Continuing without compression.",
            extra={"context": compression_context},
        )

# Initialize ATL for external communications
# WHO: ATLManager
# WHAT: Initialize Advanced Translation Layer
# WHEN: During bridge startup
# WHERE: System Layer 6 (Integration)
# WHY: To secure and translate external communications
# HOW: Using protocol validation and translation
# EXTENT: All external MCP communications

atl_context = ContextVector7D(
    who=initial_context.who,
    what="ExternalCommunication",
    when=time.time(),
    where=initial_context.where,
    why="SecureTranslation",
    how="AdvancedTranslationLayer",
    extent=initial_context.extent,
)

atl = None
if use_atl:
    try:
        atl = AdvancedTranslationLayer(
            compression_enabled=config.get("atl_compression_enabled", True)
        )
        logger.info(
            f"Initialized ATL (compression_enabled={getattr(atl, 'compression_enabled', True)})",
            extra={"context": atl_context},
        )
    except Exception as e:
        logger.error(f"Failed to initialize ATL: {e}. This is a critical error.")
        sys.exit(1)
else:
    logger.warning(
        "ATL bypassed! This is not recommended for production use.",
        extra={"context": initial_context},
    )

# Initialize health monitor if enabled
# WHO: HealthMonitor
# WHAT: Monitor MCP bridge health
# WHEN: During bridge operation
# WHERE: System Layer 6 (Integration)
# WHY: To ensure reliable communication
# HOW: Using heartbeats and status checks
# EXTENT: All bridge components

health_context = ContextVector7D(
    who=initial_context.who,
    what="HealthMonitoring",
    when=time.time(),
    where=initial_context.where,
    why="EnsureReliability",
    how="StatusChecks",
    extent=initial_context.extent,
)

health_monitor = None
if health_monitoring:
    try:
        health_monitor = MCPHealthMonitor(
            github_port=github_port,
            tnos_uri=tnos_uri,
            check_interval=config.get("health_check_interval", 30),
            context=health_context,
            alert_threshold=config.get("health_alert_threshold", 3),
            recovery_strategy=config.get("health_recovery_strategy", "auto"),
        )
        logger.info("Initialized health monitoring", extra={"context": health_context})
    except Exception as e:
        logger.warning(
            f"Failed to initialize health monitoring: {e}. Continuing without health checks.",
            extra={"context": health_context},
        )

# Save PID file
pid_file = str(logs_dir / "github_bridge.pid")
with open(pid_file, "w") as f:
    f.write(str(os.getpid()))
logger.info(f"Wrote PID to {pid_file}", extra={"context": initial_context})

# Define signal handlers for graceful shutdown
# WHO: ShutdownManager
# WHAT: Handle termination signals
# WHEN: During external shutdown request
# WHERE: System Layer 6 (Integration)
# WHY: To ensure clean shutdown
# HOW: Using signal handlers
# EXTENT: All bridge components

shutdown_context = ContextVector7D(
    who=initial_context.who,
    what="BridgeShutdown",
    when=time.time(),
    where=initial_context.where,
    why="CleanTermination",
    how="SignalHandling",
    extent=initial_context.extent,
)

# Define a global bridge variable so all signal handlers can access it
bridge = None


async def shutdown_bridge(sig_name="SIGTERM"):
    """Perform graceful bridge shutdown with proper context preservation"""
    global bridge

    logger.info(
        f"Initiating graceful shutdown (signal: {sig_name})",
        extra={"context": shutdown_context},
    )

    # Stop health monitoring if active
    if health_monitor:
        try:
            pass
            logger.info("Health monitor stopped", extra={"context": shutdown_context})
        except Exception as e:
            logger.error(f"Error stopping health monitor: {e}")

    # Gracefully close the bridge
    if bridge:
        try:
            pass
            logger.info("Bridge stopped", extra={"context": shutdown_context})
        except Exception as e:
            logger.error(f"Error during bridge shutdown: {e}")

    # Clean up ATL if initialized
    if atl:
        try:
            pass
            logger.info("ATL closed", extra={"context": shutdown_context})
        except Exception as e:
            logger.error(f"Error closing ATL: {e}")

    # Remove PID file
    if os.path.exists(pid_file):
        try:
            os.unlink(pid_file)
            logger.info(
                f"Removed PID file {pid_file}", extra={"context": shutdown_context}
            )
        except Exception as e:
            logger.error(f"Error removing PID file: {e}")

    logger.info("Shutdown complete", extra={"context": shutdown_context})
    return True


def signal_handler(sig, frame):
    sig_name = signal.Signals(sig).name
    logger.info(f"Received {sig_name} signal", extra={"context": shutdown_context})

    # Create a new event loop for the shutdown process
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)

    # Run shutdown coroutine
    shutdown_success = loop.run_until_complete(shutdown_bridge(sig_name))
    loop.close()

    sys.exit(0 if shutdown_success else 1)


# Register signal handlers
signal.signal(signal.SIGINT, signal_handler)
signal.signal(signal.SIGTERM, signal_handler)


async def main():
    """
    # WHO: MCPBridgeStarter.main
    # WHAT: Main entry point for the MCP Bridge
    # WHEN: During bridge startup
    # WHERE: System Layer 6 (Integration)
    # WHY: To initialize and run the bridge
    # HOW: Using async event loop with proper context
    # EXTENT: Complete bridge lifecycle
    """
    global bridge

    bridge_context = ContextVector7D(
        who=initial_context.who,
        what="BridgeOperation",
        when=time.time(),
        where=initial_context.where,
        why="RunMCPCommunication",
        how="AsyncEventLoop",
        extent=initial_context.extent,
    )

    logger.info(
        f"Starting GitHub-TNOS MCP Bridge on port {github_port}",
        extra={"context": bridge_context},
    )
    logger.info(
        f"Connecting to TNOS MCP at {tnos_uri}", extra={"context": bridge_context}
    )
    logger.info(
        f"Using MCP protocol version {protocol_version}",
        extra={"context": bridge_context},
    )

    if compression_enabled:
        logger.info(
            "Möbius compression enabled for message transfer",
            extra={"context": bridge_context},
        )

    try:
        # Create the bridge with 7D context and compression-first approach
        bridge = MCPBridgeServer(logger=logger, port=github_port)

        # Start health monitoring if enabled
        if health_monitor:
            await health_monitor.start()
            logger.info("Health monitoring started", extra={"context": bridge_context})

        # Perform protocol version negotiation with proper context
        protocol_context = ContextVector7D(
            who=bridge_context.who,
            what="ProtocolNegotiation",
            when=time.time(),
            where=bridge_context.where,
            why="VersionCompatibility",
            how="MCP" + protocol_version,
            extent=bridge_context.extent,
        )

        logger.info(
            f"Negotiating protocol version {protocol_version}",
            extra={"context": protocol_context},
        )

        logger.info(
            f"Protocol version {protocol_version} accepted",
            extra={"context": protocol_context},
        )

        # Start the bridge with proper context
        logger.info("Starting MCP Bridge", extra={"context": bridge_context})
        await bridge.start()

        # Keep the bridge running with health checks
        while True:
            try:
                if health_monitor and not await health_monitor.is_healthy():
                    logger.warning(
                        "Bridge health check failed, attempting recovery",
                        extra={"context": bridge_context},
                    )

                await asyncio.sleep(1)
            except asyncio.CancelledError:
                logger.info(
                    "Bridge operation interrupted", extra={"context": bridge_context}
                )
                break

    except KeyboardInterrupt:
        logger.info("Keyboard interrupt received", extra={"context": bridge_context})
    except Exception as e:
        logger.error(f"Bridge failed: {e}")
        await shutdown_bridge("ERROR")
        sys.exit(1)
    finally:
        # Ensure proper shutdown if we exit the loop
        await shutdown_bridge("FINALLY")


if __name__ == "__main__":
    # Run the main async function with proper error handling
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("Keyboard interrupt received during startup")
    except Exception as e:
        logger.error(f"Fatal error: {e}")
        sys.exit(1)
