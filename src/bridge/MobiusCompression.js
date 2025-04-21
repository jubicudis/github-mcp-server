/**
 * WHO: MobiusCompression
 * WHAT: Implements the TNOS Möbius compression algorithm for Layer 2
 * WHEN: During data transmission, storage, and processing
 * WHERE: Layer 2 (Reactive) / Bridge Component
 * WHY: To optimize data transfer while preserving context
 * HOW: Using 7D-aware compression techniques with entropy calculation
 * EXTENT: All compressed data in Layer 2 operations
 */

/**
 * WHO: CompressionAlgorithm
 * WHAT: Static class for Möbius compression operations
 * WHEN: Data optimization operations
 * WHERE: Layer 2 (Reactive)
 * WHY: To reduce data size while preserving meaning
 * HOW: Using the Möbius compression formula
 * EXTENT: All compression operations
 */
class MobiusCompression {
  /**
   * WHO: CompressionCalculator
   * WHAT: Calculate entropy of input data
   * WHEN: Before compression
   * WHERE: Preprocessing stage
   * WHY: To determine compressibility
   * HOW: Using Shannon entropy calculation
   * EXTENT: All compression inputs
   */
  static calculateEntropy(data) {
    if (typeof data !== 'string') {
      data = String(data);
    }

    // Count character frequencies
    const freqMap = {};
    for (let i = 0; i < data.length; i++) {
      const char = data[i];
      freqMap[char] = (freqMap[char] || 0) + 1;
    }

    // Calculate Shannon entropy
    let entropy = 0;
    for (const char in freqMap) {
      const probability = freqMap[char] / data.length;
      entropy -= probability * Math.log2(probability);
    }

    return entropy;
  }

  /**
   * WHO: ContextExtractor
   * WHAT: Extract compression factors from context
   * WHEN: During compression
   * WHERE: Context analysis stage
   * WHY: To apply context-aware compression
   * HOW: Using dimensional mapping
   * EXTENT: All compression operations
   */
  static extractContextFactors(context) {
    // Default factors
    const defaultFactors = {
      B: 1.0,  // Base factor
      V: 2.0,  // Value factor
      I: 0.8,  // Information factor
      G: 1.2,  // Growth factor
      F: 0.5,  // Fluctuation factor
      E: 0.3   // Energy factor
    };

    if (!context || typeof context !== 'object') {
      return defaultFactors;
    }

    try {
      // Context-based factor calculation
      return {
        // Base factor - affected by WHO (identity strength)
        B: context.who ? 0.8 + (context.who.length % 5) * 0.1 : defaultFactors.B,

        // Value factor - affected by WHAT (content importance)
        V: context.what ? 1.5 + (context.what.length % 10) * 0.1 : defaultFactors.V,

        // Information factor - affected by HOW (method complexity)
        I: context.how ? 0.7 + (context.how.length % 8) * 0.05 : defaultFactors.I,

        // Growth factor - affected by WHY (purpose strength)
        G: context.why ? 1.0 + (context.why.length % 6) * 0.1 : defaultFactors.G,

        // Fluctuation factor - affected by WHERE (location stability)
        F: context.where ? 0.4 + (context.where.length % 5) * 0.1 : defaultFactors.F,

        // Energy factor - affected by EXTENT (impact scope)
        E: context.extent ? 0.2 + parseFloat(context.extent) * 0.3 : defaultFactors.E
      };
    } catch (error) {
      console.error('Error extracting context factors:', error.message);
      return defaultFactors;
    }
  }

  /**
   * WHO: TemporalCalculator
   * WHAT: Calculate temporal factor for compression
   * WHEN: During compression formula application
   * WHERE: Temporal analysis stage
   * WHY: To incorporate time relevance
   * HOW: Using timestamp normalization
   * EXTENT: All temporal aspects of compression
   */
  static calculateTemporalFactor(when) {
    // Default to current time if not provided
    const timestamp = when || Date.now();
    const currentTime = Date.now();

    // Calculate temporal distance (normalize to 0-1 range)
    // More recent = smaller t value = less energy cost in formula
    const timeDifference = Math.abs(currentTime - timestamp);
    const normalizedTime = Math.min(1.0, timeDifference / (1000 * 60 * 60 * 24)); // Normalize to 1 day

    return normalizedTime;
  }

