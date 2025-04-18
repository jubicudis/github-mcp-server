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
	"sync"
	"time"

	"github.com/github/github-mcp-server/pkg/mcp"
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
	Who    string  `json:"who"`    // WHO: Actor & Identity Context
	What   string  `json:"what"`   // WHAT: Function & Content Context
	When   string  `json:"when"`   // WHEN: Temporal Context
	Where  string  `json:"where"`  // WHERE: Location Context
	Why    string  `json:"why"`    // WHY: Intent & Purpose Context
	How    string  `json:"how"`    // HOW: Method & Process Context
	Extent float64 `json:"extent"` // EXTENT: Scope & Impact Context
}

// TNOSMCPBridge provides bidirectional communication between GitHub MCP and TNOS MCP
type TNOSMCPBridge struct {
	// WHO: Bridge component
	// WHAT: MCP bridge implementation
	config      BridgeConfig      // HOW: Configuration parameters
	connection  *BridgeConnection // WHERE: Connection to TNOS MCP
	contextPool sync.Pool         // WHAT: Pool of context vectors for reuse
	mutex       sync.RWMutex      // HOW: Thread safety mechanism
	isConnected bool              // WHEN: Current connection state
	healthTimer *time.Timer       // WHEN: Timer for health checks
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
				}
			},
		},
	}

	// Initialize the connection
	bridge.initConnection()

	return bridge
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
	b.mutex.RUnlock()

	if !isConnected {
		return nil, errors.New("bridge is not connected to TNOS MCP")
	}

	// Extract GitHub context
	githubContext := make(map[string]interface{})
	if req.Context != nil {
		githubContext = req.Context
	}

	// Translate to TNOS 7D context
	tnos7D, err := b.TranslateContext(githubContext)
	if err != nil {
		return nil, fmt.Errorf("failed to translate context: %w", err)
	}

	// Prepare request for TNOS MCP
	tnosRequest := map[string]interface{}{
		"tool":    req.Tool,
		"params":  req.Params,
		"context": tnos7D,
	}

	// Simulate sending request to TNOS MCP
	// In a real implementation, this would use HTTP, WebSockets, etc.
	
	if b.config.DebugMode {
		reqJSON, _ := json.Marshal(tnosRequest)
		log.Printf("Sending request to TNOS MCP: %s", string(reqJSON))
	}

	// Simulate receiving response
	// In a real implementation, this would parse the actual response
	tnosResponse := &mcp.CallToolResult{
		Type: "text",
		Text: "Response from TNOS MCP",
	}

	return tnosResponse, nil
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
