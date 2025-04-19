#!/usr/bin/env node

// WHO: DirectMCPConnector
// WHAT: Direct connection implementation for GitHub MCP
// WHEN: During direct connection mode operation
// WHERE: Layer 2 / Direct Connection Layer
// WHY: To provide direct GitHub MCP connections bypassing bridge
// HOW: Using Node.js WebSocket client with 7D context translation
// EXTENT: All GitHub direct MCP communications

/**
 * GitHub MCP Direct Connection
 * ----------------------------
 * This module provides direct connection capabilities between
 * GitHub Copilot and the TNOS MCP server, bypassing the normal
 * bridge for special use cases.
 * 
 * Configuration is loaded from config/github_mcp_config.json.
 */

'use strict';

const fs = require('fs');
const path = require('path');
const WebSocket = require('ws');
const http = require('http');
const crypto = require('crypto');

// Determine project root directory
const projectRoot = path.resolve(__dirname, '../..');

// Configure logging
const logsDir = path.join(projectRoot, 'logs');
if (!fs.existsSync(logsDir)) {
  fs.mkdirSync(logsDir, { recursive: true });
}

const logFile = path.join(logsDir, 'github_mcp_direct.log');
const logger = {
  info: (message) => logToFile('INFO', message),
  warn: (message) => logToFile('WARN', message),
  error: (message) => logToFile('ERROR', message),
  debug: (message) => logToFile('DEBUG', message)
};

function logToFile(level, message) {
  const timestamp = new Date().toISOString();
  const logMessage = `${timestamp} [${level}] ${message}\n`;

  console.log(logMessage.trim());

  try {
    fs.appendFileSync(logFile, logMessage);
  } catch (err) {
    console.error(`Failed to write to log file: ${err.message}`);
  }
}

// Load configuration
const configPath = path.join(projectRoot, 'config', 'github_mcp_config.json');
let config;

try {
  config = JSON.parse(fs.readFileSync(configPath, 'utf8'));
  logger.info(`Loaded configuration from ${configPath}`);
} catch (err) {
  logger.error(`Failed to load configuration: ${err.message}`);
  logger.info('Using default configuration');

  // Default configuration
  config = {
    server: {
      host: 'localhost',
      port: 8080,
      direct_connection: true
    },
    connection: {
      type: 'direct',
      tnos_mcp_host: 'localhost',
      tnos_mcp_port: 9001,
      use_compression: true
    },
    logging: {
      level: 'info',
      file: logFile
    }
  };
}

// Save PID file
const pidFile = path.join(logsDir, 'github_mcp_direct.pid');
try {
  fs.writeFileSync(pidFile, process.pid.toString());
  logger.info(`Wrote PID ${process.pid} to ${pidFile}`);
} catch (err) {
  logger.error(`Failed to write PID file: ${err.message}`);
}

// Clean up PID file on exit
process.on('SIGINT', () => cleanup('SIGINT'));
process.on('SIGTERM', () => cleanup('SIGTERM'));
process.on('exit', () => cleanup('exit'));

function cleanup(signal) {
  logger.info(`Received ${signal} signal, cleaning up...`);
  if (fs.existsSync(pidFile)) {
    fs.unlinkSync(pidFile);
    logger.info(`Removed PID file: ${pidFile}`);
  }

  // Close any open connections
  if (tnosWsConnection && tnosWsConnection.readyState === WebSocket.OPEN) {
    tnosWsConnection.close();
    logger.info('Closed TNOS WebSocket connection');
  }

  if (server) {
    server.close(() => {
      logger.info('Closed HTTP server');
      process.exit(0);
    });

    // Force exit after 1 second if server.close() doesn't complete
    setTimeout(() => {
      logger.info('Forcing exit after timeout');
      process.exit(0);
    }, 1000);
  } else {
    process.exit(0);
  }
}

// Context Translation Functions
function createContext(who, what, when, where, why, how, extent) {
  return {
    who: who || 'GitHubMCPDirect',
    what: what || 'Connect',
    when: when || new Date().toISOString(),
    where: where || 'DirectConnector',
    why: why || 'DirectOperation',
    how: how || 'WSConnection',
    extent: extent || 'Tool',
    metadata: {}
  };
}

