#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
GitHub MCP to TNOS MCP Bridge
------------------------------
This module provides bidirectional communication between the GitHub MCP Server
and the TNOS Model Context Protocol implementation. It translates context and
capabilities between the two systems, allowing GitHub Copilot in VS Code to
access and utilize TNOS capabilities.

Created: April 11, 2025
"""

import collections
import copy
import hashlib
import hmac
import importlib.util
import json
import logging
import os
import re
import socket
import sys
import time
import uuid
from pathlib import Path
from typing import Any, Dict, List, Optional, Set, Tuple, Union


# Define a PathManager class for robust path handling
class PathManager:
    """
    Manages paths for the GitHub-TNOS MCP Bridge to ensure
    consistent and reliable path resolution across different environments.
    """

    def __init__(self):
        # Get the absolute path to the script directory
        self.script_dir = os.path.dirname(os.path.abspath(__file__))

        # Resolve the project root (parent of parent directory)
        self.project_root = os.path.abspath(os.path.join(self.script_dir, "../.."))

        # Verify the project structure to ensure we found the correct root
        self._verify_project_structure()

        # Define standard paths
        self.logs_dir = os.path.join(self.project_root, "logs")
        self.config_dir = os.path.join(self.project_root, "config", "mcp")
        self.mcp_dir = os.path.join(self.project_root, "mcp")

        # Create required directories
        self._ensure_directories()

    def _verify_project_structure(self):
        """Verify we have the correct project root by checking for essential directories"""
        expected_dirs = ["mcp", "config", "scripts"]
        missing_dirs = []

        for directory in expected_dirs:
            if not os.path.isdir(os.path.join(self.project_root, directory)):
                missing_dirs.append(directory)

        if missing_dirs:
            logging.error(
                f"Project structure verification failed. Missing: {
    ', '.join(missing_dirs)}"
            )
            logging.error("Project root was detected as: %s", self.project_root)
            logging.error(
                "This might indicate incorrect path resolution or project setup issues."
            )
            # We'll continue anyway since we may be in a non-standard setup

    def _ensure_directories(self):
        """Create required directories if they don't exist"""
        os.makedirs(self.logs_dir, exist_ok=True)
        os.makedirs(self.config_dir, exist_ok=True)

        # Create schemas directory
        schemas_dir = os.path.join(
            os.path.dirname(os.path.abspath(__file__)), "schemas"
        )
        os.makedirs(schemas_dir, exist_ok=True)

    def get_log_path(self, service_name: str) -> str:
        """Get the path to a log file for a given service"""
        return os.path.join(self.logs_dir, f"{service_name}.log")

    def get_config_path(self, filename: str) -> str:
        """Get the path to a configuration file"""
        return os.path.join(self.config_dir, filename)

    def get_pid_path(self, service_name: str) -> str:
        """Get the path to a PID file for a given service"""
        return os.path.join(self.logs_dir, f"{service_name}.pid")

    def add_to_python_path(self) -> bool:
        """Add project root to Python path if not already added"""
        if self.project_root not in sys.path:
            sys.path.insert(0, self.project_root)
            return True
        return False

    def resolve_path(self, path: str) -> str:
        """Resolve a path relative to the project root"""
        if os.path.isabs(path):
            return path
        return os.path.normpath(os.path.join(self.project_root, path))


# Initialize the path manager early
path_manager = PathManager()

# Configure logging early so we can log import errors
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[
        logging.FileHandler(path_manager.get_log_path("github_tnos_bridge")),
        logging.StreamHandler(),
    ],
)
logger = logging.getLogger("github_mcp_bridge")

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

# Now import TNOS MCP components (only if all required packages are available)
try:
    from mcp.integration.mcp_integration import TNOSLayerIntegration
    from mcp.protocol.mcp_protocol import MCPContext, MCPMessage, MCPProtocolVersion
    from mcp.server.mcp_server import MCPServer
except ImportError as e:
    logger.error(f"Failed to import TNOS MCP components: {e}")
    logger.error(
        "Make sure you're running this script from the project root or have PYTHONPATH set correctly"
    )
    logger.error(f"Current Python path: {sys.path}")
    sys.exit(1)


