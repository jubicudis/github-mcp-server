// WHO: GitHubMCPPullRequestTests
// WHAT: Pull Request API Testing
// WHEN: During test execution
// WHERE: MCP Bridge Layer Testing
// WHY: To verify PR functionality
// HOW: By testing MCP protocol handlers
// EXTENT: All PR operations
package github_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v71/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	githubpkg "github.com/jubicudis/github-mcp-server/pkg/github"
	"github.com/jubicudis/github-mcp-server/pkg/github/testutil"
)

// Alias helpers for legacy test code
var expectRequestBody = testutil.MockResponse
var expectQueryParams = testutil.CreateQueryParamExpectation

func TestGetPullRequest(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := githubpkg.GetPullRequest(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "get_pull_request", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "pullNumber")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "pullNumber"})

	// Setup mock PR for success case
	mockPR := &github.PullRequest{
		Number:  testutil.PtrTo(42),
		Title:   testutil.PtrTo(prTitle),
		State:   testutil.PtrTo(prOpenState),
		HTMLURL: testutil.PtrTo(prHTMLURL),
		Head: &github.PullRequestBranch{
			SHA: testutil.PtrTo(prSHA1),
			Ref: testutil.PtrTo(prFeatureBranch),
		},
		Base: &github.PullRequestBranch{
			Ref: testutil.PtrTo(prMainBranch),
		},
		Body: testutil.PtrTo(prBody),
		User: &github.User{
			Login: testutil.PtrTo(prUser),
		},
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedPR     *github.PullRequest
		expectedErrMsg string
	}{
		{
			name: "successful PR fetch",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposPullsByOwnerByRepoByPullNumber,
					mockPR,
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
			},
			expectError: false,
			expectedPR:  mockPR,
		},
		{
			name: "PR fetch fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposPullsByOwnerByRepoByPullNumber,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(999),
			},
			expectError:    true,
			expectedErrMsg: "failed to get pull request",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := githubpkg.GetPullRequest(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

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
			var returnedPR github.PullRequest
			err = json.Unmarshal([]byte(textContent), &returnedPR)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedPR.Number, *returnedPR.Number)
			assert.Equal(t, *tc.expectedPR.Title, *returnedPR.Title)
			assert.Equal(t, *tc.expectedPR.State, *returnedPR.State)
			assert.Equal(t, *tc.expectedPR.HTMLURL, *returnedPR.HTMLURL)
		})
	}
}

func assertPullRequestResult(t *testing.T, expected, actual *github.PullRequest) {
	require.NotNil(t, actual)
	assert.Equal(t, *expected.Number, *actual.Number)
	if expected.Title != nil {
		assert.Equal(t, *expected.Title, *actual.Title)
	}
	if expected.State != nil {
		assert.Equal(t, *expected.State, *actual.State)
	}
	if expected.HTMLURL != nil {
		assert.Equal(t, *expected.HTMLURL, *actual.HTMLURL)
	}
	if expected.Body != nil {
		assert.Equal(t, *expected.Body, *actual.Body)
	}
	if expected.MaintainerCanModify != nil {
		assert.Equal(t, *expected.MaintainerCanModify, *actual.MaintainerCanModify)
	}
	if expected.Base != nil && expected.Base.Ref != nil {
		assert.Equal(t, *expected.Base.Ref, *actual.Base.Ref)
	}
}

