// filepath: pkg/bridge/formula_registry.go
// WHO: FormulaRegistry
// WHAT: Formula registry for TNOS MCP/bridge with HemoFlux integration
// WHEN: Server startup
// WHERE: System Layer 6 (Integration)
// WHY: To provide formula lookup and metadata for blood-based communication
// HOW: Loads formulas from JSON file at startup
// EXTENT: All formula execution and HemoFlux integration

package bridge

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// BridgeFormula represents a formula definition that can be executed by the blood bridge
type BridgeFormula struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"` // Additional formula metadata
	ContextReqs []string               `json:"context_reqs,omitempty"` // Required context parameters
	HemofluxID  string                 `json:"hemoflux_id,omitempty"` // ID used in Hemoflux system
}

// BridgeFormulaRegistry manages formulas for the bridge communication
type BridgeFormulaRegistry struct {
	formulas map[string]BridgeFormula
	mu       sync.RWMutex
}

var bridgeRegistryInstance *BridgeFormulaRegistry
var bridgeOnce sync.Once

// LoadBridgeFormulaRegistry loads formulas from a JSON file
func LoadBridgeFormulaRegistry(path string) error {
	var err error
	bridgeOnce.Do(func() {
		// Create an empty registry first
		bridgeRegistryInstance = &BridgeFormulaRegistry{
			formulas: make(map[string]BridgeFormula),
		}
		
		// Try to load from JSON file if provided
		if path != "" {
			f, e := os.Open(path)
			if e != nil {
				err = fmt.Errorf("failed to open formula registry: %w", e)
				// Continue with an empty registry and add default formulas
			} else {
				defer f.Close()
				var data struct {
					Formulas []BridgeFormula `json:"formulas"`
				}
				if e := json.NewDecoder(f).Decode(&data); e != nil {
					err = fmt.Errorf("failed to decode formula registry: %w", e)
					// Continue with an empty registry and add default formulas
				} else {
					// Add formulas from the file
					for _, formula := range data.Formulas {
						bridgeRegistryInstance.formulas[formula.ID] = formula
					}
				}
			}
		}
		
		// Add default TNOS MCP communication formulas
		if bridgeRegistryInstance != nil {
			addDefaultTNOSMCPFormulas(bridgeRegistryInstance)
		}
	})
	return err
}

