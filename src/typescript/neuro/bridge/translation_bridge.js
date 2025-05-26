/**
 * Translation Bridge for TNOS
 * 
 * This module connects the Layer 2 MCP client with the DynamicTranslationLayer,
 * allowing efficient code and message translation between different language contexts.
 * 
 * @layer: Layer 2 (JavaScript Reactive Layer)
 * @system: Advanced Translation Layer (ATL)
 * @doc Rule - 7D Documentation:
 * - Who: Translation Bridge (connector component)
 * - What: Integration between MCP client and DynamicTranslationLayer
 * - When: Runtime, during translation operations
 * - Where: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
 * - Why: To enable efficient cross-language communication
 * - How: Using optimized compression and translation algorithms
 * - To What Extent: Full bidirectional translation with context preservation
 */

const path = require('path');
const fs = require('fs');
const EventEmitter = require('events');

// Import the DynamicTranslationLayer
const DynamicTranslationLayer = require('../src/main/javascript/DynamicTranslationLayer.js');

// Import the MCP client
const getMCPClient = require('../src/main/javascript/mcp_client_layer2.js');

// Singleton pattern
let instance = null;

class TranslationBridge extends EventEmitter {
  /**
   * Initialize the Translation Bridge
   * @param {Object} config - Configuration object
   */
  constructor(config = {}) {
    super();

    if (instance) {
      return instance;
    }

    this.config = {
      mcpConfig: config.mcpConfig || {},
      translationCacheSize: config.translationCacheSize || 500,
      compressionLevel: config.compressionLevel || 6,
      debugMode: config.debugMode || false,
      autoConnect: config.autoConnect !== false
    };

    // Initialize components
    this.mcpClient = getMCPClient(this.config.mcpConfig);
    this.translationLayer = DynamicTranslationLayer;
    this.translationCache = new Map();
    this.logger = this._initLogger();

    // Configure translation layer
    this.translationLayer.setCompressionLevel(this.config.compressionLevel);

    // Set up event handlers for MCP client
    this._setupMCPEventHandlers();

    // Save instance
    instance = this;

    this.logger.info('Translation Bridge initialized');

    // Auto connect if enabled
    if (this.config.autoConnect) {
      this.connect();
    }

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

    const logPath = path.join(logDir, 'translation_bridge.log');

    return {
      info: (message) => {
        const timestamp = new Date().toISOString();
        fs.appendFileSync(logPath, `[INFO][${timestamp}] ${message}\n`);
        if (this.config.debugMode) {
          console.log(`[TRANSLATION-BRIDGE-INFO] ${message}`);
        }
      },
      error: (message, error) => {
        const timestamp = new Date().toISOString();
        const errorDetails = error ? `\n${error.stack}` : '';
        fs.appendFileSync(logPath, `[ERROR][${timestamp}] ${message}${errorDetails}\n`);
        console.error(`[TRANSLATION-BRIDGE-ERROR] ${message}`);
      },
      debug: (message) => {
        if (this.config.debugMode) {
          const timestamp = new Date().toISOString();
          fs.appendFileSync(logPath, `[DEBUG][${timestamp}] ${message}\n`);
        }
      }
    };
  }

  /**
   * Set up event handlers for MCP client
   * @private
   */
  _setupMCPEventHandlers() {
    // Listen for MCP client events
    this.mcpClient.on('connected', () => {
      this.logger.info('MCP client connected');
      this.emit('connected');
    });

    this.mcpClient.on('disconnected', () => {
      this.logger.info('MCP client disconnected');
      this.emit('disconnected');
    });

    this.mcpClient.on('registered', (sessionId) => {
      this.logger.info(`MCP client registered with session ID: ${sessionId}`);
      this.emit('registered', sessionId);

      // Register translation capabilities
      this._registerTranslationCapabilities();
    });

    this.mcpClient.on('message', (message) => {
      // Check if it's a translation-related message
      if (message.type === 'translation_request') {
        this._handleTranslationRequest(message);
      }
    });

    this.mcpClient.on('command', (command) => {
      // Handle translation-related commands
      if (command.action === 'translate') {
        this._handleTranslationCommand(command);
      }
    });

    this.mcpClient.on('error', (error) => {
      this.logger.error('MCP client error', error);
      this.emit('error', error);
    });
  }

  /**
   * Register translation capabilities with the MCP server
   * @private
   */
  _registerTranslationCapabilities() {
    const supportInfo = this.translationLayer.getSupportInfo();

    // Send capabilities to MCP server
    this.mcpClient.sendEvent('capabilities_update', {
      component: 'translation_layer',
      capabilities: {
        translation: {
          languages: ['javascript', 'typescript', 'python', 'java', 'cpp'],
          compression: true,
          batch_processing: true,
          streaming: true
        },
        system_info: supportInfo
      }
    });

    this.logger.info('Registered translation capabilities with MCP server');
  }

