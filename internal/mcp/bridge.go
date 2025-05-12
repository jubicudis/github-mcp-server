// WHO: InternalBridge
// WHAT: Internal MCP bridge implementation for GitHub MCP server
// WHEN: During MCP request processing
// WHERE: Bridge between GitHub MCP and TNOS MCP
// WHY: To enable bidirectional communication with TNOS
// HOW: Using Go with 7D context translation
// EXTENT: All bridge operations for GitHub MCP

package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// BridgeConfig contains configuration options for the TNOS MCP Bridge
type BridgeConfig struct {
	// WHO: Configuration component
	// WHAT: Bridge configuration parameters
	Endpoint          string        // WHERE: TNOS MCP endpoint URL
	Timeout           time.Duration // WHEN: Connection timeout duration
	ReconnectInterval time.Duration // WHEN: Time between reconnection attempts
	MaxRetries        int           // EXTENT: Maximum number of retry attempts
	DebugMode         bool          // HOW: Enable detailed logging
}

// ContextVector7D represents the 7D context framework used in TNOS
type ContextVector7D struct {
	// WHO: Context vector component
	// WHAT: 7D context representation
	Who    string                 `json:"who"`              // WHO: Actor & Identity Context
	What   string                 `json:"what"`             // WHAT: Function & Content Context
	When   string                 `json:"when"`             // WHEN: Temporal Context
	Where  string                 `json:"where"`            // WHERE: Location Context
	Why    string                 `json:"why"`              // WHY: Intent & Purpose Context
	How    string                 `json:"how"`              // HOW: Method & Process Context
	Extent float64                `json:"extent"`           // EXTENT: Scope & Impact Context
	Meta   map[string]interface{} `json:"meta,omitempty"`   // Additional context metadata for compression
	Source string                 `json:"source,omitempty"` // Source system identifier
}

// Compress applies the Möbius Compression Formula to this context vector
func (cv *ContextVector7D) Compress() *ContextVector7D {
	// WHO: ContextCompressor
	// WHAT: Context compression
	// WHEN: During context transmission
	// WHERE: Between systems
	// WHY: To optimize context storage and transmission
	// HOW: Using Möbius compression formula
	// EXTENT: Full context representation

	if cv.Meta == nil {
		cv.Meta = make(map[string]interface{})
	}

	// Extract or set default contextual factors
	B := getMetaFloat(cv.Meta, "B", 0.8) // Base factor
	V := getMetaFloat(cv.Meta, "V", 0.7) // Value factor
	I := getMetaFloat(cv.Meta, "I", 0.9) // Intent factor
	G := getMetaFloat(cv.Meta, "G", 1.2) // Growth factor
	F := getMetaFloat(cv.Meta, "F", 0.6) // Flexibility factor

	// Calculate time factor (how "fresh" the context is)
	now := time.Now().Format(time.RFC3339)
	t := calculateTimeDelta(cv.When, now)

	// Calculate energy factor (computational cost)
	E := 0.5

	// Calculate context entropy (simplified)
	contextBytes, _ := json.Marshal(cv)
	entropy := float64(len(contextBytes)) / 100.0

	// Apply Möbius compression formula
	alignment := (B + V*I) * math.Exp(-t*E)
	compressionFactor := (B * I * (1.0 - (entropy / math.Log2(1.0+V))) * (G + F)) /
		(E*t + entropy + alignment)

	// Store compression metadata
	cv.Meta["compressedAt"] = now
	cv.Meta["compressionFactor"] = compressionFactor
	cv.Meta["entropy"] = entropy
	cv.Meta["B"] = B
	cv.Meta["V"] = V
	cv.Meta["I"] = I
	cv.Meta["G"] = G
	cv.Meta["F"] = F
	cv.Meta["E"] = E
	cv.Meta["t"] = t
	cv.Meta["alignment"] = alignment

	return cv
}

// Merge combines this context vector with another, preserving important information
func (cv ContextVector7D) Merge(other ContextVector7D) ContextVector7D {
	// WHO: ContextMerger
	// WHAT: Context combination
	// WHEN: During multi-source operations
	// WHERE: Context integration points
	// WHY: To combine information from multiple contexts
	// HOW: Using weighted merging strategy
	// EXTENT: Comprehensive context integration

	result := cv

	// Use the more recent When value
	t1 := calculateTimeDelta("", cv.When)
	t2 := calculateTimeDelta("", other.When)
	if t2 < t1 {
		result.When = other.When
	}

	// Merge metadata
	if result.Meta == nil {
		result.Meta = make(map[string]interface{})
	}

	if other.Meta != nil {
		for k, v := range other.Meta {
			if _, exists := result.Meta[k]; !exists {
				result.Meta[k] = v
			}
		}
	}

	// Keep track of sources
	if result.Source == "" {
		result.Source = "merged"
	} else if other.Source != "" && result.Source != other.Source {
		result.Source = result.Source + "+" + other.Source
	}

	return result
}

