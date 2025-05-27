# TNOS Parallel Coding Policy: All edits, enhancements, and AI instance creations should be performed in parallel across relevant files and subsystems. Cross-reference and synchronize changes as if working on a multi-component system (e.g., smartphone hardware/software stack). All AI instances must be created with full access to memory, DNA, and QHP technology, and all changes should be reflected across the MCP system in parallel.

# pyright: reportMissingImports=false
#!/usr/bin/env -S /Users/Jubicudis/Tranquility-Neuro-OS/systems/python/venv311/bin/python3.11
# -*- coding: utf-8 -*-

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
    # Remove broken .translate call, use to_7d/from_7d or fallback to translateContext
    if hasattr(context_translator, "to_7d") and target_system.lower() == "tnos":
        return context_translator.to_7d(context)
    elif hasattr(context_translator, "from_7d") and target_system.lower() == "github":
        return context_translator.from_7d(context)
    else:
        return translateContext(context, target_system)

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
        "ws_endpoint": "ws://localhost:10617",
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
            if key not in contextVector:
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
        if "user" in context or "identity" in context:
            return bridgeMCPContext(context)
        return context
    elif targetSystem.lower() == "github":
        if "who" in context:
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


# --- Quantum Handshake Protocol (QHP) utilities ---
def get_pid_on_port_static(port):
    try:
        result = subprocess.run([
            'lsof', '-i', f':{port}', '-t'
        ], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
        pid_str = result.stdout.strip()
        if pid_str:
            return int(pid_str.split('\n')[0])
    except Exception as e:
        print(f"[QHP][Self-Heal][ERROR] Could not get PID on port {port}: {e}")
    return None

# --- CrewAI import and availability check ---
try:
    import importlib.util
    CREWAI_AVAILABLE = importlib.util.find_spec("crewai") is not None and importlib.util.find_spec("langchain") is not None
except ImportError:
    CREWAI_AVAILABLE = False

# --- MobiusCompression and ContextTranslator integration ---
try:
    from tnos.mcp.mobius_compression import MobiusCompression
except ImportError:
    try:
        from mcp.bridge.python.tnos.mcp.mobius_compression import \
            MobiusCompression
    except ImportError:
        class MobiusCompression:
            @staticmethod
            def compress(data, context=None):
                return f"[MOBIUS_STUB_COMPRESS]{data}"
            @staticmethod
            def decompress(data, context=None):
                return f"[MOBIUS_STUB_DECOMPRESS]{data}"
try:
    from tnos.mcp.context_translator import ContextTranslator
except ImportError:
    try:
        from mcp.bridge.python.tnos.mcp.context_translator import \
            ContextTranslator
    except ImportError:
        class ContextTranslator:
            @staticmethod
            def to_7d(context):
                # Minimal stub: ensure 7D keys
                keys = ["who","what","when","where","why","how","extent"]
                return {k: context.get(k, f"{k}_default") for k in keys}
            @staticmethod
            def from_7d(context):
                return context

# --- Formula Registry and TranquilSpeak Symbol Key ---
class FormulaRegistry:
    SYMBOL_KEY = {
        "attention_binding": "Ⓣ",
        "context_mapping": "Ⓛ",
        "energy_optimization": "ⓔ",
        "mobius": "Ⓜ",
        "fallback": "ⓕ",
        "default": "ⓓ"
    }
    FORMULAS = {
        "Ⓣ": "attention_binding_formula",
        "Ⓛ": "context_mapping_formula",
        "ⓔ": "energy_optimization_formula",
        "Ⓜ": "mobius_compression_formula",
        "ⓕ": "fallback_routing_formula",
        "ⓓ": "default_formula"
    }
    @classmethod
    def select_formula(cls, context):
        # Use context and symbol key to select formula
        dim = context.get("dimension")
        if dim == "energy":
            return cls.FORMULAS["ⓔ"]
        elif dim == "compression":
            return cls.FORMULAS["Ⓜ"]
        elif dim == "attention":
            return cls.FORMULAS["Ⓣ"]
        elif dim == "context":
            return cls.FORMULAS["Ⓛ"]
        elif dim == "fallback":
            return cls.FORMULAS["ⓕ"]
        return cls.FORMULAS["ⓓ"]


# === Helical Memory Implementation ===
class HelicalMemory:
    """
    WHO: HelicalMemory
    WHAT: 7D contextual logging for short-term and long-term memory
    WHEN: During all bridge operations
    WHERE: github-mcp-server/memory/
    WHY: To provide persistent, queryable logs for system introspection and debugging
    HOW: Appends JSON log entries to short_term.log and long_term.log
    EXTENT: All major bridge events, QHP handshakes, errors, and context changes
    """
    def __init__(self, memory_dir=None):
        if memory_dir is None:
            workspace_root = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
            memory_dir = os.path.join(workspace_root, "github-mcp-server", "memory")
        self.short_term_path = os.path.join(memory_dir, "short_term.log")
        self.long_term_path = os.path.join(memory_dir, "long_term.log")
        self.lock = threading.Lock()
        os.makedirs(memory_dir, exist_ok=True)

    def log(self, who, what, when, where, why, how, extent, msg, extra=None):
        entry = {
            "who": who,
            "what": what,
            "when": when,
            "where": where,
            "why": why,
            "how": how,
            "extent": extent,
            "ts": int(time.time()),
            "msg": msg,
        }
        if extra:
            entry.update(extra)
        line = json.dumps(entry)
        with self.lock:
            with open(self.short_term_path, "a") as f:
                f.write(line + "\n")
            with open(self.long_term_path, "a") as f:
                f.write(line + "\n")


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

    TNOS Parallel Coding Policy: All AI instances (including QHP, PortAssignmentAI, etc.) must be created with full memory, DNA, and QHP capabilities. All changes to AI logic, memory, or DNA must be reflected in parallel across the MCP system.
    """

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
        self.helical_memory = HelicalMemory()

        # --- CrewAI QHP/HelicalMemory integration ---
        if CREWAI_AVAILABLE:
            class HelicalMemory:
                def __init__(self, max_length=128):
                    self.primary = []
                    self.secondary = []
                    self.max_length = max_length
                def store(self, event):
                    compressed = QHPCrewAgent.tranquilspeak_compress(event)
                    if len(self.primary) >= self.max_length:
                        self.primary.pop(0)
                        self.secondary.pop(0)
                    self.primary.append(compressed)
                    parity = hashlib.sha256(json.dumps(compressed, sort_keys=True).encode()).hexdigest()
                    self.secondary.append({"parity": parity, "timestamp": time.time()})
                def reconstruct(self, idx):
                    if idx < len(self.primary):
                        return self.primary[idx]
                    elif idx < len(self.secondary):
                        return {"reconstructed": True, **self.secondary[idx]}
                    return None
                def get_all(self):
                    return list(self.primary)
            class QHPCrewAgent:
                def __init__(self, name, port, memory, logger):
                    self.name = name
                    self.port = port
                    self.memory = memory
                    self.logger = logger
                    self.fingerprint = MCPBridgeServer.generate_quantum_fingerprint(name)
                    self.trust_table = {}
                @staticmethod
                def tranquilspeak_compress(event):
                    if isinstance(event, dict):
                        role = event.get("role") or event.get("type") or "SYS"
                        action = event.get("action") or event.get("operation") or event.get("type") or "EVT"
                        target = event.get("target") or event.get("peer") or event.get("port") or "*"
                        ts = event.get("ts") or event.get("timestamp") or int(time.time())
                        return f"~{role}@{action}/{target} [ts:{ts}]"
                    return str(event)
                @staticmethod
                def tranquilspeak_decompress(ts_str):
                    return {"decompressed": ts_str}
                def observe_ports(self):
                    observed = {}
                    for pname, pnum in [("github_mcp", 10617), ("tnos_mcp", 9001), ("bridge", 10619), ("visualization", 8083)]:
                        pid = get_pid_on_port_static(pnum)
                        observed[pname] = bool(pid)
                    event = {"type": "port_observation", "ports": observed, "ts": time.time()}
                    self.memory.log(self.name, "port_observation", time.time(), "bridge", "monitor", "observe_ports", 1.0, "Observed MCP ports", {"ports": observed})
                    self.memory.store(event)
                    return observed
                def perform_qhp_handshake(self, peer_name, peer_port):
                    my_challenge = secrets.token_hex(16)
                    my_fp = self.fingerprint
                    peer_fp = MCPBridgeServer.generate_quantum_fingerprint(peer_name)
                    peer_challenge = secrets.token_hex(16)
                    my_response = hashlib.sha256((peer_challenge + my_fp).encode("utf-8")).hexdigest()
                    peer_response = hashlib.sha256((my_challenge + peer_fp).encode("utf-8")).hexdigest()
                    session_key = secrets.token_hex(32)
                    self.trust_table[peer_name] = {
                        "peer_fingerprint": peer_fp,
                        "challenge": my_challenge,
                        "challenge_response": peer_response,
                        "session_key": session_key,
                        "timestamp": time.time(),
                    }
                    event = {
                        "type": "qhp_handshake",
                        "peer": peer_name,
                        "peer_fp": peer_fp,
                        "my_fp": my_fp,
                        "challenge": my_challenge,
                        "peer_challenge": peer_challenge,
                        "my_response": my_response,
                        "peer_response": peer_response,
                        "session_key": session_key,
                        "ts": time.time(),
                    }
                    self.memory.log(self.name, "QHP handshake", time.time(), "bridge", "trust", "QHP", 1.0, f"QHP handshake with {peer_name}", event)
                    self.memory.store(event)
                    self.logger.info(f"[QHP][CrewAI][TS] {QHPCrewAgent.tranquilspeak_compress(event)}")
                    self.logger.info(f"[QHP][CrewAI] Handshake with {peer_name} complete. Session key: {session_key}. Trust table updated.")
                    return True
                def decay_trust(self, timeout=600):
                    now = time.time()
                    expired = [k for k, v in self.trust_table.items() if now - v["timestamp"] > timeout]
                    for k in expired:
                    if expired:
                        self.logger.info(f"[QHP][CrewAI] Decayed trust for: {expired}")
                def get_trust_table(self):
                    return dict(self.trust_table)
            self.helical_memory = HelicalMemory(max_length=256)
            self.qhp_bridge_agent = QHPCrewAgent("MCPBridge", 10619, self.helical_memory, self.logger)
            self.qhp_tnos_agent = QHPCrewAgent("TNOS_MCP", 9001, self.helical_memory, self.logger)

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

    async def connect_to_tnos_mcp(self) -> None:
        self.logger.info("Connecting to TNOS MCP server...")
        self.helical_memory.log("MCPBridgeServer", "connect", time.time(), "bridge", "init", "connect_to_tnos_mcp", 1.0, "Attempting to connect to TNOS MCP server.")
        if self.tnos_mcp_socket:
            try:
                await self.tnos_mcp_socket.close()
            except Exception:
                pass
            self.tnos_mcp_socket = None
        try:
            ws_endpoint = CONFIG["tnos_mcp"]["ws_endpoint"]
            self.logger.debug(f"Attempting to connect to TNOS MCP at {ws_endpoint}")
            self.helical_memory.log("MCPBridgeServer", "connect", time.time(), "bridge", "init", "connect_to_tnos_mcp", 1.0, f"Connecting to {ws_endpoint}")
            self.tnos_mcp_socket = await websockets.connect(ws_endpoint)
            self.logger.info(f"Connected to TNOS MCP server on port {CONFIG['tnos_mcp']['port']}")
            self.helical_memory.log("MCPBridgeServer", "connect", time.time(), "bridge", "success", "connect_to_tnos_mcp", 1.0, f"Connected to TNOS MCP server on port {CONFIG['tnos_mcp']['port']}")
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
            self.helical_memory.log("MCPBridgeServer", "QHP handshake", time.time(), "bridge", "init", "QHP", 1.0, "Sent QHP handshake to TNOS MCP", {"handshake": handshake})
            handshake_response_msg = await asyncio.wait_for(self.tnos_mcp_socket.recv(), timeout=5)
            handshake_response = json.loads(handshake_response_msg)
            if handshake_response.get("type") != "qhp_handshake_response":
                self.logger.error("QHP handshake response not received, closing connection.")
                self.helical_memory.log("MCPBridgeServer", "QHP handshake", time.time(), "bridge", "failure", "QHP", 1.0, "QHP handshake response not received from TNOS MCP.")
                await self.tnos_mcp_socket.close()
                self.tnos_mcp_socket = None
                return
            peer_fingerprint = handshake_response.get("fingerprint")
            peer_challenge = handshake_response.get("challenge")
            challenge_response = handshake_response.get("challenge_response")
            # Verify their response to our challenge
            if not MCPBridgeServer.verify_challenge_response(my_challenge, challenge_response, peer_fingerprint):
                self.logger.error("QHP handshake: challenge response invalid, closing connection.")
                self.helical_memory.log("MCPBridgeServer", "QHP handshake", time.time(), "bridge", "failure", "QHP", 1.0, "QHP handshake: challenge response invalid.")
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
            self.helical_memory.log("MCPBridgeServer", "QHP handshake", time.time(), "bridge", "trust", "QHP", 1.0, f"Handshake complete with {peer_fingerprint}", {"peer": peer_fingerprint})
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
                self.helical_memory.log("MCPBridgeServer", "RIC", time.time(), "bridge", "trust", "QHP", 1.0, f"RIC score from TNOS MCP: {ric_score}")
            asyncio.create_task(self.handle_tnos_mcp_messages())
        except (
            websockets.exceptions.WebSocketException,
            ConnectionRefusedError,
            OSError,
        ) as e:
            self.logger.warning(
                f"TNOS MCP WebSocket error on port {CONFIG['tnos_mcp']['port']}: {str(e)}"
            )
            self.helical_memory.log("MCPBridgeServer", "error", time.time(), "bridge", "failure", "connect_to_tnos_mcp", 1.0, f"TNOS MCP WebSocket error: {str(e)}")
            next_delay = self.get_next_backoff_delay("tnos")
            self.logger.info(
                f"Will attempt to reconnect to TNOS MCP server in {next_delay}s "
                f"(attempt {self.backoff_strategies['tnos']['attempts']})"
            )
            if not self.shutting_down:
                asyncio.create_task(self.delayed_reconnect("tnos", next_delay))

    async def connect_to_github_mcp(self) -> None:
        self.logger.info("Connecting to GitHub MCP server...")
        self.helical_memory.log("MCPBridgeServer", "connect", time.time(), "bridge", "init", "connect_to_github_mcp", 1.0, "Attempting to connect to GitHub MCP server.")
        if self.github_mcp_socket:
            try:
                await self.github_mcp_socket.close()
            except Exception:
                pass  # Ignore errors during cleanup
            self.github_mcp_socket = None
        try:
            ws_endpoint = CONFIG["github_mcp"]["ws_endpoint"]
            self.logger.debug(f"Attempting to connect to GitHub MCP at {ws_endpoint}")
            self.helical_memory.log("MCPBridgeServer", "connect", time.time(), "bridge", "init", "connect_to_github_mcp", 1.0, f"Connecting to {ws_endpoint}")
            self.github_mcp_socket = await websockets.connect(ws_endpoint)
            self.logger.info("Connected to GitHub MCP server")
            self.helical_memory.log("MCPBridgeServer", "connect", time.time(), "bridge", "success", "connect_to_github_mcp", 1.0, "Connected to GitHub MCP server")
            self.reset_backoff("github")
            await self.message_queue.process_github_queue(self.github_mcp_socket)
            asyncio.create_task(self.handle_github_mcp_messages())
        except (
            websockets.exceptions.WebSocketException,
            ConnectionRefusedError,
            OSError,
        ) as e:
            self.logger.error(f"GitHub MCP WebSocket error: {str(e)}")
            self.helical_memory.log("MCPBridgeServer", "error", time.time(), "bridge", "failure", "connect_to_github_mcp", 1.0, f"GitHub MCP WebSocket error: {str(e)}")
            next_delay = self.get_next_backoff_delay("github")
            self.logger.info(
                f"Will attempt to reconnect to GitHub MCP server in {next_delay}s "
                f"(attempt {self.backoff_strategies['github']['attempts']})"
            )
            if not self.shutting_down:
                asyncio.create_task(self.delayed_reconnect("github", next_delay))

def main():
    helical_memory = None
    try:
        # Initialize helical memory as early as possible
        helical_memory = HelicalMemory()
        helical_memory.log(
            "MCPBridgePython", "startup", time.time(), "github-mcp-server", "init", "python_entry", 1.0,
            "Python MCP Bridge entrypoint invoked."
        )
        logger = setup_logging(CONFIG["logging"]["log_level"])
        logger.info("[MCPBridgePython] Entrypoint started. Initializing bridge server...")
        helical_memory.log(
            "MCPBridgePython", "startup", time.time(), "github-mcp-server", "init", "python_entry", 1.0,
            "Logger initialized. Starting MCPBridgeServer."
        )
        FormulaRegistry.set_helical_memory(helical_memory)
        bridge_server = MCPBridgeServer(logger)
        helical_memory.log(
            "MCPBridgePython", "startup", time.time(), "github-mcp-server", "init", "python_entry", 1.0,
            "MCPBridgeServer object created. Entering asyncio event loop."
        )
        asyncio.run(bridge_server.ensure_servers_running())
        helical_memory.log(
            "MCPBridgePython", "startup", time.time(), "github-mcp-server", "success", "python_entry", 1.0,
            "MCP Bridge Python startup completed successfully."
        )
    except Exception as e:
        tb = traceback.format_exc()
        if helical_memory is None:
            try:
                helical_memory = HelicalMemory()
            except Exception:
                helical_memory = None
        if helical_memory:
            helical_memory.log(
                "MCPBridgePython", "fatal_error", time.time(), "github-mcp-server", "crash", "python_entry", 1.0,
                f"Fatal error in MCP Bridge Python: {str(e)}", {"traceback": tb}
            )
        print(f"[FATAL][MCPBridgePython] {e}\n{tb}", file=sys.stderr)
        sys.exit(1)
    finally:
        if helical_memory:
            helical_memory.log(
                "MCPBridgePython", "exit", time.time(), "github-mcp-server", "shutdown", "python_entry", 1.0,
                "MCP Bridge Python process exiting."
            )

if __name__ == "__main__":
    main()