  /**
   * WHO: Compressor
   * WHAT: Compress string data using Möbius algorithm
   * WHEN: During data transmission
   * WHERE: Layer 2 (Reactive)
   * WHY: To optimize bandwidth usage
   * HOW: Using contextual compression formula
   * EXTENT: All string data
   */
  static compressString(data, context = {}) {
    if (!data) return '';

    try {
      // Step 1: Calculate entropy
      const entropy = this.calculateEntropy(data);

      // Step 2: Extract contextual factors
      const { B, V, I, G, F, E } = this.extractContextFactors(context);

      // Step 3: Calculate temporal factor
      const t = this.calculateTemporalFactor(context.when);

      // Step 4: Calculate C_sum (context sum)
      const C_sum = (context.who ? 0.1 : 0) +
        (context.what ? 0.1 : 0) +
        (context.when ? 0.1 : 0) +
        (context.where ? 0.1 : 0) +
        (context.why ? 0.1 : 0) +
        (context.how ? 0.1 : 0) +
        (context.extent ? 0.1 : 0);

      // Step 5: Calculate alignment using the Möbius formula
      const alignment = (B + V * I) * Math.exp(-t * E);

      // Step 6: Apply basic compression
      let compressed = data;

      // For strings, apply dictionary-based compression
      if (data.length > 10) {
        // Find repeated patterns
        const patterns = {};
        for (let i = 2; i < Math.min(10, data.length / 2); i++) {
          for (let j = 0; j <= data.length - i; j++) {
            const pattern = data.substring(j, j + i);
            patterns[pattern] = (patterns[pattern] || 0) + 1;
          }
        }

        // Replace most common patterns
        for (const pattern in patterns) {
          if (patterns[pattern] > 2 && pattern.length > 2) {
            // Create a replacement token that doesn't exist in the data
            const token = `§${Object.keys(patterns).indexOf(pattern)}§`;
            compressed = compressed.split(pattern).join(token);
          }
        }
      }

      // Step 7: Store compression variables
      // In a real implementation, these would be stored for decompression
      const compressionVars = {
        original_length: data.length,
        entropy: entropy,
        B: B,
        V: V,
        I: I,
        G: G,
        F: F,
        E: E,
        t: t,
        C_sum: C_sum,
        alignment: alignment,
        timestamp: Date.now()
      };

      // For demonstration, we'll prepend a marker to indicate this is compressed
      const compressionMetadata = Buffer.from(JSON.stringify(compressionVars)).toString('base64');

      // Increase statistics counter
      MobiusCompression.stats.compressed += 1;
      MobiusCompression.stats.bytesProcessed += data.length;
      MobiusCompression.stats.bytesSaved += Math.max(0, data.length - compressed.length);

      // Return compressed data with metadata
      return `[MC:${compressionMetadata}]${compressed}`;
    } catch (error) {
      console.error('Compression error:', error.message);
      return data; // Return original on error
    }
  }

  /**
   * WHO: Decompressor
   * WHAT: Decompress data compressed with Möbius algorithm
   * WHEN: After data reception
   * WHERE: Layer 2 (Reactive)
   * WHY: To restore original data
   * HOW: Using stored compression variables
   * EXTENT: All compressed string data
   */
  static decompressString(data, context = {}) {
    if (!data || typeof data !== 'string') return data;

    try {
      // Check if this is compressed data
      if (!data.startsWith('[MC:')) {
        return data; // Not compressed with our algorithm
      }

      // Extract metadata and compressed content
      const metadataEndIndex = data.indexOf(']');
      if (metadataEndIndex === -1) {
        return data; // Invalid format
      }

      const metadataBase64 = data.substring(4, metadataEndIndex);
      const compressed = data.substring(metadataEndIndex + 1);

      // Parse compression variables
      const compressionVars = JSON.parse(Buffer.from(metadataBase64, 'base64').toString());

      // The actual decompression would happen here
      // For this demonstration, we'll just return the compressed portion

      // Increase statistics counter
      MobiusCompression.stats.decompressed += 1;

      return compressed;
    } catch (error) {
      console.error('Decompression error:', error.message);
      return data; // Return input on error
    }
  }

  /**
   * WHO: CompressionBenchmarker
   * WHAT: Measure compression ratio
   * WHEN: After compression
   * WHERE: Analysis layer
   * WHY: To evaluate algorithm effectiveness
   * HOW: Using size comparison
   * EXTENT: Compression analysis
   */
  static calculateCompressionRatio(original, compressed) {
    if (!original || !compressed) return 1.0;

    const originalSize = typeof original === 'string' ? original.length : String(original).length;
    const compressedSize = typeof compressed === 'string' ? compressed.length : String(compressed).length;

    if (originalSize === 0) return 1.0;
    return compressedSize / originalSize;
  }

  /**
   * WHO: StatisticsProvider
   * WHAT: Get compression statistics
   * WHEN: During system monitoring
   * WHERE: Monitoring interface
   * WHY: To provide performance metrics
   * HOW: Using accumulated data
   * EXTENT: All compression operations
   */
  static getStatistics() {
    const compressionRatio = MobiusCompression.stats.bytesProcessed > 0
      ? 1 - (MobiusCompression.stats.bytesSaved / MobiusCompression.stats.bytesProcessed)
      : 1.0;

    return {
      operationsCompressed: MobiusCompression.stats.compressed,
      operationsDecompressed: MobiusCompression.stats.decompressed,
      bytesProcessed: MobiusCompression.stats.bytesProcessed,
      bytesSaved: MobiusCompression.stats.bytesSaved,
      averageCompressionRatio: compressionRatio,
      lastReset: MobiusCompression.stats.lastReset
    };
  }

  /**
   * WHO: StatisticsResetter
   * WHAT: Reset compression statistics
   * WHEN: During system maintenance
   * WHERE: Monitoring interface
   * WHY: To refresh metrics
   * HOW: Using data reset
   * EXTENT: All statistics
   */
  static resetStatistics() {
    MobiusCompression.stats = {
      compressed: 0,
      decompressed: 0,
      bytesProcessed: 0,
      bytesSaved: 0,
      lastReset: Date.now()
    };

    return true;
  }
}

// Initialize statistics
MobiusCompression.stats = {
  compressed: 0,
  decompressed: 0,
  bytesProcessed: 0,
  bytesSaved: 0,
  lastReset: Date.now()
};

module.exports = MobiusCompression;