// ToMap converts the context vector to a map representation
func (cv *ContextVector7D) ToMap() map[string]interface{} {
	// WHO: DataTransformer
	// WHAT: Format conversion
	// WHEN: During serialization
	// WHERE: Data exchange boundaries
	// WHY: To enable interoperability
	// HOW: Using standardized mapping
	// EXTENT: Complete context representation

	return map[string]interface{}{
		"who":    cv.Who,
		"what":   cv.What,
		"when":   cv.When,
		"where":  cv.Where,
		"why":    cv.Why,
		"how":    cv.How,
		"extent": cv.Extent,
		"meta":   cv.Meta,
		"source": cv.Source,
	}
}

// TranslateToMCP converts the context vector to MCP format
func (cv *ContextVector7D) TranslateToMCP() map[string]interface{} {
	// WHO: FormatTranslator
	// WHAT: Protocol format translation
	// WHEN: During protocol interchange
	// WHERE: Bridge communication
	// WHY: For protocol compatibility
	// HOW: Using format mapping
	// EXTENT: All required MCP fields

	result := map[string]interface{}{
		"identity":  cv.Who,
		"operation": cv.What,
		"timestamp": cv.When,
		"location":  cv.Where,
		"purpose":   cv.Why,
		"method":    cv.How,
		"scope":     cv.Extent,
	}

	if cv.Meta != nil {
		result["metadata"] = cv.Meta
	}

	if cv.Source != "" {
		result["source"] = cv.Source
	}

	return result
}

// TNOSMCPBridge provides bidirectional communication between GitHub MCP and TNOS MCP
type TNOSMCPBridge struct {
	// WHO: Bridge component
	// WHAT: MCP bridge implementation
	config            BridgeConfig                                  // HOW: Configuration parameters
	connection        *BridgeConnection                             // WHERE: Connection to TNOS MCP
	contextPool       sync.Pool                                     // WHAT: Pool of context vectors for reuse
	mutex             sync.RWMutex                                  // HOW: Thread safety mechanism
	isConnected       bool                                          // WHEN: Current connection state
	healthTimer       *time.Timer                                   // WHEN: Timer for health checks
	protocolVersion   ProtocolVersion                               // WHAT: Negotiated protocol version
	supportedVersions []ProtocolVersion                             // WHAT: Supported protocol versions
	serverFeatures    map[string]interface{}                        // WHAT: Server-advertised feature set
	contextStorage    map[string]ContextVector7D                    // WHERE: Persistent context storage
	messageHandlers   map[string]func(map[string]interface{}) error // HOW: Message type handlers
}

// BridgeConnection manages the underlying connection to TNOS MCP
type BridgeConnection struct {
	// WHO: Connection component
	// WHAT: Connection management
	ctx        context.Context    // WHEN: Connection context
	cancelFunc context.CancelFunc // HOW: Function to cancel context
	// Additional connection fields would be implemented here
}

// NewTNOSMCPBridge creates a new MCP bridge with the specified configuration
func NewTNOSMCPBridge(config BridgeConfig) *TNOSMCPBridge {
	// WHO: Bridge factory
	// WHAT: Bridge instance creation
	// WHEN: During server initialization
	// WHERE: GitHub MCP Server startup
	// WHY: To establish connection with TNOS MCP
	// HOW: Using provided configuration
	// EXTENT: Single bridge instance

	// Use default values if not specified
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.ReconnectInterval == 0 {
		config.ReconnectInterval = 5 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 5
	}

	// Default supported versions
	supportedVersions := []ProtocolVersion{
		{Major: 3, Minor: 0, Patch: 0},
		{Major: 2, Minor: 0, Patch: 0},
		{Major: 1, Minor: 0, Patch: 0},
	}

	// Create and initialize the bridge
	bridge := &TNOSMCPBridge{
		config: config,
		contextPool: sync.Pool{
			New: func() interface{} {
				return &ContextVector7D{
					Who:    "System",
					What:   "Operation",
					When:   time.Now().Format(time.RFC3339),
					Where:  "MCP_Bridge",
					Why:    "Protocol_Communication",
					How:    "Bridge_Connection",
					Extent: 1.0,
					Meta:   make(map[string]interface{}),
					Source: "github_mcp_bridge",
				}
			},
		},
		supportedVersions: supportedVersions,
		protocolVersion:   supportedVersions[0], // Default to highest supported version
		serverFeatures:    make(map[string]interface{}),
		contextStorage:    make(map[string]ContextVector7D),
		messageHandlers:   make(map[string]func(map[string]interface{}) error),
	}

	// Register default message handlers
	bridge.registerDefaultHandlers()

	// Initialize the connection
	bridge.initConnection()

	return bridge
}

