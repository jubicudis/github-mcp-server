/**
 * TNOS7DContextBridge.js
 * 
 * This module serves as a bridge between the TNOS 7D Context format and the MCP context format.
 * It handles translation, transformation, and compatibility between the two context models,
 * ensuring seamless integration between TNOS and GitHub MCP.
 */

const fs = require('fs');
const path = require('path');
const crypto = require('crypto');

// Import the context persistence module
const ContextPersistence = require('./MCPContextPersistence.js');

// Configuration
const CONFIG = {
  tnos7dContextPath: '/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/core/context_master.7d',
  tempDir: '/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/core/reactive/temp',
  logPath: '/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/logs/tnos7d_context_bridge.log'
};

// Initialize logging
const logMessage = (message) => {
  const timestamp = new Date().toISOString();
  const entry = `[${timestamp}] ${message}\n`;
  fs.appendFileSync(CONFIG.logPath, entry);
  console.log(message);
};

// Ensure directories exist
function ensureDirectories() {
  const dirs = [
    CONFIG.tempDir,
    path.dirname(CONFIG.logPath)
  ];

  for (const dir of dirs) {
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
      logMessage(`Created directory: ${dir}`);
    }
  }
}

// Initialize bridge
function initialize() {
  try {
    ensureDirectories();
    logMessage('TNOS 7D Context Bridge initialized');
    return true;
  } catch (error) {
    logMessage(`Error initializing TNOS 7D Context Bridge: ${error.message}`);
    return false;
  }
}

/**
 * Read the TNOS 7D context file
 * The 7D format is a specialized hierarchical context format used by TNOS
 */
async function readTNOS7DContext() {
  try {
    if (!fs.existsSync(CONFIG.tnos7dContextPath)) {
      logMessage(`TNOS 7D context file not found at ${CONFIG.tnos7dContextPath}`);
      return null;
    }

    const data = fs.readFileSync(CONFIG.tnos7dContextPath, 'utf8');
    logMessage(`Read TNOS 7D context file (${data.length} bytes)`);

    return data;
  } catch (error) {
    logMessage(`Error reading TNOS 7D context: ${error.message}`);
    return null;
  }
}

/**
 * Parse the TNOS 7D context into a structured format
 * This converts the specialized 7D format into a structured object
 */
