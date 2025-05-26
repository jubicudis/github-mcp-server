/**
 * MCPContextPersistence.js
 * 
 * This module provides functionality for persisting and retrieving MCP context data
 * between different system components. It ensures context continuity across
 * bridge operations and provides methods for context manipulation.
 */

// WHO: ContextPersistenceManager
// WHAT: MCP Context Persistence Module
// WHEN: During MCP bridge operations
// WHERE: TNOS Layer 2 / GitHub MCP Bridge
// WHY: To maintain context continuity across MCP operations
// HOW: File system persistence with memory caching
// EXTENT: All MCP context operations

const fs = require('fs');
const path = require('path');
const crypto = require('crypto');

// Configuration
const CONFIG = {
  persistencePath: '/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server/data/context',
  logPath: '/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/logs/mcp_context_persistence.log',
  contextSources: ['github', 'tnos7d', 'system', 'user'],
  retention: {
    maxFileAge: 7 * 24 * 60 * 60 * 1000, // 7 days in milliseconds
    maxFileSize: 10 * 1024 * 1024, // 10 MB
    pruneInterval: 24 * 60 * 60 * 1000 // 24 hours in milliseconds
  }
};

// In-memory cache for fast access
const contextCache = {
  github: {},
  tnos7d: {},
  system: {},
  user: {}
};

/**
 * Initialize context persistence system
 */
async function initialize() {
  try {
    // Ensure context directory exists
    if (!fs.existsSync(CONFIG.persistencePath)) {
      fs.mkdirSync(CONFIG.persistencePath, { recursive: true });
    }

    // Ensure log directory exists
    const logDir = path.dirname(CONFIG.logPath);
    if (!fs.existsSync(logDir)) {
      fs.mkdirSync(logDir, { recursive: true });
    }

    // Initialize log file with header
    if (!fs.existsSync(CONFIG.logPath)) {
      fs.writeFileSync(CONFIG.logPath, `# MCP Context Persistence Log\n# Created: ${new Date().toISOString()}\n\n`);
    }

    // Initialize context files for each source if they don't exist
    for (const source of CONFIG.contextSources) {
      const contextPath = path.join(CONFIG.persistencePath, `${source}_context.json`);
      if (!fs.existsSync(contextPath)) {
        const initialContext = {
          _meta: {
            source: source,
            created: new Date().toISOString(),
            updated: new Date().toISOString(),
            version: 1
          }
        };
        fs.writeFileSync(contextPath, JSON.stringify(initialContext, null, 2));
        contextCache[source] = initialContext;
      } else {
        try {
          // Load existing context into memory cache
          const data = fs.readFileSync(contextPath, 'utf8');
          contextCache[source] = JSON.parse(data);
        } catch (err) {
          logMessage(`Error loading context file for ${source}: ${err.message}`);
          // Initialize with empty context if file exists but is corrupted
          contextCache[source] = {
            _meta: {
              source: source,
              created: new Date().toISOString(),
              updated: new Date().toISOString(),
              version: 1,
              note: "Initialized after error loading previous context"
            }
          };
        }
      }
    }

    // Schedule periodic pruning of old context files
    setTimeout(pruneOldContextFiles, CONFIG.retention.pruneInterval);

    logMessage("Context persistence system initialized successfully");
    return true;
  } catch (error) {
    logMessage(`Error initializing context persistence system: ${error.message}`);
    return false;
  }
}

/**
 * Log a message to the context persistence log file
 */
function logMessage(message) {
  const timestamp = new Date().toISOString();
  const logEntry = `[${timestamp}] ${message}\n`;

  try {
    fs.appendFileSync(CONFIG.logPath, logEntry);
  } catch (error) {
    console.error(`Failed to write to context persistence log: ${error.message}`);
  }

  // Also log to console for immediate visibility
  console.log(`[ContextPersistence] ${message}`);
}

/**
 * Load context from persistence storage
 */
async function loadContext(source) {
  try {
    if (!CONFIG.contextSources.includes(source)) {
      logMessage(`Invalid context source: ${source}`);
      return {};
    }

    // Check if context is in memory cache
    if (contextCache[source] && Object.keys(contextCache[source]).length > 0) {
      return contextCache[source];
    }

    // Load from disk if not in cache
    const contextPath = path.join(CONFIG.persistencePath, `${source}_context.json`);
    if (!fs.existsSync(contextPath)) {
      logMessage(`Context file does not exist for source: ${source}`);
      return {};
    }

    const data = fs.readFileSync(contextPath, 'utf8');
    try {
      const context = JSON.parse(data);
      // Update cache
      contextCache[source] = context;
      logMessage(`Loaded context for ${source} (${Object.keys(context).length} keys)`);
      return context;
    } catch (error) {
      logMessage(`Error parsing context file for ${source}: ${error.message}`);
      return {};
    }
  } catch (error) {
    logMessage(`Error loading context for ${source}: ${error.message}`);
    return {};
  }
}

/**
 * Save context to persistence storage
 */
async function saveContext(context, source) {
  try {
    if (!CONFIG.contextSources.includes(source)) {
      logMessage(`Invalid context source: ${source}`);
      return false;
    }

    if (!context || typeof context !== 'object') {
      logMessage(`Invalid context data for ${source}: not an object`);
      return false;
    }

    // Update metadata
    if (!context._meta) {
      context._meta = {};
    }

    context._meta.updated = new Date().toISOString();
    context._meta.source = source;

    if (!context._meta.created) {
      context._meta.created = new Date().toISOString();
    }

    if (!context._meta.version) {
      context._meta.version = 1;
    } else {
      context._meta.version++;
    }

    // Create a backup before overwriting
    const contextPath = path.join(CONFIG.persistencePath, `${source}_context.json`);
    if (fs.existsSync(contextPath)) {
      const backupPath = path.join(CONFIG.persistencePath, `${source}_context_${Date.now()}.bak.json`);
      fs.copyFileSync(contextPath, backupPath);
    }

    // Write updated context to file
    fs.writeFileSync(contextPath, JSON.stringify(context, null, 2));

    // Update cache
    contextCache[source] = context;

    logMessage(`Saved context for ${source} (${Object.keys(context).length} keys)`);
    return true;
  } catch (error) {
    logMessage(`Error saving context for ${source}: ${error.message}`);
    return false;
  }
}

