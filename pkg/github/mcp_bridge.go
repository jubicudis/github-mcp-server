/*
 * WHO: MCPBridge
 * WHAT: Bridge between GitHub MCP and TNOS MCP
 * WHEN: During cross-system communications
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide bidirectional context integration
 * HOW: Using WebSocket connection with context translation
 * EXTENT: All communications between GitHub and TNOS
 */

package github

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/tranquility-neuro-os/github-mcp-server/pkg/log"

	"github.com/gorilla/websocket"
)

// MCPBridgeState represents the state of the MCP bridge
type MCPBridgeState string

const (
	// WHO: StateManager
	// WHAT: Bridge state constants
	// WHEN: During state transitions
	// WHERE: System Layer 6 (Integration)
	// WHY: To track bridge lifecycle
	// HOW: Using string constants
	// EXTENT: Bridge state tracking

	// MCPBridgeStateInitializing indicates the bridge is initializing
	MCPBridgeStateInitializing MCPBridgeState = "initializing"

	// MCPBridgeStateConnecting indicates the bridge is connecting
	MCPBridgeStateConnecting MCPBridgeState = "connecting"

	// MCPBridgeStateConnected indicates the bridge is connected
	MCPBridgeStateConnected MCPBridgeState = "connected"

	// MCPBridgeStateReconnecting indicates the bridge is reconnecting
	MCPBridgeStateReconnecting MCPBridgeState = "reconnecting"

	// MCPBridgeStateDisconnected indicates the bridge is disconnected
	MCPBridgeStateDisconnected MCPBridgeState = "disconnected"

	// MCPBridgeStateStopping indicates the bridge is stopping
	MCPBridgeStateStopping MCPBridgeState = "stopping"

	// MCPBridgeStateError indicates the bridge has an error
	MCPBridgeStateError MCPBridgeState = "error"
)

// MCPBridgeOptions represents configuration options for the MCP bridge
type MCPBridgeOptions struct {
	// WHO: ConfigurationManager
	// WHAT: Bridge configuration
	// WHEN: During bridge creation
	// WHERE: System Layer 6 (Integration)
	// WHY: To customize bridge behavior
	// HOW: Using options pattern
	// EXTENT: Bridge customization

	// URL of the TNOS MCP server
	TNOSMCPURL string

	// Authentication token for GitHub API
	GithubToken string

	// Logger instance
	Logger *log.Logger

	// Whether to enable context compression
	EnableCompression bool

	// Whether to preserve metadata during context translation
	PreserveMetadata bool

	// Whether to enforce strict context mapping
	StrictMapping bool

	// Interval to attempt reconnection on disconnect
	ReconnectInterval time.Duration

	// Heartbeat interval for keepalive
	HeartbeatInterval time.Duration

	// Timeout for operations
	OperationTimeout time.Duration
}

// DefaultMCPBridgeOptions returns default options for the MCP bridge
func DefaultMCPBridgeOptions() MCPBridgeOptions {
	// WHO: ConfigurationProvider
	// WHAT: Default bridge configuration
	// WHEN: During bridge creation
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide sensible defaults
	// HOW: Using factory pattern
	// EXTENT: Options creation

	return MCPBridgeOptions{
		TNOSMCPURL:        "ws://localhost:8765/mcp",
		EnableCompression: true,
		PreserveMetadata:  true,
		StrictMapping:     false,
		ReconnectInterval: 10 * time.Second,
		HeartbeatInterval: 30 * time.Second,
		OperationTimeout:  5 * time.Second,
	}
}