func TestUpdatePullRequest(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := UpdatePullRequest(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "update_pull_request", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "pullNumber")
	assert.Contains(t, tool.InputSchema.Properties, "title")
	assert.Contains(t, tool.InputSchema.Properties, "body")
	assert.Contains(t, tool.InputSchema.Properties, "state")
	assert.Contains(t, tool.InputSchema.Properties, "base")
	assert.Contains(t, tool.InputSchema.Properties, "maintainer_can_modify")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "pullNumber"})

	// Setup mock PR for success case
	mockUpdatedPR := &github.PullRequest{
		Number:              testutil.PtrTo(42),
		Title:               testutil.PtrTo(prUpdatedTitle),
		State:               testutil.PtrTo(prOpenState),
		HTMLURL:             testutil.PtrTo(prHTMLURL),
		Body:                testutil.PtrTo(prUpdatedBody),
		MaintainerCanModify: testutil.PtrTo(false),
		Base: &github.PullRequestBranch{
			Ref: testutil.PtrTo(prDevelopBranch),
		},
	}

	mockClosedPR := &github.PullRequest{
		Number: testutil.PtrTo(42),
		Title:  testutil.PtrTo(prTitle),
		State:  testutil.PtrTo(prClosedState), // State updated
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedPR     *github.PullRequest
		expectedErrMsg string
	}{
		{
			name: "successful PR update (title, body, base, maintainer_can_modify)",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PatchReposPullsByOwnerByRepoByPullNumber,
					// Expect the flat string based on previous test failure output and API docs
					expectRequestBody(t, map[string]interface{}{
						"title":                 prUpdatedTitle,
						"body":                  prUpdatedBody,
						"base":                  prDevelopBranch,
						"maintainer_can_modify": false,
					}).andThen(
						mockResponse(t, http.StatusOK, mockUpdatedPR),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":                 "owner",
				"repo":                  "repo",
				"pullNumber":            float64(42),
				"title":                 prUpdatedTitle,
				"body":                  prUpdatedBody,
				"base":                  prDevelopBranch,
				"maintainer_can_modify": false,
			},
			expectError: false,
			expectedPR:  mockUpdatedPR,
		},
		{
			name: "successful PR update (state)",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PatchReposPullsByOwnerByRepoByPullNumber,
					expectRequestBody(t, map[string]interface{}{
						"state": prClosedState,
					}).andThen(
						mockResponse(t, http.StatusOK, mockClosedPR),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
				"state":      prClosedState,
			},
			expectError: false,
			expectedPR:  mockClosedPR,
		},
		{
			name:         "no update parameters provided",
			mockedClient: mock.NewMockedHTTPClient(), // No API call expected
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
				// No update fields
			},
			expectError:    false, // Error is returned in the result, not as Go error
			expectedErrMsg: "No update parameters provided",
		},
		{
			name: "PR update fails (API error)",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PatchReposPullsByOwnerByRepoByPullNumber,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusUnprocessableEntity)
						_, _ = w.Write([]byte(`{"message": "Validation Failed"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
				"title":      "Invalid Title Causing Error",
			},
			expectError:    true,
			expectedErrMsg: "failed to update pull request",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := github.NewClient(tc.mockedClient)
			_, handler := UpdatePullRequest(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)
			request := testutil.CreateMCPRequest(tc.requestArgs)
			result, err := handler(context.Background(), request)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			textContent := testutil.GetTextResult(t, result)

			if tc.expectedErrMsg != "" {
				assert.Contains(t, textContent, tc.expectedErrMsg)
				return
			}

			var returnedPR github.PullRequest
			err = json.Unmarshal([]byte(textContent), &returnedPR)
			require.NoError(t, err)
			assertPullRequestResult(t, tc.expectedPR, &returnedPR)
		})
	}
}

func TestListPullRequests(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := githubpkg.ListPullRequests(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "list_pull_requests", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "state")
	assert.Contains(t, tool.InputSchema.Properties, "head")
	assert.Contains(t, tool.InputSchema.Properties, "base")
	assert.Contains(t, tool.InputSchema.Properties, "sort")
	assert.Contains(t, tool.InputSchema.Properties, "direction")
	assert.Contains(t, tool.InputSchema.Properties, "perPage")
	assert.Contains(t, tool.InputSchema.Properties, "page")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo"})

	// Setup mock PRs for success case
	mockPRs := []*github.PullRequest{
		{
			Number:  testutil.PtrTo(42),
			Title:   testutil.PtrTo(prFirstPRTitle),
			State:   testutil.PtrTo(prOpenState),
			HTMLURL: testutil.PtrTo(prHTMLURL),
		},
		{
			Number:  testutil.PtrTo(43),
			Title:   testutil.PtrTo(prSecondPRTitle),
			State:   testutil.PtrTo(prClosedState),
			HTMLURL: testutil.PtrTo(prSecondHTMLURL),
		},
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedPRs    []*github.PullRequest
		expectedErrMsg string
	}{
		{
			name: "successful PRs listing",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposPullsByOwnerByRepo,
					expectQueryParams(t, map[string]string{
						"state":     "all",
						"sort":      "created",
						"direction": "desc",
						"per_page":  "30",
						"page":      "1",
					}).andThen(
						mockResponse(t, http.StatusOK, mockPRs),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":     "owner",
				"repo":      "repo",
				"state":     "all",
				"sort":      "created",
				"direction": "desc",
				"perPage":   float64(30),
				"page":      float64(1),
			},
			expectError: false,
			expectedPRs: mockPRs,
		},
		{
			name: "PRs listing fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposPullsByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusBadRequest)
						_, _ = w.Write([]byte(`{"message": "Invalid request"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner": "owner",
				"repo":  "repo",
				"state": "invalid",
			},
			expectError:    true,
			expectedErrMsg: "failed to list pull requests",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := githubpkg.ListPullRequests(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

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
			var returnedPRs []*github.PullRequest
			err = json.Unmarshal([]byte(textContent), &returnedPRs)
			require.NoError(t, err)
			assert.Len(t, returnedPRs, 2)
			assert.Equal(t, *tc.expectedPRs[0].Number, *returnedPRs[0].Number)
			assert.Equal(t, *tc.expectedPRs[0].Title, *returnedPRs[0].Title)
			assert.Equal(t, *tc.expectedPRs[0].State, *returnedPRs[0].State)
			assert.Equal(t, *tc.expectedPRs[1].Number, *returnedPRs[1].Number)
			assert.Equal(t, *tc.expectedPRs[1].Title, *returnedPRs[1].Title)
			assert.Equal(t, *tc.expectedPRs[1].State, *returnedPRs[1].State)
		})
	}
}

