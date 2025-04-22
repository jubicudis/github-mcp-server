/**
 * MobiusCompression.js - Standardized Möbius 7D Compression Implementation
 * 
 * This is the reference JavaScript implementation of the Möbius Compression
 * algorithm for Tranquility-Neuro-OS, following the standardized specification.
 */

/**
 * WHO: CompressionManager
 * WHAT: Implement Möbius 7D compression algorithm 
 * WHEN: During data transmission and storage operations
 * WHERE: System Layer 2 (Reactive)
 * WHY: To optimize data size while preserving context
 * HOW: Using the Möbius mathematical formula with 7D awareness
 * EXTENT: All compressible data across system boundaries
 */

// Ensure compatibility with both Node.js and browser environments
const isNode = typeof module !== 'undefined' && module.exports;

/**
 * ContextVector7D class - Represents a 7-dimensional context vector
 * 
 * WHO: ContextVector7D
 * WHAT: Store 7D context information
 * WHEN: During compression/decompression operations
 * WHERE: System Layer 2 (Reactive)
 * WHY: To provide context for compression operations
 * HOW: Using object properties for each dimension
 * EXTENT: All context-aware compression
 */
class ContextVector7D {
  constructor(who, what, when, where, why, how, extent) {
    this.who = who;
    this.what = what;
    this.when = when;
    this.where = where;
    this.why = why;
    this.how = how;
    this.extent = extent;
  }

  // Get serialized representation
  toString() {
    return JSON.stringify({
      who: this.who,
      what: this.what,
      when: this.when,
      where: this.where,
      why: this.why,
      how: this.how,
      extent: this.extent
    });
  }

  // Clone the context vector
  clone() {
    return new ContextVector7D(
      this.who, this.what, this.when, this.where,
      this.why, this.how, this.extent
    );
  }
}

/**
 * MobiusCompression class - Implements the standardized Möbius Compression algorithm
 * 
 * WHO: MobiusCompression
 * WHAT: Compress and decompress data
 * WHEN: During data transmission and storage
 * WHERE: System Layer 2 (Reactive)
 * WHY: To reduce data size while preserving integrity
 * HOW: Using the Möbius compression formula with context factors
 * EXTENT: All data exchanged between systems
 */
class MobiusCompression {
  /**
   * Initialize the Möbius Compression module
   * @param {Object} config - Configuration options
   */
  static initialize(config = {}) {
    MobiusCompression.config = {
      defaultFactors: {
        B: 1.5,  // Base factor for who/what dimensions
        V: 2.3,  // Variance factor for where dimensions
        I: 1.2,  // Intent factor for why dimension
        G: 0.8,  // Temporal gradient for when dimension
        F: 1.1,  // Fidelity factor for extent dimension
        E: 0.5,  // Entropy coefficient
        C_sum: 0.7 // Cumulative context factor
      },
      version: "1.0",
      algorithm: "mobius7d",
      useTimeFactor: true,
      useEnergyFactor: true,
      contextAware: true,
      ...config
    };

    // Statistics tracking
    MobiusCompression.stats = {
      compressions: 0,
      decompressions: 0,
      averageCompressionRatio: 0,
      totalOriginalSize: 0,
      totalCompressedSize: 0,
      averageEntropyReduction: 0,
      errors: 0,
      recoveries: 0
    };

    // Log initialization
    if (isNode) {
      console.log(`[MobiusCompression] Initialized with compression-first approach (v${MobiusCompression.config.version})`);
    }

    return MobiusCompression.config;
  }