class VersionManager:
    """
    Manages MCP protocol version compatibility
    """

    SUPPORTED_VERSIONS = ["1.0", "1.5", "2.0", "2.5", "3.0"]

    @staticmethod
    def get_latest_version():
        """Get the latest supported version"""
        return VersionManager.SUPPORTED_VERSIONS[-1]

    @staticmethod
    def is_supported(version: str) -> bool:
        """Check if a version is supported"""
        return version in VersionManager.SUPPORTED_VERSIONS

    @staticmethod
    def get_fallback_version(requested_version: str) -> str:
        """Get a fallback version for an unsupported version"""
        if requested_version not in VersionManager.SUPPORTED_VERSIONS:
            # Find the closest older version
            for v in reversed(VersionManager.SUPPORTED_VERSIONS):
                if v < requested_version:
                    return v
            # If no older version, use the oldest supported version
            return VersionManager.SUPPORTED_VERSIONS[0]
        return requested_version

    @staticmethod
    def negotiate_version(client_versions: List[str]) -> str:
        """Negotiate a common version between client and server"""
        # Find the highest version supported by both
        for v in reversed(VersionManager.SUPPORTED_VERSIONS):
            if v in client_versions:
                return v

        # If no common version, return the latest supported version
        # The client will need to adapt
        return VersionManager.get_latest_version()


class ConfigManager:
    """
    Manages configuration for the GitHub-TNOS MCP Bridge
    """

    def __init__(self, path_manager: PathManager):
        self.path_manager = path_manager
        self.config_file = path_manager.get_config_path("github_tnos_tools.json")
        self.config = self._load_config()

    def _load_config(self) -> Dict[str, Any]:
        """Load configuration from file or create default"""
        if os.path.exists(self.config_file):
            try:
                with open(self.config_file, "r") as f:
                    config = json.load(f)
                logger.info(f"Loaded configuration from {self.config_file}")
                return config
            except (json.JSONDecodeError, FileNotFoundError) as e:
                logger.warning(f"Error loading configuration: {e}")

        # Return empty config if we couldn't load one
        return {}

    def save_config(self, config: Dict[str, Any]) -> bool:
        """Save configuration to file"""
        try:
            with open(self.config_file, "w") as f:
                json.dump(config, f, indent=2)
            logger.info(f"Saved configuration to {self.config_file}")
            return True
        except Exception as e:
            logger.error(f"Error saving configuration: {e}")
            return False

    def validate_config(self, config: Dict[str, Any]) -> bool:
        """Validate that the configuration has the required structure"""
        try:
            # Use jsonschema for validation if available
            if "jsonschema" in sys.modules:
                schema = {
                    "type": "object",
                    "required": ["toolDefinitions", "serverSettings"],
                    "properties": {
                        "toolDefinitions": {
                            "type": "array",
                            "items": {
                                "type": "object",
                                "required": ["name", "description", "parameters"],
                                "properties": {
                                    "name": {"type": "string"},
                                    "description": {"type": "string"},
                                    "parameters": {"type": "object"},
                                },
                            },
                        },
                        "serverSettings": {
                            "type": "object",
                            "required": ["bridgeUri"],
                            "properties": {"bridgeUri": {"type": "string"}},
                        },
                    },
                }
                jsonschema.validate(config, schema)
                return True

            # Fallback validation if jsonschema is not available
            else:
                # Check for required top-level keys
                required_keys = ["toolDefinitions", "serverSettings"]
                for key in required_keys:
                    if key not in config:
                        logger.error(f"Configuration missing required key: {key}")
                        return False

                # Check serverSettings
                server_settings = config.get("serverSettings", {})
                if not server_settings.get("bridgeUri"):
                    logger.error("Configuration missing serverSettings.bridgeUri")
                    return False

                # Check toolDefinitions
                tool_definitions = config.get("toolDefinitions", [])
                if not isinstance(tool_definitions, list) or not tool_definitions:
                    logger.error("Configuration missing or invalid toolDefinitions")
                    return False

                # Check each tool definition has required keys
                for tool in tool_definitions:
                    if not all(
                        key in tool for key in ["name", "description", "parameters"]
                    ):
                        logger.error(f"Tool definition missing required keys: {tool}")
                        return False

                return True

        except Exception as e:
            logger.error(f"Configuration validation error: {e}")
            return False


class PortManager:
    """
    Manages port status checking and verification
    """

    @staticmethod
    def check_port_in_use(host: str, port: int) -> bool:
        """Check if a port is in use"""
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            return s.connect_ex((host, port)) == 0

    @staticmethod
    def wait_for_port(host: str, port: int, timeout: int = 10) -> bool:
        """Wait for a port to become available or in use"""
        start_time = time.time()
        while time.time() - start_time < timeout:
            if PortManager.check_port_in_use(host, port):
                return True
            time.sleep(0.5)
        return False