// registerDefaultHandlers registers the default message type handlers
func (b *TNOSMCPBridge) registerDefaultHandlers() {
	// WHO: HandlerRegistrar
	// WHAT: Message handler registration
	// WHEN: During bridge initialization
	// WHERE: Message processing setup
	// WHY: To handle different message types
	// HOW: Using function mapping
	// EXTENT: All supported message types

	b.messageHandlers["handshake"] = b.handleHandshake
	b.messageHandlers["healthcheck"] = b.handleHealthCheck
	b.messageHandlers["context_sync"] = b.handleContextSync
	b.messageHandlers["error"] = b.handleError
}

// handleHandshake processes handshake messages
func (b *TNOSMCPBridge) handleHandshake(message map[string]interface{}) error {
	// WHO: HandshakeProcessor
	// WHAT: Protocol handshake handling
	// WHEN: During connection establishment
	// WHERE: Protocol initialization
	// WHY: To negotiate protocol parameters
	// HOW: Using version comparison
	// EXTENT: Complete handshake process

	// Extract supported versions from the message
	versions := []string{"1.0"}
	if versionsArray, ok := message["supportedVersions"].([]interface{}); ok {
		versions = make([]string, len(versionsArray))
		for i, v := range versionsArray {
			if strVersion, ok := v.(string); ok {
				versions[i] = strVersion
			}
		}
	}

	// Find highest compatible version
	serverVersions := make([]ProtocolVersion, 0)
	for _, vStr := range versions {
		if v, err := NewProtocolVersion(vStr); err == nil {
			serverVersions = append(serverVersions, v)
		}
	}

	// Find highest compatible version
	negotiatedVersion, err := b.negotiateVersion(serverVersions)
	if err != nil {
		return fmt.Errorf("version negotiation failed: %w", err)
	}

	// Update bridge state
	b.mutex.Lock()
	b.protocolVersion = negotiatedVersion

	// Extract server features if available
	if features, ok := message["serverFeatures"].(map[string]interface{}); ok {
		b.serverFeatures = features
	}
	b.mutex.Unlock()

	// Log successful handshake
	if b.config.DebugMode {
		log.Printf("Handshake successful, using protocol version %s", negotiatedVersion.String())
	}

	return nil
}

// handleHealthCheck processes health check messages
func (b *TNOSMCPBridge) handleHealthCheck(message map[string]interface{}) error {
	// WHO: HealthcheckProcessor
	// WHAT: Health verification
	// WHEN: During health monitoring
	// WHERE: Bridge connection
	// WHY: To confirm connection health
	// HOW: Using status response
	// EXTENT: Single health check

	// Update last activity timestamp
	timestamp := time.Now().Format(time.RFC3339)

	if b.config.DebugMode {
		log.Printf("Health check received at %s", timestamp)
	}

	// Could implement additional health metrics here

	return nil
}

// handleContextSync processes context synchronization messages
func (b *TNOSMCPBridge) handleContextSync(message map[string]interface{}) error {
	// WHO: ContextSynchronizer
	// WHAT: Context state synchronization
	// WHEN: During context updates
	// WHERE: Bridge context storage
	// WHY: To maintain context continuity
	// HOW: Using context storage
	// EXTENT: Shared context state

	// Extract context from the message
	contextMap, ok := message["context"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid context in sync message")
	}

	// Convert to 7D context
	context := TranslateFromMCP(contextMap)

	// Store or update the context
	id := extractStringWithDefault(message, "id", "default")

	b.mutex.Lock()
	b.contextStorage[id] = context
	b.mutex.Unlock()

	if b.config.DebugMode {
		log.Printf("Context synchronized: %s", id)
	}

	return nil
}

// handleError processes error messages
func (b *TNOSMCPBridge) handleError(message map[string]interface{}) error {
	// WHO: ErrorHandler
	// WHAT: Error processing
	// WHEN: During error conditions
	// WHERE: Bridge communication
	// WHY: To handle error states
	// HOW: Using error reporting
	// EXTENT: Error recovery

	errorCode := extractStringWithDefault(message, "code", "unknown")
	errorMsg := extractStringWithDefault(message, "message", "Unknown error")

	log.Printf("Error received from TNOS MCP: %s - %s", errorCode, errorMsg)

	// Implement error-specific recovery logic here

	return fmt.Errorf("mcp error: [%s] %s", errorCode, errorMsg)
}

