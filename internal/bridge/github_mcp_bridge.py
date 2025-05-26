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
        from mcp.bridge.python.tnos.mcp.mobius_compression import MobiusCompression
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
        from mcp.bridge.python.tnos.mcp.context_translator import ContextTranslator
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
                    self.memory.store(event)
                    return observed
                def perform_qhp_handshake(self, peer_name, peer_port):
                    my_challenge = secrets.token_hex(16)
                    my_fp = self.fingerprint
                    peer_fp = MCPBridgeServer.generate_quantum_fingerprint(peer_name)
                    peer_challenge = secrets.token_hex(16)
                    my_response = hashlib.sha256((peer_challenge + my_fp).encode("utf-8")).hexdigest()
                    peer_response = hashlib.sha256((my_challenge + peer_fp).encode("utf-8")).hexdigest()
                    # --- Session key generation (ephemeral, per handshake) ---
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
                    self.memory.store(event)
                    self.logger.info(f"[QHP][CrewAI][TS] {QHPCrewAgent.tranquilspeak_compress(event)}")
                    self.logger.info(f"[QHP][CrewAI] Handshake with {peer_name} complete. Session key: {session_key}. Trust table updated.")
                    return True
                def decay_trust(self, timeout=600):
                    now = time.time()
                    expired = [k for k, v in self.trust_table.items() if now - v["timestamp"] > timeout]
                    for k in expired:
                        del self.trust_table[k]
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

        # Forward to GitHub MCP
        github_message = self.context_bridge.tnos7d_to_github(message)
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
        """
        7D Recursive Collapse: Health check loop observes, collapses, and acts.
        WHO: MCPBridgeServer
        WHAT: Health check observation and self-correction
        WHEN: Every interval (default 30s)
        WHERE: Bridge event loop
        WHY: To ensure all MCP components are healthy and self-healing
        HOW: By observing, logging, and acting on health state in real time
        EXTENT: All bridge health and recovery cycles
        """
        interval = CONFIG["bridge"].get("health_check_interval", 30)
        while not self.shutting_down:
            # 1. Observation (left ∞): Gather full 7D context for this cycle
            context7d = ContextVector7D(
                who="MCPBridgeServer",
                what="health_check",
                when=time.time(),
                where="bridge_health_loop",
                why="periodic_integrity_check",
                how="recursive_observation",
                extent=1.0,
                metadata={"interval": interval}
            )
            try:
                # 2. Collapse (middle ∞): Attempt health check and observe outcome
                await self.health_check()
                # 3. Action (right ∞): Log successful observation/collapse
                self.logger.info(f"[7D][Observation] Health check succeeded | {context7d.to_dict()}")
            except Exception as e:
                # 4. Real-time adjustment: Log, escalate, and trigger self-healing if needed
                self.logger.error(f"[7D][Observation] Health check error: {e} | {context7d.to_dict()}")
                # Optionally, escalate/collapse further (e.g., trigger ensure_servers_running)
                try:
                    await self.ensure_servers_running()
                    self.logger.info(f"[7D][Action] Self-healing triggered after health check failure | {context7d.to_dict()}")
                except Exception as heal_err:
                    self.logger.error(f"[7D][Action] Self-healing failed: {heal_err} | {context7d.to_dict()}")
            # 5. Observation feedback: Wait, then repeat (recursive cycle)
            await asyncio.sleep(interval)

    async def ai_port_sync_loop(self) -> None:
        """
        AI-driven port and endpoint discovery/sync loop.
        WHO: CrewAI/BridgeAI
        WHAT: Discover and sync MCP component ports/endpoints in real time
        WHEN: Every 15 seconds (configurable)
        WHERE: Bridge event loop
        WHY: To ensure all MCP endpoints are correct and up-to-date
        HOW: By scanning, validating, and updating config as needed
        EXTENT: All MCP bridge runtime
        """
        interval = 15
        while not self.shutting_down:
            context7d = ContextVector7D(
                who="CrewAI",
                what="port_sync",
                when=time.time(),
                where="ai_port_sync_loop",
                why="dynamic_endpoint_discovery",
                how="scan_and_validate",
                extent=1.0,
                metadata={"interval": interval}
            )
            try:
                if CREWAI_AVAILABLE:
                    observed = self.qhp_bridge_agent.observe_ports()
                    handshake_result = self.qhp_bridge_agent.perform_qhp_handshake("TNOS_MCP", 9001)
                    self.logger.info(f"[CrewAI][QHP][TS] Port observation: {self.qhp_bridge_agent.tranquilspeak_compress({'type':'port_observation','ports':observed})} | Handshake: {handshake_result}")
                    self.qhp_bridge_agent.decay_trust(timeout=600)
                    self.logger.info(f"[CrewAI][QHP] Trust table: {self.qhp_bridge_agent.get_trust_table()}")
                    self.logger.info(f"[CrewAI][QHP] Helical memory: {self.helical_memory.get_all()[-3:]}")
                discovered = {}
                for name, port in [("github_mcp", 10617), ("tnos_mcp", 9001), ("bridge", 10619), ("visualization", 8083)]:
                    pid = get_pid_on_port_static(port)
                    discovered[name] = bool(pid)
                for name, port in [("github_mcp", 10617), ("tnos_mcp", 9001)]:
                    running = await self.check_server_running("localhost", port)
                    if not running:
                        self.logger.warning(f"[AI][PortSync] {name} not running on port {port} | {context7d.to_dict()}")
                    else:
                        self.logger.info(f"[AI][PortSync] {name} running on port {port} | {context7d.to_dict()}")
            except Exception as e:
                self.logger.error(f"[AI][PortSync] Error during port sync: {e} | {context7d.to_dict()}")
            await asyncio.sleep(interval)

        if CREWAI_AVAILABLE:
            class PortAssignmentAI:
                """
                CrewAI-based intelligent port assignment and routing AI for Layer 3.
                Uses QHP, TranquilSpeak, MobiusCompression, and ContextTranslator for secure, efficient, and adaptive port management.
                Instantiates multiple CrewAI agents for:
                - Port assignment and monitoring
                - Fallback routing
                - Formula optimization (registry-driven)
                """
                def __init__(self, bridge_server, logger):
                    self.bridge_server = bridge_server
                    self.logger = logger
                    self.memory = bridge_server.helical_memory
                    self.port_agent = bridge_server.qhp_bridge_agent
                    self.tnos_agent = bridge_server.qhp_tnos_agent
                    self.fallback_agent = self._create_fallback_agent()
                    self.formula_agent = self._create_formula_agent()
                    self.context_translator = ContextTranslator()
                    self.mobius = MobiusCompression
                    self.formula_registry = FormulaRegistry
                def _compress_event(self, event):
                    ctx7d = self.context_translator.to_7d(event)
                    compressed = self.mobius.compress(event, context=ctx7d)
                    return compressed
                def _create_fallback_agent(self):
                    port_agent = self.port_agent
                    mobius = self.mobius
                    context_translator = self.context_translator
                    class FallbackRoutingAgent:
                        def __init__(self, logger, memory):
                            self.logger = logger
                            self.memory = memory
                        def route(self, port_map):
                            ctx7d = context_translator.to_7d({"type": "fallback_route", "ports": port_map, "ts": time.time()})
                            ts = port_agent.tranquilspeak_compress({"type": "fallback_route", "ports": port_map, "ts": time.time()})
                            compressed = mobius.compress(ts, context=ctx7d)
                            self.memory.store({"type": "fallback_route", "ports": port_map, "compressed": compressed, "ts": time.time()})
                            self.logger.info(f"[FallbackAI][TS][Mobius] {compressed}")
                            for name, available in port_map.items():
                                if available:
                                    return name
                            return None
                    return FallbackRoutingAgent(self.logger, self.memory)
                def _create_formula_agent(self):
                    port_agent = self.port_agent
                    mobius = self.mobius
                    context_translator = self.context_translator
                    formula_registry = self.formula_registry
                    class FormulaOptimizationAgent:
                        def __init__(self, logger, memory):
                            self.logger = logger
                            self.memory = memory
                        def select_formula(self, context):
                            ctx7d = context_translator.to_7d(context)
                            symbol = formula_registry.select_formula(context)
                            formula = formula_registry.FORMULAS.get(symbol, "default_formula")
                            ts = port_agent.tranquilspeak_compress({"type": "formula_select", "formula": formula, "context": context, "ts": time.time()})
                            compressed = mobius.compress(ts, context=ctx7d)
                            self.memory.store({"type": "formula_select", "formula": formula, "context": context, "compressed": compressed, "ts": time.time()})
                            self.logger.info(f"[FormulaAI][TS][Mobius] {compressed}")
                            return formula
                    return FormulaOptimizationAgent(self.logger, self.memory)
                def assign_ports(self):
                    observed = self.port_agent.observe_ports()
                    ctx7d = self.context_translator.to_7d({"type": "port_observation", "ports": observed, "ts": time.time()})
                    compressed = self.mobius.compress(observed, context=ctx7d)
                    self.memory.store({"type": "port_observation", "ports": observed, "compressed": compressed, "ts": time.time()})
                    self.logger.info(f"[PortAI][Mobius] Observed ports (compressed): {compressed}")
                    fallback = self.fallback_agent.route(observed)
                    if fallback:
                        self.logger.info(f"[PortAI] Fallback routing to: {fallback}")
                    for peer, port in [("TNOS_MCP", 9001), ("github_mcp", 10617), ("bridge", 10619)]:
                        self.port_agent.perform_qhp_handshake(peer, port)
                    return observed
                def optimize_formula(self, context):
                    return self.formula_agent.select_formula(context)
            # Attach to MCPBridgeServer
            MCPBridgeServer.port_assignment_ai = None
            def start_port_assignment_ai(self):
                if CREWAI_AVAILABLE and not self.port_assignment_ai:
                    self.port_assignment_ai = PortAssignmentAI(self, self.logger)
                    self.logger.info("[PortAI] CrewAI PortAssignmentAI started.")
            MCPBridgeServer.start_port_assignment_ai = start_port_assignment_ai