class EnhancedMessageValidator:
    """
    Enhanced message validator with stronger authentication and security checks
    """

    def __init__(self, config_manager):
        self.config_manager = config_manager
        self.token_manager = None  # Will be set later
        self.message_schemas = {}
        self.load_message_schemas()

        # Track potentially malicious patterns
        self.suspicious_patterns = []

        # HMAC for message authentication
        self._hmac_key = self._generate_hmac_key()

        # Message replay protection
        self.message_ids = collections.deque(
            maxlen=1000
        )  # Remember last 1000 message IDs

    def set_token_manager(self, token_manager):
        """Set the token manager for validation operations that require token access"""
        self.token_manager = token_manager

    def _generate_hmac_key(self):
        """Generate a HMAC key for message authentication"""
        # Generate a machine-specific HMAC key
        try:
            machine_id = str(uuid.getnode())
            base_path = os.path.dirname(os.path.abspath(__file__))
            key_material = f"{machine_id}:{base_path}:{os.getlogin()}"
            return hashlib.sha256(key_material.encode()).digest()
        except Exception as e:
            # Fall back to a random key if machine-specific information isn't
            # available
            logger.warning(f"Using random HMAC key due to error: {str(e)}")
            return os.urandom(32)

    def load_message_schemas(self):
        """Load the JSON schemas for message validation"""
        schema_dir = os.path.join(os.path.dirname(os.path.abspath(__file__)), "schemas")
        try:
            if os.path.exists(schema_dir):
                for schema_file in os.listdir(schema_dir):
                    if schema_file.endswith(".json"):
                        schema_path = os.path.join(schema_dir, schema_file)
                        with open(schema_path, "r") as f:
                            schema_name = os.path.splitext(schema_file)[0]
                            self.message_schemas[schema_name] = json.load(f)
            else:
                logger.warning(f"Schema directory not found: {schema_dir}")
        except Exception as e:
            logger.error(f"Error loading message schemas: {str(e)}")

    def sign_message(self, message):
        """
        Add authentication signature to outgoing messages

        Args:
            message: The message to sign

        Returns:
            dict: The message with signature added
        """
        if not isinstance(message, dict):
            return message

        # Create a copy of the message to avoid modifying the original
        signed_message = copy.deepcopy(message)

        # Add message ID and timestamp for replay protection
        signed_message["meta"] = signed_message.get("meta", {})
        signed_message["meta"]["message_id"] = str(uuid.uuid4())
        signed_message["meta"]["timestamp"] = time.time()

        # Create the signature
        message_str = json.dumps(signed_message, sort_keys=True)
        signature = hmac.new(
            self._hmac_key, message_str.encode(), hashlib.sha256
        ).hexdigest()

        # Add the signature
        signed_message["meta"]["signature"] = signature

        return signed_message

    def verify_message(self, message):
        """
        Verify the authenticity of incoming messages

        Args:
            message: The message to verify

        Returns:
            tuple: (is_valid, reason)
        """
        if not isinstance(message, dict):
            return False, "Message is not a dictionary"

        # Check for meta information
        if "meta" not in message:
            return False, "Missing meta information"

        meta = message.get("meta", {})
        if not isinstance(meta, dict):
            return False, "Meta is not a dictionary"

        # Check for required fields
        if "message_id" not in meta:
            return False, "Missing message ID"

        if "timestamp" not in meta:
            return False, "Missing timestamp"

        if "signature" not in meta:
            return False, "Missing signature"

        # Check message freshness (within 5 minutes)
        message_time = meta["timestamp"]
        current_time = time.time()
        if current_time - message_time > 300:  # 5 minutes
            return False, "Message expired"

        # Check for replay attacks
        message_id = meta["message_id"]
        if message_id in self.message_ids:
            logger.warning(
                f"⚠️ SECURITY ALERT: Possible replay attack detected with message ID: {message_id}"
            )
            return False, "Message replay detected"
        self.message_ids.append(message_id)

        # Verify signature
        original_signature = meta["signature"]
        verification_message = copy.deepcopy(message)
        verification_message["meta"] = copy.deepcopy(meta)
        del verification_message["meta"]["signature"]

        message_str = json.dumps(verification_message, sort_keys=True)
        expected_signature = hmac.new(
            self._hmac_key, message_str.encode(), hashlib.sha256
        ).hexdigest()

        if not hmac.compare_digest(original_signature, expected_signature):
            logger.warning(f"⚠️ SECURITY ALERT: Invalid message signature detected")
            return False, "Invalid signature"

        return True, "Message verified"

    def validate_message_schema(self, message, schema_name):
        """
        Validate a message against its schema

        Args:
            message: The message to validate
            schema_name: The name of the schema to validate against

        Returns:
            tuple: (is_valid, error_message)
        """
        if schema_name not in self.message_schemas:
            return False, f"Schema not found: {schema_name}"

        try:
            jsonschema.validate(message, self.message_schemas[schema_name])
            return True, "Message validated"
        except jsonschema.exceptions.ValidationError as e:
            return False, str(e)

    def scan_for_security_threats(self, message):
        """
        Scan messages for potential security threats

        Args:
            message: The message to scan

        Returns:
            tuple: (is_safe, threat_details)
        """
        # Convert message to string for scanning
        if isinstance(message, dict):
            message_str = json.dumps(message)
        else:
            message_str = str(message)

        # Check for common injection patterns
        injection_patterns = [
            r"(?i)(?:--.*?$|\/\*.*?\*\/|;.*?$)",  # SQL comments/semicolon
            r"(?i)(?:<script.*?>.*?<\/script>)",  # XSS scripts
            r"(?i)(?:\w+\s*\(\s*\)\s*\{)",  # Command injection
            r"(?:\$\{.*?\})",  # Template injection
            r"(?:%[0-9a-fA-F]{2})",  # URL encoded attacks
        ]

        for pattern in injection_patterns:
            if re.search(pattern, message_str):
                threat = f"Potential injection attack detected: {pattern}"
                self.suspicious_patterns.append(
                    {
                        "pattern": pattern,
                        "timestamp": time.time(),
                        "sample": message_str[:100]
                        + ("..." if len(message_str) > 100 else ""),
                    }
                )
                logger.warning(f"⚠️ SECURITY ALERT: {threat}")
                return False, threat

        # Check for excessive data lengths (potential DoS)
        if len(message_str) > 10 * 1024 * 1024:  # 10MB
            threat = f"Message size exceeds limit: {len(message_str)} bytes"
            logger.warning(f"⚠️ SECURITY ALERT: {threat}")
            return False, threat

        return True, "No threats detected"

    def get_security_statistics(self):
        """Get statistics about security scans for monitoring"""
        return {
            "suspicious_patterns_detected": len(self.suspicious_patterns),
            "recent_suspicious_patterns": [
                p
                for p in self.suspicious_patterns
                if time.time() - p["timestamp"] < 3600
            ],  # Last hour
            "message_replay_attempts": sum(1 for _ in self.message_ids),
        }

    # WHO: EnhancedMessageValidator
    # WHAT: GitHub request validation method
    # WHEN: During message processing
    # WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
    # WHY: To ensure request format compliance
    # HOW: Using JSON schema validation
    # EXTENT: All GitHub MCP requests

    def validate_github_request(self, request):
        """
        Validate a GitHub MCP request against the schema

        Args:
            request: The GitHub MCP request to validate

        Returns:
            tuple: (is_valid, error_message)
        """
        # Basic type checking
        if not isinstance(request, dict):
            return False, "Request must be a dictionary"

        # Check for required fields
        if "name" not in request:
            return False, "Missing required field: name"

        if "parameters" not in request:
            return False, "Missing required field: parameters"

        # Check field types
        if not isinstance(request.get("name"), str):
            return False, "Field 'name' must be a string"

        if not isinstance(request.get("parameters"), dict):
            return False, "Field 'parameters' must be an object"

        # Check for security threats
        is_safe, threat_details = self.scan_for_security_threats(request)
        if not is_safe:
            return False, f"Security threat detected: {threat_details}"

        # Use schema validation if available
        if "request" in self.message_schemas:
            try:
                jsonschema.validate(request, self.message_schemas["request"])
            except jsonschema.exceptions.ValidationError as e:
                return False, f"Schema validation failed: {e}"

        # If we made it here, the request is valid
        return True, "Request validated successfully"