// negotiateVersion finds the highest compatible version between client and server
func (b *TNOSMCPBridge) negotiateVersion(serverVersions []ProtocolVersion) (ProtocolVersion, error) {
	// WHO: VersionNegotiator
	// WHAT: Protocol version selection
	// WHEN: During handshake
	// WHERE: Protocol initialization
	// WHY: To ensure compatible communication
	// HOW: Using version matching
	// EXTENT: Protocol version compatibility

	// If no server versions, return error
	if len(serverVersions) == 0 {
		return ProtocolVersion{}, errors.New("no compatible protocol version")
	}

	// Try to find highest compatible version
	for _, clientVersion := range b.supportedVersions {
		for _, serverVersion := range serverVersions {
			if clientVersion.Major == serverVersion.Major &&
				clientVersion.Minor >= serverVersion.Minor {
				return clientVersion, nil
			}
		}
	}

	// No compatible version found
	return ProtocolVersion{}, errors.New("no compatible protocol version found")
}

// GetPersistentContext retrieves a stored context by ID
func (b *TNOSMCPBridge) GetPersistentContext(id string) (*ContextVector7D, bool) {
	// WHO: ContextProvider
	// WHAT: Context retrieval
	// WHEN: During context operations
	// WHERE: Context storage
	// WHY: To access persistent context
	// HOW: Using context lookup
	// EXTENT: Single context instance

	b.mutex.RLock()
	defer b.mutex.RUnlock()

	context, exists := b.contextStorage[id]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid concurrent modification
	contextCopy := context
	return &contextCopy, true
}

// StorePersistentContext saves a context with the given ID
func (b *TNOSMCPBridge) StorePersistentContext(id string, context ContextVector7D) {
	// WHO: ContextStorer
	// WHAT: Context persistence
	// WHEN: During context updates
	// WHERE: Context storage
	// WHY: To maintain context state
	// HOW: Using context mapping
	// EXTENT: Single context instance

	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.contextStorage[id] = context
}

// initConnection establishes the initial connection to TNOS MCP
func (b *TNOSMCPBridge) initConnection() {
	// WHO: Bridge component
	// WHAT: Connection initialization
	// WHEN: During bridge creation
	// WHERE: Connection establishment
	// WHY: To prepare for communication
	// HOW: Using context and cancellation
	// EXTENT: Single connection instance

	b.mutex.Lock()
	defer b.mutex.Unlock()

	ctx, cancelFunc := context.WithCancel(context.Background())
	b.connection = &BridgeConnection{
		ctx:        ctx,
		cancelFunc: cancelFunc,
	}

	// Start health check timer
	b.startHealthCheck()

	// Attempt initial connection
	go b.connectWithRetry()
}

// startHealthCheck initiates periodic health checks for the bridge connection
func (b *TNOSMCPBridge) startHealthCheck() {
	// WHO: Health monitor component
	// WHAT: Connection health monitoring
	// WHEN: Periodic intervals
	// WHERE: Bridge connection
	// WHY: To ensure connection stability
	// HOW: Using timer-based checks
	// EXTENT: Continuous monitoring

	if b.healthTimer != nil {
		b.healthTimer.Stop()
	}

	b.healthTimer = time.AfterFunc(30*time.Second, func() {
		if err := b.performHealthCheck(); err != nil {
			log.Printf("Health check failed: %v", err)
			// Attempt reconnection
			go b.connectWithRetry()
		}
		// Schedule next health check
		b.startHealthCheck()
	})
}

// performHealthCheck verifies the bridge connection is functioning
func (b *TNOSMCPBridge) performHealthCheck() error {
	// WHO: Health check component
	// WHAT: Connection validation
	// WHEN: Scheduled health check
	// WHERE: Bridge connection
	// WHY: To verify connection status
	// HOW: Using ping mechanism
	// EXTENT: Single connection verification

	b.mutex.RLock()
	isConnected := b.isConnected
	b.mutex.RUnlock()

	if !isConnected {
		return errors.New("bridge is not connected")
	}

	// Perform actual health check (ping)
	// Implementation would depend on the actual protocol
	return nil
}

