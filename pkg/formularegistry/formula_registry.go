// filepath: pkg/bridge/formula_registry.go
// WHO: FormulaRegistry
// WHAT: Formula registry for TNOS MCP/bridge with HemoFlux integration
// WHEN: Server startup
// WHERE: System Layer 6 (Integration)
// WHY: To provide formula lookup and metadata for blood-based communication
// HOW: Loads formulas from JSON file at startup
// EXTENT: All formula execution and HemoFlux integration

// TRANQUILSPEAK SYMBOL CLUSTER: [𝒯🦴ζℏƒ𓆑#SK1𝑾𝑾𝑯𝑾𝑯𝑬𝑾𝑯𝑬𝑹𝑾𝑯𝒀𝑯𝑶𝑾𝑬𝑿⏳📍𝒮𝓔𝓗]
// This file is part of the 'skeletal' biosystem. See circulatory/github-mcp-server/symbolic_mapping_registry_autogen_20250603.tsq for details.

package formularegistry

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// SymbolMapping holds the mapping from formula names to canonical TranquilSpeak symbols
var SymbolMapping = make(map[string]string)

// FormulaRegistryData represents the structure of formulas.json
type FormulaRegistryData struct {
	RegistryVersion     string `json:"registry_version"`
	LastUpdated         string `json:"last_updated"`
	CanonicalReferences []string `json:"canonical_references"`
	Formulas           []FormulaData `json:"formulas"`
}

// FormulaData represents a single formula entry from formulas.json
type FormulaData struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	Inputs              []string `json:"inputs"`
	Outputs             []string `json:"outputs"`
	Tags                []string `json:"tags"`
	Reference           string   `json:"reference"`
	TranquilSpeakSymbol string   `json:"tranquilspeak_symbol"`
}

// LoadSymbolMapping loads the TranquilSpeak symbol mapping from the canonical mapping file
func LoadSymbolMapping(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open symbol mapping file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "|") && strings.Contains(line, "|") {
			fields := strings.Split(line, "|")
			if len(fields) > 5 {
				formulaName := strings.TrimSpace(fields[1])
				tranquilSym := strings.TrimSpace(fields[5])
				if formulaName != "" && tranquilSym != "" {
					SymbolMapping[formulaName] = tranquilSym
				}
			}
		}
	}
	return scanner.Err()
}