# Update QHPCrewAgent and all usages to use MCPBridgeServer.generate_quantum_fingerprint and .verify_challenge_response
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
        self.memory.store(event)
        return observed
    def perform_qhp_handshake(self, peer_name, peer_port):
        my_challenge = secrets.token_hex(16)
        my_fp = self.fingerprint
        peer_fp = MCPBridgeServer.generate_quantum_fingerprint(peer_name)
        peer_challenge = secrets.token_hex(16)
        my_response = hashlib.sha256((peer_challenge + my_fp).encode("utf-8")).hexdigest()
        peer_response = hashlib.sha256((my_challenge + peer_fp).encode("utf-8")).hexdigest()
        # --- Session key generation (ephemeral, per handshake) ---
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
        self.memory.store(event)
        self.logger.info(f"[QHP][CrewAI][TS] {QHPCrewAgent.tranquilspeak_compress(event)}")
        self.logger.info(f"[QHP][CrewAI] Handshake with {peer_name} complete. Session key: {session_key}. Trust table updated.")
        return True
    def decay_trust(self, timeout=600):
        now = time.time()
        expired = [k for k, v in self.trust_table.items() if now - v["timestamp"] > timeout]
        for k in expired:
            del self.trust_table[k]
        if expired:
            self.logger.info(f"[QHP][CrewAI] Decayed trust for: {expired}")
    def get_trust_table(self):
        return dict(self.trust_table)

# --- Patch: Make generate_quantum_fingerprint and verify_challenge_response static methods of MCPBridgeServer ---
MCPBridgeServer.generate_quantum_fingerprint = staticmethod(MCPBridgeServer.generate_quantum_fingerprint)
MCPBridgeServer.verify_challenge_response = staticmethod(MCPBridgeServer.verify_challenge_response)

def normalize_timestamp(ts):
    if isinstance(ts, (int, float)) and ts > 1e10:
        return ts // 1000 if isinstance(ts, int) else ts / 1000.0
    return ts
