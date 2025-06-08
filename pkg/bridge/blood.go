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
	"net/url"
	"sync"
	"time"

	"github-mcp-server/pkg/common"
	"github-mcp-server/pkg/log"
	"github-mcp-server/pkg/translations"

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
	tnosMCPURL     string
	options        common.ConnectionOptions
	logger         *log.Logger
	state          common.ConnectionState
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

// NewBloodCirculation creates a new blood circulation bridge
func NewBloodCirculation(ctx context.Context, options common.ConnectionOptions) (*BloodCirculation, error) {
	ctx, cancel := context.WithCancel(ctx)
	
	// Default to TNOS MCP server URL if not provided
	tnosMCPURL := fmt.Sprintf("ws://localhost:%d", TNOSMCPPort)
	if options.ServerURL != "" && options.ServerPort > 0 {
		tnosMCPURL = fmt.Sprintf("ws://%s:%d", options.ServerURL, options.ServerPort)
	}
	
	// Create the blood circulation instance
	bc := &BloodCirculation{
		ctx:          ctx,
		cancelFunc:   cancel,
		tnosMCPURL:   tnosMCPURL,
		options:      options,
		logger:       options.Logger.(*log.Logger),
		state:        common.StateDisconnected,
		cellQueue:    make([]*BloodCell, 0, 100),
		formulaCache: make(map[string]BridgeFormula),
		metrics: &BloodMetrics{
			CirculationStartTime: time.Now(),
			LastPulse:            time.Now(),
			CurrentOxygenLevel:   1.0,
		},
	}
	
	// Start the circulation pump
	if err := bc.startCirculation(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start circulation: %w", err)
	}
	
	// Pre-load commonly used formulas
	if err := bc.loadCoreFormulas(); err != nil {
		bc.logger.Warn("Failed to load core formulas: %v", err)
	}
	
	return bc, nil
}

// startCirculation initiates the blood pump and circulation system
func (bc *BloodCirculation) startCirculation() error {
    bc.mutex.Lock()
    defer bc.mutex.Unlock()

    // Start heartbeat ticker and pumps immediately, connection retries run in background
    bc.pulseTicker = time.NewTicker(CirculationInterval)
    go bc.circulationPump()
    go bc.receiveBloodCells()

    // Initial connection attempt (non-blocking)
    go func() {
        if err := bc.connect(); err != nil {
            bc.logger.Warn("Initial connect failed, will retry in background", "error", err)
            bc.reconnectLoop()
        } else {
            bc.logger.Info("Connected on first attempt", "url", bc.tnosMCPURL)
        }
    }()

    bc.logger.Info("Blood circulation started (async connect)", "tnos_mcp_url", bc.tnosMCPURL)
    return nil
}

// reconnectLoop attempts connection retries until success or max retries reached
func (bc *BloodCirculation) reconnectLoop() {
   retries := 0
   for {
       if bc.options.MaxRetries > 0 && retries >= bc.options.MaxRetries {
           bc.logger.Error("Max reconnect attempts reached", "attempts", retries)
           return
       }
       time.Sleep(bc.options.RetryDelay)
       bc.logger.Info("Reconnection attempt", "attempt", retries+1)
       if err := bc.connect(); err != nil {
           retries++
           bc.logger.Warn("Reconnect failed", "attempt", retries, "error", err)
           bc.metrics.ClottingEvents++
           bc.logger.Info("Hemoflux checkpoint", "event", "reconnect_failure", "attempt", retries)
           continue
       }
       bc.logger.Info("Reconnected successfully")
       return
   }
}

