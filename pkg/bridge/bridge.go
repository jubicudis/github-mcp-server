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
	"net/http"
	"sync"
	"time"

	"github-mcp-server/pkg/log"
	"github-mcp-server/pkg/translations"

	"github.com/gorilla/websocket"
)

// Canonical QHP endpoints for fallback and connection logic
const (
	TnosMCPURL   = "ws://localhost:9001"
	BridgeURL    = "ws://localhost:10619"
	GitHubMCPURL = "ws://localhost:10617"
	CopilotURL   = "ws://localhost:8083"
)

// ConnectionOptions defines options for creating a new bridge client connection
// ConnectionOptions defines options for creating a new bridge client connection
// Import from common.go to avoid redeclaration

// Client represents a connection to the MCP bridge
// Client represents a connection to the MCP bridge
// Imported from common.go to avoid redeclaration

// Message represents a message sent over the bridge
// Message is imported from mcp_bridge.go to avoid redeclaration
// WHO: MessageCarrier
// WHAT: Bridge message structure
// WHEN: During message exchange
// WHERE: System Layer 6 (Integration)
// WHY: To standardize message format
// HOW: Using structured format with context
// EXTENT: Single message lifecycle

// ConnectionState represents the current state of the bridge connection
// ConnectionState is imported from common.go
// to represent the current state of the bridge connection

// Connection states are imported from common.go
// Instead of redeclaring them here, we use them directly
// See common.go for the following connection states:
// StateDisconnected
// StateConnecting
// StateConnected
// StateReconnecting
// StateError

// BridgeStats is imported from common.go
// For tracking operational statistics

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
	logger               *log.Logger // Add logger field
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
		logger: log.NewLogger(), // Initialize logger
	}
}

