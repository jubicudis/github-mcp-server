# pyright: reportMissingImports=false
#!/usr/bin/env -S /Users/Jubicudis/Tranquility-Neuro-OS/systems/python/venv311/bin/python3.11
# -*- coding: utf-8 -*-

import argparse
import asyncio
import hashlib
import json
import logging
import os
import secrets
import subprocess
import sys
import threading
import time
import traceback
from typing import Any, Dict, Optional, Union

"""
WHO: GitHubMCPBridgePython
WHAT: Python bridge between GitHub MCP and TNOS MCP
WHEN: During IDE runtime and system operations
WHERE: System Layer 3 (Higher Thought)
WHY: Enable communication between GitHub Copilot and TNOS with advanced
Python capabilities
HOW: WebSocket protocol with context translation and Python-specific
enhancements
EXTENT: All Python-based MCP communications
"""

# --- Ensure canonical TNOS MCP modules are importable at the very top ---
CANONICAL_MCP_PATH = "/Users/Jubicudis/Tranquility-Neuro-OS/mcp/bridge/python"
if CANONICAL_MCP_PATH not in sys.path:
    sys.path.insert(0, CANONICAL_MCP_PATH)
print(f"[DIAG][early] sys.path: {sys.path}")
print(f"[DIAG][early] os.listdir(CANONICAL_MCP_PATH): {os.listdir(CANONICAL_MCP_PATH)}")
print(f"[DIAG][early] sys.executable: {sys.executable}")
print(f"[DIAG][early] sys.path: {sys.path}")
print(f"[DIAG][early] PYTHONPATH: {os.environ.get('PYTHONPATH', None)}")

# --- Import websockets (required for async bridge) ---
try:
    import websockets
except ImportError as e:
    print("[ERROR] websockets module not found in venv311. Check that venv311 is active and installed.")
    print(e)
    print(f"[DIAG] Python executable: {sys.executable}")
    print(f"[DIAG] sys.path: {sys.path}")
    raise

# --- Canonical TNOS MCP module imports (must be after sys.path logic) ---
try:
    from tnos.mcp.context_translator import ContextTranslator
    from tnos.mcp.mobius_compression import MobiusCompression

    # Use ContextTranslator to resolve linting error
    _context_translator_instance = ContextTranslator()
except ImportError as e:
    print(f"[FATAL] Could not import TNOS MCP modules: {e}")
    print(f"[DIAG] sys.path: {sys.path}")
    print(f"[DIAG] os.listdir(CANONICAL_MCP_PATH): {os.listdir(CANONICAL_MCP_PATH)}")
    print(f"[DIAG] os.listdir(CANONICAL_MCP_PATH + '/tnos'): {os.listdir(os.path.join(CANONICAL_MCP_PATH, 'tnos'))}")
    exit(1)

# Use ContextTranslator to demonstrate import and provide a translation utility
context_translator = ContextTranslator()

def translate_context_with_class(context: dict, target_system: str = "tnos") -> dict:
    """
    Use the imported ContextTranslator class to translate context between systems.
    """
    return context_translator.translate(context, target_system)

# Helper: Always use venv311 python for subprocesses
VENV311_PATH = "/Users/Jubicudis/Tranquility-Neuro-OS/systems/python/venv311"
VENV311_BIN = os.path.join(VENV311_PATH, "bin")
VENV311_PYTHON = os.path.join(VENV311_BIN, "python3.11")

def run_in_venv311(args, **kwargs):
    if args[0] == sys.executable:
        args = [VENV311_PYTHON] + args[1:]
    elif args[0] in ("python", "python3", "python3.11"):
        args[0] = VENV311_PYTHON
    return subprocess.run(args, **kwargs)

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
        "port": 9001,
        "ws_endpoint": "ws://localhost:9001",
        "api_endpoint": "http://localhost:9001/api",
    },
    "bridge": {
        "port": 10619,
        "context_sync_interval": 60,
        "health_check_interval": 30,
        "reconnect_attempts": 5,
        "reconnect_delay": 5,
    },
    "visualization": {
        "port": 8083
    },
    "logging": {
        "log_dir": os.path.join(
            os.environ.get(
                "TNOS_ROOT",
                "/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS"
            ),
            "logs",
        ),
        "log_file": "mcp_bridge_python.log",
        "log_level": "info",
    },
    "paths": {
        "javascript_bridge": "/Users/Jubicudis/Tranquility-Neuro-OS/github-mcp-server/src/bridge/MCPBridge.js",
        "diagnostic_bridge": "/Users/Jubicudis/Tranquility-Neuro-OS/github-mcp-server/bridge/mcp_bridge.js",
    },
    "compression": {"enabled": True, "level": 7, "preserve_original": False},
}


