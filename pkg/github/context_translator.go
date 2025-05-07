/*
 * WHO: ContextTranslator
 * WHAT: Translates context between GitHub and TNOS MCP systems
 * WHEN: During cross-system communications
 * WHERE: System Layer 6 (Integration)
 * WHY: To maintain context coherence across systems
 * HOW: Using bidirectional context mapping with compression
 * EXTENT: All context transitions between systems
 */

package github

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/translations"
)

// ContextVector7D represents the 7-dimensional context vector used in GitHub MCP
type ContextVector7D struct {
	// WHO: ContextVectorManager
	// WHAT: 7D context structure
	// WHEN: During context operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To represent context dimensions
	// HOW: Using structured dimensions
	// EXTENT: Context representation

	Who    string                 `json:"who"`
	What   string                 `json:"what"`
	When   int64                  `json:"when"`
	Where  string                 `json:"where"`
	Why    string                 `json:"why"`
	How    string                 `json:"how"`
	Extent float64                `json:"extent"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
}

// ContextTranslator handles context translation between systems
type ContextTranslator struct {
	// WHO: TranslationManager
	// WHAT: Context translation service
	// WHEN: During context exchange
	// WHERE: System Layer 6 (Integration)
	// WHY: To ensure proper context mapping
	// HOW: Using translation algorithms
	// EXTENT: All context translations

	logger *log.Logger

	// Configuration options
	enableCompression bool
	preserveMetadata  bool
	strictMapping     bool

	// Statistics
	translationCount int64
	lastTranslation  time.Time
}

// NewContextTranslator creates a new context translator
func NewContextTranslator(
	logger *log.Logger,
	enableCompression bool,
	preserveMetadata bool,
	strictMapping bool,
) *ContextTranslator {
	// WHO: TranslatorFactory
	// WHAT: Create context translator
	// WHEN: During system initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To initialize translation service
	// HOW: Using factory pattern
	// EXTENT: Translator lifecycle

	if logger == nil {
		// Create a minimal logger if none provided
		logger = log.NewLogger(log.Config{
			Level:      log.LevelInfo,
			ConsoleOut: true,
		})
	}

	return &ContextTranslator{
		logger:            logger,
		enableCompression: enableCompression,
		preserveMetadata:  preserveMetadata,
		strictMapping:     strictMapping,
		translationCount:  0,
	}
}

// TranslateGitHubToTNOS translates GitHub context to TNOS context
func (ct *ContextTranslator) TranslateGitHubToTNOS(
	githubContext *ContextVector7D,
) (*translations.ContextVector7D, error) {
	// WHO: ForwardTranslator
	// WHAT: Convert GitHub to TNOS context
	// WHEN: During outbound communications
	// WHERE: System Layer 6 (Integration)
	// WHY: To adapt context for TNOS
	// HOW: Using forward mapping
	// EXTENT: GitHub→TNOS translations

	if githubContext == nil {
		return nil, fmt.Errorf("cannot translate nil GitHub context")
	}

	// Create TNOS context from GitHub context
	tnosContext := &translations.ContextVector7D{
		Who:    githubContext.Who,
		What:   githubContext.What,
		When:   githubContext.When,
		Where:  githubContext.Where,
		Why:    githubContext.Why,
		How:    githubContext.How,
		Extent: githubContext.Extent,
		Source: "github_mcp",
	}

	// Handle metadata
	if ct.preserveMetadata && githubContext.Meta != nil {
		tnosContext.Meta = githubContext.Meta
	} else if githubContext.Meta != nil {
		// Extract essential metadata only
		tnosContext.Meta = make(map[string]interface{})

		// Extract Möbius factors
		factorKeys := []string{"B", "V", "I", "G", "F"}
		for _, key := range factorKeys {
			if val, ok := githubContext.Meta[key]; ok {
				tnosContext.Meta[key] = val
			}
		}
	}

	// Apply additional TNOS-specific enrichment
	now := time.Now().Unix()
	tnosContext.Meta["translated_at"] = now
	tnosContext.Meta["translation_direction"] = "github_to_tnos"
	tnosContext.Meta["translator_version"] = "1.0"

	// Update statistics
	ct.translationCount++
	ct.lastTranslation = time.Now()

	ct.logger.Debug("Translated GitHub context to TNOS",
		"who", githubContext.Who,
		"what", githubContext.What)

	return tnosContext, nil
}

// TranslateTNOSToGitHub translates TNOS context to GitHub context
func (ct *ContextTranslator) TranslateTNOSToGitHub(
	tnosContext *translations.ContextVector7D,
) (*ContextVector7D, error) {
	// WHO: ReverseTranslator
	// WHAT: Convert TNOS to GitHub context
	// WHEN: During inbound communications
	// WHERE: System Layer 6 (Integration)
	// WHY: To adapt context for GitHub
	// HOW: Using reverse mapping
	// EXTENT: TNOS→GitHub translations

	if tnosContext == nil {
		return nil, fmt.Errorf("cannot translate nil TNOS context")
	}

	// Create GitHub context from TNOS context
	githubContext := &ContextVector7D{
		Who:    tnosContext.Who,
		What:   tnosContext.What,
		When:   tnosContext.When,
		Where:  tnosContext.Where,
		Why:    tnosContext.Why,
		How:    tnosContext.How,
		Extent: tnosContext.Extent,
	}

	// Handle metadata
	if ct.preserveMetadata && tnosContext.Meta != nil {
		githubContext.Meta = tnosContext.Meta
	} else if tnosContext.Meta != nil {
		// Extract essential metadata only
		githubContext.Meta = make(map[string]interface{})

		// Extract Möbius factors
		factorKeys := []string{"B", "V", "I", "G", "F"}
		for _, key := range factorKeys {
			if val, ok := tnosContext.Meta[key]; ok {
				githubContext.Meta[key] = val
			}
		}
	}

	// Apply additional GitHub-specific enrichment
	now := time.Now().Unix()
	if githubContext.Meta == nil {
		githubContext.Meta = make(map[string]interface{})
	}
	githubContext.Meta["translated_at"] = now
	githubContext.Meta["translation_direction"] = "tnos_to_github"
	githubContext.Meta["translator_version"] = "1.0"

	// Update statistics
	ct.translationCount++
	ct.lastTranslation = time.Now()

	ct.logger.Debug("Translated TNOS context to GitHub",
		"who", tnosContext.Who,
		"what", tnosContext.What)

	return githubContext, nil
}

// TranslateMapToTNOS converts a map context to TNOS context
func (ct *ContextTranslator) TranslateMapToTNOS(
	contextMap map[string]interface{},
) (*translations.ContextVector7D, error) {
	// WHO: MapTranslator
	// WHAT: Convert map to TNOS context
	// WHEN: During map conversions
	// WHERE: System Layer 6 (Integration)
	// WHY: To adapt map data for TNOS
	// HOW: Using map extraction
	// EXTENT: Map→TNOS translations

	if contextMap == nil {
		return nil, fmt.Errorf("cannot translate nil context map")
	}

	// Extract fields from the map
	githubContext := &ContextVector7D{
		Who:    getMapString(contextMap, "who", "Unknown"),
		What:   getMapString(contextMap, "what", "Unknown"),
		When:   getMapInt64(contextMap, "when", time.Now().Unix()),
		Where:  getMapString(contextMap, "where", "Unknown"),
		Why:    getMapString(contextMap, "why", "Unknown"),
		How:    getMapString(contextMap, "how", "Unknown"),
		Extent: getMapFloat(contextMap, "extent", 1.0),
	}

	// Handle meta field
	if meta, ok := contextMap["meta"].(map[string]interface{}); ok {
		githubContext.Meta = meta
	}

	// Translate to TNOS format
	return ct.TranslateGitHubToTNOS(githubContext)
}

// TranslateTNOSToMap converts a TNOS context to a map
func (ct *ContextTranslator) TranslateTNOSToMap(
	tnosContext *translations.ContextVector7D,
) (map[string]interface{}, error) {
	// WHO: MapConverter
	// WHAT: Convert TNOS to map context
	// WHEN: During context serialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To adapt TNOS data for transmission
	// HOW: Using map creation
	// EXTENT: TNOS→Map translations

	if tnosContext == nil {
		return nil, fmt.Errorf("cannot translate nil TNOS context")
	}

	// Convert to GitHub context first
	githubContext, err := ct.TranslateTNOSToGitHub(tnosContext)
	if err != nil {
		return nil, err
	}

	// Convert to map
	contextMap := map[string]interface{}{
		"who":    githubContext.Who,
		"what":   githubContext.What,
		"when":   githubContext.When,
		"where":  githubContext.Where,
		"why":    githubContext.Why,
		"how":    githubContext.How,
		"extent": githubContext.Extent,
	}

	// Add metadata if present
	if githubContext.Meta != nil {
		contextMap["meta"] = githubContext.Meta
	}

	// Apply compression if enabled
	if ct.enableCompression {
		compressedMap, err := ct.CompressContext(contextMap)
		if err != nil {
			ct.logger.Warn("Failed to compress context", "error", err.Error())
			// Continue with uncompressed map
		} else {
			contextMap = compressedMap
		}
	}

	return contextMap, nil
}

// CompressContext applies Möbius compression to a context map
func (ct *ContextTranslator) CompressContext(
	context map[string]interface{},
) (map[string]interface{}, error) {
	// WHO: ContextCompressor
	// WHAT: Apply context compression
	// WHEN: During data optimization
	// WHERE: System Layer 6 (Integration)
	// WHY: To reduce context size
	// HOW: Using Möbius compression
	// EXTENT: Context optimization

	// Extract context as bytes for entropy calculation
	contextBytes, err := json.Marshal(context)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal context: %w", err)
	}

	// Extract Möbius factors or use defaults
	meta, _ := context["meta"].(map[string]interface{})

	B := 0.8 // Base factor
	V := 0.7 // Value factor
	I := 0.9 // Intent factor
	G := 1.2 // Growth factor
	F := 0.6 // Flexibility factor

	// Override with meta values if available
	if meta != nil {
		if val, ok := meta["B"].(float64); ok {
			B = val
		}
		if val, ok := meta["V"].(float64); ok {
			V = val
		}
		if val, ok := meta["I"].(float64); ok {
			I = val
		}
		if val, ok := meta["G"].(float64); ok {
			G = val
		}
		if val, ok := meta["F"].(float64); ok {
			F = val
		}
	}

	// Calculate entropy (simplified)
	entropy := float64(len(contextBytes)) / 100.0

	// Calculate time factor
	now := time.Now().Unix()
	when := getMapInt64(context, "when", now)
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

	// Create compressed context with compression metadata
	compressed := map[string]interface{}{
		"compressed": true,
		"context":    context,
		"compression": map[string]interface{}{
			"algorithm":         "mobius",
			"version":           "1.0",
			"originalSize":      len(contextBytes),
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

// GetTranslationStats returns statistics about the translator
func (ct *ContextTranslator) GetTranslationStats() map[string]interface{} {
	// WHO: StatsProvider
	// WHAT: Get translation statistics
	// WHEN: During monitoring
	// WHERE: System Layer 6 (Integration)
	// WHY: To track translation activity
	// HOW: Using statistics collection
	// EXTENT: Monitoring data

	stats := map[string]interface{}{
		"translation_count": ct.translationCount,
		"last_translation":  ct.lastTranslation.Unix(),
		"configuration": map[string]interface{}{
			"enable_compression": ct.enableCompression,
			"preserve_metadata":  ct.preserveMetadata,
			"strict_mapping":     ct.strictMapping,
		},
	}

	return stats
}

// Helper function to safely extract a string from a map
func getMapString(m map[string]interface{}, key string, defaultValue string) string {
	// WHO: MapAccessor
	// WHAT: Extract string from map
	// WHEN: During map operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To safely access map values
	// HOW: Using type assertion
	// EXTENT: Map value extraction

	if val, ok := m[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

// Helper function to safely extract an int64 from a map
func getMapInt64(m map[string]interface{}, key string, defaultValue int64) int64 {
	// WHO: MapAccessor
	// WHAT: Extract int64 from map
	// WHEN: During map operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To safely access map values
	// HOW: Using type assertion
	// EXTENT: Map value extraction

	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return int64(v)
		case float32:
			return int64(v)
		case int:
			return int64(v)
		case int32:
			return int64(v)
		case int64:
			return v
		}
	}
	return defaultValue
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