// connectWithRetry attempts to connect to TNOS MCP with retry logic
func (b *TNOSMCPBridge) connectWithRetry() {
	// WHO: Connection manager
	// WHAT: Connection establishment with retry
	// WHEN: After connection failure
	// WHERE: Bridge initialization or recovery
	// WHY: To establish reliable connection
	// HOW: Using exponential backoff
	// EXTENT: Until connected or max retries

	retries := 0
	for retries < b.config.MaxRetries {
		if err := b.connect(); err == nil {
			// Successfully connected
			return
		}

		retries++
		if b.config.DebugMode {
			log.Printf("Connection attempt %d/%d failed, retrying in %v",
				retries, b.config.MaxRetries, b.config.ReconnectInterval)
		}

		// Wait before retrying
		time.Sleep(b.config.ReconnectInterval)
	}

	if b.config.DebugMode {
		log.Printf("Failed to connect after %d attempts", retries)
	}
}

// connect establishes a connection to TNOS MCP
func (b *TNOSMCPBridge) connect() error {
	// WHO: Connection component
	// WHAT: Connection establishment
	// WHEN: During initialization or reconnection
	// WHERE: Bridge to TNOS MCP
	// WHY: To establish communication channel
	// HOW: Using configured endpoint
	// EXTENT: Single connection attempt

	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Implement actual connection logic here
	// This would depend on the actual protocol used

	// For now, simulate a successful connection
	b.isConnected = true

	if b.config.DebugMode {
		log.Printf("Connected to TNOS MCP at %s", b.config.Endpoint)
	}

	return nil
}

// Disconnect closes the bridge connection
func (b *TNOSMCPBridge) Disconnect() error {
	// WHO: Connection manager
	// WHAT: Connection termination
	// WHEN: During shutdown
	// WHERE: Bridge connection
	// WHY: To cleanly terminate connection
	// HOW: Using context cancellation
	// EXTENT: Complete disconnection

	b.mutex.Lock()
	defer b.mutex.Unlock()

	if !b.isConnected {
		return nil
	}

	// Stop health check timer
	if b.healthTimer != nil {
		b.healthTimer.Stop()
		b.healthTimer = nil
	}

	// Cancel the connection context
	if b.connection != nil && b.connection.cancelFunc != nil {
		b.connection.cancelFunc()
	}

	b.isConnected = false

	if b.config.DebugMode {
		log.Println("Disconnected from TNOS MCP")
	}

	return nil
}

// TranslateContext converts between GitHub MCP context and TNOS 7D context
func (b *TNOSMCPBridge) TranslateContext(githubContext map[string]interface{}) (*ContextVector7D, error) {
	// WHO: Context translator
	// WHAT: Context format translation
	// WHEN: During request processing
	// WHERE: Between GitHub and TNOS
	// WHY: To maintain 7D context integrity
	// HOW: Using mapping rules
	// EXTENT: Complete context translation

	// Get a context vector from the pool
	tnos7D := b.contextPool.Get().(*ContextVector7D)

	// Extract values from GitHub context with defaults
	tnos7D.Who = extractStringWithDefault(githubContext, "identity", "System")
	tnos7D.What = extractStringWithDefault(githubContext, "operation", "Process")
	tnos7D.When = extractStringWithDefault(githubContext, "timestamp",
		time.Now().Format(time.RFC3339))
	tnos7D.Where = "GitHub_MCP_Bridge"
	tnos7D.Why = extractStringWithDefault(githubContext, "purpose", "Protocol_Compliance")
	tnos7D.How = "Bridge_Translation"
	tnos7D.Extent = extractFloatWithDefault(githubContext, "scope", 1.0)

	return tnos7D, nil
}

// TranslateContextBack converts TNOS 7D context back to GitHub MCP context
func (b *TNOSMCPBridge) TranslateContextBack(tnos7D *ContextVector7D) (map[string]interface{}, error) {
	// WHO: Context translator
	// WHAT: Reverse context translation
	// WHEN: During response processing
	// WHERE: Between TNOS and GitHub
	// WHY: To present standardized context
	// HOW: Using reverse mapping
	// EXTENT: Complete context translation

	githubContext := make(map[string]interface{})

	// Convert 7D context to GitHub context
	githubContext["identity"] = tnos7D.Who
	githubContext["operation"] = tnos7D.What
	githubContext["timestamp"] = tnos7D.When
	githubContext["location"] = tnos7D.Where
	githubContext["purpose"] = tnos7D.Why
	githubContext["method"] = tnos7D.How
	githubContext["scope"] = tnos7D.Extent

	// Return context to pool for reuse
	b.contextPool.Put(tnos7D)

	return githubContext, nil
}

