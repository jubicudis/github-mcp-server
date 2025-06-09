// Canonical test file for repositories.go
// Remove all duplicate imports, fix import cycles, and ensure all tests reference only canonical helpers from /pkg/common and /pkg/testutil
// All test cases must be robust, DRY, and match the implementation in repositories.go

package github

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
	"github.com/jubicudis/github-mcp-server/pkg/log"
	"github.com/jubicudis/github-mcp-server/pkg/testutil"
)

// Replace with real logger
var logger = log.NewLogger().WithLevel(log.LevelDebug)

const readmeFileName = "README.md"

// Test constants for repeated string literals
const (
	branchMain          = "main"
	branchNewFeature    = "new-feature"
	testUser            = "Test User"
	testUserEmail       = "test@example.com"
	testRepoName        = "test-repo"
	testRepoDescription = "Test repository"
	docsExamplePath     = "docs/example.md"
	addExampleMsg       = "Add example file"
	updateFileMsg       = "Update file"
	updateMultipleMsg   = "Update multiple files"
	readmeContent       = "# README"
	exampleContent      = "# Example\n\nThis is an example file."
	refsHeadsMain       = "refs/heads/main"
	userReposPattern    = "/user/repos"
	testReadmeFileName  = "README.md"
)

const (
	ownerKey       = "owner"
	repoKey        = "repo"
	pathKey        = "path"
	branchKey      = "branch"
	shaKey         = "sha"
	commitMessage  = "message"
	contentKey     = "content"
	filesKey       = "files"
	nameKey        = "name"
	descriptionKey = "description"
	privateKey     = "private"
	autoInitKey    = "autoInit"
	pageKey        = "page"
	perPageKey     = "perPage"
)

// Remove legacy/duplicate helper aliases and use only canonical helpers from testutil
// Remove expectRequestBody and expectQueryParams aliases
// Replace all usage of mockResponse, expectQueryParams, Ptr, etc. with testutil.MockResponse, testutil.CreateQueryParamMatcher, testutil.Ptr, etc.

func TestGetFileContents(t *testing.T) {
	mockClient := githubpkg.NewClient("", logger)
	tool, _ := githubpkg.GetFileContents(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "get_file_contents", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, ownerKey)
	assert.Contains(t, tool.InputSchema.Properties, repoKey)
	assert.Contains(t, tool.InputSchema.Properties, pathKey)
	assert.Contains(t, tool.InputSchema.Properties, branchKey)
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{ownerKey, repoKey, pathKey})

	// Setup mock file content for success case
	mockFileContent := &github.RepositoryContent{
		Type:        testutil.Ptr("file"),
		Name:        testutil.Ptr(testReadmeFileName),
		Path:        testutil.Ptr(testReadmeFileName),
		Content:     testutil.Ptr("IyBUZXN0IFJlcG9zaXRvcnkKClRoaXMgaXMgYSB0ZXN0IHJlcG9zaXRvcnku"), // Base64 encoded "# Test Repository\n\nThis is a test repository."
		SHA:         testutil.Ptr("abc123"),
		Size:        testutil.Ptr(42),
		HTMLURL:     testutil.Ptr("https://github.com/owner/repo/blob/main/README.md"),
		DownloadURL: testutil.Ptr("https://raw.githubusercontent.com/owner/repo/main/README.md"),
	}

	// Setup mock directory content for success case
	mockDirContent := []*github.RepositoryContent{
		{
			Type:    testutil.Ptr("file"),
			Name:    testutil.Ptr(readmeFileName),
			Path:    testutil.Ptr(readmeFileName),
			SHA:     testutil.Ptr("abc123"),
			Size:    testutil.Ptr(42),
			HTMLURL: testutil.Ptr("https://github.com/owner/repo/blob/main/README.md"),
		},
		{
			Type:    testutil.Ptr("dir"),
			Name:    testutil.Ptr("src"),
			Path:    testutil.Ptr("src"),
			SHA:     testutil.Ptr("def456"),
			HTMLURL: testutil.Ptr("https://github.com/owner/repo/tree/main/src"),
		},
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedResult interface{}
		expectedErrMsg string
	}{
		{
			name: "successful file content fetch",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposContentsByOwnerByRepoByPath,
					testutil.CreateQueryParamMatcher(t, map[string]string{
						"ref": branchMain,
					}).AndThen(
						testutil.MockResponse(t, http.StatusOK, mockFileContent),
					),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:  "owner",
				repoKey:   "repo",
				pathKey:   readmeFileName,
				branchKey: branchMain,
			},
			expectError:    false,
			expectedResult: mockFileContent,
		},
		{
			name: "successful directory content fetch",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposContentsByOwnerByRepoByPath,
					testutil.CreateQueryParamMatcher(t, map[string]string{}).AndThen(
						testutil.MockResponse(t, http.StatusOK, mockDirContent),
					),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey: "owner",
				repoKey:  "repo",
				pathKey:  "src",
			},
			expectError:    false,
			expectedResult: mockDirContent,
		},
		{
			name: "content fetch fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposContentsByOwnerByRepoByPath,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:  "owner",
				repoKey:   "repo",
				pathKey:   "nonexistent.md",
				branchKey: branchMain,
			},
			expectError:    true,
			expectedErrMsg: "failed to get file contents",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := githubpkg.NewClient("", logger)
			_, handler := githubpkg.GetFileContents(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

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

			// Verify based on expected type
			switch expected := tc.expectedResult.(type) {
			case *github.RepositoryContent:
				var returnedContent github.RepositoryContent
				err = json.Unmarshal([]byte(textContent), &returnedContent)
				require.NoError(t, err)
				assert.Equal(t, *expected.Name, *returnedContent.Name)
				assert.Equal(t, *expected.Path, *returnedContent.Path)
				assert.Equal(t, *expected.Type, *returnedContent.Type)
			case []*github.RepositoryContent:
				var returnedContents []*github.RepositoryContent
				err = json.Unmarshal([]byte(textContent), &returnedContents)
				require.NoError(t, err)
				assert.Len(t, returnedContents, len(expected))
				for i, content := range returnedContents {
					assert.Equal(t, *expected[i].Name, *content.Name)
					assert.Equal(t, *expected[i].Path, *content.Path)
					assert.Equal(t, *expected[i].Type, *content.Type)
				}
			}
		})
	}
}

