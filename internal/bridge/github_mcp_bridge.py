#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
WHO: GitHubMCPBridgePython
WHAT: Python bridge between GitHub MCP and TNOS MCP
WHEN: During IDE runtime and system operations
WHERE: System Layer 3 (Higher Thought)
WHY: Enable communication between GitHub Copilot and TNOS with advanced Python capabilities
HOW: WebSocket protocol with context translation and Python-specific enhancements
EXTENT: All Python-based MCP communications
"""

import os
import sys
import json
import time
import logging
import argparse
import asyncio
import traceback
import threading
import websockets
from datetime import datetime
from typing import Dict, List, Any, Optional, Tuple, Union

# Import subprocess for server starting functionality
import subprocess

# Configuration
CONFIG = {
    "github_mcp": {
        "host": "localhost",
        "port": 10617,
        "ws_endpoint": "ws://localhost:10617/ws",
        "api_endpoint": "http://localhost:10617/api",
    },
    "tnos_mcp": {
        "host": "localhost",
        "port": 8888,
        "ws_endpoint": "ws://localhost:8888/ws",
        "api_endpoint": "http://localhost:8888/api",
    },
    "bridge": {
        "port": 10619,
        "context_sync_interval": 60,  # seconds
        "health_check_interval": 30,  # seconds
        "reconnect_attempts": 5,
        "reconnect_delay": 5,  # seconds
    },
    "logging": {
        "log_dir": os.path.join(
            os.environ.get("TNOS_ROOT", "/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS"),
            "logs",
        ),
        "log_file": "mcp_bridge_python.log",
        "log_level": "info",  # debug, info, warning, error
    },
    "paths": {
        "javascript_bridge": os.path.join(
            os.environ.get("TNOS_ROOT", "/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS"),
            "github-mcp-server/src/bridge/MCPBridge.js",
        ),
        "diagnostic_bridge": os.path.join(
            os.environ.get("TNOS_ROOT", "/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS"),
            "github-mcp-server/bridge/mcp_bridge.js",
        ),
    },
    "compression": {"enabled": True, "level": 7, "preserve_original": False},
}


# Configure logging
def setup_logging(level_name: str, log_file: str = None) -> logging.Logger:
    """
    WHO: LoggingInitializer
    WHAT: Set up Python logging system
    WHEN: During bridge initialization
    WHERE: System Layer 3 (Higher Thought)
    WHY: To record bridge operation details
    HOW: Using Python's logging module with file and console handlers
    EXTENT: All bridge events and messages
    """
    log_levels = {
        "debug": logging.DEBUG,
        "info": logging.INFO,
        "warning": logging.WARNING,
        "error": logging.ERROR,
    }
    level = log_levels.get(level_name.lower(), logging.INFO)

    logger = logging.getLogger("github_mcp_bridge")
    logger.setLevel(level)

    # Create formatter
    formatter = logging.Formatter(
        "[%(asctime)s] [%(levelname)s] [MCP-Python] %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )

    # Ensure log directory exists
    log_dir = CONFIG["logging"]["log_dir"]
    os.makedirs(log_dir, exist_ok=True)

    # Create file handler
    log_path = (
        log_file if log_file else os.path.join(log_dir, CONFIG["logging"]["log_file"])
    )
    file_handler = logging.FileHandler(log_path)
    file_handler.setLevel(level)
    file_handler.setFormatter(formatter)

    # Create console handler
    console_handler = logging.StreamHandler()
    console_handler.setLevel(level)
    console_handler.setFormatter(formatter)

    # Add handlers to logger
    logger.addHandler(file_handler)
    logger.addHandler(console_handler)

    return logger


# Message queue for reliability
class MessageQueue:
    """
    WHO: PythonMessageQueue
    WHAT: Queue message system for reliable delivery in Python
    WHEN: During connection interruptions
    WHERE: System Layer 3 (Higher Thought)
    WHY: To prevent message loss during disconnections
    HOW: Using thread-safe queue with disk persistence
    EXTENT: All Python-managed messages during connection issues
    """

    def __init__(self, logger: logging.Logger):
        self.github_queue = []
        self.tnos_queue = []
        self.logger = logger
        self.lock = threading.Lock()

    def queue_for_github(self, message: Dict[str, Any]) -> None:
        """Add message to GitHub MCP queue"""
        with self.lock:
            self.github_queue.append({"message": message, "timestamp": time.time()})
            self._persist_queues()
        self.logger.info(
            f"Message queued for GitHub MCP (queue size: {len(self.github_queue)})"
        )

    def queue_for_tnos(self, message: Dict[str, Any]) -> None:
        """Add message to TNOS MCP queue"""
        with self.lock:
            self.tnos_queue.append({"message": message, "timestamp": time.time()})
            self._persist_queues()
        self.logger.info(
            f"Message queued for TNOS MCP (queue size: {len(self.tnos_queue)})"
        )

    def _persist_queues(self) -> None:
        """Persist message queues to disk"""
        try:
            queue_data = {
                "github": self.github_queue,
                "tnos": self.tnos_queue,
                "timestamp": time.time(),
            }

            queue_file = os.path.join(
                CONFIG["logging"]["log_dir"], "mcp_message_queue_python.json"
            )
            with open(queue_file, "w") as f:
                json.dump(queue_data, f, indent=2)
        except Exception as e:
            self.logger.error(f"Failed to persist message queues: {str(e)}")

    def load_queues(self) -> None:
        """Load message queues from disk"""
        try:
            queue_file = os.path.join(
                CONFIG["logging"]["log_dir"], "mcp_message_queue_python.json"
            )

            if os.path.exists(queue_file):
                with open(queue_file, "r") as f:
                    queue_data = json.load(f)

                with self.lock:
                    self.github_queue = queue_data.get("github", [])
                    self.tnos_queue = queue_data.get("tnos", [])

                    # Filter out messages that are too old (over 1 hour)
                    now = time.time()
                    max_age = 60 * 60  # 1 hour
                    self.github_queue = [
                        item
                        for item in self.github_queue
                        if (now - item["timestamp"]) <= max_age
                    ]
                    self.tnos_queue = [
                        item
                        for item in self.tnos_queue
                        if (now - item["timestamp"]) <= max_age
                    ]

                    self.logger.info(
                        f"Loaded message queues: GitHub ({len(self.github_queue)}), TNOS ({len(self.tnos_queue)})"
                    )

                    # Persist filtered queues
                    self._persist_queues()
        except Exception as e:
            self.logger.error(f"Failed to load message queues: {str(e)}")
            # Initialize empty queues on error
            with self.lock:
                self.github_queue = []
                self.tnos_queue = []

    async def process_github_queue(self, websocket) -> None:
        """Process queued messages for GitHub"""
        with self.lock:
            if not self.github_queue:
                return

            self.logger.info(
                f"Processing {len(self.github_queue)} queued messages for GitHub MCP"
            )

            # Filter out messages that are too old (over 1 hour)
            now = time.time()
            max_age = 60 * 60  # 1 hour
            self.github_queue = [
                item
                for item in self.github_queue
                if (now - item["timestamp"]) <= max_age
            ]

            # Process remaining messages
            if websocket and not websocket.closed:
                # Take a copy of the queue and clear it before processing
                messages_to_process = self.github_queue.copy()
                self.github_queue = []
                self._persist_queues()

                for i, item in enumerate(messages_to_process):
                    try:
                        await websocket.send(json.dumps(item["message"]))
                        self.logger.debug(
                            f"Sent queued message {i + 1}/{len(messages_to_process)} to GitHub MCP"
                        )
                    except Exception as e:
                        self.logger.error(
                            f"Error sending queued message to GitHub MCP: {str(e)}"
                        )
                        # Re-queue the failed message
                        self.queue_for_github(item["message"])

                self.logger.info("Finished processing GitHub MCP message queue")

    async def process_tnos_queue(self, websocket) -> None:
        """Process queued messages for TNOS"""
        with self.lock:
            if not self.tnos_queue:
                return

            self.logger.info(
                f"Processing {len(self.tnos_queue)} queued messages for TNOS MCP"
            )

            # Filter out messages that are too old (over 1 hour)
            now = time.time()
            max_age = 60 * 60  # 1 hour
            self.tnos_queue = [
                item for item in self.tnos_queue if (now - item["timestamp"]) <= max_age
            ]

            # Process remaining messages
            if websocket and not websocket.closed:
                # Take a copy of the queue and clear it before processing
                messages_to_process = self.tnos_queue.copy()
                self.tnos_queue = []
                self._persist_queues()

                for i, item in enumerate(messages_to_process):
                    try:
                        await websocket.send(json.dumps(item["message"]))
                        self.logger.debug(
                            f"Sent queued message {i + 1}/{len(messages_to_process)} to TNOS MCP"
                        )
                    except Exception as e:
                        self.logger.error(
                            f"Error sending queued message to TNOS MCP: {str(e)}"
                        )
                        # Re-queue the failed message
                        self.queue_for_tnos(item["message"])

                self.logger.info("Finished processing TNOS MCP message queue")


# Context bridge module for Python
class ContextBridge:
    """
    WHO: PythonContextBridge
    WHAT: Context translation functions in Python
    WHEN: During context transformation operations
    WHERE: System Layer 3 (Higher Thought)
    WHY: To translate between GitHub and TNOS 7D contexts
    HOW: Using context mapping algorithms with enhanced compatibility
    EXTENT: All context translation operations
    """

    @staticmethod
    def github_to_tnos7d(message: Dict[str, Any]) -> Dict[str, Any]:
        """Transform GitHub MCP message to TNOS 7D format"""
        # Extract context information from GitHub message
        context = {
            "who": message.get("user", "System"),
            "what": message.get("type", "Transform"),
            "when": message.get("timestamp", int(time.time() * 1000)),
            "where": "MCP_Bridge",
            "why": message.get("purpose", "Protocol_Compliance"),
            "how": "Context_Translation",
            "extent": message.get("scope", 1.0),
        }

        # Create TNOS 7D message
        tnos7d_message = {
            "type": message.get("type", "unknown"),
            "data": message.get("data", {}),
            "context": context,
            "source": "github_mcp",
            "timestamp": int(time.time() * 1000),
            "id": message.get("id", f"bridge-{int(time.time() * 1000)}"),
        }

        return tnos7d_message

    @staticmethod
    def tnos7d_to_github(message: Dict[str, Any]) -> Dict[str, Any]:
        """Transform TNOS 7D message to GitHub MCP format"""
        # Extract context from TNOS message
        context = message.get("context", {})

        # Create GitHub MCP message
        github_message = {
            "type": message.get("type", "unknown"),
            "data": message.get("data", {}),
            "user": context.get("who", "System"),
            "purpose": context.get("why", "Protocol_Compliance"),
            "scope": context.get("extent", 1.0),
            "timestamp": message.get("timestamp", int(time.time() * 1000)),
            "id": message.get("id", f"bridge-{int(time.time() * 1000)}"),
        }

        return github_message


# Möbius Compression Module for Python
class MobiusCompression:
    """
    WHO: PythonMobiusCompression
    WHAT: Python implementation of Möbius compression
    WHEN: During data compression operations
    WHERE: System Layer 3 (Higher Thought)
    WHY: To provide efficient data compression
    HOW: Using Möbius compression formula with context-awareness
    EXTENT: All Python-managed compression operations
    """

    @staticmethod
    def compress(data: Any, params: Dict[str, Any]) -> Dict[str, Any]:
        """Compress data using Möbius formula"""
        context = params.get("context", {})
        use_time_factor = params.get("useTimeFactor", True)
        use_energy_factor = params.get("useEnergyFactor", True)

        # Convert data to string for compression if it's not already
        if not isinstance(data, str):
            data_str = json.dumps(data)
        else:
            data_str = data

        # Calculate data size
        original_size = len(data_str)

        # For demonstration, we'll implement a simple placeholder compression
        # In a real implementation, this would use the actual Möbius formula
        import base64
        import zlib

        # Get compression factors from context
        B = 0.8  # Base factor
        V = 0.7  # Value factor
        I = 0.9  # Intent factor
        G = 1.2  # Growth factor
        F = 0.6  # Flexibility factor

        # Extract potential contextual factors
        try:
            if "who" in context and context["who"] == "MCPBridge":
                B += 0.1
            if "what" in context and "Compression" in context["what"]:
                V += 0.1
            if "why" in context and "Optimize" in context["why"]:
                I += 0.1
            if "extent" in context:
                F = max(0.2, min(1.0, float(context["extent"])))
        except (ValueError, TypeError):
            pass  # Ignore context extraction errors

        # Apply compression
        compressed_data = base64.b64encode(
            zlib.compress(
                data_str.encode("utf-8"), level=CONFIG["compression"]["level"]
            )
        ).decode("utf-8")

        compressed_size = len(compressed_data)

        # Calculate compression ratio
        if original_size > 0:
            compression_ratio = 1.0 - (compressed_size / original_size)
        else:
            compression_ratio = 0.0

        # Return compression result
        return {
            "success": True,
            "originalSize": original_size,
            "compressedSize": compressed_size,
            "compressionRatio": compression_ratio,
            "data": compressed_data,
            "metadata": {
                "B": B,
                "V": V,
                "I": I,
                "G": G,
                "F": F,
                "useTimeFactor": use_time_factor,
                "useEnergyFactor": use_energy_factor,
            },
            "contextVector": context,
            "timestamp": int(time.time() * 1000),
        }

    @staticmethod
    def decompress(compressed_data: Dict[str, Any]) -> Dict[str, Any]:
        """Decompress data using Möbius formula"""
        # Extract compressed data and metadata
        data = compressed_data.get("data", "")
        metadata = compressed_data.get("metadata", {})

        # For demonstration, we'll implement a simple placeholder decompression
        import base64
        import zlib

        try:
            # Apply decompression
            decompressed_data = zlib.decompress(base64.b64decode(data)).decode("utf-8")

            # Try to parse as JSON if it looks like JSON
            if decompressed_data.strip().startswith(("{", "[")):
                try:
                    decompressed_data = json.loads(decompressed_data)
                except json.JSONDecodeError:
                    # If not valid JSON, keep as string
                    pass

            return {
                "success": True,
                "data": decompressed_data,
                "metadata": metadata,
                "timestamp": int(time.time() * 1000),
            }
        except Exception as e:
            return {
                "success": False,
                "error": f"Decompression failed: {str(e)}",
                "timestamp": int(time.time() * 1000),
            }

    @staticmethod
    def get_statistics() -> Dict[str, Any]:
        """Get compression statistics"""
        return {
            "compressionRatio": 0.75,  # Placeholder
            "averageCompressionTime": 0.05,  # seconds
            "totalCompressed": 1000,  # bytes
            "totalDecompressed": 4000,  # bytes
            "timestamp": int(time.time() * 1000),
        }


# MCP Bridge Server class
class MCPBridgeServer:
    """
    WHO: PythonMCPBridgeServer
    WHAT: Main Python bridge server implementation
    WHEN: During bridge operations
    WHERE: System Layer 3 (Higher Thought)
    WHY: To manage bridge lifecycle and communications
    HOW: Using async WebSocket connections with error handling
    EXTENT: All bridge server operations
    """

    def __init__(self, logger: logging.Logger, port: int = None):
        self.logger = logger
        self.port = port or CONFIG["bridge"]["port"]
        self.github_mcp_socket = None
        self.tnos_mcp_socket = None
        self.client_sockets = set()
        self.message_queue = MessageQueue(logger)
        self.context_bridge = ContextBridge()
        self.shutting_down = False

        # Initialize backoff strategies
        self.backoff_strategies = {
            "github": {
                "attempts": 0,
                "max_attempts": float("inf"),  # Never stop trying
                "base_delay": 1.0,  # second
                "max_delay": 30.0,  # seconds
            },
            "tnos": {
                "attempts": 0,
                "max_attempts": float("inf"),  # Never stop trying
                "base_delay": 1.0,  # second
                "max_delay": 30.0,  # seconds
            },
        }

    def get_next_backoff_delay(self, target: str) -> float:
        """Calculate next backoff delay using exponential strategy"""
        strategy = self.backoff_strategies[target]
        delay = min(
            strategy["base_delay"] * (2 ** strategy["attempts"]), strategy["max_delay"]
        )
        strategy["attempts"] += 1
        return delay

    def reset_backoff(self, target: str) -> None:
        """Reset backoff counter on successful connection"""
        strategy = self.backoff_strategies[target]
        strategy["attempts"] = 0

    async def check_server_running(self, host: str, port: int) -> bool:
        """Check if a server is running using a connection test"""
        import socket

        try:
            reader, writer = await asyncio.open_connection(host, port)
            writer.close()
            await writer.wait_closed()
            return True
        except (ConnectionRefusedError, OSError):
            return False

    async def ensure_servers_running(self) -> bool:
        """Ensure MCP servers are running"""
        self.logger.info("Checking if MCP servers are running...")

        github_mcp_running = await self.check_server_running(
            CONFIG["github_mcp"]["host"], CONFIG["github_mcp"]["port"]
        )
        tnos_mcp_running = await self.check_server_running(
            CONFIG["tnos_mcp"]["host"], CONFIG["tnos_mcp"]["port"]
        )

        if not github_mcp_running or not tnos_mcp_running:
            self.logger.info("Starting MCP servers...")

            # Try to start servers
            try:
                # Get TNOS root directory
                tnos_root = os.environ.get(
                    "TNOS_ROOT", "/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS"
                )

                # Updated path to the correct script location
                start_script = os.path.join(
                    tnos_root, "scripts/shell/start_tnos_github_integration.sh"
                )

                if os.path.exists(start_script):
                    process = await asyncio.create_subprocess_exec(
                        "bash",
                        start_script,
                        stdout=subprocess.PIPE,
                        stderr=subprocess.PIPE,
                    )
                    stdout, stderr = await process.communicate()

                    if process.returncode != 0:
                        self.logger.error(
                            f"Failed to start MCP servers: {stderr.decode()}"
                        )
                        return False

                    self.logger.info("MCP servers started successfully")
                    return True
                else:
                    self.logger.error(f"Start script not found at {start_script}")
                    return False

            except Exception as e:
                self.logger.error(f"Error starting MCP servers: {str(e)}")
                return False

        self.logger.info("MCP servers are running")
        return True

    async def connect_to_github_mcp(self) -> None:
        """Connect to GitHub MCP server"""
        self.logger.info("Connecting to GitHub MCP server...")

        if self.github_mcp_socket:
            try:
                await self.github_mcp_socket.close()
            except Exception:
                pass  # Ignore errors during cleanup
            self.github_mcp_socket = None

        try:
            self.github_mcp_socket = await websockets.connect(
                CONFIG["github_mcp"]["ws_endpoint"]
            )
            self.logger.info("Connected to GitHub MCP server")

            # Reset backoff counter on successful connection
            self.reset_backoff("github")

            # Process queued messages
            await self.message_queue.process_github_queue(self.github_mcp_socket)

            # Start message processing loop
            asyncio.create_task(self.handle_github_mcp_messages())

        except (
            websockets.exceptions.WebSocketException,
            ConnectionRefusedError,
            OSError,
        ) as e:
            self.logger.error(f"GitHub MCP WebSocket error: {str(e)}")

            # Try to reconnect using exponential backoff
            next_delay = self.get_next_backoff_delay("github")
            self.logger.info(
                f"Will attempt to reconnect to GitHub MCP server in {next_delay}s "
                f"(attempt {self.backoff_strategies['github']['attempts']})"
            )

            if not self.shutting_down:
                # Schedule reconnection
                asyncio.create_task(self.delayed_reconnect("github", next_delay))

    async def connect_to_tnos_mcp(self) -> None:
        """Connect to TNOS MCP server"""
        self.logger.info("Connecting to TNOS MCP server...")

        if self.tnos_mcp_socket:
            try:
                await self.tnos_mcp_socket.close()
            except Exception:
                pass  # Ignore errors during cleanup
            self.tnos_mcp_socket = None

        try:
            ws_endpoint = (
                f"ws://{CONFIG['tnos_mcp']['host']}:{CONFIG['tnos_mcp']['port']}/ws"
            )
            self.logger.debug(f"Attempting to connect to TNOS MCP at {ws_endpoint}")

            self.tnos_mcp_socket = await websockets.connect(ws_endpoint)
            self.logger.info(
                f"Connected to TNOS MCP server on port {CONFIG['tnos_mcp']['port']}"
            )

            # Reset backoff counter on successful connection
            self.reset_backoff("tnos")

            # Process queued messages
            await self.message_queue.process_tnos_queue(self.tnos_mcp_socket)

            # Start message processing loop
            asyncio.create_task(self.handle_tnos_mcp_messages())

        except (
            websockets.exceptions.WebSocketException,
            ConnectionRefusedError,
            OSError,
        ) as e:
            self.logger.warn(
                f"TNOS MCP WebSocket error on port {CONFIG['tnos_mcp']['port']}: {str(e)}"
            )

            # If connection fails, try to reconnect using exponential backoff
            next_delay = self.get_next_backoff_delay("tnos")
            self.logger.info(
                f"Will attempt to reconnect to TNOS MCP server in {next_delay}s "
                f"(attempt {self.backoff_strategies['tnos']['attempts']})"
            )

            if not self.shutting_down:
                # Schedule reconnection
                asyncio.create_task(self.delayed_reconnect("tnos", next_delay))

    async def delayed_reconnect(self, target: str, delay: float) -> None:
        """Handle delayed reconnection using asyncio sleep"""
        await asyncio.sleep(delay)
        if self.shutting_down:
            return

        self.logger.info(f"Attempting to reconnect to {target.upper()} MCP server...")
        if target == "github":
            await self.connect_to_github_mcp()
        else:
            await self.connect_to_tnos_mcp()

    async def handle_github_mcp_messages(self) -> None:
        """Handle messages from GitHub MCP server"""
        if not self.github_mcp_socket:
            return

        try:
            async for message_data in self.github_mcp_socket:
                try:
                    message = json.loads(message_data)
                    self.logger.debug(
                        f"Received message from GitHub MCP: {json.dumps(message)}"
                    )

                    # Process message from GitHub MCP
                    await self.process_github_mcp_message(message)
                except json.JSONDecodeError:
                    self.logger.error(
                        f"Received invalid JSON from GitHub MCP: {message_data}"
                    )
                except Exception as e:
                    self.logger.error(
                        f"Error processing message from GitHub MCP: {str(e)}"
                    )
                    self.logger.debug(f"Stack trace: {traceback.format_exc()}")

        except websockets.exceptions.ConnectionClosedError as e:
            self.logger.warn(f"GitHub MCP WebSocket connection closed: {str(e)}")

            # Try to reconnect
            if not self.shutting_down:
                next_delay = self.get_next_backoff_delay("github")
                self.logger.info(
                    f"Will attempt to reconnect to GitHub MCP server in {next_delay}s "
                    f"(attempt {self.backoff_strategies['github']['attempts']})"
                )
                asyncio.create_task(self.delayed_reconnect("github", next_delay))

        except Exception as e:
            self.logger.error(f"Error in GitHub MCP message handler: {str(e)}")
            self.logger.debug(f"Stack trace: {traceback.format_exc()}")

    async def handle_tnos_mcp_messages(self) -> None:
        """Handle messages from TNOS MCP server"""
        if not self.tnos_mcp_socket:
            return

        try:
            async for message_data in self.tnos_mcp_socket:
                try:
                    message = json.loads(message_data)
                    self.logger.debug(
                        f"Received message from TNOS MCP: {json.dumps(message)}"
                    )

                    # Process message from TNOS MCP
                    await self.process_tnos_mcp_message(message)
                except json.JSONDecodeError:
                    self.logger.error(
                        f"Received invalid JSON from TNOS MCP: {message_data}"
                    )
                except Exception as e:
                    self.logger.error(
                        f"Error processing message from TNOS MCP: {str(e)}"
                    )
                    self.logger.debug(f"Stack trace: {traceback.format_exc()}")

        except websockets.exceptions.ConnectionClosedError as e:
            self.logger.warn(f"TNOS MCP WebSocket connection closed: {str(e)}")

            # Try to reconnect
            if not self.shutting_down:
                next_delay = self.get_next_backoff_delay("tnos")
                self.logger.info(
                    f"Will attempt to reconnect to TNOS MCP server in {next_delay}s "
                    f"(attempt {self.backoff_strategies['tnos']['attempts']})"
                )
                asyncio.create_task(self.delayed_reconnect("tnos", next_delay))

        except Exception as e:
            self.logger.error(f"Error in TNOS MCP message handler: {str(e)}")
            self.logger.debug(f"Stack trace: {traceback.format_exc()}")

    async def process_github_mcp_message(self, message: Dict[str, Any]) -> None:
        """Process incoming messages from GitHub MCP"""
        # Check for specific message types
        if message.get("type") == "execute_formula" or (
            message.get("path") and "formula" in message.get("path", "")
        ):
            await self.handle_formula_message(message)
            return

        if message.get("type") == "compression" or (
            message.get("path") and "compression" in message.get("path", "")
        ):
            await self.handle_compression_message(message)
            return

        if message.get("type") == "query_context" or (
            message.get("path") and "context" in message.get("path", "")
        ):
            await self.handle_context_query_message(message)
            return

        # Save context data if present
        if message.get("type") == "context" or (
            message.get("path") and message.get("path") == "/context"
        ):
            # Context persistence would be implemented here
            pass

        # Forward message to TNOS MCP
        await self.forward_to_tnos_mcp(message)

        # Forward to connected clients
        await self.broadcast_to_clients(message)

    async def process_tnos_mcp_message(self, message: Dict[str, Any]) -> None:
        """Process incoming messages from TNOS MCP"""
        # Save context data if present
        if message.get("type") == "context" or (
            message.get("path") and message.get("path") == "/context"
        ):
            # Context persistence would be implemented here
            pass

        # Transform TNOS 7D message to GitHub MCP format
        github_message = self.context_bridge.tnos7d_to_github(message)

        # Forward to GitHub MCP
        if self.github_mcp_socket and not self.github_mcp_socket.closed:
            try:
                await self.github_mcp_socket.send(json.dumps(github_message))
                self.logger.debug(
                    f"Forwarded message to GitHub MCP: {json.dumps(github_message)}"
                )
            except Exception as e:
                self.logger.error(f"Error forwarding message to GitHub MCP: {str(e)}")
                # Queue message for later
                self.message_queue.queue_for_github(github_message)
        else:
            self.logger.warn(
                "GitHub MCP socket not ready, queuing message for later delivery"
            )
            # Queue message for later
            self.message_queue.queue_for_github(github_message)

        # Forward to clients
        await self.broadcast_to_clients(message)

    async def handle_formula_message(self, message: Dict[str, Any]) -> None:
        """Handle formula execution requests"""
        formula_params = message.get("data", {}) or message.get("parameters", {})
        formula_name = formula_params.get("formulaName", "unknown")

        self.logger.info(f"Formula execution request received: {formula_name}")

        # In a real implementation, this would delegate to the actual formula execution
        result = {
            "success": True,
            "result": f"Formula {formula_name} execution handled by Python bridge",
            "timestamp": int(time.time() * 1000),
        }

        # Send response back to GitHub MCP
        if self.github_mcp_socket and not self.github_mcp_socket.closed:
            await self.github_mcp_socket.send(
                json.dumps(
                    {
                        "type": "formula_result",
                        "requestId": message.get("requestId", "unknown"),
                        "data": result,
                    }
                )
            )

    async def handle_compression_message(self, message: Dict[str, Any]) -> None:
        """Handle compression requests"""
        compression_params = message.get("data", {}) or message.get("parameters", {})

        # Perform compression
        result = MobiusCompression.compress(
            compression_params.get("inputData", ""),
            {
                "context": compression_params.get("context", {}),
                "useTimeFactor": compression_params.get("useTimeFactor", True),
                "useEnergyFactor": compression_params.get("useEnergyFactor", True),
            },
        )

        # Send response back to GitHub MCP
        if self.github_mcp_socket and not self.github_mcp_socket.closed:
            await self.github_mcp_socket.send(
                json.dumps(
                    {
                        "type": "compression_result",
                        "requestId": message.get("requestId", "unknown"),
                        "data": result,
                    }
                )
            )

    async def handle_context_query_message(self, message: Dict[str, Any]) -> None:
        """Handle context query requests"""
        query_params = message.get("data", {}) or message.get("parameters", {})
        dimension = query_params.get("dimension", "unknown")
        query = query_params.get("query", "")

        self.logger.info(f"Context query received: {dimension} - {query}")

        # In a real implementation, this would query the actual context system
        result = {
            "success": True,
            "dimension": dimension,
            "query": query,
            "result": f"Context query for '{dimension}' handled by Python bridge",
            "timestamp": int(time.time() * 1000),
        }

        # Send response back to GitHub MCP
        if self.github_mcp_socket and not self.github_mcp_socket.closed:
            await self.github_mcp_socket.send(
                json.dumps(
                    {
                        "type": "context_result",
                        "requestId": message.get("requestId", "unknown"),
                        "data": result,
                    }
                )
            )

    async def forward_to_tnos_mcp(self, message: Dict[str, Any]) -> None:
        """Forward messages to TNOS MCP with transformation"""
        # Transform GitHub MCP message to TNOS 7D format
        tnos7d_message = self.context_bridge.github_to_tnos7d(message)

        # Forward to TNOS MCP
        if self.tnos_mcp_socket and not self.tnos_mcp_socket.closed:
            try:
                await self.tnos_mcp_socket.send(json.dumps(tnos7d_message))
                self.logger.debug(
                    f"Forwarded message to TNOS MCP: {json.dumps(tnos7d_message)}"
                )
            except Exception as e:
                self.logger.error(f"Error forwarding message to TNOS MCP: {str(e)}")
                # Queue message for later
                self.message_queue.queue_for_tnos(tnos7d_message)
        else:
            self.logger.warn(
                "TNOS MCP socket not ready, queuing message for later delivery"
            )
            # Queue message for later
            self.message_queue.queue_for_tnos(tnos7d_message)

    async def broadcast_to_clients(self, message: Dict[str, Any]) -> None:
        """Broadcast messages to connected monitoring clients"""
        message_json = json.dumps(message)
        await asyncio.gather(
            *[
                client.send(message_json)
                for client in list(self.client_sockets)
                if not client.closed
            ],
            return_exceptions=True,
        )

    async def sync_context(self) -> None:
        """Synchronize context between MCP servers"""
        self.logger.info("Syncing context between MCP servers...")

        # In a real implementation, this would load and sync actual contexts
        # For now, we'll just log the synchronization

        if self.tnos_mcp_socket and not self.tnos_mcp_socket.closed:
            await self.tnos_mcp_socket.send(
                json.dumps(
                    {
                        "type": "context",
                        "data": {
                            "source": "github",
                            "timestamp": int(time.time() * 1000),
                        },
                    }
                )
            )
            self.logger.debug("Sent GitHub context to TNOS MCP")

        if self.github_mcp_socket and not self.github_mcp_socket.closed:
            await self.github_mcp_socket.send(
                json.dumps(
                    {
                        "type": "context",
                        "data": {
                            "source": "tnos7d",
                            "timestamp": int(time.time() * 1000),
                        },
                    }
                )
            )
            self.logger.debug("Sent TNOS 7D context to GitHub MCP")

        self.logger.info("Context sync complete")

    async def health_check(self) -> None:
        """Perform health check on MCP servers"""
        self.logger.debug("Performing health check on MCP servers...")

        github_mcp_running = await self.check_server_running(
            CONFIG["github_mcp"]["host"], CONFIG["github_mcp"]["port"]
        )
        tnos_mcp_running = await self.check_server_running(
            CONFIG["tnos_mcp"]["host"], CONFIG["tnos_mcp"]["port"]
        )

        if not github_mcp_running:
            self.logger.warn("GitHub MCP server is not responding")
            await self.connect_to_github_mcp()

        if not tnos_mcp_running:
            self.logger.warn("TNOS MCP server is not responding")
            await self.connect_to_tnos_mcp()

        if not github_mcp_running or not tnos_mcp_running:
            self.logger.info("Attempting to restart MCP servers...")
            await self.ensure_servers_running()

    async def handle_client(self, websocket, path) -> None:
        """Handle client connections for monitoring"""
        self.logger.info("Client connected to bridge")

        # Add to client set for broadcasting
        self.client_sockets.add(websocket)

        try:
            # Send initial status
            await websocket.send(
                json.dumps(
                    {
                        "type": "status",
                        "status": "connected",
                        "githubConnection": (
                            self.github_mcp_socket and not self.github_mcp_socket.closed
                        ),
                        "tnosConnection": (
                            self.tnos_mcp_socket and not self.tnos_mcp_socket.closed
                        ),
                        "timestamp": int(time.time() * 1000),
                    }
                )
            )

            # Process client messages
            async for message_data in websocket:
                try:
                    message = json.loads(message_data)

                    if message.get("type") == "command":
                        await self.process_client_command(message, websocket)

                    elif message.get("target") == "github":
                        if self.github_mcp_socket and not self.github_mcp_socket.closed:
                            await self.github_mcp_socket.send(
                                json.dumps(message.get("data", {}))
                            )
                        else:
                            self.message_queue.queue_for_github(message.get("data", {}))
                            await websocket.send(
                                json.dumps(
                                    {
                                        "type": "response",
                                        "status": "queued",
                                        "original": message,
                                    }
                                )
                            )

                    elif message.get("target") == "tnos":
                        if self.tnos_mcp_socket and not self.tnos_mcp_socket.closed:
                            await self.tnos_mcp_socket.send(
                                json.dumps(message.get("data", {}))
                            )
                        else:
                            self.message_queue.queue_for_tnos(message.get("data", {}))
                            await websocket.send(
                                json.dumps(
                                    {
                                        "type": "response",
                                        "status": "queued",
                                        "original": message,
                                    }
                                )
                            )

                except json.JSONDecodeError:
                    self.logger.error(
                        f"Received invalid JSON from client: {message_data}"
                    )

                except Exception as e:
                    self.logger.error(f"Error processing client message: {str(e)}")
                    await websocket.send(json.dumps({"type": "error", "error": str(e)}))

        except websockets.exceptions.ConnectionClosedError:
            self.logger.info("Client disconnected from bridge")

        except Exception as e:
            self.logger.error(f"Client connection error: {str(e)}")

        finally:
            self.client_sockets.discard(websocket)

    async def process_client_command(self, message: Dict[str, Any], websocket) -> None:
        """Process commands from clients"""
        command = message.get("command")

        if command == "status":
            await self.handle_status_command(message, websocket)
        elif command == "reconnect":
            await self.handle_reconnect_command(message, websocket)
        elif command == "sync":
            await self.handle_sync_command(message, websocket)
        else:
            await websocket.send(
                json.dumps(
                    {
                        "type": "response",
                        "command": command,
                        "status": "unknown",
                        "error": "Unknown command",
                    }
                )
            )

    async def handle_status_command(self, message: Dict[str, Any], websocket) -> None:
        """Handle status command"""
        status = {
            "type": "response",
            "command": "status",
            "status": {
                "githubConnection": (
                    self.github_mcp_socket and not self.github_mcp_socket.closed
                ),
                "tnosConnection": (
                    self.tnos_mcp_socket and not self.tnos_mcp_socket.closed
                ),
                "queuedMessages": {
                    "github": len(self.message_queue.github_queue),
                    "tnos": len(self.message_queue.tnos_queue),
                },
                "compressionStats": MobiusCompression.get_statistics(),
                "uptime": time.time() - self.start_time,
                "timestamp": int(time.time() * 1000),
            },
        }

        await websocket.send(json.dumps(status))

    async def handle_reconnect_command(
        self, message: Dict[str, Any], websocket
    ) -> None:
        """Handle reconnection command"""
        params = message.get("params", {})
        target = params.get("target", "all")

        if target == "github" or target == "all":
            await self.connect_to_github_mcp()

        if target == "tnos" or target == "all":
            await self.connect_to_tnos_mcp()

        await websocket.send(
            json.dumps(
                {
                    "type": "response",
                    "command": "reconnect",
                    "status": "reconnecting",
                    "target": target,
                }
            )
        )

    async def handle_sync_command(self, message: Dict[str, Any], websocket) -> None:
        """Handle sync command"""
        try:
            await self.sync_context()
            await websocket.send(
                json.dumps(
                    {"type": "response", "command": "sync", "status": "completed"}
                )
            )
        except Exception as e:
            await websocket.send(
                json.dumps(
                    {
                        "type": "response",
                        "command": "sync",
                        "status": "error",
                        "error": str(e),
                    }
                )
            )

    async def ensure_bridge_files(self) -> bool:
        """Validate that bridge files exist and create them if needed"""
        self.logger.info("Validating bridge file paths...")

        try:
            # Ensure directories exist
            python_bridge_dir = os.path.dirname(CONFIG["paths"]["javascript_bridge"])
            js_bridge_dir = os.path.dirname(CONFIG["paths"]["python_bridge_dir"])

            # Create directories if they don't exist
            os.makedirs(python_bridge_dir, exist_ok=True)
            os.makedirs(js_bridge_dir, exist_ok=True)

            # Create the diagnostics-expected file path
            diagnostic_dir = os.path.dirname(CONFIG["paths"]["diagnostic_bridge"])
            os.makedirs(diagnostic_dir, exist_ok=True)

            # Instead of creating a symlink, create a file that imports the real one
            diagnostic_path = CONFIG["paths"]["diagnostic_bridge"]
            if not os.path.exists(diagnostic_path):
                self.logger.info(
                    f"Creating bridge file for diagnostics at: {diagnostic_path}"
                )
                bridge_reference = (
                    "// Reference to actual bridge implementation\n"
                    f"const actualBridgePath = '{CONFIG['paths']['javascript_bridge']}';\n"
                    "try {\n"
                    "  const actualBridge = require(actualBridgePath);\n"
                    "  module.exports = actualBridge;\n"
                    "} catch (error) {\n"
                    "  console.error(`Error importing actual bridge: ${error.message}`);\n"
                    "  module.exports = { status: 'error' };\n"
                    "}\n"
                )

                with open(diagnostic_path, "w") as f:
                    f.write(bridge_reference)

            return True
        except Exception as e:
            self.logger.error(f"Error ensuring bridge files: {str(e)}")
            return False

    async def start(self) -> None:
        """Start the MCP bridge server"""
        self.start_time = time.time()

        try:
            # Load message queues
            self.message_queue.load_queues()

            # Ensure MCP servers are running
            servers_running = await self.ensure_servers_running()
            if not servers_running:
                self.logger.error("Failed to start MCP servers, bridge cannot start")
                return False

            # Connect to MCP servers
            await self.connect_to_github_mcp()
            await self.connect_to_tnos_mcp()

            # Start the WebSocket server
            self.logger.info(f"Starting MCP Bridge server on port {self.port}")
            async with websockets.serve(self.handle_client, "0.0.0.0", self.port):
                # Set up periodic tasks
                context_sync_task = asyncio.create_task(self.context_sync_loop())
                health_check_task = asyncio.create_task(self.health_check_loop())

                # Wait forever (or until interrupted)
                await asyncio.Future()  # Run forever

        except Exception as e:
            self.logger.error(f"Error starting MCP Bridge: {str(e)}")
            self.logger.debug(traceback.format_exc())
            return False

    async def context_sync_loop(self) -> None:
        """Run context sync at regular intervals"""
        while True:
            try:
                await self.sync_context()
            except Exception as e:
                self.logger.error(f"Error in context sync: {str(e)}")

            await asyncio.sleep(CONFIG["bridge"]["context_sync_interval"])

    async def health_check_loop(self) -> None:
        """Run health checks at regular intervals"""
        while True:
            try:
                await self.health_check()
            except Exception as e:
                self.logger.error(f"Error in health check: {str(e)}")

            await asyncio.sleep(CONFIG["bridge"]["health_check_interval"])

    def shutdown(self) -> None:
        """Shutdown the bridge server gracefully"""
        self.logger.info("Shutting down MCP Bridge...")
        self.shutting_down = True


def main() -> None:
    """
    WHO: BridgeLauncher
    WHAT: Main function to launch the MCP bridge
    WHEN: During program execution
    WHERE: System Layer 3 (Higher Thought)
    WHY: To initialize and start the bridge server
    HOW: Using asyncio for concurrent operations
    EXTENT: Bridge server lifecycle
    """
    # Parse command line arguments
    parser = argparse.ArgumentParser(description="GitHub MCP Bridge for TNOS")
    parser.add_argument(
        "--port",
        type=int,
        default=CONFIG["bridge"]["port"],
        help=f'Port to run the bridge on (default: {CONFIG["bridge"]["port"]})',
    )
    parser.add_argument(
        "--github-port",
        type=int,
        default=CONFIG["github_mcp"]["port"],
        help=f'Port for GitHub MCP server (default: {CONFIG["github_mcp"]["port"]})',
    )
    parser.add_argument(
        "--tnos-port",
        type=int,
        default=CONFIG["tnos_mcp"]["port"],
        help=f'Port for TNOS MCP server (default: {CONFIG["tnos_mcp"]["port"]})',
    )
    parser.add_argument(
        "--log-level",
        choices=["debug", "info", "warning", "error"],
        default=CONFIG["logging"]["log_level"],
        help=f'Log level (default: {CONFIG["logging"]["log_level"]})',
    )
    parser.add_argument(
        "--log-file",
        type=str,
        help="Path to log file (default: auto-generated based on configuration)",
    )
    parser.add_argument(
        "--context-vector",
        type=str,
        help="JSON string containing 7D context vector for bridge initialization",
    )
    args = parser.parse_args()

    # Update configuration based on arguments
    CONFIG["bridge"]["port"] = args.port
    CONFIG["github_mcp"]["port"] = args.github_port
    CONFIG["github_mcp"]["ws_endpoint"] = f"ws://localhost:{args.github_port}/ws"
    CONFIG["tnos_mcp"]["port"] = args.tnos_port
    CONFIG["tnos_mcp"]["ws_endpoint"] = f"ws://localhost:{args.tnos_port}/ws"

    # Parse context vector if provided
    context_vector = None
    if args.context_vector:
        try:
            context_vector = json.loads(args.context_vector)
        except json.JSONDecodeError as e:
            print(f"Error parsing context vector: {e}")
            context_vector = {
                "who": "MCPBridge",
                "what": "Connectivity",
                "when": "Runtime",
                "where": "Layer6_Integration",
                "why": "CrossSystemCommunication",
                "how": "ProtocolTranslation",
                "extent": "AllMCPCommunications",
            }

    # Setup logging
    logger = setup_logging(args.log_level, args.log_file)
    logger.info(f"Starting GitHub MCP Bridge on port {args.port}")
    logger.info(f"GitHub MCP port: {args.github_port}, TNOS MCP port: {args.tnos_port}")

    if context_vector:
        logger.info(f"Using context vector: {json.dumps(context_vector)}")

    # Create and start the bridge server
    bridge_server = MCPBridgeServer(logger, args.port)

    # Handle Ctrl+C gracefully
    def signal_handler():
        logger.info("Received shutdown signal, terminating...")
        bridge_server.shutdown()
        sys.exit(0)

    # Register signal handlers for graceful shutdown
    for sig in ("SIGINT", "SIGTERM"):
        try:
            loop = asyncio.get_event_loop()
            loop.add_signal_handler(
                getattr(signal, sig),
                lambda: asyncio.create_task(loop.run_in_executor(None, signal_handler)),
            )
        except (NotImplementedError, RuntimeError, AttributeError):
            pass  # Signal handling not available on this platform

    # Run the server
    asyncio.run(bridge_server.start())


if __name__ == "__main__":
    # For testing the script can be run directly
    import signal

    main()
