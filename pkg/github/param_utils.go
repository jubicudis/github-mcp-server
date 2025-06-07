/*
 * WHO: ParameterUtils
 * WHAT: Utility functions for parameter handling in GitHub MCP
 * WHEN: During MCP request handling
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide consistent parameter extraction
 * HOW: Using Go generics for type safety
 * EXTENT: All parameter handling in GitHub MCP
 */

package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v71/github"
	"github.com/mark3labs/mcp-go/mcp"
)

// GetClientFn is a function type for getting a GitHub client
type GetClientFn func(ctx context.Context) (*github.Client, error)

// RequiredParam retrieves a required parameter with generic type
func RequiredParam[T comparable](r mcp.CallToolRequest, p string) (T, error) {
    var zero T
    v, ok := r.Params.Arguments[p]
    if !ok {
        return zero, fmt.Errorf("missing required parameter: %s", p)
    }
    value, ok2 := v.(T)
    if !ok2 || value == zero {
        return zero, fmt.Errorf("invalid parameter: %s", p)
    }
    return value, nil
}

// OptionalParam retrieves an optional parameter
func OptionalParam[T any](r mcp.CallToolRequest, p string) (value T, _ error) {
    val, exists := r.Params.Arguments[p]
    if !exists {
        return value, nil // Return zero value for missing params
    }
    
    result, ok := val.(T)
    if !ok {
        return value, fmt.Errorf("parameter %s is not of type %T", p, value)
    }
    
    return result, nil
}
