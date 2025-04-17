package github

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v69/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetIssue(t *testing.T) {
	// Verify tool definition once
	mockClient := githubMCP.NewClient(nil)
	tool, _ := GetIssue(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	assert.Equal(t, "get_issue", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "issue_number")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "issue_number"})

	// Setup mock issue for success case
	mockIssue := &github.Issue{
		Number:  githubMCP.Ptr(42),
		Title:   githubMCP.Ptr("Test Issue"),
		Body:    githubMCP.Ptr("This is a test issue"),
		State:   githubMCP.Ptr("open"),
		HTMLURL: githubMCP.Ptr("https://github.com/owner/repo/issues/42"),
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
					mockResponse(t, http.StatusNotFound, `{"message": "Issue not found"}`),
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
			client := githubMCP.NewClient(tc.mockedClient)
			_, handler := GetIssue(stubGetClientFn(client), translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			textContent := getTextResult(t, result)

			// Unmarshal and verify the result
			var returnedIssue github.Issue
			err = json.Unmarshal([]byte(textContent.Text), &returnedIssue)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedIssue.Number, *returnedIssue.Number)
			assert.Equal(t, *tc.expectedIssue.Title, *returnedIssue.Title)
			assert.Equal(t, *tc.expectedIssue.Body, *returnedIssue.Body)
		})
	}
}

func Test_AddIssueComment(t *testing.T) {
	// Verify tool definition once
	mockClient := githubMCP.NewClient(nil)
	tool, _ := AddIssueComment(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	assert.Equal(t, "add_issue_comment", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "issue_number")
	assert.Contains(t, tool.InputSchema.Properties, "body")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "issue_number", "body"})

	// Setup mock comment for success case
	mockComment := &github.IssueComment{
		ID:   githubMCP.Ptr(int64(123)),
		Body: githubMCP.Ptr("This is a test comment"),
		User: &github.User{
			Login: githubMCP.Ptr("testuser"),
		},
		HTMLURL: githubMCP.Ptr("https://github.com/owner/repo/issues/42#issuecomment-123"),
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
					mockResponse(t, http.StatusCreated, mockComment),
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
			client := githubMCP.NewClient(tc.mockedClient)
			_, handler := AddIssueComment(stubGetClientFn(client), translations.NullTranslationHelper)

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
				textContent := getTextResult(t, result)
				assert.Contains(t, textContent.Text, tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)

			// Parse the result and get the text content if no error
			textContent := getTextResult(t, result)

			// Unmarshal and verify the result
			var returnedComment github.IssueComment
			err = json.Unmarshal([]byte(textContent.Text), &returnedComment)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedComment.ID, *returnedComment.ID)
			assert.Equal(t, *tc.expectedComment.Body, *returnedComment.Body)
			assert.Equal(t, *tc.expectedComment.User.Login, *returnedComment.User.Login)

		})
	}
}

