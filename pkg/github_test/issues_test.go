// Canonical test file for issues.go
// All tests must directly and robustly test the canonical logic in issues.go
// Remove all legacy, duplicate, or non-canonical tests
// Reference only helpers from /pkg/common and /pkg/testutil
// No import cycles, duplicate imports, or undefined helpers
// All test cases must match the actual signatures and logic of issues.go
package ghmcp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v71/github"
	ghmcp "github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/github"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/testutil"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Define constants for repeated string literals to resolve duplication warnings
const (
	urlIssue123             = "https://github.com/owner/repo/issues/123"
	helpWanted              = "help wanted"
	minimalIssue            = "Minimal Issue"
	updatedIssueTitle       = "Updated Issue Title"
	updatedIssueDescription = "Updated issue description"
	onlyTitleUpdated        = "Only Title Updated"
	testIssueTitle          = "Test Issue"
	testIssueBody           = "This is a test issue"
	repoOpenQuery           = "repo:owner/repo is:issue is:open"
)

var mockIssue = &github.Issue{
	Number:    testutil.Ptr(123),
	Title:     testutil.Ptr(testIssueTitle),
	Body:      testutil.Ptr(testIssueBody),
	State:     testutil.Ptr("open"),
	HTMLURL:   testutil.Ptr(urlIssue123),
	Assignees: []*github.User{{Login: testutil.Ptr("user1")}, {Login: testutil.Ptr("user2")}},
	Labels:    []*github.Label{{Name: testutil.Ptr("bug")}, {Name: testutil.Ptr(helpWanted)}},
	Milestone: &github.Milestone{Number: testutil.Ptr(5)},
}

func TestGetIssue(t *testing.T) {
	// Verify tool definition once
	mockHTTP := mock.NewMockedHTTPClient()
	translateFn := func(key, defaultValue string) string { return key }
	clientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(mockHTTP), nil }
	tool, _ := ghmcp.GetIssue(clientFn, translateFn)

	assert.Equal(t, "get_issue", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "issue_number")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "issue_number"})

	// Setup mock issue for success case
	mockIssue := &github.Issue{
		Number:  testutil.Ptr(42),
		Title:   testutil.Ptr(testIssueTitle),
		Body:    testutil.Ptr(testIssueBody),
		State:   testutil.Ptr("open"),
		HTMLURL: testutil.Ptr("https://github.com/owner/repo/issues/42"),
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedIssue  *github.Issue
		expectedErrMsg string
	}{
		{
			name: "successful issue retrieval",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposIssuesByOwnerByRepoByIssueNumber,
					mockIssue,
				),
			),
			requestArgs: map[string]interface{}{
				"owner":        "owner",
				"repo":         "repo",
				"issue_number": float64(42),
			},
			expectError:   false,
			expectedIssue: mockIssue,
		},
		{
			name: "issue not found",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposIssuesByOwnerByRepoByIssueNumber,
					testutil.MockResponse(t, http.StatusNotFound, `{"message": "Issue not found"}`),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":        "owner",
				"repo":         "repo",
				"issue_number": float64(999),
			},
			expectError:    true,
			expectedErrMsg: "failed to get issue",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			clientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(tc.mockedClient), nil }
			_, handler := ghmcp.GetIssue(clientFn, translateFn)

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
			textContent := testutil.GetTextResult(t, result) // Changed to testutil.GetTextResult

			// Unmarshal and verify the result
			var returnedIssue github.Issue
			err = json.Unmarshal([]byte(textContent), &returnedIssue)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedIssue.Number, *returnedIssue.Number)
			assert.Equal(t, *tc.expectedIssue.Title, *returnedIssue.Title)
			assert.Equal(t, *tc.expectedIssue.Body, *returnedIssue.Body)
		})
	}
}

