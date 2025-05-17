/*
 * WHO: GitHubMCPScanningTests
 * WHAT: Code Scanning API Testing
 * WHEN: During test execution
 * WHERE: MCP Bridge Layer Testing
 * WHY: To verify code scanning functionality
 * HOW: By testing MCP protocol handlers
 * EXTENT: All code scanning operations
 */
package github

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jubicudis/github-mcp-server/pkg/github/testutil"

	"github.com/google/go-github/v49/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCodeScanningAlert(t *testing.T) {
	// Verify tool definition once
	mockClient := mock.NewMockedHTTPClient()
	translateFn := testutil.CreateTestTranslateFunc()
	adaptedTranslateFn := func(key string, defaultValue string) string {
		return translateFn(context.Background(), key)
	}
	tool, _ := GetCodeScanningAlert(testutil.StubGetClientFnWithClient(mockClient), adaptedTranslateFn)
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
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedAlert  *github.Alert
		expectedErrMsg string
	}{
		{
			name: "successful alert fetch",
			mockedClient: mock.NewMockedHTTPClient(
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
			expectError:   false,
			expectedAlert: mockAlert,
		},
		{
			name: "alert fetch fails",
			mockedClient: mock.NewMockedHTTPClient(
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
			expectError:    true,
			expectedErrMsg: "failed to get alert",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup with mock client and translation function
			translateFn := func(key string, defaultValue string) string {
				return key // Simple translation function for testing
			}
			_, handler := GetCodeScanningAlert(testutil.StubGetClientFnWithClient(tc.mockedClient), translateFn)

			// Create call request
			request := testutil.CreateMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)

			// Parse the result and get the text content if no error
			textContent := testutil.GetTextResult(t, result)

			// Unmarshal and verify the result
			var returnedAlert github.Alert
			err = json.Unmarshal([]byte(textContent), &returnedAlert)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedAlert.Number, *returnedAlert.Number)
			assert.Equal(t, *tc.expectedAlert.State, *returnedAlert.State)
			assert.Equal(t, *tc.expectedAlert.Rule.ID, *returnedAlert.Rule.ID)
			assert.Equal(t, *tc.expectedAlert.HTMLURL, *returnedAlert.HTMLURL)
		})
	}
}

func TestListCodeScanningAlerts(t *testing.T) {
	// Verify tool definition once
	mockHttpClient := mock.NewMockedHTTPClient()
	translateFn := func(key string, defaultValue string) string {
		return key // Simple translation function for testing
	}
	tool, _ := ListCodeScanningAlerts(testutil.StubGetClientFnWithClient(mockHttpClient), translateFn)
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
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedAlerts []*github.Alert
		expectedErrMsg string
	}{
		{
			name: "successful alerts listing",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposCodeScanningAlertsByOwnerByRepo,
					testutil.ExpectQueryParams(t, map[string]string{
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
			expectError:    false,
			expectedAlerts: mockAlerts,
		},
		{
			name: "alerts listing fails",
			mockedClient: mock.NewMockedHTTPClient(
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
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup with mock client and translation function
			translateFn := func(key string, defaultValue string) string {
				return key // Simple translation function for testing
			}
			_, handler := ListCodeScanningAlerts(testutil.StubGetClientFnWithClient(tc.mockedClient), translateFn)
			request := testutil.CreateMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)

			// Parse the result and get the text content if no error
			textContent := testutil.GetTextResult(t, result)

			// Unmarshal and verify the result
			var returnedAlerts []*github.Alert
			err = json.Unmarshal([]byte(textContent), &returnedAlerts)
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