func TestForkRepository(t *testing.T) {
	mockClient := githubpkg.NewClient("", logger)
	tool, _ := githubpkg.ForkRepository(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "fork_repository", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, ownerKey)
	assert.Contains(t, tool.InputSchema.Properties, repoKey)
	assert.Contains(t, tool.InputSchema.Properties, "organization")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{ownerKey, repoKey})

	// Setup mock forked repo for success case
	mockForkedRepo := &github.Repository{
		ID:       testutil.Ptr(int64(123456)),
		Name:     testutil.Ptr("repo"),
		FullName: testutil.Ptr("new-owner/repo"),
		Owner: &github.User{
			Login: testutil.Ptr("new-owner"),
		},
		HTMLURL:       testutil.Ptr("https://github.com/new-owner/repo"),
		DefaultBranch: testutil.Ptr(branchMain),
		Fork:          testutil.Ptr(true),
		ForksCount:    testutil.Ptr(0),
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedRepo   *github.Repository
		expectedErrMsg string
	}{
		{
			name: "successful repository fork",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposForksByOwnerByRepo,
					testutil.MockResponse(t, http.StatusAccepted, mockForkedRepo),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey: "owner",
				repoKey:  "repo",
			},
			expectError:  false,
			expectedRepo: mockForkedRepo,
		},
		{
			name: "repository fork fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostReposForksByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusForbidden)
						_, _ = w.Write([]byte(`{"message": "Forbidden"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey: "owner",
				repoKey:  "repo",
			},
			expectError:    true,
			expectedErrMsg: "failed to fork repository",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := githubpkg.NewClient("", logger)
			_, handler := githubpkg.ForkRepository(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

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

			assert.Contains(t, textContent, "Fork is in progress")
		})
	}
}

func TestCreateBranch(t *testing.T) {
	mockClient := githubpkg.NewClient("", logger)
	tool, _ := githubpkg.CreateBranch(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "create_branch", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, ownerKey)
	assert.Contains(t, tool.InputSchema.Properties, repoKey)
	assert.Contains(t, tool.InputSchema.Properties, "branch")
	assert.Contains(t, tool.InputSchema.Properties, "from_branch")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{ownerKey, repoKey, "branch"})

	// Setup mock repository for default branch test
	mockRepo := &github.Repository{
		DefaultBranch: testutil.Ptr(branchMain),
	}

	// Setup mock reference for from_branch tests
	mockSourceRef := &github.Reference{
		Ref: testutil.Ptr(refsHeadsMain),
		Object: &github.GitObject{
			SHA: testutil.Ptr("abc123def456"),
		},
	}

	// Setup mock created reference
	mockCreatedRef := &github.Reference{
		Ref: testutil.Ptr("refs/heads/" + branchNewFeature),
		Object: &github.GitObject{
			SHA: testutil.Ptr("abc123def456"),
		},
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedRef    *github.Reference
		expectedErrMsg string
	}{
		{
			name: "successful branch creation with from_branch",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposGitRefByOwnerByRepoByRef,
					mockSourceRef,
				),
				mock.WithRequestMatch(
					mock.PostReposGitRefsByOwnerByRepo,
					mockCreatedRef,
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:      "owner",
				repoKey:       "repo",
				"branch":      branchNewFeature,
				"from_branch": branchMain,
			},
			expectError: false,
			expectedRef: mockCreatedRef,
		},
		{
			name: "successful branch creation with default branch",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposByOwnerByRepo,
					mockRepo,
				),
				mock.WithRequestMatch(
					mock.GetReposGitRefByOwnerByRepoByRef,
					mockSourceRef,
				),
				mock.WithRequestMatchHandler(
					mock.PostReposGitRefsByOwnerByRepo,
					testutil.MockResponse(t, http.StatusCreated, mockCreatedRef),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey: "owner",
				repoKey:  "repo",
				"branch": branchNewFeature,
			},
			expectError: false,
			expectedRef: mockCreatedRef,
		},
		{
			name: "fail to get repository",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Repository not found"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey: "owner",
				repoKey:  "nonexistent-repo",
				"branch": branchNewFeature,
			},
			expectError:    true,
			expectedErrMsg: "failed to get repository",
		},
		{
			name: "fail to get reference",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposGitRefByOwnerByRepoByRef,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Reference not found"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:      "owner",
				repoKey:       "repo",
				"branch":      branchNewFeature,
				"from_branch": "nonexistent-branch",
			},
			expectError:    true,
			expectedErrMsg: "failed to get reference",
		},
		{
			name: "fail to create branch",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposGitRefByOwnerByRepoByRef,
					mockSourceRef,
				),
				mock.WithRequestMatchHandler(
					mock.PostReposGitRefsByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusUnprocessableEntity)
						_, _ = w.Write([]byte(`{"message": "Reference already exists"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:      "owner",
				repoKey:       "repo",
				"branch":      "existing-branch",
				"from_branch": branchMain,
			},
			expectError:    true,
			expectedErrMsg: "failed to create branch",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := githubpkg.NewClient("", logger)
			_, handler := githubpkg.CreateBranch(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

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
			var returnedRef github.Reference
			err = json.Unmarshal([]byte(textContent), &returnedRef)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedRef.Ref, *returnedRef.Ref)
			assert.Equal(t, *tc.expectedRef.Object.SHA, *returnedRef.Object.SHA)
		})
	}
}