class TokenManager:
    """
    Securely handles and manages authentication tokens
    """

    def __init__(self, path_manager: PathManager):
        self.path_manager = path_manager
        self.token_env_var = "GITHUB_PERSONAL_ACCESS_TOKEN"
        self.token_file = path_manager.get_config_path("github_token.enc")
        self.token = None
        self.token_expiry = None
        self.token_refresh_threshold = 24 * 60 * 60  # 24 hours in seconds
        self.max_token_age = 7 * 24 * 60 * 60  # 7 days in seconds
        self.last_validation_time = 0
        self.validation_cache_ttl = 60 * 60  # 1 hour in seconds

    def get_token(self) -> Optional[str]:
        """
        Retrieve the GitHub personal access token securely

        First checks environment variable, then secure storage.
        Performs validation and checks for token expiration.
        """
        # First try environment variable (highest priority)
        token = os.environ.get(self.token_env_var)
        if token:
            logger.info(f"Using GitHub token from environment variable")
            # Validate the token before using it
            if not self.validate_token(token):
                logger.error(f"Invalid GitHub token format in environment variable")
                return None
            self.token = token
            return token

        # Next try to load from secure storage
        try:
            if os.path.exists(self.token_file):
                import base64

                from cryptography.fernet import Fernet, InvalidToken
                from cryptography.hazmat.primitives import hashes
                from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2HMAC

                # Get machine-specific key
                hostname = socket.gethostname()
                user = os.getlogin()
                salt = b"TNOS_MCP_Bridge_Salt"

                # Generate key from machine-specific information
                kdf = PBKDF2HMAC(
                    algorithm=hashes.SHA256(),
                    length=32,
                    salt=salt,
                    iterations=100000,
                )
                key = base64.urlsafe_b64encode(kdf.derive((hostname + user).encode()))
                cipher = Fernet(key)

                with open(self.token_file, "rb") as f:
                    try:
                        encrypted_data = f.read()
                        decrypted_data = cipher.decrypt(encrypted_data)

                        # Parse token and expiry information
                        data = json.loads(decrypted_data.decode())
                        token = data.get("token")
                        self.token_expiry = data.get("expiry")

                        # Check token expiration
                        if self.token_expiry and time.time() > self.token_expiry:
                            logger.warning(
                                "Stored GitHub token has expired, need to refresh"
                            )
                            return None

                        # Validate the token format
                        if not self.validate_token(token):
                            logger.error(
                                "Invalid GitHub token format in secure storage"
                            )
                            return None

                        logger.info(f"Loaded GitHub token from secure storage")
                        self.token = token
                        return token
                    except InvalidToken:
                        logger.error("Failed to decrypt token: invalid encryption")
                        # Delete corrupted token file
                        os.unlink(self.token_file)
                        return None
                    except json.JSONDecodeError:
                        logger.error("Failed to parse token data")
                        return None
        except Exception as e:
            logger.warning(f"Failed to load token from secure storage: {e}")

        logger.error(
            f"{self.token_env_var} not set. Please set this environment variable or use store_token()."
        )
        return None

    def store_token(self, token: str, expiry_days: int = 30) -> bool:
        """
        Securely store a GitHub personal access token with expiration

        Args:
            token: The GitHub personal access token to store
            expiry_days: Number of days until the token expires (default: 30)

        Returns:
            bool: True if token was stored successfully, False otherwise
        """
        # Validate the token first
        if not self.validate_token(token):
            logger.error("Cannot store invalid token format")
            return False

        try:
            import base64

            from cryptography.fernet import Fernet
            from cryptography.hazmat.primitives import hashes
            from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2HMAC

            # Get machine-specific key
            hostname = socket.gethostname()
            user = os.getlogin()
            salt = b"TNOS_MCP_Bridge_Salt"

            # Generate key from machine-specific information
            kdf = PBKDF2HMAC(
                algorithm=hashes.SHA256(),
                length=32,
                salt=salt,
                iterations=100000,
            )
            key = base64.urlsafe_b64encode(kdf.derive((hostname + user).encode()))
            cipher = Fernet(key)

            # Calculate expiry time
            expiry_time = int(time.time() + (expiry_days * 24 * 60 * 60))

            # Create token data with expiration
            token_data = {
                "token": token,
                "expiry": expiry_time,
                "created": int(time.time()),
            }

            # Encrypt the token data
            encrypted_data = cipher.encrypt(json.dumps(token_data).encode())

            # Create directory if it doesn't exist
            os.makedirs(os.path.dirname(self.token_file), exist_ok=True)

            # Write with secure permissions
            with open(self.token_file, "wb") as f:
                f.write(encrypted_data)

            # Set secure permissions (only owner can read)
            os.chmod(self.token_file, 0o600)

            logger.info(
                f"Stored GitHub token in secure storage with expiration in {expiry_days} days"
            )

            # Update instance variables
            self.token = token
            self.token_expiry = expiry_time

            return True
        except Exception as e:
            logger.error(f"Failed to store token securely: {e}")
            return False

    def validate_token(self, token: str) -> bool:
        """
        Validate a GitHub personal access token format

        Args:
            token: The token to validate

        Returns:
            bool: True if the token has a valid format, False otherwise
        """
        if not token or len(token) < 30:
            return False

        # GitHub tokens start with 'ghp_', 'github_pat_' or 'gho_'
        if not token.startswith(("ghp_", "github_pat_", "gho_")):
            return False

        # Basic format validation
        if not all(c.isalnum() or c == "_" for c in token):
            return False

        return True

    def is_token_valid(self) -> bool:
        """
        Check if the current token is valid and not expired

        Returns:
            bool: True if the token is valid and not expired, False otherwise
        """
        # Get the token first
        if not self.token:
            self.get_token()

        if not self.token:
            return False

        # Check expiration if available
        if self.token_expiry and time.time() > self.token_expiry:
            return False

        # Only perform online validation if we haven't done so recently
        current_time = time.time()
        if current_time - self.last_validation_time > self.validation_cache_ttl:
            # Implement GitHub API validation here if needed
            # For now, just check the format
            is_valid = self.validate_token(self.token)
            self.last_validation_time = current_time
            return is_valid

        return True

    def rotate_token(self, new_token: str) -> bool:
        """
        Rotate to a new GitHub personal access token

        Args:
            new_token: The new token to store

        Returns:
            bool: True if rotation was successful, False otherwise
        """
        # Validate the new token first
        if not self.validate_token(new_token):
            logger.error("Cannot rotate to invalid token format")
            return False

        # Store the new token
        if self.store_token(new_token):
            logger.info("Successfully rotated GitHub token")
            return True
        else:
            logger.error("Failed to rotate GitHub token")
            return False


