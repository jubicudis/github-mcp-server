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

	"github.com/gorilla/websocket"
	"github.com/jubicudis/github-mcp-server/pkg/log"
	"github.com/jubicudis/github-mcp-server/pkg/translations"
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
	
	logger := options.Logger
	if logger == nil {
		logger = log.NewLogger()
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
		stats: BridgeStats{
			StartTime:  time.Now(),
			LastActive: time.Now(),
		},
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
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.connected && c.conn != nil {
		return nil
	}
	
	u, err := url.Parse(c.options.ServerURL)
	if err != nil {
		return fmt.Errorf("invalid bridge URL: %w", err)
	}
	
	headers := http.Header{}
	for k, v := range c.options.Headers {
		headers.Add(k, v)
	}
	
	dialer := websocket.Dialer{
		HandshakeTimeout: WriteTimeout, // Use the timeout constant from common.go
	}
	
	// Use a context with timeout for the dial operation
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
		return fmt.Errorf("failed to dial WebSocket: %w (status: %d)", err, statusCode)
	}
	
	c.conn = conn
	c.connected = true
	c.stats.LastActive = time.Now()
	
	return nil
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
	// WHAT: Send bridge messages
	// WHEN: During message transmission
	// WHERE: System Layer 6 (Integration)
	// WHY: To send messages to bridge
	// HOW: Using WebSocket
	// EXTENT: Message transmission

	c.mu.Lock()
	defer c.mu.Unlock()
	
	if !c.connected || c.conn == nil {
		return ErrBridgeNotConnected
	}
	
	// Ensure context is properly set
	if msg.Context == nil && c.options.Context.Who != "" {
		msg.Context = c.options.Context
	}
	
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	// Set write deadline using WriteTimeout from common.go
	if WriteTimeout > 0 {
		err = c.conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
		if err != nil {
			c.logger.Error("Failed to set write deadline", "error", err.Error())
		}
	}
	
	err = c.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		c.connected = false
		return fmt.Errorf("failed to write message: %w", err)
	}
	
	c.stats.MessagesSent++
	c.stats.LastActive = time.Now()
	
	return nil
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
	// WHAT: Reconnect to bridge
	// WHEN: During connection loss
	// WHERE: System Layer 6 (Integration)
	// WHY: To restore connectivity
	// HOW: Using retry logic
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
		c.logger.Info("Attempting to reconnect to bridge", "attempt", i+1)
		
		err := c.connect()
		if err == nil {
			c.logger.Info("Successfully reconnected to bridge")
			c.stats.ReconnectCount++
			
			// Restart the read pump
			go c.readPump()
			return
		}
		
		c.logger.Error("Failed to reconnect", "attempt", i+1, "error", err.Error())
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
