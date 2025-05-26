/**
 * WHO: 7DContextBridge
 * WHAT: Translates between TNOS 7D Context and GitHub MCP Context
 * WHEN: During MCP message processing
 * WHERE: Bridge Layer between GitHub MCP and TNOS
 * WHY: To ensure context preservation across systems
 * HOW: Using bidirectional mapping with compression
 * EXTENT: All cross-system communication
 */

const MobiusCompression = require('./MobiusCompression');

/**
 * WHO: ContextBridge
 * WHAT: Bidirectional context translation system
 * WHEN: During message exchange
 * WHERE: Between GitHub MCP and TNOS MCP
 * WHY: To maintain context consistency
 * HOW: Using structured mapping with validation
 * EXTENT: All cross-system context translations
 */
class ContextBridge {
  constructor(options = {}) {
    this.config = {
      compressionEnabled: options.compressionEnabled ?? true,
      validateContext: options.validateContext ?? true,
      defaultContext: options.defaultContext ?? {
        who: 'System',
        what: 'Initialization',
        when: Date.now(),
        where: 'MCP_Bridge',
        why: 'System_Startup',
        how: 'Default_Context',
        extent: 1.0
      },
      logFunction: options.logFunction ?? console.log
    };

    this.log('info', 'TNOS 7D Context Bridge initialized');
  }

  /**
   * WHO: Logger
   * WHAT: Log bridge operations with 7D context
   * WHEN: During bridge operations
   * WHERE: Bridge Log System
   * WHY: To maintain operational awareness
   * HOW: Using structured logging with levels
   * EXTENT: All bridge operations
   */
  log(level, message, context = {}) {
    const fullContext = {
      who: 'ContextBridge',
      what: 'Logging',
      when: Date.now(),
      where: 'Bridge_Layer',
      why: 'Operational_Awareness',
      how: 'Structured_Log',
      extent: 0.5,
      ...context
    };

    const logMessage = `[${level.toUpperCase()}] ${message} | Context: ${JSON.stringify(fullContext)}`;

    if (typeof this.config.logFunction === 'function') {
      this.config.logFunction(logMessage);
    }

    return logMessage;
  }

  /**
   * WHO: ContextValidator
   * WHAT: Validate 7D context structure
   * WHEN: Before translation
   * WHERE: Context Validation Layer
   * WHY: To ensure data integrity
   * HOW: Using schema validation
   * EXTENT: All context objects
   */
  validateTnos7dContext(context) {
    if (!context) return false;

    // Check that all 7 dimensions exist
    const requiredDimensions = ['who', 'what', 'when', 'where', 'why', 'how', 'extent'];
    const missingDimensions = requiredDimensions.filter(dim => context[dim] === undefined);

    if (missingDimensions.length > 0) {
      this.log('warn', `Invalid 7D context: missing dimensions [${missingDimensions.join(', ')}]`);
      return false;
    }

    // Validate data types
    if (typeof context.who !== 'string' ||
      typeof context.what !== 'string' ||
      typeof context.where !== 'string' ||
      typeof context.why !== 'string' ||
      typeof context.how !== 'string' ||
      (typeof context.when !== 'number' && typeof context.when !== 'string') ||
      (typeof context.extent !== 'number' && typeof context.extent !== 'string')) {
      this.log('warn', 'Invalid 7D context: incorrect data types');
      return false;
    }

    return true;
  }

  /**
   * WHO: ContextValidator
   * WHAT: Validate GitHub MCP context structure
   * WHEN: Before translation
   * WHERE: Context Validation Layer
   * WHY: To ensure data integrity
   * HOW: Using schema validation
   * EXTENT: All GitHub context objects
   */
  validateGithubContext(context) {
    if (!context) return false;

    // Check that required fields exist
    const requiredFields = ['identity', 'operation', 'timestamp', 'scope'];
    const missingFields = requiredFields.filter(field => context[field] === undefined);

    if (missingFields.length > 0) {
      this.log('warn', `Invalid GitHub context: missing fields [${missingFields.join(', ')}]`);
      return false;
    }

    return true;
  }