  /**
   * Handle translation request from MCP
   * @param {Object} message - The translation request message
   * @private
   */
  async _handleTranslationRequest(message) {
    try {
      const { source_language, target_language, code, options = {} } = message.data;

      this.logger.debug(`Received translation request: ${source_language} → ${target_language}`);

      // Generate cache key
      const cacheKey = `${source_language}:${target_language}:${code.substring(0, 100)}`;

      // Check cache first
      if (this.translationCache.has(cacheKey)) {
        const cachedResult = this.translationCache.get(cacheKey);
        this.logger.debug('Using cached translation result');

        // Send response with cached result
        this._sendTranslationResponse(message.id, {
          status: 'success',
          translated_code: cachedResult,
          source_language,
          target_language,
          metrics: {
            cache_hit: true,
            translation_time: 0
          }
        });

        return;
      }

      // Perform the translation
      const startTime = Date.now();
      const translatedCode = await this.translationLayer.translateBetweenLanguages(
        code,
        source_language,
        target_language,
        options
      );
      const translationTime = Date.now() - startTime;

      // Cache the result
      if (this.translationCache.size >= this.config.translationCacheSize) {
        // Remove oldest entry if cache is full
        const oldestKey = this.translationCache.keys().next().value;
        this.translationCache.delete(oldestKey);
      }
      this.translationCache.set(cacheKey, translatedCode);

      // Send response
      this._sendTranslationResponse(message.id, {
        status: 'success',
        translated_code: translatedCode,
        source_language,
        target_language,
        metrics: {
          cache_hit: false,
          translation_time: translationTime
        }
      });

    } catch (error) {
      this.logger.error('Error handling translation request', error);

      // Send error response
      this._sendTranslationResponse(message.id, {
        status: 'error',
        error: error.message
      });
    }
  }

  /**
   * Send translation response to MCP server
   * @param {string} requestId - The original request ID
   * @param {Object} responseData - The response data
   * @private
   */
  _sendTranslationResponse(requestId, responseData) {
    this.mcpClient.sendCommandResponse(requestId, responseData.status === 'success' ? 'success' : 'error', responseData);
  }

