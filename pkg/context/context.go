// TRANQUILSPEAK SYMBOL CLUSTER: [𝒯🧬ζℏƒ𓆑#CT1𝑪𝑪𝑶𝑪𝑶𝑵𝑪𝑶𝑵𝑻𝑪𝑶𝑵𝑻𝑬𝑪𝑶𝑵𝑬𝑿𝑬𝑿⏳📍𝒮𝓔𝓗]
// HEMOFLUX_FILE_ID: "_USERS_JUBICUDIS_TRANQUILITY-NEURO-OS_GITHUB-MCP-SERVER_PKG_CONTEXT_CONTEXT.GO"
// HEMOFLUX_FORMULA: mobius.7d.orchestration
/*
 * WHO: ContextOrchestrator
 * WHAT: Core 7D context orchestration and management using biological neural patterns
 * WHEN: During all inter-system communication and context operations
 * WHERE: System Layer 6 (Integration) - Neural System mimicry
 * WHY: To provide the central nervous system for context awareness across TNOS
 * HOW: Using 7D Context Framework with ATM triggers and biological patterns
 * EXTENT: All context operations throughout the entire TNOS ecosystem
 */

package context

import (
	"fmt"
	"sync"
	"time"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log"
	tspeak "github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/tranquilspeak"
)

// ContextVector7D represents the fundamental 7-dimensional context structure
// This is the neural data structure that flows through our biological system
type ContextVector7D = log.ContextVector7D

// ContextOrchestrator - The central nervous system for context management
// WHO: Neural Context Orchestrator 
// WHAT: Manages all 7D context operations using biological neural patterns
// WHY: To provide centralized, neural-pattern-based context coordination
// HOW: Through ATM triggers, biological state management, and 7D awareness
type ContextOrchestrator struct {
	// Neural system components
	neuralTriggerMatrix *tspeak.TriggerMatrix  // The trigger matrix - our neural pathway system
	logger             log.LoggerInterface     // Biological logging system
	
	// Neural state management
	contextMemory      map[string]ContextVector7D  // Neural memory storage
	synapticConnections map[string][]string        // Inter-context connections
	neuralActivity     map[string]time.Time        // Neural activity tracking
	
	// Biological state protection
	synapticMutex      sync.RWMutex               // Neural pathway protection
	memoryMutex        sync.RWMutex               // Memory protection
	
	// Neural metrics (biological monitoring)
	contextProcessed   int64                      // Synaptic firing count
	errorCount         int64                      // Neural error count
	lastActivity       time.Time                  // Last neural activity
}

// NewContextOrchestrator creates a new neural context orchestrator
// Following biological neural system initialization patterns
func NewContextOrchestrator(logger log.LoggerInterface) *ContextOrchestrator {
	orchestrator := &ContextOrchestrator{
		neuralTriggerMatrix: tspeak.NewTriggerMatrix(),
		logger:             logger,
		contextMemory:      make(map[string]ContextVector7D),
		synapticConnections: make(map[string][]string),
		neuralActivity:     make(map[string]time.Time),
		lastActivity:       time.Now(),
	}
	
	// Initialize neural pathways (register ATM triggers)
	orchestrator.initializeNeuralPathways()
	
	// Log neural system initialization
	orchestrator.logNeuralActivity("Neural Context Orchestrator initialized", map[string]interface{}{
		"who":     "ContextOrchestrator",
		"what":    "neural_initialization",
		"when":    time.Now().Unix(),
		"where":   "context_package",
		"why":     "biological_context_management",
		"how":     "neural_pathway_setup",
		"extent":  1.0,
		"biological_system": "neural_context_system",
	})
	
	return orchestrator
}

// initializeNeuralPathways sets up the neural trigger pathways for context operations
func (co *ContextOrchestrator) initializeNeuralPathways() {
	// Context storage neural pathway
	co.neuralTriggerMatrix.RegisterTrigger("context.store", func(trigger tspeak.ATMTrigger) error {
		return co.handleContextStorage(trigger)
	})
	
	// Context retrieval neural pathway  
	co.neuralTriggerMatrix.RegisterTrigger("context.retrieve", func(trigger tspeak.ATMTrigger) error {
		return co.handleContextRetrieval(trigger)
	})
	
	// Context transformation neural pathway
	co.neuralTriggerMatrix.RegisterTrigger("context.transform", func(trigger tspeak.ATMTrigger) error {
		return co.handleContextTransformation(trigger)
	})
	
	// Context mapping neural pathway (to/from maps)
	co.neuralTriggerMatrix.RegisterTrigger("context.to_map", func(trigger tspeak.ATMTrigger) error {
		return co.handleContextToMap(trigger)
	})
	
	// Context creation neural pathway (from maps)
	co.neuralTriggerMatrix.RegisterTrigger("context.from_map", func(trigger tspeak.ATMTrigger) error {
		return co.handleContextFromMap(trigger)
	})
	
	// Neural logging pathway
	co.neuralTriggerMatrix.RegisterTrigger("context.log", func(trigger tspeak.ATMTrigger) error {
		return co.handleNeuralLogging(trigger)
	})
}

