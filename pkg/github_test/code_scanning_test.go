/*
 * WHO: GitHubMCPScanningTests
 * WHAT: Code Scanning API Testing
 * WHEN: During test execution
 * WHERE: MCP Bridge Layer Testing
 * WHY: To verify code scanning functionality
 * HOW: By testing MCP protocol handlers
 * EXTENT: All code scanning operations
 */
package ghmcp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/go-github/v71/github"
	ghmcp "github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/github"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/testutil"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Canonical test file for code_scanning.go
// All tests must directly and robustly test the canonical logic in code_scanning.go
// Remove all legacy, duplicate, or non-canonical tests
// Reference only helpers from /pkg/common and /pkg/testutil
// No import cycles, duplicate imports, or undefined helpers
// All test cases must match the actual signatures and logic of code_scanning.go

// Using functions from the github package

func TestGetCodeScanningAlert(t *testing.T) {
	// Verify tool definition once
	mockGHClient := github.NewClient(mock.NewMockedHTTPClient())
	// Use testutil.NullTranslationHelperFunc as it has the correct signature
	toolDefTranslateFn := testutil.NullTranslationHelperFunc
	clientFn := func(ctx context.Context) (*github.Client, error) { return mockGHClient, nil }
	tool, _ := ghmcp.GetCodeScanningAlert(clientFn, toolDefTranslateFn)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "alertNumber")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "alertNumber"})

	// Setup mock alert for success case
	mockAlert := &github.Alert{
		Number:  testutil.Ptr(42),
		State:   testutil.Ptr("open"),
		Rule:    &github.Rule{ID: testutil.Ptr("test-rule"), Description: testutil.Ptr("Test Rule Description")},
		HTMLURL: testutil.Ptr("https://github.com/owner/repo/security/code-scanning/42"),
	}

	tests := []struct {
		name                 string
		mockedHTTPClient     *http.Client
		requestArgs          map[string]interface{}
		expectToolError      bool
		expectedToolErrMsg   string
		expectHandlerError   bool
		expectedHandlerErrMsg string
		expectedAlert        *github.Alert
	}{
		{
			name: "successful alert fetch",
			mockedHTTPClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposCodeScanningAlertsByOwnerByRepoByAlertNumber,
					mockAlert,
				),
			),
			requestArgs: map[string]interface{}{
				"owner":       "owner",
				"repo":        "repo",
				"alertNumber": float64(42),
			},
			expectToolError: false,
			expectedAlert:   mockAlert,
		},
		{
			name: "alert fetch fails",
			mockedHTTPClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposCodeScanningAlertsByOwnerByRepoByAlertNumber,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":       "owner",
				"repo":        "repo",
				"alertNumber": float64(9999),
			},
			expectToolError:    true,
			expectedToolErrMsg: "failed to get alert", // Matches default message in CreateErrorResponse
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Corrected signature for translateFn used by the handler
			handlerTranslateFn := func(key string, defaultValue string) string {
				return key // Causes CreateErrorResponse to use its default message format
			}
			ghClient := github.NewClient(tc.mockedHTTPClient)
			clientFn := func(ctx context.Context) (*github.Client, error) { return ghClient, nil }
			_, handler := ghmcp.GetCodeScanningAlert(clientFn, handlerTranslateFn)

			request := testutil.CreateMCPRequest(tc.requestArgs)
			result, handlerErr := handler(context.Background(), request)

			if tc.expectHandlerError {
				require.Error(t, handlerErr)
				assert.Contains(t, handlerErr.Error(), tc.expectedHandlerErrMsg)
				return
			}
			require.NoError(t, handlerErr, "Handler itself returned an unexpected error")
			require.NotNil(t, result, "Result should not be nil if handler error is nil")

			if tc.expectToolError {
				assert.True(t, result.IsError, "result.IsError should be true for a tool error")
				require.NotEmpty(t, result.Content, "result.Content should not be empty for an error")
				textContent, ok := result.Content[0].(mcp.TextContent)
				require.True(t, ok, "Expected mcp.TextContent for error message")
				assert.Contains(t, textContent.Text, tc.expectedToolErrMsg, "Error message mismatch")
				return
			}

			// Success case
			assert.False(t, result.IsError, "result.IsError should be false for a successful call")
			text := testutil.GetTextResult(t, result)

			var returnedAlert github.Alert
			err := json.Unmarshal([]byte(text), &returnedAlert)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedAlert.Number, *returnedAlert.Number)
			assert.Equal(t, *tc.expectedAlert.State, *returnedAlert.State)
			assert.Equal(t, *tc.expectedAlert.Rule.ID, *returnedAlert.Rule.ID)
			assert.Equal(t, *tc.expectedAlert.HTMLURL, *returnedAlert.HTMLURL)
		})
	}
}

