// WHO: FallbackUtilities
// WHAT: Fallback routing utilities for TNOS
// WHEN: During failure handling
// WHERE: System Layer 6 (Integration)
// WHY: To provide robust operation with fallbacks
// HOW: Using multiple alternative routes
// EXTENT: All fallback needs

package crypto

import (
	"context"
	"fmt"

	"github-mcp-server/pkg/log"
)

// FallbackFunction is a function type for fallback routes
type FallbackFunction func() (interface{}, error)

// FallbackRoute implements the fallback logic for operations that need to try multiple paths
// WHO: FallbackRouter
// WHAT: Route operations through fallbacks
// WHEN: During operation failures
// WHERE: System Layer 6 (Integration)
// WHY: To ensure operation success
// HOW: Using prioritized fallbacks
// EXTENT: All critical operations
func FallbackRoute(
	ctx context.Context,
	operationName string,
	operationContext interface{},
	primary FallbackFunction,
	fallbacks ...FallbackFunction,
) (interface{}, error) {
	// Try primary function first
	result, err := primary()
	if err == nil {
		return result, nil
	}

	logger, ok := operationContext.(log.Logger)
	var loggerPtr *log.Logger
	if ok {
		loggerPtr = &logger
	}

	// If primary failed, try fallbacks in order
	var lastErr error = err
	for i, fallback := range fallbacks {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if loggerPtr != nil {
				loggerPtr.Warn("Primary route failed for %s, trying fallback %d", operationName, i+1)
			}
			result, err := fallback()
			if err == nil {
				return result, nil
			}
			lastErr = err
		}
	}

	// All fallbacks failed
	return nil, fmt.Errorf("all routes failed for operation %s: %w", operationName, lastErr)
}