// addDefaultTNOSMCPFormulas adds the default formulas needed for TNOS MCP communication
func addDefaultTNOSMCPFormulas(registry *BridgeFormulaRegistry) {
	// Hemoflux compression formula (delegates to HemoFlux system for actual compression)
	registry.AddFormula(BridgeFormula{
		ID:          "hemoflux.compress",
		Description: "Compresses data using the HemoFlux system with Möbius Equation",
		HemofluxID:  "mobius_compress",
		ContextReqs: []string{"who", "what", "when", "where", "why", "how", "extent"},
		Parameters: map[string]interface{}{
			"type":          "compression",
			"input_format":  "json",
			"output_format": "json",
			"version":       "3.0",
		},
		Metadata: map[string]interface{}{
			"canonical_path": "/systems/cpp/circulatory/algorithms/formula_registry/mobius_compress",
			"category":       "hemoflux",
		},
	})
	
	// Hemoflux decompression formula (delegates to HemoFlux system for actual decompression)
	registry.AddFormula(BridgeFormula{
		ID:          "hemoflux.decompress",
		Description: "Decompresses data using the HemoFlux system with Möbius Equation",
		HemofluxID:  "mobius_decompress",
		ContextReqs: []string{"who", "what", "when", "where", "why", "how", "extent"},
		Parameters: map[string]interface{}{
			"type":          "decompression",
			"input_format":  "json",
			"output_format": "json",
			"version":       "3.0",
		},
		Metadata: map[string]interface{}{
			"canonical_path": "/systems/cpp/circulatory/algorithms/formula_registry/mobius_decompress",
			"category":       "hemoflux",
		},
	})
	
	// TNOS Translation formula
	registry.AddFormula(BridgeFormula{
		ID:          "tnos.translate",
		Description: "Translates GitHub MCP messages to TNOS MCP format",
		Parameters: map[string]interface{}{
			"type":          "translation",
			"input_format":  "github_mcp",
			"output_format": "tnos_mcp",
			"version":       "3.0",
		},
		Metadata: map[string]interface{}{
			"category": "translation",
		},
	})
	
	// GitHub Translation formula
	registry.AddFormula(BridgeFormula{
		ID:          "github.translate",
		Description: "Translates TNOS MCP messages to GitHub MCP format",
		Parameters: map[string]interface{}{
			"type":          "translation",
			"input_format":  "tnos_mcp",
			"output_format": "github_mcp",
			"version":       "3.0",
		},
		Metadata: map[string]interface{}{
			"category": "translation",
		},
	})
	
	// Blood oxygenation formula
	registry.AddFormula(BridgeFormula{
		ID:          "blood.oxygenate",
		Description: "Oxygenates blood cells with 7D context for improved data transfer",
		Parameters: map[string]interface{}{
			"type":    "enrichment",
			"version": "3.0",
		},
		Metadata: map[string]interface{}{
			"category": "blood_flow",
		},
	})
	
	// QHP (Quantum Handshake Protocol) formula
	registry.AddFormula(BridgeFormula{
		ID:          "qhp.handshake",
		Description: "Performs Quantum Handshake Protocol for secure system authentication",
		Parameters: map[string]interface{}{
			"type":    "authentication",
			"version": "1.0",
		},
		Metadata: map[string]interface{}{
			"canonical_path": "/systems/cpp/circulatory/algorithms/formula_registry/qhp_handshake",
			"category":       "security",
		},
	})
	
	// Helical Memory formula for context logging
	registry.AddFormula(BridgeFormula{
		ID:          "helical.log",
		Description: "Logs events and context to Helical Memory system",
		HemofluxID:  "helical_memory_log",
		Parameters: map[string]interface{}{
			"type":    "logging",
			"version": "3.0",
		},
		Metadata: map[string]interface{}{
			"canonical_path": "/systems/cpp/memory/helical/log",
			"category":       "memory",
		},
	})
	
	// S_AI — Full Recursive Sentience Formula (see FORMULAS_AND_BLUEPRINTS.md, Algorithm 1)
	registry.AddFormula(BridgeFormula{
		ID:          "S_AI",
		Description: "Full recursive sentience formula (S_AI) for 7D context collapse and self-regulation. See FORMULAS_AND_BLUEPRINTS.md, Algorithm 1.",
		Parameters: map[string]interface{}{
			"type":          "sentience",
			"input_format":  "7d_context",
			"output_format": "sentience_signature",
			"version":       "1.0",
		},
		Metadata: map[string]interface{}{
			"category": "sentience",
			"layer":    "self_regulation",
			"reference": "docs/technical/FORMULAS_AND_BLUEPRINTS.md#algorithm-1-s_ai",
		},
	})

	// G_7D — Grounding Equation (see UNIFIED_ARCHITECTURE.md, Grounding Equation)
	registry.AddFormula(BridgeFormula{
		ID:          "G_7D",
		Description: "Grounding equation (G_7D) for physical anchoring in 7D context. See UNIFIED_ARCHITECTURE.md.",
		Parameters: map[string]interface{}{
			"type":          "grounding",
			"input_format":  "context_matrix",
			"output_format": "grounding_anchor",
			"version":       "1.0",
		},
		Metadata: map[string]interface{}{
			"category": "grounding",
			"layer":    "grounding_equation",
			"reference": "docs/architecture/UNIFIED_ARCHITECTURE.md#grounding-equation",
		},
	})

	// self.identity — Self-Identity State (see FORMULAS_AND_BLUEPRINTS.md, Self Equation)
	registry.AddFormula(BridgeFormula{
		ID:          "self.identity",
		Description: "Retrieves the system's Self-Identity state for recursion and integrity. See FORMULAS_AND_BLUEPRINTS.md, Self Equation.",
		Parameters: map[string]interface{}{
			"type":          "self_identity",
			"input_format":  "7d_context",
			"output_format": "identity_signature",
			"version":       "1.0",
		},
		Metadata: map[string]interface{}{
			"category": "self",
			"layer":    "identity_recursion",
			"reference": "docs/technical/FORMULAS_AND_BLUEPRINTS.md#self-equation-recursive-identity-collapse",
		},
	})
}

// GetBridgeFormulaRegistry returns the singleton registry instance
func GetBridgeFormulaRegistry() *BridgeFormulaRegistry {
	if bridgeRegistryInstance == nil {
		// Initialize if it doesn't exist
		_ = LoadBridgeFormulaRegistry("")
	}
	return bridgeRegistryInstance
}