  /**
   * Handle translation command from MCP
   * @param {Object} command - The translation command
   * @private
   */
  async _handleTranslationCommand(command) {
    try {
      const { action, source_language, target_language, code, batch, options = {} } = command;

      this.logger.debug(`Received translation command: ${action}`);

      let result;
      if (action === 'translate') {
        if (batch && Array.isArray(code)) {
          // Handle batch translation
          result = await this.translationLayer.processBatch(
            'translate',
            code.map(item => ({
              code: item,
              sourceLanguage: source_language,
              targetLanguage: target_language
            })),
            options
          );
        } else {
          // Handle single translation
          result = await this.translationLayer.translateBetweenLanguages(
            code,
            source_language,
            target_language,
            options
          );
        }
      } else if (action === 'compress') {
        if (batch && Array.isArray(code)) {
          // Handle batch compression
          result = await this.translationLayer.processBatch(
            'compress',
            code.map(item => ({
              data: item,
              sourceLanguage: source_language
            })),
            options
          );
        } else {
          // Handle single compression
          result = await this.translationLayer.compressData(
            code,
            source_language,
            options
          );
        }
      } else if (action === 'decompress') {
        if (batch && Array.isArray(code)) {
          // Handle batch decompression
          result = await this.translationLayer.processBatch(
            'decompress',
            code.map(item => ({
              compressed: item.compressed,
              entropy: item.entropy,
              targetLanguage: target_language
            })),
            options
          );
        } else {
          // Handle single decompression
          result = await this.translationLayer.decompressData(
            code.compressed,
            code.entropy,
            target_language,
            options
          );
        }
      }

      // Send success response
      this.mcpClient.sendCommandResponse(command.id, 'success', {
        action,
        result,
        metrics: this.translationLayer.getCompressionStats()
      });

    } catch (error) {
      this.logger.error('Error handling translation command', error);

      // Send error response
      this.mcpClient.sendCommandResponse(command.id, 'error', {
        error: error.message
      });
    }
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
  async translate(code, sourceLanguage, targetLanguage, options = {}) {
    try {
      this.logger.debug(`Translating code: ${sourceLanguage} → ${targetLanguage}`);

      // Generate cache key
      const cacheKey = `${sourceLanguage}:${targetLanguage}:${code.substring(0, 100)}`;

      // Check cache first
      if (this.translationCache.has(cacheKey)) {
        this.logger.debug('Using cached translation result');
        return this.translationCache.get(cacheKey);
      }

      // Perform the translation
      const translatedCode = await this.translationLayer.translateBetweenLanguages(
        code,
        sourceLanguage,
        targetLanguage,
        options
      );

      // Cache the result
      if (this.translationCache.size >= this.config.translationCacheSize) {
        // Remove oldest entry if cache is full
        const oldestKey = this.translationCache.keys().next().value;
        this.translationCache.delete(oldestKey);
      }
      this.translationCache.set(cacheKey, translatedCode);

      return translatedCode;
    } catch (error) {
      this.logger.error('Translation error', error);
      throw error;
    }
  }

  /**
   * Compress data using the translation layer
   * @param {string} data - The data to compress
   * @param {string} sourceLanguage - The source language
   * @param {Object} options - Compression options
   * @returns {Promise<Object>} - The compressed data result
   * @public
   */
  async compress(data, sourceLanguage, options = {}) {
    try {
      return await this.translationLayer.compressData(data, sourceLanguage, options);
    } catch (error) {
      this.logger.error('Compression error', error);
      throw error;
    }
  }

  /**
   * Decompress data using the translation layer
   * @param {string} compressed - The compressed data
   * @param {number} entropy - The entropy value
   * @param {string} targetLanguage - The target language
   * @param {Object} options - Decompression options
   * @returns {Promise<string>} - The decompressed data
   * @public
   */
  async decompress(compressed, entropy, targetLanguage, options = {}) {
    try {
      return await this.translationLayer.decompressData(compressed, entropy, targetLanguage, options);
    } catch (error) {
      this.logger.error('Decompression error', error);
      throw error;
    }
  }

  /**
   * Process a batch of translation or compression operations
   * @param {string} operation - The operation type ('translate', 'compress', or 'decompress')
   * @param {Array} items - The items to process
   * @param {Object} options - Processing options
   * @returns {Promise<Array>} - The processed results
   * @public
   */
  async processBatch(operation, items, options = {}) {
    try {
      return await this.translationLayer.processBatch(operation, items, options);
    } catch (error) {
      this.logger.error(`Batch processing error (${operation})`, error);
      throw error;
    }
  }

  /**
   * Convert reactive context to 7D context with translation optimization
   * @param {Object} reactiveContext - The reactive context
   * @returns {Promise<Object>} - The optimized 7D context
   * @public
   */
  async reactiveContextTo7D(reactiveContext) {
    try {
      // First convert using MCP client's built-in converter
      const context7D = this.mcpClient.reactiveContextTo7D(reactiveContext);

      // Check if any code blocks need translation
      if (reactiveContext.code) {
        // Compress code blocks for more efficient transmission
        const compressed = await this.translationLayer.compressData(
          reactiveContext.code,
          reactiveContext.codeLanguage || 'javascript'
        );

        // Add compressed code to the context
        context7D.dimensions.WHAT.compressed_code = compressed.compressed;
        context7D.dimensions.WHAT.code_entropy = compressed.entropy;
        context7D.dimensions.WHAT.code_language = reactiveContext.codeLanguage || 'javascript';
        context7D.dimensions.WHAT.code_metadata = compressed.metadata;
      }

      return context7D;
    } catch (error) {
      this.logger.error('Error converting reactive context to 7D', error);
      // Fall back to basic conversion if optimization fails
      return this.mcpClient.reactiveContextTo7D(reactiveContext);
    }
  }

  /**
   * Convert 7D context to reactive context with translation optimization
   * @param {Object} context7D - The 7D context
   * @returns {Promise<Object>} - The optimized reactive context
   * @public
   */
  async context7DToReactive(context7D) {
    try {
      // First convert using MCP client's built-in converter
      const reactiveContext = this.mcpClient.context7DToReactive(context7D);

      // Check if any compressed code blocks need decompression
      if (context7D.dimensions?.WHAT?.compressed_code) {
        // Decompress code blocks
        const decompressed = await this.translationLayer.decompressData(
          context7D.dimensions.WHAT.compressed_code,
          context7D.dimensions.WHAT.code_entropy,
          context7D.dimensions.WHAT.code_language || 'javascript'
        );

        // Add decompressed code to the reactive context
        reactiveContext.code = decompressed;
        reactiveContext.codeLanguage = context7D.dimensions.WHAT.code_language;
      }

      return reactiveContext;
    } catch (error) {
      this.logger.error('Error converting 7D context to reactive', error);
      // Fall back to basic conversion if optimization fails
      return this.mcpClient.context7DToReactive(context7D);
    }
  }

  /**
   * Get translation layer statistics and status
   * @returns {Object} - Statistics and status information
   * @public
   */
  getStatistics() {
    return {
      translationStats: this.translationLayer.getCompressionStats(),
      cacheSize: this.translationCache.size,
      cacheMaxSize: this.config.translationCacheSize,
      supportInfo: this.translationLayer.getSupportInfo(),
      mcpConnected: this.mcpClient.isConnected(),
      mcpSessionId: this.mcpClient.sessionId
    };
  }

  /**
   * Connect to the MCP server
   * @public
   */
  connect() {
    // The MCP client auto-connects in its constructor,
    // but we provide this method for explicit connection control
    if (!this.mcpClient.isConnected()) {
      this.mcpClient._connect();
    }
  }

  /**
   * Disconnect from the MCP server
   * @public
   */
  disconnect() {
    this.mcpClient.disconnect();
  }

  /**
   * Clear translation caches
   * @public
   */
  clearCaches() {
    this.translationCache.clear();
    this.translationLayer.clearCaches();
    this.logger.info('All translation caches cleared');
  }
}

// Export a factory function
module.exports = function createTranslationBridge(config) {
  return new TranslationBridge(config);
};


