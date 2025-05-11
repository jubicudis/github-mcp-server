/*
 * WHO: LogIOTester
 * WHAT: Tests for log I/O operations with 7D Context awareness
 * WHEN: During test execution
 * WHERE: System Layer 6 (Integration) / Test Environment
 * WHY: To validate context-aware I/O for logging
 * HOW: Using Go's testing framework with context validation
 * EXTENT: All log I/O components
 */

package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// WHO: TestUtility
// WHAT: Create test context
// WHEN: During test preparation
// WHERE: Test Environment
// WHY: To standardize test contexts
// HOW: Using factory function
// EXTENT: Test setup
func createTestContext() *ContextVector7D {
	return &ContextVector7D{
		Who:    "TestComponent",
		What:   "LoggingOperation",
		When:   time.Now().Unix(),
		Where:  "TestEnvironment",
		Why:    "ValidationPurpose",
		How:    "AutomatedTesting",
		Extent: 1.0,
		Meta: map[string]interface{}{
			"testID": fmt.Sprintf("test-%d", time.Now().UnixNano()),
			"B":      0.8, // Base factor for compression
			"V":      0.7, // Value factor
			"I":      0.9, // Intent factor
		},
		Source: "test_framework",
	}
}

// TestIOLogger tests the IOLogger functionality
// WHO: IOLoggerTester
// WHAT: Test IO Logger
// WHEN: During test execution
// WHERE: Test Environment
// WHY: To validate IO logging
// HOW: Using mock reader/writer
// EXTENT: IOLogger component
func TestIOLogger(t *testing.T) {
	// Set up test data
	testInput := []byte("test input data")
	testOutput := []byte("test output data")

	// Create buffers for testing
	inputBuf := bytes.NewBuffer(testInput)
	outputBuf := &bytes.Buffer{}

	// Create a mock logger that captures log messages
	var logMessages []string
	mockLogger := &MockLogger{
		InfofFunc: func(format string, args ...interface{}) {
			logMessages = append(logMessages, fmt.Sprintf(format, args...))
		},
	}

	// Create the IOLogger with our mock components
	ioLogger := NewIOLogger(inputBuf, outputBuf, mockLogger)

	// Test Read
	readBuf := make([]byte, len(testInput)+5) // Buffer larger than input
	n, err := ioLogger.Read(readBuf)

	if err != nil && err != io.EOF {
		t.Errorf("Expected no error or EOF, got: %v", err)
	}

	if n != len(testInput) {
		t.Errorf("Expected to read %d bytes, got %d", len(testInput), n)
	}

	if !bytes.Equal(readBuf[:n], testInput) {
		t.Errorf("Expected to read %q, got %q", testInput, readBuf[:n])
	}

	// Verify read was logged
	if len(logMessages) != 1 {
		t.Errorf("Expected 1 log message, got %d", len(logMessages))
	}

	expectedLogPrefix := "[stdin]: received"
	if !strings.Contains(logMessages[0], expectedLogPrefix) {
		t.Errorf("Expected log message to contain %q, got: %q",
			expectedLogPrefix, logMessages[0])
	}

	// Test Write
	n, err = ioLogger.Write(testOutput)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if n != len(testOutput) {
		t.Errorf("Expected to write %d bytes, got %d", len(testOutput), n)
	}

	if !bytes.Equal(outputBuf.Bytes(), testOutput) {
		t.Errorf("Expected buffer to contain %q, got %q", testOutput, outputBuf.Bytes())
	}

	// Verify write was logged
	if len(logMessages) != 2 {
		t.Errorf("Expected 2 log messages, got %d", len(logMessages))
	}

	expectedLogPrefix = "[stdout]: sending"
	if !strings.Contains(logMessages[1], expectedLogPrefix) {
		t.Errorf("Expected log message to contain %q, got: %q",
			expectedLogPrefix, logMessages[1])
	}
}

