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
	"encoding/json"
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

	// For testing purposes, we'll convert the first content item to JSON
	contentJSON, err := json.Marshal(result.Content[0])
	require.NoError(t, err)
	return string(contentJSON)
}
