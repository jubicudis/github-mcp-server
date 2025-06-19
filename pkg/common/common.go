// Package common provides shared utilities consolidated from various subpackages.
// Merged content from pkg/bridge/common.go, pkg/github/common.go,
// pkg/translations/common.go, and pkg/testhelper/common.go.
package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mark3labs/mcp-go/mcp"
)

// Use websocket.Conn to satisfy import and avoid unused import lint error
var _ = (*websocket.Conn)(nil)

// TranslationHelperFunc is a function type for simple string translations
// Moved from translations package to eliminate dependency
type TranslationHelperFunc func(key string, defaultValue string) string

// NullTranslationHelperFunc is a no-op implementation that returns the default value
var NullTranslationHelperFunc TranslationHelperFunc = func(key string, defaultValue string) string {
	return defaultValue
}

// ========== From bridge/common.go ==========
// Remove all bridge URLs, port assignments, and any logic not truly shared or canonical.
// Only keep minimal, truly shared constants and types.
const (
	MCPVersion10           = "1.0"
	MCPVersion20           = "2.0"
	MCPVersion30           = "3.0"
	MaxReconnectAttempts   = 10
	ReconnectDelay         = 5 * time.Second
	HealthCheckInterval    = 30 * time.Second
	MessageBufferSize      = 100
	WriteTimeout           = 10 * time.Second
	ReadTimeout            = 15 * time.Second
	PingInterval           = 30 * time.Second
	MaxMessageSize         = 10485760
	DefaultProtocolVersion = "3.0"
	AppVersion             = "1.2.0-tnos.alpha" // Updated version
)

type ConnectionState string

const (
	StateDisconnected ConnectionState = "DISCONNECTED"
	StateConnecting   ConnectionState = "CONNECTING"
	StateConnected    ConnectionState = "CONNECTED"
	StateReconnecting ConnectionState = "RECONNECTING"
	StateError        ConnectionState = "error"
)

// ConnectionOptions holds parameters for establishing a WebSocket connection.
type ConnectionOptions struct {
	ServerURL               string
	ServerPort              int         // Can be part of ServerURL, but kept for potential specific use
	Logger                  interface{} // Expecting log.LoggerInterface, but using interface{} for broader compatibility initially
	MaxRetries              int
	RetryDelay              time.Duration
	Timeout                 time.Duration
	QHPIntent               string // New: For X-QHP-Intent header (e.g., "TNOS MCP", "Copilot LLM", "GitHub MCP")
	SourceComponent         string // New: For X-QHP-Source header (e.g., "GitHubMCPServer-TNOSBridge")
	CustomHeaders           map[string]string
	CompressionEnabled      bool
	CompressionThreshold    int
	CompressionAlgorithm    string
	CompressionLevel        int
	CompressionPreserveKeys []string
	FilterCallback          func(cell interface{}) bool // Using interface{} for cell type initially
	// Additional fields required by client.go
	Context     map[string]interface{} // 7D context data
	Credentials map[string]string      // Authentication credentials
	Headers     map[string]string      // Additional HTTP headers (alias for CustomHeaders)
	// Add other relevant options like TLS config, proxy settings, etc.
}

