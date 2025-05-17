/*
 * WHO: TranslationManager
 * WHAT: Translation utilities for GitHub MCP server
 * WHEN: During context translation between systems
 * WHERE: System Layer 6 (Integration)
 * WHY: To facilitate interoperability between GitHub and TNOS
 * HOW: Using translation helpers with context awareness
 * EXTENT: All translation operations
 */

package translations

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Import and use ContextVector7D from context.go
// WHO: ContextContainer
// WHAT: Store 7D context data
// WHEN: During context operations
// WHERE: System Layer 6 (Integration)
// WHY: To store contextual information
// HOW: Using structured data
// EXTENT: All context operations

// The ToMap method is already defined in context.go
// Use that implementation instead of duplicating it here

// NewContextVector7D is already defined in context.go
// Use that implementation instead of duplicating it here.
// This comment references the function for documentation purposes:
// WHO: ContextFactory
// WHAT: Create context vector
// WHEN: During context initialization
// WHERE: System Layer 6 (Integration)
// WHY: For standardized creation
// HOW: Using factory pattern
// EXTENT: Context creation operations

// Compress is already defined in context.go
// WHO: CompressionEngine
// WHAT: Compress context vector
// WHEN: During optimization
// WHERE: System Layer 6 (Integration)
// WHY: For efficient transmission
// HOW: Using Möbius Compression Formula
// EXTENT: Compression operations
// func (cv *ContextVector7D) Compress() *ContextVector7D {
// This function is now imported from context.go

// Decompress reverses the Möbius Compression Formula
// WHO: DecompressionEngine
// WHAT: Decompress context vector
// WHEN: During restoration
// WHERE: System Layer 6 (Integration)
// WHY: For full information access
// HOW: Using inverse Möbius Compression Formula
// EXTENT: Decompression operations
// This function is now imported from context.go
// func (cv *ContextVector7D) Decompress() *ContextVector7D {

// calculateContextEntropy is already defined in context.go
// WHO: EntropyCalculator
// WHAT: Calculate context entropy
// WHEN: During compression optimization
// WHERE: System Layer 6 (Integration)
// WHY: To inform compression algorithm
// HOW: Using information theory formulas
// EXTENT: Entropy calculation operations
// This function is imported from context.go

// String returns a string representation of the context vector
// calculateContextEntropy exists in context.go if needed
// WHO: EntropyCalculator
// WHAT: Calculate context entropy
// WHEN: During compression optimization
// WHERE: System Layer 6 (Integration)
// WHY: To inform compression algorithm
// HOW: Using information theory formulas
// EXTENT: Entropy calculation operations
// This function should be imported from context.go

// ToMap converts a ContextVector7D to a map
// WHO: MapExporter
// WHAT: Convert context vector to map
// WHEN: During context serialization
// WHERE: System Layer 6 (Integration)
// WHY: To transform to external format
// HOW: Using structured mapping
// EXTENT: Map-based context conversion
// This method is already defined in context.go and should be imported rather than duplicated here

// FromMap converts a map to ContextVector7D
// WHO: MapImporter
// WHAT: Convert map to context vector
// WHEN: During context deserialization
// WHERE: System Layer 6 (Integration)
// WHY: To transform external format
// HOW: Using structured conversion
// EXTENT: Map-based context conversion
// This function is already defined in context.go and should be imported rather than duplicated here

