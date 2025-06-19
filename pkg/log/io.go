/*
 * WHO: LogIO
 * WHAT: Log I/O operations with 7D Context awareness
 * WHEN: During log file operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide context-aware I/O for logging
 * HOW: Using Go's io interfaces with context enrichment
 * EXTENT: All log file operations
 */

package log

import (
	"fmt"
	"io"
	"sync"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/tranquilspeak"
)

// All logging and error output in this file is routed through the TNOS 7D-aware logger infrastructure.
// This ensures compliance with the TNOS 7D architecture, biological mimicry, and formula registry best practices.
// See docs/architecture/7D_CONTEXT_FRAMEWORK.md and docs/generated/mapping_registry.json for details.

// WHO: LoggerInterface
// WHAT: Interface for logging operations
// WHEN: During logging operations
// WHERE: System Layer 6 (Integration)
// WHY: To provide a common interface for different logger implementations
// HOW: Using Go's interface mechanism
// EXTENT: All logger implementations
type LoggerInterface interface {
	Info(message string, args ...interface{})
	Debug(message string, args ...interface{}) // Added Debug
	Warn(message string, args ...interface{})  // Added Warn
	Error(message string, args ...interface{}) // Added Error
	// DNA/identity-aware logging methods
	InfoWithIdentity(message string, dna interface{}, args ...interface{})
	DebugWithIdentity(message string, dna interface{}, args ...interface{})
	ErrorWithIdentity(message string, dna interface{}, args ...interface{})
	CriticalWithIdentity(message string, dna interface{}, args ...interface{})
	WithDNA(dna interface{}) LoggerInterface
}