  /**
   * Compress data using the Möbius Compression algorithm
   * 
   * WHO: MobiusCompression
   * WHAT: Compress data with context awareness
   * WHEN: When data needs to be transmitted or stored
   * WHERE: System Layer 2 (Reactive)
   * WHY: To reduce data size with lossless recovery
   * HOW: Applying the Möbius formula with entropy calculation
   * EXTENT: All compression operations
   * 
   * @param {*} data - Data to compress
   * @param {Object} options - Compression options including context
   * @returns {Object} Compressed data package with metadata
   */
  static compress(data, options = {}) {
    try {
      // Ensure initialization
      if (!MobiusCompression.config) {
        MobiusCompression.initialize();
      }

      // Track statistics
      MobiusCompression.stats.compressions++;

      // Extract context from options
      const context = options.context || {
        who: "MobiusCompression",
        what: "DataCompression",
        when: Date.now(),
        where: "Layer2_Bridge",
        why: "DataOptimization",
        how: "MobiusFormula",
        extent: 1.0
      };

      // Normalize context
      const ctx = MobiusCompression._normalizeContext(context);

      // Calculate numeric representation of the data
      const value = MobiusCompression._getNumericRepresentation(data);

      // Calculate entropy
      const entropy = MobiusCompression._calculateEntropy(data);

      // Extract compression factors based on context
      const factors = MobiusCompression._extractContextFactors(ctx);
      const { B, V, I, G, F, E, t, C_sum } = factors;

      // Calculate alignment
      const alignment = (B + V * I) * Math.exp(-t * E);

      // Apply the Möbius Compression Formula
      const entropyFactor = 1 - (entropy / Math.log2(1 + V));
      const numerator = value * B * I * entropyFactor * (G + F);
      const denominator = E * t + C_sum * entropy + alignment;

      // Guard against division by zero
      if (Math.abs(denominator) < 1e-10) {
        throw new Error("Compression denominator too close to zero");
      }

      const compressed = numerator / denominator;

      // Package with metadata
      const compressionVars = { value, entropy, ...factors, alignment };
      const result = MobiusCompression._packageWithMetadata(compressed, compressionVars, ctx, data);

      // Update statistics
      MobiusCompression._updateCompressionStats(data, result);

      return result;
    } catch (error) {
      MobiusCompression.stats.errors++;
      console.error(`[MobiusCompression] Compression error: ${error.message}`);

      // Fallback: return original data with error metadata
      return {
        success: false,
        error: error.message,
        data: data,
        metadata: {
          algorithm: MobiusCompression.config.algorithm,
          version: MobiusCompression.config.version,
          error: error.message,
          timestamp: Date.now()
        }
      };
    }
  }

  /**
   * Decompress previously compressed data
   * 
   * WHO: MobiusCompression
   * WHAT: Decompress data with context-awareness
   * WHEN: When compressed data needs to be restored
   * WHERE: System Layer 2 (Reactive)
   * WHY: To restore original data from compressed form
   * HOW: Applying the inverse Möbius formula
   * EXTENT: All decompression operations
   * 
   * @param {string|Object} compressedData - Compressed data string or object
   * @param {Object} options - Decompression options
   * @returns {*} Original data (or best approximation)
   */
  static decompress(compressedData, options = {}) {
    try {
      // Ensure initialization
      if (!MobiusCompression.config) {
        MobiusCompression.initialize();
      }

      // Parse input if it's a string
      let compressed, metadata;

      if (typeof compressedData === 'string') {
        try {
          // Try to parse as JSON
          const parsed = JSON.parse(compressedData);
          compressed = parsed.data || parsed.compressed;
          metadata = parsed.metadata;
        } catch (e) {
          // Not JSON, log the parsing error and treat as compressed string
          console.debug(`[MobiusCompression] Input not valid JSON, treating as compressed string: ${e.message}`);
          compressed = compressedData;
          metadata = options.metadata || {};
        }
      } else if (typeof compressedData === 'object') {
        compressed = compressedData.data || compressedData.compressed;
        metadata = compressedData.metadata || {};
      } else {
        compressed = compressedData;
        metadata = options.metadata || {};
      }

      // Track statistics
      MobiusCompression.stats.decompressions++;

      // Handle error cases
      if (compressed === null && metadata.original !== undefined) {
        MobiusCompression.stats.recoveries++;
        return metadata.original;
      }

      // Check if we have the required metadata
      if (!metadata.compressionVars) {
        throw new Error("Missing compression variables in metadata");
      }

      // Extract variables from metadata
      const { entropy, B, V, I, G, F, E, t, C_sum, alignment } = metadata.compressionVars;

      // Apply the inverse formula
      const entropyFactor = 1 - (entropy / Math.log2(1 + V));
      const numerator = compressed * (E * t + C_sum * entropy + alignment);
      const denominator = B * I * entropyFactor * (G + F);

      // Guard against division by zero
      if (Math.abs(denominator) < 1e-10) {
        throw new Error("Decompression denominator too close to zero");
      }

      const decompressedValue = numerator / denominator;

      // Convert back to original data type
      return MobiusCompression._valueToOriginalType(decompressedValue, metadata);
    } catch (error) {
      MobiusCompression.stats.errors++;
      console.error(`[MobiusCompression] Decompression error: ${error.message}`);

      // If metadata contains the original, use it as fallback
      if (metadata && metadata.original !== undefined) {
        MobiusCompression.stats.recoveries++;
        return metadata.original;
      }

      // Last resort fallback: return the compressed value
      return compressedData;
    }
  }

