/*
 * WHO: ContextBridge
 * WHAT: Bridge between different context implementations
 * WHEN: During context operations across systems
 * WHERE: System Layer 6 (Integration)
 * WHY: To ensure compatibility between context implementations
 * HOW: Using translation functions
 * EXTENT: All context bridging operations
 */

package translations

import (
	"context"

	logpkg "github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log"
)

// TranslationsContextKey is the key used in translations.go
type stringContextKey string

const TranslationsContextKey stringContextKey = "7d_context"

// BridgeContextsAndSync synchronizes context implementations
// WHO: ContextSynchronizer
// WHAT: Synchronize context implementations
// WHEN: During system operation
// WHERE: System Layer 6 (Integration)
// WHY: To maintain context consistency
// HOW: Using bidirectional synchronization
// EXTENT: All cross-system context operations
func BridgeContextsAndSync(ctx context.Context) context.Context {
	// Check if context exists in either implementation
	var cv ContextVector7D
	var exists bool

	// Try getting from context.go implementation (preferred)
	cv, exists = VectorFromContext(ctx)

	if !exists {
		// Try getting from translations.go implementation
		strVal, ok := ctx.Value(TranslationsContextKey).(ContextVector7D)
		if ok {
			// Found in translations.go, use it
			cv = strVal
			exists = true
		}
	}

	if !exists {
		// No context found, create a default one
		cv = NewContextVector7D(map[string]interface{}{
			"who":    "System",
			"what":   "ContextBridge",
			"where":  "Integration",
			"why":    "Compatibility",
			"how":    "BridgeOperation",
			"extent": 1.0,
			"source": "context_bridge",
		})
	}

	// Apply compression-first approach
	// NOTE: Must specify standalone/blood-connected mode for Mobius compression
	standalone := true // TODO: wire to actual server state
	compressed := cv.Compress(standalone)

	// Store in both implementations for compatibility
	newCtx := context.WithValue(ctx, TranslationsContextKey, *compressed)
	return context.WithValue(newCtx, contextVector7DKey, *compressed)
}

// SaveToTranslationsContext saves a context vector to the translations.go format
// WHO: ContextConverter
// WHAT: Save context in translations format
// WHEN: During cross-implementation operations
// WHERE: System Layer 6 (Integration)
// WHY: For backward compatibility
// HOW: Using string key storage
// EXTENT: All compatibility operations
func SaveToTranslationsContext(ctx context.Context, cv ContextVector7D) context.Context {
	return context.WithValue(ctx, TranslationsContextKey, cv)
}

// GetFromTranslationsContext gets a context vector from the translations.go format
// WHO: ContextRetriever
// WHAT: Get context from translations format
// WHEN: During cross-implementation operations
// WHERE: System Layer 6 (Integration)
// WHY: For backward compatibility
// HOW: Using string key lookup
// EXTENT: All compatibility operations
func GetFromTranslationsContext(ctx context.Context) (ContextVector7D, bool) {
	cv, ok := ctx.Value(TranslationsContextKey).(ContextVector7D)
	return cv, ok
}

// MigrateAllContextReferences updates translations to use standardized context implementation
func MigrateAllContextReferences(ctx context.Context, logger logpkg.LoggerInterface) context.Context {
	// This function should be called at the beginning of operations
	// to ensure all code uses the standardized context implementation

	// Check if there's a context in translations.go format
	translationsCV, translationsExists := GetFromTranslationsContext(ctx)

	// Check if there's a context in context.go format
	contextCV, contextExists := VectorFromContext(ctx)

	if translationsExists && !contextExists {
		// Only exists in translations.go format, migrate to context.go
		if logger != nil {
			logger.Info("Migrating context from translations.go to context.go format")
		}
		return context.WithValue(ctx, contextVector7DKey, translationsCV)
	}

	if !translationsExists && contextExists {
		// Only exists in context.go format, add to translations.go for compatibility
		if logger != nil {
			logger.Info("Adding context.go format to translations.go for compatibility")
		}
		return context.WithValue(ctx, TranslationsContextKey, contextCV)
	}

	if translationsExists && contextExists {
		// Both exist, check if they're the same and sync if not
		if translationsCV.When != contextCV.When ||
			translationsCV.Who != contextCV.Who ||
			translationsCV.What != contextCV.What {

			// They differ, use the newer one
			var newerCV ContextVector7D
			if translationsCV.When > contextCV.When {
				if logger != nil {
					logger.Info("Using translations.go context (newer) for synchronization")
				}
				newerCV = translationsCV
			} else {
				if logger != nil {
					logger.Info("Using context.go context (newer) for synchronization")
				}
				newerCV = contextCV
			}

			// Sync both implementations
			newCtx := context.WithValue(ctx, TranslationsContextKey, newerCV)
			return context.WithValue(newCtx, contextVector7DKey, newerCV)
		}

		// They're the same, no action needed
		return ctx
	}

	// No context exists in either format, create a default one
	defaultCV := NewContextVector7D(map[string]interface{}{
		"who":    "System",
		"what":   "ContextMigration",
		"where":  "Integration",
		"why":    "Standardization",
		"how":    "MigrationOperation",
		"extent": 1.0,
		"source": "context_bridge",
	})

	if logger != nil {
		logger.Info("Creating default context for both implementations")
	}

	// Store in both implementations
	newCtx := context.WithValue(ctx, TranslationsContextKey, defaultCV)
	return context.WithValue(newCtx, contextVector7DKey, defaultCV)
}

// BridgeMCPContext bridges GitHub and TNOS 7D context vectors
func BridgeMCPContext(githubCtx GitHubContext, tnosCtx *ContextVector7D, logger logpkg.LoggerInterface) ContextVector7D {
	// Create a new 7D vector from GitHub context fields
	params := map[string]interface{}{ // 7D context mapping
		"who":    githubCtx.User,
		"what":   githubCtx.Operation,
		"when":   githubCtx.Timestamp,
		"where":  tnosCtx.Where,
		"why":    githubCtx.Purpose,
		"how":    githubCtx.Type,
		"extent": githubCtx.Scope,
		"source": githubCtx.Source,
	}
	newCV := NewContextVector7D(params)
	// Merge with existing TNOS context, prioritizing the more recent
	merged := tnosCtx.Merge(newCV)
	if logger != nil {
		logger.Info("Bridged 7D context: %+v", merged)
	}
	return merged
}
