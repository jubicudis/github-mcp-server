package github

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v49/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Ptr[T any](v T) *T { return &v }

var NullTranslationHelperFunc = func(key, defaultValue string) string { return defaultValue }

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

func Test_GetFileContents(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := GetFileContents(stubGetClientFn(mockClient), NullTranslationHelperFunc)

	assert.Equal(t, "get_file_contents", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, ownerKey)
	assert.Contains(t, tool.InputSchema.Properties, repoKey)
	assert.Contains(t, tool.InputSchema.Properties, pathKey)
	assert.Contains(t, tool.InputSchema.Properties, branchKey)
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{ownerKey, repoKey, pathKey})

	// Setup mock file content for success case
	mockFileContent := &github.RepositoryContent{
		Type:        Ptr("file"),
		Name:        Ptr("README.md"),
		Path:        Ptr("README.md"),
		Content:     Ptr("IyBUZXN0IFJlcG9zaXRvcnkKClRoaXMgaXMgYSB0ZXN0IHJlcG9zaXRvcnku"), // Base64 encoded "# Test Repository\n\nThis is a test repository."
		SHA:         Ptr("abc123"),
		Size:        Ptr(42),
		HTMLURL:     Ptr("https://github.com/owner/repo/blob/main/README.md"),
		DownloadURL: Ptr("https://raw.githubusercontent.com/owner/repo/main/README.md"),
	}

	// Setup mock directory content for success case
	mockDirContent := []*github.RepositoryContent{
		{
			Type:    Ptr("file"),
			Name:    Ptr("README.md"),
			Path:    Ptr("README.md"),
			SHA:     Ptr("abc123"),
			Size:    Ptr(42),
			HTMLURL: Ptr("https://github.com/owner/repo/blob/main/README.md"),
		},
		{
			Type:    Ptr("dir"),
			Name:    Ptr("src"),
			Path:    Ptr("src"),
			SHA:     Ptr("def456"),
			HTMLURL: Ptr("https://github.com/owner/repo/tree/main/src"),
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
					expectQueryParams(t, map[string]string{
						"ref": "main",
					}).andThen(
						mockResponse(t, http.StatusOK, mockFileContent),
					),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:  "owner",
				repoKey:   "repo",
				pathKey:   "README.md",
				branchKey: "main",
			},
			expectError:    false,
			expectedResult: mockFileContent,
		},
		{
			name: "successful directory content fetch",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposContentsByOwnerByRepoByPath,
					expectQueryParams(t, map[string]string{}).andThen(
						mockResponse(t, http.StatusOK, mockDirContent),
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
				branchKey: "main",
			},
			expectError:    true,
			expectedErrMsg: "failed to get file contents",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := GetFileContents(stubGetClientFn(client), NullTranslationHelperFunc)

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

			require.NoError(t, err)

			// Parse the result and get the text content if no error
			textContent := getTextResult(t, result)

			// Verify based on expected type
			switch expected := tc.expectedResult.(type) {
			case *github.RepositoryContent:
				var returnedContent github.RepositoryContent
				err = json.Unmarshal([]byte(textContent.Text), &returnedContent)
				require.NoError(t, err)
				assert.Equal(t, *expected.Name, *returnedContent.Name)
				assert.Equal(t, *expected.Path, *returnedContent.Path)
				assert.Equal(t, *expected.Type, *returnedContent.Type)
			case []*github.RepositoryContent:
				var returnedContents []*github.RepositoryContent
				err = json.Unmarshal([]byte(textContent.Text), &returnedContents)
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

func Test_ForkRepository(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := ForkRepository(stubGetClientFn(mockClient), NullTranslationHelperFunc)

	assert.Equal(t, "fork_repository", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, ownerKey)
	assert.Contains(t, tool.InputSchema.Properties, repoKey)
	assert.Contains(t, tool.InputSchema.Properties, "organization")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{ownerKey, repoKey})

	// Setup mock forked repo for success case
	mockForkedRepo := &github.Repository{
		ID:       Ptr(int64(123456)),
		Name:     Ptr("repo"),
		FullName: Ptr("new-owner/repo"),
		Owner: &github.User{
			Login: Ptr("new-owner"),
		},
		HTMLURL:       Ptr("https://github.com/new-owner/repo"),
		DefaultBranch: Ptr("main"),
		Fork:          Ptr(true),
		ForksCount:    Ptr(0),
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
					mockResponse(t, http.StatusAccepted, mockForkedRepo),
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
			client := github.NewClient(tc.mockedClient)
			_, handler := ForkRepository(stubGetClientFn(client), NullTranslationHelperFunc)

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

			assert.Contains(t, textContent.Text, "Fork is in progress")
		})
	}
}

