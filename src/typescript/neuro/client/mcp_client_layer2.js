/**
 * MCP Client for Layer 2 (JavaScript Reactive Layer)
 * 
 * This module implements the Model Context Protocol client for the JavaScript Reactive Layer,
 * allowing bidirectional communication between the MCP server and the JavaScript components
 * of the TNOS system.
 * 
 * @layer: Layer 2 (JavaScript Reactive Layer)
 * @system: Nervous System (Signal Processing)
 * @doc Rule - 7D Documentation:
 * - Who: Reactive Layer (UI/communication component)
 * - What: MCP client for real-time event handling
 * - When: Runtime, during system operation
 * - Where: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
 * - Why: To integrate reactive components with the MCP server
 * - How: Using WebSocket for bidirectional communication
 * - To What Extent: Full integration with 7D context framework
 */

const WebSocket = require('websocket').w3cwebsocket;
const fs = require('fs');
const path = require('path');
const EventEmitter = require('events');

// Singleton pattern to ensure only one client instance exists across the system
let instance = null;

class MCPClient extends EventEmitter {
  /**
   * Initialize the MCP Client
   * @param {Object} config - Configuration object
   */
  constructor(config = {}) {
    super();

    if (instance) {
      return instance;
    }

    this.config = {
      host: config.host || 'localhost',
      port: config.port || 8765,
      reconnectInterval: config.reconnectInterval || 5000,
      heartbeatInterval: config.heartbeatInterval || 15000,
      layerId: 'layer2',
      layerDescription: 'JavaScript Reactive Layer'
    };

    this.connected = false;
    this.sessionId = null;
    this.socket = null;
    this.reconnectTimer = null;
    this.heartbeatTimer = null;
    this.pendingMessages = [];
    this.logger = this._initLogger();

    // Add support for translation bridge
    this.translationBridge = null;
    this.translationEnabled = config.enableTranslation !== false;

    // Set up connection
    this._connect();

    // Save the instance
    instance = this;

    return instance;
  }

  /**
   * Initialize logger
   * @private
   */
  _initLogger() {
    const logDir = path.join(__dirname, '..', '..', 'logs');

    // Create logs directory if it doesn't exist
    if (!fs.existsSync(logDir)) {
      fs.mkdirSync(logDir, { recursive: true });
    }

    const logPath = path.join(logDir, 'mcp_client_layer2.log');

    return {
      info: (message) => {
        const timestamp = new Date().toISOString();
        fs.appendFileSync(logPath, `[INFO][${timestamp}] ${message}\n`);
        console.log(`[MCP-INFO] ${message}`);
      },
      error: (message, error) => {
        const timestamp = new Date().toISOString();
        const errorDetails = error ? `\n${error.stack}` : '';
        fs.appendFileSync(logPath, `[ERROR][${timestamp}] ${message}${errorDetails}\n`);
        console.error(`[MCP-ERROR] ${message}`);
      },
      debug: (message) => {
        const timestamp = new Date().toISOString();
        fs.appendFileSync(logPath, `[DEBUG][${timestamp}] ${message}\n`);
      }
    };
  }

