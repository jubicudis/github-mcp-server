// WHO: GitHubMCPSearchTests
// WHAT: Search API Testing
// WHEN: During test execution
// WHERE: MCP Bridge Layer Testing
// WHY: To verify search functionality
// HOW: By testing MCP protocol handlers
// EXTENT: All search operations
package ghmcp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/go-github/v71/github"
	ghmcp "github.com/jubicudis/github-mcp-server/pkg/github"
	"github.com/jubicudis/github-mcp-server/pkg/testutil"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constants for repeated literals
const (
	repoSearchQuery = "golang test"
	invalidQuery    = "invalid:query"
	codeSearchQuery = "fmt.Println language:go"
	userSearchQuery = "location:finland language:go"
)

type repoTestCase struct {
	name           string
	mockedClient   *http.Client
	requestArgs    map[string]interface{}
	expectError    bool
	expectedResult *github.RepositoriesSearchResult
	expectedErrMsg string
}

func TestSearchRepositories(t *testing.T) {
	// Use a default *github.Client for metadata assertions
	defaultClient := github.NewClient(nil)
	clientFn := func(ctx context.Context) (*github.Client, error) { return defaultClient, nil }
	tool, _ := ghmcp.SearchRepositories(clientFn, testutil.NullTranslationHelperFunc)

	assert.Equal(t, "search_repositories", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "query")
	assert.Contains(t, tool.InputSchema.Properties, "page")
	assert.Contains(t, tool.InputSchema.Properties, "perPage")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"query"})

	// Setup mock search results
	mockSearchResult := &github.RepositoriesSearchResult{
		Total:             testutil.Ptr(2),
		IncompleteResults: testutil.Ptr(false),
		Repositories: []*github.Repository{
			{
				ID:              testutil.Ptr(int64(12345)),
				Name:            testutil.Ptr("repo-1"),
				FullName:        testutil.Ptr("owner/repo-1"),
				HTMLURL:         testutil.Ptr("https://github.com/owner/repo-1"),
				Description:     testutil.Ptr("Test repository 1"),
				StargazersCount: testutil.Ptr(100),
			},
			{
				ID:              testutil.Ptr(int64(67890)),
				Name:            testutil.Ptr("repo-2"),
				FullName:        testutil.Ptr("owner/repo-2"),
				HTMLURL:         testutil.Ptr("https://github.com/owner/repo-2"),
				Description:     testutil.Ptr("Test repository 2"),
				StargazersCount: testutil.Ptr(200),
			},
		},
	}

	tests := []repoTestCase{
		{
			name: "successful repository search",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetSearchRepositories,
					testutil.CreateQueryParamMatcher(t, map[string]string{
						"q":        repoSearchQuery,
						"page":     "2",
						"per_page": "10",
					}).AndThen(
						testutil.MockResponse(t, http.StatusOK, mockSearchResult),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"query":   repoSearchQuery,
				"page":    float64(2),
				"perPage": float64(10),
			},
			expectError:    false,
			expectedResult: mockSearchResult,
		},
		{
			name: "repository search with default pagination",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetSearchRepositories,
					testutil.CreateQueryParamMatcher(t, map[string]string{
						"q":        repoSearchQuery,
						"page":     "1",
						"per_page": "30",
					}).AndThen(
						testutil.MockResponse(t, http.StatusOK, mockSearchResult),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"query": repoSearchQuery,
			},
			expectError:    false,
			expectedResult: mockSearchResult,
		},
		{
			name: "search fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetSearchRepositories,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusBadRequest)
						_, _ = w.Write([]byte(`{"message": "Invalid query"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"query": invalidQuery,
			},
			expectError:    true,
			expectedErrMsg: "failed to search repositories",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock HTTP and wrap in *github.Client
			client := github.NewClient(tc.mockedClient)
			clientFn := func(ctx context.Context) (*github.Client, error) { return client, nil }
			_, handler := ghmcp.SearchRepositories(clientFn, testutil.NullTranslationHelperFunc)

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
			text := testutil.GetTextResult(t, result)

			// Unmarshal and verify the result
			var returnedResult github.RepositoriesSearchResult
			err = json.Unmarshal([]byte(text), &returnedResult)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedResult.Total, *returnedResult.Total)
			assert.Equal(t, *tc.expectedResult.IncompleteResults, *returnedResult.IncompleteResults)
			assert.Len(t, returnedResult.Repositories, len(tc.expectedResult.Repositories))
			for i, repo := range returnedResult.Repositories {
				assert.Equal(t, *tc.expectedResult.Repositories[i].ID, *repo.ID)
				assert.Equal(t, *tc.expectedResult.Repositories[i].Name, *repo.Name)
				assert.Equal(t, *tc.expectedResult.Repositories[i].FullName, *repo.FullName)
				assert.Equal(t, *tc.expectedResult.Repositories[i].HTMLURL, *repo.HTMLURL)
			}

		})
	}
}

type codeTestCase struct {
	name           string
	mockedClient   *http.Client
	requestArgs    map[string]interface{}
	expectError    bool
	expectedResult *github.CodeSearchResult
	expectedErrMsg string
}

func TestSearchCode(t *testing.T) {
	// Use a default *github.Client for metadata assertions
	defaultClient := github.NewClient(nil)
	clientFn := func(ctx context.Context) (*github.Client, error) { return defaultClient, nil }
	tool, _ := ghmcp.SearchCode(clientFn, testutil.NullTranslationHelperFunc)

	assert.Equal(t, "search_code", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "q")
	assert.Contains(t, tool.InputSchema.Properties, "sort")
	assert.Contains(t, tool.InputSchema.Properties, "order")
	assert.Contains(t, tool.InputSchema.Properties, "perPage")
	assert.Contains(t, tool.InputSchema.Properties, "page")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"q"})

	// Setup mock search results
	mockSearchResult := &github.CodeSearchResult{
		Total:             testutil.Ptr(2),
		IncompleteResults: testutil.Ptr(false),
		CodeResults: []*github.CodeResult{
			{
				Name:       testutil.Ptr("file1.go"),
				Path:       testutil.Ptr("path/to/file1.go"),
				SHA:        testutil.Ptr("abc123def456"),
				HTMLURL:    testutil.Ptr("https://github.com/owner/repo/blob/main/path/to/file1.go"),
				Repository: &github.Repository{Name: testutil.Ptr("repo"), FullName: testutil.Ptr("owner/repo")},
			},
			{
				Name:       testutil.Ptr("file2.go"),
				Path:       testutil.Ptr("path/to/file2.go"),
				SHA:        testutil.Ptr("def456abc123"),
				HTMLURL:    testutil.Ptr("https://github.com/owner/repo/blob/main/path/to/file2.go"),
				Repository: &github.Repository{Name: testutil.Ptr("repo"), FullName: testutil.Ptr("owner/repo")},
			},
		},
	}

	tests := []codeTestCase{
		{
			name: "successful code search with all parameters",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetSearchCode,
					testutil.CreateQueryParamMatcher(t, map[string]string{
						"q":        codeSearchQuery,
						"sort":     "indexed",
						"order":    "desc",
						"page":     "1",
						"per_page": "30",
					}).AndThen(
						testutil.MockResponse(t, http.StatusOK, mockSearchResult),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"q":       codeSearchQuery,
				"sort":    "indexed",
				"order":   "desc",
				"page":    float64(1),
				"perPage": float64(30),
			},
			expectError:    false,
			expectedResult: mockSearchResult,
		},
		{
			name: "code search with minimal parameters",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetSearchCode,
					testutil.CreateQueryParamMatcher(t, map[string]string{
						"q":        codeSearchQuery,
						"page":     "1",
						"per_page": "30",
					}).AndThen(
						testutil.MockResponse(t, http.StatusOK, mockSearchResult),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"q": codeSearchQuery,
			},
			expectError:    false,
			expectedResult: mockSearchResult,
		},
		{
			name: "search code fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetSearchCode,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusBadRequest)
						_, _ = w.Write([]byte(`{"message": "Validation Failed"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"q": invalidQuery,
			},
			expectError:    true,
			expectedErrMsg: "failed to search code",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			clientFn := func(ctx context.Context) (*github.Client, error) { return client, nil }
			_, handler := ghmcp.SearchCode(clientFn, testutil.NullTranslationHelperFunc)

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
			text := testutil.GetTextResult(t, result)

			// Unmarshal and verify the result
			var returnedResult github.CodeSearchResult
			err = json.Unmarshal([]byte(text), &returnedResult)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedResult.Total, *returnedResult.Total)
			assert.Equal(t, *tc.expectedResult.IncompleteResults, *returnedResult.IncompleteResults)
			assert.Len(t, returnedResult.CodeResults, len(tc.expectedResult.CodeResults))
			for i, code := range returnedResult.CodeResults {
				assert.Equal(t, *tc.expectedResult.CodeResults[i].Name, *code.Name)
				assert.Equal(t, *tc.expectedResult.CodeResults[i].Path, *code.Path)
				assert.Equal(t, *tc.expectedResult.CodeResults[i].SHA, *code.SHA)
				assert.Equal(t, *tc.expectedResult.CodeResults[i].HTMLURL, *code.HTMLURL)
				assert.Equal(t, *tc.expectedResult.CodeResults[i].Repository.FullName, *code.Repository.FullName)
			}
		})
	}
}