func Test_SearchIssues(t *testing.T) {
	// Verify tool definition once
	mockClient := githubMCP.NewClient(nil)
	tool, _ := SearchIssues(stubGetClientFn(mockClient), translations.NullTranslationHelper)

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
		Total:             githubMCP.Ptr(2),
		IncompleteResults: githubMCP.Ptr(false),
		Issues: []*github.Issue{
			{
				Number:   githubMCP.Ptr(42),
				Title:    githubMCP.Ptr("Bug: Something is broken"),
				Body:     githubMCP.Ptr("This is a bug report"),
				State:    githubMCP.Ptr("open"),
				HTMLURL:  githubMCP.Ptr("https://github.com/owner/repo/issues/42"),
				Comments: githubMCP.Ptr(5),
				User: &github.User{
					Login: githubMCP.Ptr("user1"),
				},
			},
			{
				Number:   githubMCP.Ptr(43),
				Title:    githubMCP.Ptr("Feature: Add new functionality"),
				Body:     githubMCP.Ptr("This is a feature request"),
				State:    githubMCP.Ptr("open"),
				HTMLURL:  githubMCP.Ptr("https://github.com/owner/repo/issues/43"),
				Comments: githubMCP.Ptr(3),
				User: &github.User{
					Login: githubMCP.Ptr("user2"),
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
					expectQueryParams(
						t,
						map[string]string{
							"q":        "repo:owner/repo is:issue is:open",
							"sort":     "created",
							"order":    "desc",
							"page":     "1",
							"per_page": "30",
						},
					).andThen(
						mockResponse(t, http.StatusOK, mockSearchResult),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"q":       "repo:owner/repo is:issue is:open",
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
				"q": "repo:owner/repo is:issue is:open",
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
			client := githubMCP.NewClient(tc.mockedClient)
			_, handler := SearchIssues(stubGetClientFn(client), translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

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
			textContent := getTextResult(t, result)

			// Unmarshal and verify the result
			var returnedResult github.IssuesSearchResult
			err = json.Unmarshal([]byte(textContent.Text), &returnedResult)
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

func Test_CreateIssue(t *testing.T) {
	// Verify tool definition once
	mockClient := githubMCP.NewClient(nil)
	tool, _ := CreateIssue(stubGetClientFn(mockClient), translations.NullTranslationHelper)

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

	// Setup mock issue for success case
	mockIssue := &github.Issue{
		Number:    githubMCP.Ptr(123),
		Title:     githubMCP.Ptr("Test Issue"),
		Body:      githubMCP.Ptr("This is a test issue"),
		State:     githubMCP.Ptr("open"),
		HTMLURL:   githubMCP.Ptr("https://github.com/owner/repo/issues/123"),
		Assignees: []*github.User{{Login: githubMCP.Ptr("user1")}, {Login: githubMCP.Ptr("user2")}},
		Labels:    []*github.Label{{Name: githubMCP.Ptr("bug")}, {Name: githubMCP.Ptr("help wanted")}},
		Milestone: &github.Milestone{Number: githubMCP.Ptr(5)},
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
			name: "successful issue creation with all fields",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposIssuesByOwnerByRepo,
					expectRequestBody(t, map[string]any{
						"title":     "Test Issue",
						"body":      "This is a test issue",
						"labels":    []any{"bug", "help wanted"},
						"assignees": []any{"user1", "user2"},
						"milestone": float64(5),
					}).andThen(
						mockResponse(t, http.StatusCreated, mockIssue),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":     "owner",
				"repo":      "repo",
				"title":     "Test Issue",
				"body":      "This is a test issue",
				"assignees": []any{"user1", "user2"},
				"labels":    []any{"bug", "help wanted"},
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
					mockResponse(t, http.StatusCreated, &github.Issue{
						Number:  githubMCP.Ptr(124),
						Title:   githubMCP.Ptr("Minimal Issue"),
						HTMLURL: githubMCP.Ptr("https://github.com/owner/repo/issues/124"),
						State:   githubMCP.Ptr("open"),
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":     "owner",
				"repo":      "repo",
				"title":     "Minimal Issue",
				"assignees": nil, // Expect no failure with nil optional value.
			},
			expectError: false,
			expectedIssue: &github.Issue{
				Number:  githubMCP.Ptr(124),
				Title:   githubMCP.Ptr("Minimal Issue"),
				HTMLURL: githubMCP.Ptr("https://github.com/owner/repo/issues/124"),
				State:   githubMCP.Ptr("open"),
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
				"title": "",
			},
			expectError:    false,
			expectedErrMsg: "missing required parameter: title",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := githubMCP.NewClient(tc.mockedClient)
			_, handler := CreateIssue(stubGetClientFn(client), translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

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
				textContent := getTextResult(t, result)
				assert.Contains(t, textContent.Text, tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			textContent := getTextResult(t, result)

			// Unmarshal and verify the result
			var returnedIssue github.Issue
			err = json.Unmarshal([]byte(textContent.Text), &returnedIssue)
			require.NoError(t, err)

			assert.Equal(t, *tc.expectedIssue.Number, *returnedIssue.Number)
			assert.Equal(t, *tc.expectedIssue.Title, *returnedIssue.Title)
			assert.Equal(t, *tc.expectedIssue.State, *returnedIssue.State)
			assert.Equal(t, *tc.expectedIssue.HTMLURL, *returnedIssue.HTMLURL)

			if tc.expectedIssue.Body != nil {
				assert.Equal(t, *tc.expectedIssue.Body, *returnedIssue.Body)
			}

			// Check assignees if expected
			if len(tc.expectedIssue.Assignees) > 0 {
				assert.Equal(t, len(tc.expectedIssue.Assignees), len(returnedIssue.Assignees))
				for i, assignee := range returnedIssue.Assignees {
					assert.Equal(t, *tc.expectedIssue.Assignees[i].Login, *assignee.Login)
				}
			}

			// Check labels if expected
			if len(tc.expectedIssue.Labels) > 0 {
				assert.Equal(t, len(tc.expectedIssue.Labels), len(returnedIssue.Labels))
				for i, label := range returnedIssue.Labels {
					assert.Equal(t, *tc.expectedIssue.Labels[i].Name, *label.Name)
				}
			}
		})
	}
}

func Test_ListIssues(t *testing.T) {
	// Verify tool definition
	mockClient := githubMCP.NewClient(nil)
	tool, _ := ListIssues(stubGetClientFn(mockClient), translations.NullTranslationHelper)

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
			Number:    githubMCP.Ptr(123),
			Title:     githubMCP.Ptr("First Issue"),
			Body:      githubMCP.Ptr("This is the first test issue"),
			State:     githubMCP.Ptr("open"),
			HTMLURL:   githubMCP.Ptr("https://github.com/owner/repo/issues/123"),
			CreatedAt: &github.Timestamp{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
		{
			Number:    githubMCP.Ptr(456),
			Title:     githubMCP.Ptr("Second Issue"),
			Body:      githubMCP.Ptr("This is the second test issue"),
			State:     githubMCP.Ptr("open"),
			HTMLURL:   githubMCP.Ptr("https://github.com/owner/repo/issues/456"),
			Labels:    []*github.Label{{Name: githubMCP.Ptr("bug")}},
			CreatedAt: &github.Timestamp{Time: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)},
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
					expectQueryParams(t, map[string]string{
						"state":     "open",
						"labels":    "bug,enhancement",
						"sort":      "created",
						"direction": "desc",
						"since":     "2023-01-01T00:00:00Z",
						"page":      "1",
						"per_page":  "30",
					}).andThen(
						mockResponse(t, http.StatusOK, mockIssues),
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
			// Setup client with mock
			client := githubMCP.NewClient(tc.mockedClient)
			_, handler := ListIssues(stubGetClientFn(client), translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				if err != nil {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
				} else {
					// For errors returned as part of the result, not as an error
					assert.NotNil(t, result)
					textContent := getTextResult(t, result)
					assert.Contains(t, textContent.Text, tc.expectedErrMsg)
				}
				return
			}

			require.NoError(t, err)

			// Parse the result and get the text content if no error
			textContent := getTextResult(t, result)

			// Unmarshal and verify the result
			var returnedIssues []*github.Issue
			err = json.Unmarshal([]byte(textContent.Text), &returnedIssues)
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

func Test_UpdateIssue(t *testing.T) {
	// Verify tool definition
	mockClient := githubMCP.NewClient(nil)
	tool, _ := UpdateIssue(stubGetClientFn(mockClient), translations.NullTranslationHelper)

	assert.Equal(t, "update_issue", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "issue_number")
	assert.Contains(t, tool.InputSchema.Properties, "title")
	assert.Contains(t, tool.InputSchema.Properties, "body")
	assert.Contains(t, tool.InputSchema.Properties, "state")
	assert.Contains(t, tool.InputSchema.Properties, "labels")
	assert.Contains(t, tool.InputSchema.Properties, "assignees")
	assert.Contains(t, tool.InputSchema.Properties, "milestone")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "issue_number"})

	// Setup mock issue for success case
	mockIssue := &github.Issue{
		Number:    githubMCP.Ptr(123),
		Title:     githubMCP.Ptr("Updated Issue Title"),
		Body:      githubMCP.Ptr("Updated issue description"),
		State:     githubMCP.Ptr("closed"),
		HTMLURL:   githubMCP.Ptr("https://github.com/owner/repo/issues/123"),
		Assignees: []*github.User{{Login: githubMCP.Ptr("assignee1")}, {Login: githubMCP.Ptr("assignee2")}},
		Labels:    []*github.Label{{Name: githubMCP.Ptr("bug")}, {Name: githubMCP.Ptr("priority")}},
		Milestone: &github.Milestone{Number: githubMCP.Ptr(5)},
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
			name: "update issue with all fields",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PatchReposIssuesByOwnerByRepoByIssueNumber,
					expectRequestBody(t, map[string]any{
						"title":     "Updated Issue Title",
						"body":      "Updated issue description",
						"state":     "closed",
						"labels":    []any{"bug", "priority"},
						"assignees": []any{"assignee1", "assignee2"},
						"milestone": float64(5),
					}).andThen(
						mockResponse(t, http.StatusOK, mockIssue),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":        "owner",
				"repo":         "repo",
				"issue_number": float64(123),
				"title":        "Updated Issue Title",
				"body":         "Updated issue description",
				"state":        "closed",
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
					mockResponse(t, http.StatusOK, &github.Issue{
						Number:  githubMCP.Ptr(123),
						Title:   githubMCP.Ptr("Only Title Updated"),
						HTMLURL: githubMCP.Ptr("https://github.com/owner/repo/issues/123"),
						State:   githubMCP.Ptr("open"),
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":        "owner",
				"repo":         "repo",
				"issue_number": float64(123),
				"title":        "Only Title Updated",
			},
			expectError: false,
			expectedIssue: &github.Issue{
				Number:  githubMCP.Ptr(123),
				Title:   githubMCP.Ptr("Only Title Updated"),
				HTMLURL: githubMCP.Ptr("https://github.com/owner/repo/issues/123"),
				State:   githubMCP.Ptr("open"),
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
				"title":        "This issue doesn't exist",
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
				"state":        "invalid_state",
			},
			expectError:    true,
			expectedErrMsg: "failed to update issue",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := githubMCP.NewClient(tc.mockedClient)
			_, handler := UpdateIssue(stubGetClientFn(client), translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				if err != nil {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
				} else {
					// For errors returned as part of the result, not as an error
					require.NotNil(t, result)
					textContent := getTextResult(t, result)
					assert.Contains(t, textContent.Text, tc.expectedErrMsg)
				}
				return
			}

			require.NoError(t, err)

			// Parse the result and get the text content if no error
			textContent := getTextResult(t, result)

			// Unmarshal and verify the result
			var returnedIssue github.Issue
			err = json.Unmarshal([]byte(textContent.Text), &returnedIssue)
			require.NoError(t, err)

			assert.Equal(t, *tc.expectedIssue.Number, *returnedIssue.Number)
			assert.Equal(t, *tc.expectedIssue.Title, *returnedIssue.Title)
			assert.Equal(t, *tc.expectedIssue.State, *returnedIssue.State)
			assert.Equal(t, *tc.expectedIssue.HTMLURL, *returnedIssue.HTMLURL)

			if tc.expectedIssue.Body != nil {
				assert.Equal(t, *tc.expectedIssue.Body, *returnedIssue.Body)
			}

			// Check assignees if expected
			if len(tc.expectedIssue.Assignees) > 0 {
				assert.Len(t, returnedIssue.Assignees, len(tc.expectedIssue.Assignees))
				for i, assignee := range returnedIssue.Assignees {
					assert.Equal(t, *tc.expectedIssue.Assignees[i].Login, *assignee.Login)
				}
			}

			// Check labels if expected
			if len(tc.expectedIssue.Labels) > 0 {
				assert.Len(t, returnedIssue.Labels, len(tc.expectedIssue.Labels))
				for i, label := range returnedIssue.Labels {
					assert.Equal(t, *tc.expectedIssue.Labels[i].Name, *label.Name)
				}
			}

			// Check milestone if expected
			if tc.expectedIssue.Milestone != nil {
				assert.NotNil(t, returnedIssue.Milestone)
				assert.Equal(t, *tc.expectedIssue.Milestone.Number, *returnedIssue.Milestone.Number)
			}
		})
	}
}

func Test_ParseISOTimestamp(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedErr  bool
		expectedTime time.Time
	}{
		{
			name:         "valid RFC3339 format",
			input:        "2023-01-15T14:30:00Z",
			expectedErr:  false,
			expectedTime: time.Date(2023, 1, 15, 14, 30, 0, 0, time.UTC),
		},
		{
			name:         "valid date only format",
			input:        "2023-01-15",
			expectedErr:  false,
			expectedTime: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:        "empty timestamp",
			input:       "",
			expectedErr: true,
		},
		{
			name:        "invalid format",
			input:       "15/01/2023",
			expectedErr: true,
		},
		{
			name:        "invalid date",
			input:       "2023-13-45",
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsedTime, err := parseISOTimestamp(tc.input)

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedTime, parsedTime)
			}
		})
	}
}

func Test_GetIssueComments(t *testing.T) {
	// Verify tool definition once
	mockClient := githubMCP.NewClient(nil)
	tool, _ := GetIssueComments(stubGetClientFn(mockClient), translations.NullTranslationHelper)

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
			ID:   githubMCP.Ptr(int64(123)),
			Body: githubMCP.Ptr("This is the first comment"),
			User: &github.User{
				Login: githubMCP.Ptr("user1"),
			},
			CreatedAt: &github.Timestamp{Time: time.Now().Add(-time.Hour * 24)},
		},
		{
			ID:   githubMCP.Ptr(int64(456)),
			Body: githubMCP.Ptr("This is the second comment"),
			User: &github.User{
				Login: githubMCP.Ptr("user2"),
			},
			CreatedAt: &github.Timestamp{Time: time.Now().Add(-time.Hour)},
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
					expectQueryParams(t, map[string]string{
						"page":     "2",
						"per_page": "10",
					}).andThen(
						mockResponse(t, http.StatusOK, mockComments),
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
					mockResponse(t, http.StatusNotFound, `{"message": "Issue not found"}`),
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
			// Setup client with mock
			client := githubMCP.NewClient(tc.mockedClient)
			_, handler := GetIssueComments(stubGetClientFn(client), translations.NullTranslationHelper)

			// Create call request
			request := createMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			textContent := getTextResult(t, result)

			// Unmarshal and verify the result
			var returnedComments []*github.IssueComment
			err = json.Unmarshal([]byte(textContent.Text), &returnedComments)
			require.NoError(t, err)
			assert.Equal(t, len(tc.expectedComments), len(returnedComments))
			if len(returnedComments) > 0 {
				assert.Equal(t, *tc.expectedComments[0].Body, *returnedComments[0].Body)
				assert.Equal(t, *tc.expectedComments[0].User.Login, *returnedComments[0].User.Login)
			}
		})
	}
}