// TestContextWriter tests the ContextWriter functionality
// WHO: ContextWriterTester
// WHAT: Test Context Writer
// WHEN: During test execution
// WHERE: Test Environment
// WHY: To validate context-aware writing
// HOW: Using buffer and context comparison
// EXTENT: ContextWriter component
func TestContextWriter(t *testing.T) {
	// Create test data
	testData := []byte("context-aware test data")
	outputBuf := &bytes.Buffer{}

	// Create initial context
	initialContext := createTestContext()

	// Create context writer
	contextWriter := NewContextWriter(outputBuf, initialContext)

	// Write data
	n, err := contextWriter.Write(testData)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if n != len(testData) {
		t.Errorf("Expected to write %d bytes, got %d", len(testData), n)
	}

	if !bytes.Equal(outputBuf.Bytes(), testData) {
		t.Errorf("Expected buffer to contain %q, got %q", testData, outputBuf.Bytes())
	}

	// Update context and write again
	updatedContext := createTestContext()
	updatedContext.What = "UpdatedOperation"
	updatedContext.Why = "UpdatedPurpose"

	contextWriter.SetContext(updatedContext)

	// Write more data
	moreData := []byte("more context-aware data")

	n, err = contextWriter.Write(moreData)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expectedCombined := append(testData, moreData...)
	if !bytes.Equal(outputBuf.Bytes(), expectedCombined) {
		t.Errorf("Expected buffer to contain %q, got %q", expectedCombined, outputBuf.Bytes())
	}

	// Test Close
	err = contextWriter.Close()
	if err != nil {
		t.Errorf("Expected no error on close, got: %v", err)
	}
}

// TestRotatingFileWriter tests the RotatingFileWriter functionality
// WHO: RotatingWriterTester
// WHAT: Test Log Rotation
// WHEN: During test execution
// WHERE: Test Environment
// WHY: To validate log rotation
// HOW: Using temp files and size/time triggers
// EXTENT: RotatingFileWriter component
func TestRotatingFileWriter(t *testing.T) {
	// Create temp directory for test logs
	tempDir, err := os.MkdirTemp("", "log_rotation_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up test parameters
	baseFilename := filepath.Join(tempDir, "test.log")
	maxSize := int64(100)           // Small size to trigger rotation
	maxAge := 10 * time.Millisecond // Short duration to trigger time-based rotation

	// Create rotating writer
	writer, err := NewRotatingFileWriter(baseFilename, maxSize, maxAge)
	if err != nil {
		t.Fatalf("Failed to create rotating writer: %v", err)
	}
	defer writer.Close()

	// Write data to trigger size-based rotation
	data := bytes.Repeat([]byte("test data that will eventually cause rotation\n"), 3)

	// Write first batch
	n, err := writer.Write(data)
	if err != nil {
		t.Errorf("Failed to write data: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, got %d", len(data), n)
	}

	// Verify file exists
	if _, err := os.Stat(baseFilename); os.IsNotExist(err) {
		t.Errorf("Log file doesn't exist")
	}

	// Write second batch - should trigger rotation
	n, err = writer.Write(data)
	if err != nil {
		t.Errorf("Failed to write data: %v", err)
	}

	// Check for rotated files
	files, err := filepath.Glob(filepath.Join(tempDir, "test.log.*"))
	if err != nil {
		t.Errorf("Failed to list log files: %v", err)
	}

	if len(files) < 1 {
		t.Errorf("Expected at least one rotated log file, found none")
	}

	// Test time-based rotation
	time.Sleep(maxAge * 2)

	// Write more data - should trigger time-based rotation
	n, err = writer.Write([]byte("small write after time delay"))
	if err != nil {
		t.Errorf("Failed to write data: %v", err)
	}

	// Check for additional rotated files
	newFiles, err := filepath.Glob(filepath.Join(tempDir, "test.log.*"))
	if err != nil {
		t.Errorf("Failed to list log files: %v", err)
	}

	if len(newFiles) <= len(files) {
		t.Errorf("Expected more rotated files after time delay, count remained at %d", len(files))
	}
}

// TestMultiContextWriter tests the MultiContextWriter functionality
// WHO: MultiWriterTester
// WHAT: Test Multi-Context Writer
// WHEN: During test execution
// WHERE: Test Environment
// WHY: To validate multi-destination writing
// HOW: Using multiple buffers
// EXTENT: MultiContextWriter component
func TestMultiContextWriter(t *testing.T) {
	// Create test data and buffers
	testData := []byte("multi-context test data")
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}

	// Create context
	context := createTestContext()

	// Create multi-context writer
	multiWriter := NewMultiContextWriter(context, buf1, buf2)

	// Write data
	n, err := multiWriter.Write(testData)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if n != len(testData) {
		t.Errorf("Expected to write %d bytes, got %d", len(testData), n)
	}

	// Verify data was written to both buffers
	if !bytes.Equal(buf1.Bytes(), testData) {
		t.Errorf("Expected buffer 1 to contain %q, got %q", testData, buf1.Bytes())
	}

	if !bytes.Equal(buf2.Bytes(), testData) {
		t.Errorf("Expected buffer 2 to contain %q, got %q", testData, buf2.Bytes())
	}

	// Update context
	updatedContext := createTestContext()
	updatedContext.What = "UpdatedMultiOperation"
	multiWriter.SetContext(updatedContext)

	// Add a context writer to test propagation
	buf3 := &bytes.Buffer{}
	contextWriter := NewContextWriter(buf3, context)

	// Create new multi writer with the context writer
	newMultiWriter := NewMultiContextWriter(updatedContext, contextWriter)

	// Write data
	moreData := []byte("more multi-context data")
	n, err = newMultiWriter.Write(moreData)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify data was written to the context writer's buffer
	if !bytes.Equal(buf3.Bytes(), moreData) {
		t.Errorf("Expected buffer 3 to contain %q, got %q", moreData, buf3.Bytes())
	}

	// Close both writers
	if err := multiWriter.Close(); err != nil {
		t.Errorf("Expected no error on close, got: %v", err)
	}

	if err := newMultiWriter.Close(); err != nil {
		t.Errorf("Expected no error on close, got: %v", err)
	}
}

