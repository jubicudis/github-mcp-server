#!/usr/bin/env node

/**
 * MCPBridge.js
 * 
 * This script acts as a bridge between the GitHub MCP server and the TNOS MCP implementation.
 * It enables bidirectional communication and context sharing between the two systems,
 * allowing for seamless integration and interoperability.
 * 
 * This is a consolidated version that combines functionality from both bridge implementations.
 */

const http = require('http');
const WebSocket = require('ws');
const fs = require('fs');
const path = require('path');
const { exec } = require('child_process');
const fetch = require('node-fetch'); // Added for compression API support
const contextPersistence = require('../../src/utils/MCPContextPersistence.js');

// Configuration
const CONFIG = {
  // GitHub MCP server configuration
  githubMcp: {
    host: 'localhost',
    port: 10617,
    wsEndpoint: 'ws://localhost:10617/ws',
    apiEndpoint: 'http://localhost:10617/api'
  },
  // TNOS MCP server configuration (support both port configurations)
  tnosMcp: {
    host: 'localhost',
    port: process.env.TNOS_MCP_PORT || 8888, // Default to standard MCP port, can be overridden with env var
    wsEndpoint: `ws://localhost:${process.env.TNOS_MCP_PORT || 8888}/ws`,
    apiEndpoint: `http://localhost:${process.env.TNOS_MCP_PORT || 8888}/api`,
    altPort: 10618, // Alternative port from the original MCPBridge.js
    altWsEndpoint: 'ws://localhost:10618/ws',
    altApiEndpoint: 'http://localhost:10618/api'
  },
  // Bridge configuration
  bridge: {
    port: 10619,
    contextSyncInterval: 60000, // 1 minute
    healthCheckInterval: 30000, // 30 seconds
    reconnectAttempts: 5,
    reconnectDelay: 5000 // 5 seconds
  },
  // Logging configuration
  logging: {
    logDir: '/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/logs',
    logFile: 'mcp_bridge.log',
    logLevel: 'info' // debug, info, warn, error
  },
  // Formula registry path (imported from bridge.js)
  formulaRegistry: {
    path: '/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/config/formulas.json'
  }
};

// Initialize logging
const logLevels = {
  debug: 0,
  info: 1,
  warn: 2,
  error: 3
};

const logPath = path.join(CONFIG.logging.logDir, CONFIG.logging.logFile);

// Ensure log directory exists
if (!fs.existsSync(CONFIG.logging.logDir)) {
  fs.mkdirSync(CONFIG.logging.logDir, { recursive: true });
}

function log(level, message) {
  if (logLevels[level] >= logLevels[CONFIG.logging.logLevel]) {
    const timestamp = new Date().toISOString();
    const logEntry = `[${timestamp}] [${level.toUpperCase()}] ${message}`;

    // Log to console
    console.log(logEntry);

    // Log to file
    fs.appendFileSync(logPath, logEntry + '\n');
  }
}

// WebSocket connections
let githubMcpSocket = null;
let tnosMcpSocket = null;
let clientSockets = new Set();

// Context bridge for 7D context mapping
const contextBridge = require('../utils/7DContextBridge');

// Load formula registry (imported from bridge.js)
let formulas = {};
try {
  if (fs.existsSync(CONFIG.formulaRegistry.path)) {
    formulas = JSON.parse(fs.readFileSync(CONFIG.formulaRegistry.path, 'utf8'));
    log('info', `Loaded ${Object.keys(formulas).length} formulas from registry`);
  } else {
    log('warn', `Formula registry not found at ${CONFIG.formulaRegistry.path}`);
  }
} catch (error) {
  log('error', `Error loading formula registry: ${error.message}`);
}

