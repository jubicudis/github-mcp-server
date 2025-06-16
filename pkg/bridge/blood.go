/*
 * WHO: BloodCirculationBridge
 * WHAT: Bidirectional data flow between GitHub MCP and TNOS MCP
 * WHEN: During inter-system communication
 * WHERE: System Layer 6 (Integration)
 * WHY: To facilitate seamless data exchange via biological mimicry
 * HOW: Using blood as a metaphor for data circulation
 * EXTENT: All cross-system data transfer
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

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/common"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/translations"

	"github.com/gorilla/websocket"
)

// BloodConstants defines the constants specific to the blood bridge
const (
	// Ports for various MCP servers, aligned with QHP (Quantum Handshake Protocol)
	TNOSMCPPort    = 9001
	BridgePort     = 10619
	GitHubMCPPort  = 10617
	CopilotLLMPort = 8083

	// Blood flow parameters
	MaxCellCapacity      = 1024 * 1024 // Maximum size of a blood cell (1MB)
	CirculationInterval  = 5 * time.Second
	OxygenationTimeout   = 10 * time.Second
	ClottingThreshold    = 3 // Number of failed attempts before clotting (temporary disconnection)
	HemofluxCompression  = true
	DefaultPulseDuration = 100 * time.Millisecond
)

// BloodCell represents a unit of data in the blood circulation system,
// metaphorically representing how data flows between systems
type BloodCell struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"` // "Red" for data, "White" for control, "Platelet" for recovery
	Payload       map[string]interface{} `json:"payload"`
	Context7D     translations.ContextVector7D `json:"context_7d"`
	Timestamp     int64                  `json:"timestamp"`
	Source        string                 `json:"source"`
	Destination   string                 `json:"destination"`
	Priority      int                    `json:"priority"` // 1-10, with 10 being highest
	OxygenLevel   float64                `json:"oxygen_level"` // 0.0-1.0, representing data completeness
	Compressed    bool                   `json:"compressed"`
	CompressionID string                 `json:"compression_id,omitempty"`
	ErrorCount    int                    `json:"error_count"`
	RetryCount    int                    `json:"retry_count"`
}

// BloodCirculation manages bidirectional data flow between GitHub MCP and TNOS MCP
type BloodCirculation struct {
	ctx            context.Context
	cancelFunc     context.CancelFunc
	conn           *websocket.Conn
	TargetURL      string // Renamed from TnosMCPURL to be generic
	options        common.ConnectionOptions // Changed to store the whole options struct
	logger         log.LoggerInterface // Changed to interface
	State          common.ConnectionState
	cellQueue      []*BloodCell
	formulaCache   map[string]BridgeFormula
	metrics        *BloodMetrics
	mutex          sync.RWMutex
	pulseTicker    *time.Ticker
	// Compression configuration
	compressionEnabled      bool
	compressionThreshold    int
	compressionAlgorithm    string
	compressionLevel        int
	compressionPreserveKeys []string
	
	// Filter callback
	filterCallback func(cell *BloodCell) bool
}

// BloodMetrics tracks the health and performance of the blood circulation
type BloodMetrics struct {
	CellsTransmitted     int64
	CellsReceived        int64
	ErrorCount           int64
	CompressionRatio     float64
	AverageResponseTime  float64
	CurrentOxygenLevel   float64
	CirculationStartTime time.Time
	LastPulse            time.Time
	ClottingEvents       int
}

// Exported connection state constants
const (
	StateDisconnected = common.StateDisconnected
	StateConnecting   = common.StateConnecting
	StateConnected    = common.StateConnected
	StateReconnecting = common.StateReconnecting
	StateError        = common.StateError
)

// NewBloodCirculation creates a new blood circulation bridge
func NewBloodCirculation(ctx context.Context, options common.ConnectionOptions) (*BloodCirculation, error) {
	if options.Logger == nil {
		return nil, fmt.Errorf("logger is required in ConnectionOptions")
	}
	if options.ServerURL == "" {
		return nil, fmt.Errorf("serverURL is required in ConnectionOptions")
	}
	if options.QHPIntent == "" {
		options.QHPIntent = "Generic MCP" // Default intent if not specified
	}
	if options.SourceComponent == "" {
		options.SourceComponent = "Unknown TNOS Component"
	}

	ctx, cancel := context.WithCancel(ctx)


	// Create the blood circulation instance
	bc := &BloodCirculation{
		ctx:          ctx,
		cancelFunc:   cancel,
		TargetURL:    options.ServerURL, // Use ServerURL from options
		options:      options,           // Store all options
		logger:       options.Logger.(log.LoggerInterface), // Type assert to LoggerInterface
		State:        common.StateDisconnected,
		cellQueue:    make([]*BloodCell, 0, 100),
		formulaCache: make(map[string]BridgeFormula),
		metrics: &BloodMetrics{
			CirculationStartTime: time.Now(),
			LastPulse:            time.Now(),
			CurrentOxygenLevel:   1.0, // Start fully oxygenated
		},
		// Initialize compression fields from options or defaults
		compressionEnabled:      options.CompressionEnabled,
		compressionThreshold:    options.CompressionThreshold,
		compressionAlgorithm:    options.CompressionAlgorithm,
		compressionLevel:        options.CompressionLevel,
		compressionPreserveKeys: options.CompressionPreserveKeys,
		filterCallback:          nil, // Will be set below if provided
	}

	// Validate logger type if possible, or ensure it's not nil
	if _, ok := bc.logger.(*log.Logger); !ok {
		// If it's not the concrete *log.Logger, we assume it's a valid LoggerInterface.
		// No error, but a warning if a specific concrete type was expected for some internal features.
		// For now, as long as it fulfills LoggerInterface, it's fine.
	}

	// Set up FilterCallback if provided
	if options.FilterCallback != nil {
		bc.filterCallback = func(cell *BloodCell) bool {
			return options.FilterCallback(cell)
		}
	}
	
	// Start the circulation pump (will attempt connection in background)
	// No need to call bc.startCirculation() here, Start() method will do it.

	// Pre-load commonly used formulas
	if err := bc.loadCoreFormulas(); err != nil {
		bc.logger.Warn("Failed to load core formulas", "error", err, "component", bc.options.SourceComponent, "target", bc.TargetURL)
	}

	bc.logger.Info("BloodCirculation instance created", "targetURL", bc.TargetURL, "qhpIntent", bc.options.QHPIntent, "source", bc.options.SourceComponent)
	return bc, nil
}

// Start initiates the connection and circulation pump.
func (bc *BloodCirculation) Start() {
	bc.logger.Info("Starting BloodCirculation", "target", bc.TargetURL, "intent", bc.options.QHPIntent, "source", bc.options.SourceComponent)
	if err := bc.startCirculationInternal(); err != nil {
		bc.logger.Error("Failed to start circulation on Start()", "error", err, "target", bc.TargetURL)
		// bc.State will be updated by connect/reconnectLoop
	}
}


// startCirculationInternal initiates the blood pump and circulation system
func (bc *BloodCirculation) startCirculationInternal() error { // Renamed to avoid export conflict if any, and clarify internal use
    bc.mutex.Lock()
    defer bc.mutex.Unlock()

    // Start heartbeat ticker and pumps immediately, connection retries run in background
    bc.pulseTicker = time.NewTicker(CirculationInterval)
    go bc.circulationPump()
    go bc.receiveBloodCells() // Renamed from receiveBloodCells

    // Initial connection attempt (non-blocking)
    go func() {
        if err := bc.connect(); err != nil {
            bc.logger.Warn("Initial connect failed, will retry in background", "error", err, "target", bc.TargetURL, "intent", bc.options.QHPIntent, "source", bc.options.SourceComponent)
            bc.reconnectLoop()
        } else {
            bc.logger.Info("Connected on first attempt", "target", bc.TargetURL, "intent", bc.options.QHPIntent, "source", bc.options.SourceComponent)
        }
    }()

    bc.logger.Info("Blood circulation started (async connect)", "target_url", bc.TargetURL, "intent", bc.options.QHPIntent, "source", bc.options.SourceComponent)
    return nil
}

// reconnectLoop attempts connection retries until success or max retries reached
func (bc *BloodCirculation) reconnectLoop() {
	retries := 0
	maxRetries := bc.options.MaxRetries
	if maxRetries <= 0 { // Default to a sensible number if not set or invalid
		maxRetries = 5
	}
	retryDelay := bc.options.RetryDelay
	if retryDelay <= 0 {
		retryDelay = 10 * time.Second // Default delay
	}

	for {
		select {
		case <-bc.ctx.Done():
			bc.logger.Info("Reconnect loop cancelled by context", "target", bc.TargetURL, "intent", bc.options.QHPIntent, "source", bc.options.SourceComponent)
			return
		default:
			// Proceed with retry logic
		}

		if retries >= maxRetries {
			bc.logger.Error("Max reconnect attempts reached", "attempts", retries, "target", bc.TargetURL, "intent", bc.options.QHPIntent, "source", bc.options.SourceComponent)
			bc.State = common.StateError // Or StateDisconnected after max retries
			return
		}
		time.Sleep(retryDelay)
		retries++
		bc.logger.Info("Reconnection attempt", "attempt", retries, "max_attempts", maxRetries, "target", bc.TargetURL, "intent", bc.options.QHPIntent, "source", bc.options.SourceComponent)
		if err := bc.connect(); err != nil {
			bc.logger.Warn("Reconnect failed", "attempt", retries, "error", err, "target", bc.TargetURL, "intent", bc.options.QHPIntent, "source", bc.options.SourceComponent)
			bc.mutex.Lock()
			bc.metrics.ClottingEvents++
			bc.mutex.Unlock()
			logFields := map[string]interface{}{
				"event":    "reconnect_failure",
				"attempt":  retries,
				"target":   bc.TargetURL,
				"intent":   bc.options.QHPIntent,
				"source":   bc.options.SourceComponent,
			}
			bc.logger.Info("Hemoflux checkpoint", logFields)
			continue
		}
		bc.logger.Info("Reconnected successfully", "target", bc.TargetURL, "intent", bc.options.QHPIntent, "source", bc.options.SourceComponent)
		return // Exit loop on successful reconnect
	}
}

// connect establishes a WebSocket connection to the TargetURL
func (bc *BloodCirculation) connect() error {
	bc.mutex.Lock()
	if bc.State == common.StateConnected {
		bc.mutex.Unlock()
		return nil // Already connected
	}
	bc.State = common.StateConnecting
	targetLogURL := bc.TargetURL // Capture for logging before unlock
	intentLog := bc.options.QHPIntent
	sourceLog := bc.options.SourceComponent
	bc.mutex.Unlock()

	bc.logger.Info("Starting QHP handshake", "url", targetLogURL, "intent", intentLog, "source", sourceLog)

	u, err := url.Parse(targetLogURL)
	if err != nil {
		bc.mutex.Lock()
		bc.State = common.StateError
		bc.mutex.Unlock()
		bc.logger.Error("Invalid Target URL", "url", targetLogURL, "error", err, "intent", intentLog, "source", sourceLog)
		return fmt.Errorf("invalid Target URL %s: %w", targetLogURL, err)
	}

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = bc.options.Timeout
	if dialer.HandshakeTimeout <= 0 {
		dialer.HandshakeTimeout = 30 * time.Second // Default timeout
	}

	header := make(http.Header) // Use http.Header for standard compliance
	header.Set("X-QHP-Version", common.MCPVersion30)
	header.Set("X-QHP-Source", sourceLog) // Use SourceComponent from options
	header.Set("X-QHP-Intent", intentLog) // Use QHPIntent from options
	// Add any other custom headers from bc.options.CustomHeaders if defined

	conn, resp, err := dialer.DialContext(bc.ctx, u.String(), header)
	if err != nil {
		bc.mutex.Lock()
		bc.State = common.StateError
		bc.metrics.ClottingEvents++
		bc.mutex.Unlock()
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		bc.logger.Error("QHP handshake failed", "url", targetLogURL, "status_code", statusCode, "error", err, "intent", intentLog, "source", sourceLog)
		logFields := map[string]interface{}{
			"event":    "qhp_failure",
			"endpoint": targetLogURL,
			"intent":   intentLog,
			"source":   sourceLog,
		}
		if resp != nil {
			logFields["status_code"] = resp.StatusCode
		}
		bc.logger.Info("Hemoflux checkpoint", logFields)
		return fmt.Errorf("QHP handshake failed for %s (intent: %s): %w", targetLogURL, intentLog, err)
	}

	bc.mutex.Lock()
	bc.conn = conn
	bc.State = common.StateConnected
	bc.mutex.Unlock()
	bc.logger.Info("Connected via QHP", "url", targetLogURL, "intent", intentLog, "source", sourceLog)
	
	// Start the receiver goroutine now that connection is established
	// This was previously in startCirculationInternal, but better here to ensure conn is not nil.
	// However, receiveBloodCells is already started in startCirculationInternal.
	// We need to ensure it handles nil conn gracefully or is started after successful connect.
	// For now, let's assume receiveBloodCells handles bc.conn being nil until connection.
	// Or, signal receiveBloodCells to start processing.
	// A simpler approach: receiveBloodCells loop checks bc.State and bc.conn.

	return nil
}

// loadCoreFormulas loads the essential formulas needed for blood circulation from the formula registry
func (bc *BloodCirculation) loadCoreFormulas() error {
	registry := GetBridgeFormulaRegistry()
	if registry == nil {
		return fmt.Errorf("formula registry is not initialized")
	}
	
	// Get essential formulas
	essentialFormulas := []string{
		"hemoflux.compress",
		"hemoflux.decompress",
		"tnos.translate",
		"github.translate",
		"blood.oxygenate",
	}
	
	for _, id := range essentialFormulas {
		formula, ok := registry.GetFormula(id)
		if !ok {
			return fmt.Errorf("essential formula not found: %s", id)
		}
		bc.formulaCache[id] = formula
	}
	
	return nil
}

// circulationPump continuously pumps blood cells through the system
func (bc *BloodCirculation) circulationPump() {
	bc.logger.Info("Blood circulation pump started")
	
	for {
		select {
		case <-bc.ctx.Done():
			bc.logger.Info("Blood circulation pump stopped")
			return
		case <-bc.pulseTicker.C:
			// Send heartbeat pulse
			bc.pulse()
			
			// Process queued cells
			bc.processQueuedCells()
		}
	}
}

// pulse sends a heartbeat to maintain connection and monitor system health
func (bc *BloodCirculation) pulse() {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if bc.State != common.StateConnected {
		return
	}

	pulseCell := &BloodCell{
		ID:          "pulse-" + time.Now().Format(time.RFC3339Nano),
		Type:        "Platelet",
		Payload:     map[string]interface{}{"action": "pulse", "timestamp": time.Now().Unix()},
		Timestamp:   time.Now().Unix(),
		Source:      "github-mcp-server",
		Destination: "tnos-mcp-server",
		Priority:    1,
		OxygenLevel: bc.metrics.CurrentOxygenLevel,
		Context7D: translations.ContextVector7D{
			Who:    "BloodCirculation",
			What:   "Pulse",
			When:   time.Now().Unix(),
			Where:  "github-mcp-server/pkg/bridge",
			Why:    "Heartbeat",
			How:    "WebSocket",
			Extent: 0.1,
			Source: "github-mcp-server",
		},
	}

	contextMap := common.GenerateContextMap(pulseCell.Context7D)
	bc.logger.Debug("Generated context map", "contextMap", contextMap)

	if err := bc.sendBloodCell(pulseCell); err != nil {
		bc.logger.Error("Failed to send pulse", "error", err)
		bc.metrics.ErrorCount++
		
		// Check if we need to attempt reconnection
		if bc.metrics.ErrorCount > ClottingThreshold {
			bc.logger.Warn("Clotting threshold reached, attempting to reconnect")
			bc.reconnect()
		}
	} else {
		bc.metrics.LastPulse = time.Now()
	}
}

// processQueuedCells processes blood cells waiting in the queue
func (bc *BloodCirculation) processQueuedCells() {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	if len(bc.cellQueue) == 0 {
		return
	}
	
	bc.logger.Debug("Processing queued cells", "count", len(bc.cellQueue))
	
	// Process cells in priority order (higher first)
	for i := len(bc.cellQueue) - 1; i >= 0; i-- {
		cell := bc.cellQueue[i]
		
		// Apply hemoflux compression if needed
		if HemofluxCompression && !cell.Compressed {
			if err := bc.compressCell(cell); err != nil {
				bc.logger.Error("Failed to compress cell", "id", cell.ID, "error", err)
			}
		}
		
		// Apply blood filtering if callback is set
		if bc.filterCallback != nil && !bc.filterCallback(cell) {
			bc.cellQueue = append(bc.cellQueue[:i], bc.cellQueue[i+1:]...)
			bc.logger.Debug("Cell filtered out", "id", cell.ID)
			continue
		}
		
		// Send the cell
		if err := bc.sendBloodCell(cell); err != nil {
			bc.logger.Error("Failed to send cell", "id", cell.ID, "error", err)
			cell.RetryCount++
			
			// Keep in queue if under retry limit
			if cell.RetryCount < 3 {
				continue
			}
		}
		
		// Remove from queue on success or exceeded retry limit
		bc.cellQueue = append(bc.cellQueue[:i], bc.cellQueue[i+1:]...)
	}
}

// receiveBloodCells continuously reads incoming blood cells
func (bc *BloodCirculation) receiveBloodCells() {
	bc.logger.Info("Starting to receive blood cells")
	
	for {
		select {
		case <-bc.ctx.Done():
			bc.logger.Info("Blood cell receiver stopped")
			return
		default:
			if bc.State != common.StateConnected || bc.conn == nil {
				time.Sleep(1 * time.Second)
				continue
			}
			
			// Read message
			_, message, err := bc.conn.ReadMessage()
			if err != nil {
				bc.logger.Error("Error reading message", "error", err)
				bc.reconnect()
				time.Sleep(1 * time.Second)
				continue
			}
			
			// Process the message
			cell := &BloodCell{}
			if err := json.Unmarshal(message, cell); err != nil {
				bc.logger.Error("Failed to unmarshal blood cell", "error", err)
				continue
			}
			
			bc.metrics.CellsReceived++
			bc.logger.Debug("Received blood cell", "id", cell.ID, "type", cell.Type)
			
			// Decompress if needed
			if cell.Compressed {
				if err := bc.decompressCell(cell); err != nil {
					bc.logger.Error("Failed to decompress cell", "id", cell.ID, "error", err)
					continue
				}
			}
			
			// Process the cell
			go bc.processReceivedCell(cell)
		}
	}
}

// processReceivedCell processes a blood cell received from the TNOS MCP server
func (bc *BloodCirculation) processReceivedCell(cell *BloodCell) {
	// Handle different cell types
	switch cell.Type {
	case "Red": // Data cell
		bc.logger.Info("Received data cell", "id", cell.ID, "from", cell.Source)
		// Process data and forward to GitHub MCP
		// Implementation depends on how data should be forwarded to GitHub MCP
		
	case "White": // Control cell
		bc.logger.Info("Received control cell", "id", cell.ID, "action", cell.Payload["action"])
		// Handle control actions
		if action, ok := cell.Payload["action"].(string); ok {
			switch action {
			case "ping":
				// Respond to ping with pong
				bc.sendResponse(cell, "pong", nil)
				
			case "formula_request":
				// Handle formula execution request
				formulaID, _ := cell.Payload["formula_id"].(string)
				params, _ := cell.Payload["parameters"].(map[string]interface{})
				
				result, err := bc.executeFormula(formulaID, params)
				if err != nil {
					bc.sendResponse(cell, "formula_error", map[string]interface{}{
						"error": err.Error(),
					})
				} else {
					bc.sendResponse(cell, "formula_result", map[string]interface{}{
						"result": result,
					})
				}
			}
		}
		
	case "Platelet": // Recovery cell
		bc.logger.Info("Received recovery cell", "id", cell.ID)
		// Handle system recovery or diagnostics
		if action, ok := cell.Payload["action"].(string); ok && action == "pulse_response" {
			// Update health metrics
			bc.mutex.Lock()
			bc.metrics.CurrentOxygenLevel = cell.OxygenLevel
			bc.mutex.Unlock()
		}
	}
}

// Replace context map generation with the shared logic function
func (bc *BloodCirculation) sendResponse(originalCell *BloodCell, responseType string, payload map[string]interface{}) {
	if payload == nil {
		payload = make(map[string]interface{})
	}
	
	// Add reference to original cell
	payload["reference_id"] = originalCell.ID
	payload["action"] = responseType
	
	responseCell := &BloodCell{
		ID:          responseType + "-" + time.Now().Format(time.RFC3339Nano),
		Type:        originalCell.Type,
		Payload:     payload,
		Timestamp:   time.Now().Unix(),
		Source:      "github-mcp-server",
		Destination: originalCell.Source,
		Priority:    originalCell.Priority,
		OxygenLevel: bc.metrics.CurrentOxygenLevel,
		Context7D: translations.ContextVector7D{
			Who:    "BloodCirculation",
			What:   "Response",
			When:   time.Now().Unix(),
			Where:  "github-mcp-server/pkg/bridge",
			Why:    "Reply",
			How:    "WebSocket",
			Extent: originalCell.Context7D.Extent,
			Source: "github-mcp-server",
		},
	}
	
	bc.QueueBloodCell(responseCell)
}

// executeFormula executes a formula from the registry with the given parameters
func (bc *BloodCirculation) executeFormula(formulaID string, params map[string]interface{}) (interface{}, error) {
	if bc.State == common.StateConnected && bc.conn != nil {
		// Delegate formula execution to TNOS MCP server via blood bridge
		request := map[string]interface{}{
			"action": "formula_request",
			"formula_id": formulaID,
			"parameters": params,
		}
		// Send request and wait for response (synchronously for simplicity)
		response, err := bc.sendFormulaRequestToTNOS(request)
		if err != nil {
			return nil, err
		}
		return response, nil
	}
	// Standalone: use local registry
	formula, ok := bc.formulaCache[formulaID]
	if !ok {
		registry := GetBridgeFormulaRegistry()
		if registry == nil {
			return nil, fmt.Errorf("formula registry is not initialized")
		}
		formula, ok = registry.GetFormula(formulaID)
		if !ok {
			return nil, fmt.Errorf("formula not found: %s", formulaID)
		}
		bc.formulaCache[formulaID] = formula
	}
	bc.logger.Info("Executing formula (standalone)", "id", formulaID, "params", params)
	if len(formula.ContextReqs) > 0 {
		ctx, ok := params["context"].(map[string]interface{})
		if !ok {
			ctx = make(map[string]interface{})
			params["context"] = ctx
		}
		for _, req := range formula.ContextReqs {
			if _, exists := ctx[req]; !exists {
				bc.logger.Warn("Missing required context parameter", "formula", formulaID, "param", req)
			}
		}
	}
	registry := GetBridgeFormulaRegistry()
	if registry == nil {
		return nil, fmt.Errorf("formula registry is not initialized")
	}
	result, err := registry.ExecuteFormula(formulaID, params)
	if err != nil {
		bc.logger.Error("Formula execution failed", "id", formulaID, "error", err)
		return nil, err
	}
	result["timestamp"] = time.Now().Unix()
	result["formula_id"] = formulaID
	return result, nil
}

// sendFormulaRequestToTNOS sends a formula execution request to the TNOS MCP server and waits for a response
func (bc *BloodCirculation) sendFormulaRequestToTNOS(request map[string]interface{}) (map[string]interface{}, error) {
	// This is a placeholder for the actual bridge logic
	// In production, this would send the request over the WebSocket and wait for a response
	// For now, return a not implemented error
	return nil, fmt.Errorf("remote formula execution via blood bridge not yet implemented")
}

// compressCell applies hemoflux compression to a blood cell
func (bc *BloodCirculation) compressCell(cell *BloodCell) error {
	// Get the registry for formula execution
	registry := GetBridgeFormulaRegistry()
	if registry == nil {
		return fmt.Errorf("formula registry not initialized")
	}
	
	// Prepare the Hemoflux compression context using shared utility
	context := common.GenerateContextMap(cell.Context7D)
	// Add extra fields for compression context
	context["destination"] = cell.Destination
	context["timestamp"] = cell.Timestamp
	context["cell_type"] = cell.Type
	context["priority"] = cell.Priority
	
	// Prepare cell payload for compression
	payloadData, err := json.Marshal(cell.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal cell payload: %w", err)
	}
	
	// Execute the hemoflux.compress formula (this delegates to the Formula Registry)
	// Note: Actual compression would be done by HemoFlux which is the sole executor of Mobius Equation
	compressionParams := map[string]interface{}{
		"data":         string(payloadData),
		"context":      context,
		"formula_key": "tnos_mcp_bridge",
	}
	
	result, err := registry.ExecuteFormula("hemoflux.compress", compressionParams)
	if err != nil {
		return fmt.Errorf("hemoflux compression failed: %w", err)
	}
	
	// In production, this compressed data would come from the actual HemoFlux system
	// For now, we just simulate compression behavior
	cell.Compressed = true
	cell.CompressionID = "hemoflux-" + time.Now().Format(time.RFC3339Nano)
	
	// Update payload to include compression metadata
	if compressedPayload, ok := result["compressed_payload"].(map[string]interface{}); ok {
		cell.Payload = compressedPayload
	}
	
	// Track compression ratio for metrics
	if ratio, ok := result["compression_ratio"].(float64); ok {
		bc.mutex.Lock()
		bc.metrics.CompressionRatio = ratio
		bc.mutex.Unlock()
	}
	
	bc.logger.Debug("Applied HemoFlux compression to cell", 
		"id", cell.ID, 
		"compression_id", cell.CompressionID)
	
	return nil
}

// decompressCell decompresses a blood cell
func (bc *BloodCirculation) decompressCell(cell *BloodCell) error {
	// Get the registry for formula execution
	registry := GetBridgeFormulaRegistry()
	if registry == nil {
		return fmt.Errorf("formula registry not initialized")
	}
	
	// Skip if not compressed
	if !cell.Compressed {
		return nil
	}
	
	// Execute the hemoflux.decompress formula (this delegates to the Formula Registry)
	// Note: Actual decompression would be done by HemoFlux which is the sole executor of Mobius Equation
	decompressionParams := map[string]interface{}{
		"compressed_payload": cell.Payload,
		"compression_id":     cell.CompressionID,
		"context":            common.GenerateContextMap(cell.Context7D),
	}
	
	result, err := registry.ExecuteFormula("hemoflux.decompress", decompressionParams)
	if err != nil {
		return fmt.Errorf("hemoflux decompression failed: %w", err)
	}
	
	// Update payload with decompressed data
	if decompressedPayload, ok := result["decompressed_payload"].(map[string]interface{}); ok {
		cell.Payload = decompressedPayload
	}
	
	cell.Compressed = false
	
	bc.logger.Debug("Applied HemoFlux decompression to cell", "id", cell.ID)
	
	return nil
}

// sendBloodCell sends a blood cell to the TNOS MCP server
func (bc *BloodCirculation) sendBloodCell(cell *BloodCell) error {
	if bc.State != common.StateConnected || bc.conn == nil {
		return fmt.Errorf("not connected to TNOS MCP")
	}
	
	data, err := json.Marshal(cell)
	if err != nil {
		return fmt.Errorf("failed to marshal blood cell: %w", err)
	}
	
	if err := bc.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("failed to send blood cell: %w", err)
	}
	
	bc.metrics.CellsTransmitted++
	return nil
}

// Export SendBloodCell method
func (bc *BloodCirculation) SendBloodCell(cell *BloodCell) error {
	return bc.sendBloodCell(cell)
}

// QueueBloodCell queues a blood cell for circulation
func (bc *BloodCirculation) QueueBloodCell(cell *BloodCell) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	bc.cellQueue = append(bc.cellQueue, cell)
	bc.logger.Debug("Blood cell queued", "id", cell.ID, "type", cell.Type, "queue_size", len(bc.cellQueue))
}

// reconnect attempts to reconnect to the TNOS MCP server
func (bc *BloodCirculation) reconnect() {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	bc.metrics.ClottingEvents++
	bc.logger.Info("Attempting to reconnect to TNOS MCP", "attempt", bc.metrics.ClottingEvents)
	
	// Close existing connection if any
	if bc.conn != nil {
		_ = bc.conn.Close()
		bc.conn = nil
	}
	
	bc.State = common.StateReconnecting
	
	// Try to reconnect
	if err := bc.connect(); err != nil {
		bc.logger.Error("Failed to reconnect", "error", err)
		bc.State = common.StateError
		
		// Schedule another attempt
		time.AfterFunc(common.ReconnectDelay, func() {
			bc.reconnect()
		})
	} else {
		bc.logger.Info("Reconnected successfully to TNOS MCP")
		bc.State = common.StateConnected
		bc.metrics.ErrorCount = 0
	}
}

// SetBloodFilter sets a filter callback for blood cells
func (bc *BloodCirculation) SetBloodFilter(callback func(cell *BloodCell) bool) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	bc.filterCallback = callback
}

// GetCirculationMetrics returns metrics for the blood circulation
func (bc *BloodCirculation) GetCirculationMetrics() BloodMetrics {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	return *bc.metrics
}

// SetCompressionOptions configures Hemoflux compression settings for the blood circulation
func (bc *BloodCirculation) SetCompressionOptions(options map[string]interface{}) error {
	// WHO: CompressionManager
	// WHAT: Configure blood cell compression
	// WHEN: During bridge initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To optimize data transfer
	// HOW: Using Hemoflux compression algorithm
	// EXTENT: All compressed blood cells

	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	// Get the formula registry
	reg := GetBridgeFormulaRegistry()
	if reg == nil {
		return fmt.Errorf("formula registry is not initialized")
	}
	
	// Check if the compression/decompression formulas exist
	if _, exists := reg.GetFormula("hemoflux.compress"); !exists {
		return fmt.Errorf("hemoflux compression formula not found")
	}
	
	if _, exists := reg.GetFormula("hemoflux.decompress"); !exists {
		return fmt.Errorf("hemoflux decompression formula not found")
	}
	
	// Extract and validate settings
	enabled, _ := options["enabled"].(bool)
	threshold := 1024 // Default 1KB
	if thresholdVal, ok := options["threshold"].(float64); ok {
		threshold = int(thresholdVal)
	} else if thresholdVal, ok := options["threshold"].(int); ok {
		threshold = thresholdVal
	}
	
	algorithm := "hemoflux.mobius"
	if algorithmVal, ok := options["algorithm"].(string); ok {
		algorithm = algorithmVal
	}
	
	level := 5 // Default compression level
	if levelVal, ok := options["level"].(float64); ok {
		if levelVal >= 1 && levelVal <= 9 {
			level = int(levelVal)
		}
	} else if levelVal, ok := options["level"].(int); ok {
		if levelVal >= 1 && levelVal <= 9 {
			level = levelVal
		}
	}
	
	// Update blood circulation configuration
	bc.compressionEnabled = enabled
	bc.compressionThreshold = threshold
	bc.compressionAlgorithm = algorithm
	bc.compressionLevel = level
	
	// Set up preserved keys if provided
	if preserveKeys, ok := options["preserve_keys"].([]string); ok {
		bc.compressionPreserveKeys = preserveKeys
	} else if preserveKeys, ok := options["preserve_keys"].([]interface{}); ok {
		// Convert from []interface{} to []string
		bc.compressionPreserveKeys = make([]string, 0, len(preserveKeys))
		for _, key := range preserveKeys {
			if strKey, ok := key.(string); ok {
				bc.compressionPreserveKeys = append(bc.compressionPreserveKeys, strKey)
			}
		}
	}
	
	bc.logger.Info("Compression settings updated",
		"enabled", enabled,
		"algorithm", algorithm,
		"level", level,
		"threshold", threshold)
	
	return nil
}

// HasCompressionEnabled returns whether Hemoflux compression is enabled
func (bc *BloodCirculation) HasCompressionEnabled() bool {
	// WHO: CompressionManager
	// WHAT: Check compression status
	// WHEN: During blood cell processing
	// WHERE: System Layer 6 (Integration)
	// WHY: To determine if compression should be applied
	// HOW: Using BloodCirculation state
	// EXTENT: All blood cells

	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	return bc.compressionEnabled
}

// GetCompressionThreshold returns the size threshold for compression
func (bc *BloodCirculation) GetCompressionThreshold() int {
	// WHO: CompressionManager
	// WHAT: Get compression threshold
	// WHEN: During blood cell processing
	// WHERE: System Layer 6 (Integration)
	// WHY: To determine if a payload is large enough to compress
	// HOW: Using BloodCirculation configuration
	// EXTENT: All blood cells

	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	return bc.compressionThreshold
}

// GetCompressionOptions returns all compression settings as a map
func (bc *BloodCirculation) GetCompressionOptions() map[string]interface{} {
	// WHO: CompressionManager
	// WHAT: Return all compression settings
	// WHEN: During configuration introspection
	// WHERE: System Layer 6 (Integration)
	// WHY: To examine current compression configuration
	// HOW: Using BloodCirculation state
	// EXTENT: All configuration parameters

	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	return map[string]interface{}{
		"enabled":        bc.compressionEnabled,
		"threshold":      bc.compressionThreshold,
		"algorithm":      bc.compressionAlgorithm,
		"level":          bc.compressionLevel,
		"preserve_keys":  bc.compressionPreserveKeys,
	}
}

// Close shuts down the blood circulation
func (bc *BloodCirculation) Close() error {
	bc.logger.Info("Shutting down blood circulation")
	
	// Stop the pulse ticker
	if bc.pulseTicker != nil {
		bc.pulseTicker.Stop()
	}
	
	// Close the connection
	if bc.conn != nil {
		if err := bc.conn.Close(); err != nil {
			bc.logger.Error("Error closing WebSocket connection", "error", err)
		}
	}
	
	// Cancel the context
	bc.cancelFunc()
	
	bc.State = common.StateDisconnected
	return nil
}

// GetState returns the current connection state of the bridge.
func (bc *BloodCirculation) GetState() common.ConnectionState {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	return bc.State
}

// GetMetrics returns the current metrics for the blood circulation.
func (bc *BloodCirculation) GetMetrics() *BloodMetrics {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	// Return a copy to prevent modification of the internal metrics struct
	metricsCopy := *bc.metrics
	return &metricsCopy
}

// Shutdown gracefully stops the blood circulation.
func (bc *BloodCirculation) Shutdown() {
	bc.logger.Info("Shutting down BloodCirculation", "target", bc.TargetURL, "intent", bc.options.QHPIntent, "source", bc.options.SourceComponent)
	bc.cancelFunc() // Signal all goroutines to stop

	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if bc.conn != nil {
		// Attempt to send a close message
		// Set a deadline for the close message
		deadline := time.Now().Add(5 * time.Second)
		err := bc.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), deadline)
		if err != nil {
			bc.logger.Warn("Failed to send close message during shutdown", "error", err, "target", bc.TargetURL)
		}
		bc.conn.Close()
		bc.conn = nil
	}
	bc.State = common.StateDisconnected
	if bc.pulseTicker != nil {
		bc.pulseTicker.Stop()
	}
	bc.logger.Info("BloodCirculation shutdown complete", "target", bc.TargetURL, "intent", bc.options.QHPIntent, "source", bc.options.SourceComponent)
}