function githubToTnosContext(githubRequest) {
  const toolName = githubRequest.name || 'unknown';
  const params = githubRequest.parameters || {};

  return createContext(
    'GitHub Copilot',
    `Tool:${toolName}`,
    new Date().toISOString(),
    'DirectConnector',
    `GitHub operation: ${toolName}`,
    'direct_connection',
    params.scope || 'single_resource'
  );
}

function tnosToGithubResponse(tnosResponse, originalRequest) {
  // Extract content from TNOS response
  let resultContent = tnosResponse.content || {};

  // Parse content if it's a string
  if (typeof resultContent === 'string') {
    try {
      resultContent = JSON.parse(resultContent);
    } catch (e) {
      resultContent = { content: resultContent };
    }
  }

  // Format response for GitHub
  return {
    result: resultContent,
    metadata: {
      tnos_context: tnosResponse.context || {},
      direct_connection: true,
      request_tool: originalRequest ? originalRequest.name : 'unknown'
    }
  };
}

// Setup TNOS WebSocket Connection
const tnosMcpUrl = `ws://${config.connection.tnos_mcp_host}:${config.connection.tnos_mcp_port}`;
let tnosWsConnection = null;
let connectionAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 5;
const RECONNECT_DELAY = 3000; // 3 seconds

function connectToTnosMcp() {
  logger.info(`Connecting to TNOS MCP at ${tnosMcpUrl}`);

  if (tnosWsConnection && (tnosWsConnection.readyState === WebSocket.OPEN || tnosWsConnection.readyState === WebSocket.CONNECTING)) {
    logger.info('Already connected or connecting to TNOS MCP');
    return;
  }

  connectionAttempts++;

  tnosWsConnection = new WebSocket(tnosMcpUrl);

  tnosWsConnection.on('open', () => {
    logger.info('Connected to TNOS MCP server');
    connectionAttempts = 0;

    // Negotiate protocol version
    const negotiationMsg = {
      message_type: 'version_negotiation',
      content: {
        supported_versions: ['1.0', '1.5', '2.0', '2.5', '3.0'],
        preferred_version: '3.0'
      },
      context: createContext()
    };

    tnosWsConnection.send(JSON.stringify(negotiationMsg));
  });

  tnosWsConnection.on('message', (data) => {
    try {
      const message = JSON.parse(data);

      // Only log non-ping messages to avoid log flooding
      if (message.message_type !== 'pong') {
        logger.debug(`Received message from TNOS MCP: ${message.message_type}`);
      }

      // Handle protocol version negotiation response
      if (message.message_type === 'version_negotiation_response') {
        const version = message.content && message.content.selected_version;
        if (version) {
          logger.info(`Negotiated protocol version: ${version}`);
        }
      }
    } catch (err) {
      logger.error(`Error processing message from TNOS MCP: ${err.message}`);
    }
  });

  tnosWsConnection.on('error', (error) => {
    logger.error(`TNOS MCP connection error: ${error.message}`);
  });

  tnosWsConnection.on('close', (code, reason) => {
    logger.warn(`TNOS MCP connection closed: ${code} - ${reason}`);

    // Attempt to reconnect if not max attempts
    if (connectionAttempts < MAX_RECONNECT_ATTEMPTS) {
      logger.info(`Attempting to reconnect in ${RECONNECT_DELAY / 1000} seconds... (Attempt ${connectionAttempts})`);
      setTimeout(connectToTnosMcp, RECONNECT_DELAY);
    } else {
      logger.error(`Failed to connect to TNOS MCP after ${MAX_RECONNECT_ATTEMPTS} attempts`);
    }
  });
}

