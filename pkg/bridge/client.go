// WHO: SharedMCPComponents
// WHAT: Common MCP protocol definitions and structures
// WHEN: During all bridge operations
// WHERE: System Layer 6 (Integration)
// WHY: To provide consistent shared components
// HOW: Using centralized definitions
// EXTENT: All MCP bridge interactions

// TRANQUILSPEAK SYMBOL CLUSTER: [𝒯🧠βℏƒ𓆑#NE1𝑾𝑾𝑯𝑾𝑯𝑬𝑾𝑯𝑬𝑹𝑾𝑯𝒀𝑯𝑶𝑾𝑬𝑿⏳📍𝒮𝓔𝓗]
// This file is part of the 'nervous' biosystem. See circulatory/github-mcp-server/symbolic_mapping_registry_autogen_20250603.tsq for details.

package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/common"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/tranquilspeak"

	"github.com/gorilla/websocket"
)

// DefaultProtocolVersion and other protocol constants are defined in common.go

// Bridge is our own client implementation that wraps the common.Client
type Bridge struct {
	conn       *websocket.Conn
	options    common.ConnectionOptions
	state      common.ConnectionState
	mutex      sync.RWMutex
	mu         sync.Mutex
	stats      common.BridgeStats
	ctx        context.Context
	cancelFunc context.CancelFunc
	logger     *log.Logger
	messages   chan common.Message
	connected  bool        // Flag to track connection status
	sessionKey string      // Session key from handshake
	triggerMatrix *tranquilspeak.TriggerMatrix // Add triggerMatrix field
}

// NewClient creates a new bridge client with the given options and triggerMatrix
func NewClient(ctx context.Context, options common.ConnectionOptions, triggerMatrix *tranquilspeak.TriggerMatrix) (*Bridge, error) {
	// WHO: ClientFactory
	// WHAT: Create bridge client
	// WHEN: During client initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To establish bridge connection
	// HOW: Using WebSocket protocol
	// EXTENT: Client creation

	clientCtx, cancel := context.WithCancel(ctx)

	// Use a basic logger if log package is missing
	var logger *log.Logger
	if loggerInterface, ok := options.Logger.(*log.Logger); ok {
		logger = loggerInterface
	} else {
		logger = log.NewLogger(triggerMatrix) // fallback to default logger (canonical)
	}

	// Ensure context is initialized
	if options.Context == nil {
		// Create a default context
		options.Context = map[string]interface{}{
			"who":    "BridgeClient",
			"what":   "Connection",
			"when":   time.Now().Unix(),
			"where":  "SystemLayer6",
			"why":    "Communication",
			"how":    "WebSocket",
			"extent": 1.0,
			"source": "GitHubMCPServer",
		}
	}

	client := &Bridge{
		ctx:        clientCtx,
		cancelFunc: cancel,
		options:    options,
		logger:     logger,
		messages:   make(chan common.Message, common.MessageBufferSize),
		stats:      common.BridgeStats{},
		state:      common.StateDisconnected,
		triggerMatrix: triggerMatrix, // Store triggerMatrix in the client
	}

	err := client.connect()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to bridge: %w", err)
	}

	go client.readPump()

	return client, nil
}

// connect establishes a WebSocket connection to the bridge
func (c *Bridge) connect() error {
	// WHO: ConnectionManager
	// WHAT: Establish WebSocket connection with fallback
	// WHEN: During client (re)connection
	// WHERE: System Layer 6 (Integration)
	// WHY: To ensure robust connection
	// HOW: Using FallbackRoute utility
	// EXTENT: All connection attempts

	// Canonical fallback URLs/ports from common.go
	const (
		TnosMCPURL = "ws://localhost:9001" // TNOS MCP server (primary)
	)
	const tnosMCPPort = 9001

	c.options.ServerPort = tnosMCPPort

	// Emit ATM event trigger for port assignment (after assignment)
	if c.triggerMatrix != nil {
		trigger := c.triggerMatrix.CreateTrigger(
			"BridgeClient", // who
			"PortAssignment", // what
			"SystemLayer6", // where
			"CanonicalPortAssignment", // why
			"ATMEvent", // how
			"1.0", // extent
			tranquilspeak.TriggerTypeSystemControl, // triggerType
			"bridge", // targetSystem
			map[string]interface{}{
				"assigned_port": tnosMCPPort,
			},
		)
		_ = c.triggerMatrix.ProcessTrigger(trigger)
	}

	// Directly try to connect to the primary URL
	_, err := c.tryConnect(TnosMCPURL)
	if err != nil {
		return fmt.Errorf("bridge connection failed: %w", err)
	}
	return nil
}

