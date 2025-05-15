//go:build testing

/*
 * WHO: TestHelperManager
 * WHAT: Test setup helper for GitHub MCP testing
 * WHEN: During test initialization
 * WHERE: System Layer 6 (Integration Testing)
 * WHY: To ensure proper package initialization
 * HOW: Using Go build tags
 * EXTENT: All GitHub MCP test files
 */

package github

import (
	// Ensure all required test dependencies are referenced
	_ "github.com/tranquility-neuro-os/github-mcp-server/pkg/github/testutil"
	_ "github.com/tranquility-neuro-os/github-mcp-server/pkg/translations"
)

// TestHelper provides a placeholder function to ensure the testing package is properly initialized
// WHO: TestHelperManager
// WHAT: Package initialization verification
// WHEN: During test setup
// WHERE: System Layer 6 (Integration Testing)
// WHY: To force proper package resolution
// HOW: Using explicit import references
// EXTENT: All GitHub MCP test dependencies
func TestHelper() bool {
	return true
}