// Add queue storage for messages during disconnection
const messageQueue = {
  github: [],
  tnos: [],

  // WHO: MessageQueue
  // WHAT: Queue message for GitHub MCP server
  // WHEN: During connection interruptions
  // WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
  // WHY: To prevent message loss during disconnections
  // HOW: Using persistent array queue with disk backup
  // EXTENT: All GitHub-bound messages
  queueForGithub(message) {
    this.githubMCP.push({
      message: message,
      timestamp: Date.now()
    });
    this._persistQueues();
    log('info', `Message queued for GitHub MCP (queue size: ${this.github.length})`);
  },

  // WHO: MessageQueue
  // WHAT: Queue message for TNOS MCP server
  // WHEN: During connection interruptions
  // WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
  // WHY: To prevent message loss during disconnections
  // HOW: Using persistent array queue with disk backup
  // EXTENT: All TNOS-bound messages
  queueForTnos(message) {
    this.tnos.push({
      message: message,
      timestamp: Date.now()
    });
    this._persistQueues();
    log('info', `Message queued for TNOS MCP (queue size: ${this.tnos.length})`);
  },

  // WHO: MessageQueue
  // WHAT: Process queued messages for GitHub
  // WHEN: After connection restoration
  // WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
  // WHY: To send stored messages after reconnection
  // HOW: Using FIFO queue processing with age limits
  // EXTENT: All queued GitHub messages
  processGithubQueue() {
    if (this.github.length === 0) return;

    log('info', `Processing ${this.github.length} queued messages for GitHub MCP`);

    // Filter out messages that are too old (over 1 hour)
    const now = Date.now();
    const maxAge = 60 * 60 * 1000; // 1 hour
    this.github = this.githubMCP.filter(item => (now - item.timestamp) <= maxAge);

    // Process remaining messages
    if (githubMcpSocket && githubMcpSocket.readyState === WebSocket.OPEN) {
      // Take a copy of the queue and clear it before processing
      // This prevents race conditions if processing is interrupted
      const messagesToProcess = [...this.github];
      this.github = [];
      this._persistQueues();

      messagesToProcess.forEach((item, index) => {
        try {
          githubMCP.send(JSON.stringify(item.message));
          log('debug', `Sent queued message ${index + 1}/${messagesToProcess.length} to GitHub MCP`);
        } catch (error) {
          log('error', `Error sending queued message to GitHub MCP: ${error.message}`);
          // Re-queue the failed message
          this.queueForGithub(item.message);
        }
      });

      log('info', `Finished processing GitHub MCP message queue`);
    }
  },

  // WHO: MessageQueue
  // WHAT: Process queued messages for TNOS
  // WHEN: After connection restoration
  // WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
  // WHY: To send stored messages after reconnection
  // HOW: Using FIFO queue processing with age limits
  // EXTENT: All queued TNOS messages
  processTnosQueue() {
    if (this.tnos.length === 0) return;

    log('info', `Processing ${this.tnos.length} queued messages for TNOS MCP`);

    // Filter out messages that are too old (over 1 hour)
    const now = Date.now();
    const maxAge = 60 * 60 * 1000; // 1 hour
    this.tnos = this.tnos.filter(item => (now - item.timestamp) <= maxAge);

    // Process remaining messages
    if (tnosMcpSocket && tnosMcpSocket.readyState === WebSocket.OPEN) {
      // Take a copy of the queue and clear it before processing
      const messagesToProcess = [...this.tnos];
      this.tnos = [];
      this._persistQueues();

      messagesToProcess.forEach((item, index) => {
        try {
          tnosMcpSocket.send(JSON.stringify(item.message));
          log('debug', `Sent queued message ${index + 1}/${messagesToProcess.length} to TNOS MCP`);
        } catch (error) {
          log('error', `Error sending queued message to TNOS MCP: ${error.message}`);
          // Re-queue the failed message
          this.queueForTnos(item.message);
        }
      });

      log('info', `Finished processing TNOS MCP message queue`);
    }
  },

  // WHO: MessageQueue
  // WHAT: Persist message queues to disk
  // WHEN: After queue modifications
  // WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
  // WHY: To ensure message durability across restarts
  // HOW: Using filesystem storage with JSON serialization
  // EXTENT: All queued messages
  _persistQueues() {
    try {
      const queueData = {
        github: this.github,
        tnos: this.tnos,
        timestamp: Date.now()
      };

      const queueFile = path.join(CONFIG.logging.logDir, 'mcp_message_queue.json');
      fs.writeFileSync(queueFile, JSON.stringify(queueData, null, 2));
    } catch (error) {
      log('error', `Failed to persist message queues: ${error.message}`);
    }
  },

  // WHO: MessageQueue
  // WHAT: Load message queues from disk
  // WHEN: During bridge initialization
  // WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
  // WHY: To recover messages after system restart
  // HOW: Using filesystem storage with JSON parsing
  // EXTENT: Previously stored message queues
  loadQueues() {
    try {
      const queueFile = path.join(CONFIG.logging.logDir, 'mcp_message_queue.json');

      if (fs.existsSync(queueFile)) {
        const queueData = JSON.parse(fs.readFileSync(queueFile, 'utf8'));
        this.github = queueData.github || [];
        this.tnos = queueData.tnos || [];

        // Filter out messages that are too old (over 1 hour)
        const now = Date.now();
        const maxAge = 60 * 60 * 1000; // 1 hour
        this.github = this.githubMCP.filter(item => (now - item.timestamp) <= maxAge);
        this.tnos = this.tnos.filter(item => (now - item.timestamp) <= maxAge);

        log('info', `Loaded message queues: GitHub (${this.github.length}), TNOS (${this.tnos.length})`);

        // Persist filtered queues
        this._persistQueues();
      }
    } catch (error) {
      log('error', `Failed to load message queues: ${error.message}`);
      // Initialize empty queues on error
      this.github = [];
      this.tnos = [];
    }
  }
};

// Helper function to check if a server is running
function checkServerRunning(host, port) {
  return new Promise((resolve) => {
    const connection = http.get({
      host: host,
      port: port,
      path: '/',
      timeout: 3000
    }, (res) => {
      resolve(true);
      connection.destroy();
    });

    connection.on('error', () => {
      resolve(false);
      connection.destroy();
    });

    connection.on('timeout', () => {
      resolve(false);
      connection.destroy();
    });
  });
}

// Function to start servers if they're not running
async function ensureServersRunning() {
  log('info', 'Checking if MCP servers are running...');

  const githubMcpRunning = await checkServerRunning(CONFIG.githubMcp.host, CONFIG.githubMcp.port);
  const tnosMcpRunning = await checkServerRunning(CONFIG.tnosMcp.host, CONFIG.tnosMcp.port);
  const tnosMcpAltRunning = await checkServerRunning(CONFIG.tnosMcp.host, CONFIG.tnosMcp.altPort);

  if (!githubMcpRunning || (!tnosMcpRunning && !tnosMcpAltRunning)) {
    log('info', 'Starting MCP servers...');

    return new Promise((resolve) => {
      const startScript = path.join(__dirname, '../../scripts/start_mcp_servers.sh');

      exec(`bash ${startScript}`, (error, stdout, stderr) => {
        if (error) {
          log('error', `Failed to start MCP servers: ${error.message}`);
          log('debug', `stdout: ${stdout}`);
          log('debug', `stderr: ${stderr}`);
          resolve(false);
        } else {
          log('info', 'MCP servers started successfully');
          resolve(true);
        }
      });
    });
  }

  log('info', 'MCP servers are running');
  return true;
}