async function parseTNOS7DContext(rawContext) {
  try {
    if (!rawContext) {
      logMessage('No raw context provided for parsing');
      return null;
    }

    // The parsing logic here depends on the actual structure of the 7D format
    // This is a simplified example - actual implementation would be more complex

    // Split the content by sections (denoted by dimension markers)
    const sections = rawContext.split(/\[D\d+\]/g).filter(Boolean);

    // Process each section
    const parsedContext = {
      dimensions: {},
      vectors: [],
      metadata: {},
      timestamp: new Date().toISOString()
    };

    // Extract dimension information
    const dimensionMatches = rawContext.match(/\[D(\d+)\]([^\[]+)/g) || [];
    dimensionMatches.forEach(match => {
      const dimension = match.match(/\[D(\d+)\]/)[1];
      const content = match.replace(/\[D\d+\]/, '').trim();
      parsedContext.dimensions[`D${dimension}`] = content;
    });

    // Extract vector information
    const vectorMatches = rawContext.match(/\<V([^>]+)>([^<]+)</g) || [];
    vectorMatches.forEach(match => {
      const id = match.match(/\<V([^>]+)>/)[1];
      const content = match.replace(/\<V[^>]+>/, '').replace(/<$/, '').trim();
      parsedContext.vectors.push({ id, content });
    });

    // Extract metadata
    const metadataMatch = rawContext.match(/\{META\}([^{]+)\{\/META\}/);
    if (metadataMatch) {
      try {
        parsedContext.metadata = JSON.parse(metadataMatch[1].trim());
      } catch (e) {
        parsedContext.metadata = { error: 'Failed to parse metadata as JSON' };
      }
    }

    logMessage(`Parsed TNOS 7D context: ${Object.keys(parsedContext.dimensions).length} dimensions, ${parsedContext.vectors.length} vectors`);
    return parsedContext;
  } catch (error) {
    logMessage(`Error parsing TNOS 7D context: ${error.message}`);
    return null;
  }
}

/**
 * Convert TNOS 7D context to MCP context format
 */
async function convertTNOS7DToMCPFormat(parsedContext) {
  try {
    if (!parsedContext) {
      logMessage('No parsed context provided for conversion');
      return null;
    }

    // Create MCP context structure
    const mcpContext = {
      // Base context information
      context: {
        session: {
          id: parsedContext.metadata.sessionId || crypto.randomUUID(),
          created: parsedContext.timestamp,
          updated: new Date().toISOString()
        },
        user: {
          preferences: parsedContext.metadata.userPreferences || {},
          history: extractHistoryFromVectors(parsedContext.vectors)
        },
        system: {
          tnos: {
            version: parsedContext.metadata.tnosVersion || "unknown",
            contextFormat: "7D",
            dimensions: Object.keys(parsedContext.dimensions).length
          }
        }
      },

      // Content extracted from dimensions
      content: {},

      // Vectors transformed to MCP format
      vectors: {},

      // Metadata and transformation info
      _meta: {
        source: "TNOS7DContextBridge",
        transformedAt: new Date().toISOString(),
        originalFormat: "7D",
        convertedFormat: "MCP"
      }
    };

    // Process dimensions into content
    Object.entries(parsedContext.dimensions).forEach(([key, value]) => {
      const dimensionNumber = key.replace('D', '');

      // Map dimensions to appropriate MCP content sections
      switch (dimensionNumber) {
        case '1':
          mcpContext.content.current = value;
          break;
        case '2':
          mcpContext.content.context = value;
          break;
        case '3':
          mcpContext.content.knowledge = value;
          break;
        case '4':
          mcpContext.content.memory = value;
          break;
        case '5':
          mcpContext.content.reasoning = value;
          break;
        case '6':
          mcpContext.content.awareness = value;
          break;
        case '7':
          mcpContext.content.integration = value;
          break;
        default:
          mcpContext.content[`dimension${dimensionNumber}`] = value;
      }
    });

    // Process vectors
    parsedContext.vectors.forEach((vector, index) => {
      mcpContext.vectors[`v${index}`] = {
        id: vector.id,
        content: vector.content,
        position: index,
        metadata: extractVectorMetadata(vector.id)
      };
    });

    logMessage(`Converted TNOS 7D context to MCP format successfully`);
    return mcpContext;
  } catch (error) {
    logMessage(`Error converting TNOS 7D context to MCP format: ${error.message}`);
    return null;
  }
}

/**
 * Extract user interaction history from context vectors
 */
function extractHistoryFromVectors(vectors) {
  // This is a simplified implementation
  const history = [];

  // Look for vectors that represent user interactions
  vectors.forEach(vector => {
    if (vector.id.includes('USER_INTERACTION') || vector.id.includes('DIALOG')) {
      const parts = vector.content.split('|');
      if (parts.length >= 2) {
        history.push({
          timestamp: new Date().toISOString(),
          type: vector.id,
          content: vector.content
        });
      }
    }
  });

  return history;
}

/**
 * Extract metadata from vector ID
 */
function extractVectorMetadata(vectorId) {
  // This is a simplified implementation
  const parts = vectorId.split('_');

  const metadata = {
    type: parts[0] || 'UNKNOWN',
    category: parts[1] || 'GENERAL',
    timestamp: new Date().toISOString()
  };

  return metadata;
}

/**
 * Convert MCP context format back to TNOS 7D format
 */
async function convertMCPToTNOS7DFormat(mcpContext) {
  try {
    if (!mcpContext) {
      logMessage('No MCP context provided for conversion');
      return null;
    }

    // Initialize 7D context structure
    let tnos7dContext = '';

    // Add metadata section
    const metadata = {
      sessionId: mcpContext.context?.session?.id || crypto.randomUUID(),
      tnosVersion: mcpContext.context?.system?.tnos?.version || "unknown",
      userPreferences: mcpContext.context?.user?.preferences || {},
      transformedAt: new Date().toISOString(),
      originalFormat: "MCP",
      convertedFormat: "7D"
    };

    tnos7dContext += `{META}${JSON.stringify(metadata, null, 2)}{/META}\n\n`;

    // Convert content sections to dimensions
    if (mcpContext.content) {
      // Map MCP content sections to 7D dimensions
      if (mcpContext.content.current) {
        tnos7dContext += `[D1]${mcpContext.content.current}\n\n`;
      }

      if (mcpContext.content.context) {
        tnos7dContext += `[D2]${mcpContext.content.context}\n\n`;
      }

      if (mcpContext.content.knowledge) {
        tnos7dContext += `[D3]${mcpContext.content.knowledge}\n\n`;
      }

      if (mcpContext.content.memory) {
        tnos7dContext += `[D4]${mcpContext.content.memory}\n\n`;
      }

      if (mcpContext.content.reasoning) {
        tnos7dContext += `[D5]${mcpContext.content.reasoning}\n\n`;
      }

      if (mcpContext.content.awareness) {
        tnos7dContext += `[D6]${mcpContext.content.awareness}\n\n`;
      }

      if (mcpContext.content.integration) {
        tnos7dContext += `[D7]${mcpContext.content.integration}\n\n`;
      }

      // Handle any additional content sections
      Object.entries(mcpContext.content).forEach(([key, value]) => {
        if (!['current', 'context', 'knowledge', 'memory', 'reasoning', 'awareness', 'integration'].includes(key)) {
          // Extract dimension number if present
          const dimensionMatch = key.match(/dimension(\d+)/);
          if (dimensionMatch) {
            tnos7dContext += `[D${dimensionMatch[1]}]${value}\n\n`;
          } else {
            // Add as a custom dimension
            tnos7dContext += `[D-${key}]${value}\n\n`;
          }
        }
      });
    }

    // Convert vectors
    if (mcpContext.vectors) {
      Object.values(mcpContext.vectors).forEach(vector => {
        tnos7dContext += `<V${vector.id}>${vector.content}<\n\n`;
      });
    }

    // Add system information at the end
    tnos7dContext += `[D-SYSTEM]Converted from MCP format at ${new Date().toISOString()}\n`;
    tnos7dContext += `Original MCP session ID: ${mcpContext.context?.session?.id || 'unknown'}\n`;
    tnos7dContext += `Conversion tool: TNOS7DContextBridge\n`;

    logMessage(`Converted MCP context to TNOS 7D format successfully`);
    return tnos7dContext;
  } catch (error) {
    logMessage(`Error converting MCP context to TNOS 7D format: ${error.message}`);
    return null;
  }
}

/**
 * Save TNOS 7D context to file
 */
async function saveTNOS7DContext(context, targetPath = CONFIG.tnos7dContextPath) {
  try {
    if (!context) {
      logMessage('No context provided for saving');
      return false;
    }

    // Ensure target directory exists
    const targetDir = path.dirname(targetPath);
    if (!fs.existsSync(targetDir)) {
      fs.mkdirSync(targetDir, { recursive: true });
    }

    // Create backup of existing file if it exists
    if (fs.existsSync(targetPath)) {
      const backupPath = `${targetPath}.${new Date().toISOString().replace(/:/g, '-')}.bak`;
      fs.copyFileSync(targetPath, backupPath);
      logMessage(`Created backup of TNOS 7D context at ${backupPath}`);
    }

    // Write context to file
    fs.writeFileSync(targetPath, context, 'utf8');
    logMessage(`Saved TNOS 7D context to ${targetPath} (${context.length} bytes)`);

    return true;
  } catch (error) {
    logMessage(`Error saving TNOS 7D context: ${error.message}`);
    return false;
  }
}

/**
 * Transform TNOS 7D context to MCP context
 */
async function transformTNOS7DToMCP() {
  try {
    // Read raw TNOS 7D context
    const rawContext = await readTNOS7DContext();
    if (!rawContext) {
      return null;
    }

    // Parse the raw context
    const parsedContext = await parseTNOS7DContext(rawContext);
    if (!parsedContext) {
      return null;
    }

    // Convert to MCP format
    const mcpContext = await convertTNOS7DToMCPFormat(parsedContext);
    if (!mcpContext) {
      return null;
    }

    // Save to persistence
    await ContextPersistence.saveContext(mcpContext, 'tnos7d');

    logMessage('Successfully transformed TNOS 7D context to MCP format');
    return mcpContext;
  } catch (error) {
    logMessage(`Error transforming TNOS 7D to MCP: ${error.message}`);
    return null;
  }
}

/**
 * Transform MCP context to TNOS 7D context
 */
async function transformMCPToTNOS7D(source = 'github') {
  try {
    // Load MCP context
    const mcpContext = await ContextPersistence.loadContext(source);
    if (!mcpContext || Object.keys(mcpContext).length === 0) {
      logMessage(`No ${source} MCP context available for transformation`);
      return null;
    }

    // Convert to TNOS 7D format
    const tnos7dContext = await convertMCPToTNOS7DFormat(mcpContext);
    if (!tnos7dContext) {
      return null;
    }

    // Save to file
    const success = await saveTNOS7DContext(tnos7dContext);
    if (!success) {
      return null;
    }

    logMessage(`Successfully transformed ${source} MCP context to TNOS 7D format`);
    return tnos7dContext;
  } catch (error) {
    logMessage(`Error transforming MCP to TNOS 7D: ${error.message}`);
    return null;
  }
}

/**
 * Sync context between TNOS 7D and MCP formats
 */
async function syncContext(direction = 'bidirectional') {
  try {
    logMessage(`Starting context synchronization (${direction})`);

    if (direction === 'tnos7d-to-mcp' || direction === 'bidirectional') {
      // Transform TNOS 7D to MCP
      const mcpContext = await transformTNOS7DToMCP();
      if (!mcpContext) {
        logMessage('Failed to transform TNOS 7D to MCP');
      }
    }

    if (direction === 'mcp-to-tnos7d' || direction === 'bidirectional') {
      // Transform MCP to TNOS 7D
      const tnos7dContext = await transformMCPToTNOS7D();
      if (!tnos7dContext) {
        logMessage('Failed to transform MCP to TNOS 7D');
      }
    }

    logMessage(`Completed context synchronization (${direction})`);
    return true;
  } catch (error) {
    logMessage(`Error synchronizing context: ${error.message}`);
    return false;
  }
}

// Export the module functions
module.exports = {
  CONFIG,
  initialize,
  readTNOS7DContext,
  parseTNOS7DContext,
  convertTNOS7DToMCPFormat,
  convertMCPToTNOS7DFormat,
  saveTNOS7DContext,
  transformTNOS7DToMCP,
  transformMCPToTNOS7D,
  syncContext
};

// Initialize on module load
initialize();

