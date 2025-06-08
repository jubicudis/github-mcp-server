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
	"github-mcp-server/pkg/common"

	"github.com/mark3labs/mcp-go/mcp"
)

// Deprecated: use common.RequiredParam instead
func RequiredParam[T comparable](r mcp.CallToolRequest, p string) (T, error) {
	return common.RequiredParam[T](r, p)
}

// Deprecated: use common.OptionalParamOK instead
func OptionalParam[T any](r mcp.CallToolRequest, p string) (value T, err error) {
	v, _, err := common.OptionalParamOK[T](r, p)
	return v, err
}
