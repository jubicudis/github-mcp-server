// WHO: GitHubMCPRepositoryResourceTests
// WHAT: Repository Resource API Testing
// WHEN: During test execution
// WHERE: MCP Bridge Layer Testing
// WHY: To verify repository file operations
// HOW: By testing MCP protocol handlers
// EXTENT: All repository resource operations
package github

import (
	"context"
	"net/http"
	"testing"

	"tranquility-neuro-os/github-mcp-server/pkg/github/testutil"
	"tranquility-neuro-os/github-mcp-server/pkg/log"

	"github.com/google/go-github/v49/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/require"
)

// Test constants for repeated string literals
const (
	readmeFileName  = "README.md"
	ownerLogin      = "owner"
	repoName        = "repo"
	mainBranch      = "main"
	dataPngFileName = "data.png"
	htmlUrlReadme   = "https://github.com/owner/repo/blob/main/README.md"
	htmlUrlDataPng  = "https://github.com/owner/repo/blob/main/data.png"
	dirHtmlUrl      = "https://github.com/owner/repo/tree/main/src"
	testRepoContent = "# Test Repository\n\nThis is a test repository."
)

func Ptr[T any](v T) *T { return &v }

var GetRawReposContentsByOwnerByRepoByPath mock.EndpointPattern = mock.EndpointPattern{
	Pattern: "/{owner}/{repo}/main/{path:.+}",
	Method:  "GET",
}

func TestRepositoryResourceContentsHandler(t *testing.T) {
	logger := log.NewLogger().WithLevel(log.LevelDebug)

	mockDirContent := []*github.RepositoryContent{
		{
			Type:        Ptr("file"),
			Name:        Ptr(readmeFileName),
			Path:        Ptr(readmeFileName),
			SHA:         Ptr("abc123"),
			Size:        Ptr(42),
			HTMLURL:     Ptr(htmlUrlReadme),
			DownloadURL: Ptr("https://raw.githubusercontent.com/owner/repo/main/README.md"),
		},
		{
			Type:        Ptr("dir"),
			Name:        Ptr("src"),
			Path:        Ptr("src"),
			SHA:         Ptr("def456"),
			HTMLURL:     Ptr(dirHtmlUrl),
			DownloadURL: Ptr("https://raw.githubusercontent.com/owner/repo/main/src"),
		},
	}
	expectedDirContent := []mcp.TextResourceContents{
		{
			URI:      htmlUrlReadme,
			MIMEType: "text/markdown",
			Text:     readmeFileName,
		},
		{
			URI:      dirHtmlUrl,
			MIMEType: "text/directory",
			Text:     "src",
		},
	}

	mockTextContent := &github.RepositoryContent{
		Type:        Ptr("file"),
		Name:        Ptr(readmeFileName),
		Path:        Ptr(readmeFileName),
		Content:     Ptr(testRepoContent),
		SHA:         Ptr("abc123"),
		Size:        Ptr(42),
		HTMLURL:     Ptr(htmlUrlReadme),
		DownloadURL: Ptr("https://raw.githubusercontent.com/owner/repo/main/README.md"),
	}

	mockFileContent := &github.RepositoryContent{
		Type:        Ptr("file"),
		Name:        Ptr(dataPngFileName),
		Path:        Ptr(dataPngFileName),
		Content:     Ptr("IyBUZXN0IFJlcG9zaXRvcnkKClRoaXMgaXMgYSB0ZXN0IHJlcG9zaXRvcnku"), // Base64 encoded "# Test Repository\n\nThis is a test repository."
		SHA:         Ptr("abc123"),
		Size:        Ptr(42),
		HTMLURL:     Ptr(htmlUrlDataPng),
		DownloadURL: Ptr("https://raw.githubusercontent.com/owner/repo/main/data.png"),
	}

	expectedFileContent := []mcp.BlobResourceContents{
		{
			Blob:     "IyBUZXN0IFJlcG9zaXRvcnkKClRoaXMgaXMgYSB0ZXN0IHJlcG9zaXRvcnku",
			MIMEType: "image/png",
			URI:      "",
		},
	}

	expectedTextContent := []mcp.TextResourceContents{
		{
			Text:     testRepoContent,
			MIMEType: "text/markdown",
			URI:      "",
		},
	}

	tests := []struct {
		name           string
		mockedClient   *http.Client
		requestArgs    map[string]any
		expectError    string
		expectedResult any
		expectedErrMsg string
	}{
		{
			name: "missing owner",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposContentsByOwnerByRepoByPath,
					mockFileContent,
				),
			),
			requestArgs: map[string]any{},
			expectError: "owner is required",
		},
		{
			name: "missing repo",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposContentsByOwnerByRepoByPath,
					mockFileContent,
				),
			),
			requestArgs: map[string]any{
				"owner": []string{ownerLogin},
			},
			expectError: "repo is required",
		},
		{
			name: "successful blob content fetch",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposContentsByOwnerByRepoByPath,
					mockFileContent,
				),
				mock.WithRequestMatchHandler(
					GetRawReposContentsByOwnerByRepoByPath,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.Header().Set("Content-Type", "image/png")
						// as this is given as a png, it will return the content as a blob
						_, err := w.Write([]byte(testRepoContent))
						require.NoError(t, err)
					}),
				),
			),
			requestArgs: map[string]any{
				"owner":  []string{ownerLogin},
				"repo":   []string{repoName},
				"path":   []string{dataPngFileName},
				"branch": []string{mainBranch},
			},
			expectedResult: expectedFileContent,
		},
		{
			name: "successful text content fetch",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposContentsByOwnerByRepoByPath,
					mockTextContent,
				),
				mock.WithRequestMatch(
					GetRawReposContentsByOwnerByRepoByPath,
					[]byte(testRepoContent),
				),
			),
			requestArgs: map[string]any{
				"owner":  []string{ownerLogin},
				"repo":   []string{repoName},
				"path":   []string{readmeFileName},
				"branch": []string{mainBranch},
			},
			expectedResult: expectedTextContent,
		},
		{
			name: "successful directory content fetch",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposContentsByOwnerByRepoByPath,
					mockDirContent,
				),
			),
			requestArgs: map[string]any{
				"owner": []string{ownerLogin},
				"repo":  []string{repoName},
				"path":  []string{"src"},
			},
			expectedResult: expectedDirContent,
		},
		{
			name: "no data",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposContentsByOwnerByRepoByPath,
				),
			),
			requestArgs: map[string]any{
				"owner": []string{ownerLogin},
				"repo":  []string{repoName},
				"path":  []string{"src"},
			},
			expectedResult: nil,
			expectError:    "no repository resource content found",
		},
		{
			name: "empty data",
			mockedClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.GetReposContentsByOwnerByRepoByPath,
					[]*github.RepositoryContent{},
				),
			),
			requestArgs: map[string]any{
				"owner": []string{ownerLogin},
				"repo":  []string{repoName},
				"path":  []string{"src"},
			},
			expectedResult: nil,
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
			requestArgs: map[string]any{
				"owner":  []string{ownerLogin},
				"repo":   []string{repoName},
				"path":   []string{"nonexistent.md"},
				"branch": []string{mainBranch},
			},
			expectError: "404 Not Found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := github.NewClient(tc.mockedClient)
			handler := RepositoryResourceContentsHandler((testutil.StubGetClientFn(client)))

			request := mcp.ReadResourceRequest{
				Params: struct {
					URI       string         `json:"uri"`
					Arguments map[string]any `json:"arguments,omitempty"`
				}{
					Arguments: tc.requestArgs,
				},
			}

			resp, err := handler(context.TODO(), request)

			if tc.expectError != "" {
				require.ErrorContains(t, err, tc.expectedErrMsg)
				return
			}

			require.NoError(t, err)
			require.ElementsMatch(t, resp, tc.expectedResult)
		})
	}
}

