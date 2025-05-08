/*
 * WHO: Logger
 * WHAT: 7D Context-aware logging system
 * WHEN: During system operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide structured logging with context
 * HOW: Using Go's structured logging with context compression
 * EXTENT: All system operations
 */

package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Config holds the logger configuration
type Config struct {
	// WHO: ConfigurationManager
	// WHAT: Logger configuration structure
	// WHEN: During logger initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide configurable logging options
	// HOW: Using structured configuration parameters
	// EXTENT: Logger initialization

	// Level is the minimum log level to output
	Level string
	// FilePath is the path to the log file
	FilePath string
	// ConsoleOut indicates whether to log to the console
	ConsoleOut bool
	// ContextMode controls how context is displayed
	ContextMode string
	// EnableCompression controls whether to compress context data
	EnableCompression bool
}

// Level represents a log level
type Level int

const (
	// Debug is used for detailed information
	Debug Level = iota
	// Info is used for general information
	Info
	// Warn is used for warnings that don't affect execution
	Warn
	// Error is used for errors that affect execution
	Error
	// Fatal is used for critical errors that stop execution
	Fatal
)

// String representation of log levels
func (l Level) String() string {
	// WHO: LevelFormatter
	// WHAT: Format log level as string
	// WHEN: During log formatting
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide human-readable log levels
	// HOW: Using string conversion
	// EXTENT: Log message formatting

	switch l {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	case Fatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// FromString converts a string to a Level
func FromString(s string) Level {
	// WHO: LevelParser
	// WHAT: Parse string to log level
	// WHEN: During logger configuration
	// WHERE: System Layer 6 (Integration)
	// WHY: To convert string configuration to level enum
	// HOW: Using string comparison
	// EXTENT: Logger initialization

	switch strings.ToUpper(s) {
	case "DEBUG":
		return Debug
	case "INFO":
		return Info
	case "WARN":
		return Warn
	case "ERROR":
		return Error
	case "FATAL":
		return Fatal
	default:
		return Info
	}
}

// ContextVector7D represents the 7-dimensional context vector used in logging
type ContextVector7D struct {
	// WHO: ContextVectorManager
	// WHAT: 7D context structure
	// WHEN: During context operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To represent context dimensions
	// HOW: Using structured dimensions
	// EXTENT: Context representation

	Who    string                 `json:"who"`
	What   string                 `json:"what"`
	When   int64                  `json:"when"`
	Where  string                 `json:"where"`
	Why    string                 `json:"why"`
	How    string                 `json:"how"`
	Extent float64                `json:"extent"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
}

// Logger implements 7D Context-aware logging
type Logger struct {
	// WHO: LoggerCore
	// WHAT: Core logging implementation
	// WHEN: During log operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To centralize logging with context
	// HOW: Using structured data with compression
	// EXTENT: All log operations

	level             Level
	contextMode       string
	enableCompression bool
	out               io.Writer
	fileOut           io.Writer
	mu                sync.Mutex
	context           *ContextVector7D
}

// NewLogger creates a new Logger with the specified configuration
func NewLogger(config Config) *Logger {
	// WHO: LoggerFactory
	// WHAT: Logger initialization
	// WHEN: During system startup
	// WHERE: System Layer 6 (Integration)
	// WHY: To create configured logger instance
	// HOW: Using provided configuration parameters
	// EXTENT: Logger lifecycle

	var out io.Writer
	var fileOut io.Writer = nil
	var writers []io.Writer

	// Set up file output if specified
	if config.FilePath != "" {
		// Create the directory if it doesn't exist
		dir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(dir, 0755); err == nil {
			// Open log file for append
			file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err == nil {
				fileOut = file
				writers = append(writers, file)
			} else {
				fmt.Fprintf(os.Stderr, "Failed to open log file %s: %v\n", config.FilePath, err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Failed to create log directory %s: %v\n", dir, err)
		}
	}

	// Set up console output if enabled
	if config.ConsoleOut {
		writers = append(writers, os.Stdout)
	}

	// Use MultiWriter if we have multiple outputs
	switch len(writers) {
	case 0:
		out = os.Stdout // Default to stdout if no writers
	case 1:
		out = writers[0]
	default:
		out = io.MultiWriter(writers...)
	}

	// Create default context
	defaultContext := &ContextVector7D{
		Who:    "LogSystem",
		What:   "Logging",
		When:   time.Now().Unix(),
		Where:  "System Layer 6",
		Why:    "SystemMonitoring",
		How:    "StructuredLogging",
		Extent: 1.0,
		Meta:   make(map[string]interface{}),
	}

	// Create and return logger
	return &Logger{
		level:             FromString(config.Level),
		contextMode:       config.ContextMode,
		enableCompression: config.EnableCompression,
		out:               out,
		fileOut:           fileOut,
		context:           defaultContext,
	}
}

// DefaultLogger returns a logger with default configuration
func DefaultLogger() *Logger {
	// WHO: DefaultLoggerProvider
	// WHAT: Create default logger
	// WHEN: During quick initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide a simple logger with defaults
	// HOW: Using standard configuration
	// EXTENT: Simple logging needs

	config := Config{
		Level:       "INFO",
		ConsoleOut:  true,
		ContextMode: "debug",
	}
	return NewLogger(config)
}

// Close closes the logger and associated resources
func (l *Logger) Close() {
	// WHO: ResourceCleaner
	// WHAT: Release logger resources
	// WHEN: During shutdown
	// WHERE: System Layer 6 (Integration)
	// WHY: To properly close file handles
	// HOW: Using io.Closer interface
	// EXTENT: Resource cleanup

	l.mu.Lock()
	defer l.mu.Unlock()

	// Close the file if it's open
	if closer, ok := l.fileOut.(io.Closer); ok && closer != nil {
		closer.Close()
	}
}

// WithContext returns a new logger with the specified context
func (l *Logger) WithContext(context *ContextVector7D) *Logger {
	// WHO: ContextualLoggerProvider
	// WHAT: Create contextualized logger
	// WHEN: During context-aware operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide context-aware logging
	// HOW: Using context inheritance
	// EXTENT: Contextual operations

	newLogger := &Logger{
		level:             l.level,
		contextMode:       l.contextMode,
		enableCompression: l.enableCompression,
		out:               l.out,
		fileOut:           l.fileOut,
		context:           context,
	}

	return newLogger
}

// log is the internal logging function
func (l *Logger) log(level Level, msg string, keyvals ...interface{}) {
	// WHO: LogRecorder
	// WHAT: Record log message with context
	// WHEN: During log operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To store log data with context
	// HOW: Using formatted output with context
	// EXTENT: All log messages

	// Skip logging if the level is below the configured level
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}

	// Format the log entry
	ts := time.Now().Format("2006-01-02 15:04:05.000")
	fileLine := fmt.Sprintf("%s:%d", filepath.Base(file), line)

	// Start with timestamp, level, and message
	fmt.Fprintf(l.out, "[%s] [%s] [%s] %s", ts, level, fileLine, msg)

	// Add key-value pairs
	var contextFields []string
	var standardFields []string

	// Separate context fields from standard fields
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			key := fmt.Sprintf("%v", keyvals[i])
			val := keyvals[i+1]

			// Check if this is a context field
			if isContextDimension(key) {
				contextFields = append(contextFields, fmt.Sprintf("%s=%v", key, val))
			} else {
				standardFields = append(standardFields, fmt.Sprintf("%s=%v", key, val))
			}
		}
	}

	// Add standard fields first
	for _, field := range standardFields {
		fmt.Fprintf(l.out, " %s", field)
	}

	// Add context fields based on mode
	if len(contextFields) > 0 && l.contextMode != "none" {
		// Format context differently based on mode
		switch l.contextMode {
		case "debug":
			// Print each context dimension on its own
			fmt.Fprintf(l.out, " [CONTEXT: %s]", strings.Join(contextFields, ", "))
		case "compressed":
			// Print compressed context representation
			fmt.Fprintf(l.out, " [CTX]")
		default:
			// Default mode: just show that context is present
			fmt.Fprintf(l.out, " [7D]")
		}
	}

	// If we have a context object and are in debug mode, add it to the log
	if l.context != nil && l.contextMode == "debug" {
		// Only print context if we didn't already have context fields
		if len(contextFields) == 0 {
			fmt.Fprintf(l.out, " [7D-CTX: who=%s what=%s where=%s why=%s]",
				l.context.Who, l.context.What, l.context.Where, l.context.Why)
		}
	}

	fmt.Fprintln(l.out)
}

// Helper function to check if a field is a 7D context dimension
func isContextDimension(field string) bool {
	// WHO: ContextFieldValidator
	// WHAT: Validate context dimension
	// WHEN: During log formatting
	// WHERE: System Layer 6 (Integration)
	// WHY: To identify context fields
	// HOW: Using dimension name comparison
	// EXTENT: Log field processing

	dimensions := []string{
		"who", "what", "when", "where", "why", "how", "extent",
		"WHO", "WHAT", "WHEN", "WHERE", "WHY", "HOW", "EXTENT",
		"context", "Context", "CONTEXT",
	}

	for _, dim := range dimensions {
		if field == dim {
			return true
		}
	}
	return false
}

// Log methods for different levels

// Debug logs a debug message
func (l *Logger) Debug(msg string, keyvals ...interface{}) {
	// WHO: DebugLogger
	// WHAT: Log debug message
	// WHEN: During detailed diagnostics
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide detailed troubleshooting
	// HOW: Using debug level logging
	// EXTENT: Diagnostic operations

	l.log(Debug, msg, keyvals...)
}

// Info logs an info message
func (l *Logger) Info(msg string, keyvals ...interface{}) {
	// WHO: InfoLogger
	// WHAT: Log info message
	// WHEN: During normal operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide operational information
	// HOW: Using info level logging
	// EXTENT: Standard operations

	l.log(Info, msg, keyvals...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, keyvals ...interface{}) {
	// WHO: WarnLogger
	// WHAT: Log warning message
	// WHEN: During potential issues
	// WHERE: System Layer 6 (Integration)
	// WHY: To highlight potential problems
	// HOW: Using warning level logging
	// EXTENT: Warning conditions

	l.log(Warn, msg, keyvals...)
}

// Error logs an error message
func (l *Logger) Error(msg string, keyvals ...interface{}) {
	// WHO: ErrorLogger
	// WHAT: Log error message
	// WHEN: During error conditions
	// WHERE: System Layer 6 (Integration)
	// WHY: To record errors for resolution
	// HOW: Using error level logging
	// EXTENT: Error conditions

	l.log(Error, msg, keyvals...)
}

// Fatal logs a fatal message and exits the program
func (l *Logger) Fatal(msg string, keyvals ...interface{}) {
	// WHO: FatalLogger
	// WHAT: Log fatal message and exit
	// WHEN: During critical failures
	// WHERE: System Layer 6 (Integration)
	// WHY: To handle unrecoverable errors
	// HOW: Using fatal level logging with exit
	// EXTENT: Critical failures

	l.log(Fatal, msg, keyvals...)
	os.Exit(1)
}

// SetLevel changes the logging level
func (l *Logger) SetLevel(level Level) {
	// WHO: LevelManager
	// WHAT: Change logging level
	// WHEN: During runtime configuration
	// WHERE: System Layer 6 (Integration)
	// WHY: To adjust verbosity dynamically
	// HOW: Using level modification
	// EXTENT: Logger configuration

	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetLevelFromString changes the logging level from a string
func (l *Logger) SetLevelFromString(levelStr string) {
	// WHO: LevelConfigurator
	// WHAT: Set log level from string
	// WHEN: During runtime configuration
	// WHERE: System Layer 6 (Integration)
	// WHY: To adjust verbosity from config
	// HOW: Using level parsing
	// EXTENT: Logger configuration

	l.SetLevel(FromString(levelStr))
}

// CompressContext applies Möbius compression to the context
func CompressContext(context map[string]interface{}) map[string]interface{} {
	// WHO: ContextCompressor
	// WHAT: Context compression
	// WHEN: During logging
	// WHERE: System Layer 6 (Integration)
	// WHY: To reduce context size while preserving meaning
	// HOW: Using Möbius Compression Formula
	// EXTENT: Context lifecycle

	// In a real implementation, this would apply the Möbius Compression Formula
	// Formula: compressed = (value * B * I * (1 - (entropy / log2(1 + V))) * (G + F)) / (E * t + entropy + alignment)
	// alignment = (B + V * I) * exp(-t * E)

	// For now, we'll just return the original context with compression metadata
	compressed := make(map[string]interface{})
	compressed["original"] = context
	compressed["compressed"] = true
	compressed["compression_algorithm"] = "mobius"
	compressed["compression_timestamp"] = time.Now().Unix()

	return compressed
}

// Debugf formats and logs a debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	// WHO: FormatDebugLogger
	// WHAT: Format and log debug message
	// WHEN: During detailed diagnostics
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide formatted debug information
	// HOW: Using printf-style formatting
	// EXTENT: Diagnostic operations

	l.Debug(fmt.Sprintf(format, args...))
}

// Infof formats and logs an info message
func (l *Logger) Infof(format string, args ...interface{}) {
	// WHO: FormatInfoLogger
	// WHAT: Format and log info message
	// WHEN: During normal operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide formatted operational information
	// HOW: Using printf-style formatting
	// EXTENT: Standard operations

	l.Info(fmt.Sprintf(format, args...))
}

// Warnf formats and logs a warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	// WHO: FormatWarnLogger
	// WHAT: Format and log warning message
	// WHEN: During potential issues
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide formatted warning information
	// HOW: Using printf-style formatting
	// EXTENT: Warning conditions

	l.Warn(fmt.Sprintf(format, args...))
}

// Errorf formats and logs an error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	// WHO: FormatErrorLogger
	// WHAT: Format and log error message
	// WHEN: During error conditions
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide formatted error information
	// HOW: Using printf-style formatting
	// EXTENT: Error conditions

	l.Error(fmt.Sprintf(format, args...))
}

// Fatalf formats and logs a fatal message and exits the program
func (l *Logger) Fatalf(format string, args ...interface{}) {
	// WHO: FormatFatalLogger
	// WHAT: Format and log fatal message and exit
	// WHEN: During critical failures
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide formatted fatal information
	// HOW: Using printf-style formatting with exit
	// EXTENT: Critical failures

	l.Fatal(fmt.Sprintf(format, args...))
}

// GetCurrentContext returns a copy of the current context
func (l *Logger) GetCurrentContext() *ContextVector7D {
	// WHO: ContextProvider
	// WHAT: Get current context
	// WHEN: During context observation
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide access to current context
	// HOW: Using context copying
	// EXTENT: Context observation

	if l.context == nil {
		return nil
	}

	// Return a copy to prevent modification
	ctx := *l.context
	return &ctx
}

// UpdateContext updates the current context with new values
func (l *Logger) UpdateContext(updates map[string]interface{}) {
	// WHO: ContextUpdater
	// WHAT: Update context values
	// WHEN: During context modification
	// WHERE: System Layer 6 (Integration)
	// WHY: To modify contextual information
	// HOW: Using context field updating
	// EXTENT: Context maintenance

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.context == nil {
		return
	}

	// Update fields based on map
	for key, value := range updates {
		switch strings.ToLower(key) {
		case "who":
			if strVal, ok := value.(string); ok {
				l.context.Who = strVal
			}
		case "what":
			if strVal, ok := value.(string); ok {
				l.context.What = strVal
			}
		case "when":
			switch v := value.(type) {
			case int64:
				l.context.When = v
			case int:
				l.context.When = int64(v)
			case time.Time:
				l.context.When = v.Unix()
			}
		case "where":
			if strVal, ok := value.(string); ok {
				l.context.Where = strVal
			}
		case "why":
			if strVal, ok := value.(string); ok {
				l.context.Why = strVal
			}
		case "how":
			if strVal, ok := value.(string); ok {
				l.context.How = strVal
			}
		case "extent":
			switch v := value.(type) {
			case float64:
				l.context.Extent = v
			case float32:
				l.context.Extent = float64(v)
			case int:
				l.context.Extent = float64(v)
			}
		case "meta":
			if metaMap, ok := value.(map[string]interface{}); ok {
				if l.context.Meta == nil {
					l.context.Meta = metaMap
				} else {
					// Merge maps
					for k, v := range metaMap {
						l.context.Meta[k] = v
					}
				}
			}
		}
	}
}