// TestBufferedContextWriter tests the BufferedContextWriter functionality
// WHO: BufferedWriterTester
// WHAT: Test Buffered Writer
// WHEN: During test execution
// WHERE: Test Environment
// WHY: To validate buffered writing
// HOW: Using buffer size triggers
// EXTENT: BufferedContextWriter component
func TestBufferedContextWriter(t *testing.T) {
	// Create test data
	smallData := []byte("small data")
	largeData := bytes.Repeat([]byte("large data that exceeds buffer size"), 10)

	// Create output buffer and context
	outputBuf := &bytes.Buffer{}
	context := createTestContext()

	// Create buffered writer with small buffer size
	bufSize := 50
	bufferedWriter := NewBufferedContextWriter(outputBuf, bufSize, context)

	// Write small data (should be buffered, not written yet)
	n, err := bufferedWriter.Write(smallData)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if n != len(smallData) {
		t.Errorf("Expected to write %d bytes, got %d", len(smallData), n)
	}

	// Output buffer should be empty as data is still in internal buffer
	if outputBuf.Len() != 0 {
		t.Errorf("Expected output buffer to be empty, got %d bytes", outputBuf.Len())
	}

	// Flush the buffer
	err = bufferedWriter.Flush()
	if err != nil {
		t.Errorf("Expected no error on flush, got: %v", err)
	}

	// Output buffer should now contain the data
	if !bytes.Equal(outputBuf.Bytes(), smallData) {
		t.Errorf("Expected output buffer to contain %q, got %q", smallData, outputBuf.Bytes())
	}

	// Reset output buffer
	outputBuf.Reset()

	// Write large data (should exceed buffer and write directly)
	n, err = bufferedWriter.Write(largeData)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if n != len(largeData) {
		t.Errorf("Expected to write %d bytes, got %d", len(largeData), n)
	}

	// Output buffer should contain the large data
	if !bytes.Equal(outputBuf.Bytes(), largeData) {
		t.Errorf("Expected output buffer to contain large data of length %d, got %d",
			len(largeData), len(outputBuf.Bytes()))
	}

	// Update context
	updatedContext := createTestContext()
	updatedContext.What = "UpdatedBufferedOperation"
	bufferedWriter.SetContext(updatedContext)

	// Test close (should flush any remaining data)
	outputBuf.Reset()
	bufferedWriter.Write(smallData) // Write more data

	err = bufferedWriter.Close()
	if err != nil {
		t.Errorf("Expected no error on close, got: %v", err)
	}

	// Output buffer should contain the small data
	if !bytes.Equal(outputBuf.Bytes(), smallData) {
		t.Errorf("Expected output buffer to contain %q, got %q", smallData, outputBuf.Bytes())
	}
}

