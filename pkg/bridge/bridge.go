/*
 * WHO: MCPBridge
 * WHAT: Connection between GitHub MCP and TNOS MCP
 * WHEN: During all cross-system operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To enable seamless communication between systems
 * HOW: Using bidirectional protocol translation
 * EXTENT: All MCP operations
 */

package bridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/translations"
)

// Supported MCP protocol versions
const (
	// WHO: VersionManager
	// WHAT: MCP protocol versions
	// WHEN: During connection setup
	// WHERE: System Layer 6 (Integration)
	// WHY: For protocol compatibility
	// HOW: Using version negotiation
	// EXTENT: All MCP operations

	MCPVersion10 = "1.0"
	MCPVersion20 = "2.0"
	MCPVersion30 = "3.0"
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

// MCPBridge provides connection between GitHub MCP and TNOS MCP
type MCPBridge struct {
	// WHO: BridgeManager
	// WHAT: Bridge component
	// WHEN: During system operation
	// WHERE: System Layer 6 (Integration)
	// WHY: For system integration
	// HOW: Using websocket connection
	// EXTENT: All cross-system operations

	conn                 *websocket.Conn
	url                  string
	negotiatedVersion    string
	state                ConnectionState
	stats                BridgeStats
	lastError            error
	reconnectAttempts    int
	maxReconnectAttempts int
	reconnectInterval    time.Duration
	callbacksMutex       sync.RWMutex
	messageHandlers      map[string]func(message map[string]interface{}) error
	stateHandlers        map[string]func(oldState, newState ConnectionState)
	ctx                  context.Context
	cancel               context.CancelFunc
	healthCheckInterval  time.Duration
	persistentContext    translations.ContextVector7D
}

// NewMCPBridge creates a new bridge instance
func NewMCPBridge(url string) *MCPBridge {
	// WHO: BridgeCreator
	// WHAT: Create bridge instance
	// WHEN: During system initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: For system integration
	// HOW: Using configuration
	// EXTENT: Bridge lifecycle

	ctx, cancel := context.WithCancel(context.Background())

	return &MCPBridge{
		url:                  url,
		negotiatedVersion:    MCPVersion30, // Start with latest version
		state:                StateDisconnected,
		reconnectAttempts:    0,
		maxReconnectAttempts: 10,
		reconnectInterval:    5 * time.Second,
		messageHandlers:      make(map[string]func(message map[string]interface{}) error),
		stateHandlers:        make(map[string]func(oldState, newState ConnectionState)),
		ctx:                  ctx,
		cancel:               cancel,
		healthCheckInterval:  30 * time.Second,
		stats: BridgeStats{
			StartTime:  time.Now(),
			LastActive: time.Now(),
		},
		persistentContext: translations.NewContext(
			"MCPBridge",
			"BridgeOperation",
			"Layer6_Integration",
			"SystemIntegration",
			"Protocol_Translation",
			1.0,
		),
	}
}

// Connect establishes connection to TNOS MCP
func (b *MCPBridge) Connect() error {
	// WHO: ConnectionManager
	// WHAT: Establish connection
	// WHEN: During bridge initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: For communication setup
	// HOW: Using WebSocket protocol
	// EXTENT: Connection establishment

	if b.state == StateConnected {
		return nil
	}

	b.changeState(StateConnecting)

	// Attempt connection
	log.Printf("Connecting to TNOS MCP at %s", b.url)
	conn, _, err := websocket.DefaultDialer.Dial(b.url, nil)

	if err != nil {
		b.lastError = fmt.Errorf("failed to connect: %w", err)
		b.changeState(StateError)
		return b.lastError
	}

	b.conn = conn
	b.changeState(StateConnected)

	// Reset reconnect attempts on successful connection
	b.reconnectAttempts = 0

	// Start message handler
	go b.handleMessages()

	// Start health checker
	go b.runHealthChecker()

	// Perform version negotiation
	err = b.negotiateVersion()
	if err != nil {
		return err
	}

	log.Printf("Connected to TNOS MCP with protocol version %s", b.negotiatedVersion)
	return nil
}

// Disconnect closes the connection
func (b *MCPBridge) Disconnect() error {
	// WHO: ConnectionManager
	// WHAT: Close connection
	// WHEN: During shutdown
	// WHERE: System Layer 6 (Integration)
	// WHY: For clean termination
	// HOW: Using controlled closure
	// EXTENT: Connection termination

	if b.state == StateDisconnected {
		return nil
	}

	if b.conn != nil {
		// Send close message
		err := b.conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Disconnecting"),
		)
		if err != nil {
			log.Printf("Error sending close message: %v", err)
		}

		// Close connection
		err = b.conn.Close()
		if err != nil {
			log.Printf("Error closing connection: %v", err)
		}

		b.conn = nil
	}

	// Cancel all operations
	b.cancel()

	// Create new context for future connections
	b.ctx, b.cancel = context.WithCancel(context.Background())

	b.changeState(StateDisconnected)
	log.Println("Disconnected from TNOS MCP")
	return nil
}