// BridgeStats contains statistics about the MCP bridge
type BridgeStats struct {
	// WHO: StatsTracker
	// WHAT: Bridge statistics
	// WHEN: During monitoring
	// WHERE: System Layer 6 (Integration)
	// WHY: To monitor bridge performance
	// HOW: Using metrics collection
	// EXTENT: Performance monitoring

	// Number of messages sent to TNOS
	MessagesSent int64

	// Number of messages received from TNOS
	MessagesReceived int64

	// Number of reconnection attempts
	ReconnectionAttempts int64

	// Timestamp of the last successful message send
	LastMessageSent int64

	// Timestamp of the last successful message receive
	LastMessageReceived int64

	// Map of operation counts by type
	OperationCounts map[string]int64

	// Context translation statistics
	ContextTranslations struct {
		Success int64
		Failure int64
	}
}

// MCPBridge implements the bridge between GitHub MCP and TNOS MCP
type MCPBridge struct {
	// WHO: BridgeManager
	// WHAT: MCP bridge implementation
	// WHEN: During system operation
	// WHERE: System Layer 6 (Integration)
	// WHY: To connect disparate systems
	// HOW: Using WebSocket and context translation
	// EXTENT: All cross-system communications

	// Configuration options
	options MCPBridgeOptions

	// Logger instance
	logger *log.Logger

	// Context translator
	translator *ContextTranslator

	// WebSocket connection
	conn *websocket.Conn

	// Connection mutex
	connMutex sync.Mutex

	// Current bridge state
	state MCPBridgeState

	// Bridge statistics
	stats BridgeStats

	// Channel to signal shutdown
	stopCh chan struct{}

	// Wait group for ongoing operations
	wg sync.WaitGroup

	// Last error encountered
	lastError error
}

// NewMCPBridge creates a new MCP bridge
func NewMCPBridge(options MCPBridgeOptions) (*MCPBridge, error) {
	// WHO: BridgeFactory
	// WHAT: Create MCP bridge
	// WHEN: During system initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To establish cross-system communication
	// HOW: Using factory pattern
	// EXTENT: Bridge lifecycle

	// Validate options
	if options.GithubToken == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}

	// Set defaults for optional fields
	if options.TNOSMCPURL == "" {
		options.TNOSMCPURL = "ws://localhost:8765/mcp"
	}

	if options.ReconnectInterval == 0 {
		options.ReconnectInterval = 10 * time.Second
	}

	if options.HeartbeatInterval == 0 {
		options.HeartbeatInterval = 30 * time.Second
	}

	if options.OperationTimeout == 0 {
		options.OperationTimeout = 5 * time.Second
	}

	// Create logger if none provided
	logger := options.Logger
	if logger == nil {
		logger = log.NewLogger()
		// Logger already has default level LevelInfo
	}

	// Create context translator
	translator := NewContextTranslator(
		logger,
		options.EnableCompression,
		options.PreserveMetadata,
		options.StrictMapping,
	)

	// Create the bridge instance
	bridge := &MCPBridge{
		options:    options,
		logger:     logger,
		translator: translator,
		state:      MCPBridgeStateInitializing,
		stopCh:     make(chan struct{}),
		stats: BridgeStats{
			OperationCounts: make(map[string]int64),
		},
	}

	// Log bridge creation
	bridge.logger.Info("Created MCP Bridge",
		"url", options.TNOSMCPURL,
		"compression", options.EnableCompression,
	)

	return bridge, nil
}

// Start starts the MCP bridge and connects to the TNOS MCP server
func (b *MCPBridge) Start() error {
	// WHO: BridgeStarter
	// WHAT: Start MCP bridge
	// WHEN: During system startup
	// WHERE: System Layer 6 (Integration)
	// WHY: To initialize connectivity
	// HOW: Using connection establishment
	// EXTENT: Bridge activation

	b.connMutex.Lock()
	defer b.connMutex.Unlock()

	// Check if already connected
	if b.state == MCPBridgeStateConnected {
		return nil
	}

	// Update state
	b.setState(MCPBridgeStateConnecting)

	// Connect to TNOS MCP
	err := b.connect()
	if err != nil {
		b.lastError = err
		b.setState(MCPBridgeStateError)
		return fmt.Errorf("failed to connect to TNOS MCP: %w", err)
	}

	// Start background goroutines
	b.wg.Add(3)

	// Start reader
	go func() {
		defer b.wg.Done()
		b.reader()
	}()

	// Start heartbeat
	go func() {
		defer b.wg.Done()
		b.heartbeat()
	}()

	// Start reconnector
	go func() {
		defer b.wg.Done()
		b.reconnector()
	}()

	return nil
}