// prepareContext handles retrieving or creating the context for a request
func (b *TNOSMCPBridge) prepareContext(githubContext map[string]interface{}) (*ContextVector7D, string, error) {
	// WHO: Context preparer
	// WHAT: Context preparation
	// WHEN: During request setup
	// WHERE: Bridge context handling
	// WHY: To establish request context
	// HOW: Using existing or new context
	// EXTENT: Single context preparation

	// Check for persistent context ID
	contextID := extractStringWithDefault(githubContext, "contextId", "")
	var tnos7D *ContextVector7D
	var err error

	if contextID != "" {
		// Try to retrieve existing context
		if existingContext, exists := b.GetPersistentContext(contextID); exists {
			tnos7D = existingContext
			// Update with any new values from the request
			if tnos7D.Meta == nil {
				tnos7D.Meta = make(map[string]interface{})
			}
			tnos7D.Meta["lastUsed"] = time.Now().Format(time.RFC3339)
			return tnos7D, contextID, nil
		}
	}

	// If no existing context, translate from GitHub context
	tnos7D, err = b.TranslateContext(githubContext)
	if err != nil {
		return nil, contextID, fmt.Errorf("failed to translate context: %w", err)
	}

	return tnos7D, contextID, nil
}

// createTNOSRequest builds the request based on protocol version
func (b *TNOSMCPBridge) createTNOSRequest(
	protocolVersion ProtocolVersion,
	req mcp.CallToolRequest,
	tnos7D *ContextVector7D,
) map[string]interface{} {
	// WHO: Request builder
	// WHAT: Protocol-specific request creation
	// WHEN: Before request transmission
	// WHERE: Bridge request preparation
	// WHY: To create protocol-compliant request
	// HOW: Using version-based format selection
	// EXTENT: Complete request structure

	switch protocolVersion.Major {
	case 3:
		// MCP 3.0 format
		return map[string]interface{}{
			"tool":    req.Params.Name,
			"params":  req.Params.Arguments,
			"context": tnos7D.ToMap(),
			"meta": map[string]interface{}{
				"version":   protocolVersion.String(),
				"requestId": generateRequestID(),
				"timestamp": time.Now().Format(time.RFC3339),
				"source":    "github_mcp_server",
			},
		}
	case 2:
		// MCP 2.0 format
		return map[string]interface{}{
			"tool":       req.Params.Name,
			"parameters": req.Params.Arguments,
			"context":    tnos7D.TranslateToMCP(),
			"requestId":  generateRequestID(),
		}
	default:
		// MCP 1.0 format (fallback)
		return map[string]interface{}{
			"tool":    req.Params.Name,
			"params":  req.Params.Arguments,
			"context": tnos7D.TranslateToMCP(),
		}
	}
}

// processResponse handles the response from TNOS MCP
func (b *TNOSMCPBridge) processResponse(
	response *mcp.CallToolResult,
	tnos7D *ContextVector7D,
	protocolVersion ProtocolVersion,
) {
	// WHO: Response processor
	// WHAT: Response handling
	// WHEN: After request completion
	// WHERE: Bridge response handling
	// WHY: To process and enrich response
	// HOW: Using protocol-specific format
	// EXTENT: Complete response processing

	// Add context information to the response based on protocol version
	if protocolVersion.Major >= 2 {
		// For MCP 2.0+, include context in the response
		contextJSON, _ := json.Marshal(tnos7D.ToMap())

		// The actual MCP library might store context in a different way
		// Here we're assuming it might use a Metadata field
		if metadataField := reflect.ValueOf(response).Elem().FieldByName("Metadata"); metadataField.IsValid() && metadataField.CanSet() {
			metadata := map[string]interface{}{
				"context": string(contextJSON),
			}
			metadataField.Set(reflect.ValueOf(metadata))
		}
	}
}

