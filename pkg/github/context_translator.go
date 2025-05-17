/*
 * WHO: ContextTranslator
 * WHAT: Translate context between GitHub MCP and TNOS 7D formats
 * WHEN: During cross-system context operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide bidirectional context translation
 * HOW: Using translations package with compression-first approach
 * EXTENT: All context translations between GitHub and TNOS
 */

package github

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jubicudis/github-mcp-server/pkg/log"
	"github.com/jubicudis/github-mcp-server/pkg/translations"
)

// ContextTranslator implements bidirectional context translation
// WHO: TranslatorManager
// WHAT: Manage context translations
// WHEN: During cross-system communication
// WHERE: System Layer 6 (Integration)
// WHY: To ensure context compatibility
// HOW: Using translations package
// EXTENT: All context conversions
type ContextTranslator struct {
	// Logger instance
	logger *log.Logger

	// Translation options
	enableCompression bool
	preserveMetadata  bool
	strictMapping     bool

	// Translation statistics
	successCount int64
	failureCount int64
}

// NewContextTranslator creates a new context translator
// WHO: TranslatorFactory
// WHAT: Create context translator
// WHEN: During translator initialization
// WHERE: System Layer 6 (Integration)
// WHY: To establish translation capability
// HOW: Using factory pattern
// EXTENT: Translator lifecycle
func NewContextTranslator(
	logger *log.Logger,
	enableCompression bool,
	preserveMetadata bool,
	strictMapping bool,
) *ContextTranslator {
	// Create default logger if none provided
	if logger == nil {
		logger = log.NewLogger()
	}

	return &ContextTranslator{
		logger:            logger,
		enableCompression: enableCompression,
		preserveMetadata:  preserveMetadata,
		strictMapping:     strictMapping,
	}
}

// TranslateMapToTNOS translates a GitHub context map to TNOS 7D context
// WHO: GitHubToTNOSTranslator
// WHAT: Convert GitHub context to TNOS
// WHEN: During inbound translation
// WHERE: System Layer 6 (Integration)
// WHY: To convert external format
// HOW: Using translations package
// EXTENT: Inbound context translation
func (t *ContextTranslator) TranslateMapToTNOS(
	githubContext map[string]interface{},
) (map[string]interface{}, error) {
	// Log translation start
	t.logger.Debug("Translating GitHub context to TNOS 7D",
		"source", githubContext["identity"],
		"operation", githubContext["operation"])

	// Convert to TNOS 7D context
	tnosContext := translations.MCPContextToTNOS(githubContext)

	// Apply compression if enabled
	if t.enableCompression {
		tnosContext = translations.CompressTranslationContext(tnosContext)
	}

	// Convert to map for return
	result := tnosContext.ToMap()

	// Add metadata about the translation process
	if result["meta"] == nil {
		result["meta"] = make(map[string]interface{})
	}

	metaMap, _ := result["meta"].(map[string]interface{})
	metaMap["translated_at"] = time.Now().Unix()
	metaMap["translation_source"] = "github_mcp"
	metaMap["translation_target"] = "tnos_7d"
	metaMap["compressed"] = t.enableCompression

	// Update success count
	atomic.AddInt64(&t.successCount, 1)

	return result, nil
}

// TranslateTNOSToMap translates a TNOS 7D context to GitHub context
// WHO: TNOSToGitHubTranslator
// WHAT: Convert TNOS context to GitHub
// WHEN: During outbound translation
// WHERE: System Layer 6 (Integration)
// WHY: To convert internal format
// HOW: Using translations package
// EXTENT: Outbound context translation
func (t *ContextTranslator) TranslateTNOSToMap(
	tnosContext map[string]interface{},
) (map[string]interface{}, error) {
	// Log translation start
	t.logger.Debug("Translating TNOS 7D context to GitHub",
		"who", tnosContext["who"],
		"what", tnosContext["what"])

	// Convert map to ContextVector7D
	cv := translations.FromMap(tnosContext)

	// Convert TNOS 7D context to MCP context
	result := translations.TNOSContextToMCP(cv)

	// Update success count
	atomic.AddInt64(&t.successCount, 1)

	return result, nil
}

// TranslateStringToTNOS translates a GitHub context as JSON string to TNOS 7D context
// WHO: StringTranslator
// WHAT: Convert JSON string to TNOS
// WHEN: During string-based translation
// WHERE: System Layer 6 (Integration)
// WHY: To handle string formats
// HOW: Using JSON parsing and translation
// EXTENT: String-based translation
func (t *ContextTranslator) TranslateStringToTNOS(jsonStr string) (map[string]interface{}, error) {
	// Parse JSON string to context vector
	cv, err := translations.FromJSON(jsonStr)
	if err != nil {
		// Update failure count
		atomic.AddInt64(&t.failureCount, 1)
		return nil, fmt.Errorf("failed to parse JSON context: %w", err)
	}

	// Apply compression if enabled
	if t.enableCompression {
		compressed := cv.Compress()
		cv = *compressed
	}

	// Convert to map
	result := cv.ToMap()

	// Update success count
	atomic.AddInt64(&t.successCount, 1)

	return result, nil
}

// GetTranslationStats returns translation statistics
// WHO: StatsProvider
// WHAT: Provide translation metrics
// WHEN: During monitoring
// WHERE: System Layer 6 (Integration)
// WHY: To monitor translation performance
// HOW: Using atomic counters
// EXTENT: Translation monitoring
func (t *ContextTranslator) GetTranslationStats() map[string]interface{} {
	return map[string]interface{}{
		"Success": atomic.LoadInt64(&t.successCount),
		"Failure": atomic.LoadInt64(&t.failureCount),
		"Total":   atomic.LoadInt64(&t.successCount) + atomic.LoadInt64(&t.failureCount),
	}
}