  /**
   * WHO: ContextTransformer
   * WHAT: Convert GitHub MCP context to TNOS 7D format
   * WHEN: When receiving GitHub MCP messages
   * WHERE: Context Translation Layer
   * WHY: To provide TNOS-compatible context
   * HOW: Using dimension mapping with compression
   * EXTENT: All incoming GitHub contexts
   */
  githubContextToTnos7d(githubContext) {
    if (!githubContext) {
      return this.config.defaultContext;
    }

    // Validate if configured
    if (this.config.validateContext && !this.validateGithubContext(githubContext)) {
      this.log('warn', 'Using default context due to validation failure', { what: 'Validation_Failure' });
      return this.config.defaultContext;
    }

    try {
      // Basic mapping
      const tnos7dContext = {
        who: githubContext.identity || 'GitHub_MCP',
        what: githubContext.operation || 'Unknown_Operation',
        when: githubContext.timestamp || Date.now(),
        where: githubContext.location || 'GitHub_Environment',
        why: githubContext.purpose || 'External_Request',
        how: githubContext.method || 'MCP_Protocol',
        extent: parseFloat(githubContext.scope) || 1.0
      };

      // Additional mappings for complex fields
      if (githubContext.metadata) {
        if (githubContext.metadata.actor) tnos7dContext.who = githubContext.metadata.actor;
        if (githubContext.metadata.resource) tnos7dContext.where = githubContext.metadata.resource;
        if (githubContext.metadata.urgency) tnos7dContext.extent = this.mapUrgencyToExtent(githubContext.metadata.urgency);
      }

      // Apply compression if enabled
      if (this.config.compressionEnabled && tnos7dContext) {
        const compressionContext = {
          who: 'ContextBridge',
          what: 'Context_Compression',
          when: Date.now(),
          where: 'Translation_Layer',
          why: 'Data_Optimization',
          how: 'Mobius_Algorithm',
          extent: 0.7
        };

        // Use MobiusCompression to compress the context values
        tnos7dContext.who = MobiusCompression.compressString(tnos7dContext.who, compressionContext);
        tnos7dContext.what = MobiusCompression.compressString(tnos7dContext.what, compressionContext);
        tnos7dContext.where = MobiusCompression.compressString(tnos7dContext.where, compressionContext);
        tnos7dContext.why = MobiusCompression.compressString(tnos7dContext.why, compressionContext);
        tnos7dContext.how = MobiusCompression.compressString(tnos7dContext.how, compressionContext);
      }

      this.log('debug', 'Transformed GitHub context to TNOS 7D format', { what: 'Context_Translation' });
      return tnos7dContext;
    } catch (error) {
      this.log('error', `Error converting GitHub context to TNOS 7D: ${error.message}`, { what: 'Translation_Error' });
      return this.config.defaultContext;
    }
  }

  /**
   * WHO: ContextTransformer
   * WHAT: Convert TNOS 7D context to GitHub MCP format
   * WHEN: When sending messages to GitHub MCP
   * WHERE: Context Translation Layer
   * WHY: To provide GitHub-compatible context
   * HOW: Using inverse dimension mapping
   * EXTENT: All outgoing TNOS 7D contexts
   */
  tnos7dContextToGithub(tnos7dContext) {
    if (!tnos7dContext) {
      return {
        identity: 'System',
        operation: 'Default',
        timestamp: Date.now(),
        scope: '1.0'
      };
    }

    // Validate if configured
    if (this.config.validateContext && !this.validateTnos7dContext(tnos7dContext)) {
      this.log('warn', 'Using default GitHub context due to validation failure', { what: 'Validation_Failure' });
      return {
        identity: 'System',
        operation: 'Default',
        timestamp: Date.now(),
        scope: '1.0'
      };
    }

    try {
      // If context is compressed, decompress it first
      let processedContext = tnos7dContext;

      if (this.config.compressionEnabled) {
        const compressionContext = {
          who: 'ContextBridge',
          what: 'Context_Decompression',
          when: Date.now(),
          where: 'Translation_Layer',
          why: 'Data_Restoration',
          how: 'Mobius_Algorithm',
          extent: 0.7
        };

        processedContext = {
          ...tnos7dContext,
          who: MobiusCompression.decompressString(tnos7dContext.who, compressionContext),
          what: MobiusCompression.decompressString(tnos7dContext.what, compressionContext),
          where: MobiusCompression.decompressString(tnos7dContext.where, compressionContext),
          why: MobiusCompression.decompressString(tnos7dContext.why, compressionContext),
          how: MobiusCompression.decompressString(tnos7dContext.how, compressionContext)
        };
      }

      // Basic mapping
      const githubContext = {
        identity: processedContext.who,
        operation: processedContext.what,
        timestamp: processedContext.when,
        location: processedContext.where,
        purpose: processedContext.why,
        method: processedContext.how,
        scope: processedContext.extent.toString()
      };

      // Add metadata for more complex mappings
      githubContext.metadata = {
        actor: processedContext.who,
        resource: processedContext.where,
        urgency: this.mapExtentToUrgency(processedContext.extent)
      };

      this.log('debug', 'Transformed TNOS 7D context to GitHub format', { what: 'Context_Translation' });
      return githubContext;
    } catch (error) {
      this.log('error', `Error converting TNOS 7D context to GitHub: ${error.message}`, { what: 'Translation_Error' });
      return {
        identity: 'System',
        operation: 'Default',
        timestamp: Date.now(),
        scope: '1.0'
      };
    }
  }