// SendRequest sends a request to TNOS MCP and returns the response
func (b *TNOSMCPBridge) SendRequest(req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// WHO: Request handler
	// WHAT: Request transmission
	// WHEN: During tool invocation
	// WHERE: From GitHub MCP to TNOS MCP
	// WHY: To execute tool operations
	// HOW: Using context translation
	// EXTENT: Complete request lifecycle

	b.mutex.RLock()
	isConnected := b.isConnected
	protocolVersion := b.protocolVersion
	b.mutex.RUnlock()

	if !isConnected {
		return nil, errors.New("bridge is not connected to TNOS MCP")
	}

	// Extract GitHub context
	githubContext := make(map[string]interface{})
	if contextValue, ok := req.Params.Arguments["context"].(map[string]interface{}); ok {
		githubContext = contextValue
	}

	// Prepare context for the request
	tnos7D, contextID, err := b.prepareContext(githubContext)
	if err != nil {
		return nil, err
	}

	// Apply compression to optimize context
	tnos7D.Compress()

	// Store context for future use if ID provided
	if contextID != "" {
		b.StorePersistentContext(contextID, *tnos7D)
	}

	// Prepare request for TNOS MCP based on protocol version
	tnosRequest := b.createTNOSRequest(protocolVersion, req, tnos7D)

	// Log the request with timing
	startTime := time.Now()
	if b.config.DebugMode {
		reqJSON, _ := json.Marshal(tnosRequest)
		log.Printf("[MCP-BRIDGE] Sending request to TNOS MCP: %s", string(reqJSON))
	}

	// Simulate sending request to TNOS MCP
	// In a real implementation, this would use HTTP, WebSockets, etc.
	time.Sleep(10 * time.Millisecond)

	// Create response
	tnosResponse := &mcp.CallToolResult{}
	responseJSON := []byte(`{"result":"Response from TNOS MCP"}`)
	if err := json.Unmarshal(responseJSON, tnosResponse); err != nil && b.config.DebugMode {
		log.Printf("[MCP-BRIDGE] Warning: Failed to construct response: %v", err)
	}

	// Process and enrich the response
	b.processResponse(tnosResponse, tnos7D, protocolVersion)

	// Log performance metrics
	if b.config.DebugMode {
		duration := time.Since(startTime)
		log.Printf("[MCP-BRIDGE] Request completed in %v", duration)
	}

	return tnosResponse, nil
}

// generateRequestID creates a unique request ID
func generateRequestID() string {
	// WHO: IDGenerator
	// WHAT: Request ID generation
	// WHEN: During request preparation
	// WHERE: Bridge communication
	// WHY: To uniquely identify requests
	// HOW: Using timestamp and random component
	// EXTENT: Single request identification

	now := time.Now().UnixNano()
	return fmt.Sprintf("req_%d_%d", now, now%1000)
}

// RegisterTools registers tools from TNOS MCP with GitHub MCP Server
func (b *TNOSMCPBridge) RegisterTools() []mcp.Tool {
	// WHO: Tool registrar
	// WHAT: Tool registration
	// WHEN: During server initialization
	// WHERE: Bridge integration point
	// WHY: To expose TNOS capabilities
	// HOW: Using tool definitions
	// EXTENT: All TNOS-specific tools

	// These would be retrieved from TNOS MCP in a real implementation
	tools := []mcp.Tool{
		mcp.NewTool(
			"formula_executor",
			mcp.WithDescription("Execute formulas from Formula Registry"),
			mcp.WithString("formula", mcp.Description("Formula identifier or expression")),
			mcp.WithObject("parameters", mcp.Description("Formula parameters")),
		),
		mcp.NewTool(
			"context_query",
			mcp.WithDescription("Query the 7D context system"),
			mcp.WithString("dimension", mcp.Description("Context dimension to query (who, what, when, where, why, how, extent)")),
			mcp.WithString("query", mcp.Description("Query string")),
		),
		mcp.NewTool(
			"mcp_bridge",
			mcp.WithDescription("Direct messaging to TNOS MCP"),
			mcp.WithString("message", mcp.Description("Message content")),
			mcp.WithObject("context", mcp.Description("Context for message")),
		),
		mcp.NewTool(
			"mobius_compression",
			mcp.WithDescription("Compress data using TNOS framework"),
			mcp.WithString("data", mcp.Description("Data to compress")),
			mcp.WithObject("context", mcp.Description("Compression context")),
		),
		mcp.NewTool(
			"context_sync",
			mcp.WithDescription("Synchronize context between GitHub MCP and TNOS"),
			mcp.WithObject("context", mcp.Description("Context to synchronize")),
		),
	}

	return tools
}

// Helper functions for context translation

// extractStringWithDefault extracts a string value from a map with a default fallback
func extractStringWithDefault(m map[string]interface{}, key, defaultVal string) string {
	if val, ok := m[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultVal
}

// extractFloatWithDefault extracts a float value from a map with a default fallback
func extractFloatWithDefault(m map[string]interface{}, key string, defaultVal float64) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		}
	}
	return defaultVal
}