// Connect establishes connection to TNOS MCP
func (b *MCPBridge) Connect() error {
	// WHO: ConnectionManager
	// WHAT: Establish connection
	// WHEN: During bridge initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: For communication setup
	// HOW: Using WebSocket protocol with fallback and 7D context
	// EXTENT: Connection establishment

	if b.state == StateConnected {
		return nil
	}

	b.changeState(StateConnecting)

	b.logger.Info("Connecting to TNOS MCP at %s", b.url)

	context7D := b.persistentContext

	// Use FallbackRoute for robust connection attempts
	_, err := FallbackRoute(
		b.ctx,
		"Connect",
		context7D,
		func() (interface{}, error) {
			// Primary connection attempt
			dialer := websocket.Dialer{
				HandshakeTimeout: 60 * time.Second,
			}
			connectCtx, cancel := context.WithTimeout(b.ctx, 60*time.Second)
			defer cancel()
			conn, _, err := dialer.DialContext(connectCtx, b.url, nil)
			if err != nil {
				b.lastError = fmt.Errorf("failed to connect: %w", err)
				b.changeState(StateError)
				b.logger.Error("[7DContext] WHO:MCPBridge WHAT:Connect WHEN:%d WHERE:Layer6 WHY:ConnectionFailure HOW:WebSocket EXTENT:1.0 ERROR:%v", time.Now().Unix(), err)
				if errors.Is(err, context.DeadlineExceeded) {
					b.logger.Warn("[7DContext] Context deadline exceeded during MCPBridge.Connect (timeout=60s)")
				}
				return nil, b.lastError
			}
			b.conn = conn
			b.changeState(StateConnected)
			b.reconnectAttempts = 0
			go b.handleMessages()
			go b.runHealthChecker()
			negErr := b.negotiateVersion()
			if negErr != nil {
				b.logger.Error("[7DContext] Version negotiation failed: %v", negErr)
				return nil, negErr
			}
			b.logger.Info("Connected to TNOS MCP with protocol version %s", b.negotiatedVersion)
			return nil, nil
		},
		func() (interface{}, error) {
			// Fallback: retry with exponential backoff
			for attempt := 1; attempt <= b.maxReconnectAttempts; attempt++ {
				delay := b.reconnectInterval * time.Duration(1<<uint(attempt-1))
				if delay > 60*time.Second {
					delay = 60 * time.Second
				}
				b.logger.Warn("[7DContext] FallbackRoute: waiting %v before reconnect attempt %d", delay, attempt)
				select {
				case <-b.ctx.Done():
					return nil, context.Canceled
				case <-time.After(delay):
				}
				dialer := websocket.Dialer{
					HandshakeTimeout: 60 * time.Second,
				}
				connectCtx, cancel := context.WithTimeout(b.ctx, 60*time.Second)
				defer cancel()
				conn, _, err := dialer.DialContext(connectCtx, b.url, nil)
				if err == nil {
					b.conn = conn
					b.changeState(StateConnected)
					b.reconnectAttempts = 0
					go b.handleMessages()
					go b.runHealthChecker()
					negErr := b.negotiateVersion()
					if negErr != nil {
						b.logger.Error("[7DContext] Version negotiation failed: %v", negErr)
						return nil, negErr
					}
					b.logger.Info("[7DContext] FallbackRoute: reconnected on attempt %d", attempt)
					return nil, nil
				}
				b.logger.Warn("[7DContext] FallbackRoute: reconnect attempt %d failed: %v", attempt, err)
			}
			b.changeState(StateError)
			return nil, errors.New("all fallback connection attempts failed")
		},
		func() (interface{}, error) {
			// GitHub MCP fallback connection logic
			b.logger.Warn("[7DContext] FallbackRoute: attempting GitHub MCP fallback connection")
			// Example: try to connect to GitHub MCP via HTTP or WebSocket (pseudo-code, replace with actual logic)
			// err := githubmcp.Connect()
			// if err != nil { return nil, err }
			// return nil, nil
			return nil, errors.New("GitHub MCP fallback connection not implemented yet")
		},
		func() (interface{}, error) {
			// Copilot LLM fallback connection logic
			b.logger.Warn("[7DContext] FallbackRoute: attempting Copilot LLM fallback connection")
			// Example: try to connect to Copilot LLM (pseudo-code, replace with actual logic)
			// err := copilotllm.Connect()
			// if err != nil { return nil, err }
			// return nil, nil
			return nil, errors.New("Copilot LLM fallback connection not implemented yet")
		},
		b.logger,
	)
	return err
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
			b.logger.Error("Error sending close message: %v", err)
		}

		// Close connection
		err = b.conn.Close()
		if err != nil {
			b.logger.Error("Error closing connection: %v", err)
		}

		b.conn = nil
	}

	// Cancel all operations
	b.cancel()

	// Create new context for future connections
	b.ctx, b.cancel = context.WithCancel(context.Background())

	b.changeState(StateDisconnected)
	b.logger.Info("Disconnected from TNOS MCP")
	return nil
}