class RateLimiter:
    """
    Implements rate limiting to prevent flooding and API abuse
    """

    def __init__(self):
        # Default rate limits
        self.global_limit = 100  # requests per minute globally
        self.tool_limits = {
            # Higher limits for read operations
            "semantic_search": 30,
            "list_code_usages": 20,
            "read_file": 40,
            "list_dir": 30,
            "get_errors": 20,
            # Lower limits for potentially expensive operations
            "run_in_terminal": 5,
            "get_terminal_output": 10,
            # Moderate limits for file operations
            "file_search": 15,
            "grep_search": 15,
            # Lower limits for file modification operations
            "insert_edit_into_file": 10,
        }

        # Storage for rate tracking
        self.global_requests = []
        self.tool_requests = {}

        # Tracking window in seconds
        self.window = 60

        # Initialize tool tracking
        for tool in self.tool_limits:
            self.tool_requests[tool] = []

        # For security auditing
        self.exceeded_attempts = {}

    def _clean_old_requests(self, request_list):
        """Remove requests older than the window from request_list"""
        current_time = time.time()
        return [
            req_time
            for req_time in request_list
            if current_time - req_time < self.window
        ]

    def check_rate_limit(self, tool_name):
        """
        Check if a request should be allowed based on rate limits

        Args:
            tool_name: The name of the tool being called

        Returns:
            bool: True if the request is allowed, False if rate limit exceeded
        """
        current_time = time.time()

        # Clean up old requests
        self.global_requests = self._clean_old_requests(self.global_requests)

        # Check global rate limit
        if len(self.global_requests) >= self.global_limit:
            logger.warning(
                f"Global rate limit exceeded: {
                    len(
                        self.global_requests)} requests in the last minute"
            )
            self._track_exceeded_attempt("global", tool_name)
            return False

        # Clean up tool-specific requests
        if tool_name in self.tool_requests:
            self.tool_requests[tool_name] = self._clean_old_requests(
                self.tool_requests[tool_name]
            )

            # Check tool-specific rate limit
            if (
                tool_name in self.tool_limits
                and len(self.tool_requests[tool_name]) >= self.tool_limits[tool_name]
            ):
                logger.warning(
                    f"Tool-specific rate limit exceeded for {tool_name}: {
                        len(
                            self.tool_requests[tool_name])} requests in the last minute"
                )
                self._track_exceeded_attempt(tool_name, tool_name)
                return False

        # Add request to tracking
        self.global_requests.append(current_time)
        if tool_name in self.tool_requests:
            self.tool_requests[tool_name].append(current_time)

        return True

    def _track_exceeded_attempt(self, limit_type, tool_name):
        """Track rate limit violations for security auditing"""
        current_time = time.time()
        key = f"{limit_type}:{tool_name}"

        if key not in self.exceeded_attempts:
            self.exceeded_attempts[key] = []

        self.exceeded_attempts[key].append(current_time)

        # If we see multiple exceeded attempts in a short period, log a
        # security warning
        recent_attempts = [
            t for t in self.exceeded_attempts[key] if current_time - t < 300
        ]  # 5 minutes
        if len(recent_attempts) >= 5:
            logger.warning(
                f"⚠️ SECURITY ALERT: Possible DoS attempt detected - {
                    len(recent_attempts)} rate limit violations for {tool_name} in the last 5 minutes"
            )

    def get_remaining_quota(self, tool_name):
        """
        Get the remaining request quota for a specific tool

        Args:
            tool_name: The name of the tool

        Returns:
            dict: Information about remaining quota
        """
        current_time = time.time()

        # Clean up tracking
        self.global_requests = self._clean_old_requests(self.global_requests)
        if tool_name in self.tool_requests:
            self.tool_requests[tool_name] = self._clean_old_requests(
                self.tool_requests[tool_name]
            )

        # Calculate remaining quota
        global_remaining = self.global_limit - len(self.global_requests)
        tool_remaining = self.tool_limits.get(tool_name, float("inf")) - len(
            self.tool_requests.get(tool_name, [])
        )

        return {
            "global_remaining": global_remaining,
            "tool_remaining": tool_remaining if tool_name in self.tool_limits else None,
            "reset_in_seconds": self.window
            - (current_time - min(self.global_requests) if self.global_requests else 0),
        }

    def update_limits(self, global_limit=None, tool_limits=None):
        """
        Update rate limits based on system load or configuration

        Args:
            global_limit: New global rate limit (requests per minute)
            tool_limits: Dictionary of new tool-specific limits
        """
        if global_limit is not None:
            self.global_limit = global_limit

        if tool_limits is not None:
            for tool, limit in tool_limits.items():
                self.tool_limits[tool] = limit
                if tool not in self.tool_requests:
                    self.tool_requests[tool] = []