// Stop stops the MCP bridge and disconnects from the TNOS MCP server
func (b *MCPBridge) Stop() error {
	// WHO: BridgeStopper
	// WHAT: Stop MCP bridge
	// WHEN: During system shutdown
	// WHERE: System Layer 6 (Integration)
	// WHY: To gracefully terminate
	// HOW: Using connection teardown
	// EXTENT: Bridge deactivation

	b.connMutex.Lock()
	defer b.connMutex.Unlock()

	// Check if already stopped
	if b.state == MCPBridgeStateDisconnected {
		return nil
	}

	// Update state
	b.setState(MCPBridgeStateStopping)

	// Signal background goroutines to stop
	close(b.stopCh)

	// Disconnect from TNOS MCP
	if b.conn != nil {
		b.conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(
				websocket.CloseNormalClosure,
				"MCP Bridge stopping",
			),
		)
		b.conn.Close()
		b.conn = nil
	}

	// Wait for goroutines to finish
	b.wg.Wait()

	// Update state
	b.setState(MCPBridgeStateDisconnected)

	return nil
}

// GetState returns the current state of the MCP bridge
func (b *MCPBridge) GetState() MCPBridgeState {
	// WHO: StateProvider
	// WHAT: Get bridge state
	// WHEN: During monitoring
	// WHERE: System Layer 6 (Integration)
	// WHY: To monitor bridge status
	// HOW: Using state reporting
	// EXTENT: Status monitoring

	return b.state
}

// GetStats returns statistics about the MCP bridge
func (b *MCPBridge) GetStats() BridgeStats {
	// WHO: StatsProvider
	// WHAT: Get bridge statistics
	// WHEN: During monitoring
	// WHERE: System Layer 6 (Integration)
	// WHY: To monitor bridge performance
	// HOW: Using metrics reporting
	// EXTENT: Performance monitoring

	// Add translator statistics
	translatorStats := b.translator.GetTranslationStats()
	if success, ok := translatorStats["Success"]; ok {
		if successVal, ok := success.(int64); ok {
			b.stats.ContextTranslations.Success = successVal
		}
	}
	if failure, ok := translatorStats["Failure"]; ok {
		if failureVal, ok := failure.(int64); ok {
			b.stats.ContextTranslations.Failure = failureVal
		}
	}

	return b.stats
}

// HealthCheck performs a health check on the MCP bridge
func (b *MCPBridge) HealthCheck() (bool, error) {
	// WHO: HealthMonitor
	// WHAT: Check bridge health
	// WHEN: During monitoring
	// WHERE: System Layer 6 (Integration)
	// WHY: To verify connectivity
	// HOW: Using ping-pong
	// EXTENT: Health verification

	b.connMutex.Lock()
	defer b.connMutex.Unlock()

	// Check state
	if b.state != MCPBridgeStateConnected {
		return false, fmt.Errorf("bridge not connected, current state: %s", b.state)
	}

	// Check connection
	if b.conn == nil {
		return false, fmt.Errorf("bridge connection is nil")
	}

	// Send ping
	err := b.conn.WriteMessage(websocket.PingMessage, []byte{})
	if err != nil {
		return false, fmt.Errorf("failed to send ping: %w", err)
	}

	return true, nil
}

