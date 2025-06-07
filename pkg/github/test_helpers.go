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

	"github-mcp-server/pkg/github/testutil"

	"github.com/mark3labs/mcp-go/mcp"
)

// Ptr function is now defined in common.go and testutil package

// CreateTestTranslateFunc returns a simple translation function for tests
// This is a wrapper around testutil.CreateTestTranslateFuncSimple for backward compatibility
func CreateTestTranslateFunc() func(key string, defaultValue string) string {
	return testutil.CreateTestTranslateFuncSimple()
}

// GetTextContent extracts text content from an MCP result
// This is a wrapper around testutil.GetTextContent for backward compatibility
func GetTextContent(t *testing.T, result *mcp.CallToolResult) string {
	return testutil.GetTextContent(t, result)
}