  /**
   * Calculate Shannon entropy of the input data
   * @param {*} data - Input data to calculate entropy for
   * @returns {number} Entropy value
   * @private
   */
  static _calculateEntropy(data) {
    // Convert data to string if not already
    const stringData = typeof data === 'string' ? data : JSON.stringify(data);

    // Count character frequencies
    const frequencies = {};
    for (const char of stringData) {
      frequencies[char] = (frequencies[char] || 0) + 1;
    }

    // Calculate entropy
    let entropy = 0;
    for (const char in frequencies) {
      const probability = frequencies[char] / stringData.length;
      entropy -= probability * Math.log2(probability);
    }

    return entropy;
  }

  /**
   * Extract contextual factors based on the context vector
   * @param {ContextVector7D} context - The 7D context vector
   * @returns {Object} The extracted factors
   * @private
   */
  static _extractContextFactors(context) {
    // Extract the default factors
    const { B, V, I, G, F, E, C_sum } = MobiusCompression.config.defaultFactors;

    // Enhance with context-specific adjustments
    const B_adjusted = MobiusCompression._getContextFactor('B', context.who, context.what, B);
    const V_adjusted = MobiusCompression._getContextFactor('V', context.where, null, V);
    const I_adjusted = MobiusCompression._getContextFactor('I', context.why, null, I);
    const G_adjusted = MobiusCompression._getContextFactor('G', context.when, null, G);
    const F_adjusted = MobiusCompression._getContextFactor('F', context.extent, null, F);
    const E_adjusted = MobiusCompression._getContextFactor('E', context, null, E);
    const t = MobiusCompression._getTemporalFactor(context.when);
    const C_sum_adjusted = MobiusCompression._getCSumFactor(context, C_sum);

    return {
      B: B_adjusted,
      V: V_adjusted,
      I: I_adjusted,
      G: G_adjusted,
      F: F_adjusted,
      E: E_adjusted,
      t,
      C_sum: C_sum_adjusted
    };
  }

  /**
   * Convert data to numeric representation
   * @param {*} data - Data to represent numerically
   * @returns {number} Numeric representation
   * @private
   */
  static _getNumericRepresentation(data) {
    if (typeof data === 'number') {
      return data;
    }

    if (typeof data === 'string') {
      // Use string length and character codes for a numeric representation
      let value = data.length;
      for (let i = 0; i < Math.min(data.length, 100); i++) {
        value += data.charCodeAt(i) / (i + 1);
      }
      return value;
    }

    if (typeof data === 'boolean') {
      return data ? 1 : 0;
    }

    if (data === null || data === undefined) {
      return 0;
    }

    if (Array.isArray(data)) {
      return data.length + MobiusCompression._calculateEntropy(data);
    }

    if (typeof data === 'object') {
      const jsonStr = JSON.stringify(data);
      return jsonStr.length + MobiusCompression._calculateEntropy(jsonStr);
    }

    // Default fallback
    return 1;
  }

