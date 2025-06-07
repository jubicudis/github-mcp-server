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

	"github-mcp-server/pkg/common"

	"github.com/google/go-github/v71/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// NullTranslationHelperFunc returns a null translation helper function
// This bridges between the TranslationHelper interface and TranslationHelperFunc type
func NullTranslationHelperFunc(key string, defaultValue string) string {
	return defaultValue
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

		// For tests, when we don't have the actual client structure,
		// we can just return a basic client that won't be called
		return github.NewClient(nil), nil
	}
}

// TranslateFunc is a type that matches the translation function signature expected by tests
type TranslateFunc func(ctx context.Context, key string, args ...interface{}) string

// CreateTestTranslateFunc creates a simple translation function for tests
func CreateTestTranslateFunc() TranslateFunc {
	return func(ctx context.Context, key string, args ...interface{}) string {
		return key // Just return the key for testing
	}
}

// CreateTestTranslateFuncSimple returns a simple translation helper function
// Compatible with translations.TranslationHelperFunc signature
func CreateTestTranslateFuncSimple() func(key string, defaultValue string) string {
	return func(key string, defaultValue string) string {
		return key // Just return the key for testing
	}
}

// CreateMCPRequest creates an MCP request with the provided arguments
func CreateMCPRequest(args map[string]interface{}) mcp.CallToolRequest {
	req := mcp.CallToolRequest{}
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

// GetTextContent extracts text content from an MCP result and converts to JSON
// This is an alternative to GetTextResult that returns JSON-formatted content
func GetTextContent(t *testing.T, result *mcp.CallToolResult) string {
	require.NotNil(t, result)
	require.NotEmpty(t, result.Content, "CallToolResult has no content")

	// For testing purposes, we'll convert the first content item to JSON
	contentJSON, err := json.Marshal(result.Content[0])
	require.NoError(t, err)
	return string(contentJSON)
}

// PtrTo is a convenience function for creating a pointer to a value
// Imported from pkg/common/common.go but renamed to avoid generic issue
func PtrTo[T any](v T) *T {
	return common.Ptr(v)
}

// QueryParamMatcher is a middleware that matches query parameters
type QueryParamMatcher struct {
	t        *testing.T
	expected map[string]string
	next     http.Handler
	handler  http.HandlerFunc
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

// CreateQueryParamExpectation creates a middleware that validates query parameters
func CreateQueryParamExpectation(t *testing.T, expected map[string]string) *QueryParamMatcher {
	return CreateQueryParamMatcher(t, expected)
}

// MockResponse creates an HTTP handler that returns the provided response
func MockResponse(t *testing.T, statusCode int, body interface{}) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)

		if body == nil {
			return
		}

		var data []byte
		var err error

		switch v := body.(type) {
		case []byte:
			data = v
		case string:
			data = []byte(v)
		default:
			data, err = json.Marshal(body)
			require.NoError(t, err, "Failed to marshal response body")
		}

		_, err = w.Write(data)
		require.NoError(t, err, "Failed to write response body")
	}

	return http.HandlerFunc(handler)
}

// NewTestClient creates a GitHub client for testing
func NewTestClient(httpClient *http.Client) *github.Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return github.NewClient(httpClient)
}

// PartialMock is a utility for creating mock HTTP handlers with expectations
type PartialMock struct {
	T                   *testing.T
	ExpectedQueryParams map[string]string
	ExpectedRequestBody any
}

// ExpectQueryParams creates a PartialMock with expected query parameters
func ExpectQueryParams(t *testing.T, expectedQueryParams map[string]string) *PartialMock {
	return &PartialMock{
		T:                   t,
		ExpectedQueryParams: expectedQueryParams,
	}
}

// ExpectRequestBody creates a PartialMock with an expected request body
func ExpectRequestBody(t *testing.T, expectedRequestBody any) *PartialMock {
	return &PartialMock{
		T:                   t,
		ExpectedRequestBody: expectedRequestBody,
	}
}

// AndThen chains a response handler to the PartialMock after validating expectations
func (p *PartialMock) AndThen(responseHandler http.HandlerFunc) http.HandlerFunc {
	p.T.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if p.ExpectedRequestBody != nil {
			var unmarshaledRequestBody any
			err := json.NewDecoder(r.Body).Decode(&unmarshaledRequestBody)
			require.NoError(p.T, err)

			require.Equal(p.T, p.ExpectedRequestBody, unmarshaledRequestBody)
		}

		if p.ExpectedQueryParams != nil {
			require.Equal(p.T, len(p.ExpectedQueryParams), len(r.URL.Query()))
			for k, v := range p.ExpectedQueryParams {
				require.Equal(p.T, v, r.URL.Query().Get(k))
			}
		}

		responseHandler(w, r)
	}
}