func TestAddIssueComment(t *testing.T) {
	// Verify tool definition once
	defaultClient := github.NewClient(nil)
	translateFn := func(key, defaultValue string) string { return key }
	clientFn := func(ctx context.Context) (*github.Client, error) { return defaultClient, nil }
	tool, _ := ghmcp.AddIssueComment(clientFn, translateFn)

	assert.Equal(t, "add_issue_comment", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "issue_number")
	assert.Contains(t, tool.InputSchema.Properties, "body")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "issue_number", "body"})

	// Setup mock comment for success case
	mockComment := &github.IssueComment{
		ID:   testutil.Ptr(int64(123)),
		Body: testutil.Ptr("This is a test comment"),
		User: &github.User{
			Login: testutil.Ptr("testuser"),
		},
		HTMLURL: testutil.Ptr("https://github.com/owner/repo/issues/42#issuecomment-123"),
	}

	tests := []struct {
		name            string
		mockedClient    *http.Client
		requestArgs     map[string]interface{}
		expectError     bool
		expectedComment *github.IssueComment
		expectedErrMsg  string
	}{
		{
			name: "successful comment creation",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposIssuesCommentsByOwnerByRepoByIssueNumber,
					testutil.MockResponse(t, http.StatusCreated, mockComment),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":        "owner",
				"repo":         "repo",
				"issue_number": float64(42),
				"body":         "This is a test comment",
			},
			expectError:     false,
			expectedComment: mockComment,
		},
		{
			name: "comment creation fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposIssuesCommentsByOwnerByRepoByIssueNumber,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusUnprocessableEntity)
						_, _ = w.Write([]byte(`{"message": "Invalid request"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":        "owner",
				"repo":         "repo",
				"issue_number": float64(42),
				"body":         "",
			},
			expectError:    false,
			expectedErrMsg: "missing required parameter: body",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			clientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(tc.mockedClient), nil }
			_, handler := ghmcp.AddIssueComment(clientFn, func(key, defaultValue string) string { return key })

			// Create call request
			request := mcp.CallToolRequest{
				Params: struct {
					Name      string                 `json:"name"`
					Arguments map[string]interface{} `json:"arguments,omitempty"`
					Meta      *struct {
						ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
					} `json:"_meta,omitempty"`
				}{
					Arguments: tc.requestArgs,
				},
			}

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
				return
			}

			if tc.expectedErrMsg != "" {
				require.NotNil(t, result)
				textContent := testutil.GetTextResult(t, result) // Changed to testutil.GetTextResult
				assert.Contains(t, textContent, tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)

			// Parse the result and get the text content if no error
			textContent := testutil.GetTextResult(t, result) // Changed to testutil.GetTextResult

			// Unmarshal and verify the result
			var returnedComment github.IssueComment
			err = json.Unmarshal([]byte(textContent), &returnedComment)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedComment.ID, *returnedComment.ID)
			assert.Equal(t, *tc.expectedComment.Body, *returnedComment.Body)
			assert.Equal(t, *tc.expectedComment.User.Login, *returnedComment.User.Login)

		})
	}
}

func TestSearchIssues(t *testing.T) {
	// Verify tool definition once
	// mockClient := github.NewClient(nil) // Removed unused variable
	// Setup tool definition with a default client
	defaultClient := github.NewClient(nil)
	clientFn := func(ctx context.Context) (*github.Client, error) { return defaultClient, nil }
	tool, _ := ghmcp.SearchIssues(clientFn, func(key, defaultValue string) string { return key })

	assert.Equal(t, "search_issues", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "q")
	assert.Contains(t, tool.InputSchema.Properties, "sort")
	assert.Contains(t, tool.InputSchema.Properties, "order")
	assert.Contains(t, tool.InputSchema.Properties, "perPage")
	assert.Contains(t, tool.InputSchema.Properties, "page")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"q"})

	// Setup mock search results
	mockSearchResult := &github.IssuesSearchResult{
		Total:             testutil.Ptr(2),
		IncompleteResults: testutil.Ptr(false),
		Issues: []*github.Issue{
			{
				Number:   testutil.Ptr(42),
				Title:    testutil.Ptr("Bug: Something is broken"),
				Body:     testutil.Ptr("This is a bug report"),
				State:    testutil.Ptr("open"),
				HTMLURL:  testutil.Ptr("https://github.com/owner/repo/issues/42"),
				Comments: testutil.Ptr(5),
				User: &github.User{
					Login: testutil.Ptr("user1"),
				},
			},
			{
				Number:   testutil.Ptr(43),
				Title:    testutil.Ptr("Feature: Add new functionality"),
				Body:     testutil.Ptr("This is a feature request"),
				State:    testutil.Ptr("open"),
				HTMLURL:  testutil.Ptr("https://github.com/owner/repo/issues/43"),
				Comments: testutil.Ptr(3),
				User: &github.User{
					Login: testutil.Ptr("user2"),
				},
			},
		},
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedResult *github.IssuesSearchResult
		expectedErrMsg string
	}{
		{
			name: "successful issues search with all parameters",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetSearchIssues,
					testutil.CreateQueryParamMatcher(
						t,
						map[string]string{
							"q":        repoOpenQuery,
							"sort":     "created",
							"order":    "desc",
							"page":     "1",
							"per_page": "30",
						},
					).AndThen(
						testutil.MockResponse(t, http.StatusOK, mockSearchResult),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"q":       repoOpenQuery,
				"sort":    "created",
				"order":   "desc",
				"page":    float64(1),
				"perPage": float64(30),
			},
			expectError:    false,
			expectedResult: mockSearchResult,
		},
		{
			name: "issues search with minimal parameters",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetSearchIssues,
					mockSearchResult,
				),
			),
			requestArgs: map[string]interface{}{
				"q": repoOpenQuery,
			},
			expectError:    false,
			expectedResult: mockSearchResult,
		},
		{
			name: "search issues fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetSearchIssues,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusBadRequest)
						_, _ = w.Write([]byte(`{"message": "Validation Failed"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"q": "invalid:query",
			},
			expectError:    true,
			expectedErrMsg: "failed to search issues",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			// client := github.NewClient(tc.mockedClient) // Removed unused variable
			// Use the clientFn defined for the test case scope, which uses tc.mockedClient
			testClientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(tc.mockedClient), nil }
			_, handler := ghmcp.SearchIssues(testClientFn, func(key, defaultValue string) string { return key })

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
			textContent := testutil.GetTextResult(t, result) // Changed to testutil.GetTextResult

			// Unmarshal and verify the result
			var returnedResult github.IssuesSearchResult
			err = json.Unmarshal([]byte(textContent), &returnedResult)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedResult.Total, *returnedResult.Total)
			assert.Equal(t, *tc.expectedResult.IncompleteResults, *returnedResult.IncompleteResults)
			assert.Len(t, returnedResult.Issues, len(tc.expectedResult.Issues))
			for i, issue := range returnedResult.Issues {
				assert.Equal(t, *tc.expectedResult.Issues[i].Number, *issue.Number)
				assert.Equal(t, *tc.expectedResult.Issues[i].Title, *issue.Title)
				assert.Equal(t, *tc.expectedResult.Issues[i].State, *issue.State)
				assert.Equal(t, *tc.expectedResult.Issues[i].HTMLURL, *issue.HTMLURL)
				assert.Equal(t, *tc.expectedResult.Issues[i].User.Login, *issue.User.Login)
			}
		})
	}
}

