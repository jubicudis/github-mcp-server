// WHO: SharedMCPComponents
// WHAT: Common MCP protocol definitions and structures
// WHEN: During all bridge operations
// WHERE: System Layer 6 (Integration)
// WHY: To provide consistent shared components
// HOW: Using centralized definitions
// EXTENT: All MCP bridge interactions

package bridge

import (
	"time"

	"tranquility-neuro-os/github-mcp-server/pkg/log"
	"tranquility-neuro-os/github-mcp-server/pkg/translations"
)

// DefaultProtocolVersion and other protocol constants are defined in common.go

// ClientMessage represents a bridge message for client operations
type ClientMessage struct {
	// WHO: MessageFormat
	// WHAT: Message structure definition
	// WHEN: During message exchange
	// WHERE: System Layer 6 (Integration)
	// WHY: To structure bridge communications
	// HOW: Using standardized message format
	// EXTENT: All bridge messages

	Type      string      `json:"type"`
	Timestamp int64       `json:"timestamp"`
	Content   interface{} `json:"content,omitempty"`
	Context   interface{} `json:"context,omitempty"`
}

// ClientBridgeStats tracks operational statistics for the client side
type ClientBridgeStats struct {
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

// ClientConnectionOptions contains options for connecting to the bridge
type ClientConnectionOptions struct {
	// WHO: ConnectionConfig
	// WHAT: Connection configuration
	// WHEN: During connection setup
	// WHERE: System Layer 6 (Integration)
	// WHY: To customize connection behavior
	// HOW: Using structured options
	// EXTENT: Connection lifecycle

	ServerURL  string
	ServerPort int
	Context    translations.ContextVector7D
	Logger     *log.Logger
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
	Headers    map[string]string
}