// MockLogger is a mock implementation of the Logger interface for testing
// WHO: MockLogComponent
// WHAT: Mock Logger Implementation
// WHEN: During testing
// WHERE: Test Environment
// WHY: To simulate logger behavior
// HOW: Using function fields
// EXTENT: Testing purposes only
type MockLogger struct {
	DebugFunc       func(args ...interface{})
	DebugfFunc      func(format string, args ...interface{})
	InfoFunc        func(args ...interface{})
	InfofFunc       func(format string, args ...interface{})
	WarnFunc        func(args ...interface{})
	WarnfFunc       func(format string, args ...interface{})
	ErrorFunc       func(args ...interface{})
	ErrorfFunc      func(format string, args ...interface{})
	FatalFunc       func(args ...interface{})
	FatalfFunc      func(format string, args ...interface{})
	PanicFunc       func(args ...interface{})
	PanicfFunc      func(format string, args ...interface{})
	WithFunc        func(key string, value interface{}) Logger
	WithMapFunc     func(fields map[string]interface{}) Logger
	WithErrorFunc   func(err error) Logger
	WithContextFunc func(ctx interface{}) Logger
}

// MockLogger implements the Logger interface
// Add a compile-time check to ensure MockLogger implements Logger
var _ Logger = (*MockLogger)(nil)

// This type assertion verifies that MockLogger implements the Logger interface

// Ensure MockLogger implements all required methods of Logger interface
func (m *MockLogger) Init() error {
	return nil
}

func (m *MockLogger) Debug(args ...interface{}) {
	if m.DebugFunc != nil {
		m.DebugFunc(args...)
	}
}

func (m *MockLogger) Debugf(format string, args ...interface{}) {
	if m.DebugfFunc != nil {
		m.DebugfFunc(format, args...)
	}
}

func (m *MockLogger) Info(args ...interface{}) {
	if m.InfoFunc != nil {
		m.InfoFunc(args...)
	}
}

func (m *MockLogger) Infof(format string, args ...interface{}) {
	if m.InfofFunc != nil {
		m.InfofFunc(format, args...)
	}
}

func (m *MockLogger) Warn(args ...interface{}) {
	if m.WarnFunc != nil {
		m.WarnFunc(args...)
	}
}

func (m *MockLogger) Warnf(format string, args ...interface{}) {
	if m.WarnfFunc != nil {
		m.WarnfFunc(format, args...)
	}
}

func (m *MockLogger) Error(args ...interface{}) {
	if m.ErrorFunc != nil {
		m.ErrorFunc(args...)
	}
}

func (m *MockLogger) Errorf(format string, args ...interface{}) {
	if m.ErrorfFunc != nil {
		m.ErrorfFunc(format, args...)
	}
}

func (m *MockLogger) Fatal(args ...interface{}) {
	if m.FatalFunc != nil {
		m.FatalFunc(args...)
	}
}