func TestGetCommit(t *testing.T) {
	mockClient := githubpkg.NewClient("", logger)
	tool, _ := githubpkg.GetCommit(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "get_commit", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, ownerKey)
	assert.Contains(t, tool.InputSchema.Properties, repoKey)
	assert.Contains(t, tool.InputSchema.Properties, shaKey)
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{ownerKey, repoKey, shaKey})

	mockCommit := &github.RepositoryCommit{
		SHA: testutil.Ptr("abc123def456"),
		Commit: &github.Commit{
			Message: testutil.Ptr("First commit"),
			Author: &github.CommitAuthor{
				Name:  testutil.Ptr(testUser),
				Email: testutil.Ptr(testUserEmail),
				Date:  &github.Timestamp{Time: time.Now().Add(-48 * time.Hour)},
			},
		},
		Author: &github.User{
			Login: testutil.Ptr("testuser"),
		},
		HTMLURL: testutil.Ptr("https://github.com/owner/repo/commit/abc123def456"),
		Stats: &github.CommitStats{
			Additions: testutil.Ptr(10),
			Deletions: testutil.Ptr(2),
			Total:     testutil.Ptr(12),
		},
		Files: []*github.CommitFile{
			{
				Filename:  testutil.Ptr("file1.go"),
				Status:    testutil.Ptr("modified"),
				Additions: testutil.Ptr(10),
				Deletions: testutil.Ptr(2),
				Changes:   testutil.Ptr(12),
				Patch:     testutil.Ptr("@@ -1,2 +1,10 @@"),
			},
		},
	}
	// This one currently isn't defined in the mock package we're using.
	var mockEndpointPattern = mock.EndpointPattern{
		Pattern: "/repos/{owner}/{repo}/commits/{sha}",
		Method:  "GET",
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedCommit *github.RepositoryCommit
		expectedErrMsg string
	}{
		{
			name: "successful commit fetch",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mockEndpointPattern,
					testutil.MockResponse(t, http.StatusOK, mockCommit),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey: "owner",
				repoKey:  "repo",
				shaKey:   "abc123def456",
			},
			expectError:    false,
			expectedCommit: mockCommit,
		},
		{
			name: "commit fetch fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mockEndpointPattern,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey: "owner",
				repoKey:  "repo",
				shaKey:   "nonexistent-sha",
			},
			expectError:    true,
			expectedErrMsg: "failed to get commit",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := githubpkg.NewClient("", logger)
			_, handler := githubpkg.GetCommit(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

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
			var returnedCommit github.RepositoryCommit
			err = json.Unmarshal([]byte(textContent), &returnedCommit)
			require.NoError(t, err)

			assert.Equal(t, *tc.expectedCommit.SHA, *returnedCommit.SHA)
			assert.Equal(t, *tc.expectedCommit.Commit.Message, *returnedCommit.Commit.Message)
			assert.Equal(t, *tc.expectedCommit.Author.Login, *returnedCommit.Author.Login)
			assert.Equal(t, *tc.expectedCommit.HTMLURL, *returnedCommit.HTMLURL)
		})
	}
}