func Test_CreateBranch(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := CreateBranch(stubGetClientFn(mockClient), NullTranslationHelperFunc)

	assert.Equal(t, "create_branch", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, ownerKey)
	assert.Contains(t, tool.InputSchema.Properties, repoKey)
	assert.Contains(t, tool.InputSchema.Properties, "branch")
	assert.Contains(t, tool.InputSchema.Properties, "from_branch")
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{ownerKey, repoKey, "branch"})

	// Setup mock repository for default branch test
	mockRepo := &github.Repository{
		DefaultBranch: Ptr("main"),
	}

	// Setup mock reference for from_branch tests
	mockSourceRef := &github.Reference{
		Ref: Ptr("refs/heads/main"),
		Object: &github.GitObject{
			SHA: Ptr("abc123def456"),
		},
	}

	// Setup mock created reference
	mockCreatedRef := &github.Reference{
		Ref: Ptr("refs/heads/new-feature"),
		Object: &github.GitObject{
			SHA: Ptr("abc123def456"),
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
				"branch":      "new-feature",
				"from_branch": "main",
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
					expectRequestBody(t, map[string]interface{}{
						"ref": "refs/heads/new-feature",
						"sha": "abc123def456",
					}).andThen(
						mockResponse(t, http.StatusCreated, mockCreatedRef),
					),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey: "owner",
				repoKey:  "repo",
				"branch": "new-feature",
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
				"branch": "new-feature",
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
				"branch":      "new-feature",
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
				"from_branch": "main",
			},
			expectError:    true,
			expectedErrMsg: "failed to create branch",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := CreateBranch(stubGetClientFn(client), NullTranslationHelperFunc)

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
			var returnedRef github.Reference
			err = json.Unmarshal([]byte(textContent.Text), &returnedRef)
			require.NoError(t, err)
			assert.Equal(t, *tc.expectedRef.Ref, *returnedRef.Ref)
			assert.Equal(t, *tc.expectedRef.Object.SHA, *returnedRef.Object.SHA)
		})
	}
}

func Test_GetCommit(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := GetCommit(stubGetClientFn(mockClient), NullTranslationHelperFunc)

	assert.Equal(t, "get_commit", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, ownerKey)
	assert.Contains(t, tool.InputSchema.Properties, repoKey)
	assert.Contains(t, tool.InputSchema.Properties, shaKey)
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{ownerKey, repoKey, shaKey})

	mockCommit := &github.RepositoryCommit{
		SHA: Ptr("abc123def456"),
		Commit: &github.Commit{
			Message: Ptr("First commit"),
			Author: &github.CommitAuthor{
				Name:  Ptr("Test User"),
				Email: Ptr("test@example.com"),
				Date:  Ptr(time.Now().Add(-48 * time.Hour)),
			},
		},
		Author: &github.User{
			Login: Ptr("testuser"),
		},
		HTMLURL: Ptr("https://github.com/owner/repo/commit/abc123def456"),
		Stats: &github.CommitStats{
			Additions: Ptr(10),
			Deletions: Ptr(2),
			Total:     Ptr(12),
		},
		Files: []*github.CommitFile{
			{
				Filename:  Ptr("file1.go"),
				Status:    Ptr("modified"),
				Additions: Ptr(10),
				Deletions: Ptr(2),
				Changes:   Ptr(12),
				Patch:     Ptr("@@ -1,2 +1,10 @@"),
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
					mockResponse(t, http.StatusOK, mockCommit),
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
			client := github.NewClient(tc.mockedClient)
			_, handler := GetCommit(stubGetClientFn(client), NullTranslationHelperFunc)

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
			var returnedCommit github.RepositoryCommit
			err = json.Unmarshal([]byte(textContent.Text), &returnedCommit)
			require.NoError(t, err)

			assert.Equal(t, *tc.expectedCommit.SHA, *returnedCommit.SHA)
			assert.Equal(t, *tc.expectedCommit.Commit.Message, *returnedCommit.Commit.Message)
			assert.Equal(t, *tc.expectedCommit.Author.Login, *returnedCommit.Author.Login)
			assert.Equal(t, *tc.expectedCommit.HTMLURL, *returnedCommit.HTMLURL)
		})
	}
}