// SendMessage sends a message to TNOS MCP
func (b *MCPBridge) SendMessage(messageType string, payload map[string]interface{}) error {
	// WHO: MessageSender
	// WHAT: Send MCP message
	// WHEN: During communication
	// WHERE: System Layer 6 (Integration)
	// WHY: For data exchange
	// HOW: Using message protocol
	// EXTENT: Single message

	if b.state != StateConnected {
		return errors.New("bridge not connected")
	}

	// Add message type and timestamp
	message := map[string]interface{}{
		"type":      messageType,
		"timestamp": time.Now().UnixMilli(),
		"payload":   payload,
		"version":   b.negotiatedVersion,
	}

	// Add context information
	contextMap := b.persistentContext.ToMap()
	message["context"] = contextMap

	// Serialize and send
	err := b.conn.WriteJSON(message)
	if err != nil {
		b.recordError(fmt.Errorf("failed to send message: %w", err))
		go b.handleConnectionFailure()
		return err
	}

	// Update stats
	b.stats.MessagesSent++
	b.stats.LastActive = time.Now()
	b.stats.Uptime = time.Since(b.stats.StartTime)

	// Log message
	logContext := translations.ContextVector7D{
		Who:    "MCPBridge",
		What:   "MessageSent",
		When:   time.Now().Unix(),
		Where:  "Layer6_Integration",
		Why:    messageType,
		How:    "WebSocket",
		Extent: float64(len(fmt.Sprintf("%v", message))),
	}

	// Compress context for logging
	compressedContext := logContext.Compress()
	log.Printf("Message sent: %s (context: %+v)", messageType, compressedContext.ToMap())

	return nil
}

// RegisterMessageHandler registers a handler for a specific message type
func (b *MCPBridge) RegisterMessageHandler(messageType string, handler func(message map[string]interface{}) error) {
	// WHO: HandlerRegistrar
	// WHAT: Register message handler
	// WHEN: During initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: For message processing
	// HOW: Using callback registration
	// EXTENT: Message handling

	b.callbacksMutex.Lock()
	defer b.callbacksMutex.Unlock()
	b.messageHandlers[messageType] = handler
}

// RegisterStateHandler registers a handler for state changes
func (b *MCPBridge) RegisterStateHandler(id string, handler func(oldState, newState ConnectionState)) {
	// WHO: StateHandlerRegistrar
	// WHAT: Register state handler
	// WHEN: During initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: For state monitoring
	// HOW: Using callback registration
	// EXTENT: State transitions

	b.callbacksMutex.Lock()
	defer b.callbacksMutex.Unlock()
	b.stateHandlers[id] = handler
}

// UnregisterStateHandler removes a state handler
func (b *MCPBridge) UnregisterStateHandler(id string) {
	// WHO: StateHandlerRegistrar
	// WHAT: Remove state handler
	// WHEN: During cleanup
	// WHERE: System Layer 6 (Integration)
	// WHY: For handler management
	// HOW: Using callback removal
	// EXTENT: State transitions

	b.callbacksMutex.Lock()
	defer b.callbacksMutex.Unlock()
	delete(b.stateHandlers, id)
}

// GetState returns the current connection state
func (b *MCPBridge) GetState() ConnectionState {
	// WHO: StateReporter
	// WHAT: Get connection state
	// WHEN: During monitoring
	// WHERE: System Layer 6 (Integration)
	// WHY: For status reporting
	// HOW: Using state accessor
	// EXTENT: Connection state

	return b.state
}

// GetStats returns bridge statistics
func (b *MCPBridge) GetStats() BridgeStats {
	// WHO: StatsReporter
	// WHAT: Get operational stats
	// WHEN: During monitoring
	// WHERE: System Layer 6 (Integration)
	// WHY: For performance analysis
	// HOW: Using stats accessor
	// EXTENT: System health

	// Update uptime before returning
	b.stats.Uptime = time.Since(b.stats.StartTime)
	return b.stats
}

// GetLastError returns the last error
func (b *MCPBridge) GetLastError() error {
	// WHO: ErrorReporter
	// WHAT: Get last error
	// WHEN: During diagnostics
	// WHERE: System Layer 6 (Integration)
	// WHY: For troubleshooting
	// HOW: Using error accessor
	// EXTENT: Error reporting

	return b.lastError
}

// SetPersistentContext updates the context that will be sent with all messages
func (b *MCPBridge) SetPersistentContext(context translations.ContextVector7D) {
	// WHO: ContextManager
	// WHAT: Update persistent context
	// WHEN: During configuration
	// WHERE: System Layer 6 (Integration)
	// WHY: For consistent context
	// HOW: Using context update
	// EXTENT: All messages

	b.persistentContext = context
}

