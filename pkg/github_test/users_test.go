// WHO: GitHubMCPUserTests
// WHAT: User API Testing
// WHEN: During test execution
// WHERE: MCP Bridge Layer Testing
// WHY: To verify user functionality
// HOW: By testing MCP protocol handlers
// EXTENT: All user operations
package ghmcp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/go-github/v71/github"
	ghmcp "github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/github"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/testutil"
	"github.com/mark3labs/mcp-go/mcp" // Added import for mcp.ErrorContent
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Canonical test file for users.go
// All tests must directly and robustly test the canonical logic in users.go
// Remove all legacy, duplicate, or non-canonical tests
// Reference only helpers from /pkg/common and /pkg/testutil
// No import cycles, duplicate imports, or undefined helpers
// All test cases must match the actual signatures and logic of users.go

func TestGetUser(t *testing.T) {
	// For tool definition testing, a placeholder client is used to initialize the tool's metadata.
	// This client is not used for actual API calls in the test cases below.
	toolDefinitionClientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(nil), nil }
	tool, _ := ghmcp.GetUser(toolDefinitionClientFn, testutil.NullTranslationHelperFunc)

	assert.Equal(t, "get_user", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.NotNil(t, tool.InputSchema) // Ensure InputSchema is not nil
	assert.Contains(t, tool.InputSchema.Properties, "username")
	// Username is optional, so it should not be in the Required list.
	// Depending on mcp.NewTool behavior, InputSchema.Required might be nil or empty if no params are mcp.Required().
	// Let's assert it's not marked as required if the field exists.
	if tool.InputSchema.Required != nil {
		assert.NotContains(t, tool.InputSchema.Required, "username")
	}

	// Setup mock user for success case
	mockUser := &github.User{
		Login:     testutil.Ptr("testuser"),
		ID:        testutil.Ptr(int64(12345)),
		Name:      testutil.Ptr("Test User"),
		Email:     testutil.Ptr("testuser@example.com"),
		Company:   testutil.Ptr("Test Inc."),
		Location:  testutil.Ptr("Test City"),
		HTMLURL:   testutil.Ptr("https://github.com/testuser"),
		Followers: testutil.Ptr(100),
		Following: testutil.Ptr(50),
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedUser   *github.User
		expectedErrMsg string
	}{
		{
			name: "successful user retrieval for authenticated user",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetUser,
					mockUser,
				),
			),
			requestArgs: map[string]interface{}{},
			expectError:  false,
			expectedUser: mockUser,
		},
		{
			name: "successful user retrieval for specified username",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetUsersByUsername,
					mockUser,
				),
			),
			requestArgs: map[string]interface{}{
				"username": "testuser",
			},
			expectError:  false,
			expectedUser: mockUser,
		},
		{
			name: "user not found",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetUsersByUsername,
					testutil.MockResponse(t, http.StatusNotFound, `{"message": "Not Found"}`),
				),
			),
			requestArgs: map[string]interface{}{
				"username": "nonexistentuser",
			},
			expectError:    true,
			expectedErrMsg: "User not found: nonexistentuser",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(tc.mockedClient), nil }
			_, handler := ghmcp.GetUser(clientFn, testutil.NullTranslationHelperFunc)
			request := testutil.CreateMCPRequest(tc.requestArgs)
			result, err := handler(context.Background(), request)

			if tc.expectError {
				require.NoError(t, err) // The handler itself should not return an error, the error is in the MCP result
				require.NotNil(t, result)
				assert.True(t, result.IsError)
				require.NotEmpty(t, result.Content, "Error result should have content")
				textContent, ok := result.Content[0].(mcp.TextContent)
				require.True(t, ok, "Error result content should be mcp.TextContent")
				assert.Contains(t, textContent.Text, tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			textContent := testutil.GetTextResult(t, result)
			var returnedUser github.User
			err = json.Unmarshal([]byte(textContent), &returnedUser)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedUser.Login, *returnedUser.Login)
			assert.Equal(t, *tc.expectedUser.ID, *returnedUser.ID)
			// Add other assertions as needed, e.g., Name, Email, if they are consistently populated
			if tc.expectedUser.Name != nil && returnedUser.Name != nil {
				assert.Equal(t, *tc.expectedUser.Name, *returnedUser.Name)
			}
			if tc.expectedUser.Email != nil && returnedUser.Email != nil {
				assert.Equal(t, *tc.expectedUser.Email, *returnedUser.Email)
			}
		})
	}
}
