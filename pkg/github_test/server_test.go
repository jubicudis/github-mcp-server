// WHO: GitHubMCPBridgeTest
// WHAT: GitHub API Integration Package Tests
// WHEN: During test execution
// WHERE: MCP Bridge Layer Testing
// WHY: To verify GitHub API integration
// HOW: By testing MCP protocol handlers
// EXTENT: Server functionality verification
package ghmcp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v71/github"
	ghmcp "github.com/jubicudis/github-mcp-server/pkg/github"
	"github.com/jubicudis/github-mcp-server/pkg/testutil"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Canonical test file for server.go
// All tests must directly and robustly test the canonical logic in server.go
// Remove all legacy, duplicate, or non-canonical tests
// Reference only helpers from /pkg/common and /pkg/testutil
// No import cycles, duplicate imports, or undefined helpers
// All test cases must match the actual signatures and logic of server.go

// All parameter extraction and test helpers in this file use canonical implementations from pkg/common and pkg/github/testutil.
// This file is fully aligned with the 7D documentation and MCP architecture vision.

// Test constants for repeated literals
const (
	testUserLogin    = "testuser"
	testUserName     = "Test User"
	testUserBio      = "GitHub user for testing"
	testUserCompany  = "Test Company"
	testUserLocation = "Test Location"
	testUserHTMLURL  = "https://github.com/testuser"
	testUserType     = "User"
	testUserPlan     = "pro"
	testValue        = "test-value"
	missingParam     = "missing parameter"
	wrongTypeParam   = "wrong type parameter"
	validNumParam    = "valid number parameter"
	notANumber       = "not-a-number"
)

// STUB: testutil.Ptr is undefined, so we stub it here for test purposes
// func Ptr[T any](v T) *T { return &v } // Removed Ptr stub

// STUB: common, githubpkg.Timestamp, githubpkg.Plan, githubpkg.ErrorResponse, githubpkg.AcceptedError are undefined, so we stub them for test purposes
// var common = struct{}{} // Removed common stub
// type Timestamp struct{ Time time.Time } // Removed Timestamp stub
// type Plan struct{ Name *string } // Removed Plan stub
// type ErrorResponse struct{} // Removed ErrorResponse stub
// type AcceptedError struct{} // Removed AcceptedError stub

func TestGetMe(t *testing.T) {
	// Verify tool definition
	defaultClient := github.NewClient(nil)
	clientFn := func(ctx context.Context) (*github.Client, error) { return defaultClient, nil }
	tool, _ := ghmcp.GetMe(clientFn, testutil.NullTranslationHelperFunc)

	assert.Equal(t, "get_me", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "reason")
	assert.Empty(t, tool.InputSchema.Required) // No required parameters

	// Setup mock user response
	mockUser := &github.User{ // Changed githubpkg.User to github.User
		Login:     testutil.Ptr(testUserLogin),
		Name:      testutil.Ptr(testUserName),
		Email:     testutil.Ptr(testUserEmail),
		Bio:       testutil.Ptr(testUserBio),
		Company:   testutil.Ptr(testUserCompany),
		Location:  testutil.Ptr(testUserLocation),
		HTMLURL:   testutil.Ptr(testUserHTMLURL),
		CreatedAt: &github.Timestamp{Time: time.Now().Add(-365 * 24 * time.Hour)}, // Changed githubpkg.Timestamp to github.Timestamp
		Type:      testutil.Ptr(testUserType),
		Plan: &github.Plan{ // Changed githubpkg.Plan to github.Plan
			Name: testutil.Ptr(testUserPlan),
		},
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedUser   *github.User // Changed githubpkg.User to github.User
		expectedErrMsg string
	}{
		{
			name: "successful get user",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetUser,
					mockUser,
				),
			),
			requestArgs:  map[string]interface{}{},
			expectError:  false,
			expectedUser: mockUser,
		},
		{
			name: "successful get user with reason",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetUser,
					mockUser,
				),
			),
			requestArgs: map[string]interface{}{
				"reason": "Testing API",
			},
			expectError:  false,
			expectedUser: mockUser,
		},
		{
			name: "get user fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetUser,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusUnauthorized)
						_, _ = w.Write([]byte(`{"message": "Unauthorized"}`))
					}),
				),
			),
			requestArgs:    map[string]interface{}{},
			expectError:    true,
			expectedErrMsg: "failed to get user",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			clientFn := func(ctx context.Context) (*github.Client, error) { return client, nil }
			_, handler := ghmcp.GetMe(clientFn, testutil.NullTranslationHelperFunc)

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

			// Parse result and get text content if no error
			text := testutil.GetTextResult(t, result)

			// Unmarshal and verify the result
			var returnedUser github.User // Changed githubpkg.User to github.User
			err = json.Unmarshal([]byte(text), &returnedUser)
			require.NoError(t, err)

			// Verify user details
			assert.Equal(t, *tc.expectedUser.Login, *returnedUser.Login)
			assert.Equal(t, *tc.expectedUser.Name, *returnedUser.Name)
			assert.Equal(t, *tc.expectedUser.Email, *returnedUser.Email)
			assert.Equal(t, *tc.expectedUser.Bio, *returnedUser.Bio)
			assert.Equal(t, *tc.expectedUser.HTMLURL, *returnedUser.HTMLURL)
			assert.Equal(t, *tc.expectedUser.Type, *returnedUser.Type)
		})
	}
}