// LoadSymbolMappingFromJSON loads TranquilSpeak symbol mapping from formulas.json
func LoadSymbolMappingFromJSON(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open formulas.json: %w", err)
	}
	defer file.Close()

	var data FormulaRegistryData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return fmt.Errorf("failed to decode formulas.json: %w", err)
	}

	// Build symbol mapping from JSON data
	for _, formula := range data.Formulas {
		if formula.TranquilSpeakSymbol != "" {
			SymbolMapping[formula.ID] = formula.TranquilSpeakSymbol
			fmt.Printf("Loaded formula symbol mapping: %s -> %s\n", formula.ID, formula.TranquilSpeakSymbol)
		}
	}

	fmt.Printf("Loaded %d TranquilSpeak symbol mappings from %s\n", len(SymbolMapping), path)
	return nil
}

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
			// First, load the symbol mapping from the same JSON file
			if e := LoadSymbolMappingFromJSON(path); e != nil {
				fmt.Printf("Warning: failed to load symbol mapping from JSON: %v\n", e)
			}
			
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
	
	// TranquilSpeak logger initialization formula
	registry.AddFormula(BridgeFormula{
		ID:          "tranquilspeak.initializeLogger",
		Description: "Initializes the TranquilSpeak logger with 7D context integration",
		ContextReqs: []string{"who", "what", "when", "where", "why", "how", "extent"},
		Parameters: map[string]interface{}{ "logLevel": "info" },
		Metadata: map[string]interface{}{ "category": "tranquilspeak", "canonical_path": "/systems/tranquilspeak/logger/initialize" },
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

// GetFormulaBySymbol returns a formula by its TranquilSpeak symbol
func (r *BridgeFormulaRegistry) GetFormulaBySymbol(symbol string) (BridgeFormula, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Find formula by matching TranquilSpeak symbol
	for _, formula := range r.formulas {
		if mappedSymbol, exists := SymbolMapping[formula.ID]; exists && mappedSymbol == symbol {
			return formula, true
		}
	}
	return BridgeFormula{}, false
}

// GetFormulaByName returns a formula by its canonical name using the symbol mapping
func (r *BridgeFormulaRegistry) GetFormulaByName(name string) (BridgeFormula, bool) {
	symbol, ok := SymbolMapping[name]
	if !ok {
		return BridgeFormula{}, false
	}
	return r.GetFormulaBySymbol(symbol)
}

// ExecuteFormula executes a formula by canonical symbol with the given parameters
func (r *BridgeFormulaRegistry) ExecuteFormula(symbol string, params map[string]interface{}) (map[string]interface{}, error) {
	formula, exists := r.GetFormula(symbol)
	if !exists {
		return nil, fmt.Errorf("formula %s not found", symbol)
	}
	fn, ok := FormulaFuncRegistry[symbol]
	if !ok {
		return nil, fmt.Errorf("no computation function registered for symbol %s", symbol)
	}
	result, err := fn(params)
	if err != nil {
		return nil, err
	}
	result["formula_id"] = symbol
	result["timestamp"] = time.Now().UnixNano()
	result["metadata"] = formula.Metadata
	return result, nil
}

// ExecuteFormulaByName executes a formula by its canonical name using the symbol mapping
func (r *BridgeFormulaRegistry) ExecuteFormulaByName(name string, params map[string]interface{}) (map[string]interface{}, error) {
	symbol, ok := SymbolMapping[name]
	if !ok {
		return nil, fmt.Errorf("symbol mapping not found for formula name: %s", name)
	}
	return r.ExecuteFormula(symbol, params)
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

// FormulaFunc is a function that performs the actual computation for a formula
// It receives the formula parameters and returns the result or error
// You can register new formulas by adding to the FormulaFuncRegistry
// Example: FormulaFuncRegistry["hemoflux.compress"] = HemofluxCompressFunc
// All keys should be canonical TranquilSpeak symbols
var FormulaFuncRegistry = make(map[string]func(map[string]interface{}) (map[string]interface{}, error))

// RegisterFormulaFunc registers a computation function for a formula symbol
func RegisterFormulaFunc(symbol string, fn func(map[string]interface{}) (map[string]interface{}, error)) {
	FormulaFuncRegistry[symbol] = fn
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

// Canonical: Helical Parity and Recovery Functions
// Implements dual-helix parity and self-healing logic per HDSA/AI-DNA/7D/ATM spec

func helicalParityFunc(params map[string]interface{}) (map[string]interface{}, error) {
	primaryBytesRaw, ok := params["primary"]
	if !ok {
		return nil, fmt.Errorf("missing 'primary' parameter for helical.parity")
	}
	var primaryBytes []byte
	switch v := primaryBytesRaw.(type) {
	case []byte:
		primaryBytes = v
	case string:
		primaryBytes = []byte(v)
	default:
		return nil, fmt.Errorf("invalid type for 'primary' parameter")
	}
	// Canonical: Parity is SHA256 of primary, for deterministic self-healing
	parity := sha256.Sum256(primaryBytes)
	return map[string]interface{}{"parity": parity[:]}, nil
}

func helicalRecoverPrimaryFunc(params map[string]interface{}) (map[string]interface{}, error) {
	parityBytesRaw, ok := params["parity"]
	if !ok {
		return nil, fmt.Errorf("missing 'parity' parameter for helical.recover_primary")
	}
	var parityBytes []byte
	switch v := parityBytesRaw.(type) {
	case []byte:
		parityBytes = v
	case string:
		parityBytes = []byte(v)
	default:
		return nil, fmt.Errorf("invalid type for 'parity' parameter")
	}
	// Canonical: In real HDSA, this would use dual-helix error correction; here, just return the hash as a stand-in
	return map[string]interface{}{"primary": parityBytes}, nil
}

func init() {
	// Load canonical symbol mapping if not already loaded
	if len(SymbolMapping) == 0 {
		// Try to load from formulas.json first (preferred)
		if err := LoadSymbolMappingFromJSON("config/formulas.json"); err != nil {
			// Fallback to legacy .tsq file
			_ = LoadSymbolMapping("/Users/Jubicudis/Tranquility-Neuro-OS/systems/tranquilspeak/circulatory/github-mcp-server/symbolic_mapping_registry_autogen_20250603.tsq")
		}
	}
	// Register canonical formula functions using the correct symbols
	if sym, ok := SymbolMapping["helical.parity"]; ok {
		RegisterFormulaFunc(sym, helicalParityFunc)
	}
	if sym, ok := SymbolMapping["helical.recover_primary"]; ok {
		RegisterFormulaFunc(sym, helicalRecoverPrimaryFunc)
	}
}

// [TNOS] All formula registration and execution must use TranquilSpeak symbols as canonical keys.
// See: ../../../../systems/cpp/circulatory/algorithms/data/tranquilspeak_symbol_key.md for the authoritative mapping.
// Example: {"helical.parity": "ⓗ"} (replace with actual symbol from mapping)
//
// Usage: Always pass the symbol, not a hardcoded name, to AddFormula/ExecuteFormula.

// Example symbol mapping (should be auto-generated or imported in production)
var FORMULA_SYMBOLS = map[string]string{
	"helical.parity": "ⓗ", // TODO: Replace with actual symbol from mapping
	"helical.recover_primary": "ⓡ", // TODO: Replace with actual symbol from mapping
	// ...add more as needed
}

// In all usages, replace calls like GetFormula("helical.parity") with GetFormula(FORMULA_SYMBOLS["helical.parity"])
// and ExecuteFormula("helical.parity", ...) with ExecuteFormula(FORMULA_SYMBOLS["helical.parity"], ...)