class GitHubTNOSBridge:
    """
    Bridge between GitHub MCP Server and TNOS MCP implementation.
    Translates requests, contexts, and capabilities between the two systems.
    """

    def __init__(
        self, tnos_server_uri: str = "ws://localhost:8888", github_mcp_port: int = 8080
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

        self.tnos_server_uri = tnos_server_uri
        self.github_mcp_port = github_mcp_port
        self.tnos_ws_connection = None
        self.github_clients = set()
        self.session_mapping = {}  # Maps GitHub session IDs to TNOS session IDs
        self.tool_mapping = self._build_tool_mapping()
        self.version_manager = VersionManager()
        self.message_validator = EnhancedMessageValidator(self.config_manager)
        self.message_validator.set_token_manager(self.token_manager)
        self.rate_limiter = RateLimiter()

        # Protocol version negotiation
        self.supported_versions = VersionManager.SUPPORTED_VERSIONS
        self.current_version = VersionManager.get_latest_version()

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
        # Extract information from GitHub request
        tool_name = github_request.get("name", "")
        params = github_request.get("parameters", {})

        # Map to 7D context dimensions
        context = MCPContext()

        # WHO: The actor making the request (GitHub Copilot or user)
        context.who = "GitHub Copilot"

        # WHAT: The capability being requested (mapped from GitHub tool)
        context.what = self.tool_mapping.get(tool_name, "unknown_operation")

        # WHEN: Current timestamp (set by MCPContext constructor)

        # WHERE: Location context
        repo_info = f"{params.get('owner', '')}/{params.get('repo', '')}"
        if "path" in params:
            repo_info += f"/{params.get('path', '')}"
        context.where = "/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp"

        # WHY: Purpose of the operation
        context.why = f"GitHub operation: {tool_name}"

        # HOW: The method being used
        context.how = "github_mcp_bridge"

        # TO_WHAT_EXTENT: Scope of the operation
        # Determine based on parameters if it's a single resource or multiple
        if any(param in params for param in ["perPage", "page", "list"]):
            context.extent = "multiple_resources"
        else:
            context.extent = "single_resource"

        # Add original request as metadata for reference
        context.metadata = {
            "original_github_request": json.dumps(github_request),
            "github_tool": tool_name,
            "mcp_version": self.current_version,  # Add protocol version information
        }

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
        result_content = (
            tnos_response.content if hasattr(tnos_response, "content") else {}
        )

        # If TNOS response is a string, try to parse as JSON
        if isinstance(result_content, str):
            try:
                result_content = json.loads(result_content)
            except json.JSONDecodeError:
                # If not valid JSON, wrap in a content field
                result_content = {"content": result_content}

        # Add the original context information as metadata
        response = {
            "result": result_content,
            "metadata": {
                "tnos_context": {
                    "who": tnos_response.context.who,
                    "what": tnos_response.context.what,
                    "when": tnos_response.context.when,
                    "where": tnos_response.context.where,
                    "why": tnos_response.context.why,
                    "how": tnos_response.context.how,
                    "extent": tnos_response.context.extent,
                },
                "mcp_version": self.current_version,  # Add protocol version information
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
            self.connection_state["tnos"]["last_connection_time"] = time.time()
            self.connection_state["tnos"]["reconnect_attempts"] = 0

            logger.info(
                f"Connected to TNOS MCP server at {
                    self.tnos_server_uri}"
            )

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
            # Create a version negotiation message
            negotiation_msg = {
                "message_type": "version_negotiation",
                "content": {
                    "supported_versions": self.supported_versions,
                    "preferred_version": self.current_version,
                },
            }

            # Send the negotiation message
            await self.tnos_ws_connection.send(json.dumps(negotiation_msg))

            # Wait for a response
            response_str = await asyncio.wait_for(
                self.tnos_ws_connection.recv(), timeout=5.0
            )

            # Parse the response
            response = json.loads(response_str)

            if response.get("message_type") == "version_negotiation_response":
                server_version = response.get("content", {}).get("selected_version")

                if server_version:
                    # Update the current version to the negotiated version
                    self.current_version = server_version
                    logger.info(
                        f"Negotiated MCP protocol version: {
                            self.current_version}"
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
        self.connection_state["github"]["last_connection_time"] = time.time()

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
                        f"Received GitHub MCP request: {
                            github_request.get(
                                'name', 'unknown')}"
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
                            f"Sent request to TNOS MCP for operation: {
                                tnos_context.what}"
                        )
                    except websockets.exceptions.ConnectionClosed:
                        # Connection to TNOS MCP server was closed, try to
                        # reconnect
                        logger.warning(
                            "Connection to TNOS MCP server was closed, attempting to reconnect..."
                        )

                        if await self.connect_to_tnos_mcp():
                            await self.tnos_ws_connection.send(tnos_message.to_json())
                        else:
                            error_msg = f"Failed to reconnect to TNOS MCP server: {
                                self.connection_state['tnos']['last_error']}"
                            logger.error(error_msg)
                            await websocket.send(
                                json.dumps(
                                    {
                                        "error": error_msg,
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
                            self.tnos_ws_connection.recv(),
                            timeout=30.0,  # 30 second timeout
                        )
                        tnos_response = MCPMessage.from_json(tnos_response_str)
                        logger.info(
                            f"Received response from TNOS MCP for operation: {
                                tnos_context.what}"
                        )
                    except asyncio.TimeoutError:
                        error_msg = f"Timeout waiting for response from TNOS MCP server for operation: {
                            tnos_context.what}"
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
                                    "recoverable": False,
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
                        f"Sent response back to GitHub MCP client for {
                            github_request.get(
                                'name', 'unknown')}"
                    )

                except json.JSONDecodeError:
                    error_msg = "Invalid JSON received from GitHub MCP"
                    logger.error(error_msg)
                    await websocket.send(
                        json.dumps(
                            {
                                "error": error_msg,
                                "status": "error",
                                "errorType": "json_parsing_failure",
                                "recoverable": True,
                            }
                        )
                    )
                except Exception as e:
                    error_msg = f"Error processing GitHub MCP request: {e}"
                    logger.error(error_msg)
                    logger.exception("Detailed exception information:")
                    await websocket.send(
                        json.dumps(
                            {
                                "error": error_msg,
                                "status": "error",
                                "errorType": "processing_failure",
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
                    try:
                        await self.tnos_ws_connection.close()
                        self.tnos_ws_connection = None
                        self.connection_state["tnos"]["connected"] = False
                        logger.info("Closed connection to TNOS MCP server")
                    except Exception as e:
                        logger.error(f"Error closing TNOS MCP connection: {e}")

    async def start(self):
        """
        Start the bridge server to handle connections from GitHub MCP.
        """
        try:
            # Check if the port is already in use
            if PortManager.check_port_in_use("localhost", self.github_mcp_port):
                logger.error(f"Port {self.github_mcp_port} is already in use")
                logger.error("Another instance of the bridge may be running")
                logger.error(
                    "Use the scripts/stop_tnos_github_integration.sh script to stop existing instances"
                )
                return

            server = await websockets.serve(
                self.handle_github_request, "localhost", self.github_mcp_port
            )
            logger.info(
                f"GitHub-TNOS MCP Bridge running on port {
                    self.github_mcp_port}"
            )

            # Keep the server running
            await server.wait_closed()
        except Exception as e:
            logger.error(f"Failed to start GitHub-TNOS MCP Bridge: {e}")


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

    # Create configuration for GitHub MCP
    config = {
        "toolDefinitions": tnos_capabilities,
        "serverSettings": {
            "bridgeUri": "ws://localhost:8080",
            "description": "TNOS capabilities exposed via GitHub MCP",
            "mcp_protocol_version": VersionManager.get_latest_version(),
        },
    }

    return config


async def main():
    """Start the GitHub-TNOS MCP bridge server."""
    # Generate tool registration config
    config = register_tnos_tools_with_github_mcp()

    # Save configuration using the predefined path constants
    config_manager = ConfigManager(path_manager)

    # Validate config before saving
    if not config_manager.validate_config(config):
        logger.error("Generated configuration is invalid")
        logger.error("This may indicate an issue with the tool registration process")
        sys.exit(1)

    # Save the configuration
    if not config_manager.save_config(config):
        logger.error("Failed to save configuration")
        logger.error("Check permissions on the config directory")
        sys.exit(1)

    # Check if TNOS MCP server is running
    if not PortManager.check_port_in_use("localhost", 8888):
        logger.warning("TNOS MCP server does not appear to be running on port 8888")
        logger.warning("The bridge will start, but connection to TNOS MCP may fail")
        logger.warning(
            "Start the TNOS MCP server with scripts/start_tnos_github_integration.sh"
        )

    # Create and start the bridge
    bridge = GitHubTNOSBridge()
    await bridge.start()


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("Bridge server stopped by user")
    except Exception as e:
        logger.error(f"Unexpected error: {e}")
        logger.exception("Detailed exception information:")
        sys.exit(1)