func (m *MockLogger) Fatalf(format string, args ...interface{}) {
	if m.FatalfFunc != nil {
		m.FatalfFunc(format, args...)
	}
}

func (m *MockLogger) With(key string, value interface{}) Logger {
	if m.WithFunc != nil {
		return m.WithFunc(key, value)
	}
	// Create a new logger instance to avoid type mismatch
	newLogger := &MockLogger{
		DebugFunc:       m.DebugFunc,
		DebugfFunc:      m.DebugfFunc,
		InfoFunc:        m.InfoFunc,
		InfofFunc:       m.InfofFunc,
		WarnFunc:        m.WarnFunc,
		WarnfFunc:       m.WarnfFunc,
		ErrorFunc:       m.ErrorFunc,
		ErrorfFunc:      m.ErrorfFunc,
		FatalFunc:       m.FatalFunc,
		FatalfFunc:      m.FatalfFunc,
		PanicFunc:       m.PanicFunc,
		PanicfFunc:      m.PanicfFunc,
		WithFunc:        m.WithFunc,
		WithMapFunc:     m.WithMapFunc,
		WithErrorFunc:   m.WithErrorFunc,
		WithContextFunc: m.WithContextFunc,
	}
	return newLogger
}

func (m *MockLogger) WithMap(fields map[string]interface{}) Logger {
	if m.WithMapFunc != nil {
		return m.WithMapFunc(fields)
	}
	// Create a new logger instance to avoid type mismatch
	newLogger := &MockLogger{
		DebugFunc:       m.DebugFunc,
		DebugfFunc:      m.DebugfFunc,
		InfoFunc:        m.InfoFunc,
		InfofFunc:       m.InfofFunc,
		WarnFunc:        m.WarnFunc,
		WarnfFunc:       m.WarnfFunc,
		ErrorFunc:       m.ErrorFunc,
		ErrorfFunc:      m.ErrorfFunc,
		FatalFunc:       m.FatalFunc,
		FatalfFunc:      m.FatalfFunc,
		PanicFunc:       m.PanicFunc,
		PanicfFunc:      m.PanicfFunc,
		WithFunc:        m.WithFunc,
		WithMapFunc:     m.WithMapFunc,
		WithErrorFunc:   m.WithErrorFunc,
		WithContextFunc: m.WithContextFunc,
	}
	return newLogger
}

func (m *MockLogger) Panic(args ...interface{}) {
	if m.PanicFunc != nil {
		m.PanicFunc(args...)
	}
}

func (m *MockLogger) Panicf(format string, args ...interface{}) {
	if m.PanicfFunc != nil {
		m.PanicfFunc(format, args...)
	}
}

func (m *MockLogger) WithError(err error) Logger {
	if m.WithErrorFunc != nil {
		return m.WithErrorFunc(err)
	}
	return m
}

func (m *MockLogger) WithContext(ctx interface{}) Logger {
	if m.WithContextFunc != nil {
		return m.WithContextFunc(ctx)
	}
	// Create a new logger instance to avoid type mismatch
	newLogger := &MockLogger{
		DebugFunc:       m.DebugFunc,
		DebugfFunc:      m.DebugfFunc,
		InfoFunc:        m.InfoFunc,
		InfofFunc:       m.InfofFunc,
		WarnFunc:        m.WarnFunc,
		WarnfFunc:       m.WarnfFunc,
		ErrorFunc:       m.ErrorFunc,
		ErrorfFunc:      m.ErrorfFunc,
		FatalFunc:       m.FatalFunc,
		FatalfFunc:      m.FatalfFunc,
		PanicFunc:       m.PanicFunc,
		PanicfFunc:      m.PanicfFunc,
		WithFunc:        m.WithFunc,
		WithMapFunc:     m.WithMapFunc,
		WithErrorFunc:   m.WithErrorFunc,
		WithContextFunc: m.WithContextFunc,
	}
	return newLogger
}
