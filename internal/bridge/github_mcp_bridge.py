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
import time
import uuid
import argparse
import websockets
from typing import Any, Dict, List, Optional, Tuple, Union

# Add project root to Python path
project_root = os.path.abspath(
    os.path.join(os.path.dirname(__file__), "..", "..", "..")
)
if project_root not in sys.path:
    sys.path.insert(0, project_root)

# Add mcp directory to Python path
mcp_path = os.path.join(project_root, "mcp")
if mcp_path not in sys.path:
    sys.path.insert(0, mcp_path)

# Import MCP components
try:
    # Try direct imports first
    from mcp.protocol.mcp_protocol import MCPContext, MCPMessage, MCPProtocolVersion
    from mcp.server.mcp_server_time import MCPServerTime
    from mcp.integration.mcp_integration import TNOSLayerIntegration
    from mcp.integration.mobius_compression_mcp import get_mcp_handler
except ImportError as e:
    try:
        # Try alternate import structure
        from protocol.mcp_protocol import MCPContext, MCPMessage, MCPProtocolVersion
        from server.mcp_server_time import MCPServerTime
        from integration.mcp_integration import TNOSLayerIntegration
        from integration.mobius_compression_mcp import get_mcp_handler
    except ImportError as e2:
        sys.stderr.write(f"Failed to import MCP modules: {e} -> {e2}\n")
        sys.stderr.write(f"Current Python path: {sys.path}\n")
        sys.exit(1)

# Configure logging
log_directory = os.path.join(project_root, "logs")
if not os.path.exists(log_directory):
    try:
        os.makedirs(log_directory)
    except Exception as e:
        sys.stderr.write(f"Failed to create log directory: {e}\n")
        log_directory = os.getcwd()

log_file_path = os.path.join(log_directory, f"github_tnos_bridge_{time.strftime('%Y%m%d_%H%M%S')}.log")
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[
        logging.FileHandler(log_file_path),
        logging.StreamHandler(sys.stdout),
    ],
)
logger = logging.getLogger("github_mcp_bridge")

# Parse command line arguments
parser = argparse.ArgumentParser(description="GitHub MCP to TNOS MCP Bridge")
parser.add_argument(
    "--tnos-port",
    type=int,
    default=9000,
    help="TCP port for TNOS MCP server (WebSocket port will be TCP port + 1)",
)
parser.add_argument(
    "--github-port", 
    type=int, 
    default=8080, 
    help="Port for GitHub MCP server"
)
parser.add_argument(
    "--log-level", 
    choices=["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"],
    default="INFO",
    help="Set logging level"
)
parser.add_argument(
    "--context-vector",
    type=str,
    default=None,
    help="JSON string with 7D context vector"
)
args = parser.parse_args()

# Set log level
log_level = getattr(logging, args.log_level)
logger.setLevel(log_level)
logging.getLogger().setLevel(log_level)

# Calculate WebSocket port (TCP port + 1)
tnos_ws_port = args.tnos_port + 1
tnos_server_uri = f"ws://localhost:{tnos_ws_port}"
logger.info(f"TNOS MCP server TCP port: {args.tnos_port}, WebSocket port: {tnos_ws_port}")
logger.info(f"GitHub MCP server port: {args.github_port}")

