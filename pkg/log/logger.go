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
	"time"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/tranquilspeak"
)

type LogLevel int

// Logger provides a 7D context-aware logging facility, event-driven via TriggerMatrix
// All log entries are emitted as ATM triggers and routed through the TriggerMatrix for HemoFlux compression and helical memory storage
// No file output, no direct memory calls
// This logger is a thin event emitter

type Logger struct {
	level         int
	context       map[string]string
	triggerMatrix *tranquilspeak.TriggerMatrix
	dna           interface{} // Added for WithDNA/identity-aware logging
}

const (
	LevelDebug = 0
	LevelInfo  = 1
	LevelWarn  = 2
	LevelError = 3
	LevelCritical = 4
)

// NewLogger creates a new 7D-aware logger with helical memory integration
func NewLogger(triggerMatrix *tranquilspeak.TriggerMatrix) *Logger {
	return &Logger{
		level:   LevelInfo,
		context: map[string]string{"who": "System", "what": "Log", "where": "TNOS_MCP_Bridge", "why": "SystemMonitoring", "how": "LoggerComponent", "extent": "SystemWide"},
		triggerMatrix: triggerMatrix,
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
func NewNopLogger(triggerMatrix *tranquilspeak.TriggerMatrix) *Logger {
	logger := NewLogger(triggerMatrix)
	logger.level = LevelCritical // Only log critical messages
	return logger
}

// WithLevel sets the minimum log level
func (l *Logger) WithLevel(level int) *Logger {
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

// WithDNA sets the DNA/identity context for the logger (required by LoggerInterface)
func (l *Logger) WithDNA(dna interface{}) LoggerInterface {
	newLogger := *l
	newLogger.dna = dna
	return &newLogger
}

// log writes a message at the specified level
func (l *Logger) log(level int, message string, args ...interface{}) {
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

	msg := message
	if len(args) > 0 {
		msg = fmt.Sprintf(message, args...)
	}
	// Build 7D context
	ctx7d := map[string]interface{}{
		"who":    l.context["who"],
		"what":   l.context["what"],
		"when":   time.Now().Unix(),
		"where":  l.context["where"],
		"why":    l.context["why"],
		"how":    l.context["how"],
		"extent": l.context["extent"],
	}
	payload := map[string]interface{}{
		"level":   level,
		"context": ctx7d,
		"message": msg,
		"time":    time.Now().Format(time.RFC3339),
	}
	atmTrigger := tranquilspeak.CreateTrigger(
		l.context["who"], l.context["what"], l.context["where"], l.context["why"], l.context["how"], l.context["extent"],
		tranquilspeak.TriggerTypeSystemControl, "helical_memory", payload,
	)
	_ = l.triggerMatrix.ProcessTrigger(atmTrigger)
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

// Identity-aware logging methods required by LoggerInterface

// InfoWithIdentity logs an info message with DNA/identity context
func (l *Logger) InfoWithIdentity(message string, dna interface{}, args ...interface{}) {
	sig := ""
	if dna != nil {
		if s, ok := getDNASignature(dna); ok {
			sig = s
		}
	} else if l.dna != nil {
		if s, ok := getDNASignature(l.dna); ok {
			sig = s
		}
	}
	identityMessage := fmt.Sprintf("[DNA:%s] %s", sig, message)
	l.log(LevelInfo, identityMessage, args...)
}

// DebugWithIdentity logs a debug message with DNA/identity context
func (l *Logger) DebugWithIdentity(message string, dna interface{}, args ...interface{}) {
	sig := ""
	if dna != nil {
		if s, ok := getDNASignature(dna); ok {
			sig = s
		}
	} else if l.dna != nil {
		if s, ok := getDNASignature(l.dna); ok {
			sig = s
		}
	}
	identityMessage := fmt.Sprintf("[DNA:%s] %s", sig, message)
	l.log(LevelDebug, identityMessage, args...)
}

// ErrorWithIdentity logs an error message with DNA/identity context
func (l *Logger) ErrorWithIdentity(message string, dna interface{}, args ...interface{}) {
	sig := ""
	if dna != nil {
		if s, ok := getDNASignature(dna); ok {
			sig = s
		}
	} else if l.dna != nil {
		if s, ok := getDNASignature(l.dna); ok {
			sig = s
		}
	}
	identityMessage := fmt.Sprintf("[DNA:%s] %s", sig, message)
	l.log(LevelError, identityMessage, args...)
}

// CriticalWithIdentity logs a critical message with DNA/identity context
func (l *Logger) CriticalWithIdentity(message string, dna interface{}, args ...interface{}) {
	sig := ""
	if dna != nil {
		if s, ok := getDNASignature(dna); ok {
			sig = s
		}
	} else if l.dna != nil {
		if s, ok := getDNASignature(l.dna); ok {
			sig = s
		}
	}
	identityMessage := fmt.Sprintf("[DNA:%s] %s", sig, message)
	l.log(LevelCritical, identityMessage, args...)
}

// getDNASignature attempts to extract a signature from a DNA-like object
func getDNASignature(dna interface{}) (string, bool) {
	if dna == nil {
		return "", false
	}
	if s, ok := dna.(interface{ Signature() string }); ok {
		return s.Signature(), true
	}
	return "", false
}

// LogWithTEI logs with TEI (Tranquility Engine Identity) awareness
func (l *Logger) LogWithTEI(level LogLevel, actionName string, message string, args ...interface{}) {
	teiMessage := fmt.Sprintf("[TEI:%s] %s", actionName, message)
	l.log(int(level), teiMessage, args...)
}

// Update logger configuration based on mode
func UpdateLoggerMode(mode string, triggerMatrix *tranquilspeak.TriggerMatrix) {
	logger := NewLogger(triggerMatrix).WithLevel(LevelInfo)
	switch mode {
	case "standalone":
		logger.Info("Logger set to standalone mode")
	case "blood-connected":
		logger.Info("Logger set to blood-connected mode")
	}
}