func TestListCodeScanningAlerts(t *testing.T) {
	mockGHClient := github.NewClient(mock.NewMockedHTTPClient())
	// Use testutil.NullTranslationHelperFunc as it has the correct signature
	toolDefTranslateFn := testutil.NullTranslationHelperFunc
	clientFn := func(ctx context.Context) (*github.Client, error) { return mockGHClient, nil }
	tool, _ := ghmcp.ListCodeScanningAlerts(clientFn, toolDefTranslateFn)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "ref")
	assert.Contains(t, tool.InputSchema.Properties, "state")
	assert.Contains(t, tool.InputSchema.Properties, "severity")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo"})

	// Setup mock alerts for success case
	mockAlerts := []*github.Alert{
		{
			Number:  testutil.Ptr(42),
			State:   testutil.Ptr("open"),
			Rule:    &github.Rule{ID: testutil.Ptr("test-rule-1"), Description: testutil.Ptr("Test Rule 1")},
			HTMLURL: testutil.Ptr("https://github.com/owner/repo/security/code-scanning/42"),
		},
		{
			Number:  testutil.Ptr(43),
			State:   testutil.Ptr("fixed"),
			Rule:    &github.Rule{ID: testutil.Ptr("test-rule-2"), Description: testutil.Ptr("Test Rule 2")},
			HTMLURL: testutil.Ptr("https://github.com/owner/repo/security/code-scanning/43"),
		},
	}

	tests := []struct {
		name                  string
		mockedHTTPClient      *http.Client
		requestArgs           map[string]interface{}
		expectToolError       bool
		expectedToolErrMsg    string
		expectHandlerError    bool
		expectedHandlerErrMsg  string
		expectedAlerts        []*github.Alert
	}{
		{
			name: "successful alerts listing",
			mockedHTTPClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposCodeScanningAlertsByOwnerByRepo,
					testutil.CreateQueryParamMatcher(t, map[string]string{
						"ref":      "main",
						"state":    "open",
						"severity": "high",
					}).AndThen(
						testutil.MockResponse(t, http.StatusOK, mockAlerts),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":    "owner",
				"repo":     "repo",
				"ref":      "main",
				"state":    "open",
				"severity": "high",
			},
			expectToolError: false,
			expectedAlerts:  mockAlerts,
		},
		{
			name: "alerts listing fails",
			mockedHTTPClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposCodeScanningAlertsByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusUnauthorized)
						_, _ = w.Write([]byte(`{"message": "Unauthorized access"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner": "owner",
				"repo":  "repo",
			},
			expectToolError:    true,
			expectedToolErrMsg: "failed to list alerts", // Matches default message in CreateErrorResponse
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Corrected signature for translateFn used by the handler
			handlerTranslateFn := func(key string, defaultValue string) string {
				return key // Causes CreateErrorResponse to use its default message format
			}
			ghClient := github.NewClient(tc.mockedHTTPClient)
			clientFn := func(ctx context.Context) (*github.Client, error) { return ghClient, nil }
			_, handler := ghmcp.ListCodeScanningAlerts(clientFn, handlerTranslateFn)
			request := testutil.CreateMCPRequest(tc.requestArgs)

			result, handlerErr := handler(context.Background(), request)

			if tc.expectHandlerError {
				require.Error(t, handlerErr)
				assert.Contains(t, handlerErr.Error(), tc.expectedHandlerErrMsg)
				return
			}
			require.NoError(t, handlerErr, "Handler itself returned an unexpected error")
			require.NotNil(t, result, "Result should not be nil if handler error is nil")

			if tc.expectToolError {
				assert.True(t, result.IsError, "result.IsError should be true for a tool error")
				require.NotEmpty(t, result.Content, "result.Content should not be empty for an error")
				textContent, ok := result.Content[0].(mcp.TextContent)
				require.True(t, ok, "Expected mcp.TextContent for error message")
				assert.Contains(t, textContent.Text, tc.expectedToolErrMsg, "Error message mismatch")
				return
			}

			assert.False(t, result.IsError, "result.IsError should be false for a successful call")
			text := testutil.GetTextResult(t, result)

			var returnedAlerts []*github.Alert
			err := json.Unmarshal([]byte(text), &returnedAlerts)
			require.NoError(t, err)
			assert.Len(t, returnedAlerts, len(tc.expectedAlerts))
			for i, alert := range returnedAlerts {
				assert.Equal(t, *tc.expectedAlerts[i].Number, *alert.Number)
				assert.Equal(t, *tc.expectedAlerts[i].State, *alert.State)
				assert.Equal(t, *tc.expectedAlerts[i].Rule.ID, *alert.Rule.ID)
				assert.Equal(t, *tc.expectedAlerts[i].HTMLURL, *alert.HTMLURL)
			}
		})
	}
}


