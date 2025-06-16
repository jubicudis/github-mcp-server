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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Note: The following context-related functions are defined in context.go:
// - MCPContextToTNOS: Converts MCP context to TNOS 7D context
// - TNOSContextToMCP: Converts TNOS 7D context to MCP context
// - ContextWithVector: Adds a 7D context vector to a Go context
// - VectorFromContext: Extracts a 7D context vector from a Go context
// - contextKey type: Used for context keys

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

// Note: FromJSON is now in helper_interfaces.go to avoid duplicate declarations

// Helper function to safely extract string values from a map
func getStringValue(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

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
	standalone := true // TODO: wire to actual server state
	compressed := cv.Compress(standalone)
	return *compressed
}

// CompressTranslationContextWithMode applies Möbius Compression to context vector with explicit mode
func CompressTranslationContextWithMode(cv ContextVector7D, standalone bool) ContextVector7D {
	compressed := cv.Compress(standalone)
	return *compressed
}

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
