#!/usr/bin/env python3
# -*- coding: utf-8 -*-

# WHO: MCPServer
# WHAT: TNOS Model Context Protocol server implementation
# WHEN: During system operation
# WHERE: Layer 3 / MCP Server Layer
# WHY: To provide MCP services to GitHub and other clients
# HOW: Using WebSocket server with 7D context framework
# EXTENT: All MCP communications

"""
TNOS MCP Server
--------------
This module implements the TNOS Model Context Protocol server.
It provides a WebSocket server that handles MCP requests, maintains
context awareness across all seven dimensions, and implements the
TNOS capabilities for use by GitHub Copilot and other clients.

Usage:
    python mcp_server.py [options]

Options:
    --port PORT     TCP port to listen on (default: 9000)
    --host HOST     Host to bind to (default: localhost)
    --debug         Enable debug mode
    --config FILE   Path to custom configuration file
"""

import argparse
import asyncio
import datetime
import json
import logging
import os
import signal
import sys
import time
import uuid
from typing import Any, Dict

# Setup argument parsing
try:
    from mcp.config import MCP_PORT_MAP
    GITHUB_MCP_SERVER_PORT = MCP_PORT_MAP["GITHUB_MCP_SERVER"]
except ImportError:
    GITHUB_MCP_SERVER_PORT = 9000  # Fallback if config not available
# WHO: MCPServer
# WHAT: Use canonical port assignment from MCP_PORT_MAP
# WHEN: During server startup
# WHERE: /github-mcp-server/internal/server/mcp_server.py
# WHY: To prevent port conflicts and centralize port management
# HOW: Import and use MCP_PORT_MAP for port assignment
# EXTENT: All MCP server startup logic

parser = argparse.ArgumentParser(description="Start the TNOS MCP Server")
parser.add_argument("--port", type=int, default=GITHUB_MCP_SERVER_PORT, help="TCP port to listen on")
parser.add_argument("--host", type=str, default="localhost", help="Host to bind to")
parser.add_argument("--debug", action="store_true", help="Enable debug mode")
parser.add_argument("--config", type=str, help="Path to custom configuration file")

args = parser.parse_args()

# Get the project root directory
project_root = os.path.abspath(os.path.join(os.path.dirname(__file__), "../.."))

# Add project root to Python path
sys.path.insert(0, project_root)

# Configure logging
log_level = logging.DEBUG if args.debug else logging.INFO
logs_dir = os.path.join(project_root, "logs")
os.makedirs(logs_dir, exist_ok=True)

logging.basicConfig(
    level=log_level,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[
        logging.FileHandler(os.path.join(logs_dir, "tnos_mcp_server.log")),
        logging.StreamHandler(),
    ],
)
logger = logging.getLogger("tnos_mcp_server")

# Check for required dependencies
try:
    import websockets
except ImportError:
    logger.error(
        "Missing required package: websockets. Please install with 'pip install websockets'"
    )
    sys.exit(1)


# Define the MCP Context class
class MCPContext:
    """
    Implements the 7D context framework for MCP operations
    """

    def __init__(self):
        self.who = "System"
        self.what = "Initialize"
        self.when = datetime.datetime.now().isoformat()
        self.where = "MCP_Server"
        self.why = "Startup"
        self.how = "Default"
        self.extent = "System"
        self.metadata = {}

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "MCPContext":
        """Create context from dictionary"""
        context = cls()
        if isinstance(data, dict):
            context.who = data.get("who", context.who)
            context.what = data.get("what", context.what)
            context.when = data.get("when", context.when)
            context.where = data.get("where", context.where)
            context.why = data.get("why", context.why)
            context.how = data.get("how", context.how)
            context.extent = data.get("extent", context.extent)
            context.metadata = data.get("metadata", {})
        return context

    def to_dict(self) -> Dict[str, Any]:
        """Convert context to dictionary"""
        return {
            "who": self.who,
            "what": self.what,
            "when": self.when,
            "where": self.where,
            "why": self.why,
            "how": self.how,
            "extent": self.extent,
            "metadata": self.metadata,
        }


# Define the MCP Message class
class MCPMessage:
    """
    Represents a message in the Model Context Protocol
    """

    def __init__(self, message_type: str, content: Any, context: MCPContext = None):
        self.message_id = str(uuid.uuid4())
        self.message_type = message_type
        self.content = content
        self.timestamp = time.time()
        self.context = context or MCPContext()

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "MCPMessage":
        """Create message from dictionary"""
        if not isinstance(data, dict):
            raise ValueError("Message data must be a dictionary")

        message_type = data.get("message_type", "unknown")
        content = data.get("content")
        context_data = data.get("context", {})
        context = MCPContext.from_dict(context_data)

        msg = cls(message_type, content, context)
        msg.message_id = data.get("message_id", msg.message_id)
        msg.timestamp = data.get("timestamp", msg.timestamp)

        return msg

    def to_dict(self) -> Dict[str, Any]:
        """Convert message to dictionary"""
        return {
            "message_id": self.message_id,
            "message_type": self.message_type,
            "content": self.content,
            "timestamp": self.timestamp,
            "context": self.context.to_dict(),
        }


