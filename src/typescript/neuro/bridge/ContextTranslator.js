/**
 * ContextTranslator.js - Translation layer between GitHub MCP and TNOS 7D context
 * 
 * WHO: ContextTranslator
 * WHAT: Bidirectional context translation between systems
 * WHEN: During inter-system communication
 * WHERE: System Bridge Layer
 * WHY: To ensure context preservation across systems
 * HOW: Using context mapping and transformation rules
 * EXTENT: All cross-system communications
 */

const { log } = require('./BridgeLogger');

/**
 * WHO: ContextTranslator
 * WHAT: Context translation service for MCP Bridge
 * WHEN: During message exchange
 * WHERE: System Bridge Layer
 * WHY: To maintain context integrity across systems
 * HOW: Using bidirectional context mapping
 * EXTENT: All cross-system messages
 */
class ContextTranslator {
  /**
   * WHO: ContextInitializer
   * WHAT: Initialize the context translator
   * WHEN: During system startup
   * WHERE: System Bridge Layer
   * WHY: To configure mapping rules
   * HOW: Using configuration settings
   * EXTENT: Translation service configuration
   */
  constructor(config) {
    /**
     * WHO: Configuration
     * WHAT: Translation system configuration
     * WHEN: At system initialization
     * WHERE: System Bridge Layer
     * WHY: To customize translation behavior
     * HOW: Using configurable parameters
     * EXTENT: Translation system configuration
     */
    this.config = config || {
      defaultContext: {
        who: 'MCP_Bridge',
        what: 'Context_Translation',
        when: 'System_Operation',
        where: 'Bridge_Layer',
        why: 'Context_Preservation',
        how: 'BiDirectional_Mapping',
        extent: 'Cross_System'
      },
      contextMappings: {
        // GitHub MCP field to TNOS 7D context dimension mappings
        github: {
          identity: 'who',
          operation: 'what',
          timestamp: 'when',
          location: 'where',
          purpose: 'why',
          method: 'how',
          scope: 'extent'
        },
        // TNOS 7D context dimension to GitHub MCP field mappings
        tnos: {
          who: 'identity',
          what: 'operation',
          when: 'timestamp',
          where: 'location',
          why: 'purpose',
          how: 'method',
          extent: 'scope'
        }
      },
      // Value transformations for specific fields
      valueTransformations: {
        github: {
          // Convert GitHub role types to TNOS actor types
          identity: {
            'user': 'Human_User',
            'system': 'GitHub_System',
            'bot': 'GitHub_Bot',
            'copilot': 'GitHub_Copilot',
            'action': 'GitHub_Action'
          },
          // Map operation types
          operation: {
            'read': 'Query',
            'write': 'Modify',
            'create': 'Create',
            'delete': 'Delete',
            'update': 'Update',
            'execute': 'Execute'
          }
        },
        tnos: {
          // Convert TNOS actor types to GitHub role types
          who: {
            'Human_User': 'user',
            'GitHub_System': 'system',
            'GitHub_Bot': 'bot',
            'GitHub_Copilot': 'copilot',
            'GitHub_Action': 'action',
            'System': 'system',
            'TNOS_Core': 'system'
          },
          // Map operation types
          what: {
            'Query': 'read',
            'Modify': 'write',
            'Create': 'create',
            'Delete': 'delete',
            'Update': 'update',
            'Execute': 'execute'
          }
        }
      }
    };

    /**
     * WHO: Metrics
     * WHAT: Translation statistics
     * WHEN: Throughout system operation
     * WHERE: System Bridge Layer
     * WHY: For monitoring translation effectiveness
     * HOW: Using counters for various translation events
     * EXTENT: All translation operations
     */
    this.metrics = {
      githubToTnosTranslations: 0,
      tnosToGithubTranslations: 0,
      fallbacksApplied: 0,
      transformationsApplied: 0,
      unknownContextFields: 0,
      lastTranslationTime: 0
    };

    log('info', 'ContextTranslator initialized');
  }

  /**
   * WHO: GitHubTranslator
   * WHAT: Translate GitHub MCP context to TNOS 7D context
   * WHEN: When receiving messages from GitHub MCP
   * WHERE: System Bridge Layer
   * WHY: To convert external to internal context format
   * HOW: Using mapping rules and transformation logic
   * EXTENT: GitHub to TNOS messages
   */
  translateGithubToTnos(githubContext) {
    const startTime = Date.now();

    // Initialize with default TNOS context
    const tnosContext = { ...this.config.defaultContext };

    if (!githubContext || typeof githubContext !== 'object') {
      log('warn', 'Invalid GitHub context provided for translation, using defaults');
      this.metrics.fallbacksApplied++;
      this.metrics.lastTranslationTime = Date.now() - startTime;
      return tnosContext;
    }

    // Map GitHub fields to TNOS dimensions
    for (const [githubField, value] of Object.entries(githubContext)) {
      const tnosDimension = this.config.contextMappings.github[githubField];

      if (tnosDimension) {
        // Apply value transformation if defined
        if (this.config.valueTransformations.github[githubField] &&
          this.config.valueTransformations.github[githubField][value]) {
          tnosContext[tnosDimension] = this.config.valueTransformations.github[githubField][value];
          this.metrics.transformationsApplied++;
        } else {
          tnosContext[tnosDimension] = value;
        }
      } else {
        // Store unknown fields in metadata to preserve information
        if (!tnosContext.metadata) {
          tnosContext.metadata = {};
        }
        tnosContext.metadata[githubField] = value;
        this.metrics.unknownContextFields++;
      }
    }

    // Ensure timestamp is in ISO format
    if (tnosContext.when && typeof tnosContext.when === 'number') {
      tnosContext.when = new Date(tnosContext.when).toISOString();
    } else if (!tnosContext.when) {
      tnosContext.when = new Date().toISOString();
      this.metrics.fallbacksApplied++;
    }

    // Post-processing and validation
    this._validateAndNormalizeTnosContext(tnosContext);

    this.metrics.githubToTnosTranslations++;
    this.metrics.lastTranslationTime = Date.now() - startTime;

    log('debug', `Translated GitHub context to TNOS 7D context in ${this.metrics.lastTranslationTime}ms`);
    return tnosContext;
  }