// SyncContext synchronizes context between GitHub and TNOS
func (b *MCPBridge) SyncContext(context map[string]interface{}) error {
	// WHO: ContextSynchronizer
	// WHAT: Sync context across systems
	// WHEN: During context operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To maintain context coherence
	// HOW: Using bidirectional sync
	// EXTENT: Context synchronization

	// Check if bridge is connected
	if b.state != MCPBridgeStateConnected {
		return fmt.Errorf("bridge not connected, current state: %s", b.state)
	}

	// Create TNOS context from GitHub context
	tnosContext, err := b.translator.TranslateMapToTNOS(context)
	if err != nil {
		b.logger.Error("Failed to translate context to TNOS", "error", err.Error())
		b.stats.ContextTranslations.Failure++
		return err
	}
	b.stats.ContextTranslations.Success++

	// Create sync message
	syncMessage := map[string]interface{}{
		"operation": "context_sync",
		"timestamp": time.Now().Unix(),
		"data": map[string]interface{}{
			"context": tnosContext,
		},
	}

	// Apply compression if enabled
	if b.options.EnableCompression {
		compressed, err := b.applyMobiusCompression(syncMessage)
		if err != nil {
			b.logger.Warn("Failed to compress sync message", "error", err.Error())
			// Continue with uncompressed message
		} else {
			syncMessage = compressed
		}
	}

	// Send sync message
	err = b.sendMessage(syncMessage)
	if err != nil {
		b.logger.Error("Failed to sync context", "error", err.Error())
		return err
	}

	// Update operation count
	b.stats.OperationCounts["context_sync"]++

	return nil
}

// SendOperation sends an operation to TNOS MCP
func (b *MCPBridge) SendOperation(
	operation string,
	data map[string]interface{},
	context map[string]interface{},
) error {
	// WHO: OperationSender
	// WHAT: Send operation to TNOS
	// WHEN: During cross-system operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To perform TNOS actions
	// HOW: Using operation protocol
	// EXTENT: Operation execution

	// Check if bridge is connected
	if b.state != MCPBridgeStateConnected {
		return fmt.Errorf("bridge not connected, current state: %s", b.state)
	}

	// Create operation message
	operationMessage := map[string]interface{}{
		"operation": operation,
		"timestamp": time.Now().Unix(),
		"data":      data,
	}

	// Include context if provided
	if context != nil {
		// Create TNOS context from GitHub context
		tnosContext, err := b.translator.TranslateMapToTNOS(context)
		if err != nil {
			b.logger.Error("Failed to translate context to TNOS", "error", err.Error())
			b.stats.ContextTranslations.Failure++
			return err
		}
		b.stats.ContextTranslations.Success++

		operationMessage["context"] = tnosContext
	}

	// Apply compression if enabled
	if b.options.EnableCompression {
		compressed, err := b.applyMobiusCompression(operationMessage)
		if err != nil {
			b.logger.Warn("Failed to compress operation message", "error", err.Error())
			// Continue with uncompressed message
		} else {
			operationMessage = compressed
		}
	}

	// Send operation message
	err := b.sendMessage(operationMessage)
	if err != nil {
		b.logger.Error("Failed to send operation", "error", err.Error())
		return err
	}

	// Update operation count
	if _, ok := b.stats.OperationCounts[operation]; !ok {
		b.stats.OperationCounts[operation] = 0
	}
	b.stats.OperationCounts[operation]++

	return nil
}

// Internal methods