  /**
   * Connect to the MCP server via WebSocket
   * @private
   */
  _connect() {
    if (this.socket) {
      this.socket.close();
    }

    const url = `ws://${this.config.host}:${this.config.port}/mcp`;
    this.logger.info(`Connecting to MCP server at ${url}`);

    try {
      this.socket = new WebSocket(url);

      this.socket.onopen = () => {
        this.connected = true;
        this.logger.info('Connected to MCP server');
        this._clearReconnectTimer();

        // Register this layer with the MCP server
        this._registerLayer();

        // Start heartbeat
        this._startHeartbeat();

        // Send any pending messages
        this._sendPendingMessages();

        // Emit connection event
        this.emit('connected');
      };

      this.socket.onclose = () => {
        this.connected = false;
        this.logger.info('Disconnected from MCP server');
        this._clearHeartbeatTimer();
        this._scheduleReconnect();

        // Emit disconnection event
        this.emit('disconnected');
      };

      this.socket.onerror = (error) => {
        this.logger.error('WebSocket error', error);
        this.emit('error', error);
      };

      this.socket.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          this._handleMessage(message);
        } catch (error) {
          this.logger.error('Error parsing message', error);
        }
      };
    } catch (error) {
      this.logger.error('Error creating WebSocket connection', error);
      this._scheduleReconnect();
    }
  }

  /**
   * Schedule a reconnection attempt
   * @private
   */
  _scheduleReconnect() {
    if (!this.reconnectTimer) {
      this.reconnectTimer = setTimeout(() => {
        this.logger.info('Attempting to reconnect...');
        this._connect();
      }, this.config.reconnectInterval);
    }
  }

  /**
   * Clear reconnect timer
   * @private
   */
  _clearReconnectTimer() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  /**
   * Start sending heartbeats to keep the connection alive
   * @private
   */
  _startHeartbeat() {
    this._clearHeartbeatTimer();
    this.heartbeatTimer = setInterval(() => {
      if (this.connected) {
        this._sendHeartbeat();
      }
    }, this.config.heartbeatInterval);
  }

  /**
   * Clear heartbeat timer
   * @private
   */
  _clearHeartbeatTimer() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  /**
   * Send a heartbeat message to the server
   * @private
   */
  _sendHeartbeat() {
    const heartbeatMessage = {
      type: 'heartbeat',
      source: this.config.layerId,
      timestamp: Date.now()
    };

    this._sendMessage(heartbeatMessage);
  }

  /**
   * Register this layer with the MCP server
   * @private
   */
  _registerLayer() {
    // Enhanced capabilities list to include translation capabilities
    const capabilities = [
      'reactive_events',
      'ui_interactions',
      'translation_layer'
    ];

    // Add advanced translation capabilities if bridge will be enabled
    if (this.translationEnabled) {
      capabilities.push('code_translation');
      capabilities.push('compression');
      capabilities.push('batch_processing');
    }

    const registerMessage = {
      type: 'register',
      source: this.config.layerId,
      data: {
        description: this.config.layerDescription,
        capabilities: capabilities
      },
      timestamp: Date.now()
    };

    this._sendMessage(registerMessage);
  }

  /**
   * Send any messages that were queued while disconnected
   * @private
   */
  _sendPendingMessages() {
    if (this.pendingMessages.length > 0) {
      this.logger.info(`Sending ${this.pendingMessages.length} pending messages`);

      while (this.pendingMessages.length > 0) {
        const message = this.pendingMessages.shift();
        this._sendMessage(message);
      }
    }
  }

  /**
   * Handle incoming messages from the MCP server
   * @param {Object} message - The parsed message object
   * @private
   */
  _handleMessage(message) {
    this.logger.debug(`Received message: ${JSON.stringify(message)}`);

    switch (message.type) {
      case 'register_response':
        this._handleRegisterResponse(message);
        break;

      case 'context_update':
        this._handleContextUpdate(message);
        break;

      case 'event_notification':
        this._handleEventNotification(message);
        break;

      case 'command':
        this._handleCommand(message);
        break;

      default:
        this.logger.debug(`Unhandled message type: ${message.type}`);
        // Emit the message for custom handling
        this.emit('message', message);
    }
  }

  /**
   * Handle registration response
   * @param {Object} message - The registration response message
   * @private
   */
  _handleRegisterResponse(message) {
    if (message.status === 'success') {
      this.sessionId = message.session_id;
      this.logger.info(`Successfully registered with MCP server. Session ID: ${this.sessionId}`);
      this.emit('registered', this.sessionId);
    } else {
      this.logger.error(`Registration failed: ${message.error}`);
      this.emit('registration_failed', message.error);
    }
  }

  /**
   * Handle context update messages
   * @param {Object} message - The context update message
   * @private
   */
  _handleContextUpdate(message) {
    this.logger.debug(`Received context update: ${JSON.stringify(message.data)}`);
    this.emit('context_update', message.data);
  }

  /**
   * Handle event notification messages
   * @param {Object} message - The event notification message
   * @private
   */
  _handleEventNotification(message) {
    this.logger.debug(`Received event notification: ${JSON.stringify(message.data)}`);
    this.emit('event', message.data);
  }

  /**
   * Handle command messages
   * @param {Object} message - The command message
   * @private
   */
  _handleCommand(message) {
    this.logger.debug(`Received command: ${JSON.stringify(message.data)}`);

    // Check if this is a translation-related command
    const command = message.data || {};
    if (this.translationBridge &&
      (command.action === 'translate' ||
        command.action === 'compress' ||
        command.action === 'decompress')) {

      // Translation-related commands are handled by the bridge
      this.emit('command', command);
      return;
    }

    // For other commands, emit the command event for application handling
    this.emit('command', command);

    // Send generic command response
    this.sendCommandResponse(message.id, 'success', { result: 'Command executed' });
  }

  /**
   * Send a message to the MCP server
   * @param {Object} message - The message to send
   * @private
   */
  _sendMessage(message) {
    if (this.connected && this.socket.readyState === WebSocket.OPEN) {
      try {
        this.socket.send(JSON.stringify(message));
      } catch (error) {
        this.logger.error('Error sending message', error);
        // Queue the message for later retry
        this.pendingMessages.push(message);
      }
    } else {
      // Queue the message for later
      this.pendingMessages.push(message);
    }
  }

  /**
   * Send context data to the MCP server
   * @param {Object} contextData - The context data to send
   * @public
   */
  sendContext(contextData) {
    const message = {
      type: 'context_update',
      source: this.config.layerId,
      data: contextData,
      timestamp: Date.now()
    };

    this._sendMessage(message);
  }

  /**
   * Send an event notification to the MCP server
   * @param {string} eventType - The type of event
   * @param {Object} eventData - The event data
   * @public
   */
  sendEvent(eventType, eventData) {
    const message = {
      type: 'event_notification',
      source: this.config.layerId,
      data: {
        event_type: eventType,
        ...eventData
      },
      timestamp: Date.now()
    };

    this._sendMessage(message);
  }

  /**
   * Send a command response to the MCP server
   * @param {string} commandId - The ID of the command being responded to
   * @param {string} status - The status of the command execution (success/error)
   * @param {Object} responseData - Additional response data
   * @public
   */
  sendCommandResponse(commandId, status, responseData = {}) {
    const message = {
      type: 'command_response',
      source: this.config.layerId,
      command_id: commandId,
      status: status,
      data: responseData,
      timestamp: Date.now()
    };

    this._sendMessage(message);
  }

  /**
   * Convert reactive context to 7D context format
   * @param {Object} reactiveContext - The reactive context data
   * @returns {Object} - The 7D context object
   * @public
   */
  reactiveContextTo7D(reactiveContext) {
    // Create 7D context structure
    const context7D = {
      dimensions: {
        WHO: {},
        WHAT: {},
        WHEN: {},
        WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
        WHY: {},
        HOW: {},
        EXTENT: {}
      },
      metadata: {
        source_layer: this.config.layerId,
        source_format: 'tnos_js_reactive',
        timestamp: Date.now()
      }
    };

    // Map reactive context properties to 7D dimensions
    if (reactiveContext.actor) {
      context7D.dimensions.WHO.actor = reactiveContext.actor;
    }

    if (reactiveContext.event) {
      context7D.dimensions.WHAT.event = reactiveContext.event;
    }

    if (reactiveContext.timing) {
      context7D.dimensions.WHEN.timing = reactiveContext.timing;
    }

    if (reactiveContext.environment) {
      context7D.dimensions.WHERE/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
    }

    if (reactiveContext.motivation) {
      context7D.dimensions.WHY.motivation = reactiveContext.motivation;
    }

    if (reactiveContext.response) {
      context7D.dimensions.HOW.response = reactiveContext.response;
    }

    if (reactiveContext.intensity) {
      context7D.dimensions.EXTENT.intensity = reactiveContext.intensity;
    }

    // Add any advanced reactive constructs
    this._mapAdvancedReactiveConstructs(reactiveContext, context7D);

    return context7D;
  }

  /**
   * Convert 7D context to reactive context format
   * @param {Object} context7D - The 7D context data
   * @returns {Object} - The reactive context object
   * @public
   */
  context7DToReactive(context7D) {
    const reactiveContext = {};

    // Extract dimensions
    const dimensions = context7D.dimensions || {};

    // Map 7D dimensions to reactive context properties
    if (dimensions.WHO && dimensions.WHO.actor) {
      reactiveContext.actor = dimensions.WHO.actor;
    }

    if (dimensions.WHAT && dimensions.WHAT.event) {
      reactiveContext.event = dimensions.WHAT.event;
    }

    if (dimensions.WHEN && dimensions.WHEN.timing) {
      reactiveContext.timing = dimensions.WHEN.timing;
    }

    if (dimensions.WHERE /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
      reactiveContext.environment = dimensions.WHERE/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
    }

    if (dimensions.WHY && dimensions.WHY.motivation) {
      reactiveContext.motivation = dimensions.WHY.motivation;
    }

    if (dimensions.HOW && dimensions.HOW.response) {
      reactiveContext.response = dimensions.HOW.response;
    }

    if (dimensions.EXTENT && dimensions.EXTENT.intensity) {
      reactiveContext.intensity = dimensions.EXTENT.intensity;
    }

    // Extract any advanced reactive constructs
    const advancedConstructs = this._extractAdvancedReactiveConstructs(dimensions);

    return {
      ...reactiveContext,
      ...advancedConstructs,
      metadata: {
        source: 'mcp',
        timestamp: context7D.timestamp || Date.now(),
        session_id: this.sessionId
      }
    };
  }

  /**
   * Map advanced reactive constructs to 7D dimensions
   * @param {Object} reactiveContext - The reactive context
   * @param {Object} context7D - The 7D context to modify
   * @private
   */
  _mapAdvancedReactiveConstructs(reactiveContext, context7D) {
    // Reactive constructs specific to Layer 2
    if (reactiveContext.eventPriority) {
      context7D.dimensions.EXTENT.eventPriority = reactiveContext.eventPriority;
    }

    if (reactiveContext.uiComponent) {
      context7D.dimensions.WHO.uiComponent = reactiveContext.uiComponent;
    }

    if (reactiveContext.eventChain) {
      context7D.dimensions.WHEN.eventChain = reactiveContext.eventChain;
    }

    if (reactiveContext.eventSource) {
      context7D.dimensions.WHERE/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
    }

    if (reactiveContext.userIntent) {
      context7D.dimensions.WHY.userIntent = reactiveContext.userIntent;
    }

    if (reactiveContext.interactionMethod) {
      context7D.dimensions.HOW.interactionMethod = reactiveContext.interactionMethod;
    }
  }

  /**
   * Extract advanced reactive constructs from 7D dimensions
   * @param {Object} dimensions - The 7D dimensions
   * @returns {Object} - Advanced reactive constructs
   * @private
   */
  _extractAdvancedReactiveConstructs(dimensions) {
    const advancedConstructs = {};

    if (dimensions.EXTENT && dimensions.EXTENT.eventPriority) {
      advancedConstructs.eventPriority = dimensions.EXTENT.eventPriority;
    }

    if (dimensions.WHO && dimensions.WHO.uiComponent) {
      advancedConstructs.uiComponent = dimensions.WHO.uiComponent;
    }

    if (dimensions.WHEN && dimensions.WHEN.eventChain) {
      advancedConstructs.eventChain = dimensions.WHEN.eventChain;
    }

    if (dimensions.WHERE /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
      advancedConstructs.eventSource = dimensions.WHERE/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
    }

    if (dimensions.WHY && dimensions.WHY.userIntent) {
      advancedConstructs.userIntent = dimensions.WHY.userIntent;
    }

    if (dimensions.HOW && dimensions.HOW.interactionMethod) {
      advancedConstructs.interactionMethod = dimensions.HOW.interactionMethod;
    }

    return advancedConstructs;
  }

  /**
   * Get the reactive state for the current environment
   * @returns {Object} - The reactive state
   * @public
   */
  getReactiveState() {
    // This would normally obtain the actual state from the JavaScript environment
    // For this implementation, we're providing a simulated state
    return {
      actor: 'ui_system',
      event: 'user_interaction',
      timing: 'realtime',
      environment: 'browser',
      motivation: 'user_request',
      response: 'visual_feedback',
      intensity: 'medium',
      eventPriority: 3,
      uiComponent: 'main_panel',
      eventChain: ['click', 'process', 'update'],
      eventSource: 'user_interface',
      userIntent: 'data_retrieval',
      interactionMethod: 'click'
    };
  }

  /**
   * Check if the client is connected to the MCP server
   * @returns {boolean} - True if connected, false otherwise
   * @public
   */
  isConnected() {
    return this.connected;
  }

  /**
   * Disconnect from the MCP server
   * @public
   */
  disconnect() {
    if (this.socket) {
      this.socket.close();
    }

    this._clearReconnectTimer();
    this._clearHeartbeatTimer();
  }

  /**
   * Get the translation bridge instance
   * @returns {Object|null} - The translation bridge instance or null if not enabled
   * @public
   */
  getTranslationBridge() {
    if (!this.translationBridge && this.translationEnabled) {
      try {
        // Dynamically import the translation bridge to avoid circular dependencies
        const createTranslationBridge = require('../src/main/javascript/translation_bridge.js');
        this.translationBridge = createTranslationBridge({
          mcpConfig: this.config
        });
        this.logger.info('Translation bridge initialized and connected to MCP client');
      } catch (error) {
        this.logger.error('Failed to initialize translation bridge', error);
        this.translationEnabled = false;
      }
    }

    return this.translationBridge;
  }

  /**
   * Enhanced reactive context to 7D conversion that uses the translation layer
   * @param {Object} reactiveContext - The reactive context data
   * @returns {Promise<Object>} - The 7D context object
   * @public
   */
  async enhancedReactiveContextTo7D(reactiveContext) {
    // If translation bridge is available, use it for enhanced conversion
    const bridge = this.getTranslationBridge();
    if (bridge) {
      return bridge.reactiveContextTo7D(reactiveContext);
    }

    // Fall back to standard conversion
    return this.reactiveContextTo7D(reactiveContext);
  }

  /**
   * Enhanced 7D context to reactive conversion that uses the translation layer
   * @param {Object} context7D - The 7D context data
   * @returns {Promise<Object>} - The reactive context object
   * @public
   */
  async enhancedContext7DToReactive(context7D) {
    // If translation bridge is available, use it for enhanced conversion
    const bridge = this.getTranslationBridge();
    if (bridge) {
      return bridge.context7DToReactive(context7D);
    }

    // Fall back to standard conversion
    return this.context7DToReactive(context7D);
  }

  /**
   * Translate code between languages
   * @param {string} code - The code to translate
   * @param {string} sourceLanguage - The source language
   * @param {string} targetLanguage - The target language
   * @param {Object} options - Translation options
   * @returns {Promise<string>} - The translated code
   * @public
   */
  async translateCode(code, sourceLanguage, targetLanguage, options = {}) {
    const bridge = this.getTranslationBridge();
    if (!bridge) {
      throw new Error('Translation bridge not available');
    }

    return bridge.translate(code, sourceLanguage, targetLanguage, options);
  }

  /**
   * Compress data
   * @param {string} data - The data to compress
   * @param {string} sourceLanguage - The source language
   * @param {Object} options - Compression options
   * @returns {Promise<Object>} - The compressed data
   * @public
   */
  async compressData(data, sourceLanguage, options = {}) {
    const bridge = this.getTranslationBridge();
    if (!bridge) {
      throw new Error('Translation bridge not available');
    }

    return bridge.compress(data, sourceLanguage, options);
  }

  /**
   * Decompress data
   * @param {string} compressed - The compressed data
   * @param {number} entropy - The entropy value
   * @param {string} targetLanguage - The target language
   * @param {Object} options - Decompression options
   * @returns {Promise<string>} - The decompressed data
   * @public
   */
  async decompressData(compressed, entropy, targetLanguage, options = {}) {
    const bridge = this.getTranslationBridge();
    if (!bridge) {
      throw new Error('Translation bridge not available');
    }

    return bridge.decompress(compressed, entropy, targetLanguage, options);
  }

  /**
   * Get translation statistics
   * @returns {Object|null} - Translation statistics or null if not available
   * @public
   */
  getTranslationStats() {
    const bridge = this.getTranslationBridge();
    if (!bridge) {
      return null;
    }

    return bridge.getStatistics();
  }
}

// Export a singleton instance
module.exports = function getClient(config) {
  return new MCPClient(config);
};