// Backoff strategies for reconnection
const backoffStrategies = {
  github: {
    attempts: 0,
    maxAttempts: Infinity, // Never stop trying to reconnect
    baseDelay: 1000,
    maxDelay: 30000,
    getNextBackoffDelay() {
      const delay = Math.min(this.baseDelay * Math.pow(2, this.attempts), this.maxDelay);
      this.attempts++;
      return delay;
    },
    resetBackoff() {
      this.attempts = 0;
    }
  },
  tnos: {
    attempts: 0,
    maxAttempts: Infinity, // Never stop trying to reconnect
    baseDelay: 1000,
    maxDelay: 30000,
    getNextBackoffDelay() {
      const delay = Math.min(this.baseDelay * Math.pow(2, this.attempts), this.maxDelay);
      this.attempts++;
      return delay;
    },
    resetBackoff() {
      this.attempts = 0;
    }
  }
};

// Flag to indicate if we're shutting down - prevent reconnection attempts during shutdown
let shuttingDown = false;

// Connect to GitHub MCP server
function connectToGithubMcp() {
  log('info', 'Connecting to GitHub MCP server...');

  // Clean up existing socket if any
  if (githubMcpSocket) {
    try {
      githubMCP.terminate();
    } catch (error) {
      // Ignore errors during cleanup
    }
    githubMcpSocket = null;
  }

  githubMcpSocket = new WebSocket(CONFIG.githubMcp.wsEndpoint);

  githubMCP.on('open', () => {
    log('info', 'Connected to GitHub MCP server');
    // Reset the backoff counter on successful connection
    backoffStrategies.githubMCP.resetBackoff();
    // Process queued messages
    messageQueue.processGithubQueue();
  });

  githubMCP.on('message', (data) => {
    try {
      const message = JSON.parse(data);
      log('debug', `Received message from GitHub MCP: ${JSON.stringify(message)}`);

      // Process message from GitHub MCP
      processGithubMcpMessage(message);
    } catch (error) {
      log('error', `Error processing message from GitHub MCP: ${error.message}`);
    }
  });

  githubMCP.on('error', (error) => {
    log('error', `GitHub MCP WebSocket error: ${error.message}`);
  });

  githubMCP.on('close', (code, reason) => {
    log('warn', `GitHub MCP WebSocket connection closed: Code ${code}${reason ? ', Reason: ' + reason : ''}`);

    // Try to reconnect using exponential backoff
    const nextDelay = backoffStrategies.githubMCP.getNextBackoffDelay();
    log('info', `Will attempt to reconnect to GitHub MCP server in ${nextDelay}ms (attempt ${backoffStrategies.github.attempts})`);

    setTimeout(() => {
      if (shuttingDown) return;
      log('info', 'Attempting to reconnect to GitHub MCP server...');
      connectToGithubMcp();
    }, nextDelay);
  });
}

// Connect to TNOS MCP server
function connectToTnosMcp() {
  log('info', 'Connecting to TNOS MCP server...');

  // Clean up existing socket if any
  if (tnosMcpSocket) {
    try {
      tnosMcpSocket.terminate();
    } catch (error) {
      // Ignore errors during cleanup
    }
    tnosMcpSocket = null;
  }

  // First try the main port
  const wsEndpoint = `ws://${CONFIG.tnosMcp.host}:${CONFIG.tnosMcp.port}/ws`;
  log('debug', `Attempting to connect to TNOS MCP at ${wsEndpoint}`);

  tnosMcpSocket = new WebSocket(wsEndpoint);

  tnosMcpSocket.on('open', () => {
    log('info', `Connected to TNOS MCP server on port ${CONFIG.tnosMcp.port}`);
    // Reset the backoff counter on successful connection
    backoffStrategies.tnos.resetBackoff();
    // Process queued messages
    messageQueue.processTnosQueue();
  });

  tnosMcpSocket.on('error', (error) => {
    log('warn', `TNOS MCP WebSocket error on port ${CONFIG.tnosMcp.port}: ${error.message}`);

    // If connection fails on primary port, try the alternate port
    if (backoffStrategies.tnos.attempts === 0) {
      log('info', `Attempting to connect to alternate TNOS MCP port ${CONFIG.tnosMcp.altPort}`);

      tnosMcpSocket = new WebSocket(CONFIG.tnosMcp.altWsEndpoint);

      tnosMcpSocket.on('open', () => {
        log('info', `Connected to TNOS MCP server on alternate port ${CONFIG.tnosMcp.altPort}`);
        // Reset the backoff counter on successful connection
        backoffStrategies.tnos.resetBackoff();
        // Process queued messages
        messageQueue.processTnosQueue();
      });

      setupTnosMcpEventHandlers();
    }
  });

  setupTnosMcpEventHandlers();
}

// Set up event handlers for TNOS MCP socket
function setupTnosMcpEventHandlers() {
  if (!tnosMcpSocket) return;

  tnosMcpSocket.on('message', (data) => {
    try {
      const message = JSON.parse(data);
      log('debug', `Received message from TNOS MCP: ${JSON.stringify(message)}`);

      // Process message from TNOS MCP
      processTnosMcpMessage(message);
    } catch (error) {
      log('error', `Error processing message from TNOS MCP: ${error.message}`);
    }
  });

  tnosMcpSocket.on('close', (code, reason) => {
    log('warn', `TNOS MCP WebSocket connection closed: Code ${code}${reason ? ', Reason: ' + reason : ''}`);

    // Try to reconnect using exponential backoff
    const nextDelay = backoffStrategies.tnos.getNextBackoffDelay();
    log('info', `Will attempt to reconnect to TNOS MCP server in ${nextDelay}ms (attempt ${backoffStrategies.tnos.attempts})`);

    setTimeout(() => {
      if (shuttingDown) return;
      log('info', 'Attempting to reconnect to TNOS MCP server...');
      connectToTnosMcp();
    }, nextDelay);
  });
}