func TestMergePullRequest(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := MergePullRequest(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "merge_pull_request", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "pullNumber")
	assert.Contains(t, tool.InputSchema.Properties, "commit_title")
	assert.Contains(t, tool.InputSchema.Properties, "commit_message")
	assert.Contains(t, tool.InputSchema.Properties, "merge_method")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "pullNumber"})

	// Setup mock merge result for success case
	mockMergeResult := &github.PullRequestMergeResult{
		Merged:  testutil.PtrTo(true),
		Message: testutil.PtrTo(prMergedMsg),
		SHA:     testutil.PtrTo(prSHA1),
	}

	tests := []struct {
		name                string
		mockedClient        *http.Client
		requestArgs         map[string]interface{}
		expectError         bool
		expectedMergeResult *github.PullRequestMergeResult
		expectedErrMsg      string
	}{
		{
			name: "successful merge",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PutReposPullsMergeByOwnerByRepoByPullNumber,
					expectRequestBody(t, map[string]interface{}{
						"commit_title":   "Merge PR #42",
						"commit_message": "Merging awesome feature",
						"merge_method":   "squash",
					}).andThen(
						mockResponse(t, http.StatusOK, mockMergeResult),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":          "owner",
				"repo":           "repo",
				"pullNumber":     float64(42),
				"commit_title":   "Merge PR #42",
				"commit_message": "Merging awesome feature",
				"merge_method":   "squash",
			},
			expectError:         false,
			expectedMergeResult: mockMergeResult,
		},
		{
			name: "merge fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PutReposPullsMergeByOwnerByRepoByPullNumber,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusMethodNotAllowed)
						_, _ = w.Write([]byte(`{"message": "Pull request cannot be merged"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
			},
			expectError:    true,
			expectedErrMsg: "failed to merge pull request",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := MergePullRequest(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

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
			var returnedResult github.PullRequestMergeResult
			err = json.Unmarshal([]byte(textContent), &returnedResult)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedMergeResult.Merged, *returnedResult.Merged)
			assert.Equal(t, *tc.expectedMergeResult.Message, *returnedResult.Message)
			assert.Equal(t, *tc.expectedMergeResult.SHA, *returnedResult.SHA)
		})
	}
}

func TestGetPullRequestFiles(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := GetPullRequestFiles(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "get_pull_request_files", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "pullNumber")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "pullNumber"})

	// Setup mock PR files for success case
	mockFiles := []*github.CommitFile{
		{
			Filename:  testutil.PtrTo(prFile1),
			Status:    testutil.PtrTo("modified"),
			Additions: testutil.PtrTo(10),
			Deletions: testutil.PtrTo(5),
			Changes:   testutil.PtrTo(15),
			Patch:     testutil.PtrTo("@@ -1,5 +1,10 @@"),
		},
		{
			Filename:  testutil.PtrTo(prFile2),
			Status:    testutil.PtrTo("added"),
			Additions: testutil.PtrTo(20),
			Deletions: testutil.PtrTo(0),
			Changes:   testutil.PtrTo(20),
			Patch:     testutil.PtrTo("@@ -0,0 +1,20 @@"),
		},
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedFiles  []*github.CommitFile
		expectedErrMsg string
	}{
		{
			name: "successful files fetch",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposPullsFilesByOwnerByRepoByPullNumber,
					mockFiles,
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
			},
			expectError:   false,
			expectedFiles: mockFiles,
		},
		{
			name: "files fetch fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposPullsFilesByOwnerByRepoByPullNumber,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(999),
			},
			expectError:    true,
			expectedErrMsg: "failed to get pull request files",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := GetPullRequestFiles(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

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
			var returnedFiles []*github.CommitFile
			err = json.Unmarshal([]byte(textContent), &returnedFiles)
			require.NoError(t, err)
			assert.Len(t, returnedFiles, len(tc.expectedFiles))
			for i, file := range returnedFiles {
				assert.Equal(t, *tc.expectedFiles[i].Filename, *file.Filename)
				assert.Equal(t, *tc.expectedFiles[i].Status, *file.Status)
				assert.Equal(t, *tc.expectedFiles[i].Additions, *file.Additions)
				assert.Equal(t, *tc.expectedFiles[i].Deletions, *file.Deletions)
			}
		})
	}
}

func TestGetPullRequestStatus(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := GetPullRequestStatus(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "get_pull_request_status", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "pullNumber")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "pullNumber"})

	// Setup mock PR for successful PR fetch
	mockPR := &github.PullRequest{
		Number:  testutil.PtrTo(42),
		Title:   testutil.PtrTo(prTitle),
		HTMLURL: testutil.PtrTo(prHTMLURL),
		Head: &github.PullRequestBranch{
			SHA: testutil.PtrTo(prSHA1),
			Ref: testutil.PtrTo(prFeatureBranch),
		},
	}

	// Setup mock status for success case
	mockStatus := &github.CombinedStatus{
		State:      testutil.PtrTo(prSuccessState),
		TotalCount: testutil.PtrTo(3),
		Statuses: []*github.RepoStatus{
			{
				State:       testutil.PtrTo(prSuccessState),
				Context:     testutil.PtrTo("continuous-integration/travis-ci"),
				Description: testutil.PtrTo("Build succeeded"),
				TargetURL:   testutil.PtrTo(prTravisCI),
			},
			{
				State:       testutil.PtrTo(prSuccessState),
				Context:     testutil.PtrTo("codecov/patch"),
				Description: testutil.PtrTo(prCoverageIncreased),
				TargetURL:   testutil.PtrTo(prCodecov),
			},
			{
				State:       testutil.PtrTo(prSuccessState),
				Context:     testutil.PtrTo("lint/golangci-lint"),
				Description: testutil.PtrTo(prNoIssues),
				TargetURL:   testutil.PtrTo(prGolangCILint),
			},
		},
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedStatus *github.CombinedStatus
		expectedErrMsg string
	}{
		{
			name: "successful status fetch",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposPullsByOwnerByRepoByPullNumber,
					mockPR,
				),
				mock.WithRequestMatch(
					mock.GetReposCommitsStatusByOwnerByRepoByRef,
					mockStatus,
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
			},
			expectError:    false,
			expectedStatus: mockStatus,
		},
		{
			name: "PR fetch fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposPullsByOwnerByRepoByPullNumber,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(999),
			},
			expectError:    true,
			expectedErrMsg: "failed to get pull request",
		},
		{
			name: "status fetch fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposPullsByOwnerByRepoByPullNumber,
					mockPR,
				),
				mock.WithRequestMatchHandler(
					mock.GetReposCommitsStatusesByOwnerByRepoByRef,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
			},
			expectError:    true,
			expectedErrMsg: "failed to get combined status",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := GetPullRequestStatus(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

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
			var returnedStatus github.CombinedStatus
			err = json.Unmarshal([]byte(textContent), &returnedStatus)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedStatus.State, *returnedStatus.State)
			assert.Equal(t, *tc.expectedStatus.TotalCount, *returnedStatus.TotalCount)
			assert.Len(t, returnedStatus.Statuses, len(tc.expectedStatus.Statuses))
			for i, status := range returnedStatus.Statuses {
				assert.Equal(t, *tc.expectedStatus.Statuses[i].State, *status.State)
				assert.Equal(t, *tc.expectedStatus.Statuses[i].Context, *status.Context)
				assert.Equal(t, *tc.expectedStatus.Statuses[i].Description, *status.Description)
			}
		})
	}
}

func TestUpdatePullRequestBranch(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := UpdatePullRequestBranch(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "update_pull_request_branch", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "pullNumber")
	assert.Contains(t, tool.InputSchema.Properties, "expectedHeadSha")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "pullNumber"})

	// Setup mock update result for success case
	mockUpdateResult := &github.PullRequestBranchUpdateResponse{
		Message: testutil.PtrTo(prBranchUpdateMsg),
		URL:     testutil.PtrTo(prBranchUpdateURL),
	}

	tests := []struct {
		name                 string
		mockedClient         *http.Client
		requestArgs          map[string]interface{}
		expectError          bool
		expectedUpdateResult *github.PullRequestBranchUpdateResponse
		expectedErrMsg       string
	}{
		{
			name: "successful branch update",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PutReposPullsUpdateBranchByOwnerByRepoByPullNumber,
					expectRequestBody(t, map[string]interface{}{
						"expected_head_sha": prSHA1,
					}).andThen(
						mockResponse(t, http.StatusAccepted, mockUpdateResult),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":           "owner",
				"repo":            "repo",
				"pullNumber":      float64(42),
				"expectedHeadSha": prSHA1,
			},
			expectError:          false,
			expectedUpdateResult: mockUpdateResult,
		},
		{
			name: "branch update without expected SHA",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PutReposPullsUpdateBranchByOwnerByRepoByPullNumber,
					expectRequestBody(t, map[string]interface{}{}).andThen(
						mockResponse(t, http.StatusAccepted, mockUpdateResult),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
			},
			expectError:          false,
			expectedUpdateResult: mockUpdateResult,
		},
		{
			name: "branch update fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PutReposPullsUpdateBranchByOwnerByRepoByPullNumber,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusConflict)
						_, _ = w.Write([]byte(`{"message": "Merge conflict"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
			},
			expectError:    true,
			expectedErrMsg: "failed to update pull request branch",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := UpdatePullRequestBranch(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

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

			assert.Contains(t, textContent, "is in progress")
		})
	}
}

func TestGetPullRequestComments(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := GetPullRequestComments(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "get_pull_request_comments", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "pullNumber")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "pullNumber"})

	// Setup mock PR comments for success case
	mockComments := []*github.PullRequestComment{
		{
			ID:      testutil.PtrTo(int64(101)),
			Body:    testutil.PtrTo(prLooksGood),
			HTMLURL: testutil.PtrTo("https://github.com/owner/repo/pull/42#discussion_r101"),
			User: &github.User{
				Login: testutil.PtrTo(prReviewer1),
			},
			Path:      testutil.PtrTo(prFile1),
			Position:  testutil.PtrTo(5),
			CommitID:  testutil.PtrTo(prCommitID),
			CreatedAt: testutil.PtrTo(github.Timestamp{Time: time.Now().Add(-24 * time.Hour)}),
			UpdatedAt: testutil.PtrTo(github.Timestamp{Time: time.Now().Add(-24 * time.Hour)}),
		},
		{
			ID:      testutil.PtrTo(int64(102)),
			Body:    testutil.PtrTo(prNeedsFix),
			HTMLURL: testutil.PtrTo("https://github.com/owner/repo/pull/42#discussion_r102"),
			User: &github.User{
				Login: testutil.PtrTo(prReviewer2),
			},
			Path:      testutil.PtrTo(prFile2),
			Position:  testutil.PtrTo(10),
			CommitID:  testutil.PtrTo(prCommitID),
			CreatedAt: testutil.PtrTo(github.Timestamp{Time: time.Now().Add(-12 * time.Hour)}),
			UpdatedAt: testutil.PtrTo(github.Timestamp{Time: time.Now().Add(-12 * time.Hour)}),
		},
	}

	tests := []struct {
		name             string
		mockedClient     *http.Client
		requestArgs      map[string]interface{}
		expectError      bool
		expectedComments []*github.PullRequestComment
		expectedErrMsg   string
	}{
		{
			name: "successful comments fetch",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposPullsCommentsByOwnerByRepoByPullNumber,
					mockComments,
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
			},
			expectError:      false,
			expectedComments: mockComments,
		},
		{
			name: "comments fetch fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposPullsCommentsByOwnerByRepoByPullNumber,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(999),
			},
			expectError:    true,
			expectedErrMsg: "failed to get pull request comments",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := GetPullRequestComments(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

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
			var returnedComments []*github.PullRequestComment
			err = json.Unmarshal([]byte(textContent), &returnedComments)
			require.NoError(t, err)
			assert.Len(t, returnedComments, len(tc.expectedComments))
			for i, comment := range returnedComments {
				assert.Equal(t, *tc.expectedComments[i].ID, *comment.ID)
				assert.Equal(t, *tc.expectedComments[i].Body, *comment.Body)
				assert.Equal(t, *tc.expectedComments[i].User.Login, *comment.User.Login)
				assert.Equal(t, *tc.expectedComments[i].Path, *comment.Path)
				assert.Equal(t, *tc.expectedComments[i].HTMLURL, *comment.HTMLURL)
			}
		})
	}
}