// GetPersistentContext returns the current persistent context
func (b *MCPBridge) GetPersistentContext() translations.ContextVector7D {
	// WHO: ContextProvider
	// WHAT: Get persistent context
	// WHEN: During monitoring
	// WHERE: System Layer 6 (Integration)
	// WHY: For context inspection
	// HOW: Using context accessor
	// EXTENT: Context reporting

	return b.persistentContext
}

// RunHealthCheck triggers an immediate health check
func (b *MCPBridge) RunHealthCheck() error {
	// WHO: HealthMonitor
	// WHAT: Perform health check
	// WHEN: During maintenance
	// WHERE: System Layer 6 (Integration)
	// WHY: For connection verification
	// HOW: Using ping message
	// EXTENT: Connection health

	if b.state != StateConnected {
		return errors.New("bridge not connected")
	}

	return b.sendHealthCheckPing()
}

// handleMessages processes incoming messages
func (b *MCPBridge) handleMessages() {
	// WHO: MessageProcessor
	// WHAT: Process incoming messages
	// WHEN: During communication
	// WHERE: System Layer 6 (Integration)
	// WHY: For data reception
	// HOW: Using message handlers
	// EXTENT: All incoming messages

	for {
		// Check if context is canceled
		select {
		case <-b.ctx.Done():
			return
		default:
			// Continue processing
		}

		if b.conn == nil {
			// Connection closed, exit loop
			return
		}

		// Read next message
		_, rawMessage, err := b.conn.ReadMessage()
		if err != nil {
			b.recordError(fmt.Errorf("failed to read message: %w", err))
			go b.handleConnectionFailure()
			return
		}

		// Update stats
		b.stats.MessagesReceived++
		b.stats.LastActive = time.Now()

		// Parse message
		var message map[string]interface{}
		if err := json.Unmarshal(rawMessage, &message); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Extract message type
		messageType, ok := message["type"].(string)
		if !ok {
			log.Printf("Received message without type: %v", message)
			continue
		}

		// Handle health check responses specially
		if messageType == "health_check_response" {
			b.handleHealthCheckResponse(message)
			continue
		}

		// Log message receipt with context
		if contextData, ok := message["context"].(map[string]interface{}); ok {
			context := translations.FromMap(contextData)
			log.Printf("Message received: %s (context: %+v)", messageType, context.ToMap())
		} else {
			log.Printf("Message received: %s (no context)", messageType)
		}

		// Process message with appropriate handler
		b.callbacksMutex.RLock()
		handler, exists := b.messageHandlers[messageType]
		b.callbacksMutex.RUnlock()

		if exists {
			go func() {
				if err := handler(message); err != nil {
					log.Printf("Error handling message type %s: %v", messageType, err)
				}
			}()
		} else {
			log.Printf("No handler for message type: %s", messageType)
		}
	}
}

// changeState updates the connection state
func (b *MCPBridge) changeState(newState ConnectionState) {
	// WHO: StateManager
	// WHAT: Update connection state
	// WHEN: During state transitions
	// WHERE: System Layer 6 (Integration)
	// WHY: For state tracking
	// HOW: Using state mutation
	// EXTENT: Connection lifecycle

	oldState := b.state
	if oldState == newState {
		return
	}

	b.state = newState
	log.Printf("Bridge state changed: %s -> %s", oldState, newState)

	// Notify state handlers
	b.callbacksMutex.RLock()
	handlers := make([]func(oldState, newState ConnectionState), 0, len(b.stateHandlers))
	for _, handler := range b.stateHandlers {
		handlers = append(handlers, handler)
	}
	b.callbacksMutex.RUnlock()

	// Call handlers outside the lock
	for _, handler := range handlers {
		go handler(oldState, newState)
	}
}