/**
 * Execute a formula from the TNOS Formula Registry
 * @param {Object} params - Formula execution parameters
 * @returns {Promise<Object>} - Formula execution result
 * 
 * # WHO: MCPBridge.executeFormula
 * # WHAT: Execute formulas from the TNOS Formula Registry
 * # WHEN: When formula execution is requested through MCP
 * # WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
 * # WHY: To provide formula execution capabilities
 * # HOW: Using formula registry and TNOS MCP
 * # EXTENT: All registered formulas
 */
async function executeFormula(params) {
  const { formulaName, parameters } = params;

  log('info', `Executing formula: ${formulaName}`);

  // Check if formula exists in registry
  if (!formulas[formulaName]) {
    log('warn', `Formula not found: ${formulaName}`);
    return {
      error: `Formula not found: ${formulaName}`,
      success: false
    };
  }

  try {
    // Send formula execution request to TNOS MCP
    if (tnosMcpSocket && tnosMcpSocket.readyState === WebSocket.OPEN) {
      return new Promise((resolve) => {
        const requestId = `formula-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;

        // Set up one-time handler for the response
        const responseHandler = (data) => {
          try {
            const message = JSON.parse(data);
            if (message.requestId === requestId) {
              tnosMcpSocket.removeEventListener('message', responseHandler);
              resolve({
                success: true,
                result: message.result,
                timestamp: new Date().toISOString()
              });
            }
          } catch (error) {
            // Ignore errors in this handler
          }
        };

        // Add temporary listener for this specific response
        tnosMcpSocket.addEventListener('message', responseHandler);

        // Send the request
        tnosMcpSocket.send(JSON.stringify({
          type: 'execute_formula',
          requestId: requestId,
          data: {
            formulaName,
            parameters: parameters || {}
          }
        }));

        // Set timeout to prevent hanging
        setTimeout(() => {
          tnosMcpSocket.removeEventListener('message', responseHandler);
          resolve({
            success: true,
            result: `Formula ${formulaName} execution initiated, but response timed out`,
            timestamp: new Date().toISOString()
          });
        }, 10000); // 10 second timeout
      });
    } else {
      // Queue formula execution for later if TNOS MCP is not connected
      messageQueue.queueForTnos({
        type: 'execute_formula',
        data: {
          formulaName,
          parameters: parameters || {}
        }
      });

      // Return placeholder result
      return {
        success: true,
        result: `Formula ${formulaName} execution queued for later processing`,
        timestamp: new Date().toISOString()
      };
    }
  } catch (error) {
    log('error', `Error executing formula: ${error.message}`);
    return {
      error: `Error executing formula: ${error.message}`,
      success: false
    };
  }
}

/**
 * Query the TNOS 7D context system
 * @param {Object} params - Query parameters
 * @returns {Promise<Object>} - Query result
 * 
 * # WHO: MCPBridge.queryDimensionalContext
 * # WHAT: Query the TNOS 7D context system
 * # WHEN: When dimensional context queries are made
 * # WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
 * # WHY: To provide access to 7D context information
 * # HOW: Using context bridge with dimensional mapping
 * # EXTENT: All 7D context dimensions
 */
async function queryDimensionalContext(params) {
  const { dimension, query, recursiveDepth = 3 } = params;

  log('info', `Querying dimensional context: ${dimension} for "${query}"`);

  try {
    // Use the 7D Context Bridge to query the context
    const result = await contextBridge.queryDimensionalContext(dimension, query, recursiveDepth);
    return result;
  } catch (error) {
    log('error', `Error querying dimensional context: ${error.message}`);
    return {
      error: `Error querying dimensional context: ${error.message}`,
      success: false
    };
  }
}

/**
 * Perform Möbius compression on data
 * @param {Object} params - Compression parameters
 * @returns {Promise<Object>} - Compression result
 * 
 * # WHO: MCPBridge.performMobiusCompression
 * # WHAT: Möbius compression
 * # WHEN: When compression is requested through the MCP bridge
 * # WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
 * # WHY: To optimize data with compression-first approach
 * # HOW: Using dimensional compression with 7D context
 * # EXTENT: All data passing through MCP
 */
async function performMobiusCompression(params) {
  const { inputData, useTimeFactor = true, useEnergyFactor = true, context = {} } = params;

  log('info', `Performing Möbius compression on ${inputData ? (typeof inputData === 'string' ? inputData.length : JSON.stringify(inputData).length) : 0} bytes of data`);

  try {
    // Convert to standard compression request format
    const compressionRequest = {
      operation: 'compress',
      data: inputData,
      context: {
        who: context.who || 'MCPBridge',
        what: context.what || 'DataCompression',
        when: context.when || Date.now(),
        where: /Users/Jubicudis / TNOS1 / Tranquility - Neuro - OS / mcp
        why: context.why || 'OptimizeDataTransfer',
        how: context.how || 'MobiusFormula',
        extent: context.extent || 1.0
      },
      target_layer: 3, // Default to Python layer
      session_id: context.sessionId || `mcp-bridge-${Date.now()}`
    };

    // Add compression config if provided
    if (useTimeFactor !== undefined || useEnergyFactor !== undefined) {
      compressionRequest.compression_config = {
        time_factor: useTimeFactor ? calculateTimeFactor() : 1.0,
        energy_factor: useEnergyFactor ? calculateEnergyFactor(inputData) : 1.0,
        preserve_structure: true
      };
    }

    // Send compression request to the MCP server's dedicated compression handler
    // Try both port configurations for compression endpoint
    let response;
    try {
      response = await fetch(`http://${CONFIG.tnosMcp.host}:${CONFIG.tnosMcp.port}/compression`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(compressionRequest)
      }).then(res => res.json());
    } catch (error) {
      // Try alternate port if main port fails
      log('warn', `Compression request failed on port ${CONFIG.tnosMcp.port}, trying alternate port ${CONFIG.tnosMcp.altPort}`);
      response = await fetch(`http://${CONFIG.tnosMcp.host}:${CONFIG.tnosMcp.altPort}/compression`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(compressionRequest)
      }).then(res => res.json());
    }

    // If the request failed, throw an error
    if (response.status === 'error') {
      throw new Error(response.error || 'Unknown compression error');
    }

    // Return the compression result in the expected format
    return {
      success: true,
      originalSize: response.metadata?.original_size || 0,
      compressedSize: response.metadata?.compressed_size || 0,
      compressionRatio: response.metadata?.compression_ratio || 1.0,
      timeFactor: response.metadata?.time_factor || 1.0,
      energyFactor: response.metadata?.energy_factor || 1.0,
      data: response.data,
      metadata: response.metadata,
      timestamp: new Date().toISOString()
    };
  } catch (error) {
    log('error', `Error performing Möbius compression: ${error.message}`);

    // Fall back to basic compression simulation if the server request fails
    // This ensures the compression-first approach continues to work even if the 
    // dedicated compression service is unavailable
    const compressionRatio = 0.7; // Typical ratio for Möbius compression
    return {
      success: false,
      originalSize: inputData ? (typeof inputData === 'string' ? inputData.length : JSON.stringify(inputData).length) : 0,
      compressedSize: inputData ? Math.floor((typeof inputData === 'string' ? inputData.length : JSON.stringify(inputData).length) * compressionRatio) : 0,
      compressionRatio,
      error: error.message,
      timestamp: new Date().toISOString()
    };
  }
}

