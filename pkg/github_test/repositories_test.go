// Canonical test file for repositories.go
// All tests must directly and robustly test the canonical logic in repositories.go
// Remove all legacy, duplicate, or non-canonical tests
// Reference only helpers from /pkg/common and /pkg/testutil
// No import cycles, duplicate imports, or undefined helpers
// All test cases must match the actual signatures and logic of repositories.go

package ghmcp_test

// References:
//   - Implementation: pkg/github/repositories.go
//   - Common helpers: pkg/common/common.go
//   - Test utilities: pkg/testutil/testutil.go

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

	ghmcp "github.com/jubicudis/github-mcp-server/pkg/github"
	"github.com/jubicudis/github-mcp-server/pkg/testutil"
)

// Implementation under test: pkg/github/repositories.go defines the canonical functions for repository operations.
// This test suite directly exercises those implementations using inline clientFn closures and live HTTP mock transports.
// Cross-reference documentation in pkg/common and pkg/testutil for helper usage.

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
	// Define client function without stubs
	clientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(nil), nil }
	tool, _ := ghmcp.GetFileContents(clientFn, testutil.NullTranslationHelperFunc)

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
			// Create client function using mocked HTTP client transport
			clientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(tc.mockedClient), nil }
			_, handler := ghmcp.GetFileContents(clientFn, testutil.NullTranslationHelperFunc)

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
	// For tool definition testing, a placeholder client is used to initialize the tool's metadata.
	// This client is not used for actual API calls in the test cases below.
	toolDefinitionClientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(nil), nil }
	tool, _ := ghmcp.ForkRepository(toolDefinitionClientFn, testutil.NullTranslationHelperFunc)

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
			// Setup client with mock for test execution
			clientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(tc.mockedClient), nil }
			_, handler := ghmcp.ForkRepository(clientFn, testutil.NullTranslationHelperFunc)

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
	// For tool definition testing, a placeholder client is used to initialize the tool's metadata.
	// This client is not used for actual API calls in the test cases below.
	toolDefinitionClientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(nil), nil }
	tool, _ := ghmcp.CreateBranch(toolDefinitionClientFn, testutil.NullTranslationHelperFunc)

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
			// Setup client with mock for test execution
			clientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(tc.mockedClient), nil }
			_, handler := ghmcp.CreateBranch(clientFn, testutil.NullTranslationHelperFunc)

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
	// For tool definition testing, a placeholder client is used to initialize the tool's metadata.
	// This client is not used for actual API calls in the test cases below.
	toolDefinitionClientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(nil), nil }
	tool, _ := ghmcp.GetCommit(toolDefinitionClientFn, testutil.NullTranslationHelperFunc)

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
			// Setup client with mock for test execution
			clientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(tc.mockedClient), nil }
			_, handler := ghmcp.GetCommit(clientFn, testutil.NullTranslationHelperFunc)

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
	// For tool definition testing, a placeholder client is used to initialize the tool's metadata.
	// This client is not used for actual API calls in the test cases below.
	toolDefinitionClientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(nil), nil }
	tool, _ := ghmcp.ListCommits(toolDefinitionClientFn, testutil.NullTranslationHelperFunc)

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
			// Setup client with mock for test execution
			clientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(tc.mockedClient), nil }
			_, handler := ghmcp.ListCommits(clientFn, testutil.NullTranslationHelperFunc)

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
	// For tool definition testing, a placeholder client is used to initialize the tool's metadata.
	// This client is not used for actual API calls in the test cases below.
	toolDefinitionClientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(nil), nil }
	tool, _ := ghmcp.CreateOrUpdateFile(toolDefinitionClientFn, testutil.NullTranslationHelperFunc)

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
			// Setup client with mock for test execution
			clientFn := func(ctx context.Context) (*github.Client, error) { return github.NewClient(tc.mockedClient), nil }
			_, handler := ghmcp.CreateOrUpdateFile(clientFn, testutil.NullTranslationHelperFunc)

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
