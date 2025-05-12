/*
 * WHO: GitHubMCPTestUtils
 * WHAT: Test utilities for GitHub MCP
 * WHEN: During test execution
 * WHERE: MCP Bridge Layer Testing
 * WHY: To support test implementations
 * HOW: By providing common test functions
 * EXTENT: All GitHub MCP testing
 */
package testutil

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/go-github/v49/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// StubGetClientFn returns a function that returns the provided client, useful for testing
func StubGetClientFn(client *github.Client) func(context.Context) (*github.Client, error) {
	return func(context.Context) (*github.Client, error) {
		return client, nil
	}
}

// Note: StubGetClientWithTokenFn is defined in test_helpers.go

// CreateMCPRequest creates an MCP request with the provided arguments
func CreateMCPRequest(args map[string]interface{}) *mcp.CallToolRequest {
	req := &mcp.CallToolRequest{}
	req.Params.Arguments = args
	return req
}

// GetTextResult extracts text content from an MCP result
func GetTextResult(t *testing.T, result *mcp.CallToolResult) string {
	require.NotNil(t, result)

	// Check if Content slice is available and has elements
	if len(result.Content) > 0 {
		// Extract the text content from the first content element
		// Assuming the first content is TextContent
		textContent, ok := result.Content[0].(mcp.TextContent)
		if ok {
			return textContent.Text
		}
	}
	return ""
}

// Ptr returns a pointer to the provided value
func Ptr[T any](v T) *T {
	return &v
}

// QueryParamMatcher is a middleware that matches query parameters
type QueryParamMatcher struct {
	t        *testing.T
	expected map[string]string
	next     http.Handler
}

// CreateQueryParamMatcher creates a middleware that verifies query parameters and then calls the next handler
func CreateQueryParamMatcher(t *testing.T, expected map[string]string) *QueryParamMatcher {
	return &QueryParamMatcher{
		t:        t,
		expected: expected,
	}
}

// AndThen chains the next handler to be called after parameter validation
func (m *QueryParamMatcher) AndThen(next http.Handler) http.Handler {
	m.next = next
	return m
}

// ServeHTTP implements the http.Handler interface
func (m *QueryParamMatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	for k, v := range m.expected {
		require.Equal(m.t, v, query.Get(k), "Query parameter %s mismatch", k)
	}
	if m.next != nil {
		m.next.ServeHTTP(w, r)
	}
}

// MockResponse creates an HTTP handler that returns the provided response
func MockResponse(t *testing.T, statusCode int, body interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		if body != nil {
			bytes, err := json.Marshal(body)
			require.NoError(t, err)
			_, err = w.Write(bytes)
			require.NoError(t, err)
		}
	})
}