func TestGetPullRequestReviews(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := GetPullRequestReviews(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "get_pull_request_reviews", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "pullNumber")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "pullNumber"})

	// Setup mock PR reviews for success case
	mockReviews := []*github.PullRequestReview{
		{
			ID:      testutil.PtrTo(int64(201)),
			State:   testutil.PtrTo("APPROVED"),
			Body:    testutil.PtrTo(prLGTM),
			HTMLURL: testutil.PtrTo("https://github.com/owner/repo/pull/42#pullrequestreview-201"),
			User: &github.User{
				Login: testutil.PtrTo(prApprover),
			},
			CommitID:    testutil.PtrTo(prCommitID),
			SubmittedAt: testutil.PtrTo(github.Timestamp{Time: time.Now().Add(-24 * time.Hour)}),
		},
		{
			ID:      testutil.PtrTo(int64(202)),
			State:   testutil.PtrTo("CHANGES_REQUESTED"),
			Body:    testutil.PtrTo("Please address the following issues"),
			HTMLURL: testutil.PtrTo("https://github.com/owner/repo/pull/42#pullrequestreview-202"),
			User: &github.User{
				Login: testutil.PtrTo(prReviewer),
			},
			CommitID:    testutil.PtrTo(prCommitID),
			SubmittedAt: testutil.PtrTo(github.Timestamp{Time: time.Now().Add(-12 * time.Hour)}),
		},
	}

	tests := []struct {
		name            string
		mockedClient    *http.Client
		requestArgs     map[string]interface{}
		expectError     bool
		expectedReviews []*github.PullRequestReview
		expectedErrMsg  string
	}{
		{
			name: "successful reviews fetch",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposPullsReviewsByOwnerByRepoByPullNumber,
					mockReviews,
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
			},
			expectError:     false,
			expectedReviews: mockReviews,
		},
		{
			name: "reviews fetch fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposPullsReviewsByOwnerByRepoByPullNumber,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(999),
			},
			expectError:    true,
			expectedErrMsg: "failed to get pull request reviews",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := GetPullRequestReviews(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

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
			var returnedReviews []*github.PullRequestReview
			err = json.Unmarshal([]byte(textContent), &returnedReviews)
			require.NoError(t, err)
			assert.Len(t, returnedReviews, len(tc.expectedReviews))
			for i, review := range returnedReviews {
				assert.Equal(t, *tc.expectedReviews[i].ID, *review.ID)
				assert.Equal(t, *tc.expectedReviews[i].State, *review.State)
				assert.Equal(t, *tc.expectedReviews[i].Body, *review.Body)
				assert.Equal(t, *tc.expectedReviews[i].User.Login, *review.User.Login)
				assert.Equal(t, *tc.expectedReviews[i].HTMLURL, *review.HTMLURL)
			}
		})
	}
}