type userTestCase struct {
	name           string
	mockedClient   *http.Client
	requestArgs    map[string]interface{}
	expectError    bool
	expectedResult *github.UsersSearchResult
	expectedErrMsg string
}

func TestSearchUsers(t *testing.T) {
	// Use a default *github.Client for metadata assertions
	defaultClient := github.NewClient(nil)
	clientFn := func(ctx context.Context) (*github.Client, error) { return defaultClient, nil }
	tool, _ := ghmcp.SearchUsers(clientFn, testutil.NullTranslationHelperFunc)

	assert.Equal(t, "search_users", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "q")
	assert.Contains(t, tool.InputSchema.Properties, "sort")
	assert.Contains(t, tool.InputSchema.Properties, "order")
	assert.Contains(t, tool.InputSchema.Properties, "perPage")
	assert.Contains(t, tool.InputSchema.Properties, "page")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"q"})

	// Setup mock search results
	mockSearchResult := &github.UsersSearchResult{
		Total:             testutil.Ptr(2),
		IncompleteResults: testutil.Ptr(false),
		Users: []*github.User{
			{
				Login:     testutil.Ptr("user1"),
				ID:        testutil.Ptr(int64(1001)),
				HTMLURL:   testutil.Ptr("https://github.com/user1"),
				AvatarURL: testutil.Ptr("https://avatars.githubusercontent.com/u/1001"),
				Type:      testutil.Ptr("User"),
				Followers: testutil.Ptr(100),
				Following: testutil.Ptr(50),
			},
			{
				Login:     testutil.Ptr("user2"),
				ID:        testutil.Ptr(int64(1002)),
				HTMLURL:   testutil.Ptr("https://github.com/user2"),
				AvatarURL: testutil.Ptr("https://avatars.githubusercontent.com/u/1002"),
				Type:      testutil.Ptr("User"),
				Followers: testutil.Ptr(200),
				Following: testutil.Ptr(75),
			},
		},
	}

	tests := []userTestCase{
		{
			name: "successful users search with all parameters",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetSearchUsers,
					testutil.CreateQueryParamMatcher(t, map[string]string{
						"q":        userSearchQuery,
						"sort":     "followers",
						"order":    "desc",
						"page":     "1",
						"per_page": "30",
					}).AndThen(
						testutil.MockResponse(t, http.StatusOK, mockSearchResult),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"q":       userSearchQuery,
				"sort":    "followers",
				"order":   "desc",
				"page":    float64(1),
				"perPage": float64(30),
			},
			expectError:    false,
			expectedResult: mockSearchResult,
		},
		{
			name: "users search with minimal parameters",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetSearchUsers,
					testutil.CreateQueryParamMatcher(t, map[string]string{
						"q":        userSearchQuery,
						"page":     "1",
						"per_page": "30",
					}).AndThen(
						testutil.MockResponse(t, http.StatusOK, mockSearchResult),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"q": userSearchQuery,
			},
			expectError:    false,
			expectedResult: mockSearchResult,
		},
		{
			name: "search users fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetSearchUsers,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusBadRequest)
						_, _ = w.Write([]byte(`{"message": "Validation Failed"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"q": invalidQuery,
			},
			expectError:    true,
			expectedErrMsg: "failed to search users",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			clientFn := func(ctx context.Context) (*github.Client, error) { return client, nil }
			_, handler := ghmcp.SearchUsers(clientFn, testutil.NullTranslationHelperFunc)

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
			require.NotNil(t, result)

			text := testutil.GetTextResult(t, result)

			// Unmarshal and verify the result
			var returnedResult github.UsersSearchResult
			err = json.Unmarshal([]byte(text), &returnedResult)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedResult.Total, *returnedResult.Total)
			assert.Equal(t, *tc.expectedResult.IncompleteResults, *returnedResult.IncompleteResults)
			assert.Len(t, returnedResult.Users, len(tc.expectedResult.Users))
			for i, user := range returnedResult.Users {
				assert.Equal(t, *tc.expectedResult.Users[i].Login, *user.Login)
				assert.Equal(t, *tc.expectedResult.Users[i].ID, *user.ID)
				assert.Equal(t, *tc.expectedResult.Users[i].HTMLURL, *user.HTMLURL)
				assert.Equal(t, *tc.expectedResult.Users[i].AvatarURL, *user.AvatarURL)
				assert.Equal(t, *tc.expectedResult.Users[i].Type, *user.Type)
				assert.Equal(t, *tc.expectedResult.Users[i].Followers, *user.Followers)
			}
		})
	}
}

// Canonical test file for search.go
// All tests must directly and robustly test the canonical logic in search.go
// Remove all legacy, duplicate, or non-canonical tests
// Reference only helpers from /pkg/common and /pkg/testutil
// No import cycles, duplicate imports, or undefined helpers
// All test cases must match the actual signatures and logic of search.go