# WHO: ContextVector7D
# WHAT: 7D context vector implementation
# WHEN: During context creation and translation
# WHERE: Layer 3 / Bridge
# WHY: To maintain TNOS 7D context format
# HOW: Object-oriented context representation
# EXTENT: All context operations
class ContextVector7D:
    """
    Represents a 7D context vector following TNOS standards.
    """
    def __init__(
        self, 
        who: str = "System",
        what: str = "Transform",
        when: Union[str, float] = None,
        where: str = "MCP_Bridge",
        why: str = "Protocol_Compliance",
        how: str = "Context_Translation",
        extent: Union[str, float] = 1.0,
    ):
        """Initialize a 7D context vector with default values."""
        self.who = who
        self.what = what
        self.when = when if when is not None else time.time()
        self.where = where
        self.why = why
        self.how = how
        self.extent = extent
        self.metadata = {}
        
    def to_dict(self) -> Dict[str, Any]:
        """Convert the context vector to a dictionary."""
        return {
            "WHO": self.who,
            "WHAT": self.what,
            "WHEN": self.when,
            "WHERE": self.where,
            "WHY": self.why,
            "HOW": self.how,
            "EXTENT": self.extent,
            "metadata": self.metadata
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'ContextVector7D':
        """Create a context vector from a dictionary."""
        context = cls()
        
        # Extract dimensions with appropriate defaults
        context.who = data.get("WHO", data.get("who", "System"))
        context.what = data.get("WHAT", data.get("what", "Transform"))
        context.when = data.get("WHEN", data.get("when", time.time()))
        context.where = data.get("WHERE", data.get("where", "MCP_Bridge"))
        context.why = data.get("WHY", data.get("why", "Protocol_Compliance"))
        context.how = data.get("HOW", data.get("how", "Context_Translation"))
        context.extent = data.get("EXTENT", data.get("extent", 1.0))
        
        # Extract metadata if available
        if "metadata" in data and isinstance(data["metadata"], dict):
            context.metadata = data["metadata"]
            
        return context


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
        self.tnos_server_uri = tnos_server_uri
        self.github_mcp_port = github_mcp_port
        self.tnos_ws_connection = None
        self.github_clients = set()
        self.session_mapping = {}  # Maps GitHub session IDs to TNOS session IDs
        self.tool_mapping = self._build_tool_mapping()
        
        # Initialize MCP time service
        self.time_service = MCPServerTime()
        
        # Protocol version negotiation
        self.supported_versions = MCPProtocolVersion.SUPPORTED_VERSIONS
        self.current_version = MCPProtocolVersion.get_latest_version()
        self.min_version = MCPProtocolVersion.get_min_supported_version()

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
        
        # Message buffer for handling disconnections
        self.message_buffer = {
            "github": [],
            "tnos": []
        }
        
        # Rate limiting settings
        self.rate_limits = {
            "default": {"requests": 60, "period": 60},
            "compression": {"requests": 20, "period": 60},
            "formula": {"requests": 30, "period": 60},
        }
        self.request_counts = {}

        # Save PID for monitoring
        self._save_pid()

        # Log the TNOS server URI we're connecting to
        logger.info(f"Initialized bridge to connect to TNOS MCP server at {self.tnos_server_uri}")

    def _save_pid(self):
        """Save the current process ID to a file for monitoring"""
        pid_path = os.path.join(project_root, "logs", "github_bridge.pid")
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
            # Formula Registry operations
            "tnos_formula_registry": "formula_registry_query",
            "tnos_mobius_formula_execution": "formula_execute",
            
            # Context operations
            "tnos_context_analysis": "context_analyze",
            "tnos_7d_context_vector": "context_vector_generate",
            
            # Compression operations
            "tnos_compression": "mobius_compression",
            
            # Code operations
            "tnos_code_optimization": "code_optimize",
            "tnos_neural_hybrid_reasoning": "neural_hybrid_reason",
            
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

    # WHO: GitHubTNOSBridge.bridgeMCPContext
    # WHAT: Convert GitHub context to TNOS 7D context
    # WHEN: During tool request handling
    # WHERE: MCP_Bridge
    # WHY: To preserve context across systems
    # HOW: Context translation with dimension mapping
    # EXTENT: All tool requests
    def bridgeMCPContext(self, github_context: Dict[str, Any]) -> ContextVector7D:
        """
        Bridge GitHub MCP context format to TNOS 7D context.
        This is a critical function for maintaining context awareness.
        
        Args:
            github_context: Context from GitHub MCP request
            
        Returns:
            TNOS 7D Context Vector
        """
        # Create a 7D context vector with defaults
        context_vector = ContextVector7D()
        
        # Extract GitHub context information if available
        if github_context:
            # Map GitHub identity to WHO dimension
            if "identity" in github_context:
                context_vector.who = github_context.get("identity") or "GitHub_Copilot"
            
            # Map GitHub operation to WHAT dimension
            if "operation" in github_context:
                context_vector.what = github_context.get("operation") or "Transform"
            
            # Map GitHub timestamp to WHEN dimension with compression
            if "timestamp" in github_context:
                timestamp = github_context.get("timestamp")
                if isinstance(timestamp, (int, float)):
                    context_vector.when = self.time_service.compress_time(timestamp)
                else:
                    context_vector.when = self.time_service.get_current_time()
            else:
                context_vector.when = self.time_service.get_current_time()
            
            # Map GitHub location to WHERE dimension
            if "location" in github_context:
                context_vector.where = github_context.get("location") or "MCP_Bridge"
            
            # Map GitHub purpose to WHY dimension
            if "purpose" in github_context:
                context_vector.why = github_context.get("purpose") or "Protocol_Compliance"
            
            # Map GitHub process to HOW dimension
            if "process" in github_context:
                context_vector.how = github_context.get("process") or "Context_Translation"
            elif "method" in github_context:
                context_vector.how = github_context.get("method") or "Context_Translation"
            
            # Map GitHub scope to EXTENT dimension
            if "scope" in github_context:
                context_vector.extent = github_context.get("scope") or 1.0
            
            # Store the original GitHub context as metadata
            context_vector.metadata["original_github_context"] = json.dumps(github_context)
        
        # Add bridge metadata
        context_vector.metadata["bridge_timestamp"] = time.time()
        context_vector.metadata["mcp_version"] = self.current_version
        
        return context_vector

    # WHO: GitHubTNOSBridge.translateContext
    # WHAT: Translate context between systems
    # WHEN: During request processing
    # WHERE: MCP_Bridge
    # WHY: To ensure context compatibility
    # HOW: Using 7D context framework
    # EXTENT: All cross-system communications
    def translateContext(self, context: Dict[str, Any], direction: str = "to_tnos") -> Dict[str, Any]:
        """
        Translate context between GitHub MCP and TNOS MCP formats.
        
        Args:
            context: Context dictionary to translate
            direction: Direction of translation ("to_tnos" or "to_github")
            
        Returns:
            Translated context dictionary
        """
        if direction == "to_tnos":
            # Convert GitHub context to TNOS 7D context
            context_vector = self.bridgeMCPContext(context)
            
            # Apply compression to dimensional data for optimal transmission
            compressed_context = self.compressContextDimensions(context_vector)
            
            return compressed_context.to_dict()
        else:
            # Convert TNOS 7D context to GitHub context
            if isinstance(context, ContextVector7D):
                tnos_context = context
            else:
                tnos_context = ContextVector7D.from_dict(context)
            
            # Create GitHub context format
            github_context = {
                "identity": tnos_context.who,
                "operation": tnos_context.what,
                "timestamp": self.time_service.decompress_time(tnos_context.when),
                "location": tnos_context.where,
                "purpose": tnos_context.why,
                "method": tnos_context.how,
                "scope": tnos_context.extent,
                "metadata": {
                    "source": "tnos_mcp",
                    "translated_by": "github_mcp_bridge",
                    "timestamp": time.time(),
                    "mcp_version": self.current_version
                }
            }
            
            # Include any additional TNOS metadata
            if tnos_context.metadata:
                github_context["metadata"]["tnos_metadata"] = tnos_context.metadata
            
            return github_context

    # WHO: GitHubTNOSBridge.compressContextDimensions
    # WHAT: Apply compression to context
    # WHEN: Before context transmission
    # WHERE: MCP_Bridge
    # WHY: To optimize data transfer
    # HOW: Using Möbius compression formula
    # EXTENT: All dimensional context data
    def compressContextDimensions(self, context_vector: ContextVector7D) -> ContextVector7D:
        """
        Apply compression to 7D context dimensions using the Möbius compression formula.
        This optimizes transmission while preserving context integrity.
        
        Args:
            context_vector: The 7D context vector to compress
            
        Returns:
            Compressed 7D context vector
        """
        # Create a new context vector for the compressed output
        compressed = ContextVector7D(
            who=context_vector.who,
            what=context_vector.what,
            when=context_vector.when,
            where=context_vector.where,
            why=context_vector.why,
            how=context_vector.how,
            extent=context_vector.extent
        )
        
        # Copy metadata
        compressed.metadata = context_vector.metadata.copy()
        
        try:
            # Get compression handler
            compression_handler = get_mcp_handler()
            
            # Add indicators that compression was applied
            compressed.metadata["compression_applied"] = True
            compressed.metadata["compression_timestamp"] = time.time()
            
            # Compress temporal dimension (WHEN) if it's a timestamp
            if isinstance(context_vector.when, (int, float)):
                compressed.when = self.time_service.compress_time(context_vector.when)
                compressed.metadata["when_compression"] = "temporal"
            
            # Store compression variables for decompression
            compressed.metadata["compression_variables"] = {
                "B": 1.0,  # Base context integrity
                "I": 0.95,  # Information retention factor
                "V": 7.0,  # Vector dimensionality
                "G": 1.0,  # Global context scale
                "F": 0.05,  # Fine-tuning parameter
                "E": 0.01,  # Entropy scaling factor
            }
            
        except Exception as e:
            logger.warning(f"Context compression failed, using uncompressed: {e}")
            # If compression fails, return the original context
            return context_vector
        
        return compressed

    def translate_github_to_tnos_context(self, github_request: Dict[str, Any]) -> MCPContext:
        """
        Translate GitHub MCP request into TNOS MCPContext object.

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
        github_context = github_request.get("context", {})

        # Translate GitHub context to 7D context
        context_7d = self.bridgeMCPContext(github_context)
        
        # Update with tool-specific information if not already in GitHub context
        if "who" not in github_context:
            context.who = "GitHub_Copilot"
        else:
            context.who = context_7d.who
            
        if "what" not in github_context:
            context.what = self.tool_mapping.get(tool_name, "unknown_operation")
        else:
            context.what = context_7d.what
            
        if "when" not in github_context:
            context.when = self.time_service.get_current_time()
        else:
            context.when = context_7d.when
            
        if "where" not in github_context:
            # Determine location from parameters if available
            repo_info = f"{params.get('owner', '')}/{params.get('repo', '')}"
            if "path" in params:
                repo_info += f"/{params.get('path', '')}"
            context.where = repo_info if repo_info != "/" else "MCP_Bridge"
        else:
            context.where = context_7d.where
            
        if "why" not in github_context:
            context.why = f"GitHub_operation_{tool_name}"
        else:
            context.why = context_7d.why
            
        if "how" not in github_context:
            context.how = "github_mcp_bridge"
        else:
            context.how = context_7d.how
            
        if "extent" not in github_context:
            # Determine based on parameters if it's a single resource or multiple
            if any(param in params for param in ["perPage", "page", "list"]):
                context.extent = "multiple_resources"
            else:
                context.extent = "single_resource"
        else:
            context.extent = context_7d.extent

        # Add original request as metadata for reference
        context.add_metadata("original_github_request", json.dumps(github_request))
        context.add_metadata("github_tool", tool_name)
        context.add_metadata("mcp_version", self.current_version)
        
        # Add any metadata from the 7D context
        for key, value in context_7d.metadata.items():
            if key not in ["original_github_request", "github_tool", "mcp_version"]:
                context.add_metadata(key, value)

        return context

    def translate_tnos_to_github_response(
        self, tnos_response: Any, original_request: Dict[str, Any]
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
        try:
            if isinstance(tnos_response, MCPMessage):
                result_content = tnos_response.content
                context_obj = tnos_response.context
            else:
                result_content = tnos_response.get("content", {})
                context_obj = MCPContext()
                
                # Try to populate context from tnos_response
                if "context" in tnos_response and isinstance(tnos_response["context"], dict):
                    context_dict = tnos_response["context"]
                    for dim in ["who", "what", "when", "where", "why", "how", "extent"]:
                        if dim in context_dict:
                            setattr(context_obj, dim, context_dict[dim])
        except Exception as e:
            logger.error(f"Error extracting content from TNOS response: {e}")
            result_content = {"error": str(e)}
            context_obj = MCPContext()

        # If TNOS response is a string, try to parse as JSON
        if isinstance(result_content, str):
            try:
                result_content = json.loads(result_content)
            except json.JSONDecodeError:
                # If not valid JSON, wrap in a content field
                result_content = {"content": result_content}

        # Extract 7D context dimensions for metadata
        try:
            context_metadata = {
                "who": getattr(context_obj, "who", "System"),
                "what": getattr(context_obj, "what", "Response"),
                "when": getattr(context_obj, "when", self.time_service.get_current_time()),
                "where": getattr(context_obj, "where", "MCP_Bridge"),
                "why": getattr(context_obj, "why", "Request_Response"),
                "how": getattr(context_obj, "how", "Bridge_Translation"),
                "extent": getattr(context_obj, "extent", "Single_Response")
            }

            # If timestamp is compressed, decompress it
            if isinstance(context_metadata["when"], (int, float)):
                try:
                    context_metadata["when"] = self.time_service.decompress_time(context_metadata["when"])
                except:
                    # If decompression fails, use as-is
                    pass
                    
        except Exception as e:
            logger.error(f"Error extracting context metadata: {e}")
            context_metadata = {
                "who": "System",
                "what": "Response",
                "when": time.time(),
                "where": "MCP_Bridge",
                "why": "Request_Response",
                "how": "Bridge_Translation",
                "extent": "Single_Response"
            }

        # Add the original context information as metadata
        response = {
            "result": result_content,
            "metadata": {
                "tnos_context": context_metadata,
                "mcp_version": self.current_version,
                "timestamp": time.time(),
                "operation": original_request.get("name", ""),
                "request_id": original_request.get("id", str(uuid.uuid4()))
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
            # Extract host and port from URI
            uri_parts = self.tnos_server_uri.split("://")[1].split(":")
            host = uri_parts[0]
            port = int(uri_parts[1]) if len(uri_parts) > 1 else 9001

            # Check if port is available (this doesn't accurately check for WebSocket specifically)
            try:
                # Try to make a quick connection to see if the port is open
                test_socket = await asyncio.open_connection(host, port)
                await asyncio.sleep(0.1)  # Brief delay
                test_socket[0].close()
                await test_socket[1].wait_closed()
            except (ConnectionRefusedError, OSError):
                error_msg = f"TNOS MCP server not available at {host}:{port}"
                logger.error(error_msg)
                self.connection_state["tnos"]["last_error"] = error_msg
                return False

            # Connect to the server
            self.tnos_ws_connection = await websockets.connect(
                self.tnos_server_uri, 
                ping_interval=30,  # Send ping every 30 seconds
                ping_timeout=10,   # Wait 10 seconds for pong
                close_timeout=5    # Wait 5 seconds for clean close
            )

            # Update connection state
            self.connection_state["tnos"]["connected"] = True
            self.connection_state["tnos"]["last_connection_time"] = time.time()
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
            context_vector = ContextVector7D(
                who="GitHubTNOSBridge",
                what="ProtocolNegotiation",
                when=self.time_service.get_current_time(),
                where="Layer3_Bridge",
                why="VersionCompatibility",
                how="WebSocket",
                extent="ProtocolCompatibility"
            )
            
            # Convert to MCPContext
            context = MCPContext()
            context.who = context_vector.who
            context.what = context_vector.what
            context.when = context_vector.when
            context.where = context_vector.where
            context.why = context_vector.why
            context.how = context_vector.how
            context.extent = context_vector.extent

            # Create a version negotiation message
            negotiation_msg = {
                "type": "version_negotiation",
                "context": context_vector.to_dict(),
                "content": {
                    "supported_versions": self.supported_versions,
                    "preferred_version": self.current_version,
                    "minimum_version": self.min_version
                }
            }

            # Send the negotiation message as JSON
            await self.tnos_ws_connection.send(json.dumps(negotiation_msg))
            logger.info(f"Sent protocol version negotiation request with preferred version {self.current_version}")

            # Wait for a response with timeout
            response_str = await asyncio.wait_for(self.tnos_ws_connection.recv(), timeout=5.0)

            # Parse the response
            try:
                tnos_response = json.loads(response_str)
                
                # Extract the negotiated version
                if isinstance(tnos_response, dict):
                    if "content" in tnos_response and isinstance(tnos_response["content"], dict):
                        server_version = tnos_response["content"].get("selected_version")
                    else:
                        server_version = None
                else:
                    server_version = None
                
                if server_version:
                    # Update the current version to the negotiated version
                    self.current_version = server_version
                    logger.info(f"Negotiated MCP protocol version: {self.current_version}")
                else:
                    logger.warning("Server did not specify a protocol version, using default")
            except json.JSONDecodeError:
                logger.error(f"Invalid JSON in version negotiation response: {response_str[:100]}...")

        except asyncio.TimeoutError:
            logger.warning(f"Timeout during version negotiation, using default version {self.current_version}")
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
        self.connection_state["github"]["last_connection_time"] = time.time()

        try:
            # Ensure TNOS MCP connection is established
            if not self.tnos_ws_connection or not self.connection_state["tnos"]["connected"]:
                if not await self.connect_to_tnos_mcp():
                    await websocket.send(json.dumps({
                        "error": f"Failed to connect to TNOS MCP server: {self.connection_state['tnos']['last_error']}",
                        "status": "error",
                        "errorType": "connection_failure",
                        "recoverable": False
                    }))
                    return

            # Handle messages from GitHub MCP client
            async for message in websocket:
                try:
                    # Parse the message
                    github_request = json.loads(message)
                    
                    # Log the request type
                    logger.info(f"Received GitHub MCP request: {github_request.get('name', 'unknown')}")
                    
                    # Check if the tool is supported
                    tool_name = github_request.get("name", "")
                    if not tool_name.startswith("tnos_") and tool_name not in self.tool_mapping:
                        await websocket.send(json.dumps({
                            "error": f"Unsupported tool: {tool_name}",
                            "status": "error",
                            "errorType": "unsupported_tool",
                            "recoverable": True,
                            "suggestions": list(self.tool_mapping.keys())
                        }))
                        continue
                    
                    # Check rate limits
                    if not self._check_rate_limit(tool_name):
                        await websocket.send(json.dumps({
                            "error": f"Rate limit exceeded for tool: {tool_name}",
                            "status": "error",
                            "errorType": "rate_limit_exceeded",
                            "recoverable": True
                        }))
                        continue
                    
                    # Translate to TNOS context
                    tnos_context = self.translate_github_to_tnos_context(github_request)
                    
                    # Create TNOS MCP message
                    tnos_message = {
                        "type": "request",
                        "context": {
                            "WHO": tnos_context.who,
                            "WHAT": tnos_context.what,
                            "WHEN": tnos_context.when,
                            "WHERE": tnos_context.where,
                            "WHY": tnos_context.why,
                            "HOW": tnos_context.how,
                            "EXTENT": tnos_context.extent
                        },
                        "content": github_request.get("parameters", {})
                    }
                    
                    # Add metadata
                    for key, value in tnos_context.get_all_metadata().items():
                        tnos_message["context"][key] = value
                    
                    # Send to TNOS MCP
                    try:
                        await self.tnos_ws_connection.send(json.dumps(tnos_message))
                    except websockets.exceptions.ConnectionClosed:
                        # Connection lost, try to reconnect
                        logger.warning("Connection to TNOS MCP server lost, attempting to reconnect...")
                        if await self.connect_to_tnos_mcp():
                            await self.tnos_ws_connection.send(json.dumps(tnos_message))
                        else:
                            await websocket.send(json.dumps({
                                "error": "Connection to TNOS MCP server lost and reconnection failed",
                                "status": "error",
                                "errorType": "connection_failure",
                                "recoverable": False
                            }))
                            continue
                    
                    # Wait for response from TNOS MCP
                    try:
                        tnos_response_str = await asyncio.wait_for(self.tnos_ws_connection.recv(), timeout=30.0)
                        tnos_response = json.loads(tnos_response_str)
                        
                        # Translate response back to GitHub format
                        github_response = self.translate_tnos_to_github_response(tnos_response, github_request)
                        
                        # Send response back to GitHub client
                        await websocket.send(json.dumps(github_response))
                        logger.info(f"Sent response back to GitHub MCP client for {tool_name}")
                        
                    except asyncio.TimeoutError:
                        await websocket.send(json.dumps({
                            "error": "Timeout waiting for TNOS MCP server response",
                            "status": "error",
                            "errorType": "timeout",
                            "recoverable": True
                        }))
                    except Exception as e:
                        logger.error(f"Error receiving response from TNOS MCP: {e}")
                        await websocket.send(json.dumps({
                            "error": f"Error receiving response from TNOS MCP: {str(e)}",
                            "status": "error",
                            "errorType": "communication_failure",
                            "recoverable": True
                        }))
                
                except json.JSONDecodeError:
                    await websocket.send(json.dumps({
                        "error": "Invalid JSON in GitHub MCP request",
                        "status": "error",
                        "errorType": "invalid_json",
                        "recoverable": True
                    }))
                except Exception as e:
                    logger.error(f"Error handling GitHub MCP request: {e}")
                    await websocket.send(json.dumps({
                        "error": f"Error handling GitHub MCP request: {str(e)}",
                        "status": "error",
                        "errorType": "general_error",
                        "recoverable": True
                    }))
                
        except websockets.exceptions.ConnectionClosed:
            logger.info(f"GitHub MCP client disconnected: {client_id}")
        except Exception as e:
            logger.error(f"Unexpected error in handle_github_request: {e}")
        finally:
            # Clean up when the connection is closed
            if websocket in self.github_clients:
                self.github_clients.remove(websocket)
            
            # Close TNOS connection if no clients are connected
            if not self.github_clients:
                self.connection_state["github"]["connected"] = False
                if self.tnos_ws_connection and self.connection_state["tnos"]["connected"]:
                    await self.tnos_ws_connection.close()
                    self.connection_state["tnos"]["connected"] = False
                    logger.info("Closed connection to TNOS MCP server (no clients connected)")

    def _check_rate_limit(self, tool_name: str) -> bool:
        """
        Check if a request exceeds rate limits.
        
        Args:
            tool_name: Name of the tool being requested
            
        Returns:
            bool: True if within rate limits, False if limit exceeded
        """
        current_time = time.time()
        
        # Get rate limit settings for this tool
        if tool_name.startswith("tnos_compression"):
            limit_key = "compression"
        elif tool_name.startswith("tnos_formula") or tool_name.startswith("tnos_mobius_formula"):
            limit_key = "formula"
        else:
            limit_key = "default"
            
        limit = self.rate_limits.get(limit_key, self.rate_limits["default"])
        
        # Initialize or clean up request counts
        if tool_name not in self.request_counts:
            self.request_counts[tool_name] = []
        else:
            # Remove entries older than the period
            self.request_counts[tool_name] = [
                timestamp for timestamp in self.request_counts[tool_name] 
                if current_time - timestamp < limit["period"]
            ]
        
        # Check if we're exceeding the limit
        if len(self.request_counts[tool_name]) >= limit["requests"]:
            return False
        
        # Add current request to the count
        self.request_counts[tool_name].append(current_time)
        return True

    async def start(self):
        """
        Start the bridge server to handle connections from GitHub MCP.
        """
        try:
            # Check if the port might already be in use
            try:
                test_socket = await asyncio.open_connection("localhost", self.github_mcp_port)
                await asyncio.sleep(0.1)  # Brief delay
                test_socket[0].close()
                await test_socket[1].wait_closed()
                logger.error(f"Port {self.github_mcp_port} is already in use. " 
                            "Another instance of the bridge may be running.")
                return
            except (ConnectionRefusedError, OSError):
                # Port is available, which is what we want
                pass

            server = await websockets.serve(
                self.handle_github_request, "localhost", self.github_mcp_port
            )
            logger.info(f"GitHub-TNOS MCP Bridge running on port {self.github_mcp_port}")

            # Keep the server running
            await server.wait_closed()
        except Exception as e:
            logger.error(f"Failed to start GitHub-TNOS MCP Bridge: {e}")

    async def stop(self) -> bool:
        """
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

            # Remove PID file
            pid_path = os.path.join(project_root, "logs", "github_bridge.pid")
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


def register_tnos_tools_with_github_mcp():
    """
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
                },
                "required": ["formulaName"],
            },
        },
    ]

    return {
        "toolDefinitions": tnos_capabilities,
        "serverSettings": {"bridgeUri": f"ws://localhost:{args.github_port}"},
    }


async def main():
    """
    Main entry point for the GitHub-TNOS MCP Bridge.
    """
    # Create and start the bridge
    bridge = GitHubTNOSBridge(tnos_server_uri, args.github_port)

    # Log startup information
    logger.info("GitHub TNOS Bridge starting up")
    logger.info(f"Connecting to TNOS MCP server at {tnos_server_uri}")
    logger.info(f"GitHub MCP port: {args.github_port}")

    # Start the bridge
    await bridge.start()


if __name__ == "__main__":
    try:
        # Create and start the bridge
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("Bridge shutdown requested by user")
        sys.exit(0)
    except Exception as e:
        logger.error(f"Bridge startup failed: {e}", exc_info=True)
        sys.exit(1)