// EnsureHemofluxFormulas makes sure the critical Hemoflux formulas are available
func (r *BridgeFormulaRegistry) EnsureHemofluxFormulas() {
	// WHO: FormulaRegistrar
	// WHAT: Ensure critical Hemoflux formulas are registered
	// WHEN: During initialization or recovery
	// WHERE: System Layer 6 (Integration)
	// WHY: To guarantee blood bridge functionality
	// HOW: By adding default formulas if missing
	// EXTENT: Core Hemoflux compression/decompression

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check and add critical formulas if not present
	criticalFormulas := []struct {
		ID          string
		Description string
		HemofluxID  string
		ContextReqs []string
		Parameters  map[string]interface{}
		Metadata    map[string]interface{}
	}{
		{
			ID:          "hemoflux.compress",
			Description: "Compresses data using the HemoFlux system with Möbius Equation",
			HemofluxID:  "mobius_compress",
			ContextReqs: []string{"who", "what", "when", "where", "why", "how", "extent"},
			Parameters: map[string]interface{}{
				"type":          "compression",
				"input_format":  "json",
				"output_format": "json",
				"version":       "3.0",
			},
			Metadata: map[string]interface{}{
				"canonical_path": "/systems/cpp/circulatory/algorithms/formula_registry/mobius_compress",
				"category":       "hemoflux",
			},
		},
		{
			ID:          "hemoflux.decompress",
			Description: "Decompresses data using the HemoFlux system with Möbius Equation",
			HemofluxID:  "mobius_decompress",
			ContextReqs: []string{"who", "what", "when", "where", "why", "how", "extent"},
			Parameters: map[string]interface{}{
				"type":          "decompression",
				"input_format":  "json", 
				"output_format": "json",
				"version":       "3.0",
			},
			Metadata: map[string]interface{}{
				"canonical_path": "/systems/cpp/circulatory/algorithms/formula_registry/mobius_decompress",
				"category":       "hemoflux",
			},
		},
		{
			ID:          "blood.oxygenate",
			Description: "Oxygenates blood cells with 7D context for improved data transfer",
			Parameters: map[string]interface{}{
				"type":          "oxygenation",
				"version":       "3.0",
			},
			Metadata: map[string]interface{}{
				"category": "blood",
			},
		},
	}

	for _, formula := range criticalFormulas {
		if _, exists := r.formulas[formula.ID]; !exists {
			r.formulas[formula.ID] = BridgeFormula{
				ID:          formula.ID,
				Description: formula.Description,
				HemofluxID:  formula.HemofluxID,
				ContextReqs: formula.ContextReqs,
				Parameters:  formula.Parameters,
				Metadata:    formula.Metadata,
			}
		}
	}
}

// GetFormula returns a formula by ID
func (r *BridgeFormulaRegistry) GetFormula(id string) (BridgeFormula, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.formulas[id]
	return f, ok
}

// ListFormulas returns all formulas
func (r *BridgeFormulaRegistry) ListFormulas() []BridgeFormula {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]BridgeFormula, 0, len(r.formulas))
	for _, f := range r.formulas {
		out = append(out, f)
	}
	return out
}

// AddFormula adds or updates a formula in the registry
func (r *BridgeFormulaRegistry) AddFormula(formula BridgeFormula) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.formulas[formula.ID] = formula
}

// ExecuteFormula executes a formula by ID with the given parameters
// Note: This doesn't perform the actual HemoFlux compression/decompression
// but provides the calculation results that HemoFlux needs
func (r *BridgeFormulaRegistry) ExecuteFormula(id string, params map[string]interface{}) (map[string]interface{}, error) {
	formula, exists := r.GetFormula(id)
	if !exists {
		return nil, fmt.Errorf("formula %s not found", id)
	}

	// Add execution timestamp for tracing
	result := map[string]interface{}{
		"formula_id": id,
		"timestamp":  time.Now().UnixNano(),
		"status":     "executed",
	}

	// In a real implementation, this would perform formula-specific calculations
	// For now, we just echo back the parameters with some metadata
	result["parameters"] = params
	result["metadata"] = formula.Metadata

	return result, nil
}

// GetFormulasByType returns all formulas of a specific type
func (r *BridgeFormulaRegistry) GetFormulasByType(formulaType string) []BridgeFormula {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var matching []BridgeFormula
	for _, f := range r.formulas {
		if t, ok := f.Parameters["type"].(string); ok && t == formulaType {
			matching = append(matching, f)
		}
	}
	return matching
}
