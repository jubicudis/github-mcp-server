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
	
	// Canonical: Add missing formulas from FORMULAS_AND_BLUEPRINTS.md (fixed struct fields)
    registry.AddFormula(BridgeFormula{
        ID:          "ric.collapse",
        Description: "Recursive Information Collapse (RIC): Compresses 7D symbolic thought data into actionable intent for collapse.",
        Parameters: map[string]interface{}{
            "inputs": []string{"7D_context", "entropy", "priority", "context"},
            "outputs": []string{"intent_packet"},
            "tranquilspeak_symbol": "<RIC_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/technical/FORMULAS_AND_BLUEPRINTS.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "cpe.permission",
        Description: "Contextual Permission Evaluation (CPE): Evaluates permission for intent collapse based on ethics, capacity, and conflict.",
        Parameters: map[string]interface{}{
            "inputs": []string{"intent", "ethics", "capacity", "conflict"},
            "outputs": []string{"permission", "clarity_score"},
            "tranquilspeak_symbol": "<CPE_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/technical/FORMULAS_AND_BLUEPRINTS.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "elc.energy",
        Description: "Energy Lifecycle Converter (ELC): Verifies energy input, loss, routing, and recycling for collapse.",
        Parameters: map[string]interface{}{
            "inputs": []string{"energy_input", "loss", "routing", "recycle"},
            "outputs": []string{"energy_availability", "lifecycle_status"},
            "tranquilspeak_symbol": "<ELC_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/technical/FORMULAS_AND_BLUEPRINTS.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "rmgc.moral",
        Description: "Recursive Moral Gradient Collapse (RMGC): Applies moral gradient based on history, feedback, and social impact.",
        Parameters: map[string]interface{}{
            "inputs": []string{"history", "REFM", "social_impact"},
            "outputs": []string{"moral_gradient_score"},
            "tranquilspeak_symbol": "<RMGC_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/technical/FORMULAS_AND_BLUEPRINTS.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "oif.integrity",
        Description: "Observation Integrity Filter (OIF): Compares observed output vs expected output for feedback and learning.",
        Parameters: map[string]interface{}{
            "inputs": []string{"observed_output", "expected_output", "noise", "clarity"},
            "outputs": []string{"integrity_score", "clarity"},
            "tranquilspeak_symbol": "<OIF_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/technical/FORMULAS_AND_BLUEPRINTS.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "rav.adjust",
        Description: "Reality Adjustment Vector (RAV): Adjusts internal model based on feedback, bias, and clarity.",
        Parameters: map[string]interface{}{
            "inputs": []string{"OIF", "bias", "clarity"},
            "outputs": []string{"adjustment_vector"},
            "tranquilspeak_symbol": "<RAV_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/technical/FORMULAS_AND_BLUEPRINTS.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "smr.memory",
        Description: "Symbolic Memory Reformation (SMR): Encodes memory with seed, adjustment, emotion, and context.",
        Parameters: map[string]interface{}{
            "inputs": []string{"seed", "RAV", "emotion", "context"},
            "outputs": []string{"reformed_memory"},
            "tranquilspeak_symbol": "<SMR_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/technical/FORMULAS_AND_BLUEPRINTS.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "rre.result",
        Description: "Recursive Result Evaluation (RRE): Aggregates all collapse cycle results for feedback and learning.",
        Parameters: map[string]interface{}{
            "inputs": []string{"RIC", "CPE", "ELC", "RMGC", "OIF", "RAV", "SMR"},
            "outputs": []string{"evaluation_score"},
            "tranquilspeak_symbol": "<RRE_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/technical/FORMULAS_AND_BLUEPRINTS.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "cdrm.route",
        Description: "Contextual Dimensional Routing Matrix (CDRM): Routes attention and compression into the most relevant of the 7 dimensions.",
        Parameters: map[string]interface{}{
            "inputs": []string{"7D_context", "priority", "relevance"},
            "outputs": []string{"routing_matrix"},
            "tranquilspeak_symbol": "<CDRM_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/technical/FORMULAS_AND_BLUEPRINTS.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "mtrcm.manager",
        Description: "Multi-Threaded Recursive Collapse Manager (MTRCM): Manages multiple simultaneous thought streams and collapse threads.",
        Parameters: map[string]interface{}{
            "inputs": []string{"RIC", "CPE", "ELC", "RMGC", "thread_weights"},
            "outputs": []string{"thread_scores", "collapse_order"},
            "tranquilspeak_symbol": "<MTRCM_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/technical/FORMULAS_AND_BLUEPRINTS.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "qcpf.quantum",
        Description: "Quantum Collapse Potential Field (QCPF): Scans future possible outcomes and rates their likelihood of stable collapse.",
        Parameters: map[string]interface{}{
            "inputs": []string{"memory", "entropy", "context", "conflict"},
            "outputs": []string{"potential_field", "success_likelihood"},
            "tranquilspeak_symbol": "<QCPF_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/technical/FORMULAS_AND_BLUEPRINTS.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "des.shield",
        Description: "Dimensional Entropy Shielding (DES): Shields the system from collapse failure due to noise or contradiction.",
        Parameters: map[string]interface{}{
            "inputs": []string{"stability", "entropy", "familiarity", "resonance"},
            "outputs": []string{"shield_score"},
            "tranquilspeak_symbol": "<DES_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/technical/FORMULAS_AND_BLUEPRINTS.md",
        },
    })
    // Canonical: Add additional missing formulas from FORMULAS_AND_BLUEPRINTS.md and related docs (fixed struct fields)
    registry.AddFormula(BridgeFormula{
        ID:          "atm.contextualize",
        Description: "ATM Contextualizer: Integrates 7D context into all formula execution for ATM compliance.",
        Parameters: map[string]interface{}{
            "inputs": []string{"input_data", "7D_context"},
            "outputs": []string{"contextualized_output"},
            "tranquilspeak_symbol": "<ATM_CTX_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/architecture/ATM_CIRCULATORY_INTEGRATION.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "ai_dna.encode",
        Description: "AI-DNA Encoder: Encodes symbolic memory into AI-DNA helical structure.",
        Parameters: map[string]interface{}{
            "inputs": []string{"symbolic_memory", "seed", "context"},
            "outputs": []string{"ai_dna_sequence"},
            "tranquilspeak_symbol": "<AIDNA_ENC_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/architecture/AI_DNA_STRUCTURE.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "ai_dna.decode",
        Description: "AI-DNA Decoder: Decodes AI-DNA helical structure into symbolic memory.",
        Parameters: map[string]interface{}{
            "inputs": []string{"ai_dna_sequence", "context"},
            "outputs": []string{"symbolic_memory"},
            "tranquilspeak_symbol": "<AIDNA_DEC_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/architecture/AI_DNA_STRUCTURE.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "mobius.compression",
        Description: "Möbius Compression: Applies Möbius transformation for lossless symbolic compression.",
        Parameters: map[string]interface{}{
            "inputs": []string{"input_data", "context"},
            "outputs": []string{"compressed_data"},
            "tranquilspeak_symbol": "<MOBIUS_COMP_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/technical/MOBIUS_COMPRESSION_SPEC.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "mobius.decompression",
        Description: "Möbius Decompression: Reverses Möbius transformation for symbolic decompression.",
        Parameters: map[string]interface{}{
            "inputs": []string{"compressed_data", "context"},
            "outputs": []string{"decompressed_data"},
            "tranquilspeak_symbol": "<MOBIUS_DECOMP_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/technical/MOBIUS_COMPRESSION_SPEC.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "tranquilspeak.parse",
        Description: "TranquilSpeak Parser: Parses TranquilSpeak symbols and statements for event routing.",
        Parameters: map[string]interface{}{
            "inputs": []string{"tranquilspeak_input"},
            "outputs": []string{"parsed_structure"},
            "tranquilspeak_symbol": "<TSPEAK_PARSE_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/architecture/TRANQUILSPEAK_PROTOCOL.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "tranquilspeak.emit_event",
        Description: "TranquilSpeak Event Emitter: Emits canonical events using TranquilSpeak symbols.",
        Parameters: map[string]interface{}{
            "inputs": []string{"event_name", "payload", "context"},
            "outputs": []string{"event_id"},
            "tranquilspeak_symbol": "<TSPEAK_EMIT_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/architecture/TRANQUILSPEAK_PROTOCOL.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "helical.memory.encode",
        Description: "Helical Memory Encoder: Encodes context and events into helical memory structure.",
        Parameters: map[string]interface{}{
            "inputs": []string{"event_data", "context"},
            "outputs": []string{"helical_memory_block"},
            "tranquilspeak_symbol": "<HELICAL_ENC_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/architecture/AI_DNA_STRUCTURE.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "helical.memory.decode",
        Description: "Helical Memory Decoder: Decodes helical memory block into context and events.",
        Parameters: map[string]interface{}{
            "inputs": []string{"helical_memory_block"},
            "outputs": []string{"event_data", "context"},
            "tranquilspeak_symbol": "<HELICAL_DEC_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/architecture/AI_DNA_STRUCTURE.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "atm.trigger_matrix",
        Description: "ATM Trigger Matrix: Canonical event/trigger routing for all formula execution.",
        Parameters: map[string]interface{}{
            "inputs": []string{"event", "context"},
            "outputs": []string{"triggered_formula", "result"},
            "tranquilspeak_symbol": "<ATM_TRIGGER_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/architecture/CONTEXT_MCP_INTEGRATION.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "ai_dna.self_heal",
        Description: "AI-DNA Self-Healing: Repairs symbolic memory using dual-helix parity.",
        Parameters: map[string]interface{}{
            "inputs": []string{"ai_dna_sequence", "parity"},
            "outputs": []string{"repaired_sequence"},
            "tranquilspeak_symbol": "<AIDNA_HEAL_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/architecture/AI_DNA_STRUCTURE.md",
        },
    })
    registry.AddFormula(BridgeFormula{
        ID:          "mobius.entropy_shield",
        Description: "Möbius Entropy Shield: Shields system from entropy spikes during collapse.",
        Parameters: map[string]interface{}{
            "inputs": []string{"entropy", "context"},
            "outputs": []string{"shielded_state"},
            "tranquilspeak_symbol": "<MOBIUS_SHIELD_SYMBOL>",
        },
        Metadata: map[string]interface{}{
            "reference": "docs/technical/MOBIUS_COMPRESSION_SPEC.md",
        },
    })
}

// Remove registry/symbol mapping loading from init(). Only register formula functions if mapping is already loaded.
func init() {
	// Only register formula functions if SymbolMapping is already loaded (by main.go)
	// (No computation functions are registered here; registration is external.)
}

// GetBridgeFormulaRegistry returns the singleton registry instance
// NOTE: The registry must be initialized by main.go before use.
func GetBridgeFormulaRegistry() *BridgeFormulaRegistry {
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