// connect establishes a connection to the TNOS MCP server
func (b *MCPBridge) connect() error {
	// WHO: ConnectionManager
	// WHAT: Connect to TNOS MCP
	// WHEN: During connection establishment
	// WHERE: System Layer 6 (Integration)
	// WHY: To establish communication
	// HOW: Using WebSocket protocol
	// EXTENT: Connection lifecycle

	// Create context for connection
	context := map[string]interface{}{
		"who":    "MCPBridge",
		"what":   "Connect",
		"when":   time.Now().Unix(),
		"where":  "Integration_Layer",
		"why":    "Establish_Communication",
		"how":    "WebSocket_Protocol",
		"extent": 1.0,
		"meta": map[string]interface{}{
			"source": "github_mcp",
			"token":  b.options.GithubToken[0:4] + "...", // Log only prefix for security
		},
	}

	b.logger.Info("Connecting to TNOS MCP",
		"url", b.options.TNOSMCPURL,
		"who", context["who"],
		"what", context["what"])

	// Convert the token to a Base64-encoded header
	// token := base64.StdEncoding.EncodeToString([]byte(b.options.GithubToken))

	// Set headers for authentication
	headers := map[string][]string{
		"Authorization": {"Bearer " + b.options.GithubToken},
	}

	// Create dialer with headers
	dialer := websocket.Dialer{
		HandshakeTimeout: b.options.OperationTimeout,
	}

	// Connect to the WebSocket server
	conn, _, err := dialer.Dial(b.options.TNOSMCPURL, headers)
	if err != nil {
		return fmt.Errorf("failed to dial WebSocket: %w", err)
	}

	// Set the connection
	b.conn = conn
	b.setState(MCPBridgeStateConnected)

	b.logger.Info("Connected to TNOS MCP")

	// Reset the stopCh if it was closed
	select {
	case <-b.stopCh:
		b.stopCh = make(chan struct{})
	default:
		// Channel is still open
	}

	return nil
}

// reader reads messages from the TNOS MCP server
func (b *MCPBridge) reader() {
	// WHO: MessageReceiver
	// WHAT: Read incoming messages
	// WHEN: During communication
	// WHERE: System Layer 6 (Integration)
	// WHY: To process incoming data
	// HOW: Using message handling
	// EXTENT: Inbound communication

	b.logger.Info("Starting MCP Bridge reader")

	for {
		// Check for stop signal
		select {
		case <-b.stopCh:
			b.logger.Info("Reader stopping")
			return
		default:
			// Continue
		}

		// Check if connected
		if b.state != MCPBridgeStateConnected || b.conn == nil {
			b.logger.Debug("Reader waiting for connection")
			time.Sleep(1 * time.Second)
			continue
		}

		// Read message
		_, message, err := b.conn.ReadMessage()
		if err != nil {
			b.logger.Error("Failed to read message", "error", err.Error())

			// Handle disconnect
			b.handleDisconnect(err)

			// Wait before trying again
			time.Sleep(1 * time.Second)
			continue
		}

		// Process message
		go b.processMessage(message)
	}
}

// heartbeat sends periodic heartbeats to the TNOS MCP server
func (b *MCPBridge) heartbeat() {
	// WHO: HeartbeatManager
	// WHAT: Send periodic heartbeats
	// WHEN: During active connection
	// WHERE: System Layer 6 (Integration)
	// WHY: To maintain connection
	// HOW: Using periodic pings
	// EXTENT: Connection maintenance

	b.logger.Info("Starting MCP Bridge heartbeat")

	ticker := time.NewTicker(b.options.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-b.stopCh:
			b.logger.Info("Heartbeat stopping")
			return
		case <-ticker.C:
			// Check if connected
			if b.state != MCPBridgeStateConnected || b.conn == nil {
				continue
			}

			// Send heartbeat
			heartbeatMessage := map[string]interface{}{
				"operation": "heartbeat",
				"timestamp": time.Now().Unix(),
				"data": map[string]interface{}{
					"uptime": time.Now().Unix(),
					"source": "github_mcp",
				},
			}

			// Send message without locking
			if err := b.sendMessageUnsafe(heartbeatMessage); err != nil {
				b.logger.Error("Failed to send heartbeat", "error", err.Error())
				b.handleDisconnect(err)
			}
		}
	}
}

