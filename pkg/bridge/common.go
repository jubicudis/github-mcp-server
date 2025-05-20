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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jubicudis/github-mcp-server/pkg/log"
	"github.com/jubicudis/github-mcp-server/pkg/translations"
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

// Message types - centralized from mcp_bridge.go
const (
	// WHO: MessageTypeManager
	// WHAT: Message type constants
	// WHEN: During message handling
	// WHERE: System Layer 6 (Integration)
	// WHY: To categorize messages
	// HOW: Using type codes
	// EXTENT: All message types

	TypeHandshake    = "handshake"
	TypeCommand      = "command"
	TypeQuery        = "query"
	TypeResponse     = "response"
	TypeNotification = "notification"
	TypeError        = "error"
	TypeHealthCheck  = "health_check"
	TypeContext      = "context"
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

// ConnectionOptions contains options for connecting to the bridge
type ConnectionOptions struct {
	// WHO: ConnectionConfig
	// WHAT: Connection configuration
	// WHEN: During connection setup
	// WHERE: System Layer 6 (Integration)
	// WHY: To customize connection behavior
	// HOW: Using structured options
	// EXTENT: Connection lifecycle

	ServerURL   string
	ServerPort  int
	Context     translations.ContextVector7D
	Logger      *log.Logger
	Timeout     time.Duration
	MaxRetries  int
	RetryDelay  time.Duration
	TLSEnabled  bool
	Credentials map[string]string
	Headers     map[string]string
}

// Message represents a message sent over the bridge
type Message struct {
	// WHO: MessageFormat
	// WHAT: Message structure definition
	// WHEN: During message exchange
	// WHERE: System Layer 6 (Integration)
	// WHY: To structure bridge communications
	// HOW: Using standardized message format
	// EXTENT: All bridge messages

	Type      string                 `json:"type"`
	Timestamp int64                  `json:"timestamp"`
	ID        string                 `json:"id,omitempty"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Content   interface{}            `json:"content,omitempty"`
	Context   interface{}            `json:"context,omitempty"`
}

// MessageHandler is a function type for handling messages
type MessageHandler func(message Message) error

// DisconnectHandler is a function type for handling disconnection events
type DisconnectHandler func(reason string)

// BridgeStats tracks operational statistics for the bridge
type BridgeStats struct {
	// WHO: StatsCollector
	// WHAT: Bridge operational statistics
	// WHEN: During bridge operations
	// WHERE: System Layer 6 (Integration)
	// WHY: For performance monitoring
	// HOW: Using counters
	// EXTENT: System health

	MessagesSent     int64
	MessagesReceived int64
	ErrorCount       int64
	ReconnectCount   int64
	LastActive       time.Time
	Uptime           time.Duration
	StartTime        time.Time
}

// Client represents a connection to the MCP bridge
type Client struct {
	// WHO: BridgeClient
	// WHAT: Client for MCP bridge communication
	// WHEN: During bridge operations
	// WHERE: System Layer 6 (Integration)
	// WHY: For communication with TNOS MCP
	// HOW: Using WebSocket protocol
	// EXTENT: All bridge operations

	// Core connection fields
	conn      *websocket.Conn
	options   ConnectionOptions
	state     ConnectionState
	
	// Synchronization
	mutex     sync.RWMutex
	mu        sync.Mutex // Alternative mutex for client.go implementation
	
	// Message handling
	messageHandler    MessageHandler
	disconnectHandler DisconnectHandler
	
	// Statistics
	stats     BridgeStats
	
	// Context management
	ctx       context.Context
	cancelFunc context.CancelFunc
	
	// State flags
	isClosed  bool
	connected bool
	
	// Logging
	logger     *log.Logger
	
	// Channel for incoming messages
	messages  chan Message
}

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

// WHO: MessageHelper
// WHAT: Create a standard message
// WHEN: During message preparation
// WHERE: System Layer 6 (Integration)
// WHY: To standardize message creation
// HOW: Using common format
// EXTENT: All message operations
func CreateMessage(messageType string, payload map[string]interface{}, context interface{}) Message {
	return Message{
		Type:      messageType,
		Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
		Payload:   payload,
		Context:   context,
		ID:        GenerateMessageID(),
	}
}

// WHO: IDGenerator
// WHAT: Generate unique message ID
// WHEN: During message creation
// WHERE: System Layer 6 (Integration)
// WHY: For message tracking
// HOW: Using timestamp and random component
// EXTENT: Message lifecycle
func GenerateMessageID() string {
	timestamp := time.Now().UnixNano()
	randomPart := timestamp % 10000
	return fmt.Sprintf("msg-%d-%d", timestamp, randomPart)
}

// WHO: ErrorHandler 
// WHAT: Create standard error response
// WHEN: During error handling
// WHERE: System Layer 6 (Integration)
// WHY: For standardized error reporting
// HOW: Using error message template
// EXTENT: Error handling
func CreateErrorResponse(originalMessage Message, errMsg string, errCode string) Message {
	return Message{
		Type:      TypeError,
		Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
		Context:   originalMessage.Context,
		ID:        GenerateMessageID(),
		Payload: map[string]interface{}{
			"originalMessage": originalMessage.ID,
			"error":           errMsg,
			"code":            errCode,
		},
	}
}

// [2025-05-20] Moved from client.go: context extraction helpers for shared use across bridge components.
func getString(m map[string]interface{}, key, defaultValue string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultValue
}

func getInt64(m map[string]interface{}, key string, defaultValue int64) int64 {
	if v, ok := m[key]; ok {
		switch t := v.(type) {
		case int64:
			return t
		case float64:
			return int64(t)
		case string:
			var parsedInt int64
			if _, err := fmt.Sscanf(t, "%d", &parsedInt); err == nil {
				return parsedInt
			}
		case int:
			return int64(t)
		}
	}
	return defaultValue
}

func getFloat64(m map[string]interface{}, key string, defaultValue float64) float64 {
	if v, ok := m[key]; ok {
		switch t := v.(type) {
		case float64:
			return t
		case int64:
			return float64(t)
		case string:
			var parsedFloat float64
			if _, err := fmt.Sscanf(t, "%f", &parsedFloat); err == nil {
				return parsedFloat
			}
		case int:
			return float64(t)
		}
	}
	return defaultValue
}
