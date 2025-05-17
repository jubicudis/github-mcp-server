/*
 * WHO: MCPBridge
 * WHAT: Bridge between GitHub MCP and TNOS MCP
 * WHEN: During cross-system communication
 * WHERE: System Layer 6 (Integration)
 * WHY: To connect GitHub and TNOS systems
 * HOW: Using bidirectional context translation
 * EXTENT: All MCP communications
 */

package bridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/tranquility-dev/github-mcp-server/pkg/log"
	"github.com/tranquility-dev/github-mcp-server/pkg/translations"

	"github.com/gorilla/websocket"
)

// Message types
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

// BridgeOptions configures the MCP bridge
type BridgeOptions struct {
	// WHO: ConfigManager
	// WHAT: Bridge configuration
	// WHEN: During bridge initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To configure bridge behavior
	// HOW: Using structured options
	// EXTENT: All bridge parameters

	URL              string
	Logger           *log.Logger
	ProtocolVersions []string
	ReconnectDelay   time.Duration
	HealthInterval   time.Duration
	MessageBuffer    int
}

// Bridge represents an MCP bridge connection
type Bridge struct {
	// WHO: BridgeManager
	// WHAT: Bridge connection
	// WHEN: During system operation
	// WHERE: System Layer 6 (Integration)
	// WHY: To maintain MCP connection
	// HOW: Using WebSocket
	// EXTENT: Bridge lifecycle

	options BridgeOptions
	conn    *websocket.Conn
	url     string

	// Channel for outgoing messages
	sendCh chan []byte

	// Synchronization
	mu           sync.Mutex
	isConnected  bool
	isShutdown   bool
	reconnecting bool

	// Message handling
	handlers        map[string]MessageHandler
	responseWaiters map[string]chan Message

	// Context
	context *translations.ContextVector7D

	// Stats
	stats struct {
		messagesSent     int64
		messagesReceived int64
		reconnects       int
		errors           int
		lastActivity     time.Time
	}

	// Protocol
	protocolVersion string
	serverFeatures  map[string]interface{}
}