  /**
   * WHO: TNOSTranslator
   * WHAT: Translate TNOS 7D context to GitHub MCP context
   * WHEN: When sending messages to GitHub MCP
   * WHERE: System Bridge Layer
   * WHY: To convert internal to external context format
   * HOW: Using mapping rules and transformation logic
   * EXTENT: TNOS to GitHub messages
   */
  translateTnosToGithub(tnosContext) {
    const startTime = Date.now();

    // Initialize with minimal GitHub context
    const githubContext = {
      version: '1.0',
      timestamp: Date.now()
    };

    if (!tnosContext || typeof tnosContext !== 'object') {
      log('warn', 'Invalid TNOS context provided for translation, using defaults');
      this.metrics.fallbacksApplied++;
      this.metrics.lastTranslationTime = Date.now() - startTime;
      return githubContext;
    }

    // Map TNOS dimensions to GitHub fields
    for (const [tnosDimension, value] of Object.entries(tnosContext)) {
      // Skip metadata field which is our custom addition
      if (tnosDimension === 'metadata') continue;

      const githubField = this.config.contextMappings.tnos[tnosDimension];

      if (githubField) {
        // Apply value transformation if defined
        if (this.config.valueTransformations.tnos[tnosDimension] &&
          this.config.valueTransformations.tnos[tnosDimension][value]) {
          githubContext[githubField] = this.config.valueTransformations.tnos[tnosDimension][value];
          this.metrics.transformationsApplied++;
        } else {
          githubContext[githubField] = value;
        }
      } else if (tnosDimension !== 'metadata') {
        // Store unknown dimensions in custom field to preserve information
        if (!githubContext.customContext) {
          githubContext.customContext = {};
        }
        githubContext.customContext[tnosDimension] = value;
        this.metrics.unknownContextFields++;
      }
    }

    // Handle metadata field if present, merging it into customContext
    if (tnosContext.metadata && typeof tnosContext.metadata === 'object') {
      if (!githubContext.customContext) {
        githubContext.customContext = {};
      }
      Object.assign(githubContext.customContext, tnosContext.metadata);
    }

    // Convert temporal context to timestamp if needed
    if (tnosContext.when && typeof tnosContext.when === 'string') {
      try {
        githubContext.timestamp = new Date(tnosContext.when).getTime();
      } catch (e) {
        // Keep default timestamp if parsing fails
        log('warn', `Failed to parse TNOS temporal context: ${tnosContext.when}`);
        this.metrics.fallbacksApplied++;
      }
    }

    this.metrics.tnosToGithubTranslations++;
    this.metrics.lastTranslationTime = Date.now() - startTime;

    log('debug', `Translated TNOS 7D context to GitHub context in ${this.metrics.lastTranslationTime}ms`);
    return githubContext;
  }

  /**
   * WHO: ContextMetricsProvider
   * WHAT: Get translation metrics
   * WHEN: During system monitoring
   * WHERE: System Bridge Layer
   * WHY: For performance monitoring
   * HOW: Using accumulated statistics
   * EXTENT: All translation operations
   */
  getMetrics() {
    return {
      ...this.metrics,
      timestamp: Date.now()
    };
  }

  /**
   * WHO: ContextValidator
   * WHAT: Validate and normalize TNOS context
   * WHEN: After translation from GitHub
   * WHERE: System Bridge Layer
   * WHY: To ensure context integrity
   * HOW: Using validation rules and normalization
   * EXTENT: Translated context objects
   */
  _validateAndNormalizeTnosContext(tnosContext) {
    // Ensure all 7D dimensions are present
    const requiredDimensions = ['who', 'what', 'when', 'where', 'why', 'how', 'extent'];

    for (const dimension of requiredDimensions) {
      if (!tnosContext[dimension]) {
        tnosContext[dimension] = this.config.defaultContext[dimension];
        this.metrics.fallbacksApplied++;
        log('debug', `Applied fallback for missing TNOS context dimension: ${dimension}`);
      }
    }

    // Normalize values to snake_case format for consistency
    for (const [dimension, value] of Object.entries(tnosContext)) {
      if (typeof value === 'string') {
        // Replace spaces with underscores for consistency
        tnosContext[dimension] = value.replace(/\s+/g, '_');
      }
    }

    return tnosContext;
  }

  /**
   * WHO: BridgeContext
   * WHAT: Create a bridge-specific context
   * WHEN: For bridge operations
   * WHERE: System Bridge Layer
   * WHY: To generate standardized bridge contexts
   * HOW: Using context templates with operation specifics
   * EXTENT: Bridge operations
   */
  createBridgeContext(operation, details = {}) {
    return {
      who: 'MCP_Bridge',
      what: operation,
      when: new Date().toISOString(),
      where: 'Bridge_Layer',
      why: details.purpose || 'System_Operation',
      how: details.method || 'Context_Bridge',
      extent: details.scope || 'Bridge_Operation',
      metadata: details.metadata || {}
    };
  }
}

module.exports = { ContextTranslator };