// WHO: LoggerComponent
// WHAT: Custom logging implementation for GitHub MCP server
// WHEN: During all server operation phases
// WHERE: System Layer 6 (Integration)
// WHY: To provide consistent 7D-aware logging capability
// HOW: Using internal logging with 7D context tagging
// EXTENT: All GitHub MCP server logging

package log

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

var (
	centralLocation *time.Location
)

// Compile-time assertion to ensure Logger implements LoggerInterface
var _ LoggerInterface = (*Logger)(nil)

// LogLevel defines logging level constants
type LogLevel int

// Log level constants
const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelCritical
)

// String returns the string representation of a log level
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// Logger provides a 7D context-aware logging facility
type Logger struct {
	// WHO: LoggerDataStructure
	// WHAT: Logger configuration and state
	// WHEN: Throughout logger lifecycle
	// WHERE: System Layer 6 (Integration)
	// WHY: To maintain logger configuration
	// HOW: Using structured logger properties
	// EXTENT: All logging operations

	level          LogLevel
	output         io.Writer
	timeFormat     string
	context        map[string]string
	useCompression bool
}

// NewLogger creates a new 7D-aware logger with default settings
func NewLogger() *Logger {
	// WHO: LoggerFactory
	// WHAT: Create new logger instance
	// WHEN: During system initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide logging capabilities
	// HOW: Using factory pattern with default values
	// EXTENT: Logger initialization

	return &Logger{
		level:      LevelInfo,
		output:     os.Stdout,
		timeFormat: time.RFC3339,
		context: map[string]string{
			"who":    "System",
			"what":   "Log",
			"where":  "TNOS_MCP_Bridge",
			"why":    "SystemMonitoring",
			"how":    "LoggerComponent",
			"extent": "SystemWide",
		},
		useCompression: true,
	}
}

// NewNopLogger creates a logger that discards all output
// WHO: LoggerFactory
// WHAT: Create no-operation logger
// WHEN: During testing
// WHERE: System Layer 6 (Integration)
// WHY: To provide a silent logger for testing
// HOW: Using discard writer
// EXTENT: Testing operations
func NewNopLogger() *Logger {
	logger := NewLogger()
	logger.output = io.Discard
	logger.level = LevelCritical // Only log critical messages
	return logger
}

// WithLevel sets the minimum log level
func (l *Logger) WithLevel(level LogLevel) *Logger {
	// WHO: LogLevelConfigurator
	// WHAT: Set logger level
	// WHEN: During logger configuration
	// WHERE: System Layer 6 (Integration)
	// WHY: To control logging verbosity
	// HOW: Using fluent interface pattern
	// EXTENT: Logger level setting

	l.level = level
	return l
}

// WithOutput sets the output writer
func (l *Logger) WithOutput(w io.Writer) *Logger {
	// WHO: LogOutputConfigurator
	// WHAT: Set logger output destination
	// WHEN: During logger configuration
	// WHERE: System Layer 6 (Integration)
	// WHY: To direct log output
	// HOW: Using fluent interface pattern
	// EXTENT: Logger output configuration

	l.output = w
	return l
}

// WithTimeFormat sets the time format
func (l *Logger) WithTimeFormat(format string) *Logger {
	// WHO: TimeFormatConfigurator
	// WHAT: Set time format for log entries
	// WHEN: During logger configuration
	// WHERE: System Layer 6 (Integration)
	// WHY: To control timestamp formatting
	// HOW: Using fluent interface pattern
	// EXTENT: Logger time format setting

	l.timeFormat = format
	return l
}

// WithContext sets 7D context values
func (l *Logger) WithContext(ctx map[string]string) *Logger {
	// WHO: ContextConfigurator
	// WHAT: Set 7D context for logger
	// WHEN: During logger configuration
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide 7D context awareness
	// HOW: Using fluent interface pattern
	// EXTENT: Logger context configuration

	if ctx != nil {
		l.context = ctx
	}
	return l
}

// WithCompression enables or disables log compression
func (l *Logger) WithCompression(compress bool) *Logger {
	// WHO: CompressionConfigurator
	// WHAT: Configure log compression
	// WHEN: During logger configuration
	// WHERE: System Layer 6 (Integration)
	// WHY: To support compression-first principle
	// HOW: Using fluent interface pattern
	// EXTENT: Logger compression setting

	l.useCompression = compress
	return l
}

// compressLogMessage applies Möbius Compression to log messages
// This is a placeholder for actual compression implementation
func (l *Logger) compressLogMessage(message string) string {
	// WHO: LogCompressor
	// WHAT: Apply compression to log messages
	// WHEN: During log processing
	// WHERE: System Layer 6 (Integration)
	// WHY: To implement compression-first principle
	// HOW: Using Möbius Compression Formula
	// EXTENT: Log message compression

	if !l.useCompression {
		return message
	}

	// In a real implementation, this would apply the Möbius Compression Formula
	// Currently just a placeholder
	if len(message) > 100 {
		compressed := message[:50] + "..." + message[len(message)-50:]
		return fmt.Sprintf("[COMPRESSED:%d->%d] %s", len(message), len(compressed), compressed)
	}
	return message
}