// Message represents an MCP message
type Message struct {
	// WHO: MessageManager
	// WHAT: Message structure
	// WHEN: During message exchange
	// WHERE: System Layer 6 (Integration)
	// WHY: To structure communications
	// HOW: Using JSON format
	// EXTENT: Single message

	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Command string                 `json:"command,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
	Error   *ErrorInfo             `json:"error,omitempty"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	// WHO: ErrorManager
	// WHAT: Error information
	// WHEN: During error handling
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide error details
	// HOW: Using structured format
	// EXTENT: Error information

	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// MessageHandler is a function that handles incoming messages
type MessageHandler func(message Message) error

// NewBridge creates a new MCP bridge
func NewBridge(options BridgeOptions) *Bridge {
	// WHO: BridgeCreator
	// WHAT: Bridge creation
	// WHEN: During system startup
	// WHERE: System Layer 6 (Integration)
	// WHY: To establish MCP connection
	// HOW: Using provided options
	// EXTENT: Bridge creation

	// Set defaults for options
	if options.URL == "" {
		options.URL = DefaultBridgeURL
	}

	if len(options.ProtocolVersions) == 0 {
		options.ProtocolVersions = []string{DefaultProtocolVersion}
	}

	if options.ReconnectDelay <= 0 {
		options.ReconnectDelay = ReconnectDelay
	}

	if options.HealthInterval <= 0 {
		options.HealthInterval = HealthCheckInterval
	}

	if options.MessageBuffer <= 0 {
		options.MessageBuffer = MessageBufferSize
	}

	// Create default logger if none provided
	if options.Logger == nil {
		options.Logger = log.NewLogger()
		// Use the correct method to configure logger
		// If no configuration methods are available, use as is
	}

	// Create 7D context for the bridge
	now := time.Now().Unix()
	context := &translations.ContextVector7D{
		Who:    "MCPBridge",
		What:   "SystemIntegration",
		When:   now,
		Where:  "System Layer 6 (Integration)",
		Why:    "Connect GitHub and TNOS",
		How:    "Bidirectional Translation",
		Extent: 1.0,
		Meta: map[string]interface{}{
			"B":         0.8, // Base factor
			"V":         0.7, // Value factor
			"I":         0.9, // Intent factor
			"G":         1.2, // Growth factor
			"F":         0.6, // Flexibility factor
			"createdAt": now,
		},
		Source: "github_mcp",
	}

	return &Bridge{
		options:         options,
		url:             options.URL,
		sendCh:          make(chan []byte, options.MessageBuffer),
		handlers:        make(map[string]MessageHandler),
		responseWaiters: make(map[string]chan Message),
		context:         context,
		serverFeatures:  make(map[string]interface{}),
		protocolVersion: DefaultProtocolVersion,
	}
}

// Connect establishes connection to the MCP server
func (b *Bridge) Connect() error {
	// WHO: ConnectionManager
	// WHAT: Establish connection
	// WHEN: During system startup
	// WHERE: System Layer 6 (Integration)
	// WHY: To connect to MCP server
	// HOW: Using WebSocket protocol
	// EXTENT: Connection establishment

	b.mu.Lock()
	if b.isConnected {
		b.mu.Unlock()
		return nil
	}
	b.reconnecting = true
	b.mu.Unlock()

	// Parse URL
	u, err := url.Parse(b.url)
	if err != nil {
		return fmt.Errorf("invalid bridge URL: %w", err)
	}

	// Connect with context
	headers := http.Header{}
	headers.Add("X-MCP-Protocol-Version", b.protocolVersion)
	headers.Add("X-MCP-Client", BridgeServiceName)

	// Add context as header
	contextJSON, _ := json.Marshal(b.context.TranslateToMCP())
	headers.Add("X-MCP-Context", string(contextJSON))

	b.options.Logger.Info("Connecting to MCP bridge",
		"url", b.url,
		"protocol", b.protocolVersion)

	// Establish connection
	dialer := websocket.Dialer{
		HandshakeTimeout: 60 * time.Second, // Increased from 10s to 60s to avoid context deadline exceeded errors
	}

	conn, resp, err := dialer.Dial(u.String(), headers)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("connection failed with status %d: %w", resp.StatusCode, err)
		}
		return fmt.Errorf("connection failed: %w", err)
	}

	// Configure connection
	conn.SetReadLimit(MaxMessageSize)

	// Store connection
	b.mu.Lock()
	b.conn = conn
	b.isConnected = true
	b.reconnecting = false
	b.stats.lastActivity = time.Now()
	b.mu.Unlock()

	// Perform handshake
	err = b.performHandshake()
	if err != nil {
		b.options.Logger.Error("Handshake failed", "error", err.Error())
		b.disconnect()
		return fmt.Errorf("handshake failed: %w", err)
	}

	// Start message processing
	go b.readPump()
	go b.writePump()
	go b.healthCheckPump()

	b.options.Logger.Info("Connected to MCP bridge",
		"url", b.url,
		"protocol", b.protocolVersion)

	return nil
}

// Disconnect closes the bridge connection
func (b *Bridge) Disconnect() {
	// WHO: DisconnectionManager
	// WHAT: Close connection
	// WHEN: During system shutdown
	// WHERE: System Layer 6 (Integration)
	// WHY: To close MCP connection
	// HOW: Using clean shutdown
	// EXTENT: Connection termination

	b.mu.Lock()
	defer b.mu.Unlock()

	b.isShutdown = true

	// Close connection
	b.disconnect()

	// Close channel
	close(b.sendCh)

	b.options.Logger.Info("Disconnected from MCP bridge")
}

// disconnect closes the connection without locking
func (b *Bridge) disconnect() {
	// WHO: ConnectionCleanup
	// WHAT: Internal disconnect
	// WHEN: During connection issues
	// WHERE: System Layer 6 (Integration)
	// WHY: For clean connection closure
	// HOW: Using WebSocket close
	// EXTENT: Connection resources

	if b.conn != nil {
		// Close with normal closure message
		msg := websocket.FormatCloseMessage(
			websocket.CloseNormalClosure,
			"client disconnecting",
		)

		// Write close message with timeout
		_ = b.conn.WriteControl(
			websocket.CloseMessage,
			msg,
			time.Now().Add(WriteTimeout),
		)

		b.conn.Close()
		b.conn = nil
	}

	b.isConnected = false
}

