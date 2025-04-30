#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
GitHub MCP to TNOS MCP Bridge
------------------------------
This module provides bidirectional communication between the GitHub MCP Server
and the TNOS Model Context Protocol implementation. It translates context and
capabilities between the two systems, allowing GitHub Copilot in VS Code to
access and utilize TNOS capabilities.

# WHO: GitHubTNOSBridge
# WHAT: Bidirectional communication bridge
# WHEN: During communication between GitHub MCP and TNOS MCP
# WHERE: Layer 3 / Bridge
# WHY: To enable integration with GitHub Copilot
# HOW: WebSocket protocol translation with 7D context preservation
# EXTENT: All GitHub MCP to TNOS MCP operations

Created: April 15, 2025
"""

import asyncio
import json
import logging
import os
import sys
import argparse
import websockets
from typing import Any, Dict, List, Optional, Set, Tuple, Union

# Add the project root to Python path to find mcp modules
project_root = os.path.abspath(
    os.path.join(os.path.dirname(__file__), "..", "..", "..")
)
if project_root not in sys.path:
    sys.path.insert(0, project_root)
    print(f"Added {project_root} to Python path")

# Import components from existing MCP modules
try:
    # Import from main MCP modules
    from mcp.config import PathManager, ConfigManager
    from mcp.security_layer3 import TokenManager, EnhancedMessageValidator
    from mcp.network_layer3 import PortManager, RateLimiter
    from mcp.protocol.mcp_protocol import MCPProtocolVersion
except ImportError as e:
    # Alternative imports if main ones fail
    try:
        from mcp import config as config_module
        from mcp import security_layer3 as security_module
        from mcp import network_layer3 as network_module
        from mcp.protocol import mcp_protocol
        
        # Create class references from modules
        PathManager = config_module.PathManager if hasattr(config_module, 'PathManager') else None
        ConfigManager = config_module.ConfigManager if hasattr(config_module, 'ConfigManager') else None
        TokenManager = security_module.TokenManager if hasattr(security_module, 'TokenManager') else None
        EnhancedMessageValidator = security_module.EnhancedMessageValidator if hasattr(security_module, 'EnhancedMessageValidator') else None
        PortManager = network_module.PortManager if hasattr(network_module, 'PortManager') else None
        RateLimiter = network_module.RateLimiter if hasattr(network_module, 'RateLimiter') else None
        MCPProtocolVersion = mcp_protocol.MCPProtocolVersion if hasattr(mcp_protocol, 'MCPProtocolVersion') else None
        
        if not all([PathManager, ConfigManager, TokenManager, EnhancedMessageValidator, 
                   PortManager, RateLimiter, MCPProtocolVersion]):
            raise ImportError("Failed to import required classes from alternative modules")
    except ImportError as e2:
        print(f"Error importing MCP modules: {e} -> {e2}")
        print(f"Current Python path: {sys.path}")
        sys.exit(1)

# Initialize the path manager early
path_manager = PathManager()

# Parse command line arguments
parser = argparse.ArgumentParser(description="GitHub MCP to TNOS MCP Bridge")
parser.add_argument(
    "--tnos-port",
    type=int,
    default=9000,
    help="TCP port for TNOS MCP server (WebSocket port will be TCP port + 1)",
)
parser.add_argument(
    "--github-port", type=int, default=8080, help="Port for GitHub MCP server"
)
parser.add_argument("--log-file", type=str, default=None, help="Path to log file")
parser.add_argument(
    "--context-vector",
    type=str,
    default=None,
    help="JSON string with 7D context vector",
)
parser.add_argument("--debug", action="store_true", help="Enable debug logging")
args = parser.parse_args()

# Define log file path
log_file_path = args.log_file
if log_file_path is None:
    log_file_path = path_manager.get_log_path("github_tnos_bridge")

# Configure logging
log_level = logging.DEBUG if args.debug else logging.INFO
logging.basicConfig(
    level=log_level,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[
        logging.FileHandler(log_file_path),
        logging.StreamHandler(),
    ],
)
logger = logging.getLogger("github_mcp_bridge")

# Calculate WebSocket port (TCP port + 1)
tnos_ws_port = args.tnos_port + 1
tnos_server_uri = f"ws://localhost:{tnos_ws_port}"
logger.info(
    f"TNOS MCP server TCP port: {args.tnos_port}, WebSocket port: {tnos_ws_port}"
)
logger.info(f"GitHub MCP server port: {args.github_port}")

# Setup PYTHONPATH to include project root directory
if path_manager.add_to_python_path():
    logger.info("Added %s to Python path", path_manager.project_root)

# Try to import required packages with detailed error handling
required_packages = {
    "asyncio": "Standard library",
    "websockets": "pip install websockets",
    "json": "Standard library",
    "logging": "Standard library",
    "jsonschema": "pip install jsonschema",
}

missing_packages = []
for package, install_cmd in required_packages.items():
    if package in [
        "json",
        "logging",
        "sys",
        "os",
    ]:  # Standard libraries, already imported
        continue

    try:
        if package == "asyncio":
            import asyncio
        elif package == "websockets":
            import websockets
        elif package == "jsonschema":
            import jsonschema
    except ImportError as e:
        missing_packages.append((package, install_cmd))
        logger.error(f"Failed to import {package}: {e}. Install with: {install_cmd}")

if missing_packages:
    logger.error("STARTUP FAILED - Missing required packages")
    for package, cmd in missing_packages:
        logger.error(f"  - {package}: Install with '{cmd}'")

    # Try to help with the most common issue (websockets)
    if any(pkg[0] == "websockets" for pkg in missing_packages):
        logger.error("\nTroubleshooting for websockets package:")
        logger.error("1. Try installing directly: pip3 install websockets")
        logger.error(
            "2. If using virtualenv, activate it first: source tnos_venv/bin/activate"
        )
        logger.error("3. Check for multiple Python installations: which -a python3")
        logger.error(
            "4. Try installing with specific Python: /path/to/python3 -m pip install websockets"
        )

    sys.exit(1)

# WHO: GithubMCPBridge
# WHAT: Import necessary TNOS MCP components
# WHEN: During bridge initialization
# WHERE: GitHub MCP Bridge
# WHY: To establish connection with TNOS core
# HOW: Import required modules with error handling
# EXTENT: Required for protocol bridge functionality
try:
    # Import the correct MCP components
    from mcp.protocol.mcp_protocol import MCPContext, MCPMessage, MCPProtocolVersion
    from mcp.server_time.server_time import MCPServerTime
    
    # Import Möbius Compression components
    from algorithms.compression.mobius_compression import MobiusCompressor
    from algorithms.compression.context_aware_compression import ContextAwareCompressor
except ImportError as e:
    logger.error(f"Failed to import TNOS MCP components: {e}")
    logger.error(
        "Make sure you're running this script from the project root or have PYTHONPATH set correctly"
    )
    logger.error(f"Current Python path: {sys.path}")
    sys.exit(1)


# WHO: GitHubTNOSBridge
# WHAT: MCP bridge implementation
# WHEN: During GitHub-TNOS communication
# WHERE: Layer 3 / Bridge
# WHY: To enable bidirectional MCP integration
# HOW: Using WebSocket communication with 7D context preservation
# EXTENT: All GitHub-TNOS interactions
class GitHubTNOSBridge:
    """
    Bridge between GitHub MCP Server and TNOS MCP implementation.
    Translates requests, contexts, and capabilities between the two systems.
    """

    def __init__(
        self, tnos_server_uri: str = "ws://localhost:9001", github_mcp_port: int = 8080
    ):
        """
        Initialize the bridge between GitHub MCP and TNOS MCP.

        Args:
            tnos_server_uri: WebSocket URI for the TNOS MCP server
            github_mcp_port: Port number the GitHub MCP server is running on
        """
        # Initialize path and config manager
        self.path_manager = path_manager  # Use the global path manager
        self.config_manager = ConfigManager(self.path_manager)

        # Initialize token manager
        self.token_manager = TokenManager(self.path_manager)

        # Initialize time service
        self.time_service = MCPServerTime()

        self.tnos_server_uri = tnos_server_uri
        self.github_mcp_port = github_mcp_port
        self.tnos_ws_connection = None
        self.github_clients = set()
        self.session_mapping = {}  # Maps GitHub session IDs to TNOS session IDs
        self.tool_mapping = self._build_tool_mapping()
        self.message_validator = EnhancedMessageValidator(self.config_manager)
        self.message_validator.set_token_manager(self.token_manager)
        self.rate_limiter = RateLimiter()

        # Protocol version negotiation
        self.supported_versions = MCPProtocolVersion.SUPPORTED_VERSIONS
        self.current_version = MCPProtocolVersion.get_latest_version()

        # Connection state tracking for improved error handling
        self.connection_state = {
            "github": {
                "connected": False,
                "last_error": None,
                "last_connection_time": None,
                "reconnect_attempts": 0,
            },
            "tnos": {
                "connected": False,
                "last_error": None,
                "last_connection_time": None,
                "reconnect_attempts": 0,
            },
        }

        # Save PID for monitoring
        self._save_pid()

        # Log the TNOS server URI we're connecting to
        logger.info(
            f"Configured to connect to TNOS MCP server at {self.tnos_server_uri}"
        )

    def _save_pid(self):
        """Save the current process ID to a file for monitoring"""
        pid_path = self.path_manager.get_pid_path("github_bridge")
        try:
            with open(pid_path, "w") as f:
                f.write(str(os.getpid()))
            logger.info(f"Wrote PID to {pid_path}")
        except Exception as e:
            logger.error(f"Failed to write PID file: {e}")

    def _build_tool_mapping(self) -> Dict[str, str]:
        """
        Build a mapping between GitHub MCP tool names and TNOS capabilities.
        This allows translation of tool calls between the systems.

        Returns:
            Dict mapping GitHub MCP tool names to TNOS capabilities
        """
        # This mapping connects GitHub tools to TNOS capabilities
        # Format: "github_tool_name": "tnos_capability_name"
        return {
            # Repository operations
            "get_file_contents": "file_read_operation",
            "list_branches": "version_control_operation",
            "create_or_update_file": "file_write_operation",
            "search_repositories": "search_operation",
            # Issue operations
            "get_issue": "task_management_operation",
            "create_issue": "task_creation_operation",
            "list_issues": "task_listing_operation",
            # Pull request operations
            "get_pull_request": "code_review_operation",
            "list_pull_requests": "code_review_listing_operation",
            "get_pull_request_files": "code_diff_operation",
            # Code scanning
            "list_code_scanning_alerts": "security_analysis_operation",
            "get_code_scanning_alert": "security_detail_operation",
        }

    def translate_github_to_tnos_context(
        self, github_request: Dict[str, Any]
    ) -> MCPContext:
        """
        Translate GitHub MCP request into TNOS 7D context.

        Args:
            github_request: GitHub MCP request object

        Returns:
            TNOS MCPContext object with 7D dimensions populated
        """
        # Create a new context with current time
        context = MCPContext()

        # Extract information from GitHub request
        tool_name = github_request.get("name", "")
        params = github_request.get("parameters", {})

        # WHO: The actor making the request (GitHub Copilot or user)
        context.who = "GitHub_Copilot"

        # WHAT: The capability being requested (mapped from GitHub tool)
        context.what = self.tool_mapping.get(tool_name, "unknown_operation")

        # WHEN: Current timestamp (get from time service for proper compression)
        current_time = self.time_service.get_current_time()
        context.when = current_time

        # WHERE: The location in the repository
        repo_info = f"{params.get('owner', '')}/{params.get('repo', '')}"
        if "path" in params:
            repo_info += f"/{params.get('path', '')}"
        context.where = repo_info if repo_info != "/" else "system"

        # WHY: Purpose of the operation
        context.why = f"GitHub_operation_{tool_name}"

        # HOW: The method being used
        context.how = "github_mcp_bridge"

        # EXTENT: Scope of the operation
        # Determine based on parameters if it's a single resource or multiple
        if any(param in params for param in ["perPage", "page", "list"]):
            context.extent = "multiple_resources"
        else:
            context.extent = "single_resource"

        # Add original request as metadata for reference
        context.set_metadata("original_github_request", json.dumps(github_request))
        context.set_metadata("github_tool", tool_name)
        context.set_metadata("mcp_version", self.current_version)

        return context

    def translate_tnos_to_github_response(
        self, tnos_response: MCPMessage, original_request: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Translate TNOS response back to GitHub MCP format.

        Args:
            tnos_response: Response from TNOS MCP server
            original_request: Original GitHub request that triggered this response

        Returns:
            Response formatted for GitHub MCP server
        """
        # Extract the result content from TNOS response
        result_content = tnos_response.content

        # If TNOS response is a string, try to parse as JSON
        if isinstance(result_content, str):
            try:
                result_content = json.loads(result_content)
            except json.JSONDecodeError:
                # If not valid JSON, wrap in a content field
                result_content = {"content": result_content}

        # Extract 7D context dimensions for metadata
        context_metadata = {
            "who": tnos_response.context.who,
            "what": tnos_response.context.what,
            "when": tnos_response.context.when,
            "where": tnos_response.context.where,
            "why": tnos_response.context.why,
            "how": tnos_response.context.how,
            "extent": tnos_response.context.extent,
        }

        # Add the original context information as metadata
        response = {
            "result": result_content,
            "metadata": {
                "tnos_context": context_metadata,
                "mcp_version": self.current_version,
                "timestamp": self.time_service.get_current_time(),
                "operation": original_request.get("name", ""),
            },
        }

        return response

    async def connect_to_tnos_mcp(self) -> bool:
        """
        Connect to the TNOS MCP server

        Returns:
            True if connected successfully, False otherwise
        """
        try:
            # Check if TNOS MCP server is available
            host = self.tnos_server_uri.split("://")[1].split(":")[0]
            port = int(self.tnos_server_uri.split(":")[-1])

            if not PortManager.check_port_in_use(host, port):
                error_msg = f"TNOS MCP server not available at {host}:{port}"
                logger.error(error_msg)
                self.connection_state["tnos"]["last_error"] = error_msg
                return False

            # Connect to the server
            self.tnos_ws_connection = await websockets.connect(self.tnos_server_uri)

            # Update connection state
            self.connection_state["tnos"]["connected"] = True
            self.connection_state["tnos"][
                "last_connection_time"
            ] = self.time_service.get_current_time()
            self.connection_state["tnos"]["reconnect_attempts"] = 0

            logger.info(f"Connected to TNOS MCP server at {self.tnos_server_uri}")

            # Negotiate protocol version
            await self._negotiate_protocol_version()

            return True

        except Exception as e:
            error_msg = f"Failed to connect to TNOS MCP server: {e}"
            logger.error(error_msg)

            # Update connection state
            self.connection_state["tnos"]["connected"] = False
            self.connection_state["tnos"]["last_error"] = str(e)
            self.connection_state["tnos"]["reconnect_attempts"] += 1

            return False

    async def _negotiate_protocol_version(self):
        """
        Negotiate MCP protocol version with the TNOS MCP server
        """
        try:
            # Create a context vector with proper 7D context for negotiation
            context = MCPContext()
            context.who = "GitHubTNOSBridge"
            context.what = "ProtocolNegotiation"
            context.when = self.time_service.get_current_time()
            context.where = "Layer3_Bridge"
            context.why = "VersionCompatibility"
            context.how = "WebSocket"
            context.extent = "ProtocolCompatibility"

            # Create a version negotiation message
            negotiation_msg = MCPMessage(
                message_type="version_negotiation",
                context=context,
                content={
                    "supported_versions": self.supported_versions,
                    "preferred_version": self.current_version,
                },
            )

            # Send the negotiation message
            await self.tnos_ws_connection.send(negotiation_msg.to_json())

            # Wait for a response
            response_str = await asyncio.wait_for(
                self.tnos_ws_connection.recv(), timeout=5.0
            )

            # Parse the response
            tnos_response = MCPMessage.from_json(response_str)

            if tnos_response.message_type == "version_negotiation_response":
                if isinstance(tnos_response.content, dict):
                    server_version = tnos_response.content.get("selected_version")
                else:
                    try:
                        content_dict = json.loads(tnos_response.content)
                        server_version = content_dict.get("selected_version")
                    except (json.JSONDecodeError, TypeError):
                        server_version = None

                if server_version:
                    # Update the current version to the negotiated version
                    self.current_version = server_version
                    logger.info(
                        f"Negotiated MCP protocol version: {self.current_version}"
                    )
                else:
                    logger.warning("Server did not specify a protocol version")
            else:
                logger.warning("Unexpected response to version negotiation")

        except Exception as e:
            logger.error(f"Error during protocol version negotiation: {e}")
            # Continue with the default version

    async def handle_github_request(self, websocket, path):
        """
        Handle incoming WebSocket connections from GitHub MCP server.

        Args:
            websocket: WebSocket connection from GitHub MCP
            path: WebSocket connection path
        """
        # Add the client to our set
        client_id = id(websocket)
        self.github_clients.add(websocket)
        logger.info(f"New GitHub MCP client connected: {client_id}")

        # Update connection state
        self.connection_state["github"]["connected"] = True
        self.connection_state["github"][
            "last_connection_time"
        ] = self.time_service.get_current_time()

        try:
            # Connect to TNOS MCP server if not already connected
            if (
                not self.tnos_ws_connection
                or not self.connection_state["tnos"]["connected"]
            ):
                if not await self.connect_to_tnos_mcp():
                    # Send error response to GitHub client
                    await websocket.send(
                        json.dumps(
                            {
                                "error": self.connection_state["tnos"]["last_error"],
                                "status": "error",
                                "errorType": "connection_failure",
                                "recoverable": False,
                                "recommendations": [
                                    "Ensure TNOS MCP server is running (scripts/start_tnos_github_integration.sh)",
                                    "Check if another process is using port 8888",
                                    "Verify network connectivity to TNOS MCP server",
                                ],
                            }
                        )
                    )
                    return

            # Handle messages from GitHub MCP
            async for message in websocket:
                try:
                    # Parse the message from GitHub MCP
                    github_request = json.loads(message)
                    logger.info(
                        f"Received GitHub MCP request: {github_request.get('name', 'unknown')}"
                    )

                    # Validate the incoming message format
                    is_valid, error_message = (
                        self.message_validator.validate_github_request(github_request)
                    )
                    if not is_valid:
                        logger.error(f"Invalid request format: {error_message}")
                        await websocket.send(
                            json.dumps(
                                {
                                    "error": error_message,
                                    "status": "error",
                                    "errorType": "validation_failure",
                                    "recoverable": True,
                                }
                            )
                        )
                        continue

                    # Check if the tool is supported
                    tool_name = github_request.get("name", "")
                    if (
                        not tool_name.startswith("tnos_")
                        and tool_name not in self.tool_mapping
                    ):
                        error_msg = f"Unsupported tool: {tool_name}"
                        logger.warning(error_msg)
                        await websocket.send(
                            json.dumps(
                                {
                                    "error": error_msg,
                                    "status": "error",
                                    "errorType": "unsupported_tool",
                                    "recoverable": True,
                                    "suggestions": list(self.tool_mapping.keys())
                                    + [
                                        t
                                        for t in self.config_manager.config.get(
                                            "toolDefinitions", []
                                        )
                                        if t.get("name", "").startswith("tnos_")
                                    ],
                                }
                            )
                        )
                        continue

                    # Check rate limit
                    if not self.rate_limiter.check_rate_limit(tool_name):
                        error_msg = f"Rate limit exceeded for tool: {tool_name}"
                        logger.warning(error_msg)
                        await websocket.send(
                            json.dumps(
                                {
                                    "error": error_msg,
                                    "status": "error",
                                    "errorType": "rate_limit_exceeded",
                                    "recoverable": True,
                                    "remaining_quota": self.rate_limiter.get_remaining_quota(
                                        tool_name
                                    ),
                                }
                            )
                        )
                        continue

                    # Translate to TNOS context format
                    tnos_context = self.translate_github_to_tnos_context(github_request)

                    # Create TNOS MCP message
                    tnos_message = MCPMessage(
                        message_type="request",
                        context=tnos_context,
                        content=github_request.get("parameters", {}),
                    )

                    # Send to TNOS MCP server
                    try:
                        await self.tnos_ws_connection.send(tnos_message.to_json())
                        logger.info(
                            f"Sent request to TNOS MCP for operation: {tnos_context.what}"
                        )
                    except websockets.exceptions.ConnectionClosed:
                        # Connection to TNOS MCP server was closed, try to reconnect
                        logger.warning(
                            "Connection to TNOS MCP server was closed, attempting to reconnect..."
                        )

                        if await self.connect_to_tnos_mcp():
                            await self.tnos_ws_connection.send(tnos_message.to_json())
                        else:
                            await websocket.send(
                                json.dumps(
                                    {
                                        "error": "Connection to TNOS MCP server lost and reconnection failed",
                                        "status": "error",
                                        "errorType": "connection_failure",
                                        "recoverable": False,
                                    }
                                )
                            )
                            continue

                    # Wait for response from TNOS with timeout
                    try:
                        # Set a reasonable timeout for waiting for a response
                        tnos_response_str = await asyncio.wait_for(
                            self.tnos_ws_connection.recv(), timeout=30.0
                        )
                        tnos_response = MCPMessage.from_json(tnos_response_str)
                    except asyncio.TimeoutError:
                        error_msg = "Timeout waiting for TNOS MCP server response"
                        logger.error(error_msg)
                        await websocket.send(
                            json.dumps(
                                {
                                    "error": error_msg,
                                    "status": "error",
                                    "errorType": "timeout",
                                    "recoverable": True,
                                }
                            )
                        )
                        continue
                    except Exception as e:
                        error_msg = (
                            f"Error receiving response from TNOS MCP server: {e}"
                        )
                        logger.error(error_msg)
                        await websocket.send(
                            json.dumps(
                                {
                                    "error": error_msg,
                                    "status": "error",
                                    "errorType": "communication_failure",
                                    "recoverable": True,
                                }
                            )
                        )
                        continue

                    # Translate TNOS response to GitHub format
                    github_response = self.translate_tnos_to_github_response(
                        tnos_response, github_request
                    )

                    # Send response back to GitHub MCP
                    await websocket.send(json.dumps(github_response))
                    logger.info(
                        f"Sent response back to GitHub MCP client for {github_request.get('name', 'unknown')}"
                    )

                except json.JSONDecodeError:
                    error_msg = "Invalid JSON in GitHub MCP request"
                    logger.error(error_msg)
                    await websocket.send(
                        json.dumps(
                            {
                                "error": error_msg,
                                "status": "error",
                                "errorType": "invalid_json",
                                "recoverable": True,
                            }
                        )
                    )
                except Exception as e:
                    error_msg = f"Error handling GitHub MCP request: {e}"
                    logger.error(error_msg)
                    logger.exception("Detailed exception information:")
                    await websocket.send(
                        json.dumps(
                            {
                                "error": error_msg,
                                "status": "error",
                                "errorType": "general_error",
                                "recoverable": True,
                            }
                        )
                    )

        except websockets.exceptions.ConnectionClosed:
            logger.info(f"GitHub MCP client disconnected: {client_id}")
        except Exception as e:
            logger.error(f"Unexpected error in handle_github_request: {e}")
            logger.exception("Detailed exception information:")
        finally:
            # Clean up when the connection is closed
            if websocket in self.github_clients:
                self.github_clients.remove(websocket)

            # Update connection state
            if not self.github_clients:
                self.connection_state["github"]["connected"] = False

                # Close TNOS connection if no clients are connected
                if (
                    self.tnos_ws_connection
                    and self.connection_state["tnos"]["connected"]
                ):
                    await self.tnos_ws_connection.close()
                    self.connection_state["tnos"]["connected"] = False
                    logger.info(
                        "Closed connection to TNOS MCP server (no clients connected)"
                    )

    async def start(self):
        """
        # WHO: GitHubTNOSBridge.start
        # WHAT: Start the bridge server
        # WHEN: During system initialization
        # WHERE: Layer 3 / Bridge
        # WHY: To begin handling connections
        # HOW: Using WebSocket server
        # EXTENT: All bridge operations

        Start the bridge server to handle connections from GitHub MCP.
        """
        try:
            # Check if the port is already in use
            if PortManager.check_port_in_use("localhost", self.github_mcp_port):
                logger.error(
                    f"Port {self.github_mcp_port} is already in use. "
                    "Another instance of the bridge may be running."
                )
                return

            server = await websockets.serve(
                self.handle_github_request, "localhost", self.github_mcp_port
            )
            logger.info(
                f"GitHub-TNOS MCP Bridge running on port {self.github_mcp_port}"
            )

            # Keep the server running
            await server.wait_closed()
        except Exception as e:
            logger.error(f"Failed to start GitHub-TNOS MCP Bridge: {e}")

    async def negotiate_protocol_version(
        self,
        requested_version: str,
        min_acceptable_version: str = "1.0",
        context: Optional[Any] = None,
    ) -> str:
        """
        # WHO: GitHubTNOSBridge.negotiate_protocol_version
        # WHAT: Negotiate protocol version compatibility
        # WHEN: During bridge initialization
        # WHERE: Layer 3 / Protocol
        # WHY: To ensure compatibility between systems
        # HOW: Version comparison and negotiation
        # EXTENT: All bridge communication

        Negotiate a compatible protocol version with the TNOS MCP server.
        This function is called by the bridge starter to establish a compatible
        protocol version for communication.

        Args:
            requested_version: Preferred protocol version to use
            min_acceptable_version: Minimum acceptable protocol version
            context: Optional context for the negotiation

        Returns:
            The negotiated protocol version as a string
        """
        # Set up a local version if this is called before connection
        if (
            not self.tnos_ws_connection
            or not self.connection_state["tnos"]["connected"]
        ):
            try:
                if not await self.connect_to_tnos_mcp():
                    logger.error("Failed to connect for protocol negotiation")
                    return requested_version  # Return requested version if we can't connect
            except Exception as e:
                logger.error(f"Connection error during protocol negotiation: {e}")
                return requested_version

        try:
            # Create a version negotiation message with proper 7D context
            request_context = MCPContext()
            request_context.who = "GitHubTNOSBridge"
            request_context.what = "ProtocolNegotiation"
            request_context.when = self.time_service.get_current_time()
            request_context.where = "Layer3_Bridge"
            request_context.why = "VersionCompatibility"
            request_context.how = f"MCP_{requested_version}"
            request_context.extent = "ProtocolCompatibility"

            # If external context is provided, use its values
            if context:
                for dim in ["who", "what", "when", "where", "why", "how", "extent"]:
                    if hasattr(context, dim):
                        setattr(request_context, dim, getattr(context, dim))

            negotiation_msg = MCPMessage(
                message_type="version_negotiation",
                context=request_context,
                content={
                    "supported_versions": self.supported_versions,
                    "preferred_version": requested_version,
                    "minimum_version": min_acceptable_version,
                },
            )

            # Send the negotiation message
            await self.tnos_ws_connection.send(negotiation_msg.to_json())
            logger.info(
                f"Sent protocol version negotiation request: {requested_version}"
            )

            # Wait for a response with timeout
            response_str = await asyncio.wait_for(
                self.tnos_ws_connection.recv(), timeout=5.0
            )

            # Parse the response
            tnos_response = MCPMessage.from_json(response_str)

            if tnos_response.message_type == "version_negotiation_response":
                content = tnos_response.content
                if isinstance(content, str):
                    try:
                        content = json.loads(content)
                    except json.JSONDecodeError:
                        content = {"selected_version": requested_version}

                server_version = content.get("selected_version", requested_version)

                if server_version:
                    # Update the current version to the negotiated version
                    self.current_version = server_version
                    logger.info(
                        f"Successfully negotiated protocol version: {self.current_version}"
                    )
                    return server_version
                else:
                    logger.warning(
                        "Server did not specify a version, using requested version"
                    )
                    return requested_version
            else:
                logger.warning(
                    f"Unexpected response type: {tnos_response.message_type}"
                )
                return requested_version

        except asyncio.TimeoutError:
            logger.error("Timeout during protocol version negotiation")
            return requested_version
        except Exception as e:
            logger.error(f"Error during protocol negotiation: {e}")
            return requested_version

    async def stop(self) -> bool:
        """
        # WHO: GitHubTNOSBridge.stop
        # WHAT: Gracefully stop the bridge
        # WHEN: During shutdown
        # WHERE: Layer 3 / Bridge
        # WHY: To ensure clean termination
        # HOW: Closing connections and resources
        # EXTENT: All bridge resources

        Gracefully stop the bridge, closing all connections
        and cleaning up resources.

        Returns:
            True if shutdown was successful, False otherwise
        """
        logger.info("Shutting down GitHub-TNOS MCP Bridge")

        try:
            # Close all GitHub client connections
            close_tasks = []
            for client in self.github_clients:
                try:
                    close_tasks.append(client.close())
                except Exception as e:
                    logger.error(f"Error closing GitHub client: {e}")

            if close_tasks:
                # Wait for all clients to close with timeout
                await asyncio.wait(close_tasks, timeout=5.0)

            # Clear the client set
            self.github_clients.clear()
            self.connection_state["github"]["connected"] = False

            # Close TNOS MCP connection if open
            if self.tnos_ws_connection and self.connection_state["tnos"]["connected"]:
                try:
                    await self.tnos_ws_connection.close()
                    self.tnos_ws_connection = None
                    self.connection_state["tnos"]["connected"] = False
                    logger.info("Closed connection to TNOS MCP server")
                except Exception as e:
                    logger.error(f"Error closing TNOS MCP connection: {e}")

            # Remove PID file if it exists
            pid_path = self.path_manager.get_pid_path("github_bridge")
            if os.path.exists(pid_path):
                try:
                    os.unlink(pid_path)
                    logger.info(f"Removed PID file: {pid_path}")
                except Exception as e:
                    logger.error(f"Error removing PID file: {e}")

            logger.info("Bridge shutdown completed successfully")
            return True

        except Exception as e:
            logger.error(f"Error during bridge shutdown: {e}")
            return False

    async def recover(self) -> bool:
        """
        # WHO: GitHubTNOSBridge.recover
        # WHAT: Attempt to recover bridge functionality
        # WHEN: After communication failure
        # WHERE: Layer 3 / Bridge
        # WHY: To restore operation after failure
        # HOW: Reconnection and state restoration
        # EXTENT: Critical bridge components

        Attempt to recover the bridge after a failure.
        This is called by the health monitor when issues are detected.

        Returns:
            True if recovery was successful, False otherwise
        """
        logger.info("Attempting bridge recovery")

        try:
            # Reconnect to TNOS MCP if needed
            if (
                not self.tnos_ws_connection
                or not self.connection_state["tnos"]["connected"]
            ):
                logger.info("Reconnecting to TNOS MCP server")
                if not await self.connect_to_tnos_mcp():
                    logger.error(
                        "Recovery failed: Unable to connect to TNOS MCP server"
                    )
                    return False

            # Report successful recovery
            logger.info("Bridge recovery completed successfully")
            return True

        except Exception as e:
            logger.error(f"Error during bridge recovery: {e}")
            return False