func TestCreatePullRequestReview(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := CreatePullRequestReview(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "create_pull_request_review", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "pullNumber")
	assert.Contains(t, tool.InputSchema.Properties, "body")
	assert.Contains(t, tool.InputSchema.Properties, "event")
	assert.Contains(t, tool.InputSchema.Properties, "commitId")
	assert.Contains(t, tool.InputSchema.Properties, "comments")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "pullNumber", "event"})

	// Setup mock review for success case
	mockReview := &github.PullRequestReview{
		ID:      testutil.PtrTo(int64(301)),
		State:   testutil.PtrTo("APPROVED"),
		Body:    testutil.PtrTo(prLooksGood),
		HTMLURL: testutil.PtrTo("https://github.com/owner/repo/pull/42#pullrequestreview-301"),
		User: &github.User{
			Login: testutil.PtrTo(prReviewer),
		},
		CommitID:    testutil.PtrTo(prCommitID),
		SubmittedAt: testutil.PtrTo(github.Timestamp{Time: time.Now()}),
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedReview *github.PullRequestReview
		expectedErrMsg string
	}{
		{
			name: "successful review creation with body only",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposPullsReviewsByOwnerByRepoByPullNumber,
					expectRequestBody(t, map[string]interface{}{
						"body":  prLooksGood,
						"event": "APPROVE",
					}).andThen(
						mockResponse(t, http.StatusOK, mockReview),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
				"body":       prLooksGood,
				"event":      "APPROVE",
			},
			expectError:    false,
			expectedReview: mockReview,
		},
		{
			name: "successful review creation with commitId",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposPullsReviewsByOwnerByRepoByPullNumber,
					expectRequestBody(t, map[string]interface{}{
						"body":      prLooksGood,
						"event":     "APPROVE",
						"commit_id": prCommitID,
					}).andThen(
						mockResponse(t, http.StatusOK, mockReview),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
				"body":       prLooksGood,
				"event":      "APPROVE",
				"commitId":   prCommitID,
			},
			expectError:    false,
			expectedReview: mockReview,
		},
		{
			name: "successful review creation with comments",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposPullsReviewsByOwnerByRepoByPullNumber,
					expectRequestBody(t, map[string]interface{}{
						"body":  "Some issues to fix",
						"event": "REQUEST_CHANGES",
						"comments": []interface{}{
							map[string]interface{}{
								"path":     prFile1,
								"position": float64(10),
								"body":     prNeedsFix,
							},
							map[string]interface{}{
								"path":     prFile2,
								"position": float64(20),
								"body":     "Consider a different approach here",
							},
						},
					}).andThen(
						mockResponse(t, http.StatusOK, mockReview),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
				"body":       "Some issues to fix",
				"event":      "REQUEST_CHANGES",
				"comments": []interface{}{
					map[string]interface{}{
						"path":     prFile1,
						"position": float64(10),
						"body":     prNeedsFix,
					},
					map[string]interface{}{
						"path":     prFile2,
						"position": float64(20),
						"body":     "Consider a different approach here",
					},
				},
			},
			expectError:    false,
			expectedReview: mockReview,
		},
		{
			name: "invalid comment format",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposPullsReviewsByOwnerByRepoByPullNumber,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusUnprocessableEntity)
						_, _ = w.Write([]byte(`{"message": "Invalid comment format"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
				"event":      "REQUEST_CHANGES",
				"comments": []interface{}{
					map[string]interface{}{
						"path": prFile1,
						// missing position
						"body": prNeedsFix,
					},
				},
			},
			expectError:    false,
			expectedErrMsg: "each comment must have either position or line",
		},
		{
			name: "successful review creation with line parameter",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposPullsReviewsByOwnerByRepoByPullNumber,
					expectRequestBody(t, map[string]interface{}{
						"body":  "Code review comments",
						"event": "COMMENT",
						"comments": []interface{}{
							map[string]interface{}{
								"path": prMainGo,
								"line": float64(42),
								"body": "Consider adding a comment here",
							},
						},
					}).andThen(
						mockResponse(t, http.StatusOK, mockReview),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
				"body":       "Code review comments",
				"event":      "COMMENT",
				"comments": []interface{}{
					map[string]interface{}{
						"path": prMainGo,
						"line": float64(42),
						"body": "Consider adding a comment here",
					},
				},
			},
			expectError:    false,
			expectedReview: mockReview,
		},
		{
			name: "successful review creation with multi-line comment",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposPullsReviewsByOwnerByRepoByPullNumber,
					expectRequestBody(t, map[string]interface{}{
						"body":  "Multi-line comment review",
						"event": "COMMENT",
						"comments": []interface{}{
							map[string]interface{}{
								"path":       prMainGo,
								"start_line": float64(10),
								"line":       float64(15),
								"side":       "RIGHT",
								"body":       "This entire block needs refactoring",
							},
						},
					}).andThen(
						mockResponse(t, http.StatusOK, mockReview),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
				"body":       "Multi-line comment review",
				"event":      "COMMENT",
				"comments": []interface{}{
					map[string]interface{}{
						"path":       prMainGo,
						"start_line": float64(10),
						"line":       float64(15),
						"side":       "RIGHT",
						"body":       "This entire block needs refactoring",
					},
				},
			},
			expectError:    false,
			expectedReview: mockReview,
		},
		{
			name:         "invalid multi-line comment - missing line parameter",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
				"event":      "COMMENT",
				"comments": []interface{}{
					map[string]interface{}{
						"path":       prMainGo,
						"start_line": float64(10),
						// missing line parameter
						"body": "Invalid multi-line comment",
					},
				},
			},
			expectError:    false,
			expectedErrMsg: "each comment must have either position or line", // Updated error message
		},
		{
			name: "invalid comment - mixing position with line parameters",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.PostReposPullsReviewsByOwnerByRepoByPullNumber,
					mockReview,
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
				"event":      "COMMENT",
				"comments": []interface{}{
					map[string]interface{}{
						"path":     prMainGo,
						"position": float64(5),
						"line":     float64(42),
						"body":     "Invalid parameter combination",
					},
				},
			},
			expectError:    false,
			expectedErrMsg: "position cannot be combined with line, side, start_line, or start_side",
		},
		{
			name:         "invalid multi-line comment - missing side parameter",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
				"event":      "COMMENT",
				"comments": []interface{}{
					map[string]interface{}{
						"path":       prMainGo,
						"start_line": float64(10),
						"line":       float64(15),
						"start_side": "LEFT",
						// missing side parameter
						"body": "Invalid multi-line comment",
					},
				},
			},
			expectError:    false,
			expectedErrMsg: "if start_side is provided, side must also be provided",
		},
		{
			name: "review creation fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposPullsReviewsByOwnerByRepoByPullNumber,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusUnprocessableEntity)
						_, _ = w.Write([]byte(`{"message": "Invalid comment format"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":      "owner",
				"repo":       "repo",
				"pullNumber": float64(42),
				"body":       prLooksGood,
				"event":      "APPROVE",
			},
			expectError:    true,
			expectedErrMsg: "failed to create pull request review",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := CreatePullRequestReview(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

			// Create call request
			request := testutil.CreateMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				if err != nil {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
					return
				}

				// If no error returned but in the result
				textContent := testutil.GetTextResult(t, result)
				assert.Contains(t, textContent, tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)

			// For error messages in the result
			if tc.expectedErrMsg != "" {
				textContent := testutil.GetTextResult(t, result)
				assert.Contains(t, textContent, tc.expectedErrMsg)
				return
			}

			// Parse the result and get the text content if no error
			textContent := testutil.GetTextResult(t, result)

			// Unmarshal and verify the result
			var returnedReview github.PullRequestReview
			err = json.Unmarshal([]byte(textContent), &returnedReview)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedReview.ID, *returnedReview.ID)
			assert.Equal(t, *tc.expectedReview.State, *returnedReview.State)
			assert.Equal(t, *tc.expectedReview.Body, *returnedReview.Body)
			assert.Equal(t, *tc.expectedReview.User.Login, *returnedReview.User.Login)
			assert.Equal(t, *tc.expectedReview.HTMLURL, *returnedReview.HTMLURL)
		})
	}
}