// calculateTimeDelta calculates the time difference between two RFC3339 timestamps in days
func calculateTimeDelta(timestamp1, timestamp2 string) float64 {
	// WHO: TimeCalculator
	// WHAT: Time difference calculation
	// WHEN: During context operations
	// WHERE: Context time comparison
	// WHY: To evaluate context freshness
	// HOW: Using time parsing and comparison
	// EXTENT: Single time comparison

	// If timestamp1 is empty, use current time
	if timestamp1 == "" {
		timestamp1 = time.Now().Format(time.RFC3339)
	}

	// If timestamp2 is empty, use current time
	if timestamp2 == "" {
		timestamp2 = time.Now().Format(time.RFC3339)
	}

	// Parse the timestamps
	t1, err1 := time.Parse(time.RFC3339, timestamp1)
	t2, err2 := time.Parse(time.RFC3339, timestamp2)

	// Default to 0 if parsing fails
	if err1 != nil || err2 != nil {
		return 0
	}

	// Calculate difference in days
	diff := t2.Sub(t1).Hours() / 24.0
	if diff < 0 {
		diff = -diff // Absolute value
	}

	return diff
}

// getMetaFloat extracts a float value from metadata with a default value
func getMetaFloat(meta map[string]interface{}, key string, defaultValue float64) float64 {
	// WHO: MetadataExtractor
	// WHAT: Extract float value from metadata
	// WHEN: During context operations
	// WHERE: Context metadata processing
	// WHY: To retrieve compression factors
	// HOW: Using type assertion
	// EXTENT: Single metadata field

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
		case int64:
			return float64(v)
		}
	}

	return defaultValue
}

// TranslateFromMCP converts an MCP context to a 7D context vector
func TranslateFromMCP(mcpContext map[string]interface{}) ContextVector7D {
	// WHO: ContextTranslator
	// WHAT: Context format conversion
	// WHEN: During context ingestion
	// WHERE: Protocol boundary
	// WHY: To standardize external context
	// HOW: Using field mapping
	// EXTENT: Complete context structure

	// Create a new context vector
	cv := ContextVector7D{
		Who:    extractStringWithDefault(mcpContext, "identity", "External"),
		What:   extractStringWithDefault(mcpContext, "operation", "Unknown"),
		When:   extractStringWithDefault(mcpContext, "timestamp", time.Now().Format(time.RFC3339)),
		Where:  extractStringWithDefault(mcpContext, "location", "MCP_Interface"),
		Why:    extractStringWithDefault(mcpContext, "purpose", "External_Request"),
		How:    extractStringWithDefault(mcpContext, "method", "MCP_Protocol"),
		Extent: extractFloatWithDefault(mcpContext, "scope", 1.0),
		Source: extractStringWithDefault(mcpContext, "source", "mcp_external"),
	}

	// Extract metadata if available
	if metadata, ok := mcpContext["metadata"].(map[string]interface{}); ok {
		cv.Meta = metadata
	} else {
		cv.Meta = make(map[string]interface{})
	}

	return cv
}

// ProtocolVersion represents an MCP protocol version
type ProtocolVersion struct {
	// WHO: VersionManager
	// WHAT: Protocol version representation
	// WHEN: During protocol negotiation
	// WHERE: Bridge communication
	// WHY: To ensure compatibility
	// HOW: Using semantic versioning
	// EXTENT: Protocol version tracking

	Major int
	Minor int
	Patch int
}

// NewProtocolVersion creates a new protocol version from a string
func NewProtocolVersion(version string) (ProtocolVersion, error) {
	// WHO: VersionCreator
	// WHAT: Protocol version parsing
	// WHEN: During protocol initialization
	// WHERE: Version handling
	// WHY: To parse version strings
	// HOW: Using format parsing
	// EXTENT: Single version instance

	var major, minor, patch int
	_, err := fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &patch)
	if err != nil {
		// Try parsing just major.minor
		_, err = fmt.Sscanf(version, "%d.%d", &major, &minor)
		if err != nil {
			return ProtocolVersion{}, fmt.Errorf("invalid version format: %s", version)
		}
	}

	return ProtocolVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

// String returns the string representation of a protocol version
func (pv ProtocolVersion) String() string {
	// WHO: VersionFormatter
	// WHAT: Version string creation
	// WHEN: During protocol communication
	// WHERE: Version exchange
	// WHY: To standardize version format
	// HOW: Using string formatting
	// EXTENT: Version representation

	return fmt.Sprintf("%d.%d.%d", pv.Major, pv.Minor, pv.Patch)
}

// IsCompatibleWith checks if this version is compatible with another
func (pv ProtocolVersion) IsCompatibleWith(other ProtocolVersion) bool {
	// WHO: CompatibilityChecker
	// WHAT: Version compatibility check
	// WHEN: During protocol negotiation
	// WHERE: Bridge communication
	// WHY: To ensure protocol compatibility
	// HOW: Using semantic version rules
	// EXTENT: Version compatibility

	// Major version must match for compatibility
	return pv.Major == other.Major
}
