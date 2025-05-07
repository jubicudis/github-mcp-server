/*
 * WHO: TestUtilityProvider
 * WHAT: Helper functions for testing GitHub API integration
 * WHEN: During test execution
 * WHERE: System Layer 6 (Integration Tests)
 * WHY: To provide reusable test components
 * HOW: Using Go generics and test utilities
 * EXTENT: All GitHub MCP test files
 */

package testutil

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"tranquility-neuro-os/github-mcp-server/pkg/github"

	"github.com/mark3labs/mcp-go/types"
	"github.com/stretchr/testify/require"
)

// NOTE: Ptr function is already defined in testutil.go
// and is imported automatically since we're in the same package

// CreateTypesMCPRequest creates a mock MCP request with the provided args
// WHO: MCPRequestFactory
// WHAT: Test request creation for types package
// WHEN: During test execution
// WHERE: System Layer 6 (Testing)
// WHY: To simulate incoming MCP requests using types.MCPCall
// HOW: Using types.MCPCall structure
// EXTENT: All MCP handler tests requiring types.MCPCall
func CreateTypesMCPRequest(args map[string]interface{}) *types.MCPCall {
	return &types.MCPCall{
		Arguments: args,
	}
}

// StubGetClientWithTokenFn creates a test stub for the GetClientFn function with token support
// WHO: ClientStubProvider
// WHAT: GitHub client stub creation with token handling
// WHEN: During test execution setup
// WHERE: System Layer 6 (Testing)
// WHY: To provide controlled client behavior with auth token
// HOW: Using function wrapping with token parameter
// EXTENT: All GitHub client tests requiring token authentication
func StubGetClientWithTokenFn(client *github.Client) github.GetClientFn {
	return func(ctx context.Context, token string) (github.Client, error) {
		return *client, nil
	}
}

// ExpectQueryParams creates a middleware to check for expected query parameters
// WHO: QueryParamValidator
// WHAT: HTTP query parameter validation
// WHEN: During HTTP request testing
// WHERE: System Layer 6 (Testing)
// WHY: To verify correct query parameters
// HOW: Using HTTP middleware pattern
// EXTENT: All GitHub API request tests
func ExpectQueryParams(t *testing.T, expected map[string]string) *QueryParamValidator {
	return &QueryParamValidator{
		T:        t,
		Expected: expected,
	}
}

// QueryParamValidator validates HTTP query parameters
// WHO: QueryValidator
// WHAT: HTTP middleware for validation
// WHEN: During HTTP request handling
// WHERE: System Layer 6 (Testing)
// WHY: To verify request parameters
// HOW: Using http.Handler implementation
// EXTENT: All HTTP request tests
type QueryParamValidator struct {
	T        *testing.T
	Expected map[string]string
	Next     http.Handler
}

// AndThen chains to the next handler after validation
func (v *QueryParamValidator) AndThen(next http.Handler) http.Handler {
	v.Next = next
	return v
}

// ServeHTTP implements the http.Handler interface
func (v *QueryParamValidator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for k, expected := range v.Expected {
		actual := r.URL.Query().Get(k)
		if expected != actual {
			v.T.Errorf("Expected query param %q = %q, got %q", k, expected, actual)
		}
	}
	if v.Next != nil {
		v.Next.ServeHTTP(w, r)
	}
}

// CreateMockResponseHandler creates a handler that returns the provided status and body
// WHO: ResponseMocker
// WHAT: HTTP response simulation
// WHEN: During API testing
// WHERE: System Layer 6 (Testing)
// WHY: To provide controlled responses
// HOW: Using http.HandlerFunc
// EXTENT: All API endpoint tests
func CreateMockResponseHandler(t *testing.T, status int, body interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)

		if body != nil {
			data, err := json.Marshal(body)
			require.NoError(t, err)
			_, err = w.Write(data)
			require.NoError(t, err)
		}
	})
}

// GetTextResult extracts the text content from an MCP result
// WHO: ResultExtractor
// WHAT: MCP result parsing
// WHEN: After handler execution
// WHERE: System Layer 6 (Testing)
// WHY: To extract usable data from results
// HOW: Using type assertion
// EXTENT: All MCP handler tests
func GetTextResult(t *testing.T, result interface{}) *types.MCPTextContent {
	textContent, ok := result.(*types.MCPContent).Content.(*types.MCPTextContent)
	require.True(t, ok, "Expected text content in result")
	return textContent
}
