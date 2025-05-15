// WHO: GitHubMCPTestUtilities
// WHAT: Test Helper Functions
// WHEN: During test execution
// WHERE: MCP Bridge Layer Testing
// WHY: To provide reusable test utilities
// HOW: By implementing common test functions
// EXTENT: All test files
package github

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/go-github/v49/github"
	mcp_go "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tranquility-neuro-os/github-mcp-server/pkg/github/testutil"
)

// Test constants
const (
	testNameParamNotPresent = "parameter not present"
)

// WHO: TestUtility
// WHAT: Client function stub for GitHub API
// WHEN: During testing
// WHERE: Test execution context
// WHY: To isolate tests from actual GitHub API
// HOW: Using function closure with predefined client
// EXTENT: For all GitHub API test cases
func stubGetClientFn(client *Client) GetClientFn {
	// Convert our custom Client to go-github Client
	githubClient := &github.Client{}

	return func(context.Context) (*github.Client, error) {
		return githubClient, nil
	}
}

// WHO: TestUtility
// WHAT: GitHub API client factory for tests
// WHEN: During test setup
// WHERE: Test execution context
// WHY: To provide mock clients with consistent interface
// HOW: Using httpclient wrapping
// EXTENT: For all GitHub API test cases
func NewTestClient(httpClient *http.Client) *github.Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return github.NewClient(httpClient)
}

// expectQueryParams is a helper function to create a partial mock that expects a
// request with the given query parameters, with the ability to chain a response handler.
func expectQueryParams(t *testing.T, expectedQueryParams map[string]string) *partialMock {
	// WHO: TestUtility
	// WHAT: Query parameter validation
	// WHEN: During HTTP request testing
	// WHERE: Mock HTTP handler
	// WHY: To validate request parameters
	// HOW: Using request inspection
	// EXTENT: For all HTTP param tests
	return &partialMock{
		t:                   t,
		expectedQueryParams: expectedQueryParams,
	}
}

// expectRequestBody is a helper function to create a partial mock that expects a
// request with the given body, with the ability to chain a response handler.
func expectRequestBody(t *testing.T, expectedRequestBody any) *partialMock {
	return &partialMock{
		t:                   t,
		expectedRequestBody: expectedRequestBody,
	}
}

type partialMock struct {
	t                   *testing.T
	expectedQueryParams map[string]string
	expectedRequestBody any
}

func (p *partialMock) andThen(responseHandler http.HandlerFunc) http.HandlerFunc {
	p.t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if p.expectedRequestBody != nil {
			var unmarshaledRequestBody any
			err := json.NewDecoder(r.Body).Decode(&unmarshaledRequestBody)
			require.NoError(p.t, err)

			require.Equal(p.t, p.expectedRequestBody, unmarshaledRequestBody)
		}

		if p.expectedQueryParams != nil {
			require.Equal(p.t, len(p.expectedQueryParams), len(r.URL.Query()))
			for k, v := range p.expectedQueryParams {
				require.Equal(p.t, v, r.URL.Query().Get(k))
			}
		}

		responseHandler(w, r)
	}
}

// mockResponse is a helper function to create a mock HTTP response handler
// that returns a specified status code and marshaled body.
func mockResponse(t *testing.T, code int, body interface{}) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(code)
		b, err := json.Marshal(body)
		require.NoError(t, err)
		_, _ = w.Write(b)
	}
}

// Use testutil.CreateMCPRequest instead of this local function

// getTextResult is a helper function that returns a text result from a tool call.
func getTextResult(t *testing.T, result *mcp_go.CallToolResult) string {
	t.Helper()
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Content, "CallToolResult has no content")

	// In a real implementation, we would extract text from various content types
	// But for our tests, we'll marshal the content as JSON and return that
	contentJson, err := json.Marshal(result.Content[0])
	require.NoError(t, err)
	return string(contentJson)
}

// WHO: TestUtility
// WHAT: Helper function to check parameter validation
// WHEN: During test execution
// WHERE: Test context for OptionalParamOK
// WHY: To reduce cognitive complexity
// HOW: By extracting common test logic
// EXTENT: All parameter validation tests
func testTypedParam[T any](t *testing.T, args map[string]interface{},
	paramName string, expectedVal T, expectedOk bool,
	expectError bool, errorMsg string) {

	t.Helper()
	request := testutil.CreateMCPRequest(args)
	val, ok, err := OptionalParamOK[T](request, paramName)

	if expectError {
		require.Error(t, err)
		assert.Contains(t, err.Error(), errorMsg)
	} else {
		require.NoError(t, err)
	}
	assert.Equal(t, expectedOk, ok)
	assert.Equal(t, expectedVal, val)
}

// Test string parameters
func TestOptionalParamOKString(t *testing.T) {
	tests := []struct {
		name        string
		args        map[string]interface{}
		paramName   string
		expectedVal string
		expectedOk  bool
		expectError bool
		errorMsg    string
	}{
		{
			name:        "present and correct type (string)",
			args:        map[string]interface{}{"myParam": "hello"},
			paramName:   "myParam",
			expectedVal: "hello",
			expectedOk:  true,
			expectError: false,
		},
		{
			name:        "present but wrong type (string expected, got bool)",
			args:        map[string]interface{}{"myParam": true},
			paramName:   "myParam",
			expectedVal: "",
			expectedOk:  false,
			expectError: false,
		},
		{
			name:        testNameParamNotPresent,
			args:        map[string]interface{}{"anotherParam": "value"},
			paramName:   "myParam",
			expectedVal: "",
			expectedOk:  false,
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testTypedParam(t, tc.args, tc.paramName, tc.expectedVal, tc.expectedOk, tc.expectError, tc.errorMsg)
		})
	}
}

// Test bool parameters
func TestOptionalParamOKBool(t *testing.T) {
	tests := []struct {
		name        string
		args        map[string]interface{}
		paramName   string
		expectedVal bool
		expectedOk  bool
		expectError bool
		errorMsg    string
	}{
		{
			name:        "present and correct type (bool)",
			args:        map[string]interface{}{"myParam": true},
			paramName:   "myParam",
			expectedVal: true,
			expectedOk:  true,
			expectError: false,
		},
		{
			name:        "present but wrong type (bool expected, got string)",
			args:        map[string]interface{}{"myParam": "true"},
			paramName:   "myParam",
			expectedVal: false,
			expectedOk:  false,
			expectError: false,
		},
		{
			name:        testNameParamNotPresent,
			args:        map[string]interface{}{"anotherParam": "value"},
			paramName:   "myParam",
			expectedVal: false,
			expectedOk:  false,
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testTypedParam(t, tc.args, tc.paramName, tc.expectedVal, tc.expectedOk, tc.expectError, tc.errorMsg)
		})
	}
}

// Test float64 parameters
func TestOptionalParamOKFloat64(t *testing.T) {
	tests := []struct {
		name        string
		args        map[string]interface{}
		paramName   string
		expectedVal float64
		expectedOk  bool
		expectError bool
		errorMsg    string
	}{
		{
			name:        "present and correct type (number)",
			args:        map[string]interface{}{"myParam": float64(123)},
			paramName:   "myParam",
			expectedVal: 123,
			expectedOk:  true,
			expectError: false,
		},
		{
			name:        testNameParamNotPresent,
			args:        map[string]interface{}{"anotherParam": "value"},
			paramName:   "myParam",
			expectedVal: 0,
			expectedOk:  false,
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testTypedParam(t, tc.args, tc.paramName, tc.expectedVal, tc.expectedOk, tc.expectError, tc.errorMsg)
		})
	}
}