func Test_ListCommits(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := ListCommits(stubGetClientFn(mockClient), NullTranslationHelperFunc)

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
			SHA: Ptr("abc123def456"),
			Commit: &github.Commit{
				Message: Ptr("First commit"),
				Author: &github.CommitAuthor{
					Name:  Ptr("Test User"),
					Email: Ptr("test@example.com"),
					Date:  Ptr(time.Now().Add(-48 * time.Hour)),
				},
			},
			Author: &github.User{
				Login: Ptr("testuser"),
			},
			HTMLURL: Ptr("https://github.com/owner/repo/commit/abc123def456"),
		},
		{
			SHA: Ptr("def456abc789"),
			Commit: &github.Commit{
				Message: Ptr("Second commit"),
				Author: &github.CommitAuthor{
					Name:  Ptr("Another User"),
					Email: Ptr("another@example.com"),
					Date:  Ptr(time.Now().Add(-24 * time.Hour)),
				},
			},
			Author: &github.User{
				Login: Ptr("anotheruser"),
			},
			HTMLURL: Ptr("https://github.com/owner/repo/commit/def456abc789"),
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
					expectQueryParams(t, map[string]string{
						"sha":      "main",
						"page":     "1",
						"per_page": "30",
					}).andThen(
						mockResponse(t, http.StatusOK, mockCommits),
					),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey: "owner",
				repoKey:  "repo",
				shaKey:   "main",
			},
			expectError:     false,
			expectedCommits: mockCommits,
		},
		{
			name: "successful commits fetch with pagination",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposCommitsByOwnerByRepo,
					expectQueryParams(t, map[string]string{
						"page":     "2",
						"per_page": "10",
					}).andThen(
						mockResponse(t, http.StatusOK, mockCommits),
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
			client := github.NewClient(tc.mockedClient)
			_, handler := ListCommits(stubGetClientFn(client), NullTranslationHelperFunc)

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
			var returnedCommits []*github.RepositoryCommit
			err = json.Unmarshal([]byte(textContent.Text), &returnedCommits)
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

func Test_CreateOrUpdateFile(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := CreateOrUpdateFile(stubGetClientFn(mockClient), NullTranslationHelperFunc)

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
			Name:        Ptr("example.md"),
			Path:        Ptr("docs/example.md"),
			SHA:         Ptr("abc123def456"),
			Size:        Ptr(42),
			HTMLURL:     Ptr("https://github.com/owner/repo/blob/main/docs/example.md"),
			DownloadURL: Ptr("https://raw.githubusercontent.com/owner/repo/main/docs/example.md"),
		},
		Commit: github.Commit{
			SHA:     Ptr("def456abc789"),
			Message: Ptr("Add example file"),
			Author: &github.CommitAuthor{
				Name:  Ptr("Test User"),
				Email: Ptr("test@example.com"),
				Date:  Ptr(time.Now()),
			},
			HTMLURL: Ptr("https://github.com/owner/repo/commit/def456abc789"),
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
					expectRequestBody(t, map[string]interface{}{
						commitMessage: "Add example file",
						contentKey:    "IyBFeGFtcGxlCgpUaGlzIGlzIGFuIGV4YW1wbGUgZmlsZS4=", // Base64 encoded content
						branchKey:     "main",
					}).andThen(
						mockResponse(t, http.StatusOK, mockFileResponse),
					),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:      "owner",
				repoKey:       "repo",
				pathKey:       "docs/example.md",
				contentKey:    "# Example\n\nThis is an example file.",
				commitMessage: "Add example file",
				branchKey:     "main",
			},
			expectError:     false,
			expectedContent: mockFileResponse,
		},
		{
			name: "successful file update with SHA",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PutReposContentsByOwnerByRepoByPath,
					expectRequestBody(t, map[string]interface{}{
						commitMessage: "Update example file",
						contentKey:    "IyBVcGRhdGVkIEV4YW1wbGUKClRoaXMgZmlsZSBoYXMgYmVlbiB1cGRhdGVkLg==", // Base64 encoded content
						branchKey:     "main",
						shaKey:        "abc123def456",
					}).andThen(
						mockResponse(t, http.StatusOK, mockFileResponse),
					),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:      "owner",
				repoKey:       "repo",
				pathKey:       "docs/example.md",
				contentKey:    "# Updated Example\n\nThis file has been updated.",
				commitMessage: "Update example file",
				branchKey:     "main",
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
				pathKey:       "docs/example.md",
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
			client := github.NewClient(tc.mockedClient)
			_, handler := CreateOrUpdateFile(stubGetClientFn(client), NullTranslationHelperFunc)

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
			var returnedContent github.RepositoryContentResponse
			err = json.Unmarshal([]byte(textContent.Text), &returnedContent)
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

func Test_CreateRepository(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := CreateRepository(stubGetClientFn(mockClient), NullTranslationHelperFunc)

	assert.Equal(t, "create_repository", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, nameKey)
	assert.Contains(t, tool.InputSchema.Properties, descriptionKey)
	assert.Contains(t, tool.InputSchema.Properties, privateKey)
	assert.Contains(t, tool.InputSchema.Properties, autoInitKey)
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{nameKey})

	// Setup mock repository response
	mockRepo := &github.Repository{
		Name:        Ptr("test-repo"),
		Description: Ptr("Test repository"),
		Private:     Ptr(true),
		HTMLURL:     Ptr("https://github.com/testuser/test-repo"),
		CloneURL:    Ptr("https://github.com/testuser/test-repo.git"),
		CreatedAt:   Ptr(time.Now()),
		Owner: &github.User{
			Login: Ptr("testuser"),
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
						Pattern: "/user/repos",
						Method:  "POST",
					},
					expectRequestBody(t, map[string]interface{}{
						nameKey:        "test-repo",
						descriptionKey: "Test repository",
						privateKey:     true,
						autoInitKey:    true,
					}).andThen(
						mockResponse(t, http.StatusCreated, mockRepo),
					),
				),
			),
			requestArgs: map[string]interface{}{
				nameKey:        "test-repo",
				descriptionKey: "Test repository",
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
						Pattern: "/user/repos",
						Method:  "POST",
					},
					expectRequestBody(t, map[string]interface{}{
						nameKey:        "test-repo",
						autoInitKey:    false,
						descriptionKey: "",
						privateKey:     false,
					}).andThen(
						mockResponse(t, http.StatusCreated, mockRepo),
					),
				),
			),
			requestArgs: map[string]interface{}{
				nameKey: "test-repo",
			},
			expectError:  false,
			expectedRepo: mockRepo,
		},
		{
			name: "repository creation fails",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.EndpointPattern{
						Pattern: "/user/repos",
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
			client := github.NewClient(tc.mockedClient)
			_, handler := CreateRepository(stubGetClientFn(client), NullTranslationHelperFunc)

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
			var returnedRepo github.Repository
			err = json.Unmarshal([]byte(textContent.Text), &returnedRepo)
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

func Test_PushFiles(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := PushFiles(stubGetClientFn(mockClient), NullTranslationHelperFunc)

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
		Ref: Ptr("refs/heads/main"),
		Object: &github.GitObject{
			SHA: Ptr("abc123"),
			URL: Ptr("https://api.github.com/repos/owner/repo/git/trees/abc123"),
		},
	}

	mockCommit := &github.Commit{
		SHA: Ptr("abc123"),
		Tree: &github.Tree{
			SHA: Ptr("def456"),
		},
	}

	mockTree := &github.Tree{
		SHA: Ptr("ghi789"),
	}

	mockNewCommit := &github.Commit{
		SHA:     Ptr("jkl012"),
		Message: Ptr("Update multiple files"),
		HTMLURL: Ptr("https://github.com/owner/repo/commit/jkl012"),
	}

	mockUpdatedRef := &github.Reference{
		Ref: Ptr("refs/heads/main"),
		Object: &github.GitObject{
			SHA: Ptr("jkl012"),
			URL: Ptr("https://api.github.com/repos/owner/repo/git/trees/jkl012"),
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
					expectRequestBody(t, map[string]interface{}{
						"base_tree": "def456",
						"tree": []interface{}{
							map[string]interface{}{
								"path":    "README.md",
								"mode":    "100644",
								"type":    "blob",
								"content": "# Updated README\n\nThis is an updated README file.",
							},
							map[string]interface{}{
								"path":    "docs/example.md",
								"mode":    "100644",
								"type":    "blob",
								"content": "# Example\n\nThis is an example file.",
							},
						},
					}).andThen(
						mockResponse(t, http.StatusCreated, mockTree),
					),
				),
				// Create commit
				mock.WithRequestMatchHandler(
					mock.PostReposGitCommitsByOwnerByRepo,
					expectRequestBody(t, map[string]interface{}{
						commitMessage: "Update multiple files",
						"tree":        "ghi789",
						"parents":     []interface{}{"abc123"},
					}).andThen(
						mockResponse(t, http.StatusCreated, mockNewCommit),
					),
				),
				// Update reference
				mock.WithRequestMatchHandler(
					mock.PatchReposGitRefsByOwnerByRepoByRef,
					expectRequestBody(t, map[string]interface{}{
						"sha":   "jkl012",
						"force": false,
					}).andThen(
						mockResponse(t, http.StatusOK, mockUpdatedRef),
					),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:  "owner",
				repoKey:   "repo",
				branchKey: "main",
				filesKey: []interface{}{
					map[string]interface{}{
						"path":     "README.md",
						contentKey: "# Updated README\n\nThis is an updated README file.",
					},
					map[string]interface{}{
						"path":     "docs/example.md",
						contentKey: "# Example\n\nThis is an example file.",
					},
				},
				commitMessage: "Update multiple files",
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
				branchKey:     "main",
				filesKey:      "invalid-files-parameter", // Not an array
				commitMessage: "Update multiple files",
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
				branchKey: "main",
				filesKey: []interface{}{
					map[string]interface{}{
						contentKey: "# Missing path",
					},
				},
				commitMessage: "Update file",
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
				branchKey: "main",
				filesKey: []interface{}{
					map[string]interface{}{
						"path": "README.md",
						// Missing content
					},
				},
				commitMessage: "Update file",
			},
			expectError:    false, // This returns a tool error, not a Go error
			expectedErrMsg: "each file must have content",
		},
		{
			name: "fails to get branch reference",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposGitRefByOwnerByRepoByRef,
					mockResponse(t, http.StatusNotFound, nil),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:  "owner",
				repoKey:   "repo",
				branchKey: "non-existent-branch",
				filesKey: []interface{}{
					map[string]interface{}{
						"path":     "README.md",
						contentKey: "# README",
					},
				},
				commitMessage: "Update file",
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
					mockResponse(t, http.StatusNotFound, nil),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:  "owner",
				repoKey:   "repo",
				branchKey: "main",
				filesKey: []interface{}{
					map[string]interface{}{
						"path":     "README.md",
						contentKey: "# README",
					},
				},
				commitMessage: "Update file",
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
					mockResponse(t, http.StatusInternalServerError, nil),
				),
			),
			requestArgs: map[string]interface{}{
				ownerKey:  "owner",
				repoKey:   "repo",
				branchKey: "main",
				filesKey: []interface{}{
					map[string]interface{}{
						"path":     "README.md",
						contentKey: "# README",
					},
				},
				commitMessage: "Update file",
			},
			expectError:    true,
			expectedErrMsg: "failed to create tree",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup client with mock
			client := github.NewClient(tc.mockedClient)
			_, handler := PushFiles(stubGetClientFn(client), NullTranslationHelperFunc)

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

			// Parse the result and get the text content if no error
			textContent := getTextResult(t, result)

			// Unmarshal and verify the result
			var returnedRef github.Reference
			err = json.Unmarshal([]byte(textContent.Text), &returnedRef)
			require.NoError(t, err)

			assert.Equal(t, *tc.expectedRef.Ref, *returnedRef.Ref)
			assert.Equal(t, *tc.expectedRef.Object.SHA, *returnedRef.Object.SHA)
		})
	}
}

