// WHO: SharedMCPComponents
// WHAT: Common MCP protocol definitions and structures
// WHEN: During all bridge operations
// WHERE: System Layer 6 (Integration)
// WHY: To provide consistent shared components
// HOW: Using centralized definitions
// EXTENT: All MCP bridge interactions

package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jubicudis/github-mcp-server/pkg/common"
	"github.com/jubicudis/github-mcp-server/pkg/log"
	"github.com/jubicudis/github-mcp-server/pkg/translations"

	"github.com/gorilla/websocket"
)

// DefaultProtocolVersion and other protocol constants are defined in common.go

// Message is imported from common.go
// Representing a message sent over the bridge

// Use Message type from common.go instead of defining a separate ClientMessage type

// BridgeStats is imported from common.go
// WHO: StatsCollector
// WHAT: Bridge operational statistics

// ConnectionOptions is imported from common.go
// Containing options for connecting to the bridge

// Client is already defined in common.go
// The methods in this file operate on the common Client type

// Remove local port constants; use those from common.go

// NewClient creates a new bridge client with the given options
func NewClient(ctx context.Context, options ConnectionOptions) (*Client, error) {
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
	if options.Logger != nil {
		logger = options.Logger
	} else {
		logger = log.NewLogger() // fallback to default logger
	}

	// Ensure context is initialized
	if options.Context.Who == "" {
		options.Context = translations.ContextVector7D{
			Who:    "BridgeClient",
			What:   "Connection",
			When:   time.Now().Unix(),
			Where:  "SystemLayer6",
			Why:    "Communication",
			How:    "WebSocket",
			Extent: 1.0,
			Source: "GitHubMCPServer",
		}
	}

	client := &Client{
		ctx:        clientCtx,
		cancelFunc: cancel,
		options:    options,
		logger:     logger,
		messages:   make(chan Message, 100),
		stats:      BridgeStats{},
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
func (c *Client) connect() error {
	// WHO: ConnectionManager
	// WHAT: Establish WebSocket connection with fallback
	// WHEN: During client (re)connection
	// WHERE: System Layer 6 (Integration)
	// WHY: To ensure robust connection
	// HOW: Using FallbackRoute utility
	// EXTENT: All connection attempts

	// Canonical fallback URLs/ports from common.go
	const (
		TnosMCPURL   = "ws://localhost:9001" // TNOS MCP server (primary) - FIXED: removed /bridge
		BridgeURL    = "ws://localhost:10619/bridge"
		GitHubMCPURL = "ws://localhost:10617/bridge"
		CopilotURL   = "ws://localhost:8083/bridge"
	)

	primaryFn := func() (interface{}, error) {
		if c.options.ServerURL == "" || c.options.ServerURL == TnosMCPURL {
			if c.logger != nil {
				c.logger.Info("[Connect] Trying TNOS MCP server at %s", TnosMCPURL)
			}
			return c.tryConnect(TnosMCPURL)
		}
		return c.tryConnect(c.options.ServerURL)
	}
	bridgeFallback := func() (interface{}, error) {
		if c.options.ServerURL == BridgeURL {
			return nil, fmt.Errorf("Already tried MCP Bridge URL")
		}
		if c.logger != nil {
			c.logger.Warn("[FallbackRoute] Trying MCP Bridge fallback at %s", BridgeURL)
		}
		return c.tryConnect(BridgeURL)
	}
	githubMCPFallback := func() (interface{}, error) {
		if c.options.ServerURL == GitHubMCPURL {
			return nil, fmt.Errorf("Already tried GitHub MCP URL")
		}
		if c.logger != nil {
			c.logger.Warn("[FallbackRoute] Trying GitHub MCP fallback at %s", GitHubMCPURL)
		}
		return c.tryConnect(GitHubMCPURL)
	}
	copilotFallback := func() (interface{}, error) {
		if c.options.ServerURL == CopilotURL {
			return nil, fmt.Errorf("Already tried Copilot LLM URL")
		}
		if c.logger != nil {
			c.logger.Warn("[FallbackRoute] Trying Copilot LLM fallback at %s", CopilotURL)
		}
		return c.tryConnect(CopilotURL)
	}

	context7d := c.options.Context
	operationName := "Connect"
	ctx := c.ctx
	_, err := FallbackRoute(
		ctx,
		operationName,
		context7d,
		primaryFn, // TNOS MCP (primary)
		bridgeFallback,
		githubMCPFallback,
		copilotFallback,
		c.logger,
	)
	if err != nil {
		return err
	}
	return nil
}

// Helper for QHP handshake (shared for all fallback routes)
func (c *Client) performQHPHandshake(conn *websocket.Conn) error {
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
func (c *Client) tryConnect(serverURL string) (interface{}, error) {
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
func (c *Client) Receive() (Message, error) {
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
			return Message{}, ErrConnectionClosed
		}
		return msg, nil
	case <-c.ctx.Done():
		return Message{}, c.ctx.Err()
	}
}

// Send sends a message to the bridge
func (c *Client) Send(msg Message) error {
	// WHO: MessageSender
	// WHAT: Send bridge messages with fallback and compression
	// WHEN: During message transmission
	// WHERE: System Layer 6 (Integration)
	// WHY: To ensure robust message delivery
	// HOW: Using FallbackRoute and FormulaRegistry
	// EXTENT: All message sends

	context7d := c.options.Context
	operationName := "SendMessage"
	ctx := c.ctx
	fallbackSend := func() (interface{}, error) {
		c.mu.Lock()
		defer c.mu.Unlock()
		if !c.connected || c.conn == nil {
			return nil, ErrBridgeNotConnected
		}
		if msg.Context == nil && c.options.Context.Who != "" {
			msg.Context = c.options.Context
		}
		// Compression logic can be added here if needed, using shared utilities
		data, err := json.Marshal(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal message: %w", err)
		}
		if WriteTimeout > 0 {
			err = c.conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
			if err != nil {
				c.logger.Error("Failed to set write deadline", "error", err.Error())
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
	_, err := FallbackRoute(
		ctx,
		operationName,
		context7d,
		fallbackSend,
		func() (interface{}, error) { return nil, fmt.Errorf("Bridge fallback not implemented") },
		func() (interface{}, error) { return nil, fmt.Errorf("GitHub MCP fallback not implemented") },
		func() (interface{}, error) { return nil, fmt.Errorf("Copilot fallback not implemented") },
		c.logger,
	)
	return err
}

// readPump continuously reads messages from the WebSocket
func (c *Client) readPump() {
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
			c.logger.Error("Failed to set read deadline", "error", err.Error())
		}
	}

	// Configure maximum message size from common.go
	c.conn.SetReadLimit(MaxMessageSize)

	for {
		// Reset the read deadline for each message
		if ReadTimeout > 0 {
			err := c.conn.SetReadDeadline(time.Now().Add(ReadTimeout))
			if err != nil {
				c.logger.Error("Failed to reset read deadline", "error", err.Error())
			}
		}

		_, data, err := c.conn.ReadMessage()
		if err != nil {
			c.logger.Error("Failed to read message", "error", err.Error())
			c.connected = false

			// Try to reconnect
			if c.ctx.Err() == nil {
				c.reconnect()
			}

			return
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			c.logger.Error("Failed to unmarshal message", "error", err.Error())
			c.stats.ErrorCount++
			continue
		}

		// Process context if needed
		if ctxMap, ok := msg.Context.(map[string]interface{}); ok {
			// Convert context map to ContextVector7D
			contextVector := translations.ContextVector7D{
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
func (c *Client) reconnect() {
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
	context7d := c.options.Context
	operationName := "Reconnect"
	ctx := c.ctx
	for i := 0; i < maxRetries; i++ {
		_, err := FallbackRoute(
			ctx,
			operationName,
			context7d,
			func() (interface{}, error) {
				c.logger.Info("Attempting to reconnect to bridge", "attempt", i+1)
				err := c.connect()
				if err == nil {
					c.logger.Info("Successfully reconnected to bridge")
					c.stats.ReconnectCount++
					go c.readPump()
					return nil, nil
				}
				c.logger.Error("Failed to reconnect", "attempt", i+1, "error", err.Error())
				return nil, err
			},
			func() (interface{}, error) { return nil, fmt.Errorf("Bridge fallback not implemented") },
			func() (interface{}, error) { return nil, fmt.Errorf("GitHub MCP fallback not implemented") },
			func() (interface{}, error) { return nil, fmt.Errorf("Copilot fallback not implemented") },
			c.logger,
		)
		if err == nil {
			return
		}
		time.Sleep(retryDelay)
	}
	c.logger.Error("Failed to reconnect after maximum attempts", "maxRetries", maxRetries)
}

// close closes the WebSocket connection and cleans up resources
func (c *Client) close() {
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