# Define the TNOS MCP Server class
class MCPServer:
    """
    TNOS Model Context Protocol Server implementation
    """

    def __init__(self, host: str = "localhost", port: int = 9000):
        self.host = host
        self.port = port
        self.websocket_port = port + 1  # WebSocket on TCP port + 1
        self.clients = set()
        self.message_handlers = {}
        self.session_contexts = {}  # Store context for each session
        self.logger = logging.getLogger("tnos_mcp_server")

        # Register default message handlers
        self._register_default_handlers()

        # Save PID file
        pid_file = os.path.join(logs_dir, "tnos_mcp_server.pid")
        with open(pid_file, "w") as f:
            f.write(str(os.getpid()))
        logger.info(f"Wrote PID to {pid_file}")

        # Signal handlers for graceful shutdown
        signal.signal(signal.SIGINT, self._signal_handler)
        signal.signal(signal.SIGTERM, self._signal_handler)

    def _signal_handler(self, sig, frame):
        """Handle shutdown signals"""
        self.logger.info("Received shutdown signal, exiting...")
        # Clean up PID file
        pid_file = os.path.join(logs_dir, "tnos_mcp_server.pid")
        if os.path.exists(pid_file):
            os.unlink(pid_file)
        sys.exit(0)

    def _register_default_handlers(self):
        """Register default message handlers"""
        self.register_handler("ping", self._handle_ping)
        self.register_handler("version_negotiation", self._handle_version_negotiation)

    def register_handler(self, message_type: str, handler_func):
        """Register a message handler function"""
        self.message_handlers[message_type] = handler_func
        self.logger.debug(f"Registered handler for message type: {message_type}")

    async def _handle_ping(self, message: MCPMessage, websocket) -> MCPMessage:
        """Handle ping requests"""
        response_context = message.context
        response_context.what = "PingResponse"

        return MCPMessage(
            "pong",
            {"server_time": time.time(), "received_message_id": message.message_id},
            response_context,
        )

    async def _handle_version_negotiation(
        self, message: MCPMessage, websocket
    ) -> MCPMessage:
        """Handle protocol version negotiation"""
        supported_versions = ["1.0", "1.5", "2.0", "2.5", "3.0"]
        client_versions = message.content.get("supported_versions", [])
        preferred_version = message.content.get("preferred_version", "1.0")

        # Find the highest version supported by both
        selected_version = None
        for v in reversed(supported_versions):
            if v in client_versions:
                selected_version = v
                break

        # If no common version, use the client's preferred if supported
        if not selected_version:
            if preferred_version in supported_versions:
                selected_version = preferred_version
            else:
                # Otherwise use the highest supported version
                selected_version = supported_versions[-1]

        response_context = message.context
        response_context.what = "VersionNegotiation"

        return MCPMessage(
            "version_negotiation_response",
            {
                "selected_version": selected_version,
                "server_supported_versions": supported_versions,
            },
            response_context,
        )

    async def handle_client(self, websocket, path):
        """Handle client WebSocket connection"""
        # Generate a unique session ID
        session_id = str(uuid.uuid4())

        # Initialize context for this session
        self.session_contexts[session_id] = MCPContext()
        self.session_contexts[session_id].where = f"Session_{session_id[:8]}"

        # Add client to set
        self.clients.add(websocket)
        self.logger.info(f"New client connected: {session_id}")

        try:
            async for message_data in websocket:
                try:
                    # Parse message data
                    data = json.loads(message_data)

                    # Create MCPMessage from the data
                    message = MCPMessage.from_dict(data)

                    # Update context with session information
                    if message.context.where == "MCP_Server":
                        message.context.where = f"Session_{session_id[:8]}"

                    # Update session context
                    self.session_contexts[session_id] = message.context

                    # Log message receipt
                    self.logger.info(f"Received message type: {message.message_type}")

                    # Look up and call the handler
                    if message.message_type in self.message_handlers:
                        handler = self.message_handlers[message.message_type]
                        response = await handler(message, websocket)

                        # Send response if one was returned
                        if response:
                            await websocket.send(json.dumps(response.to_dict()))
                    else:
                        self.logger.warning(
                            f"No handler for message type: {message.message_type}"
                        )

                        # Send error response
                        error_context = message.context
                        error_context.what = "Error"
                        error_response = MCPMessage(
                            "error",
                            {
                                "error": f"Unsupported message type: {message.message_type}",
                                "received_message_id": message.message_id,
                            },
                            error_context,
                        )
                        await websocket.send(json.dumps(error_response.to_dict()))

                except json.JSONDecodeError:
                    self.logger.error("Failed to decode JSON message")
                    await websocket.send(
                        json.dumps(
                            {
                                "message_type": "error",
                                "content": {"error": "Invalid JSON format"},
                                "context": MCPContext().to_dict(),
                            }
                        )
                    )

                except Exception as e:
                    self.logger.error(f"Error processing message: {e}")
                    await websocket.send(
                        json.dumps(
                            {
                                "message_type": "error",
                                "content": {"error": f"Server error: {str(e)}"},
                                "context": MCPContext().to_dict(),
                            }
                        )
                    )

        except websockets.exceptions.ConnectionClosed:
            self.logger.info(f"Client disconnected: {session_id}")

        finally:
            # Clean up
            self.clients.remove(websocket)
            if session_id in self.session_contexts:
                del self.session_contexts[session_id]

    async def start(self):
        """Start the TNOS MCP server"""
        self.logger.info(
            f"Starting TNOS MCP server on {self.host}:{self.websocket_port} (WebSocket)"
        )

        # WebSocket server
        async with websockets.serve(self.handle_client, self.host, self.websocket_port):
            self.logger.info("Server started successfully")
            # Run indefinitely
            await asyncio.Future()


# Main function
async def main():
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

    # Create and start the server
    server = MCPServer(
        host=args.host or config.get("host", "localhost"),
        port=args.port or config.get("port", 9000),
    )

    await server.start()


if __name__ == "__main__":
    # Run the main async function
    asyncio.run(main())