// ProcessContext - Main neural processing function for context operations
func (co *ContextOrchestrator) ProcessContext(operation string, data map[string]interface{}) error {
	// Update neural activity
	co.updateNeuralActivity(operation)
	
	// Create ATM trigger
	trigger := tspeak.ATMTrigger{
		Who:           "ContextOrchestrator",
		What:          "ProcessContext_" + operation,
		When:          time.Now().Unix(),
		Where:         "SystemLayer6_Integration",
		Why:           "Context_Processing_Operation",
		How:           "Neural_Trigger_Matrix",
		Extent:        "Context_Operation",
		TriggerType:   operation,
		Priority:      tspeak.PriorityAwareness,
		BloodCellType: tspeak.BloodCellRed,
		TargetSystem:  "context",
		Payload:       data,
		Compressed:    false,
		DNA_ID:        "context_orchestrator_dna",
	}
	
	// Fire neural pathway
	err := co.neuralTriggerMatrix.ProcessTrigger(trigger)
	
	// Update metrics
	co.synapticMutex.Lock()
	co.contextProcessed++
	if err != nil {
		co.errorCount++
	}
	co.synapticMutex.Unlock()
	
	return err
}

// Neural pathway handlers - these implement the biological neural processing

func (co *ContextOrchestrator) handleContextStorage(trigger tspeak.ATMTrigger) error {
	contextID, ok := trigger.Payload["context_id"].(string)
	if !ok {
		return fmt.Errorf("neural storage failed: missing context_id")
	}
	
	contextVector, ok := trigger.Payload["context"].(ContextVector7D)
	if !ok {
		return fmt.Errorf("neural storage failed: invalid context vector")
	}
	
	// Store in neural memory
	co.memoryMutex.Lock()
	co.contextMemory[contextID] = contextVector
	co.neuralActivity[contextID] = time.Now()
	co.memoryMutex.Unlock()
	
	return nil
}

func (co *ContextOrchestrator) handleContextRetrieval(trigger tspeak.ATMTrigger) error {
	contextID, ok := trigger.Payload["context_id"].(string)
	if !ok {
		return fmt.Errorf("neural retrieval failed: missing context_id")
	}
	co.memoryMutex.RLock()
	contextVector, exists := co.contextMemory[contextID]
	lastActivity := co.neuralActivity[contextID]
	co.memoryMutex.RUnlock()
	if !exists {
		return fmt.Errorf("context not found in neural memory: %s", contextID)
	}
	_ = contextVector
	_ = lastActivity
	// Optionally: log or process contextVector, lastActivity
	return nil
}

func (co *ContextOrchestrator) handleContextTransformation(trigger tspeak.ATMTrigger) error {
	transformationType, _ := trigger.Payload["transformation"].(string)
	sourceContext, ok := trigger.Payload["source"].(ContextVector7D)
	if !ok {
		return fmt.Errorf("neural transformation failed: invalid source context")
	}
	// Apply neural transformation based on type
	var transformedContext ContextVector7D
	switch transformationType {
	case "mcp_to_tnos":
		transformedContext = co.transformMCPToTNOS(sourceContext, trigger.Payload)
	case "tnos_to_mcp":
		transformedContext = co.transformTNOSToMCP(sourceContext, trigger.Payload)
	default:
		transformedContext = sourceContext // Identity transformation
	}
	_ = transformedContext
	// Optionally: log or process transformedContext
	return nil
}

