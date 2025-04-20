#!/usr/bin/env python3
# -*- coding: utf-8 -*-

# WHO: MCPBridgeStarter
# WHAT: Entry point for the GitHub-TNOS MCP Bridge
# WHEN: During system startup or manual execution
# WHERE: System Layer 6 (Integration)
# WHY: To establish bidirectional MCP communication between GitHub and TNOS
# HOW: Using WebSocket connections with protocol negotiation and context translation
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
from pathlib import Path

# Third-party imports
import websockets

# Add project root to Python path for module imports
try:
    project_root = Path(__file__).resolve().parent.parent.parent
    sys.path.insert(0, str(project_root))

    # Import ATL for external communications
    from core.integration.atl import AdvancedTranslationLayer

    # Import 7D context framework
    from systems.context.context_vector import ContextVector7D

    # Import the MCP bridge with proper Layer 6 boundary respect
    from mcp.internal.bridge.github_mcp_bridge import GitHubTNOSBridge

    # Import Layer 0 compression for compression-first approach
    from algorithms.compression.mobius_compression import MobiusCompression

    # Import health monitoring
    from mcp.utils.health_monitor import MCPHealthMonitor
except ImportError as e:
    print(f"Failed to import required modules: {e}")
    print("Please ensure TNOS core modules are correctly installed")
    sys.exit(1)

# Setup argument parsing
parser = argparse.ArgumentParser(description="Start the GitHub-TNOS MCP Bridge")
parser.add_argument(
    "--github-port", type=int, default=8080, help="Port for the GitHub MCP server"
)
parser.add_argument(
    "--tnos-uri",
    type=str,
    default="ws://localhost:8888",
    help="URI for the TNOS MCP server",
)
parser.add_argument("--verbose", action="store_true", help="Enable verbose logging")
parser.add_argument(
    "--compression", action="store_true", help="Enable Möbius compression"
)
parser.add_argument("--debug", action="store_true", help="Enable debug mode")
parser.add_argument("--config", type=str, help="Path to custom configuration file")
parser.add_argument(
    "--protocol",
    type=str,
    default="3.0",
    choices=["1.0", "2.0", "3.0"],
    help="MCP protocol version to use",
)
parser.add_argument("--monitor", action="store_true", help="Enable health monitoring")
parser.add_argument(
    "--no-atl",
    action="store_true",
    help="Bypass Advanced Translation Layer (not recommended)",
)
parser.add_argument(
    "--context-file", type=str, help="Path to initial 7D context configuration"
)

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

log_level = (
    logging.DEBUG if args.debug else (logging.INFO if args.verbose else logging.WARNING)
)


# Advanced formatter with 7D context awareness
class ContextAwareFormatter(logging.Formatter):
    def format(self, record):
        # Add 7D context if available
        if hasattr(record, "context"):
            context = record.context
            context_str = (
                f"[{context.who}|{context.what}|{context.where}|{context.why}]"
            )
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
            initial_context.update(context_data)
        logger.info("Loaded 7D context from file", extra={"context": initial_context})
    except Exception as e:
        logger.error(
            f"Failed to load context file: {e}", extra={"context": initial_context}
        )

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
        logger.error(
            f"Failed to load configuration: {e}", extra={"context": initial_context}
        )
        sys.exit(1)

# Combine command line arguments with configuration
github_port = args.github_port or config.get("github_port", 8080)
tnos_uri = args.tnos_uri or config.get("tnos_uri", "ws://localhost:8888")
compression_enabled = args.compression or config.get("compression_enabled", False)
protocol_version = args.protocol or config.get("protocol_version", "3.0")
health_monitoring = args.monitor or config.get("health_monitoring", False)
use_atl = not args.no_atl and not config.get("bypass_atl", False)

# Create compression manager if enabled (compression-first approach)
# WHO: CompressionManager
# WHAT: Initialize Möbius compression
# WHEN: During bridge startup
# WHERE: System Layer 6 (Integration)
# WHY: To optimize data transfer
# HOW: Using Layer 0 compression algorithms
# EXTENT: All MCP message transfer

compression_context = initial_context.derive(
    what="DataCompression", why="OptimizeDataTransfer", how="MobiusFormula"
)

compression_manager = None
if compression_enabled:
    try:
        compression_manager = MobiusCompression(context=compression_context)
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

atl_context = initial_context.derive(
    what="ExternalCommunication",
    why="SecureTranslation",
    how="AdvancedTranslationLayer",
)

atl = None
if use_atl:
    try:
        atl = AdvancedTranslationLayer(
            protocol_version=protocol_version, context=atl_context
        )
        logger.info(
            f"Initialized ATL with protocol version {protocol_version}",
            extra={"context": atl_context},
        )
    except Exception as e:
        logger.error(
            f"Failed to initialize ATL: {e}. This is a critical error.",
            extra={"context": atl_context},
        )
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

health_context = initial_context.derive(
    what="HealthMonitoring", why="EnsureReliability", how="StatusChecks"
)