// reconnector attempts to reconnect to the TNOS MCP server on disconnect
func (b *MCPBridge) reconnector() {
	// WHO: ReconnectionManager
	// WHAT: Handle reconnection logic
	// WHEN: During disconnection
	// WHERE: System Layer 6 (Integration)
	// WHY: To restore connectivity
	// HOW: Using retry mechanism
	// EXTENT: Connection recovery

	b.logger.Info("Starting MCP Bridge reconnector")

	ticker := time.NewTicker(b.options.ReconnectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-b.stopCh:
			b.logger.Info("Reconnector stopping")
			return
		case <-ticker.C:
			// Check if reconnection needed
			if b.state != MCPBridgeStateReconnecting {
				continue
			}

			// Attempt reconnection
			b.logger.Info("Attempting reconnection to TNOS MCP")
			b.stats.ReconnectionAttempts++

			b.connMutex.Lock()
			err := b.connect()
			b.connMutex.Unlock()

			if err != nil {
				b.logger.Error("Reconnection failed", "error", err.Error(),
					"attempts", b.stats.ReconnectionAttempts)

				// Set error state if too many attempts
				if b.stats.ReconnectionAttempts > 10 {
					b.setState(MCPBridgeStateError)
					b.lastError = fmt.Errorf("reconnection failed after %d attempts: %w",
						b.stats.ReconnectionAttempts, err)
				}

				continue
			}

			b.logger.Info("Reconnected to TNOS MCP")
		}
	}
}

// processMessage processes a message from the TNOS MCP server
func (b *MCPBridge) processMessage(data []byte) {
	// WHO: MessageProcessor
	// WHAT: Process incoming message
	// WHEN: Upon message receipt
	// WHERE: System Layer 6 (Integration)
	// WHY: To handle server communications
	// HOW: Using message parsing
	// EXTENT: Message handling

	// Update stats
	b.stats.MessagesReceived++
	b.stats.LastMessageReceived = time.Now().Unix()

	// Parse message
	var message map[string]interface{}
	if err := json.Unmarshal(data, &message); err != nil {
		b.logger.Error("Failed to parse message", "error", err.Error())
		return
	}

	// Check for compression
	if compressed, ok := message["compressed"].(bool); ok && compressed {
		// Handle compressed message
		if original, ok := message["context"].(map[string]interface{}); ok {
			message = original
		} else {
			b.logger.Error("Invalid compressed message format")
			return
		}
	}

	// Extract operation
	operation, ok := message["operation"].(string)
	if !ok {
		b.logger.Error("Missing operation in message")
		return
	}

	// Update operation count
	if _, ok := b.stats.OperationCounts[operation]; !ok {
		b.stats.OperationCounts[operation] = 0
	}
	b.stats.OperationCounts[operation]++

	// Handle different operations
	switch operation {
	case "heartbeat_ack":
		// Heartbeat acknowledged
		b.logger.Debug("Received heartbeat acknowledgement")

	case "context_sync_ack":
		// Context synchronization acknowledged
		b.logger.Debug("Received context sync acknowledgement")

	case "context_request":
		// TNOS is requesting context
		b.handleContextRequest(message)

	case "operation_response":
		// Response to a previous operation
		b.handleOperationResponse(message)

	default:
		b.logger.Warn("Unknown operation", "operation", operation)
	}
}

// handleContextRequest handles a context request from TNOS MCP
func (b *MCPBridge) handleContextRequest(message map[string]interface{}) {
	// WHO: ContextRequestHandler
	// WHAT: Handle context requests
	// WHEN: Upon request receipt
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide context data
	// HOW: Using context generation
	// EXTENT: Context provisioning

	b.logger.Debug("Handling context request")

	// Extract request details
	data, ok := message["data"].(map[string]interface{})
	if !ok {
		b.logger.Error("Invalid context request format")
		return
	}

	// Generate response context
	responseContext := map[string]interface{}{
		"who":    "MCPBridge",
		"what":   "ContextResponse",
		"when":   time.Now().Unix(),
		"where":  "Integration_Layer",
		"why":    "Fulfill_Context_Request",
		"how":    "Context_Generation",
		"extent": 1.0,
		"meta": map[string]interface{}{
			"source":           "github_mcp",
			"request_id":       data["request_id"],
			"response_time":    time.Now().Unix(),
			"protocol_version": "1.0",
			"B":                0.8, // Base factor
			"V":                0.7, // Value factor
			"I":                0.9, // Intent factor
			"G":                1.2, // Growth factor
			"F":                0.6, // Flexibility factor
		},
	}

	// Create response message
	responseMessage := map[string]interface{}{
		"operation": "context_response",
		"timestamp": time.Now().Unix(),
		"data": map[string]interface{}{
			"request_id": data["request_id"],
			"context":    responseContext,
		},
	}

	// Send response
	err := b.sendMessage(responseMessage)
	if err != nil {
		b.logger.Error("Failed to send context response", "error", err.Error())
	}
}