  /**
   * Package the compressed value with metadata
   * @param {number} compressed - Compressed value
   * @param {Object} compressionVars - Variables used in compression
   * @param {ContextVector7D} context - Context vector
   * @param {*} originalData - Original data for fallback
   * @returns {Object} Packaged data with metadata
   * @private
   */
  static _packageWithMetadata(compressed, compressionVars, context, originalData) {
    // Determine original data type
    const originalType = typeof originalData;
    const originalSize = MobiusCompression._getDataSize(originalData);
    const compressedSize = String(compressed).length;
    const compressionRatio = originalSize / compressedSize;

    return {
      success: true,
      data: compressed,
      compressionRatio,
      originalSize,
      compressedSize,
      metadata: {
        algorithm: MobiusCompression.config.algorithm,
        version: MobiusCompression.config.version,
        timestamp: Date.now(),
        originalType,
        compressionVars,
        context: {
          who: context.who,
          what: context.what,
          when: context.when,
          where: context.where,
          why: context.why,
          how: context.how,
          extent: context.extent
        },
        original: MobiusCompression.config.preserveOriginal ? originalData : undefined
      }
    };
  }

  /**
   * Get the size of data in bytes (approximate)
   * @param {*} data - Data to measure
   * @returns {number} Size in bytes
   * @private
   */
  static _getDataSize(data) {
    if (typeof data === 'string') {
      return data.length;
    }

    if (typeof data === 'number') {
      return String(data).length;
    }

    if (typeof data === 'boolean') {
      return 1;
    }

    if (data === null || data === undefined) {
      return 0;
    }

    if (Array.isArray(data) || typeof data === 'object') {
      return JSON.stringify(data).length;
    }

    return String(data).length;
  }

  /**
   * Normalize context parameter to ContextVector7D
   * @param {*} context - Context parameter
   * @returns {ContextVector7D} Normalized context vector
   * @private
   */
  static _normalizeContext(context) {
    // If context is already a ContextVector7D, return it
    if (context instanceof ContextVector7D) {
      return context;
    }

    // If context is an object with 7D properties, convert it
    if (context && typeof context === 'object') {
      return new ContextVector7D(
        context.who || null,
        context.what || null,
        context.when || null,
        context.where || null,
        context.why || null,
        context.how || null,
        context.extent || null
      );
    }

    // Default empty context
    return new ContextVector7D(null, null, null, null, null, null, null);
  }

  /**
   * Get context factor based on dimension values
   * @param {string} factor - The factor name
   * @param {*} primary - Primary dimension value
   * @param {*} secondary - Secondary dimension value
   * @param {number} defaultValue - Default value if not found
   * @returns {number} The context factor
   * @private
   */
  static _getContextFactor(factor, primary, secondary, defaultValue) {
    // Default for secondary if not provided
    secondary = secondary || null;
    // This would typically query a factor registry
    // For now, use a simple algorithm based on input values
    let base = defaultValue;

    // Adjust by string length or numeric value to simulate context awareness
    if (typeof primary === 'string') {
      base += (primary.length % 10) * 0.1;
    } else if (typeof primary === 'number') {
      base += (primary % 10) * 0.1;
    }

    // Further adjust by secondary value if present
    if (secondary !== null) {
      if (typeof secondary === 'string') {
        base += (secondary.length % 5) * 0.05;
      } else if (typeof secondary === 'number') {
        base += (secondary % 5) * 0.05;
      }
    }

    return base;
  }

  /**
   * Get temporal factor based on when dimension
   * @param {*} when - The temporal value
   * @returns {number} The temporal factor
   * @private
   */
  static _getTemporalFactor(when) {
    // Convert to a number representing time
    let t = 1.0;

    if (typeof when === 'number') {
      // Use the value directly
      t = Math.max(0.1, when);
    } else if (when instanceof Date) {
      // Use timestamp
      t = when.getTime() / 1000;
    } else if (typeof when === 'string') {
      // Try to parse as date or use string length
      const timestamp = Date.parse(when);
      if (!isNaN(timestamp)) {
        t = timestamp / 1000;
      } else {
        t = when.length * 0.1;
      }
    }

    // Normalize to a reasonable range
    return Math.max(0.1, Math.min(10, t / 1000));
  }