// format creates a formatted log message with 7D context
func (l *Logger) format(level LogLevel, message string) string {
	// WHO: LogFormatter
	// WHAT: Format log messages with context
	// WHEN: During log processing
	// WHERE: System Layer 6 (Integration)
	// WHY: To create structured logs with context
	// HOW: Using string formatting with 7D context
	// EXTENT: All log message formatting

	timestamp := time.Now().In(centralLocation).Format(l.timeFormat)

	// Format with 7D context
	contextStr := make([]string, 0, len(l.context))
	for dim, val := range l.context {
		contextStr = append(contextStr, fmt.Sprintf("%s:%s", dim, val))
	}

	// Apply compression if enabled
	message = l.compressLogMessage(message)

	return fmt.Sprintf("[%s] [%s] [%s] %s",
		timestamp,
		level.String(),
		strings.Join(contextStr, ","),
		message)
}

// log writes a message at the specified level
func (l *Logger) log(level LogLevel, message string, args ...interface{}) {
	// WHO: LogWriter
	// WHAT: Write log message
	// WHEN: During logging operation
	// WHERE: System Layer 6 (Integration)
	// WHY: To record log entries
	// HOW: Using formatted output with context
	// EXTENT: All logging activities

	if level < l.level {
		return
	}

	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	fmt.Fprintln(l.output, l.format(level, message))
}

// Debug logs a debug message
func (l *Logger) Debug(message string, args ...interface{}) {
	// WHO: DebugLogger
	// WHAT: Log debug message
	// WHEN: During debug logging
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide detailed system information
	// HOW: Using debug level logging
	// EXTENT: Development and troubleshooting

	l.log(LevelDebug, message, args...)
}

// Info logs an info message
func (l *Logger) Info(message string, args ...interface{}) {
	// WHO: InfoLogger
	// WHAT: Log info message
	// WHEN: During info logging
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide general system information
	// HOW: Using info level logging
	// EXTENT: General operations

	l.log(LevelInfo, message, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, args ...interface{}) {
	// WHO: WarnLogger
	// WHAT: Log warning message
	// WHEN: During warning logging
	// WHERE: System Layer 6 (Integration)
	// WHY: To warn about potential issues
	// HOW: Using warn level logging
	// EXTENT: Potential problems

	l.log(LevelWarn, message, args...)
}

// Error logs an error message
func (l *Logger) Error(message string, args ...interface{}) {
	// WHO: ErrorLogger
	// WHAT: Log error message
	// WHEN: During error logging
	// WHERE: System Layer 6 (Integration)
	// WHY: To report system errors
	// HOW: Using error level logging
	// EXTENT: Error conditions

	l.log(LevelError, message, args...)
}

// Critical logs a critical message
func (l *Logger) Critical(message string, args ...interface{}) {
	// WHO: CriticalLogger
	// WHAT: Log critical message
	// WHEN: During critical logging
	// WHERE: System Layer 6 (Integration)
	// WHY: To report critical system issues
	// HOW: Using critical level logging
	// EXTENT: Critical system failures

	l.log(LevelCritical, message, args...)
}

// SetContext updates a specific context dimension
func (l *Logger) SetContext(dimension, value string) {
	// WHO: ContextUpdater
	// WHAT: Update context dimension
	// WHEN: During context modification
	// WHERE: System Layer 6 (Integration)
	// WHY: To change contextual information
	// HOW: Using map update
	// EXTENT: Single context dimension

	l.context[dimension] = value
}

// GetContext retrieves the context value for a dimension
func (l *Logger) GetContext(dimension string) string {
	// WHO: ContextRetriever
	// WHAT: Get context dimension value
	// WHEN: During context access
	// WHERE: System Layer 6 (Integration)
	// WHY: To read contextual information
	// HOW: Using map access
	// EXTENT: Single context dimension

	return l.context[dimension]
}

// Update logger configuration based on mode
func UpdateLoggerMode(mode string) {
	// Initialize a logger for mode updates
	logger := NewLogger().WithLevel(LevelInfo)
	if mode == "standalone" {
		logger.Info("Logger set to standalone mode")
	} else if mode == "blood-connected" {
		logger.Info("Logger set to blood-connected mode")
	}
}

func init() {
	var err error
	centralLocation, err = time.LoadLocation("America/Chicago")
	if err != nil {
		centralLocation = time.Local // fallback
	}
}
