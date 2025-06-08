// Package common provides shared utilities consolidated from various subpackages.
// Merged content from pkg/bridge/common.go, pkg/github/common.go,
// pkg/translations/common.go, and pkg/testhelper/common.go.
package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/go-github/v71/github"
	"github.com/gorilla/websocket"
	"github.com/mark3labs/mcp-go/mcp"
)

// ========== From bridge/common.go ==========
const (
	MCPVersion10          = "1.0"
	MCPVersion20          = "2.0"
	MCPVersion30          = "3.0"
	DefaultBridgeURL      = "ws://localhost:10619/bridge"
	BridgeServiceName     = "github_mcp_bridge"
	MaxReconnectAttempts  = 10
	ReconnectDelay        = 5 * time.Second
	HealthCheckInterval   = 30 * time.Second
	MessageBufferSize     = 100
	WriteTimeout          = 10 * time.Second
	ReadTimeout           = 15 * time.Second
	PingInterval          = 30 * time.Second
	MaxMessageSize        = 10485760
	DefaultProtocolVersion= "3.0"
)

type ConnectionState string

const (
	StateDisconnected ConnectionState = "DISCONNECTED"
	StateConnecting   ConnectionState = "CONNECTING"
	StateConnected    ConnectionState = "CONNECTED"
	StateReconnecting ConnectionState = "RECONNECTING"
	StateError        ConnectionState = "ERROR"
)

type ConnectionOptions struct {
	ServerURL   string
	ServerPort  int
	// PythonMCPURL specifies the fallback stub endpoint for Python MCP
	PythonMCPURL string
	Context     map[string]interface{}
	Logger      interface{}
	Timeout     time.Duration
	MaxRetries  int
	RetryDelay  time.Duration
	TLSEnabled  bool
	Credentials map[string]string
	Headers     map[string]string
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

type Client struct {
	conn    *websocket.Conn
	options ConnectionOptions
	state   ConnectionState
	mutex   sync.RWMutex
	mu      sync.Mutex
	messageHandler    MessageHandler
	disconnectHandler DisconnectHandler
	stats BridgeStats
	ctx    context.Context
	cancelFunc context.CancelFunc
}

var (
	ErrInvalidMessageFormat = errors.New("invalid message format")
	ErrUnsupportedVersion   = errors.New("unsupported protocol version")
	ErrConnectionClosed     = errors.New("connection closed")
	ErrContextTimeout       = errors.New("context deadline exceeded")
	ErrBridgeNotConnected   = errors.New("bridge not connected")
)

// ========== From github/common.go ==========

type StringTranslationFunc func(key, defaultValue string) string

type ContextTranslationFunc func(ctx context.Context, contextData map[string]interface{}) (map[string]interface{}, error)

type PaginationParams struct { page, perPage int }

type prParams struct { owner, repo string; number int }

func OptionalParamOK[T any](r mcp.CallToolRequest, p string) (value T, ok bool, err error) {
    val, exists := r.Params.Arguments[p]
    if !exists { return }
    value, ok = val.(T)
    if !ok { err = fmt.Errorf("parameter %s is not of type %T, is %T", p, value, val); ok = true }
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

func RequiredIntParam(r mcp.CallToolRequest, p string) (int, error) { v, err := RequiredParam[float64](r, p); if err != nil { return 0, err }; return int(v), nil }
func OptionalIntParam(r mcp.CallToolRequest, p string) (int, error) { v, _, err := OptionalParamOK[float64](r, p); return int(v), err }
func OptionalIntParamWithDefault(r mcp.CallToolRequest, p string, d int) (int, error) { v, err := OptionalIntParam(r, p); if err != nil { return 0, err }; if v == 0 { return d, nil }; return v, nil }

// isAcceptedError checks if an error is an AcceptedError
func IsAcceptedError(err error) bool {
    var acceptedError *github.AcceptedError
    return errors.As(err, &acceptedError)
}

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
	if v == nil { return "{}", nil }
	data, err := json.Marshal(v); return string(data), err
}

func DecodeJSON(data string, v interface{}) error {
	if data == "" { return ErrEmptyMessage }
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

func ToJSONTest(v interface{}) (string, error) { b, err := json.Marshal(v); return string(b), err }
func FromJSONTest(data string, v interface{}) error { return json.Unmarshal([]byte(data), v) }
func NewTestContext() (context.Context, context.CancelFunc) { return context.WithTimeout(context.Background(), TestTimeout) }
func ErrorsEqual(err1, err2 error) bool { if err1 == nil && err2 == nil { return true }; if err1 == nil || err2 == nil { return false }; return err1.Error() == err2.Error() }

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

// RequiredInt extracts a required integer parameter from the request.
func RequiredInt(request *http.Request, paramName string) (int, error) {
	// Implementation to extract and validate integer parameter (omitted for brevity)
	return 0, errors.New("not implemented")
}

// OptionalPaginationParams extracts optional pagination parameters from the request.
func OptionalPaginationParams(request *http.Request) (map[string]int, error) {
	// Implementation to extract pagination parameters (omitted for brevity)
	return nil, errors.New("not implemented")
}

// WithPagination is a middleware function to handle pagination logic.
func WithPagination() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Placeholder logic for pagination middleware
			// Extract pagination parameters using OptionalPaginationParams
			pagination, err := OptionalPaginationParams(r)
			if err != nil {
				http.Error(w, "Invalid pagination parameters", http.StatusBadRequest)
				return
			}

			// Add pagination info to the request context
			ctx := context.WithValue(r.Context(), "pagination", pagination)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserEndpoint is the base endpoint for user-related API calls.
const UserEndpoint = "/api/v1/users"

// OptionalStringArrayParam extracts an optional string array parameter from the request.
func OptionalStringArrayParam(request mcp.CallToolRequest, key string) ([]string, error) {
	if value, ok := request.Params[key]; ok {
		if array, ok := value.([]string); ok {
			return array, nil
		}
		return nil, fmt.Errorf("parameter '%s' is not a string array", key)
	}
	return nil, nil
}

// OptionalIntParam extracts an optional integer parameter from the request.
func OptionalIntParam(request mcp.CallToolRequest, key string) (*int, error) {
	if value, ok := request.Params[key]; ok {
		if number, ok := value.(int); ok {
			return &number, nil
		}
		return nil, fmt.Errorf("parameter '%s' is not an integer", key)
	}
	return nil, nil
}

// OptionalIntParamWithDefault extracts an optional integer parameter with a default value.
func OptionalIntParamWithDefault(request mcp.CallToolRequest, key string, defaultValue int) (int, error) {
	if value, ok := request.Params[key]; ok {
		if number, ok := value.(int); ok {
			return number, nil
		}
		return 0, fmt.Errorf("parameter '%s' is not an integer", key)
	}
	return defaultValue, nil
}

// WithPagination adds pagination parameters to a request.
func WithPagination(request mcp.CallToolRequest) (page, perPage int, err error) {
	page, err = OptionalIntParamWithDefault(request, "page", 1)
	if err != nil {
		return 0, 0, err
	}
	perPage, err = OptionalIntParamWithDefault(request, "per_page", 30)
	if err != nil {
		return 0, 0, err
	}
	return page, perPage, nil
}