// HTTP Server for GitHub MCP requests
const { host, port } = config.server;
const server = http.createServer((req, res) => {
  if (req.method === 'POST') {
    let body = '';

    req.on('data', (chunk) => {
      body += chunk.toString();
    });

    req.on('end', () => {
      try {
        // Parse GitHub request
        const githubRequest = JSON.parse(body);
        const requestId = crypto.randomUUID();

        logger.info(`Received GitHub MCP request: ${githubRequest.name} (${requestId})`);

        // Check TNOS connection
        if (!tnosWsConnection || tnosWsConnection.readyState !== WebSocket.OPEN) {
          logger.error(`TNOS MCP connection not available for request ${requestId}`);

          res.writeHead(503, { 'Content-Type': 'application/json' });
          res.end(JSON.stringify({
            error: 'TNOS MCP server connection unavailable',
            status: 'error',
            requestId
          }));

          // Try to reconnect
          connectToTnosMcp();
          return;
        }

        // Translate context
        const tnosContext = githubToTnosContext(githubRequest);

        // Create TNOS message
        const tnosMessage = {
          message_type: 'tool_request',
          content: {
            tool: githubRequest.name,
            parameters: githubRequest.parameters
          },
          context: tnosContext
        };

        // Send request to TNOS MCP
        tnosWsConnection.send(JSON.stringify(tnosMessage));

        // Wait for response with timeout
        const RESPONSE_TIMEOUT = 30000; // 30 seconds
        let responseTimer = null;
        let responseHandler = null;

        const handleResponse = new Promise((resolve, reject) => {
          responseHandler = (data) => {
            try {
              const response = JSON.parse(data);

              // Check if this is a response to our request
              if (response.message_type === 'tool_response' ||
                response.message_type === 'error') {

                // Clear the timeout
                clearTimeout(responseTimer);

                // Remove this response handler
                tnosWsConnection.removeListener('message', responseHandler);

                // Translate the response
                const githubResponse = tnosToGithubResponse(response, githubRequest);

                // Add request ID for tracking
                githubResponse.requestId = requestId;

                // Return the response
                resolve(githubResponse);
              }
            } catch (err) {
              logger.error(`Error processing response: ${err.message}`);
            }
          };

          // Add the response handler
          tnosWsConnection.on('message', responseHandler);

          // Set a timeout
          responseTimer = setTimeout(() => {
            tnosWsConnection.removeListener('message', responseHandler);
            reject(new Error('Response timeout'));
          }, RESPONSE_TIMEOUT);
        });

        handleResponse.then(responseData => {
          logger.info(`Sending response for request ${requestId}`);

          res.writeHead(200, { 'Content-Type': 'application/json' });
          res.end(JSON.stringify(responseData));
        }).catch(err => {
          logger.error(`Request ${requestId} failed: ${err.message}`);

          res.writeHead(500, { 'Content-Type': 'application/json' });
          res.end(JSON.stringify({
            error: `Request failed: ${err.message}`,
            status: 'error',
            requestId
          }));
        });

      } catch (err) {
        logger.error(`Error processing request: ${err.message}`);

        res.writeHead(400, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({
          error: `Invalid request: ${err.message}`,
          status: 'error'
        }));
      }
    });
  } else {
    // Return basic status information for GET requests
    if (req.method === 'GET') {
      const status = {
        service: 'GitHub MCP Direct Connection',
        status: 'running',
        mode: 'direct_connection',
        tnos_connection: tnosWsConnection ?
          ['CONNECTING', 'OPEN', 'CLOSING', 'CLOSED'][tnosWsConnection.readyState] :
          'NOT_CONNECTED',
        uptime: process.uptime(),
        pid: process.pid
      };

      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify(status));
    } else {
      res.writeHead(405, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ error: 'Method not allowed' }));
    }
  }
});

// Start the server and connect to TNOS MCP
server.listen(port, host, () => {
  logger.info(`GitHub MCP Direct server listening at http://${host}:${port}`);
  connectToTnosMcp();
});

// Periodic health check
setInterval(() => {
  if (tnosWsConnection && tnosWsConnection.readyState === WebSocket.OPEN) {
    // Send ping to keep connection alive
    const pingMessage = {
      message_type: 'ping',
      content: { timestamp: Date.now() },
      context: createContext('GitHubMCPDirect', 'HealthCheck')
    };

    tnosWsConnection.send(JSON.stringify(pingMessage));
  } else if (!tnosWsConnection || tnosWsConnection.readyState === WebSocket.CLOSED) {
    // Try to reconnect if connection is closed
    if (connectionAttempts < MAX_RECONNECT_ATTEMPTS) {
      logger.info('Connection closed, attempting to reconnect...');
      connectToTnosMcp();
    }
  }
}, 30000); // Every 30 seconds