def register_tnos_tools_with_github_mcp():
    """
    # WHO: register_tnos_tools_with_github_mcp
    # WHAT: Generate TNOS tool registrations
    # WHEN: During GitHub MCP initialization
    # WHERE: Layer 3 / Bridge
    # WHY: To expose TNOS capabilities
    # HOW: JSON configuration generation
    # EXTENT: All TNOS tools

    Generate a configuration that registers TNOS capabilities as tools
    in the GitHub MCP server.

    Returns:
        Configuration dictionary for GitHub MCP server
    """
    # Define TNOS capabilities to expose to GitHub Copilot
    tnos_capabilities = [
        {
            "name": "tnos_compression",
            "description": "Compress and optimize data using TNOS Möbius Compression",
            "parameters": {
                "type": "object",
                "properties": {
                    "content": {"type": "string", "description": "Content to compress"},
                    "compressionLevel": {
                        "type": "integer",
                        "description": "Compression level (1-10)",
                    },
                    "preserveContext": {
                        "type": "boolean",
                        "description": "Whether to preserve 7D context during compression",
                    },
                },
                "required": ["content"],
            },
        },
        {
            "name": "tnos_context_analysis",
            "description": "Analyze content using TNOS 7D contextual framework",
            "parameters": {
                "type": "object",
                "properties": {
                    "content": {"type": "string", "description": "Content to analyze"},
                    "dimensionFocus": {
                        "type": "string",
                        "description": "Dimension to focus on (WHO, WHAT, WHEN, WHERE, WHY, HOW, EXTENT)",
                        "enum": [
                            "WHO",
                            "WHAT",
                            "WHEN",
                            "WHERE",
                            "WHY",
                            "HOW",
                            "EXTENT",
                        ],
                    },
                },
                "required": ["content"],
            },
        },
        {
            "name": "tnos_neural_hybrid_reasoning",
            "description": "Apply TNOS Neural Hybrid Cognitive Model for reasoning",
            "parameters": {
                "type": "object",
                "properties": {
                    "query": {
                        "type": "string",
                        "description": "Query or problem statement",
                    },
                    "context": {
                        "type": "string",
                        "description": "Additional context information",
                    },
                    "reasoningType": {
                        "type": "string",
                        "description": "Type of reasoning to apply",
                        "enum": ["bayesian", "rule-based", "neural", "hybrid"],
                    },
                },
                "required": ["query"],
            },
        },
        {
            "name": "tnos_code_optimization",
            "description": "Optimize code using TNOS understanding of patterns and efficiency",
            "parameters": {
                "type": "object",
                "properties": {
                    "code": {"type": "string", "description": "Code to optimize"},
                    "language": {
                        "type": "string",
                        "description": "Programming language",
                    },
                    "optimizationGoal": {
                        "type": "string",
                        "description": "Optimization goal (speed, memory, readability)",
                        "enum": ["speed", "memory", "readability"],
                    },
                },
                "required": ["code", "language"],
            },
        },
        {
            "name": "tnos_formula_registry",
            "description": "Access TNOS formula registry for mathematical and computational formulas",
            "parameters": {
                "type": "object",
                "properties": {
                    "formulaType": {
                        "type": "string",
                        "description": "Type of formula to retrieve",
                        "enum": ["mathematical", "computational", "scientific", "all"],
                    },
                    "search": {
                        "type": "string",
                        "description": "Search term for finding formulas",
                    },
                    "domain": {
                        "type": "string",
                        "description": "Domain area for the formula",
                    },
                },
                "required": ["formulaType"],
            },
        },
        {
            "name": "tnos_7d_context_vector",
            "description": "Generate or transform 7D context vectors for advanced context operations",
            "parameters": {
                "type": "object",
                "properties": {
                    "content": {
                        "type": "string",
                        "description": "Content to generate context vector for",
                    },
                    "operation": {
                        "type": "string",
                        "description": "Context operation to perform",
                        "enum": [
                            "generate",
                            "transform",
                            "combine",
                            "extract",
                            "analyze",
                        ],
                    },
                    "dimensionalFocus": {
                        "type": "array",
                        "description": "Specific dimensions to focus on",
                        "items": {
                            "type": "string",
                            "enum": [
                                "WHO",
                                "WHAT",
                                "WHEN",
                                "WHERE",
                                "WHY",
                                "HOW",
                                "EXTENT",
                            ],
                        },
                    },
                    "contextVector": {
                        "type": "object",
                        "description": "Existing context vector to use in operation (for transform/combine operations)",
                    },
                },
                "required": ["content", "operation"],
            },
        },
        {
            "name": "tnos_mobius_formula_execution",
            "description": "Execute Möbius equations and formulas from the TNOS Formula Registry",
            "parameters": {
                "type": "object",
                "properties": {
                    "formulaName": {
                        "type": "string",
                        "description": "Name of the formula to execute",
                    },
                    "parameters": {
                        "type": "object",
                        "description": "Parameters to pass to the formula",
                    },
                    "context": {
                        "type": "object",
                        "description": "Context information for formula execution",
                    },
                    "returnFormat": {
                        "type": "string",
                        "description": "Format to return the results in",
                        "enum": ["json", "text", "detailed", "simplified"],
                    },
                },
                "required": ["formulaName"],
            },
        },
    ]

    return {
        "toolDefinitions": tnos_capabilities,
        "serverSettings": {"bridgeUri": "ws://localhost:8080"},
    }


