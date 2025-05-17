/*
 * WHO: MCPBridgeClient
 * WHAT: Client for connecting to TNOS MCP Bridge
 * WHEN: During bridge connection operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide client-side bridge connectivity
 * HOW: Using WebSocket with structured message passing
 * EXTENT: All client-side bridge operations
 */

package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"tranquility-neuro-os/github-mcp-server/pkg/log"
	"tranquility-neuro-os/github-mcp-server/pkg/translations"

	"github.com/gorilla/websocket"
)

// ConnectionState represents the current state of the bridge connection
type ConnectionState string

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

// Protocol version constants
const (
	MCPVersion10 = "1.0"
	MCPVersion20 = "2.0"
	MCPVersion30 = "3.0"

	// Default protocol version
	DefaultProtocolVersion = MCPVersion30
)

// BridgeStats tracks operational statistics
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

// Message represents a bridge message
type Message struct {
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

// ConnectionOptions contains options for connecting to the bridge
type ConnectionOptions struct {
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

// Client represents a connection to the MCP bridge
type Client struct {
	// WHO: BridgeClient
	// WHAT: Client for bridge connection
	// WHEN: During client operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To manage bridge connectivity
	// HOW: Using WebSocket connection
	// EXTENT: Client lifecycle

	conn              *websocket.Conn
	options           ConnectionOptions
	logger            *log.Logger
	sendMutex         sync.Mutex
	receiveMutex      sync.Mutex
	mutex             sync.RWMutex
	context           translations.ContextVector7D
	state             ConnectionState
	messageHandler    func(Message) error
	disconnectHandler func(string)
	isClosed          bool
	cancelFunc        context.CancelFunc
	stats             BridgeStats
}

// NewClient creates a new bridge client and connects to the bridge
func NewClient(ctx context.Context, options ConnectionOptions) (*Client, error) {
	// WHO: ClientInitializer
	// WHAT: Initialize bridge client
	// WHEN: During client creation
	// WHERE: System Layer 6 (Integration)
	// WHY: To establish bridge connection
	// HOW: Using WebSocket protocol
	// EXTENT: Client initialization

	// Default logger if none provided
	logger := options.Logger
	if logger == nil {
		logger = log.NewLogger()
	}

	// Parse the URL
	u, err := url.Parse(options.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bridge URL: %v", err)
	}

	// Connect to WebSocket with context timeout
	logger.Info("Connecting to bridge", "url", options.ServerURL)

	var c *websocket.Conn
	var resp *http.Response

	dialer := websocket.Dialer{
		HandshakeTimeout: options.Timeout,
	}

	c, resp, err = dialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("failed to connect to bridge (status: %d): %v", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("failed to connect to bridge: %v", err)
	}

	// Create the client
	client := &Client{
		conn:    c,
		options: options,
		logger:  logger,
		context: options.Context,
	}

	// Send initial handshake
	handshakeMsg := Message{
		Type:      "handshake",
		Timestamp: time.Now().Unix(),
		Content: map[string]interface{}{
			"version": MCPVersion30, // Use latest version
			"client":  "github_mcp_server",
		},
		Context: options.Context.ToMap(),
	}

	if err := client.Send(handshakeMsg); err != nil {
		c.Close()
		return nil, fmt.Errorf("failed to send handshake: %v", err)
	}

	// Wait for handshake response
	var handshakeResponse Message
	handshakeResponse, err = client.Receive()
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("failed to receive handshake response: %v", err)
	}

	if handshakeResponse.Type != "handshake_response" {
		c.Close()
		return nil, fmt.Errorf("unexpected response to handshake: %s", handshakeResponse.Type)
	}

	logger.Info("Connected to bridge", "type", handshakeResponse.Type)
	return client, nil
}

// Send sends a message to the bridge
func (c *Client) Send(message Message) error {
	// WHO: MessageSender
	// WHAT: Send message to bridge
	// WHEN: During message transmission
	// WHERE: System Layer 6 (Integration)
	// WHY: To communicate with bridge
	// HOW: Using WebSocket protocol
	// EXTENT: Message lifecycle

	c.sendMutex.Lock()
	defer c.sendMutex.Unlock()

	// Set timestamp if not already set
	if message.Timestamp == 0 {
		message.Timestamp = time.Now().Unix()
	}

	// Add context if not provided
	if message.Context == nil {
		message.Context = c.context.ToMap()
	}

	// Marshal the message
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// Send the message
	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	c.logger.Debug("Sent message to bridge", "type", message.Type)
	return nil
}

// Receive receives a message from the bridge
func (c *Client) Receive() (Message, error) {
	// WHO: MessageReceiver
	// WHAT: Receive message from bridge
	// WHEN: During message reception
	// WHERE: System Layer 6 (Integration)
	// WHY: To receive bridge communications
	// HOW: Using WebSocket protocol
	// EXTENT: Message lifecycle

	c.receiveMutex.Lock()
	defer c.receiveMutex.Unlock()

	// Read message
	_, data, err := c.conn.ReadMessage()
	if err != nil {
		return Message{}, fmt.Errorf("failed to read message: %v", err)
	}

	// Parse message
	var message Message
	if err := json.Unmarshal(data, &message); err != nil {
		return Message{}, fmt.Errorf("failed to parse message: %v", err)
	}

	c.logger.Debug("Received message from bridge", "type", message.Type)
	return message, nil
}

// Close closes the connection to the bridge
func (c *Client) Close() error {
	// WHO: ConnectionManager
	// WHAT: Close bridge connection
	// WHEN: During connection termination
	// WHERE: System Layer 6 (Integration)
	// WHY: To release resources
	// HOW: Using WebSocket close protocol
	// EXTENT: Connection lifecycle

	if c.conn != nil {
		c.logger.Info("Closing bridge connection")
		return c.conn.Close()
	}
	return nil
}

// IsConnected returns true if the client is connected to the bridge
func (c *Client) IsConnected() bool {
	// WHO: ConnectionMonitor
	// WHAT: Check connection status
	// WHEN: During connection monitoring
	// WHERE: System Layer 6 (Integration)
	// WHY: To determine connectivity
	// HOW: Using connection state check
	// EXTENT: Connection lifecycle

	return c.conn != nil
}

// GetContext returns the client's context
func (c *Client) GetContext() translations.ContextVector7D {
	// WHO: ContextProvider
	// WHAT: Provide client context
	// WHEN: During context retrieval
	// WHERE: System Layer 6 (Integration)
	// WHY: To access context information
	// HOW: Using context accessor
	// EXTENT: Context lifecycle

	return c.context
}

// SetContext sets the client's context
func (c *Client) SetContext(context translations.ContextVector7D) {
	// WHO: ContextManager
	// WHAT: Update client context
	// WHEN: During context update
	// WHERE: System Layer 6 (Integration)
	// WHY: To modify context information
	// HOW: Using context mutator
	// EXTENT: Context lifecycle

	c.context = context
}