// reconnect attempts to reestablish the connection
func (b *Bridge) reconnect() {
	// WHO: ReconnectionManager
	// WHAT: Reestablish connection
	// WHEN: During connection failures
	// WHERE: System Layer 6 (Integration)
	// WHY: To recover from disconnection
	// HOW: Using retry mechanism
	// EXTENT: Connection recovery

	b.mu.Lock()
	if b.isShutdown || b.reconnecting {
		b.mu.Unlock()
		return
	}
	b.reconnecting = true
	b.mu.Unlock()

	attempt := 0
	maxAttempts := MaxReconnectAttempts

	for attempt < maxAttempts {
		attempt++
		b.options.Logger.Info("Attempting reconnection",
			"attempt", attempt,
			"max", maxAttempts)

		// Pause before retry
		time.Sleep(b.options.ReconnectDelay)

		// Try to connect
		err := b.Connect()
		if err == nil {
			b.options.Logger.Info("Reconnected successfully")

			b.mu.Lock()
			b.stats.reconnects++
			b.reconnecting = false
			b.mu.Unlock()

			return
		}

		b.options.Logger.Error("Reconnection failed",
			"attempt", attempt,
			"error", err.Error())
	}

	b.options.Logger.Error("Failed to reconnect after maximum attempts",
		"attempts", maxAttempts)

	b.mu.Lock()
	b.reconnecting = false
	b.mu.Unlock()
}

// performHandshake sends a handshake to establish protocol version
func (b *Bridge) performHandshake() error {
	// WHO: HandshakeManager
	// WHAT: Protocol handshake
	// WHEN: During connection initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To negotiate protocol
	// HOW: Using handshake message
	// EXTENT: Protocol negotiation

	handshake := Message{
		ID:   generateMessageID(),
		Type: TypeHandshake,
		Payload: map[string]interface{}{
			"protocol":          "mcp",
			"supportedVersions": b.options.ProtocolVersions,
			"client":            BridgeServiceName,
			"clientFeatures": map[string]interface{}{
				"context7D":      true,
				"compression":    true,
				"binaryMessages": false,
				"healthCheck":    true,
			},
		},
		Context: b.context.TranslateToMCP(),
	}

	// Send handshake and wait for response
	response, err := b.SendAndWait(handshake, 10*time.Second)
	if err != nil {
		return fmt.Errorf("handshake failed: %w", err)
	}

	// Check for errors
	if response.Error != nil {
		return fmt.Errorf("handshake rejected: %s", response.Error.Message)
	}

	// Extract protocol version and features
	if version, ok := response.Payload["version"].(string); ok {
		b.protocolVersion = version
	}

	if features, ok := response.Payload["serverFeatures"].(map[string]interface{}); ok {
		b.serverFeatures = features
	}

	// Log successful handshake
	b.options.Logger.Info("Handshake successful",
		"protocol", b.protocolVersion,
		"features", len(b.serverFeatures))

	return nil
}

// RegisterHandler adds a message handler for a specific type
func (b *Bridge) RegisterHandler(messageType string, handler MessageHandler) {
	// WHO: HandlerManager
	// WHAT: Register message handler
	// WHEN: During initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To process specific messages
	// HOW: Using handler registration
	// EXTENT: Message handling

	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[messageType] = handler
	b.options.Logger.Debug("Registered handler", "type", messageType)
}

// Send sends a message through the bridge
func (b *Bridge) Send(message Message) error {
	// WHO: MessageSender
	// WHAT: Send message
	// WHEN: During communication
	// WHERE: System Layer 6 (Integration)
	// WHY: To transmit data
	// HOW: Using message serialization
	// EXTENT: Single message

	// Check if connected
	b.mu.Lock()
	if !b.isConnected {
		b.mu.Unlock()
		return errors.New("bridge not connected")
	}
	b.mu.Unlock()

	// Generate ID if not provided
	if message.ID == "" {
		message.ID = generateMessageID()
	}

	// Add context if not provided
	if message.Context == nil {
		message.Context = b.context.TranslateToMCP()
	}

	// Convert to JSON
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Send to channel
	select {
	case b.sendCh <- data:
		b.mu.Lock()
		b.stats.messagesSent++
		b.mu.Unlock()
		return nil
	default:
		return errors.New("message buffer full")
	}
}

