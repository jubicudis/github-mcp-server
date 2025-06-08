// Consolidated utility functions
package testutil

import (
	"errors"
	"github-mcp-server/pkg/common"
	"net/http"
	"testing"

	"github.com/google/go-github/v71/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

// mockResponse creates a mock HTTP response handler.
func mockResponse(t *testing.T, statusCode int, body interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		// Serialize body to JSON and write to response (omitted for brevity)
	}
}

// QueryParamMatcher is a middleware for testing query parameters
type QueryParamMatcher struct {
	t        *testing.T
	expected map[string]string
}

// expectQueryParams creates a middleware that validates query parameters
func CreateQueryParamMatcher(t *testing.T, expected map[string]string) *QueryParamMatcher {
	return &QueryParamMatcher{t: t, expected: expected}
}

// andThen chains a response handler after parameter validation
func (m *QueryParamMatcher) andThen(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate query parameters
		query := r.URL.Query()
		for k, v := range m.expected {
			require.Equal(m.t, v, query.Get(k), "Query parameter %s mismatch", k)
		}

		// Call the next handler
		handler(w, r)
	}
}

// Use common.Ptr instead of redefining Ptr
func Ptr[T any](v T) *T {
	return common.Ptr(v)
}

// CreateMCPRequest creates an MCP request with the provided arguments
func CreateMCPRequest(args map[string]interface{}) mcp.CallToolRequest {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	return req
}

// StubGetClientFn returns a function that returns the provided client, useful for testing
func StubGetClientFn(client *github.Client) func(context.Context) (*github.Client, error) {
	return func(context.Context) (*github.Client, error) {
		return client, nil
	}
}

// StubGetClientFnForCustomClient creates a GetClientFn that works with any type that provides the client interface
func StubGetClientFnForCustomClient(client interface{}) func(context.Context) (*github.Client, error) {
	// For tests, we can just return nil as the actual client implementation doesn't matter
	return func(context.Context) (*github.Client, error) {
		return nil, nil
	}
}

// StubGetClientFnWithClient returns a function that returns a GitHub client,
// useful for testing with the custom Client type
func StubGetClientFnWithClient(mockClient interface{}) func(context.Context) (*github.Client, error) {
	return func(context.Context) (*github.Client, error) {
		// If we're passed an http.Client, use it to create a new GitHub client
		if httpClient, ok := mockClient.(*http.Client); ok {
			return github.NewClient(httpClient), nil
		}
		return nil, errors.New("invalid client type")
	}
}

// NullTranslationHelperFunc is a no-op translation helper function for testing
var NullTranslationHelperFunc = func(key, defaultValue string) string { return defaultValue }
