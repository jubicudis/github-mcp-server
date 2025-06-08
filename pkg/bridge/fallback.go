/*
 * WHO: FallbackRouteManager
 * WHAT: Fallback routing functionality
 * WHEN: During operations requiring fallback
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide robust operation with fallbacks
 * HOW: Using multiple execution paths
 * EXTENT: All fallback operations
 */

package bridge

import (
	"context"
	"fmt"
	"time"

	"github-mcp-server/pkg/log"
)

// RouteOperation represents a function that can be executed with fallback
type RouteOperation func() (interface{}, error)

// FallbackFunction is a function that can be executed as a fallback (legacy type)
type FallbackFunction func() (interface{}, error)

// FallbackRoute executes the primary operation and falls back to alternatives if it fails
// The context argument is passed as-is to operations
func FallbackRoute(
	ctx context.Context,
	operationName string,
	contextData map[string]interface{},
	primaryOp FallbackFunction, // Changed to use FallbackFunction for compatibility
	fallback1 FallbackFunction, // Changed to use FallbackFunction for compatibility
	fallback2 FallbackFunction, // Changed to use FallbackFunction for compatibility
	fallback3 FallbackFunction, // Changed to use FallbackFunction for compatibility
	logger *log.Logger,
) (interface{}, error) {
	startTime := time.Now()

	result, err := primaryOp()
	if err == nil {
		if logger != nil {
			logger.Debug("[%s] Primary operation succeeded in %s", 
				operationName, time.Since(startTime))
		}
		return result, nil
	}

	if logger != nil {
		logger.Warn("[%s] Primary operation failed: %v", operationName, err)
	}

	for i, fallback := range []FallbackFunction{fallback1, fallback2, fallback3} {
		result, err = fallback()
		if err == nil {
			if logger != nil {
				logger.Debug("[%s] Fallback #%d succeeded in %s", 
					operationName, i+1, time.Since(startTime))
			}
			return result, nil
		}
		if logger != nil {
			logger.Warn("[%s] Fallback #%d failed: %v", operationName, i+1, err)
		}
	}

	return nil, fmt.Errorf("[%s] All fallback operations failed", operationName)
}