func TestListCommits(t *testing.T) {
	mockClient := githubpkg.NewClient("", logger)
	tool, _ := githubpkg.ListCommits(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "list_commits", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, ownerKey)
	assert.Contains(t, tool.InputSchema.Properties, repoKey)
	assert.Contains(t, tool.InputSchema.Properties, shaKey)
	assert.Contains(t, tool.InputSchema.Properties, pageKey)
	assert.Contains(t, tool.InputSchema.Properties, perPageKey)
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{ownerKey, repoKey})

	// Setup mock commits for success case
	mockCommits := []*github.RepositoryCommit{
		{
			SHA: testutil.Ptr("abc123def456"),
			Commit: &github.Commit{
				Message: testutil.Ptr("First commit"),
				Author: &github.CommitAuthor{
					Name:  testutil.Ptr(testUser),
					Email: testutil.Ptr(testUserEmail),
					Date:  &github.Timestamp{Time: time.Now().Add(-48 * time.Hour)},
				},
			},
			Author: &github.User{
				Login: testutil.Ptr("testuser"),
			},
			HTMLURL: testutil.Ptr("https://github.com/owner/repo/commit/abc123def456"),
		},
		{
			SHA: testutil.Ptr("def456abc789"),
			Commit: &github.Commit{
				Message: testutil.Ptr("Second commit"),
				Author: &github.CommitAuthor{
					Name:  testutil.Ptr("Another User"),
					Email: testutil.Ptr("another@example.com"),
					Date:  &github.Timestamp{Time: time.Now().Add(-24 * time.Hour)},
				},
			},
			Author: &github.User{
				Login: testutil.Ptr("anotheruser"),
			},
			HTMLURL: testutil.Ptr("https://github.com/owner/repo/commit/def456abc789"),
		},
	}

	tests := []struct {
		name            string
		mockedClient    *http.Client
		requestArgs     map[string]interface{}
		expectError     bool
		expectedCommits []*github.RepositoryCommit
		expectedErrMsg  string
	}{
		{
			name: "successful commits fetch with default params",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposCommitsByOwnerByRepo,
					mockCommits,
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey: "owner",
				repoKey:  "repo",
			},
			expectError:     false,
			expectedCommits: mockCommits,
		},
		{
			name: "successful commits fetch with branch",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposCommitsByOwnerByRepo,
					testutil.CreateQueryParamMatcher(t, map[string]string{
						"sha":      branchMain,
						"page":     "1",
						"per_page": "30",
					}).AndThen(
						testutil.MockResponse(t, http.StatusOK, mockCommits),
					),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey: "owner",
				repoKey:  "repo",
				shaKey:   branchMain,
			},
			expectError:     false,
			expectedCommits: mockCommits,
		},
		{
			name: "successful commits fetch with pagination",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposCommitsByOwnerByRepo,
					testutil.CreateQueryParamMatcher(t, map[string]string{
						"page":     "2",
						"per_page": "10",
					}).AndThen(
						testutil.MockResponse(t, http.StatusOK, mockCommits),
					),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:   "owner",
				repoKey:    "repo",
				pageKey:    float64(2),
				perPageKey: float64(10),
			},
			expectError:     false,
			expectedCommits: mockCommits,
		},
		{
			name: "commits fetch fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposCommitsByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"message": "Not Found"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey: "owner",
				repoKey:  "nonexistent-repo",
			},
			expectError:    true,
			expectedErrMsg: "failed to list commits",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := githubpkg.NewClient("", logger)
			_, handler := githubpkg.ListCommits(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

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
			var returnedCommits []*github.RepositoryCommit
			err = json.Unmarshal([]byte(textContent), &returnedCommits)
			require.NoError(t, err)
			assert.Len(t, returnedCommits, len(tc.expectedCommits))
			for i, commit := range returnedCommits {
				assert.Equal(t, *tc.expectedCommits[i].SHA, *commit.SHA)
				assert.Equal(t, *tc.expectedCommits[i].Commit.Message, *commit.Commit.Message)
				assert.Equal(t, *tc.expectedCommits[i].Author.Login, *commit.Author.Login)
				assert.Equal(t, *tc.expectedCommits[i].HTMLURL, *commit.HTMLURL)
			}
		})
	}
}