health_monitor = None
if health_monitoring:
    try:
        health_monitor = MCPHealthMonitor(
            github_port=github_port,
            tnos_uri=tnos_uri,
            check_interval=config.get("health_check_interval", 30),
            context=health_context,
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

shutdown_context = initial_context.derive(
    what="BridgeShutdown", why="CleanTermination", how="SignalHandling"
)


async def shutdown_bridge(bridge, sig_name="SIGTERM"):
    """Perform graceful bridge shutdown with proper context preservation"""
    logger.info(
        f"Initiating graceful shutdown (signal: {sig_name})",
        extra={"context": shutdown_context},
    )

    # Stop health monitoring if active
    if health_monitor:
        await health_monitor.stop()
        logger.info("Health monitor stopped", extra={"context": shutdown_context})

    # Gracefully close the bridge
    if bridge:
        try:
            await bridge.stop()
            logger.info("Bridge stopped", extra={"context": shutdown_context})
        except Exception as e:
            logger.error(
                f"Error during bridge shutdown: {e}",
                extra={"context": shutdown_context},
            )

    # Clean up ATL if initialized
    if atl:
        try:
            atl.close()
            logger.info("ATL closed", extra={"context": shutdown_context})
        except Exception as e:
            logger.error(f"Error closing ATL: {e}", extra={"context": shutdown_context})

    # Remove PID file
    if os.path.exists(pid_file):
        os.unlink(pid_file)
        logger.info(f"Removed PID file {pid_file}", extra={"context": shutdown_context})

    logger.info("Shutdown complete", extra={"context": shutdown_context})
    return True


def signal_handler(sig, frame):
    sig_name = signal.Signals(sig).name
    logger.info(f"Received {sig_name} signal", extra={"context": shutdown_context})

    # Create a new event loop for the shutdown process
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)

    # Run shutdown coroutine
    shutdown_success = loop.run_until_complete(shutdown_bridge(None, sig_name))
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
    bridge_context = initial_context.derive(
        what="BridgeOperation", why="RunMCPCommunication", how="AsyncEventLoop"
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

    bridge = None
    try:
        # Configure bridge with all components
        bridge_config = {
            "tnos_server_uri": tnos_uri,
            "github_mcp_port": github_port,
            "protocol_version": protocol_version,
            "context": bridge_context,
            "compression_manager": compression_manager,
            "atl": atl,
            "health_monitor": health_monitor,
            "config": config,
        }

        # Create the bridge with 7D context and compression-first approach
        bridge = GitHubTNOSBridge(**bridge_config)

        # Register shutdown handler that includes the bridge instance
        original_sigint = signal.getsignal(signal.SIGINT)
        original_sigterm = signal.getsignal(signal.SIGTERM)

        # Update signal handlers with bridge reference
        def enhanced_signal_handler(sig, frame):
            sig_name = signal.Signals(sig).name
            logger.info(
                f"Received {sig_name} signal during bridge operation",
                extra={"context": shutdown_context},
            )

            # Create a new event loop for the shutdown process
            loop = asyncio.new_event_loop()
            asyncio.set_event_loop(loop)

            # Run shutdown coroutine with bridge reference
            shutdown_success = loop.run_until_complete(
                shutdown_bridge(bridge, sig_name)
            )
            loop.close()

            # Restore original handlers before exiting
            signal.signal(signal.SIGINT, original_sigint)
            signal.signal(signal.SIGTERM, original_sigterm)

            sys.exit(0 if shutdown_success else 1)

        # Register updated signal handlers
        signal.signal(signal.SIGINT, enhanced_signal_handler)
        signal.signal(signal.SIGTERM, enhanced_signal_handler)

        # Start health monitoring if enabled
        if health_monitor:
            await health_monitor.start()
            logger.info("Health monitoring started", extra={"context": bridge_context})

        # Perform protocol version negotiation
        protocol_context = bridge_context.derive(
            what="ProtocolNegotiation",
            why="VersionCompatibility",
            how="MCP" + protocol_version,
        )

        logger.info(
            f"Negotiating protocol version {protocol_version}",
            extra={"context": protocol_context},
        )

        negotiated_version = await bridge.negotiate_protocol_version(protocol_version)

        if negotiated_version != protocol_version:
            logger.warning(
                f"Protocol version negotiated down from {protocol_version} to {negotiated_version}",
                extra={"context": protocol_context},
            )
        else:
            logger.info(
                f"Protocol version {protocol_version} accepted",
                extra={"context": protocol_context},
            )

        # Start the bridge with proper context
        logger.info("Starting MCP Bridge", extra={"context": bridge_context})
        await bridge.start()

        # Keep the bridge running
        while True:
            try:
                await asyncio.sleep(1)
            except asyncio.CancelledError:
                logger.info(
                    "Bridge operation interrupted", extra={"context": bridge_context}
                )
                break

    except KeyboardInterrupt:
        logger.info("Keyboard interrupt received", extra={"context": bridge_context})
    except Exception as e:
        logger.error(f"Bridge failed: {e}", extra={"context": bridge_context})
        if bridge:
            await shutdown_bridge(bridge, "ERROR")
        sys.exit(1)
    finally:
        # Ensure proper shutdown if we exit the loop
        if bridge:
            await shutdown_bridge(bridge, "FINALLY")


if __name__ == "__main__":
    # Run the main async function with proper error handling
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("Keyboard interrupt received during startup")
    except Exception as e:
        logger.error(f"Fatal error: {e}")
        sys.exit(1)
