/**
 * BridgeLogger.js - Logging system for MCP Bridge
 * 
 * WHO: BridgeLogger
 * WHAT: Centralized logging for bridge operations
 * WHEN: Throughout system operation
 * WHERE: System Bridge Layer
 * WHY: To provide monitoring and diagnostics
 * HOW: Using structured logging with context
 * EXTENT: All bridge components
 */

const fs = require('fs');
const path = require('path');

/**
 * WHO: LoggerConfiguration
 * WHAT: Configure logging parameters
 * WHEN: During system initialization
 * WHERE: System Bridge Layer
 * WHY: To customize logging behavior
 * HOW: Using environment variables and defaults
 * EXTENT: Logging system configuration
 */
const config = {
  logLevel: process.env.MCP_BRIDGE_LOG_LEVEL || 'info',
  logFile: process.env.MCP_BRIDGE_LOG_FILE || path.join(process.cwd(), 'logs', 'mcp-bridge.log'),
  consoleOutput: process.env.MCP_BRIDGE_CONSOLE_LOG !== 'false',
  includeTnos7dContext: true,
  includeTimestamp: true,
  maxLogFileSize: parseInt(process.env.MCP_BRIDGE_MAX_LOG_SIZE || '10485760', 10), // 10MB default
  logRotation: parseInt(process.env.MCP_BRIDGE_LOG_ROTATION || '5', 10) // Keep 5 files by default
};

/**
 * WHO: LogLevelManager
 * WHAT: Define log levels and their priorities
 * WHEN: For filtering log messages
 * WHERE: System Bridge Layer
 * WHY: To control logging verbosity
 * HOW: Using numerical priorities
 * EXTENT: Log filtering decisions
 */
const LOG_LEVELS = {
  error: 0,
  warn: 1,
  info: 2,
  debug: 3,
  trace: 4
};

/**
 * WHO: LogDirectoryCreator
 * WHAT: Ensure log directory exists
 * WHEN: Before writing logs
 * WHERE: System Bridge Layer
 * WHY: To prevent file system errors
 * HOW: Using recursive directory creation
 * EXTENT: Log file storage
 */
function ensureLogDirectory() {
  const logDir = path.dirname(config.logFile);
  if (!fs.existsSync(logDir)) {
    try {
      fs.mkdirSync(logDir, { recursive: true });
    } catch (err) {
      console.error(`Failed to create log directory ${logDir}:`, err);
    }
  }
}

/**
 * WHO: LogRotator
 * WHAT: Implement log file rotation
 * WHEN: When log file exceeds size limit
 * WHERE: System Bridge Layer
 * WHY: To manage disk space
 * HOW: Using file renaming and cleanup
 * EXTENT: Log file maintenance
 */
function rotateLogFileIfNeeded() {
  try {
    if (!fs.existsSync(config.logFile)) return;

    const stats = fs.statSync(config.logFile);

    if (stats.size < config.maxLogFileSize) return;

    // Shift existing rotation files
    for (let i = config.logRotation - 1; i > 0; i--) {
      const oldFile = `${config.logFile}.${i}`;
      const newFile = `${config.logFile}.${i + 1}`;

      if (fs.existsSync(oldFile)) {
        fs.renameSync(oldFile, newFile);
      }
    }

    // Rename current log to .1
    fs.renameSync(config.logFile, `${config.logFile}.1`);

    // Cleanup files beyond rotation limit
    try {
      const excessFile = `${config.logFile}.${config.logRotation + 1}`;
      if (fs.existsSync(excessFile)) {
        fs.unlinkSync(excessFile);
      }
    } catch (cleanupErr) {
      // Non-critical, just print to console
      console.error('Error cleaning up excess log files:', cleanupErr);
    }
  } catch (err) {
    console.error('Error rotating log files:', err);
  }
}

/**
 * WHO: LogFormatter
 * WHAT: Format log messages for output
 * WHEN: During log writing
 * WHERE: System Bridge Layer
 * WHY: To provide consistent log format
 * HOW: Using structured log formatting
 * EXTENT: All log messages
 */
function formatLogMessage(level, message, context, error) {
  const timestamp = config.includeTimestamp ? new Date().toISOString() : null;

  // Basic structure
  const logEntry = {
    timestamp,
    level,
    message
  };

  // Add 7D context if enabled and available
  if (config.includeTnos7dContext && context) {
    logEntry.context = context;
  }

  // Add error information if available
  if (error) {
    logEntry.error = {
      name: error.name,
      message: error.message,
      stack: error.stack
    };
  }

  return JSON.stringify(logEntry);
}

/**
 * WHO: LogWriter
 * WHAT: Write logs to file and console
 * WHEN: During logging operations
 * WHERE: System Bridge Layer
 * WHY: To persist log information
 * HOW: Using file system and console output
 * EXTENT: All log messages
 */
function writeLog(formattedMessage) {
  // Write to console if enabled
  if (config.consoleOutput) {
    console.log(formattedMessage);
  }

  // Write to file
  ensureLogDirectory();
  rotateLogFileIfNeeded();

  try {
    fs.appendFileSync(config.logFile, formattedMessage + '\n');
  } catch (err) {
    console.error('Failed to write to log file:', err);
  }
}

/**
 * WHO: Logger
 * WHAT: Primary logging function
 * WHEN: During system operation
 * WHERE: System Bridge Layer
 * WHY: To record system events
 * HOW: Using leveled logging with context
 * EXTENT: All logging operations
 */
function log(level, message, context = null, error = null) {
  // Check log level
  if (LOG_LEVELS[level] > LOG_LEVELS[config.logLevel]) {
    return; // Skip logging if level is too verbose
  }

  const formattedMessage = formatLogMessage(level, message, context, error);
  writeLog(formattedMessage);
}

/**
 * WHO: ConfigSetter
 * WHAT: Update logger configuration
 * WHEN: During runtime
 * WHERE: System Bridge Layer
 * WHY: To adjust logging behavior
 * HOW: Using configuration parameter updates
 * EXTENT: Logger configuration
 */
function setLoggerConfig(newConfig) {
  Object.assign(config, newConfig);
}

/**
 * WHO: ConfigGetter
 * WHAT: Retrieve current logger configuration
 * WHEN: When configuration needed
 * WHERE: System Bridge Layer
 * WHY: For monitoring and diagnostics
 * HOW: By returning configuration object
 * EXTENT: Logger configuration view
 */
function getLoggerConfig() {
  return { ...config };
}

module.exports = {
  log,
  setLoggerConfig,
  getLoggerConfig,
  LOG_LEVELS
};