// SendMessage sends a message to TNOS MCP using FallbackRoute
func (b *MCPBridge) SendMessage(messageType string, payload map[string]interface{}) error {
	// WHO: MessageSender
	// WHAT: Send MCP message
	// WHEN: During communication
	// WHERE: System Layer 6 (Integration)
	// WHY: For data exchange
	// HOW: Using message protocol with fallback and 7D context
	// EXTENT: Single message

	if b.state != StateConnected {
		return errors.New("bridge not connected")
	}

	message := map[string]interface{}{
		"type":      messageType,
		"timestamp": time.Now().UnixMilli(),
		"payload":   payload,
		"version":   b.negotiatedVersion,
	}
	contextMap := b.persistentContext.ToMap()
	message["context"] = contextMap

	if WriteTimeout > 0 {
		err := b.conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
		if err != nil {
			b.logger.Warn("Failed to set write deadline: %v", err)
		}
	}

	context7D := b.persistentContext

	_, err := FallbackRoute(
		b.ctx,
		"SendMessage",
		context7D,
		func() (interface{}, error) {
			// Primary send
			err := b.conn.WriteJSON(message)
			if err != nil {
				b.recordError(fmt.Errorf("failed to send message: %w", err))
				go b.handleConnectionFailure()
				return nil, err
			}
			b.stats.MessagesSent++
			b.stats.LastActive = time.Now()
			b.stats.Uptime = time.Since(b.stats.StartTime)
			logContext := translations.ContextVector7D{
				Who:    "MCPBridge",
				What:   "MessageSent",
				When:   time.Now().Unix(),
				Where:  "Layer6_Integration",
				Why:    messageType,
				How:    "WebSocket",
				Extent: float64(len(fmt.Sprintf("%v", message))),
			}
			compressedContext := logContext.Compress()
			b.logger.Info("Message sent: %s (context: %+v)", messageType, compressedContext.ToMap())
			return nil, nil
		},
		func() (interface{}, error) {
			// Fallback: try to resend after short delay
			b.logger.Warn("[7DContext] FallbackRoute: retrying message send after 2s")
			time.Sleep(2 * time.Second)
			err := b.conn.WriteJSON(message)
			if err != nil {
				b.recordError(fmt.Errorf("fallback send failed: %w", err))
				go b.handleConnectionFailure()
				return nil, err
			}
			b.stats.MessagesSent++
			b.stats.LastActive = time.Now()
			b.stats.Uptime = time.Since(b.stats.StartTime)
			b.logger.Info("[7DContext] FallbackRoute: message sent after retry")
			return nil, nil
		},
		func() (interface{}, error) {
			// GitHub MCP fallback send logic
			b.logger.Warn("[7DContext] FallbackRoute: attempting GitHub MCP fallback send")
			// Example: send message to GitHub MCP (pseudo-code, replace with actual logic)
			// err := githubmcp.SendMessage(messageType, payload)
			// if err != nil { return nil, err }
			// return nil, nil
			return nil, errors.New("GitHub MCP fallback send not implemented yet")
		},
		func() (interface{}, error) {
			// Copilot LLM fallback send logic
			b.logger.Warn("[7DContext] FallbackRoute: attempting Copilot LLM fallback send")
			// Example: send message to Copilot LLM (pseudo-code, replace with actual logic)
			// err := copilotllm.SendMessage(messageType, payload)
			// if err != nil { return nil, err }
			// return nil, nil
			return nil, errors.New("Copilot LLM fallback send not implemented yet")
		},
		b.logger,
	)
	return err
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

		// Set read deadline using timeout from common.go
		if ReadTimeout > 0 {
			err := b.conn.SetReadDeadline(time.Now().Add(ReadTimeout))
			if err != nil {
				b.logger.Warn("Failed to set read deadline: %v", err)
			}
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
			b.logger.Error("Error parsing message: %v", err)
			continue
		}

		// Extract message type
		messageType, ok := message["type"].(string)
		if !ok {
			b.logger.Warn("Received message without type: %v", message)
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
			b.logger.Info("Message received: %s (context: %+v)", messageType, context.ToMap())
		} else {
			b.logger.Info("Message received: %s (no context)", messageType)
		}

		// Process message with appropriate handler
		b.callbacksMutex.RLock()
		handler, exists := b.messageHandlers[messageType]
		b.callbacksMutex.RUnlock()

		if exists {
			go func() {
				if err := handler(message); err != nil {
					b.logger.Error("Error handling message type %s: %v", messageType, err)
				}
			}()
		} else {
			b.logger.Warn("No handler for message type: %s", messageType)
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
	b.logger.Info("Bridge state changed: %s -> %s", oldState, newState)

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

	b.logger.Warn("Connection failure detected. Reconnect attempt %d of %d",
		b.reconnectAttempts, b.maxReconnectAttempts)

	// Clean up old connection
	if b.conn != nil {
		b.conn.Close()
		b.conn = nil
	}

	// Check if we should keep trying
	if b.reconnectAttempts > b.maxReconnectAttempts {
		b.logger.Error("Maximum reconnect attempts reached (%d). Giving up.",
			b.maxReconnectAttempts)
		b.changeState(StateError)
		return
	}

	// Calculate backoff delay (exponential with cap)
	delay := b.reconnectInterval * time.Duration(1<<uint(b.reconnectAttempts-1))
	if delay > 60*time.Second {
		delay = 60 * time.Second
	}

	b.logger.Warn("Waiting %v before reconnect attempt", delay)

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
		b.logger.Error("Reconnect failed: %v", err)
		go b.handleConnectionFailure() // Schedule another attempt
	} else {
		b.logger.Info("Reconnected successfully")
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
	b.logger.Error("Bridge error: %v", err)
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
					b.logger.Error("Health check failed: %v", err)
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
		b.logger.Warn("Invalid health check response format")
		return
	}

	// Extract timestamps to calculate latency
	pingTime, ok := payload["request_timestamp"].(float64)
	if !ok {
		b.logger.Warn("Health check response missing request timestamp")
		return
	}

	// Calculate round-trip time
	rtt := time.Since(time.UnixMilli(int64(pingTime)))
	b.logger.Info("Health check RTT: %v", rtt)
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

		b.logger.Info("Updated persistent context from message: %+v",
			b.persistentContext.ToMap())
	}
}

