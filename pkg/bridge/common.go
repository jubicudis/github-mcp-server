/*
 * WHO: BridgeCommon
 * WHAT: Shared utilities for bridge operations
 * WHEN: Throughout all bridge operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To centralize shared functionality
 * HOW: By defining common constants, types, and functions
 * EXTENT: All bridge components
 */

package bridge

import (
	"encoding/json"
	"errors"
	"time"
)

// Protocol constants - centralized from bridge.go and client.go
const (
	// WHO: ProtocolManager
	// WHAT: Protocol versions
	// WHEN: During version negotiation
	// WHERE: System Layer 6 (Integration)
	// WHY: For protocol compatibility
	// HOW: Using version constants
	// EXTENT: All MCP operations

	// Protocol versions
	MCPVersion10 = "1.0"
	MCPVersion20 = "2.0"
	MCPVersion30 = "3.0"
	
	// Bridge service configuration
	DefaultBridgeURL       = "ws://localhost:9000/mcp/bridge"
	BridgeServiceName      = "github_mcp_bridge"
	MaxReconnectAttempts   = 10
	ReconnectDelay         = 5 * time.Second
	HealthCheckInterval    = 30 * time.Second
	MessageBufferSize      = 100
	WriteTimeout           = 30 * time.Second
	ReadTimeout            = 120 * time.Second
	PingInterval           = 30 * time.Second
	MaxMessageSize         = 10485760 // 10MB
	DefaultProtocolVersion = "3.0"
)

// Common error definitions
var (
	ErrInvalidMessageFormat = errors.New("invalid message format")
	ErrUnsupportedVersion   = errors.New("unsupported protocol version")
	ErrConnectionClosed     = errors.New("connection closed")
	ErrContextTimeout       = errors.New("context deadline exceeded")
	ErrBridgeNotConnected   = errors.New("bridge not connected")
)

// ConnectionState represents the current state of the bridge connection
type ConnectionState string

// Connection states
const (
	// WHO: StateManager
	// WHAT: Bridge connection states
	// WHEN: During connection lifecycle
	// WHERE: System Layer 6 (Integration)
	// WHY: For connection monitoring
	// HOW: Using state transitions
	// EXTENT: Connection lifecycle

	StateDisconnected ConnectionState = "DISCONNECTED"
	StateConnecting   ConnectionState = "CONNECTING"
	StateConnected    ConnectionState = "CONNECTED"
	StateReconnecting ConnectionState = "RECONNECTING"
	StateError        ConnectionState = "ERROR"
)

// Utility functions

// WHO: JSONProcessor
// WHAT: Deep clone for JSON serializable objects
// WHEN: During message processing
// WHERE: System Layer 6 (Integration)
// WHY: To create deep copies of messages
// HOW: Using JSON marshaling/unmarshaling
// EXTENT: All message operations
func DeepClone(src, dst interface{}) error {
	if src == nil {
		return nil
	}
	
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(data, dst)
}

// WHO: VersionNegotiator
// WHAT: Check if a version is supported
// WHEN: During connection setup
// WHERE: System Layer 6 (Integration)
// WHY: To validate protocol versions
// HOW: Using version comparison
// EXTENT: Connection establishment
func IsSupportedVersion(version string) bool {
	switch version {
	case MCPVersion10, MCPVersion20, MCPVersion30:
		return true
	default:
		return false
	}
}
