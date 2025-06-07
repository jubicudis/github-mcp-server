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

// FallbackRoute executes the primary operation and falls back to alternatives if it fails
// The context argument is passed as-is to operations
func FallbackRoute(
	ctx context.Context,
	operationName string,
	contextData map[string]interface{},
	primaryOp RouteOperation,
	fallback1 RouteOperation,
	fallback2 RouteOperation,
	fallback3 RouteOperation,
	logger *log.Logger,
) (interface{}, error) {
	startTime := time.Now()

	// Try the primary operation
	result, err := primaryOp()
	if err == nil {
		if logger != nil {
			logger.Debug("[%s] Primary operation succeeded in %s", 
				operationName, time.Since(startTime))
		}
		return result, nil
	}

	// Log the primary failure
	if logger != nil {
		logger.Warn("[%s] Primary operation failed: %s", operationName, err.Error())
	}

	// Try fallback 1
	result, err = fallback1()
	if err == nil {
		if logger != nil {
			logger.Info("[%s] Fallback 1 succeeded in %s", 
				operationName, time.Since(startTime))
		}
		return result, nil
	}

	// Log fallback 1 failure
	if logger != nil {
		logger.Warn("[%s] Fallback 1 failed: %s", operationName, err.Error())
	}

	// Try fallback 2
	result, err = fallback2()
	if err == nil {
		if logger != nil {
			logger.Info("[%s] Fallback 2 succeeded in %s", 
				operationName, time.Since(startTime))
		}
		return result, nil
	}

	// Log fallback 2 failure
	if logger != nil {
		logger.Warn("[%s] Fallback 2 failed: %s", operationName, err.Error())
	}

	// Try fallback 3 (final attempt)
	result, err = fallback3()
	if err == nil {
		if logger != nil {
			logger.Info("[%s] Fallback 3 succeeded in %s", 
				operationName, time.Since(startTime))
		}
		return result, nil
	}

	// All attempts failed
	if logger != nil {
		logger.Error("[%s] All fallback routes failed: %s", operationName, err.Error())
	}
	
	return nil, fmt.Errorf("all fallback routes failed for operation %s: %w", operationName, err)
}