// ConnectionOptions is imported from common.go
// Centralized reusable type for connection configuration

// MessageHandler is imported from common.go
// DisconnectHandler is imported from common.go

// BridgeClient provides a compatibility wrapper for transitioning to the new client.go implementation
// This will be removed in a future release
type BridgeClient = Client

// NewClient creates a new MCP bridge client
// WHO: ClientCreator
// WHAT: Create bridge client instance
// WHEN: During client initialization
// WHERE: System Layer 6 (Integration)
// WHY: For system integration
// HOW: Using configuration options
// EXTENT: Client lifecycle
func NewBridgeMCPClient(ctx context.Context, options ConnectionOptions) (*Client, error) {
	// WHO: BridgeClientCreator
	// WHAT: Create MCP bridge client instance
	// WHEN: During client initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: For system integration with unique naming
	// HOW: Using configuration options
	// EXTENT: Client lifecycle
	return newClientImpl(ctx, options)
}

// newClientImpl is the internal implementation of NewClient
func newClientImpl(ctx context.Context, options ConnectionOptions) (*Client, error) {
	// Validation and error checking is handled by the calling function
	if options.ServerURL == "" {
		return nil, errors.New("server URL is required")
	}

	// Set default values if not provided
	if options.Timeout == 0 {
		options.Timeout = 60 * time.Second // Increased from 10s to 60s to avoid context deadline exceeded errors
	}
	if options.MaxRetries == 0 {
		options.MaxRetries = 3
	}
	if options.RetryDelay == 0 {
		options.RetryDelay = time.Second
	}

	// Create a new client instance
	client := &Client{
		options: options,
		state:   StateDisconnected,
		mutex:   sync.RWMutex{},
		stats: BridgeStats{
			StartTime: time.Now(),
		},
	}

	return client, nil
}

// Connect establishes a connection to the MCP bridge
func (c *Client) Connect(ctx context.Context) error {
	// WHO: ConnectionManager
	// WHAT: Connect to MCP bridge
	// WHEN: During client initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To establish communication
	// HOW: Using WebSocket protocol
	// EXTENT: Connection lifetime

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.state == StateConnected {
		return nil // Already connected
	}

	c.setState(StateConnecting)

	// Store the context for later use
	c.ctx, c.cancelFunc = context.WithCancel(ctx)

	// Create headers with context information if available
	headers := http.Header{}
	if c.options.Headers != nil {
		for k, v := range c.options.Headers {
			headers.Set(k, v)
		}
	}

	// Add context information if available
	if c.options.Context.Who != "" {
		contextJSON, err := json.Marshal(c.options.Context)
		if err == nil {
			headers.Set("X-TNOS-Context", string(contextJSON))
		}
	}

	// Connect to the WebSocket server with timeout
	dialer := websocket.Dialer{
		HandshakeTimeout: c.options.Timeout,
	}

	conn, _, err := dialer.DialContext(ctx, c.options.ServerURL, headers)
	if err != nil {
		c.setState(StateError)
		return fmt.Errorf("failed to connect to MCP bridge: %w", err)
	}

	c.conn = conn
	c.setState(StateConnected)
	c.stats.LastActive = time.Now()

	// Start reading messages in a goroutine
	go c.readMessages()

	return nil
}

// OnMessage registers a handler for incoming messages
func (c *Client) OnMessage(handler MessageHandler) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.messageHandler = handler
}