/**
 * Calculate time factor for Möbius compression
 * @returns {number} - Time factor value between 0.5 and 1.0
 * 
 * # WHO: MCPBridge.calculateTimeFactor
 * # WHAT: Calculate time factor for compression
 * # WHEN: During Möbius compression
 * # WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
 * # WHY: To optimize compression based on temporal context
 * # HOW: Using temporal wave function calculation
 * # EXTENT: All compression operations
 */
function calculateTimeFactor() {
  // Use the fraction of the current second as a basis (0.0-1.0)
  const now = new Date();
  const baseFactor = now.getMilliseconds() / 1000;

  // Apply 7D time factor formula: 0.5 + (value * 0.5)
  // This ensures the time factor is between 0.5 and 1.0
  return 0.5 + (baseFactor * 0.5);
}

/**
 * Calculate energy factor for Möbius compression based on data entropy
 * @param {*} data - The data to calculate energy factor for
 * @returns {number} - Energy factor value between 0.6 and 0.9
 * 
 * # WHO: MCPBridge.calculateEnergyFactor
 * # WHAT: Calculate energy factor based on data entropy
 * # WHEN: During Möbius compression
 * # WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
 * # WHY: To optimize compression efficiency based on data structure
 * # HOW: Using Shannon entropy calculation
 * # EXTENT: All compression operations
 */
function calculateEnergyFactor(data) {
  if (!data) return 0.8; // Default

  try {
    // Convert data to string for entropy calculation
    const str = typeof data === 'string' ? data : JSON.stringify(data);

    // Calculate Shannon entropy
    const len = str.length;
    const charCounts = {};

    // Count character frequencies
    for (let i = 0; i < len; i++) {
      charCounts[str[i]] = (charCounts[str[i]] || 0) + 1;
    }

    // Calculate entropy
    let entropy = 0;
    for (const char in charCounts) {
      const freq = charCounts[char] / len;
      entropy -= freq * Math.log2(freq);
    }

    // Normalize entropy to energy factor (0.6-0.9)
    // Higher entropy (more randomness) = lower compression benefit
    const maxEntropy = 5; // Typical maximum for text data
    const normalizedEntropy = Math.min(entropy, maxEntropy) / maxEntropy;
    return 0.9 - (normalizedEntropy * 0.3);
  } catch (e) {
    return 0.8; // Default on error
  }
}