// SendAndWait sends a message and waits for a response
func (b *Bridge) SendAndWait(message Message, timeout time.Duration) (Message, error) {
	// WHO: SyncCommunicator
	// WHAT: Synchronous communication
	// WHEN: During request-response
	// WHERE: System Layer 6 (Integration)
	// WHY: To get synchronized response
	// HOW: Using message correlation
	// EXTENT: Request-response cycle

	// Generate ID if not provided
	if message.ID == "" {
		message.ID = generateMessageID()
	}

	// Create response channel
	responseCh := make(chan Message, 1)

	// Register waiter
	b.mu.Lock()
	b.responseWaiters[message.ID] = responseCh
	b.mu.Unlock()

	// Cleanup on exit
	defer func() {
		b.mu.Lock()
		delete(b.responseWaiters, message.ID)
		b.mu.Unlock()
	}()

	// Send the message
	if err := b.Send(message); err != nil {
		return Message{}, err
	}

	// Wait for response with timeout
	select {
	case response := <-responseCh:
		return response, nil
	case <-time.After(timeout):
		return Message{}, errors.New("response timeout")
	}
}

// readPump processes incoming messages
func (b *Bridge) readPump() {
	// WHO: MessageReader
	// WHAT: Process incoming messages
	// WHEN: During active connection
	// WHERE: System Layer 6 (Integration)
	// WHY: To receive communications
	// HOW: Using message deserialization
	// EXTENT: All incoming messages

	defer b.handleReadPumpShutdown()

	// Configure read deadline
	_ = b.conn.SetReadDeadline(time.Now().Add(ReadTimeout))

	// Set pong handler to reset deadline
	b.conn.SetPongHandler(func(string) error {
		_ = b.conn.SetReadDeadline(time.Now().Add(ReadTimeout))
		return nil
	})

	for {
		// Read message
		_, data, err := b.conn.ReadMessage()
		if err != nil {
			b.handleReadError(err)
			break
		}

		// Process the message data
		b.processIncomingMessage(data)
	}
}

// handleReadPumpShutdown manages the readPump shutdown process
func (b *Bridge) handleReadPumpShutdown() {
	// WHO: ShutdownManager
	// WHAT: Manage read pump shutdown
	// WHEN: When read pump exits
	// WHERE: System Layer 6 (Integration)
	// WHY: For clean shutdown handling
	// HOW: Using deferred execution
	// EXTENT: Read pump shutdown process

	b.options.Logger.Info("Read pump stopping")

	b.mu.Lock()
	isShutdown := b.isShutdown
	b.mu.Unlock()

	if !isShutdown {
		// Try to reconnect if not explicitly shut down
		go b.reconnect()
	}
}

// handleReadError processes WebSocket read errors
func (b *Bridge) handleReadError(err error) {
	// WHO: ErrorHandler
	// WHAT: Process WebSocket errors
	// WHEN: During read failures
	// WHERE: System Layer 6 (Integration)
	// WHY: To handle connection issues
	// HOW: Using error classification
	// EXTENT: Read error handling

	if websocket.IsUnexpectedCloseError(
		err,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway) {

		b.options.Logger.Error("WebSocket read error",
			"error", err.Error())

		b.mu.Lock()
		b.stats.errors++
		b.mu.Unlock()
	}
}

// processIncomingMessage parses and handles an incoming message
func (b *Bridge) processIncomingMessage(data []byte) {
	// WHO: MessageProcessor
	// WHAT: Process message data
	// WHEN: When message is received
	// WHERE: System Layer 6 (Integration)
	// WHY: To handle incoming messages
	// HOW: Using message parsing
	// EXTENT: Single message processing

	// Parse message
	var message Message
	if err := json.Unmarshal(data, &message); err != nil {
		b.options.Logger.Error("Failed to parse message",
			"error", err.Error())
		return
	}

	// Update stats
	b.updateMessageStats()

	// Process incoming context
	b.processIncomingContext(message)

	// Log message receipt
	b.options.Logger.Debug("Received message",
		"id", message.ID,
		"type", message.Type)

	// Handle the parsed message
	b.routeMessage(message)
}

// updateMessageStats updates the message statistics
func (b *Bridge) updateMessageStats() {
	// WHO: StatsUpdater
	// WHAT: Update message statistics
	// WHEN: After message receipt
	// WHERE: System Layer 6 (Integration)
	// WHY: To track message activity
	// HOW: Using atomic counter updates
	// EXTENT: Message statistics

	b.mu.Lock()
	b.stats.messagesReceived++
	b.stats.lastActivity = time.Now()
	b.mu.Unlock()
}

