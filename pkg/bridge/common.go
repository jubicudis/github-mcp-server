/*
 * WHO: BridgeCommon
 * WHAT: Shared utilitiesWriteTimeout           = 10 * time.Second
ReadTimeout            = 15 * time.Secondor bridge operations
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
	"math"
	"os/exec"
	"sync"
	"time"

	"github-mcp-server/pkg/log"
	"github-mcp-server/pkg/translations"

	"github.com/gorilla/websocket"
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
	DefaultBridgeURL       = "ws://localhost:10619/bridge"
	BridgeServiceName      = "github_mcp_bridge"
	MaxReconnectAttempts   = 10
	ReconnectDelay         = 5 * time.Second
	HealthCheckInterval    = 30 * time.Second
	MessageBufferSize      = 100
	WriteTimeout           = 10 * time.Second
	ReadTimeout            = 15 * time.Second
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
	conn    *websocket.Conn
	options ConnectionOptions
	state   ConnectionState

	// Synchronization
	mutex sync.RWMutex
	mu    sync.Mutex // Alternative mutex for client.go implementation

	// Message handling
	messageHandler    MessageHandler
	disconnectHandler DisconnectHandler

	// Statistics
	stats BridgeStats

	// Context management
	ctx        context.Context
	cancelFunc context.CancelFunc

	// State flags
	isClosed  bool
	connected bool

	// Logging
	logger *log.Logger

	// Channel for incoming messages
	messages chan Message

	// QHP session key for secure channel
	sessionKey string
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

// WHO: FallbackRouter
// WHAT: Shared fallback routing for MCP operations
// WHEN: On MCP operation failure or unreachability
// WHERE: System Layer 6 (Integration)
// WHY: To ensure robust multi-path routing between TNOS MCP, Bridge, GitHub MCP, and Copilot LLM
// HOW: Tries each route in order, logging 7D context at each step
// EXTENT: All MCP operations requiring fallback
func FallbackRoute(
	ctx context.Context,
	operationName string,
	context7D translations.ContextVector7D,
	tnosMCPFunc func() (interface{}, error),
	bridgeFunc func() (interface{}, error),
	githubMCPFunc func() (interface{}, error),
	copilotFunc func() (interface{}, error),
	logger *log.Logger,
) (interface{}, error) {
	// Try TNOS MCP first
	result, err := tnosMCPFunc()
	if err == nil {
		if logger != nil {
			logger.Info("FallbackRoute: TNOS MCP succeeded", "who", context7D.Who, "what", operationName)
		}
		return result, nil
	}
	if logger != nil {
		logger.Warn("FallbackRoute: TNOS MCP failed, trying Bridge", "error", err.Error(), "who", context7D.Who, "what", operationName)
	}
	// Try Bridge
	result, err = bridgeFunc()
	if err == nil {
		if logger != nil {
			logger.Info("FallbackRoute: Bridge succeeded", "who", context7D.Who, "what", operationName)
		}
		return result, nil
	}
	if logger != nil {
		logger.Warn("FallbackRoute: Bridge failed, trying GitHub MCP", "error", err.Error(), "who", context7D.Who, "what", operationName)
	}
	// Try GitHub MCP
	result, err = githubMCPFunc()
	if err == nil {
		if logger != nil {
			logger.Info("FallbackRoute: GitHub MCP succeeded", "who", context7D.Who, "what", operationName)
		}
		return result, nil
	}
	if logger != nil {
		logger.Warn("FallbackRoute: GitHub MCP failed, trying Copilot LLM", "error", err.Error(), "who", context7D.Who, "what", operationName)
	}
	// Try Copilot LLM
	result, err = copilotFunc()
	if err == nil {
		if logger != nil {
			logger.Info("FallbackRoute: Copilot LLM succeeded", "who", context7D.Who, "what", operationName)
		}
		return result, nil
	}
	if logger != nil {
		logger.Error("FallbackRoute: All routes failed", "error", err.Error(), "who", context7D.Who, "what", operationName)
	}
	return nil, errors.New("All fallback routes failed: " + err.Error())
}

// WHO: FormulaRegistryInterface
// WHAT: Interface for formula registry compression/decompression
// WHEN: During compression/decompression operations
// WHERE: System Layer 6 (Integration)
// WHY: To provide Möbius compression via formula registry
// HOW: Calls out to Python helper via subprocess
// EXTENT: All bridge compression/decompression

type FormulaRegistry interface {
	CompressValue(value, entropy, B, V, I, G, F, purpose, t, C_sum float64) (compressed, E float64, err error)
	DecompressValue(compressed, entropy, B, V, I, G, F, purpose, t, E, C_sum float64) (original float64, err error)
}

// formulaRegistryPython implements FormulaRegistry using the Python helper
// (Assumes formula_registry_helper.py is available in scripts/shell/helpers)
type formulaRegistryPython struct{}

func (f *formulaRegistryPython) CompressValue(value, entropy, B, V, I, G, F, purpose, t, C_sum float64) (float64, float64, error) {
	// Call the Python helper script
	args := []string{"scripts/shell/helpers/formula_registry_helper.py", "compress", fmt.Sprintf(`{"value":%f,"entropy":%f,"B":%f,"V":%f,"I":%f,"G":%f,"F":%f,"purpose":%f,"t":%f,"C_sum":%f}`,
		value, entropy, B, V, I, G, F, purpose, t, C_sum)}
	out, err := runPythonHelper(args)
	if err != nil {
		return 0, 0, err
	}
	var compressed, E float64
	n, _ := fmt.Sscanf(string(out), "%f %f", &compressed, &E)
	if n != 2 {
		return 0, 0, fmt.Errorf("unexpected output from formula_registry_helper.py: %s", out)
	}
	return compressed, E, nil
}

func (f *formulaRegistryPython) DecompressValue(compressed, entropy, B, V, I, G, F, purpose, t, E, C_sum float64) (float64, error) {
	args := []string{"scripts/shell/helpers/formula_registry_helper.py", "decompress", fmt.Sprintf(`{"compressed":%f,"entropy":%f,"B":%f,"V":%f,"I":%f,"G":%f,"F":%f,"purpose":%f,"t":%f,"E":%f,"C_sum":%f}`,
		compressed, entropy, B, V, I, G, F, purpose, t, E, C_sum)}
	out, err := runPythonHelper(args)
	if err != nil {
		return 0, err
	}
	var original float64
	n, _ := fmt.Sscanf(string(out), "%f", &original)
	if n != 1 {
		return 0, fmt.Errorf("unexpected output from formula_registry_helper.py: %s", out)
	}
	return original, nil
}

// runPythonHelper runs the Python helper script and returns output
func runPythonHelper(args []string) ([]byte, error) {
	// Use exec.Command to call python3.11
	// NOTE: This assumes python3.11 is in PATH and scripts are present
	cmd := exec.Command("python3.11", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("python3.11 %v failed: %v\nOutput: %s", args, err, out)
	}
	return out, nil
}

// GetFormulaRegistry returns a FormulaRegistry implementation
func GetFormulaRegistry() FormulaRegistry {
	return &formulaRegistryPython{}
}

// WHO: MobiusCompressor
// WHAT: Native Go implementation of Möbius compression formula (7D-aware)
// WHEN: During compression/collapse operations
// WHERE: System Layer 6 (Integration)
// WHY: To provide direct, context-aware compression for MCP
// HOW: Implements Möbius formula and helpers
// EXTENT: All bridge/client/server compression/collapse

// MobiusCompressionParams holds all 7D context variables for compression
// WHO: MobiusCompressionParams
// WHAT: 7D context for Möbius compression
// WHEN: On compression/collapse
// WHERE: System Layer 6 (Integration)
// WHY: To pass all context to compression logic
// HOW: Struct with all formula variables
// EXTENT: All compression/collapse ops

type MobiusCompressionParams struct {
	Value     float64 // Input value
	B         float64 // Biological/priority factor
	I         float64 // Identity/context factor
	V         float64 // Validity/trust factor
	G         float64 // Global context weight
	F         float64 // Feedback/context weight
	Entropy   float64 // Symbolic entropy
	E         float64 // Energy/time factor
	T         float64 // Time since activation
	Csum      float64 // Context sum/overlap
	Alignment float64 // Alignment factor
}

// MobiusCompress applies the Möbius compression formula
func MobiusCompress(params MobiusCompressionParams) (compressed float64, alignment float64) {
	// Calculate alignment
	alignment = (params.B + params.V*params.I) * expNeg(params.T*params.E)
	// Möbius compression formula
	compressed = (params.Value * params.B * params.I * (1 - (params.Entropy / log2(1+params.V))) * (params.G + params.F)) /
		(params.E*params.T + params.Csum*params.Entropy + alignment)
	return
}

// expNeg returns exp(-x)
func expNeg(x float64) float64 {
	return math.Exp(-x)
}

// log2 returns log base 2
func log2(x float64) float64 {
	return math.Log2(x)
}

// CollapseCondition checks if collapse should occur based on sum and entropy
func CollapseCondition(RICSum, thetaCollapse, H_eff, thetaClarity float64) bool {
	return RICSum >= thetaCollapse && H_eff <= thetaClarity
}

// WeightedEntropy computes effective entropy (H_eff)
func WeightedEntropy(RIC, H []float64) float64 {
	var sumRIC, sumWeighted float64
	for i := range RIC {
		sumRIC += RIC[i]
		sumWeighted += RIC[i] * H[i]
	}
	if sumRIC == 0 {
		return 0
	}
	return sumWeighted / sumRIC
}