func TestGetRepositoryResourceContent(t *testing.T) {
	tmpl, _ := GetRepositoryResourceContent(nil, testutil.NullTranslationHelperFunc)
	require.Equal(t, "repo://{owner}/{repo}/contents{/path*}", tmpl.URITemplate.Raw())
}

func TestGetRepositoryResourceBranchContent(t *testing.T) {
	tmpl, _ := GetRepositoryResourceBranchContent(nil, testutil.NullTranslationHelperFunc)
	require.Equal(t, "repo://{owner}/{repo}/refs/heads/{branch}/contents{/path*}", tmpl.URITemplate.Raw())
}

func TestGetRepositoryResourceCommitContent(t *testing.T) {
	tmpl, _ := GetRepositoryResourceCommitContent(nil, testutil.NullTranslationHelperFunc)
	require.Equal(t, "repo://{owner}/{repo}/sha/{sha}/contents{/path*}", tmpl.URITemplate.Raw())
}

func TestGetRepositoryResourceTagContent(t *testing.T) {
	tmpl, _ := GetRepositoryResourceTagContent(nil, testutil.NullTranslationHelperFunc)
	require.Equal(t, "repo://{owner}/{repo}/refs/tags/{tag}/contents{/path*}", tmpl.URITemplate.Raw())
}

func TestGetRepositoryResourcePrContent(t *testing.T) {
	tmpl, _ := GetRepositoryResourcePrContent(nil, testutil.NullTranslationHelperFunc)
	require.Equal(t, "repo://{owner}/{repo}/refs/pull/{prNumber}/head/contents{/path*}", tmpl.URITemplate.Raw())
}