// processIncomingContext processes context from incoming messages
func (b *Bridge) processIncomingContext(message Message) {
	// WHO: ContextProcessor
	// WHAT: Process message context
	// WHEN: During message handling
	// WHERE: System Layer 6 (Integration)
	// WHY: To maintain context state
	// HOW: Using context translation
	// EXTENT: Context processing

	if message.Context != nil {
		// Translate MCP context to 7D context
		incomingContext := translations.TranslateFromMCP(message.Context)

		// Merge with existing context
		if b.context != nil {
			mergedContext := incomingContext.Merge(*b.context)
			b.context = &mergedContext
		} else {
			b.context = &incomingContext
		}
	}
}

// routeMessage sends the message to waiter or handler
func (b *Bridge) routeMessage(message Message) {
	// WHO: MessageRouter
	// WHAT: Route message to destination
	// WHEN: After message processing
	// WHERE: System Layer 6 (Integration)
	// WHY: To deliver message to handler
	// HOW: Using message routing logic
	// EXTENT: Message routing

	// Check if this is a response to a waiting request
	if b.sendToWaiter(message) {
		return
	}

	// Otherwise, dispatch to handler
	b.dispatchToHandler(message)
}

// sendToWaiter sends message to waiting response channel
func (b *Bridge) sendToWaiter(message Message) bool {
	// WHO: ResponseHandler
	// WHAT: Handle response messages
	// WHEN: During message routing
	// WHERE: System Layer 6 (Integration)
	// WHY: To complete request-response cycle
	// HOW: Using channel communication
	// EXTENT: Response handling

	b.mu.Lock()
	responseCh, isWaiting := b.responseWaiters[message.ID]
	b.mu.Unlock()

	if isWaiting {
		// Send to waiter
		responseCh <- message
		return true
	}
	return false
}

// dispatchToHandler sends message to appropriate handler
func (b *Bridge) dispatchToHandler(message Message) {
	// WHO: HandlerDispatcher
	// WHAT: Dispatch to message handler
	// WHEN: During message routing
	// WHERE: System Layer 6 (Integration)
	// WHY: To process message by type
	// HOW: Using handler lookup
	// EXTENT: Handler dispatching

	b.mu.Lock()
	handler, hasHandler := b.handlers[message.Type]
	b.mu.Unlock()

	if hasHandler {
		// Handle in goroutine to avoid blocking
		go func() {
			if err := handler(message); err != nil {
				b.options.Logger.Error("Handler error",
					"type", message.Type,
					"id", message.ID,
					"error", err.Error())
			}
		}()
	} else {
		b.options.Logger.Debug("No handler for message",
			"type", message.Type)
	}
}

// writePump sends outgoing messages
func (b *Bridge) writePump() {
	// WHO: MessageWriter
	// WHAT: Send outgoing messages
	// WHEN: During active connection
	// WHERE: System Layer 6 (Integration)
	// WHY: To transmit communications
	// HOW: Using WebSocket protocol
	// EXTENT: All outgoing messages

	ticker := time.NewTicker(PingInterval)
	defer func() {
		ticker.Stop()
		b.options.Logger.Info("Write pump stopping")
	}()

	for {
		select {
		case message, ok := <-b.sendCh:
			// Check if channel is closed
			if !ok {
				return
			}

			// Get connection with lock
			b.mu.Lock()
			conn := b.conn
			b.mu.Unlock()

			if conn == nil {
				b.options.Logger.Error("Cannot send, not connected")
				continue
			}

			// Set write deadline
			_ = conn.SetWriteDeadline(time.Now().Add(WriteTimeout))

			// Write message
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				b.options.Logger.Error("Write error", "error", err.Error())

				b.mu.Lock()
				b.stats.errors++
				b.mu.Unlock()

				return
			}

			// Update last activity
			b.mu.Lock()
			b.stats.lastActivity = time.Now()
			b.mu.Unlock()

		case <-ticker.C:
			// Send ping
			b.mu.Lock()
			conn := b.conn
			b.mu.Unlock()

			if conn != nil {
				_ = conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					b.options.Logger.Error("Ping error", "error", err.Error())
					return
				}
			}
		}
	}
}