func Test_ListBranches(t *testing.T) {
	// Verify tool definition once
	mockClient := github.NewClient(nil)
	tool, _ := ListBranches(stubGetClientFn(mockClient), NullTranslationHelperFunc)

	assert.Equal(t, "list_branches", tool.Name)
	assert.NotEmpty(t, tool.Description)
	assert.Contains(t, tool.InputSchema.Properties, ownerKey)
	assert.Contains(t, tool.InputSchema.Properties, repoKey)
	assert.Contains(t, tool.InputSchema.Properties, pageKey)
	assert.Contains(t, tool.InputSchema.Properties, perPageKey)
	assert.ElementsMatch(t, tool.InputSchema.Required, []string{ownerKey, repoKey})

	// Setup mock branches for success case
	mockBranches := []*github.Branch{
		{
			Name:   Ptr("main"),
			Commit: &github.RepositoryCommit{SHA: Ptr("abc123")},
		},
		{
			Name:   Ptr("develop"),
			Commit: &github.RepositoryCommit{SHA: Ptr("def456")},
		},
	}

	// Test cases
	tests := []struct {
		name          string
		args          map[string]interface{}
		mockResponses []mock.MockBackendOption
		wantErr       bool
		errContains   string
	}{
		{
			name: "success",
			args: map[string]interface{}{
				ownerKey: "owner",
				repoKey:  "repo",
				pageKey:  float64(2),
			},
			mockResponses: []mock.MockBackendOption{
				mock.WithRequestMatch(
					mock.GetReposBranchesByOwnerByRepo,
					mockBranches,
				),
			},
			wantErr: false,
		},
		{
			name: "missing owner",
			args: map[string]interface{}{
				repoKey: "repo",
			},
			mockResponses: []mock.MockBackendOption{},
			wantErr:       false,
			errContains:   "missing required parameter: owner",
		},
		{
			name: "missing repo",
			args: map[string]interface{}{
				ownerKey: "owner",
			},
			mockResponses: []mock.MockBackendOption{},
			wantErr:       false,
			errContains:   "missing required parameter: repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := github.NewClient(mock.NewMockedHTTPClient(tt.mockResponses...))
			_, handler := ListBranches(stubGetClientFn(mockClient), NullTranslationHelperFunc)

			// Create request
			request := createMCPRequest(tt.args)

			// Call handler
			result, err := handler(context.Background(), request)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			if tt.errContains != "" {
				textContent := getTextResult(t, result)
				assert.Contains(t, textContent.Text, tt.errContains)
				return
			}

			textContent := getTextResult(t, result)
			require.NotEmpty(t, textContent.Text)

			// Verify response
			var branches []*github.Branch
			err = json.Unmarshal([]byte(textContent.Text), &branches)
			require.NoError(t, err)
			assert.Len(t, branches, 2)
			assert.Equal(t, "main", *branches[0].Name)
			assert.Equal(t, "develop", *branches[1].Name)
		})
	}
}