// Process message from GitHub MCP
function processGithubMcpMessage(message) {
  // WHO: MCPBridge.processGithubMcpMessage
  // WHAT: Process incoming messages from GitHub MCP
  // WHEN: When messages arrive from GitHub MCP server
  // WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
  // WHY: To transform and forward GitHub messages to TNOS
  // HOW: Using context transformation and reliable delivery
  // EXTENT: All GitHub-to-TNOS message traffic

  // Check if the message is a formula execution request
  if (message.type === 'execute_formula' || (message.path && message.path.includes('formula'))) {
    const formulaParams = message.data || message.parameters || {};

    // Execute the formula asynchronously
    executeFormula(formulaParams).then(result => {
      // Send the result back to GitHub MCP
      if (githubMcpSocket && githubMcpSocket.readyState === WebSocket.OPEN) {
        githubMCP.send(JSON.stringify({
          type: 'formula_result',
          requestId: message.requestId,
          data: result
        }));
      }
    }).catch(error => {
      log('error', `Error executing formula: ${error.message}`);
    });
  }

  // Check if the message is a compression request
  else if (message.type === 'compression' || (message.path && message.path.includes('compression'))) {
    const compressionParams = message.data || message.parameters || {};

    // Perform compression asynchronously
    performMobiusCompression(compressionParams).then(result => {
      // Send the result back to GitHub MCP
      if (githubMcpSocket && githubMcpSocket.readyState === WebSocket.OPEN) {
        githubMCP.send(JSON.stringify({
          type: 'compression_result',
          requestId: message.requestId,
          data: result
        }));
      }
    }).catch(error => {
      log('error', `Error performing compression: ${error.message}`);
    });
  }

  // Save context data to persistence
  else if (message.type === 'context' || (message.path && message.path === '/context')) {
    contextPersistence.mergeContext(message.data, 'github');
  }

  // Transform GitHub MCP message to TNOS 7D format
  const tnos7dMessage = contextBridge.githubToTnos7d(message);

  // Forward to TNOS MCP with reliability
  if (tnosMcpSocket && tnosMcpSocket.readyState === WebSocket.OPEN) {
    try {
      tnosMcpSocket.send(JSON.stringify(tnos7dMessage));
      log('debug', `Forwarded message to TNOS MCP: ${JSON.stringify(tnos7dMessage)}`);
    } catch (error) {
      log('error', `Error forwarding message to TNOS MCP: ${error.message}`);
      // Queue message for later processing on error
      messageQueue.queueForTnos(tnos7dMessage);
    }
  } else {
    log('warn', 'TNOS MCP socket not ready, queuing message for later delivery');
    // Queue message for later processing
    messageQueue.queueForTnos(tnos7dMessage);
  }

  // Forward to connected clients
  broadcastToClients(message);
}

// Process message from TNOS MCP
function processTnosMcpMessage(message) {
  // WHO: MCPBridge.processTnosMcpMessage
  // WHAT: Process incoming messages from TNOS MCP
  // WHEN: When messages arrive from TNOS MCP server
  // WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
  // WHY: To transform and forward TNOS messages to GitHub
  // HOW: Using context transformation and reliable delivery
  // EXTENT: All TNOS-to-GitHub message traffic

  // Save context data to persistence
  if (message.type === 'context' || (message.path && message.path === '/context')) {
    contextPersistence.mergeContext(message.data, 'tnos7d');
  }

  // Transform TNOS 7D message to GitHub MCP format
  const githubMessage = contextBridge.tnos7dToGithub(message);

  // Forward to GitHub MCP with reliability
  if (githubMcpSocket && githubMcpSocket.readyState === WebSocket.OPEN) {
    try {
      githubMCP.send(JSON.stringify(githubMessage));
      log('debug', `Forwarded message to GitHub MCP: ${JSON.stringify(githubMessage)}`);
    } catch (error) {
      log('error', `Error forwarding message to GitHub MCP: ${error.message}`);
      // Queue message for later processing on error
      messageQueue.queueForGithub(githubMessage);
    }
  } else {
    log('warn', 'GitHub MCP socket not ready, queuing message for later delivery');
    // Queue message for later processing
    messageQueue.queueForGithub(githubMessage);
  }

  // Forward to clients
  broadcastToClients(message);
}

// Broadcast message to all connected clients
function broadcastToClients(message) {
  clientSockets.forEach((client) => {
    if (client.readyState === WebSocket.OPEN) {
      client.send(JSON.stringify(message));
    }
  });
}

// Sync context between the two servers
async function syncContext() {
  try {
    log('info', 'Syncing context between MCP servers...');

    // Load persisted contexts
    const githubContext = await contextPersistence.loadContext('github');
    const tnos7dContext = await contextPersistence.loadContext('tnos7d');

    // Transform GitHub context to TNOS 7D format
    const tnos7dFormatted = contextBridge.githubContextToTnos7d(githubContext);

    // Transform TNOS 7D context to GitHub format
    const githubFormatted = contextBridge.tnos7dContextToGithub(tnos7dContext);

    // Send context to respective servers
    if (tnosMcpSocket && tnosMcpSocket.readyState === WebSocket.OPEN) {
      tnosMcpSocket.send(JSON.stringify({
        type: 'context',
        data: tnos7dFormatted
      }));
      log('debug', 'Sent GitHub context to TNOS MCP');
    }

    if (githubMcpSocket && githubMcpSocket.readyState === WebSocket.OPEN) {
      githubMCP.send(JSON.stringify({
        type: 'context',
        data: githubFormatted
      }));
      log('debug', 'Sent TNOS 7D context to GitHub MCP');
    }

    log('info', 'Context sync complete');
  } catch (error) {
    log('error', `Error syncing context: ${error.message}`);
  }
}

// Perform health check on MCP servers
async function healthCheck() {
  try {
    log('debug', 'Performing health check on MCP servers...');

    const githubMcpRunning = await checkServerRunning(CONFIG.githubMcp.host, CONFIG.githubMcp.port);
    const tnosMcpRunning = await checkServerRunning(CONFIG.tnosMcp.host, CONFIG.tnosMcp.port);
    const tnosMcpAltRunning = await checkServerRunning(CONFIG.tnosMcp.host, CONFIG.tnosMcp.altPort);

    if (!githubMcpRunning) {
      log('warn', 'GitHub MCP server is not responding');

      // Try to reconnect WebSocket
      if (githubMcpSocket) {
        githubMCP.terminate();
      }
      connectToGithubMcp();
    }

    if (!tnosMcpRunning && !tnosMcpAltRunning) {
      log('warn', 'TNOS MCP server is not responding on any port');

      // Try to reconnect WebSocket
      if (tnosMcpSocket) {
        tnosMcpSocket.terminate();
      }
      connectToTnosMcp();
    }

    if (!githubMcpRunning || (!tnosMcpRunning && !tnosMcpAltRunning)) {
      log('info', 'Attempting to restart MCP servers...');
      await ensureServersRunning();
    }
  } catch (error) {
    log('error', `Error performing health check: ${error.message}`);
  }
}