  /**
   * Calculate cumulative context factor
   * @param {ContextVector7D} context - Context vector
   * @param {number} defaultValue - Default factor value
   * @returns {number} Calculated C_sum factor
   * @private
   */
  static _getCSumFactor(context, defaultValue) {
    let factor = defaultValue;
    let dimensionCount = 0;

    // Count how many dimensions have values
    if (context.who) dimensionCount++;
    if (context.what) dimensionCount++;
    if (context.when) dimensionCount++;
    if (context.where) dimensionCount++;
    if (context.why) dimensionCount++;
    if (context.how) dimensionCount++;
    if (context.extent) dimensionCount++;

    // Adjust factor based on information completeness
    factor *= (1 + (dimensionCount / 7) * 0.5);

    return factor;
  }

  /**
   * Convert numeric value back to original data type
   * @param {number} value - Numeric value
   * @param {Object} metadata - Compression metadata
   * @returns {*} Original typed data
   * @private
   */
  static _valueToOriginalType(value, metadata) {
    // This is a best-effort conversion and may not be lossless
    const { originalType } = metadata;

    switch (originalType) {
      case 'number':
        return value;
      case 'boolean':
        return value >= 0.5;
      case 'string':
        // For strings, if we have the original, return it
        if (metadata.original && typeof metadata.original === 'string') {
          return metadata.original;
        }
        // Otherwise return a placeholder
        return `[Decompressed value: ${value.toFixed(2)}]`;
      case 'object':
        // For objects, if we have the original, return it
        if (metadata.original && typeof metadata.original === 'object') {
          return metadata.original;
        }
        // Otherwise return a placeholder object
        return {
          _decompressed: value,
          _type: 'decompressed_object',
          _timestamp: Date.now()
        };
      default:
        return value;
    }
  }

  /**
   * Update statistics after compression
   * @param {*} original - Original data
   * @param {Object} compressed - Compressed package
   * @private
   */
  static _updateCompressionStats(original, compressedPackage) {
    const originalSize = MobiusCompression._getDataSize(original);
    const compressedSize = MobiusCompression._getDataSize(compressedPackage.data);

    MobiusCompression.stats.totalOriginalSize += originalSize;
    MobiusCompression.stats.totalCompressedSize += compressedSize;

    // Update average compression ratio
    const ratio = originalSize / compressedSize;
    const prevCompressions = MobiusCompression.stats.compressions - 1;

    if (prevCompressions > 0) {
      MobiusCompression.stats.averageCompressionRatio = (
        (MobiusCompression.stats.averageCompressionRatio * prevCompressions) + ratio
      ) / MobiusCompression.stats.compressions;
    } else {
      MobiusCompression.stats.averageCompressionRatio = ratio;
    }
  }

  /**
   * Get statistics about compression operations
   * @returns {Object} Statistics
   */
  static getStatistics() {
    return { ...MobiusCompression.stats };
  }

  /**
   * Reset compression statistics
   * WHO: MobiusCompression
   * WHAT: Reset statistics
   * WHEN: When stats need to be cleared
   * WHERE: MobiusCompression module
   * WHY: To start fresh statistics tracking
   * HOW: Preserving structure and resetting counters
   * EXTENT: All tracked statistics
   */
  static resetStatistics() {
    for (const key in MobiusCompression.stats) {
      if (!key.startsWith('average')) {
        MobiusCompression.stats[key] = 0; // Only reset non-average statistics
      }
    }
  }
}

// Initialize the static configuration with defaults
MobiusCompression.initialize();

// Export for Node.js or assign to window for browser
if (isNode) {
  module.exports = { MobiusCompression, ContextVector7D };
} else {
  window.MobiusCompression = MobiusCompression;
  window.ContextVector7D = ContextVector7D;
}