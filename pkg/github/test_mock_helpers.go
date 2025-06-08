/*
 * WHO: TestMockHelpers
 * WHAT: Test mock helper functions for GitHub tests
 * WHEN: During test execution
 * WHERE: MCP Bridge Layer Testing
 * WHY: To provide consistent mock implementations
 * HOW: By implementing test helper functions
 * EXTENT: All GitHub test mocks
 */
package github

import (
	"github-mcp-server/pkg/testutil"
	"net/http"
)

// Ptr provides a pointer to the provided value
func Ptr[T any](v T) *T {
	return &v
}

// mockResponse creates an HTTP handler that returns the provided response
// This is a compatibility wrapper around testutil.MockResponse
var mockResponse = testutil.MockResponse

// Removed QueryParamMatcher and related functions as they are now in testutil.go

// andThen chains a response handler after parameter validation
func andThen(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Call the next handler
		handler(w, r)
	}
}