func (co *ContextOrchestrator) handleContextToMap(trigger tspeak.ATMTrigger) error {
	contextVector, ok := trigger.Payload["context"].(ContextVector7D)
	if !ok {
		return fmt.Errorf("neural mapping failed: invalid context vector")
	}
	// Convert to map using neural processing
	resultMap := map[string]interface{}{
		"who":    contextVector.Who,
		"what":   contextVector.What,
		"when":   contextVector.When,
		"where":  contextVector.Where,
		"why":    contextVector.Why,
		"how":    contextVector.How,
		"extent": contextVector.Extent,
	}
	if contextVector.Source != "" {
		resultMap["source"] = contextVector.Source
	}
	if len(contextVector.Meta) > 0 {
		resultMap["meta"] = contextVector.Meta
	}
	// Optionally: log or process resultMap
	return nil
}

func (co *ContextOrchestrator) handleContextFromMap(trigger tspeak.ATMTrigger) error {
	contextMap, ok := trigger.Payload["map"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("neural creation failed: invalid map data")
	}
	contextVector := ContextVector7D{
		Who:    co.safeGetString(contextMap, "who", "Unknown"),
		What:   co.safeGetString(contextMap, "what", "Unknown"),
		When:   contextMap["when"],
		Where:  co.safeGetString(contextMap, "where", "Unknown"),
		Why:    co.safeGetString(contextMap, "why", "Unknown"),
		How:    co.safeGetString(contextMap, "how", "Unknown"),
		Extent: co.safeGetFloat64(contextMap, "extent", 0.0),
		Source: co.safeGetString(contextMap, "source", ""),
	}
	if meta, exists := contextMap["meta"]; exists {
		if metaMap, ok := meta.(map[string]interface{}); ok {
			contextVector.Meta = metaMap
		}
	}
	// Optionally: log or process contextVector
	return nil
}

func (co *ContextOrchestrator) handleNeuralLogging(trigger tspeak.ATMTrigger) error {
	message, _ := trigger.Payload["message"].(string)
	metadata, _ := trigger.Payload["metadata"].(map[string]interface{})
	co.logNeuralActivity(message, metadata)
	return nil
}

// Neural utility functions

func (co *ContextOrchestrator) updateNeuralActivity(operation string) {
	co.synapticMutex.Lock()
	co.lastActivity = time.Now()
	co.synapticMutex.Unlock()
}

func (co *ContextOrchestrator) logNeuralActivity(message string, metadata map[string]interface{}) {
	if co.logger != nil {
		co.logger.InfoWithIdentity("ContextOrchestrator", message, metadata)
	}
}

func (co *ContextOrchestrator) transformMCPToTNOS(source ContextVector7D, data map[string]interface{}) ContextVector7D {
	// Neural transformation from MCP format to TNOS format
	transformed := source
	transformed.Source = "mcp_transformed"
	if transformed.Meta == nil {
		transformed.Meta = make(map[string]interface{})
	}
	transformed.Meta["transformation"] = "mcp_to_tnos"
	transformed.Meta["timestamp"] = time.Now().Unix()
	return transformed
}

func (co *ContextOrchestrator) transformTNOSToMCP(source ContextVector7D, data map[string]interface{}) ContextVector7D {
	// Neural transformation from TNOS format to MCP format  
	transformed := source
	transformed.Source = "tnos_transformed"
	if transformed.Meta == nil {
		transformed.Meta = make(map[string]interface{})
	}
	transformed.Meta["transformation"] = "tnos_to_mcp"
	transformed.Meta["timestamp"] = time.Now().Unix()
	return transformed
}

func (co *ContextOrchestrator) safeGetString(m map[string]interface{}, key, defaultVal string) string {
	if val, exists := m[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultVal
}

func (co *ContextOrchestrator) safeGetFloat64(m map[string]interface{}, key string, defaultVal float64) float64 {
	if val, exists := m[key]; exists {
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
	return defaultVal
}

// GetNeuralMetrics returns current neural system metrics
func (co *ContextOrchestrator) GetNeuralMetrics() map[string]interface{} {
	co.synapticMutex.RLock()
	defer co.synapticMutex.RUnlock()
	
	co.memoryMutex.RLock()
	memoryCount := len(co.contextMemory)
	connectionCount := len(co.synapticConnections)
	co.memoryMutex.RUnlock()
	
	return map[string]interface{}{
		"contexts_processed":    co.contextProcessed,
		"neural_errors":        co.errorCount,
		"memory_contexts":      memoryCount,
		"synaptic_connections": connectionCount,
		"last_activity":        co.lastActivity.Unix(),
		"biological_system":    "neural_context_orchestrator",
	}
}