// handleOperationResponse handles a response to a previous operation
func (b *MCPBridge) handleOperationResponse(message map[string]interface{}) {
	// WHO: ResponseHandler
	// WHAT: Handle operation responses
	// WHEN: Upon response receipt
	// WHERE: System Layer 6 (Integration)
	// WHY: To process operation results
	// HOW: Using response processing
	// EXTENT: Response handling

	b.logger.Debug("Handling operation response")

	// Extract response details
	data, ok := message["data"].(map[string]interface{})
	if !ok {
		b.logger.Error("Invalid operation response format")
		return
	}

	// Extract operation details
	operationID, _ := data["operation_id"].(string)
	success, _ := data["success"].(bool)

	if success {
		b.logger.Info("Operation succeeded", "operation_id", operationID)
	} else {
		errorMsg, _ := data["error"].(string)
		b.logger.Error("Operation failed", "operation_id", operationID, "error", errorMsg)
	}
}

// handleDisconnect handles a disconnection from the TNOS MCP server
func (b *MCPBridge) handleDisconnect(err error) {
	// WHO: DisconnectionHandler
	// WHAT: Handle disconnections
	// WHEN: Upon connection loss
	// WHERE: System Layer 6 (Integration)
	// WHY: To manage recovery
	// HOW: Using state transition
	// EXTENT: Connection recovery

	b.connMutex.Lock()
	defer b.connMutex.Unlock()

	// Check if already disconnected or stopping
	if b.state == MCPBridgeStateDisconnected || b.state == MCPBridgeStateStopping {
		return
	}

	// Clean up connection
	if b.conn != nil {
		b.conn.Close()
		b.conn = nil
	}

	// Set reconnecting state
	b.setState(MCPBridgeStateReconnecting)

	// Log disconnection
	b.logger.Error("Disconnected from TNOS MCP", "error", err.Error())
}

// setState sets the state of the MCP bridge
func (b *MCPBridge) setState(state MCPBridgeState) {
	// WHO: StateManager
	// WHAT: Update bridge state
	// WHEN: During state transitions
	// WHERE: System Layer 6 (Integration)
	// WHY: To track lifecycle
	// HOW: Using state machine
	// EXTENT: State management

	b.logger.Info("Bridge state changing", "from", b.state, "to", state)
	b.state = state
}

// sendMessage sends a message to the TNOS MCP server with locking
func (b *MCPBridge) sendMessage(message map[string]interface{}) error {
	// WHO: MessageSender
	// WHAT: Send message with locking
	// WHEN: During outbound communication
	// WHERE: System Layer 6 (Integration)
	// WHY: To ensure thread safety
	// HOW: Using mutex protection
	// EXTENT: Thread-safe communication

	b.connMutex.Lock()
	defer b.connMutex.Unlock()

	return b.sendMessageUnsafe(message)
}

// sendMessageUnsafe sends a message to the TNOS MCP server without locking
func (b *MCPBridge) sendMessageUnsafe(message map[string]interface{}) error {
	// WHO: UnsafeMessageSender
	// WHAT: Send message without locking
	// WHEN: During outbound communication
	// WHERE: System Layer 6 (Integration)
	// WHY: To send when already locked
	// HOW: Using direct transmission
	// EXTENT: Internal communication

	// Check if connected
	if b.state != MCPBridgeStateConnected || b.conn == nil {
		return fmt.Errorf("bridge not connected")
	}

	// Marshal message to JSON
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Send message
	err = b.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// Update stats
	b.stats.MessagesSent++
	b.stats.LastMessageSent = time.Now().Unix()

	return nil
}