// connect establishes a WebSocket connection to the TNOS MCP server
func (bc *BloodCirculation) connect() error {
    bc.state = common.StateConnecting
    bc.logger.Info("Starting QHP handshake sequence")

    // Define endpoints in priority order
    type endpoint struct{ name, url string }
    endpoints := []endpoint{
        {"TNOS MCP", fmt.Sprintf("ws://localhost:%d", TNOSMCPPort)},
        {"Bridge MCP", fmt.Sprintf("ws://localhost:%d", BridgePort)},
        {"GitHub MCP", fmt.Sprintf("ws://localhost:%d", GitHubMCPPort)},
        {"Copilot LLM", fmt.Sprintf("ws://localhost:%d", CopilotLLMPort)},
    }
    // Include Python MCP stub fallback
    if bc.options.PythonMCPURL != "" {
        endpoints = append(endpoints, endpoint{"Python MCP Stub", bc.options.PythonMCPURL})
    } else {
        endpoints = append(endpoints, endpoint{"Python MCP Stub", fmt.Sprintf("ws://localhost:%d", TNOSMCPPort)})
    }

    var lastErr error
    for _, ep := range endpoints {
        bc.logger.Info("Attempting QHP handshake", "endpoint", ep.name, "url", ep.url)
        u, err := url.Parse(ep.url)
        if err != nil {
            bc.logger.Error("Invalid endpoint URL", "endpoint", ep.name, "error", err)
            lastErr = err
            continue
        }
        dialer := websocket.DefaultDialer
        dialer.HandshakeTimeout = bc.options.Timeout

        // Apply QHP headers
        header := make(map[string][]string)
        header["X-QHP-Version"] = []string{common.MCPVersion30}
        header["X-QHP-Source"] = []string{"github-mcp-server"}
        header["X-QHP-Intent"] = []string{ep.name}

        conn, _, err := dialer.Dial(u.String(), header)
        if err != nil {
            bc.metrics.ClottingEvents++
            bc.logger.Error("QHP handshake failed", "endpoint", ep.name, "error", err)
            bc.logger.Info("Hemoflux checkpoint", "event", "qhp_failure", "endpoint", ep.name)
            lastErr = err
            continue
        }
        // Successful connection
        bc.conn = conn
        bc.state = common.StateConnected
        bc.logger.Info("Connected via QHP", "endpoint", ep.name, "url", ep.url)
        return nil
    }
    // All fallbacks exhausted
    bc.state = common.StateError
    return fmt.Errorf("all QHP fallback attempts failed, last error: %w", lastErr)
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

	if bc.state != common.StateConnected {
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
			if bc.state != common.StateConnected || bc.conn == nil {
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
	// Get formula from cache or registry
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
		
		// Cache for future use
		bc.formulaCache[formulaID] = formula
	}
	
	// Log formula execution
	bc.logger.Info("Executing formula", "id", formulaID, "params", params)
	
	// Check if formula requires specific context parameters
	if len(formula.ContextReqs) > 0 {
		// Ensure context is provided
		ctx, ok := params["context"].(map[string]interface{})
		if !ok {
			ctx = make(map[string]interface{})
			params["context"] = ctx
		}
		
		// Check for missing required context parameters
		for _, req := range formula.ContextReqs {
			if _, exists := ctx[req]; !exists {
				bc.logger.Warn("Missing required context parameter", "formula", formulaID, "param", req)
			}
		}
	}
	
	// Execute the formula using the registry
	registry := GetBridgeFormulaRegistry()
	if registry == nil {
		return nil, fmt.Errorf("formula registry is not initialized")
	}
	
	result, err := registry.ExecuteFormula(formulaID, params)
	if err != nil {
		bc.logger.Error("Formula execution failed", "id", formulaID, "error", err)
		return nil, err
	}
	
	// Add execution timestamp and metadata
	result["timestamp"] = time.Now().Unix()
	result["formula_id"] = formulaID
	
	// Track formula usage in metrics
	// TODO: Implement formula usage tracking
	
	return result, nil
}

// compressCell applies hemoflux compression to a blood cell
func (bc *BloodCirculation) compressCell(cell *BloodCell) error {
	// Get the registry for formula execution
	registry := GetBridgeFormulaRegistry()
	if registry == nil {
		return fmt.Errorf("formula registry not initialized")
	}
	
	// Prepare the Hemoflux compression context
	context := map[string]interface{}{
		"who":     cell.Context7D.Who,
		"what":    cell.Context7D.What,
		"when":    cell.Context7D.When,
		"where":   cell.Context7D.Where,
		"why":     cell.Context7D.Why,
		"how":     cell.Context7D.How,
		"extent":  cell.Context7D.Extent,
		"source":  cell.Source,
		"destination": cell.Destination,
		"timestamp":   cell.Timestamp,
		"cell_type":   cell.Type,
		"priority":    cell.Priority,
	}
	
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
		"context": map[string]interface{}{
			"who":     cell.Context7D.Who,
			"what":    cell.Context7D.What,
			"when":    cell.Context7D.When,
			"where":   cell.Context7D.Where,
			"why":     cell.Context7D.Why,
			"how":     cell.Context7D.How,
			"extent":  cell.Context7D.Extent,
		},
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
	if bc.state != common.StateConnected || bc.conn == nil {
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
	
	bc.state = common.StateReconnecting
	
	// Try to reconnect
	if err := bc.connect(); err != nil {
		bc.logger.Error("Failed to reconnect", "error", err)
		bc.state = common.StateError
		
		// Schedule another attempt
		time.AfterFunc(common.ReconnectDelay, func() {
			bc.reconnect()
		})
	} else {
		bc.logger.Info("Reconnected successfully to TNOS MCP")
		bc.state = common.StateConnected
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
	
	bc.state = common.StateDisconnected
	return nil
}