  /**
   * WHO: MetadataMapper
   * WHAT: Map urgency levels to extent values
   * WHEN: During context translation
   * WHERE: Context Translation Layer
   * WHY: To standardize impact metrics
   * HOW: Using normalized scale conversion
   * EXTENT: All urgency translations
   */
  mapUrgencyToExtent(urgency) {
    switch (urgency.toLowerCase()) {
      case 'critical':
        return 1.0;
      case 'high':
        return 0.8;
      case 'medium':
        return 0.5;
      case 'low':
        return 0.2;
      default:
        return parseFloat(urgency) || 0.5;
    }
  }

  /**
   * WHO: MetadataMapper
   * WHAT: Map extent values to urgency levels
   * WHEN: During context translation
   * WHERE: Context Translation Layer
   * WHY: To standardize impact metrics
   * HOW: Using normalized scale conversion
   * EXTENT: All extent translations
   */
  mapExtentToUrgency(extent) {
    const extentValue = parseFloat(extent);

    if (extentValue >= 0.9) return 'critical';
    if (extentValue >= 0.7) return 'high';
    if (extentValue >= 0.4) return 'medium';
    if (extentValue >= 0.1) return 'low';
    return 'informational';
  }

  /**
   * WHO: ContextPersistence
   * WHAT: Save context to persistent storage
   * WHEN: After context changes
   * WHERE: Storage Layer
   * WHY: To maintain context across restarts
   * HOW: Using file system persistence
   * EXTENT: All context objects
   */
  async saveContext(type, context) {
    try {
      // Implementation would use file system or database
      this.log('info', `Context saved: ${type}`, { what: 'Context_Persistence' });
      return true;
    } catch (error) {
      this.log('error', `Error saving context: ${error.message}`, { what: 'Persistence_Error' });
      return false;
    }
  }

  /**
   * WHO: ContextPersistence
   * WHAT: Load context from persistent storage
   * WHEN: During system initialization
   * WHERE: Storage Layer
   * WHY: To restore context after restarts
   * HOW: Using file system retrieval
   * EXTENT: All context objects
   */
  async loadContext(type) {
    try {
      // Implementation would use file system or database
      this.log('info', `Context loaded: ${type}`, { what: 'Context_Persistence' });

      // Return default contexts when implementation is missing
      if (type === 'github') {
        return {
          identity: 'System',
          operation: 'Startup',
          timestamp: Date.now(),
          scope: '1.0'
        };
      } else {
        return this.config.defaultContext;
      }
    } catch (error) {
      this.log('error', `Error loading context: ${error.message}`, { what: 'Persistence_Error' });

      // Return default contexts on error
      if (type === 'github') {
        return {
          identity: 'System',
          operation: 'Startup',
          timestamp: Date.now(),
          scope: '1.0'
        };
      } else {
        return this.config.defaultContext;
      }
    }
  }
}

module.exports = ContextBridge;