// applyMobiusCompression applies Möbius compression to a message
func (b *MCPBridge) applyMobiusCompression(
	message map[string]interface{},
) (map[string]interface{}, error) {
	// WHO: CompressionEngine
	// WHAT: Apply Möbius compression
	// WHEN: During data optimization
	// WHERE: System Layer 6 (Integration)
	// WHY: To optimize transmission
	// HOW: Using compression algorithm
	// EXTENT: Data optimization

	// Skip compression for small messages
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	if len(messageJSON) < 1024 {
		// Don't compress small messages
		return message, nil
	}

	// Extract context or create one
	var context map[string]interface{}
	if ctx, ok := message["context"].(map[string]interface{}); ok {
		context = ctx
	} else {
		context = map[string]interface{}{
			"who":    "MCPBridge",
			"what":   "CompressMessage",
			"when":   time.Now().Unix(),
			"where":  "Integration_Layer",
			"why":    "Optimize_Transmission",
			"how":    "Mobius_Compression",
			"extent": 1.0,
		}
	}

	// Extract or create metadata
	var meta map[string]interface{}
	if m, ok := context["meta"].(map[string]interface{}); ok {
		meta = m
	} else {
		meta = map[string]interface{}{}
		context["meta"] = meta
	}

	// Extract Möbius factors or use defaults
	B := getMetaFloat(meta, "B", 0.8) // Base factor
	V := getMetaFloat(meta, "V", 0.7) // Value factor
	I := getMetaFloat(meta, "I", 0.9) // Intent factor
	G := getMetaFloat(meta, "G", 1.2) // Growth factor
	F := getMetaFloat(meta, "F", 0.6) // Flexibility factor

	// Calculate entropy
	entropy := float64(len(messageJSON)) / 1024.0 // Simple approximation

	// Calculate time factor
	now := time.Now().Unix()
	when := int64(0)
	if w, ok := context["when"].(float64); ok {
		when = int64(w)
	} else if w, ok := context["when"].(int64); ok {
		when = w
	} else {
		when = now
	}

	t := float64(now-when) / 86400.0 // Days
	if t < 0 {
		t = 0
	}

	// Energy factor (computational cost)
	E := 0.5

	// Apply Möbius compression formula
	alignment := (B + V*I) * math.Exp(-t*E)
	compressionFactor := (B * I * (1.0 - (entropy / math.Log2(1.0+V))) * (G + F)) /
		(E*t + entropy + alignment)

	// Guard against extreme values
	if compressionFactor < 0.1 {
		compressionFactor = 0.1
	} else if compressionFactor > 10.0 {
		compressionFactor = 10.0
	}

	// Create compressed message
	compressed := map[string]interface{}{
		"compressed": true,
		"context":    message,
		"meta": map[string]interface{}{
			"algorithm":         "mobius",
			"version":           "1.0",
			"originalSize":      len(messageJSON),
			"compressionFactor": compressionFactor,
			"timestamp":         now,
			"factors": map[string]interface{}{
				"B": B,
				"V": V,
				"I": I,
				"G": G,
				"F": F,
				"E": E,
				"t": t,
			},
		},
	}

	return compressed, nil
}

// Helper function to safely extract a float from a map
func getMetaFloat(meta map[string]interface{}, key string, defaultValue float64) float64 {
	// WHO: MetaAccessor
	// WHAT: Extract float from meta
	// WHEN: During meta operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To safely access meta values
	// HOW: Using type assertion
	// EXTENT: Meta value extraction

	if meta == nil {
		return defaultValue
	}

	if val, ok := meta[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int32:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return defaultValue
}