/**
 * Merge new context data with existing context
 */
async function mergeContext(newContext, source) {
  try {
    if (!CONFIG.contextSources.includes(source)) {
      logMessage(`Invalid context source for merge: ${source}`);
      return false;
    }

    if (!newContext || typeof newContext !== 'object') {
      logMessage(`Invalid context data for merge (${source}): not an object`);
      return false;
    }

    // Load existing context
    const existingContext = await loadContext(source);

    // Perform deep merge
    const mergedContext = deepMerge(existingContext, newContext);

    // Save merged context
    const result = await saveContext(mergedContext, source);

    logMessage(`Merged context for ${source} (${Object.keys(newContext).length} new keys, ${Object.keys(mergedContext).length} total keys)`);
    return result;
  } catch (error) {
    logMessage(`Error merging context for ${source}: ${error.message}`);
    return false;
  }
}

/**
 * Deep merge two objects, combining their properties recursively
 */
function deepMerge(target, source) {
  // Make a copy so we don't modify the original objects
  const output = Object.assign({}, target);

  if (isObject(target) && isObject(source)) {
    Object.keys(source).forEach(key => {
      if (isObject(source[key])) {
        if (!(key in target)) {
          Object.assign(output, { [key]: source[key] });
        } else {
          output[key] = deepMerge(target[key], source[key]);
        }
      } else {
        Object.assign(output, { [key]: source[key] });
      }
    });
  }

  return output;
}

/**
 * Check if a value is an object
 */
function isObject(item) {
  return (item && typeof item === 'object' && !Array.isArray(item));
}

/**
 * Remove a specific context or context field
 */
async function removeContext(source, path = null) {
  try {
    if (!CONFIG.contextSources.includes(source)) {
      logMessage(`Invalid context source for removal: ${source}`);
      return false;
    }

    // Load existing context
    const context = await loadContext(source);

    if (!path) {
      // Remove entire context except metadata
      const metadata = context._meta || {};
      const newContext = { _meta: metadata };
      return await saveContext(newContext, source);
    }

    // Remove specific path within context
    const pathParts = path.split('.');
    let current = context;

    for (let i = 0; i < pathParts.length - 1; i++) {
      const part = pathParts[i];
      if (!current[part] || typeof current[part] !== 'object') {
        logMessage(`Cannot remove path ${path} (${source}): path does not exist`);
        return false;
      }
      current = current[part];
    }

    const lastPart = pathParts[pathParts.length - 1];
    if (!(lastPart in current)) {
      logMessage(`Cannot remove path ${path} (${source}): path does not exist`);
      return false;
    }

    delete current[lastPart];

    // Save updated context
    const result = await saveContext(context, source);

    logMessage(`Removed context path ${path} for ${source}`);
    return result;
  } catch (error) {
    logMessage(`Error removing context for ${source}: ${error.message}`);
    return false;
  }
}

/**
 * Query context data using a path
 */
async function queryContext(source, path) {
  try {
    if (!CONFIG.contextSources.includes(source)) {
      logMessage(`Invalid context source for query: ${source}`);
      return null;
    }

    // Load context
    const context = await loadContext(source);

    if (!path) {
      // Return entire context
      return context;
    }

    // Navigate to specific path
    const pathParts = path.split('.');
    let current = context;

    for (const part of pathParts) {
      if (!current[part]) {
        return null;
      }
      current = current[part];
    }

    return current;
  } catch (error) {
    logMessage(`Error querying context for ${source}: ${error.message}`);
    return null;
  }
}

/**
 * Prune old context backup files according to retention policy
 */
function pruneOldContextFiles() {
  try {
    const now = Date.now();
    const files = fs.readdirSync(CONFIG.persistencePath);

    let pruned = 0;

    for (const file of files) {
      // Only process backup files
      if (!file.includes('.bak.json')) {
        continue;
      }

      const filePath = path.join(CONFIG.persistencePath, file);
      const stats = fs.statSync(filePath);

      // Check file age
      const fileAge = now - stats.mtimeMs;
      if (fileAge > CONFIG.retention.maxFileAge) {
        fs.unlinkSync(filePath);
        pruned++;
        continue;
      }

      // Check file size
      if (stats.size > CONFIG.retention.maxFileSize) {
        fs.unlinkSync(filePath);
        pruned++;
      }
    }

    if (pruned > 0) {
      logMessage(`Pruned ${pruned} old context files`);
    }

    // Schedule next pruning
    setTimeout(pruneOldContextFiles, CONFIG.retention.pruneInterval);
  } catch (error) {
    logMessage(`Error pruning old context files: ${error.message}`);
    // Reschedule despite error
    setTimeout(pruneOldContextFiles, CONFIG.retention.pruneInterval);
  }
}

// Export the module functions
module.exports = {
  initialize,
  loadContext,
  saveContext,
  mergeContext,
  removeContext,
  queryContext,
  CONFIG
};

// Initialize on module load
initialize().catch(error => {
  console.error(`Failed to initialize context persistence system: ${error.message}`);
});
