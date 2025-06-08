// WHO: GitHubMCPUserTests
// WHAT: User API Testing
// WHEN: During test execution
// WHERE: MCP Bridge Layer Testing
// WHY: To verify user functionality
// HOW: By testing MCP protocol handlers
// EXTENT: All user operations
package github_test

import (
	"github.com/jubicudis/github-mcp-server/pkg/github/testutil"
)

// Helper aliases for legacy test expectations
var expectRequestBody = testutil.MockResponse
var expectQueryParams = testutil.CreateQueryParamExpectation

// Prefix tool calls:
// tool, _ := SearchUsers -> tool, _ := githubpkg.SearchUsers
// client := NewClient -> client := githubpkg.NewClient