# Configure logging
def setup_logging(level_name: str) -> logging.Logger:
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
    # Create formatter
    formatter = logging.Formatter(
        "[%(asctime)s] [%(levelname)s] [MCP-Python] %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )

    # Ensure log directory exists
    try:
        os.makedirs(CONFIG["logging"]["log_dir"], exist_ok=True)
    except Exception as e:
        print(f"[LOGGING][ERROR] Could not create log directory: {e}")

    # Create file handler, handle missing file gracefully
    log_path = os.path.join(
        CONFIG["logging"]["log_dir"],
        CONFIG["logging"]["log_file"]
    )
    try:
        file_handler = logging.FileHandler(log_path)
        file_handler.setLevel(level)
        file_handler.setFormatter(formatter)
        logger.addHandler(file_handler)
    except Exception as e:
        print(f"[LOGGING][ERROR] Could not create log file: {log_path} ({e})")

    # Create console handler
    console_handler = logging.StreamHandler()
    console_handler.setLevel(level)
    console_handler.setFormatter(formatter)
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
            self.github_queue.append({
                "message": message,
                "timestamp": time.time()
            })
            self._persist_queues()
        self.logger.info(
            f"Message queued for GitHub MCP "
            f"(queue size: {len(self.github_queue)})"
        )

    def queue_for_tnos(self, message: Dict[str, Any]) -> None:
        """Add message to TNOS MCP queue"""
        with self.lock:
            self.tnos_queue.append({
                "message": message,
                "timestamp": time.time()
            })
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
                        f"Loaded message queues: "
                        f"GitHub ({len(self.github_queue)}), "
                        f"TNOS ({len(self.tnos_queue)})"
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
                f"Processing {len(self.github_queue)} "
                f"queued messages for GitHub MCP"
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
            if websocket and not getattr(websocket, "closed", False):
                # Take a copy of the queue and clear it before processing
                messages_to_process = self.github_queue.copy()
                self.github_queue = []
                self._persist_queues()

                for i, item in enumerate(messages_to_process):
                    try:
                        await websocket.send(json.dumps(item["message"]))
                        self.logger.debug(
                            (f"Sent queued message {i + 1}/"
                             f"{len(messages_to_process)} to GitHub MCP")
                        )
                    except Exception as e:
                        self.logger.error(
                            f"Error sending queued message to GitHub MCP: "
                            f"{str(e)}"
                        )
                        # Re-queue the failed message
                        self.queue_for_github(item["message"])

                self.logger.info(
                    "Finished processing GitHub MCP message queue"
                )

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
            if websocket and not getattr(websocket, "closed", False):
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


# 7D Context Vector class
class ContextVector7D:
    """
    WHO: ContextVector7D
    WHAT: 7D Context vector implementation
    WHEN: During protocol operations
    WHERE: Layer 3 / Python
    WHY: To provide structured context for all operations
    HOW: Using the 7D context framework
    EXTENT: All MCP communication
    """
    
    def __init__(self, 
                who: str = "System", 
                what: str = "Operation", 
                when: Optional[float] = None, 
                where: str = "MCP_Bridge", 
                why: str = "SystemOperation", 
                how: str = "Default", 
                extent: Union[float, str] = 1.0,
                metadata: Optional[Dict[str, Any]] = None):
        """
        Initialize a new 7D context vector.
        
        Args:
            who: Actor identity
            what: Function or content
            when: Timestamp or temporal marker (defaults to current time)
            where: System location
            why: Intent or purpose
            how: Method or process
            extent: Scope or impact (1.0 = full scope)
            metadata: Additional contextual information
        """
        self.who = who
        self.what = what
        self.when = when if when is not None else time.time()
        self.where = where
        self.why = why
        self.how = how
        self.extent = extent
        self.metadata = metadata or {}
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert context vector to dictionary representation."""
        return {
            "who": self.who,
            "what": self.what,
            "when": self.when,
            "where": self.where,
            "why": self.why,
            "how": self.how,
            "extent": self.extent,
            "metadata": self.metadata
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'ContextVector7D':
        """Create context vector from dictionary representation."""
        return cls(
            who=data.get("who", "System"),
            what=data.get("what", "Operation"),
            when=data.get("when", time.time()),
            where=data.get("where", "MCP_Bridge"),
            why=data.get("why", "Protocol_Compliance"),
            how=data.get("how", "Default"),
            extent=data.get("extent", 1.0),
            metadata=data.get("metadata", {})
        )
    
    def derive(self, **kwargs) -> 'ContextVector7D':
        """Create a derived context with specified overrides."""
        data = self.to_dict()
        data.update(kwargs)
        return ContextVector7D.from_dict(data)


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
            "when": normalize_timestamp(message.get("timestamp", int(time.time() * 1000))),
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
            "timestamp": normalize_timestamp(message.get("timestamp", int(time.time() * 1000))),
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
            "timestamp": normalize_timestamp(message.get("timestamp", int(time.time() * 1000))),
            "id": message.get("id", f"bridge-{int(time.time() * 1000)}"),
        }

        return github_message


def bridgeMCPContext(githubContext: Dict[str, Any], tnosContext: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
    """
    WHO: ContextBridge
    WHAT: Bridge between GitHub MCP and TNOS MCP contexts
    WHEN: During protocol translations
    WHERE: Layer 3 / Python
    WHY: To ensure context preservation across systems
    HOW: By mapping between context formats
    EXTENT: All cross-system communications
    
    Bridge function to convert between GitHub MCP context format and TNOS 7D context.
    
    Args:
        githubContext: The GitHub context to convert
        tnosContext: Optional existing TNOS context to merge with
        
    Returns:
        A TNOS 7D context dictionary
    """
    # Convert external MCP format to internal 7D context
    contextVector = {
        "who": githubContext.get("identity") or githubContext.get("user") or "System",
        "what": githubContext.get("operation") or githubContext.get("type") or "Transform",
        "when": githubContext.get("timestamp") or time.time(),
        "where": "MCP_Bridge",
        "why": githubContext.get("purpose") or "Protocol_Compliance",
        "how": "Context_Translation",
        "extent": githubContext.get("scope") or 1.0,
        "metadata": {
            "original_github_context": githubContext,
            "bridge_timestamp": time.time(),
            "protocol_version": CONFIG.get("protocol_version", "3.0")
        }
    }
    
    # If we have an existing TNOS context, merge with it
    if tnosContext:
        for key, value in tnosContext.items():
            if key not in contextVector or not contextVector[key]:
                contextVector[key] = value
        
        # Merge metadata
        if "metadata" in tnosContext:
            contextVector["metadata"].update(tnosContext["metadata"])
    
    return contextVector


def translateContext(context: Dict[str, Any], targetSystem: str = "tnos") -> Dict[str, Any]:
    """
    WHO: ContextTranslator
    WHAT: Context translation function
    WHEN: During cross-system communication
    WHERE: MCP_Bridge
    WHY: To ensure proper context mapping
    HOW: Using system-specific mapping rules
    EXTENT: All context translations
    
    Translate context format between systems.
    
    Args:
        context: The context to translate
        targetSystem: The target system format ("tnos" or "github")
        
    Returns:
        Translated context dictionary
    """
    if not context:
        return {}
        
    if targetSystem.lower() == "tnos":
        # GitHub to TNOS translation
        if "user" in context or "identity" in context:
            return bridgeMCPContext(context)
        return context  # Already in TNOS format
    
    elif targetSystem.lower() == "github":
        # TNOS to GitHub translation
        if "who" in context:  # It's a TNOS 7D context
            return {
                "user": context.get("who", "System"),
                "type": context.get("what", "Transform"),
                "timestamp": context.get("when", time.time()),
                "purpose": context.get("why", "Protocol_Compliance"),
                "scope": context.get("extent", 1.0),
                "metadata": {
                    "tnos_context": context,
                    "translation_timestamp": time.time()
                }
            }
    
    # Unknown target system, return original
    return context


# Utility and fallback functions for MCP bridge

def normalize_timestamp(ts):
    """Convert ms timestamps to seconds if needed (robust for int/float)."""
    if isinstance(ts, (int, float)) and ts > 1e10:
        return ts // 1000 if isinstance(ts, int) else ts / 1000.0
    return ts

def get_compression_stats():
    if MobiusCompression and hasattr(MobiusCompression, 'get_statistics'):
        return MobiusCompression.get_statistics()
    return {"error": "MobiusCompression unavailable"}

# --- Visualization Policy ---
# Visualization features are now provided as an internal component of the TNOS MCP server.
# Do not attempt to start or connect to a separate visualization server process.
# TODO: Integrate all visualization endpoints and logic into the TNOS MCP server codebase.
#       The bridge and GitHub MCP server should treat visualization as a TNOS MCP feature only.
#
# TODO: Future upgrades (CrewAI, ATL, ATM, formula registry, etc.) should be added as internal
#       components of the TNOS MCP server or bridge, not as separate servers or scripts.

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

    def __init__(self, logger: logging.Logger, port: Optional[int] = None):
        self.logger = logger
        self.port = port or CONFIG["bridge"]["port"]
        self.github_mcp_socket = None
        self.tnos_mcp_socket = None
        self.client_sockets = set()
        self.message_queue = MessageQueue(logger)
        self.context_bridge = ContextBridge()
        self.shutting_down = False
        self.start_time = time.time()

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
            self.logger.info("MCP servers are not running. Please start all MCP components using setup_dev_env.sh (start-all). Auto-start via legacy shell script is now disabled.")
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
        """Connect to TNOS MCP server with QHP handshake"""
        self.logger.info("Connecting to TNOS MCP server...")
        if self.tnos_mcp_socket:
            try:
                await self.tnos_mcp_socket.close()
            except Exception:
                pass
            self.tnos_mcp_socket = None
        try:
            ws_endpoint = CONFIG["tnos_mcp"]["ws_endpoint"]
            self.logger.debug(f"Attempting to connect to TNOS MCP at {ws_endpoint}")
            self.tnos_mcp_socket = await websockets.connect(ws_endpoint)
            self.logger.info(f"Connected to TNOS MCP server on port {CONFIG['tnos_mcp']['port']}")
            # --- QHP handshake ---
            node_name = "GITHUB_MCP_BRIDGE"
            my_fingerprint = MCPBridgeServer.generate_quantum_fingerprint(node_name)
            my_challenge = secrets.token_hex(16)
            handshake = {
                "type": "qhp_handshake",
                "fingerprint": my_fingerprint,
                "challenge": my_challenge,
                "supported_versions": ["3.0"],
            }
            await self.tnos_mcp_socket.send(json.dumps(handshake))
            handshake_response_msg = await asyncio.wait_for(self.tnos_mcp_socket.recv(), timeout=5)
            handshake_response = json.loads(handshake_response_msg)
            if handshake_response.get("type") != "qhp_handshake_response":
                self.logger.error("QHP handshake response not received, closing connection.")
                await self.tnos_mcp_socket.close()
                self.tnos_mcp_socket = None
                return
            peer_fingerprint = handshake_response.get("fingerprint")
            peer_challenge = handshake_response.get("challenge")
            challenge_response = handshake_response.get("challenge_response")
            # Verify their response to our challenge
            if not MCPBridgeServer.verify_challenge_response(my_challenge, challenge_response, peer_fingerprint):
                self.logger.error("QHP handshake: challenge response invalid, closing connection.")
                await self.tnos_mcp_socket.close()
                self.tnos_mcp_socket = None
                return
            # Respond to their challenge
            my_response = hashlib.sha256((peer_challenge + my_fingerprint).encode("utf-8")).hexdigest()
            ack = {
                "type": "qhp_handshake_ack",
                "challenge_response": my_response,
            }
            await self.tnos_mcp_socket.send(json.dumps(ack))
            self.logger.info(f"QHP handshake complete with TNOS MCP peer {peer_fingerprint}")
            # --- End QHP handshake, proceed to normal ops ---
            self.reset_backoff("tnos")
            await self.message_queue.process_tnos_queue(self.tnos_mcp_socket)
            # After handshake, demonstrate RIC request for quantum symmetry
            ric_context = {
                "I": 1.0, "H": 0.5, "V": 1.0, "gamma": 0.1, "C": 0.5, "rho": 0.01, "t": 1.0,
                "kappa": 0.1, "eta": 1.0, "alpha": 1.0, "beta": 0.5, "sigma": 0.2, "tau": 1.0, "Lambda1": 1.0
            }
            ric_score = await self.request_ric_from_tnos(ric_context)
            if ric_score is not None:
                self.logger.info(f"Quantum handshake: RIC score from TNOS MCP: {ric_score}")
            asyncio.create_task(self.handle_tnos_mcp_messages())
        except (
            websockets.exceptions.WebSocketException,
            ConnectionRefusedError,
            OSError,
        ) as e:
            self.logger.warning(
                f"TNOS MCP WebSocket error on port {CONFIG['tnos_mcp']['port']}: {str(e)}"
            )
            next_delay = self.get_next_backoff_delay("tnos")
            self.logger.info(
                f"Will attempt to reconnect to TNOS MCP server in {next_delay}s "
                f"(attempt {self.backoff_strategies['tnos']['attempts']})"
            )
            if not self.shutting_down:
                asyncio.create_task(self.delayed_reconnect("tnos", next_delay))

    async def request_ric_from_tnos(self, context: dict) -> Optional[float]:
        """Request RIC score from TNOS MCP server for a given 7D context."""
        if not self.tnos_mcp_socket or getattr(self.tnos_mcp_socket, "closed", False):
            self.logger.warning("TNOS MCP socket not connected for RIC request.")
            return None
        try:
            request = {
                "operation": "compute_ric",
                "context": context,
            }
            await self.tnos_mcp_socket.send(json.dumps(request))
            response_msg = await asyncio.wait_for(self.tnos_mcp_socket.recv(), timeout=5)
            response = json.loads(response_msg)
            if response.get("status") == "ok" and "ric" in response:
                return response["ric"]
            else:
                self.logger.warning(f"RIC request failed: {response}")
                return None
        except Exception as e:
            self.logger.error(f"RIC request error: {e}")
            return None

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
            self.logger.warning(f"GitHub MCP WebSocket connection closed: {str(e)}")

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
            self.logger.warning(f"TNOS MCP WebSocket connection closed: {str(e)}")

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
            # Defensive: normalize any timestamps in context
            if "timestamp" in message:
                message["timestamp"] = normalize_timestamp(message["timestamp"])
            if "context" in message and isinstance(message["context"], dict):
                ctx = message["context"]
                if "when" in ctx:
                    ctx["when"] = normalize_timestamp(ctx["when"])

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
            # Defensive: normalize any timestamps in context
            if "timestamp" in message:
                message["timestamp"] = normalize_timestamp(message["timestamp"])
            if "context" in message and isinstance(message["context"], dict):
                ctx = message["context"]
                if "when" in ctx:
                    ctx["when"] = normalize_timestamp(ctx["when"])

        # Transform TNOS 7D message to GitHub MCP format
        github_message = self.context_bridge.tnos7d_to_github(message)

        # Forward to GitHub MCP
        if self.github_mcp_socket and not getattr(self.github_mcp_socket, "closed", False):
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
            self.logger.warning(
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
        if self.github_mcp_socket and not getattr(self.github_mcp_socket, "closed", False):
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
        if self.github_mcp_socket and not getattr(self.github_mcp_socket, "closed", False):
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
        if self.github_mcp_socket and not getattr(self.github_mcp_socket, "closed", False):
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
        if self.tnos_mcp_socket and not getattr(self.tnos_mcp_socket, "closed", False):
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
            self.logger.warning(
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
                if not getattr(client, "closed", False)
            ],
            return_exceptions=True,
        )

    async def sync_context(self) -> None:
        """Synchronize context between MCP servers"""
        self.logger.info("Syncing context between MCP servers...")

        # In a real implementation, this would load and sync actual contexts
        # For now, we'll just log the synchronization

        if self.tnos_mcp_socket and not getattr(self.tnos_mcp_socket, "closed", False):
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

        if self.github_mcp_socket and not getattr(self.github_mcp_socket, "closed", False):
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
            self.logger.warning("GitHub MCP server is not responding")
            await self.connect_to_github_mcp()

        if not tnos_mcp_running:
            self.logger.warning("TNOS MCP server is not responding")
            await self.connect_to_tnos_mcp()

        if not github_mcp_running or not tnos_mcp_running:
            self.logger.info("Attempting to restart MCP servers...")
            await self.ensure_servers_running()

    async def ensure_bridge_files(self) -> bool:
        self.logger.info("Validating bridge file paths...")
        try:
            js_bridge = CONFIG["paths"]["javascript_bridge"]
            diag_bridge = CONFIG["paths"]["diagnostic_bridge"]
            missing = []
            if not os.path.exists(js_bridge):
                missing.append(js_bridge)
            if not os.path.exists(diag_bridge):
                missing.append(diag_bridge)
            if missing:
                self.logger.warning(f"Missing bridge files: {missing}")
                return False
            self.logger.info("All bridge files validated.")
            return True
        except Exception as exc:
            self.logger.error(f"Error validating bridge files: {exc}")
            return False

    async def context_sync_loop(self) -> None:
        interval = CONFIG["bridge"].get("context_sync_interval", 60)
        while not self.shutting_down:
            try:
                await self.sync_context()
            except Exception as e:
                self.logger.error(f"Context sync error: {e}")
            await asyncio.sleep(interval)

    async def health_check_loop(self) -> None:
        interval = CONFIG["bridge"].get("health_check_interval", 30)
        while not self.shutting_down:
            try:
                await self.health_check()
            except Exception as e:
                self.logger.error(f"Health check error: {e}")
            await asyncio.sleep(interval)

    async def start_websocket_server(self) -> None:
        self.logger.info(f"Starting bridge WebSocket server on port {self.port}...")
        async with websockets.serve(self.handle_client, "0.0.0.0", self.port):
            self.logger.info(f"Bridge WebSocket server running on port {self.port}")
            while not self.shutting_down:
                await asyncio.sleep(1)

    async def handle_client(self, websocket):
        """Handle incoming client WebSocket connections."""
        self.client_sockets.add(websocket)
        self.logger.info(f"Client connected: {getattr(websocket, 'remote_address', None)}")
        try:
            idle_count = 0
            while True:
                try:
                    # Wait for a message with timeout (idle detection)
                    message = await asyncio.wait_for(websocket.recv(), timeout=5)
                    idle_count = 0  # Reset idle counter on message
                    data = json.loads(message)
                    self.logger.info(f"Received message from client: {data}")
                    # Symmetric routing: classify and forward
                    target = None
                    if data.get("target") == "github":
                        target = "github"
                    elif data.get("target") == "tnos":
                        target = "tnos"
                    elif data.get("type") in ("github", "copilot", "mcp"):
                        target = "github"
                    elif data.get("type") in ("tnos", "7d", "context"):
                        target = "tnos"
                    else:
                        target = "both"
                    if target in ("github", "both") and self.github_mcp_socket and not getattr(self.github_mcp_socket, "closed", False):
                        await self.github_mcp_socket.send(json.dumps(data))
                        self.logger.info("Forwarded client message to GitHub MCP")
                    if target in ("tnos", "both") and self.tnos_mcp_socket and not getattr(self.tnos_mcp_socket, "closed", False):
                        await self.tnos_mcp_socket.send(json.dumps(data))
                        self.logger.info("Forwarded client message to TNOS MCP")
                    await websocket.send(json.dumps({"status": "forwarded", "target": target}))
                except asyncio.TimeoutError:
                    idle_count += 1
                    if idle_count == 1:
                        # First timeout: send ping/prompt, do not close
                        self.logger.info("Client connection idle, sending ping.")
                        try:
                            await websocket.send(json.dumps({"type": "ping", "message": "Are you still there?"}))
                        except Exception as e:
                            self.logger.warning(f"Failed to send ping to client: {e}")
                        continue
                    else:
                        self.logger.warning("Client connection idle for two intervals, closing cleanly.")
                        await websocket.close()
                        break
                except Exception as e:
                    self.logger.error(f"Error processing client message: {e}")
                    await websocket.send(json.dumps({"status": "error", "error": str(e)}))
        except Exception as e:
            self.logger.error(f"Error in client handler: {e}")
        finally:
            self.client_sockets.discard(websocket)
            self.logger.info(f"Client disconnected: {getattr(websocket, 'remote_address', None)}")

    async def start(self):
        """Start the MCP bridge server and background tasks."""
        self.logger.info("Starting MCPBridgeServer...")
        self.shutting_down = False
        # Start background tasks
        self._tasks = []
        self._tasks.append(asyncio.create_task(self.context_sync_loop()))
        self._tasks.append(asyncio.create_task(self.health_check_loop()))
        # Start websocket server (blocks until shutdown)
        await self.start_websocket_server()

    def shutdown(self):
        """Shutdown the MCP bridge server and cleanup."""
        self.logger.info("Shutting down MCPBridgeServer...")
        self.shutting_down = True
        # Cancel background tasks
        if hasattr(self, '_tasks'):
            for task in self._tasks:
                task.cancel()
        # Close all client sockets
        for ws in list(self.client_sockets):
            try:
                asyncio.create_task(ws.close())
            except Exception:
                pass
        # Close connections to MCP servers
        if self.github_mcp_socket:
            try:
                asyncio.create_task(self.github_mcp_socket.close())
            except Exception:
                pass
        if self.tnos_mcp_socket:
            try:
                asyncio.create_task(self.tnos_mcp_socket.close())
            except Exception:
                pass
        self.logger.info("MCPBridgeServer shutdown complete.")

    # --- QHP helpers ---
    @staticmethod
    def generate_quantum_fingerprint(node_name: str) -> str:
        entropy = secrets.token_bytes(16)
        h = hashlib.sha256()
        h.update(node_name.encode("utf-8"))
        h.update(entropy)
        return h.hexdigest()
    @staticmethod
    def verify_challenge_response(challenge: str, response: str, peer_fingerprint: str) -> bool:
        h = hashlib.sha256()
        h.update((challenge + peer_fingerprint).encode("utf-8"))
        return h.hexdigest() == response


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
    parser = argparse.ArgumentParser(description="GitHub MCP Bridge for TNOS")
    parser.add_argument(
        "--port", type=int, default=CONFIG["bridge"]["port"], help="Port to run the bridge on (default: 10619)"
    )
    parser.add_argument(
        "--log-level", type=str, choices=["debug", "info", "warning", "error"], default=CONFIG["logging"]["log_level"], help="Log level (default: info)"
    )
    args = parser.parse_args()

    logger = setup_logging(args.log_level)
    logger.info(f"Launching GitHub MCP Bridge on port {args.port} with log level {args.log_level}")

    server = MCPBridgeServer(logger, port=args.port)

    loop = asyncio.get_event_loop()
    try:
        loop.run_until_complete(server.start())
    except KeyboardInterrupt:
        logger.info("KeyboardInterrupt received. Shutting down...")
        server.shutdown()
    except Exception as exc:
        logger.error(f"Fatal error in bridge: {exc}")
        logger.debug(traceback.format_exc())
        server.shutdown()
    finally:
        logger.info("Bridge server exiting.")
        loop.close()

if __name__ == "__main__":
    main()
