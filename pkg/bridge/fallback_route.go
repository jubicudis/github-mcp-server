// WHO: FallbackRouter
// WHAT: Fallback route mechanism for MCP bridge
// WHEN: During connection attempts
// WHERE: System Layer 6 (Integration)
// WHY: To provide robust connection with fallbacks
// HOW: Using fallback chain execution
// EXTENT: All connection attempts with fallback support

package bridge

import (
	"context"

	"github-mcp-server/pkg/log"
	"github-mcp-server/pkg/translations"
)

// FallbackFunction is a function that can be executed as a fallback
type FallbackFunction func() (interface{}, error)

// FallbackRoute executes operations with fallbacks in order
// It attempts the primary function first, and if it fails, tries each fallback in order
func FallbackRoute(
	ctx context.Context,
	operationName string,
	contextInfo translations.ContextVector7D,
	primaryFn FallbackFunction, // Primary operation
	bridgeFallback FallbackFunction, // Bridge fallback
	githubFallback FallbackFunction, // GitHub MCP fallback
	copilotFallback FallbackFunction, // Copilot LLM fallback
	logger *log.Logger,
) (interface{}, error) {
	// Try primary function first
	result, err := primaryFn()
	if err == nil {
		return result, nil
	}
	
	if logger != nil {
		logger.Warn("[FallbackRoute] Primary %s failed: %v", operationName, err)
	}
	
	// Try bridge fallback
	result, err = bridgeFallback()
	if err == nil {
		return result, nil
	}
	
	if logger != nil {
		logger.Warn("[FallbackRoute] Bridge fallback for %s failed: %v", operationName, err)
	}
	
	// Try GitHub MCP fallback
	result, err = githubFallback()
	if err == nil {
		return result, nil
	}
	
	if logger != nil {
		logger.Warn("[FallbackRoute] GitHub MCP fallback for %s failed: %v", operationName, err)
	}
	
	// Try Copilot fallback as last resort
	result, err = copilotFallback()
	if err == nil {
		return result, nil
	}
	
	if logger != nil {
		logger.Error("[FallbackRoute] All fallbacks for %s failed: %v", operationName, err)
	}
	
	return nil, err // Return the last error if all fallbacks fail
}
