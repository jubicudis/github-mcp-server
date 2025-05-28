/*
 * WHO: TestCommons
 * WHAT: Common test utilities
 * WHEN: During test execution
 * WHERE: GitHub MCP Server Tests
 * WHY: To provide consistent test helpers
 * HOW: By centralizing test utilities
 * EXTENT: All test cases
 */

package testhelper

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// Test constants
const (
	TestTimeout       = 5 * time.Second
	TestServerAddress = "localhost:8080"
	TestBridgeAddress = "localhost:9000"
)

// WHO: TestHelper
// WHAT: JSON conversion utility
// WHEN: During test execution
// WHERE: All test suites
// WHY: To convert structs to JSON for testing
// HOW: Using Go's encoding/json
// EXTENT: All JSON-related tests
func ToJSON(t *testing.T, v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}
	return string(data)
}

// WHO: TestHelper
// WHAT: JSON parsing utility
// WHEN: During test execution
// WHERE: All test suites
// WHY: To parse JSON in tests
// HOW: Using Go's encoding/json
// EXTENT: All JSON-related tests
func FromJSON(t *testing.T, data string, v interface{}) {
	if err := json.Unmarshal([]byte(data), v); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
}

// WHO: TestHelper
// WHAT: Test context creation
// WHEN: During test setup
// WHERE: All test suites
// WHY: To create consistent test contexts
// HOW: Using context with timeout
// EXTENT: All context-using tests
func NewTestContext(t *testing.T) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), TestTimeout)
}

// WHO: WebSocketHelper
// WHAT: Test WebSocket creation
// WHEN: During WebSocket tests
// WHERE: WebSocket test suites
// WHY: To create test WebSocket connections
// HOW: Using gorilla/websocket
// EXTENT: All WebSocket tests
func NewTestWebSocketClient(t *testing.T, url string) *websocket.Conn {
	dialer := websocket.Dialer{
		HandshakeTimeout: TestTimeout,
	}

	conn, _, err := dialer.Dial(url, http.Header{})
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}

	return conn
}

// WHO: ErrorComparer
// WHAT: Error comparison utility
// WHEN: During error checking in tests
// WHERE: All test suites
// WHY: To compare errors consistently
// HOW: Using error string comparison
// EXTENT: All error-checking tests
func ErrorsEqual(err1, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true
	}
	if err1 == nil || err2 == nil {
		return false
	}
	return err1.Error() == err2.Error()
}