// Helper for QHP handshake (shared for all fallback routes)
func (c *Bridge) performQHPHandshake(conn *websocket.Conn) error {
	fingerprint := ""
	if c.options.Credentials != nil {
		if val, ok := c.options.Credentials["fingerprint"]; ok {
			fingerprint = val
		}
	}
	if fingerprint == "" {
		fingerprint = GenerateQuantumFingerprint("tnos-mcp") // Use node/system id as seed
	}
	handshakeMsg := map[string]interface{}{
		"type":        "qhp_handshake",
		"fingerprint": fingerprint,
		"timestamp":   time.Now().Unix(),
	}
	if c.options.Credentials != nil {
		if val, ok := c.options.Credentials["developer_override_token"]; ok && val != "" {
			handshakeMsg["override_token"] = val
		}
	}
	handshakeData, err := json.Marshal(handshakeMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal QHP handshake: %w", err)
	}
	err = conn.WriteMessage(websocket.TextMessage, handshakeData)
	if err != nil {
		return fmt.Errorf("failed to send QHP handshake: %w", err)
	}
	ackCh := make(chan error, 1)
	go func() {
		type handshakeAck struct {
			Type        string      `json:"type"`
			Fingerprint string      `json:"fingerprint"`
			TrustTable  interface{} `json:"trust_table"`
			SessionKey  string      `json:"session_key"`
		}
		_, data, err := conn.ReadMessage()
		if err != nil {
			ackCh <- fmt.Errorf("failed to read QHP handshake ack: %w", err)
			return
		}
		var ack handshakeAck
		err = json.Unmarshal(data, &ack)
		if err != nil {
			ackCh <- fmt.Errorf("invalid QHP handshake ack: %w", err)
			return
		}
		if ack.Type != "qhp_handshake_ack" {
			ackCh <- fmt.Errorf("unexpected handshake response type: %s", ack.Type)
			return
		}
		c.sessionKey = ack.SessionKey
		ackCh <- nil
	}()
	select {
	case err := <-ackCh:
		if err != nil {
			return err
		}
	case <-time.After(5 * time.Second):
		return fmt.Errorf("QHP handshake timed out")
	}
	return nil
}

// tryConnect attempts a WebSocket connection to the given URL and updates the client state
func (c *Bridge) tryConnect(serverURL string) (interface{}, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.connected && c.conn != nil {
		return nil, nil
	}
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("invalid bridge URL: %w", err)
	}
	headers := http.Header{}
	for k, v := range c.options.Headers {
		headers.Add(k, v)
	}
	dialer := websocket.Dialer{
		HandshakeTimeout: WriteTimeout,
	}
	dialCtx := c.ctx
	if c.options.Timeout > 0 {
		var cancel context.CancelFunc
		dialCtx, cancel = context.WithTimeout(c.ctx, c.options.Timeout)
		defer cancel()
	}
	conn, resp, err := dialer.DialContext(dialCtx, u.String(), headers)
	if err != nil {
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		return nil, fmt.Errorf("failed to dial WebSocket: %w (status: %d)", err, statusCode)
	}
	// Always perform QHP handshake after connection
	err = c.performQHPHandshake(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}
	c.conn = conn
	c.connected = true
	c.stats.LastActive = time.Now()
	return nil, nil
}

// Receive returns the next message from the bridge
func (c *Bridge) Receive() (common.Message, error) {
	// WHO: MessageReceiver
	// WHAT: Receive bridge messages
	// WHEN: During message processing
	// WHERE: System Layer 6 (Integration)
	// WHY: To get messages from bridge
	// HOW: Using channel
	// EXTENT: Message reception

	select {
	case msg, ok := <-c.messages:
		if !ok {
			return common.Message{}, ErrConnectionClosed
		}
		return msg, nil
	case <-c.ctx.Done():
		return common.Message{}, c.ctx.Err()
	}
}

// Send sends a message to the bridge
func (c *Bridge) Send(msg common.Message) error {
	// WHO: MessageSender
	// WHAT: Send bridge messages with fallback and compression
	// WHEN: During message transmission
	// WHERE: System Layer 6 (Integration)
	// WHY: To ensure robust message delivery
	// HOW: Using FallbackRoute and FormulaRegistry
	// EXTENT: All message sends

	fallbackSend := func() (interface{}, error) {
		c.mu.Lock()
		defer c.mu.Unlock()
		if !c.connected || c.conn == nil {
			return nil, ErrBridgeNotConnected
		}
		if msg.Context == nil && c.options.Context != nil {
			// Check if who key exists in the context map
			if _, ok := c.options.Context["who"]; ok {
				msg.Context = c.options.Context
			}
		}
		// Compression logic can be added here if needed, using shared utilities
		data, err := json.Marshal(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal message: %w", err)
		}
		if WriteTimeout > 0 {
			err = c.conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
			if err != nil {
				c.logger.Error("Failed to set write deadline: %v", err)
			}
		}
		err = c.conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			c.connected = false
			return nil, fmt.Errorf("failed to write message: %w", err)
		}
		c.stats.MessagesSent++
		c.stats.LastActive = time.Now()
		return nil, nil
	}
	_, err := fallbackSend()
	return err
}