// handleConnectionFailure manages reconnection attempts
func (b *MCPBridge) handleConnectionFailure() {
	// WHO: ReconnectManager
	// WHAT: Handle connection failure
	// WHEN: During connection issues
	// WHERE: System Layer 6 (Integration)
	// WHY: For connection resilience
	// HOW: Using reconnect strategy
	// EXTENT: Connection recovery

	// Already reconnecting
	if b.state == StateReconnecting {
		return
	}

	b.changeState(StateReconnecting)
	b.reconnectAttempts++
	b.stats.ReconnectCount++

	log.Printf("Connection failure detected. Reconnect attempt %d of %d",
		b.reconnectAttempts, b.maxReconnectAttempts)

	// Clean up old connection
	if b.conn != nil {
		b.conn.Close()
		b.conn = nil
	}

	// Check if we should keep trying
	if b.reconnectAttempts > b.maxReconnectAttempts {
		log.Printf("Maximum reconnect attempts reached (%d). Giving up.",
			b.maxReconnectAttempts)
		b.changeState(StateError)
		return
	}

	// Calculate backoff delay (exponential with cap)
	delay := b.reconnectInterval * time.Duration(1<<uint(b.reconnectAttempts-1))
	if delay > 60*time.Second {
		delay = 60 * time.Second
	}

	log.Printf("Waiting %v before reconnect attempt", delay)

	// Wait before reconnecting
	select {
	case <-b.ctx.Done():
		return
	case <-time.After(delay):
		// Continue with reconnect
	}

	// Attempt reconnection
	err := b.Connect()
	if err != nil {
		log.Printf("Reconnect failed: %v", err)
		go b.handleConnectionFailure() // Schedule another attempt
	} else {
		log.Printf("Reconnected successfully")
	}
}

// negotiateVersion handles protocol version negotiation
func (b *MCPBridge) negotiateVersion() error {
	// WHO: ProtocolNegotiator
	// WHAT: Negotiate protocol version
	// WHEN: During connection setup
	// WHERE: System Layer 6 (Integration)
	// WHY: For protocol compatibility
	// HOW: Using version exchange
	// EXTENT: Protocol compatibility

	message := map[string]interface{}{
		"supportedVersions": []string{MCPVersion10, MCPVersion20, MCPVersion30},
		"preferredVersion":  MCPVersion30,
	}

	err := b.SendMessage("version_negotiation", message)
	if err != nil {
		return fmt.Errorf("failed to send version negotiation: %w", err)
	}

	// Version negotiation response will be processed by the message handler
	// We'll default to the highest version we support initially
	return nil
}

// recordError tracks error statistics
func (b *MCPBridge) recordError(err error) {
	// WHO: ErrorTracker
	// WHAT: Record error
	// WHEN: During error handling
	// WHERE: System Layer 6 (Integration)
	// WHY: For error tracking
	// HOW: Using error registration
	// EXTENT: Error management

	b.lastError = err
	b.stats.ErrorCount++
	log.Printf("Bridge error: %v", err)
}

// runHealthChecker performs periodic health checks
func (b *MCPBridge) runHealthChecker() {
	// WHO: HealthMonitor
	// WHAT: Regular health checks
	// WHEN: During system operation
	// WHERE: System Layer 6 (Integration)
	// WHY: For connection monitoring
	// HOW: Using periodic pings
	// EXTENT: Connection health

	ticker := time.NewTicker(b.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			if b.state == StateConnected {
				// Only run health checks when connected
				if err := b.sendHealthCheckPing(); err != nil {
					log.Printf("Health check failed: %v", err)
				}
			}
		}
	}
}

// sendHealthCheckPing sends a health check message
func (b *MCPBridge) sendHealthCheckPing() error {
	// WHO: HealthChecker
	// WHAT: Send health ping
	// WHEN: During health monitoring
	// WHERE: System Layer 6 (Integration)
	// WHY: For connection verification
	// HOW: Using ping message
	// EXTENT: Connection health

	return b.SendMessage("health_check_ping", map[string]interface{}{
		"timestamp": time.Now().UnixMilli(),
	})
}

// handleHealthCheckResponse processes health check responses
func (b *MCPBridge) handleHealthCheckResponse(message map[string]interface{}) {
	// WHO: HealthResponder
	// WHAT: Process health response
	// WHEN: During health monitoring
	// WHERE: System Layer 6 (Integration)
	// WHY: For connection verification
	// HOW: Using response handling
	// EXTENT: Connection health

	payload, ok := message["payload"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid health check response format")
		return
	}

	// Extract timestamps to calculate latency
	pingTime, ok := payload["request_timestamp"].(float64)
	if !ok {
		log.Printf("Health check response missing request timestamp")
		return
	}

	// Calculate round-trip time
	rtt := time.Since(time.UnixMilli(int64(pingTime)))
	log.Printf("Health check RTT: %v", rtt)
}

// UpdateContextFromMessage updates the persistent context from a message
func (b *MCPBridge) UpdateContextFromMessage(message map[string]interface{}) {
	// WHO: ContextUpdater
	// WHAT: Update context from message
	// WHEN: During message processing
	// WHERE: System Layer 6 (Integration)
	// WHY: For context synchronization
	// HOW: Using context extraction
	// EXTENT: Context maintenance

	if contextData, ok := message["context"].(map[string]interface{}); ok {
		newContext := translations.FromMap(contextData)

		// Merge with existing context
		mergedContext := b.persistentContext.Merge(newContext)
		b.persistentContext = mergedContext

		log.Printf("Updated persistent context from message: %+v",
			b.persistentContext.ToMap())
	}
}
