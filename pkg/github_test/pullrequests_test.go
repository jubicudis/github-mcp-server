// Canonical test file for pullrequests.go
// All tests must directly and robustly test the canonical logic in pullrequests.go
// Remove all legacy, duplicate, or non-canonical tests
// Reference only helpers from /pkg/common and /pkg/testutil
// No import cycles, duplicate imports, or undefined helpers
// All test cases must match the actual signatures and logic of pullrequests.go

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
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Constants for repeated string literals
const (
	prTitle             = "Test PR"
	prHTMLURL           = "https://github.com/owner/repo/pull/42"
	prFeatureBranch     = "feature-branch"
	prMainBranch        = "main"
	prBody              = "This is a test PR"
	prUser              = "testuser"
	prUpdatedTitle      = "Updated Test PR Title"
	prUpdatedBody       = "Updated test PR body."
	prDevelopBranch     = "develop"
	prClosedState       = "closed"
	prOpenState         = "open"
	prFirstPRTitle      = "First PR"
	prSecondPRTitle     = "Second PR"
	prSecondHTMLURL     = "https://github.com/owner/repo/pull/43"
	prLooksGood         = "Looks good!"
	prNeedsFix          = "This needs to be fixed"
	prMainGo            = "main.go"
	prFile1             = "file1.go"
	prFile2             = "file2.go"
	prLGTM              = "LGTM"
	prApprover          = "approver"
	prReviewer          = "reviewer"
	prReviewer1         = "reviewer1"
	prReviewer2         = "reviewer2"
	prBranchUpdateMsg   = "Branch was updated successfully"
	prBranchUpdateURL   = "https://api.github.com/repos/owner/repo/pulls/42"
	prMergedMsg         = "Pull Request successfully merged"
	prSHA1              = "abcd1234"
	prSHA2              = "efgh5678"
	prCommitID          = "abcdef123456"
	prTravisCI          = "https://travis-ci.org/owner/repo/builds/123"
	prCodecov           = "https://codecov.io/gh/owner/repo/pull/42"
	prGolangCILint      = "https://golangci.com/r/owner/repo/pull/42"
	prSuccessState      = "success"
	prCoverageIncreased = "Coverage increased"
	prNoIssues          = "No issues found"
)