// readPump continuously reads messages from the WebSocket
func (c *Bridge) readPump() {
	// WHO: MessagePump
	// WHAT: Read WebSocket messages
	// WHEN: During connection lifetime
	// WHERE: System Layer 6 (Integration)
	// WHY: To process incoming messages
	// HOW: Using WebSocket events
	// EXTENT: Connection lifetime

	defer func() {
		c.close()
	}()

	// Configure read deadline from common.go
	if ReadTimeout > 0 {
		err := c.conn.SetReadDeadline(time.Now().Add(ReadTimeout))
		if err != nil {
			c.logger.Error("Failed to set read deadline: %v", err)
		}
	}

	// Configure maximum message size from common.go
	c.conn.SetReadLimit(MaxMessageSize)

	for {
		// Reset the read deadline for each message
		if ReadTimeout > 0 {
			err := c.conn.SetReadDeadline(time.Now().Add(ReadTimeout))
			if err != nil {
				c.logger.Error("Failed to reset read deadline: %v", err)
			}
		}

		_, data, err := c.conn.ReadMessage()
		if err != nil {
			c.logger.Error("Failed to read message: %v", err)
			c.connected = false

			// Try to reconnect
			if c.ctx.Err() == nil {
				c.reconnect()
			}

			return
		}

		var msg common.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			c.logger.Error("Failed to unmarshal message: %v", err)
			c.stats.ErrorCount++
			continue
		}

		// Process context if needed
		if ctxMap, ok := msg.Context.(map[string]interface{}); ok {
			// Convert context map to ContextVector7D
			contextVector := log.ContextVector7D{
				Who:    getString(ctxMap, "who", "RemoteSystem"),
				What:   getString(ctxMap, "what", "Communication"),
				When:   getInt64(ctxMap, "when", time.Now().Unix()),
				Where:  getString(ctxMap, "where", "RemoteLayer"),
				Why:    getString(ctxMap, "why", "Response"),
				How:    getString(ctxMap, "how", "WebSocket"),
				Extent: getFloat64(ctxMap, "extent", 1.0),
				Source: getString(ctxMap, "source", "External"),
			}
			msg.Context = contextVector
		}

		c.stats.MessagesReceived++
		c.stats.LastActive = time.Now()

		select {
		case c.messages <- msg:
		default:
			c.logger.Warn("Message buffer full, dropping message")
		}
	}
}

// reconnect attempts to reestablish the WebSocket connection
func (c *Bridge) reconnect() {
	// WHO: ReconnectionManager
	// WHAT: Reconnect to bridge with fallback
	// WHEN: During connection loss
	// WHERE: System Layer 6 (Integration)
	// WHY: To restore connectivity robustly
	// HOW: Using FallbackRoute and retry logic
	// EXTENT: Reconnection process

	maxRetries := c.options.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 5
	}
	retryDelay := c.options.RetryDelay
	if retryDelay <= 0 {
		retryDelay = 2 * time.Second
	}
	for i := 0; i < maxRetries; i++ {
		_, err := func() (interface{}, error) {
			c.logger.Info("Attempting to reconnect to bridge (attempt %d)", i+1)
			err := c.connect()
			if err == nil {
				c.logger.Info("Successfully reconnected to bridge")
				c.stats.ReconnectCount++
				go c.readPump()
				return nil, nil
			}
			c.logger.Error("Failed to reconnect (attempt %d): %v", i+1, err)
			return nil, err
		}()
		if err == nil {
			return
		}
		time.Sleep(retryDelay)
	}
	c.logger.Error("Failed to reconnect after maximum attempts (%d)", maxRetries)
}

// close closes the WebSocket connection and cleans up resources
func (c *Bridge) close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	c.connected = false

	// Close the message channel when we're shutting down
	if c.ctx.Err() != nil {
		close(c.messages)
	}
}

// Comments: Mobius/7D context and QHP handshake logic are now enforced for all fallback routes and canonical ports. Quantum symmetry is maintained by mirrored logic and context propagation across all MCP layers.