func TestCreatePullRequest(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := CreatePullRequest(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "create_pull_request", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "title")
	assert.Contains(t, tool.InputSchema.Properties, "body")
	assert.Contains(t, tool.InputSchema.Properties, "head")
	assert.Contains(t, tool.InputSchema.Properties, "base")
	assert.Contains(t, tool.InputSchema.Properties, "draft")
	assert.Contains(t, tool.InputSchema.Properties, "maintainer_can_modify")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "title", "head", "base"})

	// Setup mock PR for success case
	mockPR := &github.PullRequest{
		Number:  testutil.PtrTo(42),
		Title:   testutil.PtrTo(prTitle),
		State:   testutil.PtrTo(prOpenState),
		HTMLURL: testutil.PtrTo(prHTMLURL),
		Head: &github.PullRequestBranch{
			SHA: testutil.PtrTo(prSHA1),
			Ref: testutil.PtrTo(prFeatureBranch),
		},
		Base: &github.PullRequestBranch{
			SHA: testutil.PtrTo(prSHA2),
			Ref: testutil.PtrTo(prMainBranch),
		},
		Body:                testutil.PtrTo(prBody),
		Draft:               testutil.PtrTo(false),
		MaintainerCanModify: testutil.PtrTo(true),
		User: &github.User{
			Login: testutil.PtrTo(prUser),
		},
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedPR     *github.PullRequest
		expectedErrMsg string
	}{
		{
			name: "successful PR creation",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposPullsByOwnerByRepo,
					expectRequestBody(t, map[string]interface{}{
						"title":                 prTitle,
						"body":                  prBody,
						"head":                  prFeatureBranch,
						"base":                  prMainBranch,
						"draft":                 false,
						"maintainer_can_modify": true,
					}).andThen(
						mockResponse(t, http.StatusCreated, mockPR),
					),
				),
			),
			requestArgs: map[string]interface{}{
				"owner":                 "owner",
				"repo":                  "repo",
				"title":                 prTitle,
				"body":                  prBody,
				"head":                  prFeatureBranch,
				"base":                  prMainBranch,
				"draft":                 false,
				"maintainer_can_modify": true,
			},
			expectError: false,
			expectedPR:  mockPR,
		},
		{
			name:         "missing required parameter",
			mockedClient: mock.NewMockedHTTPClient(),
			requestArgs: map[string]interface{}{
				"owner": "owner",
				"repo":  "repo",
				// missing title, head, base
			},
			expectError:    true,
			expectedErrMsg: "missing required parameter: title",
		},
		{
			name: "PR creation fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposPullsByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusUnprocessableEntity)
						_, _ = w.Write([]byte(`{"message":"Validation failed","errors":[{"resource":"PullRequest","code":"invalid"}]}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				"owner": "owner",
				"repo":  "repo",
				"title": prTitle,
				"head":  prFeatureBranch,
				"base":  prMainBranch,
			},
			expectError:    true,
			expectedErrMsg: "failed to create pull request",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := CreatePullRequest(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

			// Create call request
			request := testutil.CreateMCPRequest(tc.requestArgs)

			// Call handler
			result, err := handler(context.Background(), request)

			// Verify results
			if tc.expectError {
				if err != nil {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
					return
				}

				// If no error returned but in the result
				textContent := testutil.GetTextResult(t, result)
				assert.Contains(t, textContent, tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)

			// Parse the result and get the text content if no error
			textContent := testutil.GetTextResult(t, result)

			// Unmarshal and verify the result
			var returnedPR github.PullRequest
			err = json.Unmarshal([]byte(textContent), &returnedPR)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedPR.Number, *returnedPR.Number)
			assert.Equal(t, *tc.expectedPR.Title, *returnedPR.Title)
			assert.Equal(t, *tc.expectedPR.State, *returnedPR.State)
			assert.Equal(t, *tc.expectedPR.HTMLURL, *returnedPR.HTMLURL)
			assert.Equal(t, *tc.expectedPR.Head.SHA, *returnedPR.Head.SHA)
			assert.Equal(t, *tc.expectedPR.Base.Ref, *returnedPR.Base.Ref)
			assert.Equal(t, *tc.expectedPR.Body, *returnedPR.Body)
			assert.Equal(t, *tc.expectedPR.User.Login, *returnedPR.User.Login)
		})
	}
}