// Start the bridge server
async function startBridge() {
  try {
    // WHO: MCPBridge.startBridge
    // WHAT: Initialize and start MCP bridge server
    // WHEN: System startup or manual restart
    // WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
    // WHY: To establish reliable bidirectional communication
    // HOW: Using WebSockets with resilient error handling
    // EXTENT: All cross-system communication

    // Ensure MCP servers are running
    const serversRunning = await ensureServersRunning();

    if (!serversRunning) {
      log('error', 'Failed to start MCP servers, bridge cannot start');

      // Instead of exiting, schedule a retry
      setTimeout(() => {
        log('info', 'Retrying bridge startup...');
        startBridge();
      }, 10000); // Try again in 10 seconds
      return;
    }

    // Initialize context persistence with retries
    let contextInitialized = false;
    for (let attempt = 1; attempt <= 3; attempt++) {
      try {
        await contextPersistence.initialize();
        contextInitialized = true;
        break;
      } catch (error) {
        log('warn', `Context persistence initialization attempt ${attempt} failed: ${error.message}`);
        if (attempt < 3) await new Promise(resolve => setTimeout(resolve, 2000));
      }
    }

    if (!contextInitialized) {
      log('warn', 'Context persistence initialization failed, continuing with limited functionality');
    }

    // Load message queues from disk
    messageQueue.loadQueues();

    // Connect to MCP servers
    connectToGithubMcp();
    connectToTnosMcp();

    // Start WebSocket server for the bridge
    const server = http.createServer((req, res) => {
      // Simple health check endpoint
      if (req.url === '/health') {
        const health = {
          status: 'running',
          githubConnection: githubMcpSocket && githubMcpSocket.readyState === WebSocket.OPEN ? 'connected' : 'disconnected',
          tnosConnection: tnosMcpSocket && tnosMcpSocket.readyState === WebSocket.OPEN ? 'connected' : 'disconnected',
          queuedMessages: {
            github: messageQueue.github.length,
            tnos: messageQueue.tnos.length
          },
          formulas: Object.keys(formulas).length,
          uptime: process.uptime(),
          timestamp: Date.now()
        };
        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify(health));
        return;
      }

      // Default response
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ status: 'MCP Bridge running' }));
    });

    // Add error handler for HTTP server
    server.on('error', (error) => {
      log('error', `HTTP server error: ${error.message}`);
      // Attempt to restart server on port error
      if (error.code === 'EADDRINUSE') {
        log('info', `Port ${CONFIG.bridge.port} is in use, attempting to restart in 5 seconds`);
        setTimeout(() => {
          try {
            server.close();
            server.listen(CONFIG.bridge.port);
          } catch (e) {
            log('error', `Failed to restart HTTP server: ${e.message}`);
          }
        }, 5000);
      }
    });

    const wss = new WebSocket.Server({ server });

    wss.on('connection', (ws) => {
      log('info', 'Client connected to bridge');
      clientSockets.add(ws);

      ws.on('message', (data) => {
        try {
          const message = JSON.parse(data);
          log('debug', `Received message from client: ${JSON.stringify(message)}`);

          // Handle special formula execution requests directly
          if (message.type === 'execute_formula' || (message.target === 'tnos' && message.data && message.data.type === 'execute_formula')) {
            const formulaParams = message.data || message.parameters || {};
            executeFormula(formulaParams).then(result => {
              ws.send(JSON.stringify({
                type: 'formula_result',
                requestId: message.requestId,
                data: result
              }));
            }).catch(error => {
              log('error', `Error executing formula: ${error.message}`);
            });
            return;
          }

          // Handle compression requests directly
          if (message.type === 'compression' || (message.target === 'tnos' && message.data && message.data.type === 'compression')) {
            const compressionParams = message.data || message.parameters || {};
            performMobiusCompression(compressionParams).then(result => {
              ws.send(JSON.stringify({
                type: 'compression_result',
                requestId: message.requestId,
                data: result
              }));
            }).catch(error => {
              log('error', `Error performing compression: ${error.message}`);
            });
            return;
          }

          // Determine target server and forward message
          if (message.target === 'github') {
            if (githubMcpSocket && githubMcpSocket.readyState === WebSocket.OPEN) {
              githubMCP.send(JSON.stringify(message.data));
              log('debug', `Forwarded message to GitHub MCP: ${JSON.stringify(message.data)}`);
            } else {
              // Queue message for later processing
              messageQueue.queueForGithub(message.data);
            }
          } else if (message.target === 'tnos') {
            if (tnosMcpSocket && tnosMcpSocket.readyState === WebSocket.OPEN) {
              tnosMcpSocket.send(JSON.stringify(message.data));
              log('debug', `Forwarded message to TNOS MCP: ${JSON.stringify(message.data)}`);
            } else {
              // Queue message for later processing
              messageQueue.queueForTnos(message.data);
            }
          } else {
            // Broadcast to both servers
            if (githubMcpSocket && githubMcpSocket.readyState === WebSocket.OPEN) {
              githubMCP.send(JSON.stringify(message.data));
            } else {
              // Queue message for later processing
              messageQueue.queueForGithub(message.data);
            }
            if (tnosMcpSocket && tnosMcpSocket.readyState === WebSocket.OPEN) {
              tnosMcpSocket.send(JSON.stringify(message.data));
            } else {
              // Queue message for later processing
              messageQueue.queueForTnos(message.data);
            }
          }
        } catch (error) {
          log('error', `Error processing client message: ${error.message}`);
        }
      });

      ws.on('close', () => {
        log('info', 'Client disconnected from bridge');
        clientSockets.delete(ws);
      });

      ws.on('error', (error) => {
        log('error', `WebSocket client error: ${error.message}`);
        // Remove problematic clients
        clientSockets.delete(ws);
      });
    });

    // Handle WebSocket server errors
    wss.on('error', (error) => {
      log('error', `WebSocket server error: ${error.message}`);
    });

    // Start HTTP server
    server.listen(CONFIG.bridge.port, () => {
      log('info', `MCP Bridge running on port ${CONFIG.bridge.port}`);
    });

    // Set up periodic context sync with error handling
    const contextSyncInterval = setInterval(() => {
      syncContext().catch(error => {
        log('error', `Error in scheduled context sync: ${error.message}`);
      });
    }, CONFIG.bridge.contextSyncInterval);

    // Set up periodic health check with error handling
    const healthCheckInterval = setInterval(() => {
      healthCheck().catch(error => {
        log('error', `Error in scheduled health check: ${error.message}`);
      });
    }, CONFIG.bridge.healthCheckInterval);

    // Perform initial context sync after a short delay
    setTimeout(() => {
      syncContext().catch(error => {
        log('error', `Error in initial context sync: ${error.message}`);
      });
    }, 5000);

    // Handle process shutdown
    process.on('SIGINT', shutdown);
    process.on('SIGTERM', shutdown);

    // Handle uncaught exceptions to prevent bridge from crashing
    process.on('uncaughtException', (error) => {
      log('error', `Uncaught exception: ${error.message}`);
      log('error', error.stack);
      // Don't exit process - try to keep the bridge running
    });

    // Handle unhandled promise rejections
    process.on('unhandledRejection', (reason, promise) => {
      log('error', `Unhandled promise rejection: ${reason}`);
      // Don't exit process - try to keep the bridge running
    });

    log('info', 'MCP Bridge started successfully');
  } catch (error) {
    log('error', `Error starting MCP Bridge: ${error.message}`);
    // Instead of exiting, schedule a retry
    setTimeout(() => {
      log('info', 'Retrying bridge startup...');
      startBridge();
    }, 10000); // Try again in 10 seconds
  }
}