// healthCheckPump performs periodic health checks
func (b *Bridge) healthCheckPump() {
	// WHO: HealthMonitor
	// WHAT: Periodic health checks
	// WHEN: During active connection
	// WHERE: System Layer 6 (Integration)
	// WHY: To monitor connection health
	// HOW: Using health messages
	// EXTENT: Connection reliability

	ticker := time.NewTicker(b.options.HealthInterval)
	defer ticker.Stop()

	for range ticker.C {
		b.mu.Lock()
		if !b.isConnected || b.isShutdown {
			b.mu.Unlock()
			return
		}

		// Skip health check if recent activity
		timeSinceActivity := time.Since(b.stats.lastActivity)
		if timeSinceActivity < b.options.HealthInterval {
			b.mu.Unlock()
			continue
		}
		b.mu.Unlock()

		// Send health check
		healthCheck := Message{
			ID:   generateMessageID(),
			Type: TypeHealthCheck,
			Payload: map[string]interface{}{
				"timestamp": time.Now().Unix(),
			},
			Context: b.context.TranslateToMCP(),
		}

		b.options.Logger.Debug("Sending health check")

		// Don't wait for response, just send
		if err := b.Send(healthCheck); err != nil {
			b.options.Logger.Error("Health check error", "error", err.Error())
		}
	}
}

// GetStats returns bridge statistics
func (b *Bridge) GetStats() map[string]interface{} {
	// WHO: StatisticsManager
	// WHAT: Collect statistics
	// WHEN: During monitoring
	// WHERE: System Layer 6 (Integration)
	// WHY: To report bridge status
	// HOW: Using metrics collection
	// EXTENT: Operational metrics

	b.mu.Lock()
	defer b.mu.Unlock()

	return map[string]interface{}{
		"connected":        b.isConnected,
		"reconnecting":     b.reconnecting,
		"messagesSent":     b.stats.messagesSent,
		"messagesReceived": b.stats.messagesReceived,
		"reconnects":       b.stats.reconnects,
		"errors":           b.stats.errors,
		"lastActivity":     b.stats.lastActivity.Format(time.RFC3339),
		"protocolVersion":  b.protocolVersion,
		"url":              b.url,
	}
}

// SendCommand sends a command message
func (b *Bridge) SendCommand(command string, payload map[string]interface{}) (Message, error) {
	// WHO: CommandSender
	// WHAT: Send command message
	// WHEN: During command operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To execute remote commands
	// HOW: Using command messages
	// EXTENT: Single command

	message := Message{
		ID:      generateMessageID(),
		Type:    TypeCommand,
		Command: command,
		Payload: payload,
	}

	return b.SendAndWait(message, 30*time.Second)
}

// SendQuery sends a query message
func (b *Bridge) SendQuery(command string, payload map[string]interface{}) (Message, error) {
	// WHO: QuerySender
	// WHAT: Send query message
	// WHEN: During query operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To request information
	// HOW: Using query messages
	// EXTENT: Single query

	message := Message{
		ID:      generateMessageID(),
		Type:    TypeQuery,
		Command: command,
		Payload: payload,
	}

	return b.SendAndWait(message, 30*time.Second)
}

// SendContextSync synchronizes context with TNOS
func (b *Bridge) SendContextSync() error {
	// WHO: ContextSynchronizer
	// WHAT: Synchronize context
	// WHEN: During context operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To align system contexts
	// HOW: Using context messages
	// EXTENT: Full context state

	// Compress context
	compressedContext := b.context.Compress()

	message := Message{
		ID:      generateMessageID(),
		Type:    TypeContext,
		Command: "sync",
		Payload: map[string]interface{}{
			"context": compressedContext.ToMap(),
			"source":  "github_mcp",
		},
		Context: compressedContext.TranslateToMCP(),
	}

	_, err := b.SendAndWait(message, 10*time.Second)
	return err
}

// Helper functions

// generateMessageID generates a unique message ID
func generateMessageID() string {
	// WHO: IDGenerator
	// WHAT: Generate message ID
	// WHEN: During message creation
	// WHERE: System Layer 6 (Integration)
	// WHY: For message identification
	// HOW: Using timestamp and random
	// EXTENT: Single message ID

	return fmt.Sprintf("msg-%d-%d", time.Now().UnixNano(), time.Now().Unix()%1000)
}