// WHO: ContextVector7D
// WHAT: 7D Context vector representation
// WHEN: During context handling operations
// WHERE: System Layer 6 (Integration)
// WHY: To maintain complete contextual awareness
// HOW: Using structured context parameters
// EXTENT: All context operations
type ContextVector7D struct {
	Who    string                 `json:"who"`
	What   string                 `json:"what"`
	When   interface{}            `json:"when"`
	Where  string                 `json:"where"`
	Why    string                 `json:"why"`
	How    string                 `json:"how"`
	Extent float64                `json:"extent"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
	Source string                 `json:"source,omitempty"`
}

// ToMap converts a ContextVector7D to a map[string]interface{} for serialization or neural mapping
func ToMap(cv ContextVector7D) map[string]interface{} {
	result := map[string]interface{}{
		"who":    cv.Who,
		"what":   cv.What,
		"when":   cv.When,
		"where":  cv.Where,
		"why":    cv.Why,
		"how":    cv.How,
		"extent": cv.Extent,
	}
	if cv.Source != "" {
		result["source"] = cv.Source
	}
	if len(cv.Meta) > 0 {
		result["meta"] = cv.Meta
	}
	return result
}

// FromMap converts a map[string]interface{} to a ContextVector7D for deserialization or neural mapping
func FromMap(m map[string]interface{}) ContextVector7D {
	cv := ContextVector7D{}
	if v, ok := m["who"].(string); ok {
		cv.Who = v
	}
	if v, ok := m["what"].(string); ok {
		cv.What = v
	}
	if v, ok := m["when"]; ok {
		cv.When = v
	}
	if v, ok := m["where"].(string); ok {
		cv.Where = v
	}
	if v, ok := m["why"].(string); ok {
		cv.Why = v
	}
	if v, ok := m["how"].(string); ok {
		cv.How = v
	}
	if v, ok := m["extent"].(float64); ok {
		cv.Extent = v
	}
	if v, ok := m["meta"].(map[string]interface{}); ok {
		cv.Meta = v
	}
	if v, ok := m["source"].(string); ok {
		cv.Source = v
	}
	return cv
}

// The LoggerInterface is defined above
// We use that interface here for IO logging operations

// IOLogger is a wrapper around io.Reader and io.Writer that can be used
// to log the data being read and written from the underlying streams
// WHO: IOLogger
// WHAT: Logging of IO operations
// WHEN: During data transfer
// WHERE: System Layer 6 (Integration)
// WHY: To provide visibility into data flows
// HOW: Using decorator pattern around io interfaces
// EXTENT: All monitored IO operations
type IOLogger struct {
	reader io.Reader
	writer io.Writer
	logger LoggerInterface
}

// NewIOLogger creates a new IOLogger instance
// WHO: LoggerFactory
// WHAT: IO Logger creation
// WHEN: During logger initialization
// WHERE: System Layer 6 (Integration)
// WHY: To facilitate IO monitoring
// HOW: Using composition of IO interfaces
// EXTENT: Logger instance lifecycle
func NewIOLogger(r io.Reader, w io.Writer, logger LoggerInterface) *IOLogger {
	return &IOLogger{
		reader: r,
		writer: w,
		logger: logger,
	}
}

// Read reads data from the underlying io.Reader and logs it.
func (l *IOLogger) Read(p []byte) (n int, err error) {
	if l.reader == nil {
		return 0, io.EOF
	}
	n, err = l.reader.Read(p)
	if n > 0 {
		l.logger.Info(fmt.Sprintf("[stdin]: received %d bytes: %s", n, string(p[:n])))
	}
	return n, err
}

// Write writes data to the underlying io.Writer and logs it.
func (l *IOLogger) Write(p []byte) (n int, err error) {
	if l.writer == nil {
		return 0, io.ErrClosedPipe
	}
	l.logger.Info(fmt.Sprintf("[stdout]: sending %d bytes: %s", len(p), string(p)))
	return l.writer.Write(p)
}

// ContextWriter extends io.Writer with context awareness
type ContextWriter struct {
	// WHO: ContextualWriter
	// WHAT: Context-aware writer implementation
	// WHEN: During write operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To add context to write operations
	// HOW: Using io.Writer with context
	// EXTENT: All log write operations

	writer  io.Writer
	context *ContextVector7D
	mu      sync.Mutex
}

// NewContextWriter creates a new context-aware writer
func NewContextWriter(w io.Writer, context *ContextVector7D) *ContextWriter {
	// WHO: WriterFactory
	// WHAT: Create context writer
	// WHEN: During writer initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide contextualized writer
	// HOW: Using composition pattern
	// EXTENT: Writer lifecycle

	return &ContextWriter{
		writer:  w,
		context: context,
	}
}

// Write implements io.Writer with context
func (cw *ContextWriter) Write(p []byte) (n int, err error) {
	// WHO: ContextualByteWriter
	// WHAT: Write bytes with context
	// WHEN: During write operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To contextualize written data
	// HOW: Using decorated write with context
	// EXTENT: All byte write operations

	cw.mu.Lock()
	defer cw.mu.Unlock()

	// Simply delegate to the underlying writer for now
	// In a more advanced implementation, this could add context
	// information to the written data
	return cw.writer.Write(p)
}

// Close implements io.Closer for proper resource cleanup
func (cw *ContextWriter) Close() error {
	// WHO: ResourceManager
	// WHAT: Close writer resources
	// WHEN: During shutdown
	// WHERE: System Layer 6 (Integration)
	// WHY: To clean up resources
	// HOW: Using io.Closer interface
	// EXTENT: Resource lifecycle

	cw.mu.Lock()
	defer cw.mu.Unlock()

	// If the underlying writer is also a Closer, close it
	if closer, ok := cw.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// SetContext updates the writer's context
func (cw *ContextWriter) SetContext(context *ContextVector7D) {
	// WHO: ContextManager
	// WHAT: Update writer context
	// WHEN: During context change
	// WHERE: System Layer 6 (Integration)
	// WHY: To modify operational context
	// HOW: Using context replacement
	// EXTENT: Context lifecycle

	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.context = context
}

// MultiContextWriter allows writing to multiple writers with context
type MultiContextWriter struct {
	// WHO: MultipleWriterManager
	// WHAT: Manage multiple context writers
	// WHEN: During multi-destination logging
	// WHERE: System Layer 6 (Integration)
	// WHY: To write to multiple destinations
	// HOW: Using writer aggregation
	// EXTENT: Multi-destination operations

	writers []io.Writer
	context *ContextVector7D
	mu      sync.Mutex
}

// NewMultiContextWriter creates a new multi-context writer
func NewMultiContextWriter(context *ContextVector7D, writers ...io.Writer) *MultiContextWriter {
	// WHO: MultiWriterFactory
	// WHAT: Create multi-writer
	// WHEN: During writer initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide multi-destination logging
	// HOW: Using factory pattern
	// EXTENT: Writer lifecycle

	return &MultiContextWriter{
		writers: writers,
		context: context,
	}
}

// Write implements io.Writer by writing to all writers
func (mw *MultiContextWriter) Write(p []byte) (n int, err error) {
	// WHO: MultiplexingWriter
	// WHAT: Write to multiple destinations
	// WHEN: During write operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To distribute log data
	// HOW: Using write broadcasting
	// EXTENT: All write operations

	mw.mu.Lock()
	defer mw.mu.Unlock()

	// Write to all writers, but return the first error
	// This follows the behavior of io.MultiWriter
	for _, w := range mw.writers {
		n, err := w.Write(p)
		if err != nil {
			return n, err
		}
		if n != len(p) {
			return n, io.ErrShortWrite
		}
	}
	return len(p), nil
}

// Close implements io.Closer by closing all writers
func (mw *MultiContextWriter) Close() error {
	// WHO: MultiResourceManager
	// WHAT: Close multiple resources
	// WHEN: During shutdown
	// WHERE: System Layer 6 (Integration)
	// WHY: To clean up multiple resources
	// HOW: Using sequential closure
	// EXTENT: Resource cleanup

	mw.mu.Lock()
	defer mw.mu.Unlock()

	var lastErr error
	for _, w := range mw.writers {
		if closer, ok := w.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				lastErr = err
			}
		}
	}
	return lastErr
}

// SetContext updates the writer's context and propagates to all context-aware writers
func (mw *MultiContextWriter) SetContext(context *ContextVector7D) {
	// WHO: MultiContextManager
	// WHAT: Update multiple contexts
	// WHEN: During context change
	// WHERE: System Layer 6 (Integration)
	// WHY: To propagate context changes
	// HOW: Using context broadcasting
	// EXTENT: Context propagation

	mw.mu.Lock()
	defer mw.mu.Unlock()

	mw.context = context

	// Propagate context to any ContextWriters
	for _, w := range mw.writers {
		if cw, ok := w.(*ContextWriter); ok {
			cw.SetContext(context)
		}
	}
}

// BufferedContextWriter provides buffered writing with context
type BufferedContextWriter struct {
	// WHO: BufferedWriteManager
	// WHAT: Buffer write operations
	// WHEN: During buffered logging
	// WHERE: System Layer 6 (Integration)
	// WHY: To improve write performance
	// HOW: Using write buffering
	// EXTENT: Buffered operations

	writer  io.Writer
	buffer  []byte
	context *ContextVector7D
	mu      sync.Mutex
	size    int
}

// NewBufferedContextWriter creates a new buffered context writer
func NewBufferedContextWriter(w io.Writer, size int, context *ContextVector7D) *BufferedContextWriter {
	// WHO: BufferedWriterFactory
	// WHAT: Create buffered writer
	// WHEN: During writer initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide buffered logging
	// HOW: Using factory pattern
	// EXTENT: Writer lifecycle

	return &BufferedContextWriter{
		writer:  w,
		buffer:  make([]byte, 0, size),
		context: context,
		size:    size,
	}
}

// Write implements io.Writer with buffering
func (bw *BufferedContextWriter) Write(p []byte) (n int, err error) {
	// WHO: BufferingWriter
	// WHAT: Buffer write operations
	// WHEN: During write operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To batch write operations
	// HOW: Using memory buffering
	// EXTENT: All buffered writes

	bw.mu.Lock()
	defer bw.mu.Unlock()

	// If the data exceeds the buffer, flush the buffer and write directly
	if len(p) >= bw.size {
		if err := bw.flushLocked(); err != nil {
			return 0, err
		}
		return bw.writer.Write(p)
	}

	// If adding the data would overflow the buffer, flush first
	if len(bw.buffer)+len(p) > bw.size {
		if err := bw.flushLocked(); err != nil {
			return 0, err
		}
	}

	// Append to buffer
	bw.buffer = append(bw.buffer, p...)
	return len(p), nil
}

// Flush writes any buffered data to the underlying writer
func (bw *BufferedContextWriter) Flush() error {
	// WHO: BufferFlusher
	// WHAT: Flush buffer contents
	// WHEN: During flush operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To ensure data persistence
	// HOW: Using forced write
	// EXTENT: Buffer flushing

	bw.mu.Lock()
	defer bw.mu.Unlock()
	return bw.flushLocked()
}

// flushLocked writes any buffered data to the underlying writer (assumes lock is held)
func (bw *BufferedContextWriter) flushLocked() error {
	// WHO: LockedBufferFlusher
	// WHAT: Flush with lock held
	// WHEN: During locked flush
	// WHERE: System Layer 6 (Integration)
	// WHY: To perform atomic flush
	// HOW: Using direct write
	// EXTENT: Internal buffer operations

	if len(bw.buffer) == 0 {
		return nil
	}

	_, err := bw.writer.Write(bw.buffer)
	bw.buffer = bw.buffer[:0] // Clear buffer but preserve capacity
	return err
}

// Close implements io.Closer
func (bw *BufferedContextWriter) Close() error {
	// WHO: BufferedResourceManager
	// WHAT: Close buffered resources
	// WHEN: During shutdown
	// WHERE: System Layer 6 (Integration)
	// WHY: To ensure all data is written
	// HOW: Using flush and close
	// EXTENT: Resource cleanup

	bw.mu.Lock()
	defer bw.mu.Unlock()

	// Flush any remaining data
	if err := bw.flushLocked(); err != nil {
		return err
	}

	// Close underlying writer if it's a closer
	if closer, ok := bw.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// SetContext updates the writer's context
func (bw *BufferedContextWriter) SetContext(context *ContextVector7D) {
	// WHO: BufferedContextManager
	// WHAT: Update buffered context
	// WHEN: During context change
	// WHERE: System Layer 6 (Integration)
	// WHY: To modify operational context
	// HOW: Using context replacement
	// EXTENT: Context lifecycle

	bw.mu.Lock()
	defer bw.mu.Unlock()
	bw.context = context

	// Propagate context to underlying writer if it's a context writer
	if cw, ok := bw.writer.(*ContextWriter); ok {
		cw.SetContext(context)
	} else if mcw, ok := bw.writer.(*MultiContextWriter); ok {
		mcw.SetContext(context)
	}
}

// Compatibility logic for bridge and subcomponent needs
func ConfigureIOForBridge(mode string, triggerMatrix *tranquilspeak.TriggerMatrix) {
	// Use unified logger instead of stdout prints
	logger := NewNopLogger(triggerMatrix).WithLevel(LevelInfo) // Event-driven only
	if mode == "standalone" {
		logger.Info("I/O configured for standalone mode")
	} else if mode == "blood-connected" {
		logger.Info("I/O configured for blood-connected mode")
	}
}

// Remove all direct file and disk I/O logic. All log I/O must be event-driven via the logger and TriggerMatrix.
// Remove any file rotation, file output, or direct disk I/O. Only event-driven log emission is allowed.
