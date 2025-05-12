/*
 * WHO: GitHubTestHelper
 * WHAT: Helper functions for GitHub tests
 * WHEN: During test execution
 * WHERE: MCP Bridge Layer Testing
 * WHY: To provide common test utilities
 * HOW: By implementing consistent helper patterns
 * EXTENT: All GitHub tests
 */
package github

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// Ptr function is now defined in common.go

// CreateTestTranslateFunc returns a simple translation function for tests
func CreateTestTranslateFunc() func(key string, defaultValue string) string {
	return func(key string, defaultValue string) string {
		return key // Just return the key for testing
	}
}

// GetTextContent extracts text content from an MCP result
func GetTextContent(t *testing.T, result *mcp.CallToolResult) string {
	require.NotNil(t, result)
	require.NotEmpty(t, result.Content, "CallToolResult has no content")

	// For now, return the Content directly as it appears to be a string in our implementation
	// This will need to be updated once the MCP protocol is fully implemented
	return result.Content
}
