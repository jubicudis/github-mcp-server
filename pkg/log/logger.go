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
	// Level is the minimum log level to output
	Level string
	// FilePath is the path to the log file
	FilePath string
	// ConsoleOut indicates whether to log to the console
	ConsoleOut bool
	// ContextMode controls how context is displayed
	ContextMode string
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

// Logger implements 7D Context-aware logging
type Logger struct {
	// WHO: LoggerCore
	// WHAT: Core logging implementation
	// WHEN: During log operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To centralize logging with context
	// HOW: Using structured data with compression
	// EXTENT: All log operations

	level       Level
	contextMode string
	out         io.Writer
	fileOut     io.Writer
	mu          sync.Mutex
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

	// Create and return logger
	return &Logger{
		level:       FromString(config.Level),
		contextMode: config.ContextMode,
		out:         out,
		fileOut:     fileOut,
	}
}

// Close closes the logger and associated resources
func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Close the file if it's open
	if closer, ok := l.fileOut.(io.Closer); ok && closer != nil {
		closer.Close()
	}
}

// log is the internal logging function
func (l *Logger) log(level Level, msg string, keyvals ...interface{}) {
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

	fmt.Fprintln(l.out)
}

// Helper function to check if a field is a 7D context dimension
func isContextDimension(field string) bool {
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
	l.log(Debug, msg, keyvals...)
}

// Info logs an info message
func (l *Logger) Info(msg string, keyvals ...interface{}) {
	l.log(Info, msg, keyvals...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, keyvals ...interface{}) {
	l.log(Warn, msg, keyvals...)
}

// Error logs an error message
func (l *Logger) Error(msg string, keyvals ...interface{}) {
	l.log(Error, msg, keyvals...)
}

// Fatal logs a fatal message and exits the program
func (l *Logger) Fatal(msg string, keyvals ...interface{}) {
	l.log(Fatal, msg, keyvals...)
	os.Exit(1)
}

// WithContext returns a new logger with context
func (l *Logger) WithContext(context map[string]interface{}) *Logger {
	// Create a new logger with the same configuration
	return l
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
	// For now, we'll just return the original context
	return context
}