type Message struct {
	Type      string                 `json:"type"`
	Timestamp int64                  `json:"timestamp"`
	ID        string                 `json:"id,omitempty"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Content   interface{}            `json:"content,omitempty"`
	Context   interface{}            `json:"context,omitempty"`
}

type MessageHandler func(message Message) error
type DisconnectHandler func(reason string)

type BridgeStats struct {
	MessagesSent     int64
	MessagesReceived int64
	ErrorCount       int64
	ReconnectCount   int64
	LastActive       time.Time
	Uptime           time.Duration
	StartTime        time.Time
}

var (
	ErrInvalidMessageFormat = errors.New("invalid message format")
	ErrUnsupportedVersion   = errors.New("unsupported protocol version")
	ErrConnectionClosed     = errors.New("connection closed")
	ErrContextTimeout       = errors.New("context deadline exceeded")
	ErrBridgeNotConnected   = errors.New("bridge not connected")
)

// ========== From github/common.go ==========

//nolint:unused
type PaginationParams struct{ page, perPage int }

//nolint:unused
type prParams struct {
	owner, repo string
	number      int
}

// Use PaginationParams and prParams to avoid unused lint errors
var _ = PaginationParams{}
var _ = prParams{}

// ===== REMOVED: Duplicate parameter extraction helpers below this point =====

// ========== From translations/common.go ==========
var (
	ErrInvalidContext = errors.New("invalid context")
	ErrInvalidMessage = errors.New("invalid message format")
	ErrEmptyMessage   = errors.New("empty message")
	ErrUnsupported    = errors.New("unsupported translation")
)

const (
	ContextKeyWho    = "who"
	ContextKeyWhat   = "what"
	ContextKeyWhen   = "when"
	ContextKeyWhere  = "where"
	ContextKeyWhy    = "why"
	ContextKeyHow    = "how"
	ContextKeyExtent = "extent"
	DefaultWho       = "System"
	DefaultWhat      = "Communication"
	DefaultWhere     = "Integration"
	DefaultWhy       = "Protocol"
	DefaultHow       = "MCP"
	DefaultExtent    = "1.0"
)

func ToJSON(v interface{}) (string, error) {
	if v == nil {
		return "{}", nil
	}
	data, err := json.Marshal(v)
	return string(data), err
}

func DecodeJSON(data string, v interface{}) error {
	if data == "" {
		return ErrEmptyMessage
	}
	return json.Unmarshal([]byte(data), v)
}

func Extract7DContext(ctx map[string]interface{}) map[string]string {
	result := make(map[string]string)
	// extraction logic ...
	return result
}

// ========== From testhelper/common.go ==========
const (
	TestTimeout       = 5 * time.Second
	TestServerAddress = "localhost:8080"
	TestBridgeAddress = "localhost:9000"
)

func ToJSONTest(v interface{}) (string, error)      { b, err := json.Marshal(v); return string(b), err }
func FromJSONTest(data string, v interface{}) error { return json.Unmarshal([]byte(data), v) }
func NewTestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), TestTimeout)
}
func ErrorsEqual(err1, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true
	}
	if err1 == nil || err2 == nil {
		return false
	}
	return err1.Error() == err2.Error()
}

// ========== Generic Map Helper Functions ==========

// GetString retrieves a string value from a map with a default fallback
func GetString(m map[string]interface{}, key, defaultValue string) string {
	if m == nil {
		return defaultValue
	}

	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetInt64 retrieves an int64 value from a map with a default fallback
func GetInt64(m map[string]interface{}, key string, defaultValue int64) int64 {
	if m == nil {
		return defaultValue
	}

	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case float64:
			return int64(v)
		}
	}
	return defaultValue
}

// GetFloat64 retrieves a float64 value from a map with a default fallback
func GetFloat64(m map[string]interface{}, key string, defaultValue float64) float64 {
	if m == nil {
		return defaultValue
	}

	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return defaultValue
}

// Ptr is a convenience function for creating a pointer to a value
func Ptr[T any](v T) *T {
	return &v
}

// ===== MCP CallToolRequest Parameter Extraction Helpers =====
// These helpers are the only supported parameter extraction utilities for MCP tools.
// All parameter extraction should use these, operating on request.Params.Arguments.

func OptionalParamOK[T any](r mcp.CallToolRequest, p string) (value T, ok bool, err error) {
	val, exists := r.Params.Arguments[p]
	if !exists {
		return
	}
	value, ok = val.(T)
	if !ok {
		err = fmt.Errorf("parameter %s is not of type %T, is %T", p, value, val)
		ok = true
	}
	return
}

func RequiredParam[T comparable](r mcp.CallToolRequest, p string) (T, error) {
	var zero T
	v, ok := r.Params.Arguments[p]
	if !ok {
		return zero, fmt.Errorf("missing required parameter: %s", p)
	}
	value, ok2 := v.(T)
	if !ok2 || value == zero {
		return zero, fmt.Errorf("invalid parameter: %s", p)
	}
	return value, nil
}

func RequiredIntParam(r mcp.CallToolRequest, p string) (int, error) {
	v, err := RequiredParam[float64](r, p)
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

func OptionalIntParam(r mcp.CallToolRequest, p string) (int, error) {
	v, _, err := OptionalParamOK[float64](r, p)
	return int(v), err
}

func OptionalIntParamWithDefault(r mcp.CallToolRequest, p string, d int) (int, error) {
	v, err := OptionalIntParam(r, p)
	if err != nil {
		return 0, err
	}
	if v == 0 {
		return d, nil
	}
	return v, nil
}

func OptionalStringArrayParam(request mcp.CallToolRequest, key string) ([]string, error) {
	val, ok := request.Params.Arguments[key]
	if !ok {
		return []string{}, nil
	}
	switch arr := val.(type) {
	case []string:
		return arr, nil
	case []interface{}:
		result := make([]string, len(arr))
		for i, v := range arr {
			s, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("parameter '%s' contains non-string element", key)
			}
			result[i] = s
		}
		return result, nil
	default:
		return nil, fmt.Errorf("parameter '%s' is not a string array", key)
	}
}

func WithPagination(request mcp.CallToolRequest) (page, perPage int, err error) {
	page, err = OptionalIntParamWithDefault(request, "page", 1)
	if err != nil {
		return 0, 0, err
	}
	perPage, err = OptionalIntParamWithDefault(request, "perPage", 30)
	if err != nil {
		return 0, 0, err
	}
	return page, perPage, nil
}

// ===== End MCP Parameter Helpers =====

// CreateErrorResponse returns a formatted MCP error response for tool handlers.
// translateFn: a translation function (may be nil), key: translation key, format: error message format, args: format args
func CreateErrorResponse(translateFn interface{}, key, format string, args ...interface{}) (*mcp.CallToolResult, error) {
	var msg string
	if translateFn != nil {
		switch fn := translateFn.(type) {
		case func(string, string) string:
			msg = fn(key, fmt.Sprintf(format, args...))
		case func(context.Context, map[string]interface{}) (map[string]interface{}, error):
			// Not used for error string, fallback below
			msg = fmt.Sprintf(format, args...)
		default:
			msg = fmt.Sprintf(format, args...)
		}
	} else {
		msg = fmt.Sprintf(format, args...)
	}
	return mcp.NewToolResultError(msg), nil
}

//nolint:unused
// Client represents a general MCP protocol client over WebSocket
// WHO: Common Client
// WHAT: WebSocket client wrapper
// WHEN: During bridge operations
// WHERE: System Layer 6 (Integration)
// WHY: To encapsulate connection logic
// HOW: Using WebSocket and common ConnectionOptions
// EXTENT: All bridge client usage

type Client struct {
	Conn              *websocket.Conn
	Options           ConnectionOptions
	State             ConnectionState
	MessageHandler    MessageHandler
	DisconnectHandler DisconnectHandler
	Stats             BridgeStats
	Ctx               context.Context
	CancelFunc        context.CancelFunc
}
