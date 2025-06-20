// TRANQUILSPEAK SYMBOL CLUSTER: [𝒯🧬ζℏƒ𓆑#CT1𝑪𝑪𝑶𝑪𝑶𝑵𝑪𝑶𝑵𝑻𝑪𝑶𝑵𝑬𝑪𝑶𝑵𝑬𝑿𝑬𝑿⏳📍𝒮𝓔𝓗]
// HEMOFLUX_FILE_ID: "_USERS_JUBICUDIS_TRANQUILITY-NEURO-OS_GITHUB-MCP-SERVER_PKG_CONTEXT_CONTEXT_TRANSLATOR.GO"
// HEMOFLUX_FORMULA: mobius.7d.collapse
/*
 * WHO: ContextTranslator7D
 * WHAT: 7D Context translation and compression engine using ATM triggers
 * WHEN: During all inter-system communication
 * WHERE: System Layer 6 (Integration)
 * WHY: To enable seamless context flow between systems using ATM-triggered compression
 * HOW: Using 7D Recursive Collapse Model with ATM integration (no direct imports)
 * EXTENT: All context translation operations
 */

package context

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log"
	tspeak "github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/tranquilspeak"
)

// Context7DEngine - The core 7D context translation system following TNOS architecture
// This engine orchestrates context operations via ATM triggers while maintaining LoggerInterface integration
type Context7DEngine struct {
	logger         log.LoggerInterface  // Essential for ATM logging integration
	triggerMatrix  *tspeak.TriggerMatrix   // ATM trigger system
	enableCompression bool
	mu             sync.RWMutex
	successCount   int64
	failureCount   int64
}

// NewContext7DEngine creates a new 7D translation engine using ATM triggers only
func NewContext7DEngine(logger log.LoggerInterface) *Context7DEngine {
	engine := &Context7DEngine{
		logger:            logger,
		triggerMatrix:     tspeak.NewTriggerMatrix(),
		enableCompression: true,
	}
	
	// Register ATM triggers for context operations
	engine.registerATMTriggers()
	
	// Log initialization via ATM trigger
	if engine.logger != nil {
		engine.logger.Info("Context7D Engine initialized with ATM triggers")
	}
	
	return engine
}

// registerATMTriggers sets up all ATM trigger handlers for context operations
func (engine *Context7DEngine) registerATMTriggers() {
	// Register compression trigger
	engine.triggerMatrix.RegisterTrigger("context.compress", func(trigger tspeak.ATMTrigger) error {
		contextData, ok := trigger.Payload["context"].(log.ContextVector7D)
		if !ok {
			return fmt.Errorf("invalid context data type")
		}
		_ = engine.compressContextViaATM(contextData)
		// Optionally, you could log or store the compressed result here
		return nil
	})

	// Register to_map trigger
	engine.triggerMatrix.RegisterTrigger("context.to_map", func(trigger tspeak.ATMTrigger) error {
		contextData, ok := trigger.Payload["context"].(log.ContextVector7D)
		if !ok {
			return fmt.Errorf("invalid context data type")
		}
		_ = log.ToMap(contextData)
		// Optionally, you could log or store the result here
		return nil
	})
}

// TranslateFromMCPContext translates GitHub MCP context to TNOS 7D context using ATM triggers
func (engine *Context7DEngine) TranslateFromMCPContext(githubContext map[string]interface{}) (map[string]interface{}, error) {
	// Convert GitHub context to TNOS 7D context
	tnosContext := engine.mcpToTNOSContext(githubContext)

	// Apply compression via ATM trigger if enabled
	if engine.enableCompression {
		trigger := engine.triggerMatrix.CreateTrigger(
			"Context7DEngine", "context.compress", "context_translator", "compress context", "atm_trigger", "context_compression", "context.compress", "context", map[string]interface{}{
				"context": log.FromMap(tnosContext),
			},
		)
		err := engine.triggerMatrix.ProcessTrigger(trigger)
		if err == nil {
			// Optionally, update tnosContext with compressed result if needed
		}
	}

	// Update success metrics
	engine.mu.Lock()
	engine.successCount++
	engine.mu.Unlock()

	return tnosContext, nil
}

// TranslateJSONContext processes JSON context data using ATM triggers
func (engine *Context7DEngine) TranslateJSONContext(jsonStr string) (map[string]interface{}, error) {
	// Parse JSON to ContextVector7D
	cv, err := engine.jsonToContextVector(jsonStr)
	if err != nil {
		engine.mu.Lock()
		engine.failureCount++
		engine.mu.Unlock()
		return nil, fmt.Errorf("failed to parse JSON context: %w", err)
	}

	// Apply compression via ATM trigger if enabled
	result := log.ToMap(cv)
	if engine.enableCompression {
		trigger := engine.triggerMatrix.CreateTrigger(
			"Context7DEngine", "context.compress", "context_translator", "compress context", "atm_trigger", "context_compression", "context.compress", "context", map[string]interface{}{
				"context": cv,
			},
		)
		err := engine.triggerMatrix.ProcessTrigger(trigger)
		if err == nil {
			// Optionally, update result with compressed result if needed
		}
	}

	// Update success metrics
	engine.mu.Lock()
	engine.successCount++
	engine.mu.Unlock()

	return result, nil
}

// Helper methods

// compressContextViaATM compresses context using ATM trigger patterns
func (engine *Context7DEngine) compressContextViaATM(cv log.ContextVector7D) map[string]interface{} {
	compressed := map[string]interface{}{
		"w": cv.Who,
		"wh": cv.What,
		"wn": cv.When,
		"wr": cv.Where,
		"wy": cv.Why,
		"hw": cv.How,
		"e": cv.Extent,
		"compressed": true,
		"algorithm": "mobius_7d",
	}
	
	if cv.Source != "" {
		compressed["s"] = cv.Source
	}
	
	if len(cv.Meta) > 0 {
		compressed["m"] = cv.Meta
	}
	
	return compressed
}

// mcpToTNOSContext converts GitHub MCP context to TNOS format
func (engine *Context7DEngine) mcpToTNOSContext(githubContext map[string]interface{}) map[string]interface{} {
	tnosContext := map[string]interface{}{
		"who":    "GitHubMCPServer",
		"what":   "context_translation",
		"when":   time.Now().Unix(),
		"where":  "mcp_bridge",
		"why":    "enable_github_integration",
		"how":    "7d_translation",
		"extent": 1.0,
		"source": "github_mcp",
		"meta": githubContext,
	}
	
	return tnosContext
}

// jsonToContextVector converts JSON string to ContextVector7D
func (engine *Context7DEngine) jsonToContextVector(jsonStr string) (log.ContextVector7D, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return log.ContextVector7D{}, err
	}
	
	return log.FromMap(data), nil
}

// GetMetrics returns current engine metrics
func (engine *Context7DEngine) GetMetrics() map[string]interface{} {
	engine.mu.RLock()
	defer engine.mu.RUnlock()
	
	return map[string]interface{}{
		"success_count": engine.successCount,
		"failure_count": engine.failureCount,
		"compression_enabled": engine.enableCompression,
	}
}