// Graceful shutdown
async function shutdown() {
  // WHO: MCPBridge.shutdown
  // WHAT: Perform graceful shutdown of MCP Bridge
  // WHEN: During system termination or manual restart
  // WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
  // WHY: To ensure clean termination without data loss
  // HOW: Using orderly connection closure and persistence
  // EXTENT: All Bridge components and connections

  log('info', 'Shutting down MCP Bridge...');

  // Set shutting down flag to prevent reconnection attempts during shutdown
  shuttingDown = true;

  // Persist any remaining queued messages
  messageQueue._persistQueues();

  // Final context sync attempt with error handling
  try {
    await syncContext();
    log('info', 'Final context sync completed');
  } catch (error) {
    log('error', `Error in final context sync: ${error.message}`);
  }

  // Close WebSocket connections gracefully
  if (githubMcpSocket) {
    try {
      githubMCP.close(1000, "Bridge shutting down");
    } catch (error) {
      log('warn', `Error closing GitHub MCP connection: ${error.message}`);
    }
  }

  if (tnosMcpSocket) {
    try {
      tnosMcpSocket.close(1000, "Bridge shutting down");
    } catch (error) {
      log('warn', `Error closing TNOS MCP connection: ${error.message}`);
    }
  }

  // Close client connections
  let clientClosePromises = [];
  clientSockets.forEach((client) => {
    if (client.readyState === WebSocket.OPEN) {
      const promise = new Promise((resolve) => {
        client.once('close', resolve);
        client.close(1000, "Bridge shutting down");
        // Ensure we don't hang forever waiting for clients
        setTimeout(resolve, 1000);
      });
      clientClosePromises.push(promise);
    }
  });

  // Wait for all clients to close or timeout
  if (clientClosePromises.length > 0) {
    try {
      await Promise.all(clientClosePromises);
      log('info', 'All client connections closed');
    } catch (error) {
      log('warn', `Error waiting for client connections to close: ${error.message}`);
    }
  }

  log('info', 'MCP Bridge shut down gracefully');
  process.exit(0);
}

// Export the functions that were in bridge.js for compatibility
module.exports = {
  executeFormula,
  queryDimensionalContext,
  performMobiusCompression,
  syncContext: synchronizeContext,
  initialize: startBridge,
  // Legacy function renamed to match new implementation
  synchronizeContext: syncContext,
  sendTnosMcpMessage: function (params) {
    const { targetLayer, messageType, payload } = params;
    if (tnosMcpSocket && tnosMcpSocket.readyState === WebSocket.OPEN) {
      try {
        tnosMcpSocket.send(JSON.stringify({
          type: 'layer_message',
          data: {
            targetLayer,
            messageType,
            payload
          }
        }));
        return {
          success: true,
          message: `Message sent to Layer ${targetLayer}`,
          timestamp: new Date().toISOString()
        };
      } catch (error) {
        log('error', `Error sending message to TNOS MCP: ${error.message}`);
        return {
          error: `Error sending message to TNOS MCP: ${error.message}`,
          success: false
        };
      }
    } else {
      messageQueue.queueForTnos({
        type: 'layer_message',
        data: {
          targetLayer,
          messageType,
          payload
        }
      });
      return {
        success: true,
        message: `Message queued for Layer ${targetLayer}`,
        timestamp: new Date().toISOString()
      };
    }
  }
};

// Start the bridge
startBridge();