// ToJSON converts the context vector to JSON
func (cv ContextVector7D) ToJSON() (string, error) {
	// WHO: JSONFormatter
	// WHAT: Convert context to JSON
	// WHEN: During context serialization
	// WHERE: System Layer 6 (Integration)
	// WHY: For interchange format
	// HOW: Using JSON marshaling
	// EXTENT: Single context serialization

	data, err := json.Marshal(cv)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON converts JSON to context vector
func FromJSON(jsonStr string) (ContextVector7D, error) {
	// WHO: JSONParser
	// WHAT: Parse JSON to context
	// WHEN: During context deserialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To process external data
	// HOW: Using JSON unmarshaling
	// EXTENT: Single context deserialization

	var cv ContextVector7D
	err := json.Unmarshal([]byte(jsonStr), &cv)
	if err != nil {
		return NewContextVector7D(map[string]interface{}{}), err
	}
	return cv, nil
}

// MCPContextToTNOS converts MCP context to TNOS 7D context
// MCPContextToTNOS converts MCP context to TNOS 7D context
// This is a forwarding function that uses the implementation in context.go
func MCPContextToTNOS(mcpCtx map[string]interface{}) ContextVector7D {
	// WHO: FormatConverter
	// WHAT: Convert MCP to TNOS format
	// WHEN: During inbound operations
	// WHERE: System Layer 6 (Integration)
	// WHY: For internal format compatibility
	// HOW: Using dimension mapping with compression-first approach
	// EXTENT: Single context conversion

	// Implement the conversion logic directly instead of calling MCPToTNOSContext
	
	// Extract MCP context fields with default values
	who := getStringValue(mcpCtx, "identity", "System")
	what := getStringValue(mcpCtx, "operation", "Transform")
	when := mcpCtx["timestamp"]
	where := getStringValue(mcpCtx, "location", "MCP_Bridge")
	why := getStringValue(mcpCtx, "purpose", "Protocol_Compliance")
	how := getStringValue(mcpCtx, "method", "Context_Translation")
	extent := mcpCtx["scope"]
	
	// Create the context vector
	cv := ContextVector7D{
		Who:    who,
		What:   what,
		When:   when,
		Where:  where,
		Why:    why,
		How:    how,
		Extent: extent,
		Source: "github-mcp",
		Meta:   map[string]interface{}{
			"translated_at": time.Now().Unix(),
			"translation_type": "mcp_to_tnos",
		},
	}
	
	// Add MCP original context for reference
	cv.Meta["original_mcp_context"] = mcpCtx
	
	return cv
}

// Helper function to safely extract string values from a map
func getStringValue(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

// TNOSContextToMCP converts TNOS 7D context to MCP context
func TNOSContextToMCP(cv ContextVector7D) map[string]interface{} {
	// WHO: FormatConverter
	// WHAT: Convert TNOS to MCP format
	// WHEN: During outbound operations
	// WHERE: System Layer 6 (Integration)
	// WHY: For external format compatibility
	// HOW: Using dimension mapping with compression-first approach
	// EXTENT: Single context conversion

	// Apply decompression if needed (check if context is compressed)
	if compressed, ok := cv.Meta["compressed"].(bool); ok && compressed {
		// Decompress to ensure we have full context information
		decompressed := cv.Decompress()
		cv = *decompressed
	}

	// Create result with full 7D context awareness
	result := map[string]interface{}{
		"identity":  cv.Who,
		"operation": cv.What,
		"timestamp": cv.When,
		"location":  cv.Where,
		"purpose":   cv.Why,
		"method":    cv.How,
		"scope":     cv.Extent,
		"source":    cv.Source,

		// Include the full TNOS context in a special field for complete context preservation
		"_tnos_context": map[string]interface{}{
			"who":    cv.Who,
			"what":   cv.What,
			"when":   cv.When,
			"where":  cv.Where,
			"why":    cv.Why,
			"how":    cv.How,
			"extent": cv.Extent,
		},
	}

	// Add metadata if available with preservation of compression factors
	if len(cv.Meta) > 0 {
		metadata := make(map[string]interface{})

		// Transfer all metadata
		for k, v := range cv.Meta {
			metadata[k] = v
		}

		// Add translation timestamp
		metadata["translated_at"] = time.Now().Unix()
		metadata["translation_type"] = "tnos_to_mcp"

		result["metadata"] = metadata
	}

	return result
}

// ContextWithVector adds a 7D context vector to a Go context
// This is a forwarding function that uses the implementation in context.go
func ContextWithVector(ctx context.Context, cv ContextVector7D) context.Context {
	// WHO: ContextEnricher
	// WHAT: Store context in Go context
	// WHEN: During operation context enrichment
	// WHERE: System Layer 6 (Integration)
	// WHY: For context propagation
	// HOW: Using context values
	// EXTENT: Go context enrichment
	
	// This function mirrors the implementation in context.go
	// We use the same key for consistency across the system
	return context.WithValue(ctx, contextKey("7d_context"), cv)
}

// VectorFromContext extracts a 7D context vector from a Go context
func VectorFromContext(ctx context.Context) (ContextVector7D, bool) {
	// WHO: ContextExtractor
	// WHAT: Extract context from Go context
	// WHEN: During operation context retrieval
	// WHERE: System Layer 6 (Integration)
	// WHY: For context utilization
	// HOW: Using context lookup
	// EXTENT: Go context extraction

	cv, ok := ctx.Value(contextKey("7d_context")).(ContextVector7D)
	return cv, ok
}

// contextKey is a private type for context keys
type contextKey string

// CompressTranslationContext applies Möbius Compression to context vector during translation
func CompressTranslationContext(cv ContextVector7D) ContextVector7D {
	// WHO: TranslationCompressor
	// WHAT: Apply compression to context
	// WHEN: During context optimization
	// WHERE: System Layer 6 (Integration)
	// WHY: For efficient transmission
	// HOW: Using compression-first approach
	// EXTENT: Context transmission optimization

	// Forward to the Compress method from context.go
	// This ensures we use a single implementation and avoid duplication
	compressed := cv.Compress()
	return *compressed
}

type TranslationHelperFunc func(key string, defaultValue string) string

// CreateTranslationHelper creates a translation helper function
func CreateTranslationHelper() (TranslationHelperFunc, func()) {
	// WHO: TranslationHelperFactory
	// WHAT: Create translation helper
	// WHEN: During initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: For translation assistance
	// HOW: Using closure pattern
	// EXTENT: Translation helper creation

	var translationKeyMap = map[string]string{}
	v := viper.New()

	v.SetEnvPrefix("GITHUB_MCP_")
	v.AutomaticEnv()

	// Load from JSON file
	v.SetConfigName("github-mcp-server-config")
	v.SetConfigType("json")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		// ignore error if file not found as it is not required
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Could not read JSON config: %v", err)
		}
	}

	// create a function that takes both a key, and a default value and returns either the default value or an override value
	return func(key string, defaultValue string) string {
			key = strings.ToUpper(key)
			if value, exists := translationKeyMap[key]; exists {
				return value
			}
			// check if the env var exists
			if value, exists := os.LookupEnv("GITHUB_MCP_" + key); exists {
				// TODO I could not get Viper to play ball reading the env var
				translationKeyMap[key] = value
				return value
			}

			v.SetDefault(key, defaultValue)
			translationKeyMap[key] = v.GetString(key)
			return translationKeyMap[key]
		}, func() {
			// dump the translationKeyMap to a json file
			if err := DumpTranslationKeyMap(translationKeyMap); err != nil {
				log.Fatalf("Could not dump translation key map: %v", err)
			}
		}
}

// dump translationKeyMap to a json file called github-mcp-server-config.json
func DumpTranslationKeyMap(translationKeyMap map[string]string) error {
	// WHO: ConfigSerializer
	// WHAT: Save translation key map
	// WHEN: During configuration persistence
	// WHERE: System Layer 6 (Integration)
	// WHY: To preserve translations
	// HOW: Using JSON serialization
	// EXTENT: Configuration persistence

	file, err := os.Create("github-mcp-server-config.json")
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer func() { _ = file.Close() }()

	// marshal the map to json
	jsonData, err := json.MarshalIndent(translationKeyMap, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling map to JSON: %v", err)
	}

	// write the json data to the file
	if _, err := file.Write(jsonData); err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}