func TestGetPullRequest(t *testing.T) {
	// Verify tool definition once
	defaultClient := github.NewClient(nil)
	clientFn := func(ctx context.Context) (*github.Client, error) { return defaultClient, nil }
	tool, _ := ghmcp.GetPullRequest(clientFn, testutil.NullTranslationHelperFunc)

	assert.Equal(t, "get_pull_request", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "pullNumber")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "pullNumber"})

	// Setup mock PR for success case
	mockPR := &github.PullRequest{
		Number:  testutil.Ptr(42),
		Title:   testutil.Ptr(prTitle),
		State:   testutil.Ptr(prOpenState),
		HTMLURL: testutil.Ptr(prHTMLURL),
		Head: &github.PullRequestBranch{
			SHA: testutil.Ptr(prSHA1),
			Ref: testutil.Ptr(prFeatureBranch),
		},
		Base: &github.PullRequestBranch{
			Ref: testutil.Ptr(prMainBranch),
		},
		Body: testutil.Ptr(prBody),
		User: &github.User{
			Login: testutil.Ptr(prUser),
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
			clientFn := func(ctx context.Context) (*github.Client, error) { return client, nil }
			_, handler := ghmcp.GetPullRequest(clientFn, testutil.NullTranslationHelperFunc)

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
	tool, _ := ghmcp.UpdatePullRequest(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

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
		Number:              testutil.Ptr(42),
		Title:               testutil.Ptr(prUpdatedTitle),
		State:               testutil.Ptr(prOpenState),
		HTMLURL:             testutil.Ptr(prHTMLURL),
		Body:                testutil.Ptr(prUpdatedBody),
		MaintainerCanModify: testutil.Ptr(false),
		Base: &github.PullRequestBranch{
			Ref: testutil.Ptr(prDevelopBranch),
		},
	}

	mockClosedPR := &github.PullRequest{
		Number: testutil.Ptr(42),
		Title:  testutil.Ptr(prTitle),
		State:  testutil.Ptr(prClosedState), // State updated
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
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						// Decode and assert request body if needed
						var requestBody map[string]interface{}
						err := json.NewDecoder(r.Body).Decode(&requestBody)
						require.NoError(t, err)
						assert.Equal(t, prUpdatedTitle, requestBody["title"])
						assert.Equal(t, prUpdatedBody, requestBody["body"])
						assert.Equal(t, prDevelopBranch, requestBody["base"])
						assert.Equal(t, false, requestBody["maintainer_can_modify"])

						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(mockUpdatedPR)
					}),
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
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						var requestBody map[string]interface{}
						err := json.NewDecoder(r.Body).Decode(&requestBody)
						require.NoError(t, err)
						assert.Equal(t, prClosedState, requestBody["state"])

						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(mockClosedPR)
					}),
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
			clientFn := func(ctx context.Context) (*github.Client, error) { return client, nil }
			_, handler := ghmcp.UpdatePullRequest(clientFn, testutil.NullTranslationHelperFunc)
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
	tool, _ := ghmcp.ListPullRequests(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

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
			Number:  testutil.Ptr(42),
			Title:   testutil.Ptr(prFirstPRTitle),
			State:   testutil.Ptr(prOpenState),
			HTMLURL: testutil.Ptr(prHTMLURL),
		},
		{
			Number:  testutil.Ptr(43),
			Title:   testutil.Ptr(prSecondPRTitle),
			State:   testutil.Ptr(prClosedState),
			HTMLURL: testutil.Ptr(prSecondHTMLURL),
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
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						// Optionally decode and assert query parameters here if needed
						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(mockPRs)
					}),
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
			_, handler := ghmcp.ListPullRequests(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

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
	tool, _ := ghmcp.MergePullRequest(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

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
		Merged:  testutil.Ptr(true),
		Message: testutil.Ptr(prMergedMsg),
		SHA:     testutil.Ptr(prSHA1),
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
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						var requestBody map[string]interface{}
						err := json.NewDecoder(r.Body).Decode(&requestBody)
						require.NoError(t, err)
						assert.Equal(t, "Merge PR #42", requestBody["commit_title"])
						assert.Equal(t, "Merging awesome feature", requestBody["commit_message"])
						assert.Equal(t, "squash", requestBody["merge_method"])

						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(mockMergeResult)
					}),
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
			clientFn := func(ctx context.Context) (*github.Client, error) { return client, nil }
			_, handler := ghmcp.MergePullRequest(clientFn, testutil.NullTranslationHelperFunc)

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
	tool, _ := ghmcp.GetPullRequestFiles(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "get_pull_request_files", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "pullNumber")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "pullNumber"})

	// Setup mock PR files for success case
	mockFiles := []*github.CommitFile{
		{
			Filename:  testutil.Ptr(prFile1),
			Status:    testutil.Ptr("modified"),
			Additions: testutil.Ptr(10),
			Deletions: testutil.Ptr(5),
			Changes:   testutil.Ptr(15),
			Patch:     testutil.Ptr("@@ -1,5 +1,10 @@"),
		},
		{
			Filename:  testutil.Ptr(prFile2),
			Status:    testutil.Ptr("added"),
			Additions: testutil.Ptr(20),
			Deletions: testutil.Ptr(0),
			Changes:   testutil.Ptr(20),
			Patch:     testutil.Ptr("@@ -0,0 +1,20 @@"),
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
			_, handler := ghmcp.GetPullRequestFiles(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

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
	tool, _ := ghmcp.GetPullRequestStatus(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "get_pull_request_status", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "pullNumber")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "pullNumber"})

	// Setup mock PR for successful PR fetch
	mockPR := &github.PullRequest{
		Number:  testutil.Ptr(42),
		Title:   testutil.Ptr(prTitle),
		HTMLURL: testutil.Ptr(prHTMLURL),
		Head: &github.PullRequestBranch{
			SHA: testutil.Ptr(prSHA1),
			Ref: testutil.Ptr(prFeatureBranch),
		},
	}

	// Setup mock status for success case
	mockStatus := &github.CombinedStatus{
		State:      testutil.Ptr(prSuccessState),
		TotalCount: testutil.Ptr(3),
		Statuses: []*github.RepoStatus{
			{
				State:       testutil.Ptr(prSuccessState),
				Context:     testutil.Ptr("continuous-integration/travis-ci"),
				Description: testutil.Ptr("Build succeeded"),
				TargetURL:   testutil.Ptr(prTravisCI),
			},
			{
				State:       testutil.Ptr(prSuccessState),
				Context:     testutil.Ptr("codecov/patch"),
				Description: testutil.Ptr(prCoverageIncreased),
				TargetURL:   testutil.Ptr(prCodecov),
			},
			{
				State:       testutil.Ptr(prSuccessState),
				Context:     testutil.Ptr("lint/golangci-lint"),
				Description: testutil.Ptr(prNoIssues),
				TargetURL:   testutil.Ptr(prGolangCILint),
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
			clientFn := func(ctx context.Context) (*github.Client, error) { return client, nil }
			_, handler := ghmcp.GetPullRequestStatus(clientFn, testutil.NullTranslationHelperFunc)

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
	tool, _ := ghmcp.UpdatePullRequestBranch(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "update_pull_request_branch", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "pullNumber")
	assert.Contains(t, tool.InputSchema.Properties, "expectedHeadSha")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "pullNumber"})

	// Setup mock update result for success case
	mockUpdateResult := &github.PullRequestBranchUpdateResponse{
		Message: testutil.Ptr(prBranchUpdateMsg),
		URL:     testutil.Ptr(prBranchUpdateURL),
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
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						var requestBody map[string]interface{}
						err := json.NewDecoder(r.Body).Decode(&requestBody)
						require.NoError(t, err)
						assert.Equal(t, prSHA1, requestBody["expected_head_sha"])

						w.WriteHeader(http.StatusAccepted)
						_ = json.NewEncoder(w).Encode(mockUpdateResult)
					}),
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
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						var requestBody map[string]interface{}
						err := json.NewDecoder(r.Body).Decode(&requestBody)
						require.NoError(t, err)
						assert.Empty(t, requestBody["expected_head_sha"])

						w.WriteHeader(http.StatusAccepted)
						_ = json.NewEncoder(w).Encode(mockUpdateResult)
					}),
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
			clientFn := func(ctx context.Context) (*github.Client, error) { return client, nil }
			_, handler := ghmcp.UpdatePullRequestBranch(clientFn, testutil.NullTranslationHelperFunc)

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
	tool, _ := ghmcp.GetPullRequestComments(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "get_pull_request_comments", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "pullNumber")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "pullNumber"})

	// Setup mock PR comments for success case
	mockComments := []*github.PullRequestComment{
		{
			ID:      testutil.Ptr(int64(101)),
			Body:    testutil.Ptr(prLooksGood),
			HTMLURL: testutil.Ptr("https://github.com/owner/repo/pull/42#discussion_r101"),
			User: &github.User{
				Login: testutil.Ptr(prReviewer1),
			},
			Path:      testutil.Ptr(prFile1),
			Position:  testutil.Ptr(5),
			CommitID:  testutil.Ptr(prCommitID),
			CreatedAt: &github.Timestamp{Time: time.Now().Add(-24 * time.Hour)},
			UpdatedAt: &github.Timestamp{Time: time.Now().Add(-24 * time.Hour)},
		},
		{
			ID:      testutil.Ptr(int64(102)),
			Body:    testutil.Ptr(prNeedsFix),
			HTMLURL: testutil.Ptr("https://github.com/owner/repo/pull/42#discussion_r102"),
			User: &github.User{
				Login: testutil.Ptr(prReviewer2),
			},
			Path:      testutil.Ptr(prFile2),
			Position:  testutil.Ptr(10),
			CommitID:  testutil.Ptr(prCommitID),
			CreatedAt: &github.Timestamp{Time: time.Now().Add(-12 * time.Hour)},
			UpdatedAt: &github.Timestamp{Time: time.Now().Add(-12 * time.Hour)},
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
			_, handler := ghmcp.GetPullRequestComments(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

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
	tool, _ := ghmcp.GetPullRequestReviews(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "get_pull_request_reviews", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, "owner")
	assert.Contains(t, tool.InputSchema.Properties, "repo")
	assert.Contains(t, tool.InputSchema.Properties, "pullNumber")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{"owner", "repo", "pullNumber"})

	// Setup mock PR reviews for success case
	mockReviews := []*github.PullRequestReview{
		{
			ID:      testutil.Ptr(int64(201)),
			State:   testutil.Ptr("APPROVED"),
			Body:    testutil.Ptr(prLGTM),
			HTMLURL: testutil.Ptr("https://github.com/owner/repo/pull/42#pullrequestreview-201"),
			User: &github.User{
				Login: testutil.Ptr(prApprover),
			},
			CommitID:    testutil.Ptr(prCommitID),
			SubmittedAt: &github.Timestamp{Time: time.Now().Add(-24 * time.Hour)},
		},
		{
			ID:      testutil.Ptr(int64(202)),
			State:   testutil.Ptr("CHANGES_REQUESTED"),
			Body:    testutil.Ptr("Please address the following issues"),
			HTMLURL: testutil.Ptr("https://github.com/owner/repo/pull/42#pullrequestreview-202"),
			User: &github.User{
				Login: testutil.Ptr(prReviewer),
			},
			CommitID:    testutil.Ptr(prCommitID),
			SubmittedAt: &github.Timestamp{Time: time.Now().Add(-12 * time.Hour)},
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
			_, handler := ghmcp.GetPullRequestReviews(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

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
	tool, _ := ghmcp.CreatePullRequestReview(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

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
		ID:      testutil.Ptr(int64(301)),
		State:   testutil.Ptr("APPROVED"),
		Body:    testutil.Ptr(prLooksGood),
		HTMLURL: testutil.Ptr("https://github.com/owner/repo/pull/42#pullrequestreview-301"),
		User: &github.User{
			Login: testutil.Ptr(prReviewer),
		},
		CommitID:    testutil.Ptr(prCommitID),
		SubmittedAt: &github.Timestamp{Time: time.Now()},
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
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						var requestBody map[string]interface{}
						err := json.NewDecoder(r.Body).Decode(&requestBody)
						require.NoError(t, err)
						assert.Equal(t, prLooksGood, requestBody["body"])
						assert.Equal(t, "APPROVE", requestBody["event"])

						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(mockReview)
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
			expectError:    false,
			expectedReview: mockReview,
		},
		{
			name: "successful review creation with commitId",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposPullsReviewsByOwnerByRepoByPullNumber,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						var requestBody map[string]interface{}
						err := json.NewDecoder(r.Body).Decode(&requestBody)
						require.NoError(t, err)
						assert.Equal(t, prLooksGood, requestBody["body"])
						assert.Equal(t, "APPROVE", requestBody["event"])
						assert.Equal(t, prCommitID, requestBody["commit_id"])

						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(mockReview)
					}),
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
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						var requestBody map[string]interface{}
						err := json.NewDecoder(r.Body).Decode(&requestBody)
						require.NoError(t, err)
						assert.Equal(t, "REQUEST_CHANGES", requestBody["event"])
						assert.Len(t, requestBody["comments"], 2)

						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(mockReview)
					}),
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
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						var requestBody map[string]interface{}
						err := json.NewDecoder(r.Body).Decode(&requestBody)
						require.NoError(t, err)
						assert.Equal(t, "COMMENT", requestBody["event"])
						assert.Len(t, requestBody["comments"], 1)

						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(mockReview)
					}),
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
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						var requestBody map[string]interface{}
						err := json.NewDecoder(r.Body).Decode(&requestBody)
						require.NoError(t, err)
						assert.Equal(t, "COMMENT", requestBody["event"])
						assert.Len(t, requestBody["comments"], 1)

						w.WriteHeader(http.StatusOK)
						_ = json.NewEncoder(w).Encode(mockReview)
					}),
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
			clientFn := func(ctx context.Context) (*github.Client, error) { return client, nil }
			_, handler := ghmcp.CreatePullRequestReview(clientFn, testutil.NullTranslationHelperFunc)

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
	tool, _ := ghmcp.CreatePullRequest(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

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
		Number:  testutil.Ptr(42),
		Title:   testutil.Ptr(prTitle),
		State:   testutil.Ptr(prOpenState),
		HTMLURL: testutil.Ptr(prHTMLURL),
		Head: &github.PullRequestBranch{
			SHA: testutil.Ptr(prSHA1),
			Ref: testutil.Ptr(prFeatureBranch),
		},
		Base: &github.PullRequestBranch{
			SHA: testutil.Ptr(prSHA2),
			Ref: testutil.Ptr(prMainBranch),
		},
		Body:                testutil.Ptr(prBody),
		Draft:               testutil.Ptr(false),
		MaintainerCanModify: testutil.Ptr(true),
		User: &github.User{
			Login: testutil.Ptr(prUser),
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
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						var requestBody map[string]interface{}
						err := json.NewDecoder(r.Body).Decode(&requestBody)
						require.NoError(t, err)
						assert.Equal(t, prTitle, requestBody["title"])
						assert.Equal(t, prBody, requestBody["body"])
						assert.Equal(t, prFeatureBranch, requestBody["head"])
						assert.Equal(t, prMainBranch, requestBody["base"])
						assert.Equal(t, false, requestBody["draft"])
						assert.Equal(t, true, requestBody["maintainer_can_modify"])

						w.WriteHeader(http.StatusCreated)
						_ = json.NewEncoder(w).Encode(mockPR)
					}),
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
			clientFn := func(ctx context.Context) (*github.Client, error) { return client, nil }
			_, handler := ghmcp.CreatePullRequest(clientFn, testutil.NullTranslationHelperFunc)

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