// OnDisconnect registers a handler for disconnection events
func (c *Client) OnDisconnect(handler DisconnectHandler) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.disconnectHandler = handler
}

// readMessages reads messages from the WebSocket connection
func (c *Client) readMessages() {
	logger := log.NewLogger()
	for {
		if c.isClosed {
			return
		}

		_, message, err := c.conn.ReadMessage()
		if err != nil {
			c.handleDisconnect(fmt.Sprintf("read error: %v", err))
			return
		}

		c.stats.MessagesReceived++
		c.stats.LastActive = time.Now()

		// Call the message handler if registered
		c.mutex.RLock()
		handler := c.messageHandler
		c.mutex.RUnlock()

		if handler != nil {
			// Unmarshal the bytes into a Message type before passing to handler
			var parsedMessage Message
			if err := json.Unmarshal(message, &parsedMessage); err != nil {
				logger.Error("Failed to parse message: %v", err)
				continue
			}
			go handler(parsedMessage)
		}
	}
}

// SendMessage sends a message to the MCP bridge
func (c *Client) SendMessage(message []byte) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.state != StateConnected || c.conn == nil {
		return errors.New("not connected to MCP bridge")
	}

	err := c.conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	c.stats.MessagesSent++
	c.stats.LastActive = time.Now()

	return nil
}

// Close closes the connection to the MCP bridge
func (c *Client) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.state == StateDisconnected || c.conn == nil {
		return nil // Already disconnected
	}

	c.isClosed = true
	if c.cancelFunc != nil {
		c.cancelFunc()
	}

	err := c.conn.Close()
	c.setState(StateDisconnected)

	return err
}

// setState updates the connection state
func (c *Client) setState(state ConnectionState) {
	oldState := c.state
	c.state = state

	// Call the disconnect handler if transitioning to disconnected
	if state == StateDisconnected && c.disconnectHandler != nil && oldState != StateDisconnected {
		reason := "normal closure"
		if c.state == StateError {
			reason = "connection error"
		}
		go c.disconnectHandler(reason)
	}
}

// handleDisconnect handles disconnection events
func (c *Client) handleDisconnect(reason string) {
	c.mutex.Lock()

	// Check if already disconnected
	if c.state == StateDisconnected || c.isClosed {
		c.mutex.Unlock()
		return
	}

	// Update state and notify handlers
	c.setState(StateDisconnected)
	c.conn = nil

	handler := c.disconnectHandler
	c.mutex.Unlock()

	if handler != nil {
		go handler(reason)
	}
}

// Example: Use FallbackRoute for connection establishment
func (b *MCPBridge) RobustConnect() error {
	context7D := b.persistentContext

	result, err := FallbackRoute(
		b.ctx,
		"Connect",
		context7D,
		func() (interface{}, error) { return nil, b.Connect() },
		func() (interface{}, error) { return nil, b.Connect() },
		func() (interface{}, error) {
			// GitHub MCP fallback connection logic
			b.logger.Warn("[7DContext] FallbackRoute: attempting GitHub MCP fallback connection")
			// Example: try to connect to GitHub MCP via HTTP or WebSocket (pseudo-code, replace with actual logic)
			// err := githubmcp.Connect()
			// if err != nil { return nil, err }
			// return nil, nil
			return nil, errors.New("GitHub MCP fallback connection not implemented yet")
		},
		func() (interface{}, error) {
			// Copilot LLM fallback connection logic
			b.logger.Warn("[7DContext] FallbackRoute: attempting Copilot LLM fallback connection")
			// Example: try to connect to Copilot LLM (pseudo-code, replace with actual logic)
			// err := copilotllm.Connect()
			// if err != nil { return nil, err }
			// return nil, nil
			return nil, errors.New("Copilot LLM fallback connection not implemented yet")
		},
		b.logger,
	)
	if err != nil {
		return err
	}
	_ = result
	return nil
}

func (b *MCPBridge) CompressWithRegistry(value, entropy, B, V, I, G, F, purpose, t, C_sum float64) (float64, float64, error) {
	registry := GetFormulaRegistry()
	return registry.CompressValue(value, entropy, B, V, I, G, F, purpose, t, C_sum)
}