func TestCreateOrUpdateFile(t *testing.T) {
	mockClient := githubpkg.NewClient("", logger)
	tool, _ := githubpkg.CreateOrUpdateFile(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "create_or_update_file", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, ownerKey)
	assert.Contains(t, tool.InputSchema.Properties, repoKey)
	assert.Contains(t, tool.InputSchema.Properties, pathKey)
	assert.Contains(t, tool.InputSchema.Properties, contentKey)
	assert.Contains(t, tool.InputSchema.Properties, commitMessage)
	assert.Contains(t, tool.InputSchema.Properties, branchKey)
	assert.Contains(t, tool.InputSchema.Properties, shaKey)
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{ownerKey, repoKey, pathKey, contentKey, commitMessage, branchKey})

	// Setup mock file content response
	mockFileResponse := &github.RepositoryContentResponse{
		Content: &github.RepositoryContent{
			Name:        testutil.Ptr("example.md"),
			Path:        testutil.Ptr(docsExamplePath),
			SHA:         testutil.Ptr("abc123def456"),
			Size:        testutil.Ptr(42),
			HTMLURL:     testutil.Ptr("https://github.com/owner/repo/blob/main/docs/example.md"),
			DownloadURL: testutil.Ptr("https://raw.githubusercontent.com/owner/repo/main/docs/example.md"),
		},
		Commit: github.Commit{
			SHA:     testutil.Ptr("def456abc789"),
			Message: testutil.Ptr(addExampleMsg),
			Author: &github.CommitAuthor{
				Name:  testutil.Ptr(testUser),
				Email: testutil.Ptr(testUserEmail),
				Date:  &github.Timestamp{Time: time.Now()},
			},
			HTMLURL: testutil.Ptr("https://github.com/owner/repo/commit/def456abc789"),
		},
	}

	tests := []struct {
		name            string
		mockedClient    *http.Client
		requestArgs     map[string]interface{}
		expectError     bool
		expectedContent *github.RepositoryContentResponse
		expectedErrMsg  string
	}{
		{
			name: "successful file creation",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PutReposContentsByOwnerByRepoByPath,
					testutil.CreateQueryParamMatcher(t, map[string]string{
						commitMessage: addExampleMsg,
						contentKey:    "IyBFeGFtcGxlCgpUaGlzIGlzIGFuIGV4YW1wbGUgZmlsZS4=", // Base64 encoded content
						branchKey:     branchMain,
					}).AndThen(
						testutil.MockResponse(t, http.StatusOK, mockFileResponse),
					),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:      "owner",
				repoKey:       "repo",
				pathKey:       docsExamplePath,
				contentKey:    exampleContent,
				commitMessage: addExampleMsg,
				branchKey:     branchMain,
			},
			expectError:     false,
			expectedContent: mockFileResponse,
		},
		{
			name: "successful file update with SHA",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PutReposContentsByOwnerByRepoByPath,
					testutil.CreateQueryParamMatcher(t, map[string]string{
						commitMessage: "Update example file",
						contentKey:    "IyBVcGRhdGVkIEV4YW1wbGUKClRoaXMgZmlsZSBoYXMgYmVlbiB1cGRhdGVkLg==", // Base64 encoded content
						branchKey:     branchMain,
						shaKey:        "abc123def456",
					}).AndThen(
						testutil.MockResponse(t, http.StatusOK, mockFileResponse),
					),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:      "owner",
				repoKey:       "repo",
				pathKey:       docsExamplePath,
				contentKey:    "# Updated Example\n\nThis file has been updated.",
				commitMessage: "Update example file",
				branchKey:     branchMain,
				shaKey:        "abc123def456",
			},
			expectError:     false,
			expectedContent: mockFileResponse,
		},
		{
			name: "file creation fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PutReposContentsByOwnerByRepoByPath,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusUnprocessableEntity)
						_, _ = w.Write([]byte(`{"message": "Invalid request"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:      "owner",
				repoKey:       "repo",
				pathKey:       docsExamplePath,
				contentKey:    "#Invalid Content",
				commitMessage: "Invalid request",
				branchKey:     "nonexistent-branch",
			},
			expectError:    true,
			expectedErrMsg: "failed to create/update file",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := githubpkg.NewClient("", logger)
			_, handler := githubpkg.CreateOrUpdateFile(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

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
			var returnedContent github.RepositoryContentResponse
			err = json.Unmarshal([]byte(textContent), &returnedContent)
			require.NoError(t, err)

			// Verify content
			assert.Equal(t, *tc.expectedContent.Content.Name, *returnedContent.Content.Name)
			assert.Equal(t, *tc.expectedContent.Content.Path, *returnedContent.Content.Path)
			assert.Equal(t, *tc.expectedContent.Content.SHA, *returnedContent.Content.SHA)

			// Verify commit
			assert.Equal(t, *tc.expectedContent.Commit.SHA, *returnedContent.Commit.SHA)
			assert.Equal(t, *tc.expectedContent.Commit.Message, *returnedContent.Commit.Message)
		})
	}
}

func TestCreateRepository(t *testing.T) {
	logger := log.NewLogger().WithLevel(log.LevelDebug)
	mockClient := githubpkg.NewClient("", logger)
	tool, _ := githubpkg.CreateRepository(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "create_repository", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, nameKey)
	assert.Contains(t, tool.InputSchema.Properties, descriptionKey)
	assert.Contains(t, tool.InputSchema.Properties, privateKey)
	assert.Contains(t, tool.InputSchema.Properties, autoInitKey)
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{nameKey})

	// Setup mock repository response
	mockRepo := &github.Repository{
		Name:        testutil.Ptr(testRepoName),
		Description: testutil.Ptr(testRepoDescription),
		Private:     testutil.Ptr(true),
		HTMLURL:     testutil.Ptr("https://github.com/testuser/test-repo"),
		CloneURL:    testutil.Ptr("https://github.com/testuser/test-repo.git"),
		CreatedAt:   &github.Timestamp{Time: time.Now()},
		Owner: &github.User{
			Login: testutil.Ptr("testuser"),
		},
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedRepo   *github.Repository
		expectedErrMsg string
	}{
		{
			name: "successful repository creation with all parameters",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.EndpointPattern{
						Pattern: userReposPattern,
						Method:  "POST",
					},
					testutil.MockResponse(t, http.StatusCreated, mockRepo),
				),
			),
			requestArgs: map[string]interface{}{
				nameKey:        testRepoName,
				descriptionKey: testRepoDescription,
				privateKey:     true,
				autoInitKey:    true,
			},
			expectError:  false,
			expectedRepo: mockRepo,
		},
		{
			name: "successful repository creation with minimal parameters",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.EndpointPattern{
						Pattern: userReposPattern,
						Method:  "POST",
					},
					testutil.MockResponse(t, http.StatusCreated, mockRepo),
				),
			),
			requestArgs: map[string]interface{}{
				nameKey: testRepoName,
			},
			expectError:  false,
			expectedRepo: mockRepo,
		},
		{
			name: "repository creation fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.EndpointPattern{
						Pattern: userReposPattern,
						Method:  "POST",
					},
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.WriteHeader(http.StatusUnprocessableEntity)
						_, _ = w.Write([]byte(`{"message": "Repository creation failed"}`))
					}),
				),
			),
			requestArgs: map[string]interface{}{
				nameKey: "invalid-repo",
			},
			expectError:    true,
			expectedErrMsg: "failed to create repository",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := githubpkg.NewClient("", logger)
			_, handler := githubpkg.CreateRepository(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

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
			var returnedRepo github.Repository
			err = json.Unmarshal([]byte(textContent), &returnedRepo)
			assert.NoError(t, err)

			// Verify repository details
			assert.Equal(t, *tc.expectedRepo.Name, *returnedRepo.Name)
			assert.Equal(t, *tc.expectedRepo.Description, *returnedRepo.Description)
			assert.Equal(t, *tc.expectedRepo.Private, *returnedRepo.Private)
			assert.Equal(t, *tc.expectedRepo.HTMLURL, *returnedRepo.HTMLURL)
			assert.Equal(t, *tc.expectedRepo.Owner.Login, *returnedRepo.Owner.Login)
		})
	}
}

func TestPushFiles(t *testing.T) {
	logger := log.NewLogger().WithLevel(log.LevelDebug)
	mockClient := githubpkg.NewClient("", logger)
	tool, _ := githubpkg.PushFiles(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

	assert.Equal(t, "push_files", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, ownerKey)
	assert.Contains(t, tool.InputSchema.Properties, repoKey)
	assert.Contains(t, tool.InputSchema.Properties, branchKey)
	assert.Contains(t, tool.InputSchema.Properties, filesKey)
	assert.Contains(t, tool.InputSchema.Properties, commitMessage)
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{ownerKey, repoKey, branchKey, filesKey, commitMessage})

	// Setup mock objects
	mockRef := &github.Reference{
		Ref: testutil.Ptr(refsHeadsMain),
		Object: &github.GitObject{
			SHA: testutil.Ptr("abc123"),
			URL: testutil.Ptr("https://api.github.com/repos/owner/repo/git/trees/abc123"),
		},
	}

	mockCommit := &github.Commit{
		SHA: testutil.Ptr("abc123"),
		Tree: &github.Tree{
			SHA: testutil.Ptr("def456"),
		},
	}

	mockTree := &github.Tree{
		SHA: testutil.Ptr("ghi789"),
	}

	mockNewCommit := &github.Commit{
		SHA:     testutil.Ptr("jkl012"),
		Message: testutil.Ptr(updateMultipleMsg),
		HTMLURL: testutil.Ptr("https://github.com/owner/repo/commit/jkl012"),
	}

	mockUpdatedRef := &github.Reference{
		Ref: testutil.Ptr(refsHeadsMain),
		Object: &github.GitObject{
			SHA: testutil.Ptr("jkl012"),
			URL: testutil.Ptr("https://api.github.com/repos/owner/repo/git/trees/jkl012"),
		},
	}

	// Define test cases
	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]interface{}
		expectError    bool
		expectedRef    *github.Reference
		expectedErrMsg string
	}{
		{
			name: "successful push of multiple files",
			mockedClient: mock.NewMockedHTTPClient(
				// Get branch reference
				mock.WithRequestMatch(
					mock.GetReposGitRefByOwnerByRepoByRef,
					mockRef,
				),
				// Get commit
				mock.WithRequestMatch(
					mock.GetReposGitCommitsByOwnerByRepoByCommitSha,
					mockCommit,
				),
				// Create tree
				mock.WithRequestMatchHandler(
					mock.PostReposGitTreesByOwnerByRepo,
					testutil.CreateQueryParamMatcher(t, map[string]string{
						"base_tree": "def456",
						"tree": []interface{}{
							map[string]interface{}{
								"path":    testReadmeFileName,
								"mode":    "100644",
								"type":    "blob",
								"content": "# Updated README\n\nThis is an updated README file.",
							},
							map[string]interface{}{
								"path":    docsExamplePath,
								"mode":    "100644",
								"type":    "blob",
								"content": exampleContent,
							},
						},
					}).AndThen(
						testutil.MockResponse(t, http.StatusCreated, mockTree),
					),
				),
				// Create commit
				mock.WithRequestMatchHandler(
					mock.PostReposGitCommitsByOwnerByRepo,
					testutil.CreateQueryParamMatcher(t, map[string]string{
						commitMessage: updateMultipleMsg,
						"tree":        "ghi789",
						"parents":     "abc123",
					}).AndThen(
						testutil.MockResponse(t, http.StatusCreated, mockNewCommit),
					),
				),
				// Update reference
				mock.WithRequestMatchHandler(
					mock.PatchReposGitRefsByOwnerByRepoByRef,
					testutil.CreateQueryParamMatcher(t, map[string]string{
						"sha":   "jkl012",
						"force": "false",
					}).AndThen(
						testutil.MockResponse(t, http.StatusOK, mockUpdatedRef),
					),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:  "owner",
				repoKey:   "repo",
				branchKey: branchMain,
				filesKey: []interface{}{
					map[string]interface{}{
						"path":     testReadmeFileName,
						contentKey: "# Updated README\n\nThis is an updated README file.",
					},
					map[string]interface{}{
						"path":     docsExamplePath,
						contentKey: exampleContent,
					},
				},
				commitMessage: updateMultipleMsg,
			},
			expectError: false,
			expectedRef: mockUpdatedRef,
		},
		{
			name:         "fails when files parameter is invalid",
			mockedClient: mock.NewMockedHTTPClient(
			// No requests expected
			),
			requestArgs: map[string]interface{}{
				ownerKey:      "owner",
				repoKey:       "repo",
				branchKey:     branchMain,
				filesKey:      "invalid-files-parameter", // Not an array
				commitMessage: updateMultipleMsg,
			},
			expectError:    false, // This returns a tool error, not a Go error
			expectedErrMsg: "files parameter must be an array",
		},
		{
			name: "fails when files contains object without path",
			mockedClient: mock.NewMockedHTTPClient(
				// Get branch reference
				mock.WithRequestMatch(
					mock.GetReposGitRefByOwnerByRepoByRef,
					mockRef,
				),
				// Get commit
				mock.WithRequestMatch(
					mock.GetReposGitCommitsByOwnerByRepoByCommitSha,
					mockCommit,
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:  "owner",
				repoKey:   "repo",
				branchKey: branchMain,
				filesKey: []interface{}{
					map[string]interface{}{
						contentKey: "# Missing path",
					},
				},
				commitMessage: updateFileMsg,
			},
			expectError:    false, // This returns a tool error, not a Go error
			expectedErrMsg: "each file must have a path",
		},
		{
			name: "fails when files contains object without content",
			mockedClient: mock.NewMockedHTTPClient(
				// Get branch reference
				mock.WithRequestMatch(
					mock.GetReposGitRefByOwnerByRepoByRef,
					mockRef,
				),
				// Get commit
				mock.WithRequestMatch(
					mock.GetReposGitCommitsByOwnerByRepoByCommitSha,
					mockCommit,
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:  "owner",
				repoKey:   "repo",
				branchKey: branchMain,
				filesKey: []interface{}{
					map[string]interface{}{
						"path": "README.md",
						// Missing content
					},
				},
				commitMessage: updateFileMsg,
			},
			expectError:    false, // This returns a tool error, not a Go error
			expectedErrMsg: "each file must have content",
		},
		{
			name: "fails to get branch reference",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposGitRefByOwnerByRepoByRef,
					testutil.MockResponse(t, http.StatusNotFound, nil),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:  "owner",
				repoKey:   "repo",
				branchKey: "non-existent-branch",
				filesKey: []interface{}{
					map[string]interface{}{
						"path":     testReadmeFileName,
						contentKey: readmeContent,
					},
				},
				commitMessage: updateFileMsg,
			},
			expectError:    true,
			expectedErrMsg: "failed to get branch reference",
		},
		{
			name: "fails to get base commit",
			mockedClient: mock.NewMockedHTTPClient(
				// Get branch reference
				mock.WithRequestMatch(
					mock.GetReposGitRefByOwnerByRepoByRef,
					mockRef,
				),
				// Fail to get commit
				mock.WithRequestMatchHandler(
					mock.GetReposGitCommitsByOwnerByRepoByCommitSha,
					testutil.MockResponse(t, http.StatusNotFound, nil),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:  "owner",
				repoKey:   "repo",
				branchKey: branchMain,
				filesKey: []interface{}{
					map[string]interface{}{
						"path":     readmeFileName,
						contentKey: readmeContent,
					},
				},
				commitMessage: updateFileMsg,
			},
			expectError:    true,
			expectedErrMsg: "failed to get base commit",
		},
		{
			name: "fails to create tree",
			mockedClient: mock.NewMockedHTTPClient(
				// Get branch reference
				mock.WithRequestMatch(
					mock.GetReposGitRefByOwnerByRepoByRef,
					mockRef,
				),
				// Get commit
				mock.WithRequestMatch(
					mock.GetReposGitCommitsByOwnerByRepoByCommitSha,
					mockCommit,
				),
				// Fail to create tree
				mock.WithRequestMatchHandler(
					mock.PostReposGitTreesByOwnerByRepo,
					testutil.MockResponse(t, http.StatusInternalServerError, nil),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:  "owner",
				repoKey:   "repo",
				branchKey: branchMain,
				filesKey: []interface{}{
					map[string]interface{}{
						"path":     readmeFileName,
						contentKey: readmeContent,
					},
				},
				commitMessage: updateFileMsg,
			},
			expectError:    true,
			expectedErrMsg: "failed to create tree",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := githubpkg.NewClient("", logger)
			_, handler := githubpkg.PushFiles(testutil.StubGetClientFnWithClient(client), testutil.NullTranslationHelperFunc)

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
				textContent := testutil.GetTextResult(t, result)
				assert.Contains(t, textContent, tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)

			// Parse the result and get the text content if no error
			textContent := testutil.GetTextResult(t, result)

			// Unmarshal and verify the result
			var returnedRef github.Reference
			err = json.Unmarshal([]byte(textContent), &returnedRef)
			require.NoError(t, err)

			assert.Equal(t, *tc.expectedRef.Ref, *returnedRef.Ref)
			assert.Equal(t, *tc.expectedRef.Object.SHA, *returnedRef.Object.SHA)
		})
	}
}

// TODO: Implement or replace with canonical logic
// func TestListBranches(t *testing.t) {
// 	logger := log.NewLogger().WithLevel(log.LevelDebug)
// 	mockClient := githubpkg.NewClient("", logger)
// 	tool, _ := githubpkg.ListBranches(testutil.StubGetClientFnWithClient(mockClient), testutil.NullTranslationHelperFunc)

// 	assert.Equal(t, "list_branches", tool.Name)
// 	assert.NotEmpty(t, tool.Description)
// 	assert.Contains(t, tool.InputSchema.Properties, ownerKey)
// 	assert.Contains(t, tool.InputSchema.Properties, repoKey)
// 	assert.Contains(t, tool.InputSchema.Properties, pageKey)
// 	assert.Contains(t, tool.InputSchema.Properties, perPageKey)
// 	assert.ElementsMatch(t, tool.InputSchema.Required, []string{ownerKey, repoKey})

// 	// Setup mock branches for success case
// 	mockBranches := []*github.Branch{
// 		{
// 			Name:   testutil.Ptr(branchMain),
// 			Commit: &github.RepositoryCommit{SHA: testutil.Ptr("abc123")},
// 		},
// 		{
// 			Name:   testutil.Ptr("develop"),
// 			Commit: &github.RepositoryCommit{SHA: testutil.Ptr("def456")},
// 		},
// 	}

// 	// Test cases
// 	tests := []struct {
// 		name          string
// 		args          map[string]interface{}
// 		mockResponses []mock.MockBackendOption
// 		wantErr       bool
// 		errContains   string
// 	}{
// 		{
// 			name: "success",
// 			args: map[string]interface{}{
// 				ownerKey: "owner",
// 				repoKey:  "repo",
// 				pageKey:  float64(2),
// 			},
// 			mockResponses: []mock.MockBackendOption{
// 				mock.WithRequestMatch(
// 					mock.GetReposBranchesByOwnerByRepo,
// 					mockBranches,
// 				),
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "missing owner",
// 			args: map[string]interface{}{
// 				repoKey: "repo",
// 			},
// 			mockResponses: []mock.MockBackendOption{},
// 			wantErr:       false,
// 			errContains:   "missing required parameter: owner",
// 		},
// 		{
// 			name: "missing repo",
// 			args: map[string]interface{}{
// 				ownerKey: "owner",
// 			},
// 			mockResponses: []mock.MockBackendOption{},
// 			wantErr:       false,
// 			errContains:   "missing required parameter: repo",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Create mock client
// 			mockClient := github.NewClient(mock.NewMockedHTTPClient(tt.mockResponses...))
// 			_, handler := ListBranches(testutil.StubGetClientFn(mockClient), testutil.NullTranslationHelperFunc)

// 			// Create request
// 			request := testutil.CreateMCPRequest(tt.args)

// 			// Call handler
// 			result, err := handler(context.Background(), request)
// 			if tt.wantErr {
// 				require.Error(t, err)
// 				if tt.errContains != "" {
// 					assert.Contains(t, err.Error(), tt.errContains)
// 				}
// 				return
// 			}

// 			require.NoError(t, err)
// 			require.NotNil(t, result)

// 			if tt.errContains != "" {
// 				textContent := testutil.GetTextResult(t, result)
// 				assert.Contains(t, textContent, tt.errContains)
// 				return
// 			}

// 			textContent := testutil.GetTextResult(t, result)
// 			require.NotEmpty(t, textContent)

// 			// Verify response
// 			var branches []*github.Branch
// 			err = json.Unmarshal([]byte(textContent), &branches)
// 			require.NoError(t, err)
// 			assert.Len(t, branches, 2)
// 			assert.Equal(t, branchMain, *branches[0].Name)
// 			assert.Equal(t, "develop", *branches[1].Name)
// 		})
// 	}
// }