func assertIssueResult(t *testing.T, expected, actual *github.Issue) {
	require.NotNil(t, actual)
	assert.Equal(t, *expected.Number, *actual.Number)
	assert.Equal(t, *expected.Title, *actual.Title)
	assert.Equal(t, *expected.State, *actual.State)
	assert.Equal(t, *expected.HTMLURL, *actual.HTMLURL)
	if expected.Body != nil {
		assert.Equal(t, *expected.Body, *actual.Body)
	}
	if len(expected.Assignees) > 0 {
		assert.Equal(t, len(expected.Assignees), len(actual.Assignees))
		for i, assignee := range actual.Assignees {
			assert.Equal(t, *expected.Assignees[i].Login, *assignee.Login)
		}
	}
	if len(expected.Labels) > 0 {
		assert.Equal(t, len(expected.Labels), len(actual.Labels))
		for i, label := range actual.Labels {
			assert.Equal(t, *expected.Labels[i].Name, *label.Name)
		}
	}
	if expected.Milestone != nil {
		assert.NotNil(t, actual.Milestone)
		assert.Equal(t, *expected.Milestone.Number, *actual.Milestone.Number)
	}
}

func TestCreateIssue(t *testing.T) {
	// Verify tool definition once
	toolDefinitionClientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(nil), nil }
	tool, _ := ghmcp.CreateIssue(toolDefinitionClientFn, func(key, defaultValue string) string { return key })

	assert.Equal(t, "create_issue", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "title")
	assert.Contains(t, tool.InputSchema.Properties, "body")
	assert.Contains(t, tool.InputSchema.Properties, "assignees")
	assert.Contains(t, tool.InputSchema.Properties, "labels")
	assert.Contains(t, tool.InputSchema.Properties, "milestone")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "title"})

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedIssue  *github.Issue
		expectedErrMsg string
	}{
		{
			name: "successful issue creation with all fields",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposIssuesByOwnerByRepo,
					testutil.MockResponse(t, http.StatusOK, map[string]any{
						"title":     testutil.Ptr(testIssueTitle),
						"body":      testutil.Ptr(testIssueBody),
						"labels":    []any{"bug", helpWanted},
						"assignees": []any{"user1", "user2"},
						"milestone": float64(5),
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":     "owner",
				"repo":      "repo",
				"title":     testutil.Ptr(testIssueTitle),
				"body":      testutil.Ptr(testIssueBody),
				"assignees": []any{"user1", "user2"},
				"labels":    []any{"bug", helpWanted},
				"milestone": float64(5),
			},
			expectError:   false,
			expectedIssue: mockIssue,
		},
		{
			name: "successful issue creation with minimal fields",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposIssuesByOwnerByRepo,
					testutil.MockResponse(t, http.StatusCreated, &github.Issue{
						Number:  testutil.Ptr(124),
						Title:   testutil.Ptr(minimalIssue),
						HTMLURL: testutil.Ptr("https://github.com/owner/repo/issues/124"),
						State:   testutil.Ptr("open"),
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":     "owner",
				"repo":      "repo",
				"title":     testutil.Ptr(minimalIssue),
				"assignees": nil, // Expect no failure with nil optional value.
			},
			expectError: false,
			expectedIssue: &github.Issue{
				Number:  testutil.Ptr(124),
				Title:   testutil.Ptr(minimalIssue),
				HTMLURL: testutil.Ptr("https://github.com/owner/repo/issues/124"),
				State:   testutil.Ptr("open"),
			},
		},
		{
			name: "issue creation fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposIssuesByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusUnprocessableEntity)
						_, _ = w.Write([]byte(`{"message": "Validation failed"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner": "owner",
				"repo":  "repo",
				"title": testutil.Ptr(""),
			},
			expectError:    false,
			expectedErrMsg: "missing required parameter: title",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock for test execution
			clientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(tc.mockedClient), nil }
			_, handler := ghmcp.CreateIssue(clientFn, func(key, defaultValue string) string { return key })

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

			if tc.expectedErrMsg != "" {
				require.NotNil(t, result)
				textContent := testutil.GetTextResult(t, result) // Changed to testutil.GetTextResult
				assert.Contains(t, textContent, tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			textContent := testutil.GetTextResult(t, result) // Changed to testutil.GetTextResult

			// Unmarshal and verify the result
			var returnedIssue github.Issue
			err = json.Unmarshal([]byte(textContent), &returnedIssue)
			require.NoError(t, err)
			assertIssueResult(t, tc.expectedIssue, &returnedIssue)
		})
	}
}

func TestListIssues(t *testing.T) {
	// Verify tool definition
	toolDefinitionClientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(nil), nil }
	tool, _ := ghmcp.ListIssues(toolDefinitionClientFn, func(key, defaultValue string) string { return key })

	assert.Equal(t, "list_issues", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "state")
	assert.Contains(t, tool.InputSchema.Properties, "labels")
	assert.Contains(t, tool.InputSchema.Properties, "sort")
	assert.Contains(t, tool.InputSchema.Properties, "direction")
	assert.Contains(t, tool.InputSchema.Properties, "since")
	assert.Contains(t, tool.InputSchema.Properties, "page")
	assert.Contains(t, tool.InputSchema.Properties, "perPage")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo"})

	// Setup mock issues for success case
	mockIssues := []*github.Issue{
		{
			Number:    testutil.Ptr(123),
			Title:     testutil.Ptr("First Issue"),
			Body:      testutil.Ptr("This is the first test issue"),
			State:     testutil.Ptr("open"),
			HTMLURL:   testutil.Ptr("https://github.com/owner/repo/issues/123"),
			CreatedAt: testutil.Ptr(github.Timestamp{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)}),
		},
		{
			Number:    testutil.Ptr(456),
			Title:     testutil.Ptr("Second Issue"),
			Body:      testutil.Ptr("This is the second test issue"),
			State:     testutil.Ptr("open"),
			HTMLURL:   testutil.Ptr("https://github.com/owner/repo/issues/456"),
			Labels:    []*github.Label{{Name: testutil.Ptr("bug")}},
			CreatedAt: testutil.Ptr(github.Timestamp{Time: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)}),
		},
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedIssues []*github.Issue
		expectedErrMsg string
	}{
		{
			name: "list issues with minimal parameters",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposIssuesByOwnerByRepo,
					mockIssues,
				),
			),
			requestArgs: map[string]interface{}{
				"owner": "owner",
				"repo":  "repo",
			},
			expectError:    false,
			expectedIssues: mockIssues,
		},
		{
			name: "list issues with all parameters",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposIssuesByOwnerByRepo,
					testutil.CreateQueryParamMatcher(t, map[string]string{
						"state":     "open",
						"labels":    "bug,enhancement",
						"sort":      "created",
						"direction": "desc",
						"since":     "2023-01-01T00:00:00Z",
						"page":      "1",
						"per_page":  "30",
					}).AndThen(
						testutil.MockResponse(t, http.StatusOK, mockIssues),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":     "owner",
				"repo":      "repo",
				"state":     "open",
				"labels":    []any{"bug", "enhancement"},
				"sort":      "created",
				"direction": "desc",
				"since":     "2023-01-01T00:00:00Z",
				"page":      float64(1),
				"perPage":   float64(30),
			},
			expectError:    false,
			expectedIssues: mockIssues,
		},
		{
			name: "invalid since parameter",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposIssuesByOwnerByRepo,
					mockIssues,
				),
			),
			requestArgs: map[string]interface{}{
				"owner": "owner",
				"repo":  "repo",
				"since": "invalid-date",
			},
			expectError:    true,
			expectedErrMsg: "invalid ISO 8601 timestamp",
		},
		{
			name: "list issues fails with error",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposIssuesByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Repository not found"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner": "nonexistent",
				"repo":  "repo",
			},
			expectError:    true,
			expectedErrMsg: "failed to list issues",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock for test execution
			clientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(tc.mockedClient), nil }
			_, handler := ghmcp.ListIssues(clientFn, func(key, defaultValue string) string { return key })

			// Create call request
			request := testutil.CreateMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				if err != nil {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
				} else {
					// For errors returned as part of the result, not as an error
					assert.NotNil(t, result)
					textContent := testutil.GetTextResult(t, result) // Changed to testutil.GetTextResult
					assert.Contains(t, textContent, tc.expectedErrMsg)
				}
				return
			}

			require.NoError(t, err)
			textContent := testutil.GetTextResult(t, result) // Changed to testutil.GetTextResult

			// Unmarshal and verify the result
			var returnedIssues []*github.Issue
			err = json.Unmarshal([]byte(textContent), &returnedIssues)
			require.NoError(t, err)

			assert.Len(t, returnedIssues, len(tc.expectedIssues))
			for i, issue := range returnedIssues {
				assert.Equal(t, *tc.expectedIssues[i].Number, *issue.Number)
				assert.Equal(t, *tc.expectedIssues[i].Title, *issue.Title)
				assert.Equal(t, *tc.expectedIssues[i].State, *issue.State)
				assert.Equal(t, *tc.expectedIssues[i].HTMLURL, *issue.HTMLURL)
			}
		})
	}
}

func TestUpdateIssue(t *testing.T) {
	// Verify tool definition once
	toolDefinitionClientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(nil), nil }
	tool, _ := ghmcp.UpdateIssue(toolDefinitionClientFn, func(key, defaultValue string) string { return key })

	assert.Equal(t, "update_issue", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "issue_number")
	assert.NotContains(t, tool.InputSchema.Required, "title") // Title is optional for updates

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedIssue  *github.Issue
		expectedErrMsg string
	}{
		{
			name: "update issue with all fields",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PatchReposIssuesByOwnerByRepoByIssueNumber,
					testutil.MockResponse(t, http.StatusOK, map[string]any{
						"title":     testutil.Ptr(updatedIssueTitle),
						"body":      testutil.Ptr(updatedIssueDescription),
						"state":     testutil.Ptr("closed"),
						"labels":    []any{"bug", "priority"},
						"assignees": []any{"assignee1", "assignee2"},
						"milestone": float64(5),
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":        "owner",
				"repo":         "repo",
				"issue_number": float64(123),
				"title":        testutil.Ptr(updatedIssueTitle),
				"body":         testutil.Ptr(updatedIssueDescription),
				"state":        testutil.Ptr("closed"),
				"labels":       []any{"bug", "priority"},
				"assignees":    []any{"assignee1", "assignee2"},
				"milestone":    float64(5),
			},
			expectError:   false,
			expectedIssue: mockIssue,
		},
		{
			name: "update issue with minimal fields",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PatchReposIssuesByOwnerByRepoByIssueNumber,
					testutil.MockResponse(t, http.StatusOK, &github.Issue{
						Number:  testutil.Ptr(123),
						Title:   testutil.Ptr(onlyTitleUpdated),
						HTMLURL: testutil.Ptr(urlIssue123),
						State:   testutil.Ptr("open"),
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":        "owner",
				"repo":         "repo",
				"issue_number": float64(123),
				"title":        testutil.Ptr(onlyTitleUpdated),
			},
			expectError: false,
			expectedIssue: &github.Issue{
				Number:  testutil.Ptr(123),
				Title:   testutil.Ptr(onlyTitleUpdated),
				HTMLURL: testutil.Ptr(urlIssue123),
				State:   testutil.Ptr("open"),
			},
		},
		{
			name: "update issue fails with not found",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PatchReposIssuesByOwnerByRepoByIssueNumber,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Issue not found"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":        "owner",
				"repo":         "repo",
				"issue_number": float64(999),
				"title":        testutil.Ptr("This issue doesn't exist"),
			},
			expectError:    true,
			expectedErrMsg: "failed to update issue",
		},
		{
			name: "update issue fails with validation error",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PatchReposIssuesByOwnerByRepoByIssueNumber,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusUnprocessableEntity)
						_, _ = w.Write([]byte(`{"message": "Invalid state value"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":        "owner",
				"repo":         "repo",
				"issue_number": float64(123),
				"state":        testutil.Ptr("invalid_state"),
			},
			expectError:    true,
			expectedErrMsg: "failed to update issue",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock for test execution
			clientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(tc.mockedClient), nil }
			_, handler := ghmcp.UpdateIssue(clientFn, func(key, defaultValue string) string { return key })
			request := testutil.CreateMCPRequest(tc.requestArgs)
			result, err := handler(context.Background(), request)

			if tc.expectError {
				if err != nil {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
				} else {
					require.NotNil(t, result)
					textContent := testutil.GetTextResult(t, result) // Changed to testutil.GetTextResult
					assert.Contains(t, textContent, tc.expectedErrMsg)
				}
				return
			}

			require.NoError(t, err)
			textContent := testutil.GetTextResult(t, result) // Changed to testutil.GetTextResult
			var returnedIssue github.Issue
			err = json.Unmarshal([]byte(textContent), &returnedIssue)
			require.NoError(t, err)
			assertIssueResult(t, tc.expectedIssue, &returnedIssue)
		})
	}
}

func TestGetIssueComments(t *testing.T) {
	// Verify tool definition once
	toolDefinitionClientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(nil), nil }
	tool, _ := ghmcp.GetIssueComments(toolDefinitionClientFn, func(key, defaultValue string) string { return key })

	assert.Equal(t, "get_issue_comments", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "issue_number")
	assert.Contains(t, tool.InputSchema.Properties, "page")
	assert.Contains(t, tool.InputSchema.Properties, "per_page")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "issue_number"})

	// Setup mock comments for success case
	mockComments := []*github.IssueComment{
		{
			ID:   testutil.Ptr(int64(123)),
			Body: testutil.Ptr("This is the first comment"),
			User: &github.User{
				Login: testutil.Ptr("user1"),
			},
			CreatedAt: testutil.Ptr(github.Timestamp{Time: time.Now().Add(-time.Hour * 24)}),
		},
		{
			ID:   testutil.Ptr(int64(456)),
			Body: testutil.Ptr("This is the second comment"),
			User: &github.User{
				Login: testutil.Ptr("user2"),
			},
			CreatedAt: testutil.Ptr(github.Timestamp{Time: time.Now().Add(-time.Hour)}),
		},
	}

	tests := []struct {
		name             string
		mockedClient     *http.Client
		requestArgs      map[string]interface{}
		expectError      bool
		expectedComments []*github.IssueComment
		expectedErrMsg   string
	}{
		{
			name: "successful comments retrieval",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber,
					mockComments,
				),
			),
			requestArgs: map[string]interface{}{
				"owner":        "owner",
				"repo":         "repo",
				"issue_number": float64(42),
			},
			expectError:      false,
			expectedComments: mockComments,
		},
		{
			name: "successful comments retrieval with pagination",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber,
					testutil.CreateQueryParamMatcher(t, map[string]string{
						"page":     "2",
						"per_page": "10",
					}).AndThen(
						testutil.MockResponse(t, http.StatusOK, mockComments),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":        "owner",
				"repo":         "repo",
				"issue_number": float64(42),
				"page":         float64(2),
				"per_page":     float64(10),
			},
			expectError:      false,
			expectedComments: mockComments,
		},
		{
			name: "issue not found",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposIssuesCommentsByOwnerByRepoByIssueNumber,
					testutil.MockResponse(t, http.StatusNotFound, `{"message": "Issue not found"}`),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":        "owner",
				"repo":         "repo",
				"issue_number": float64(999),
			},
			expectError:    true,
			expectedErrMsg: "failed to get issue comments",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock for test execution
			clientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(tc.mockedClient), nil }
			_, handler := ghmcp.GetIssueComments(clientFn, func(key, defaultValue string) string { return key })
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
			textContent := testutil.GetTextResult(t, result) // Changed to testutil.GetTextResult

			// Unmarshal and verify the result
			var returnedComments []*github.IssueComment
			err = json.Unmarshal([]byte(textContent), &returnedComments)
			require.NoError(t, err)
			assert.Equal(t, len(tc.expectedComments), len(returnedComments))
			if len(returnedComments) > 0 {
				assert.Equal(t, *tc.expectedComments[0].Body, *returnedComments[0].Body)
				assert.Equal(t, *tc.expectedComments[0].User.Login, *returnedComments[0].User.Login)
			}
		})
	}
}