async def main():
    """
    # WHO: main
    # WHAT: Main entry point
    # WHEN: During bridge execution
    # WHERE: GitHub MCP Bridge
    # WHY: To initialize the bridge
    # HOW: Initialize and start bridge instance
    # EXTENT: Complete application lifecycle

    Main entry point for the GitHub-TNOS MCP Bridge.
    """
    # Create and start the bridge
    bridge = GitHubTNOSBridge(tnos_server_uri, args.github_port)
    await bridge.start()


if __name__ == "__main__":
    try:
        # Create and initialize the bridge
        bridge = GitHubTNOSBridge(tnos_server_uri, args.github_port)

        # Parse context vector if provided
        if args.context_vector:
            try:
                context_dict = json.loads(args.context_vector)
                logger.info(f"Using provided context vector: {context_dict}")
                # Initialize with the provided context if needed
            except json.JSONDecodeError:
                logger.error(f"Failed to parse context vector: {args.context_vector}")

        # Log startup information
        logger.info("GitHub TNOS Bridge starting up")
        logger.info(f"Connecting to TNOS MCP server at {tnos_server_uri}")
        logger.info(f"GitHub MCP port: {args.github_port}")

        # Start the bridge
        asyncio.run(bridge.start())
    except KeyboardInterrupt:
        logger.info("Bridge shutdown requested by user")
        sys.exit(0)
    except Exception as e:
        logger.error(f"Bridge startup failed: {e}", exc_info=True)
        sys.exit(1)
