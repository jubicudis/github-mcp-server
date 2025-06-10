// WHO: GitHubMCPServer
// WHAT: GitHub API Server Implementation
// WHEN: During request processing
// WHERE: MCP Bridge Layer
// WHY: To handle GitHub API integration
// HOW: Using MCP protocol handlers
// EXTENT: All GitHub API operations
package ghmcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jubicudis/github-mcp-server/pkg/common"
	"github.com/jubicudis/github-mcp-server/pkg/translations"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// WHO: ConstantsManager
// WHAT: String Constants Definition
// WHEN: Compilation time
// WHERE: GitHub MCP Server
// WHY: To avoid string duplication
// HOW: By defining reusable constants
// EXTENT: All shared string literals
const (
	// Parameter descriptions
	RepositoryOwnerDesc = "Repository owner"
	RepositoryNameDesc  = "Repository name"

	// Error messages
	ErrGetGitHubClient  = "failed to get GitHub client: %w"
	ErrReadResponseBody = "failed to read response body: %w"
	ErrMarshalIssue     = "failed to marshal issue: %w"
	ErrMarshalComment   = "failed to marshal comment: %w"
	ErrMarshalPR        = "failed to marshal pull request: %w"
	ErrMarshalSearchRes = "failed to marshal search results: %w"
	ErrMarshalUser      = "failed to marshal user: %w"
	ErrMarshalComments  = "failed to marshal comments: %w"
	ErrMarshalIssues    = "failed to marshal issues: %w"
)

// GetClientFn is defined in common.go
// TranslationHelperFunc from translations package is used for string translations

// WHO: ServerInitializer
// WHAT: MCP Server Configuration
// WHEN: System startup
// WHERE: GitHub Bridge
// WHY: To configure all available GitHub MCP tools
// HOW: By registering resource templates and tools
// EXTENT: All GitHub API functionalities
func NewServer(getClient common.GetClientFn, version string, readOnly bool, t translations.TranslationHelperFunc, opts ...server.ServerOption) *server.MCPServer {
	// Add default options
	defaultOpts := []server.ServerOption{
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	}
	opts = append(defaultOpts, opts...)

	// Create a new MCP server
	s := server.NewMCPServer(
		"github-mcp-server",
		version,
		opts...,
	)

	// Add GitHub Resources
	registerResourceTemplates(s, getClient, t)

	// Register all tools by category
	registerIssueTools(s, getClient, t, readOnly)
	registerPRTools(s, getClient, t, readOnly)
	registerRepositoryTools(s, getClient, t, readOnly)
	registerSearchTools(s, getClient, t)
	registerUserTools(s, getClient, t)
	registerCodeScanningTools(s, getClient, t)

	return s
}

// WHO: ResourceRegistrar
// WHAT: Resource Template Registration
// WHEN: Server initialization
// WHERE: GitHub MCP Server
// WHY: To provide access to GitHub resources
// HOW: By defining resource templates for different access patterns
// EXTENT: All GitHub resource types
func registerResourceTemplates(s *server.MCPServer, getClient common.GetClientFn, t translations.TranslationHelperFunc) {
	s.AddResourceTemplate(GetRepositoryResourceContent(getClient, t))
	// TODO: Implement resource templates
	// // TODO: Implement resource templates
	// s.AddResourceTemplate(GetRepositoryResourceBranchContent(getClient, t))
	// // s.AddResourceTemplate(GetRepositoryResourceCommitContent(getClient, t))
	// // s.AddResourceTemplate(GetRepositoryResourceTagContent(getClient, t))
	// // s.AddResourceTemplate(GetRepositoryResourcePrContent(getClient, t))
}

// WHO: IssueToolRegistrar
// WHAT: Issue Tool Registration
// WHEN: Server initialization
// WHERE: GitHub MCP Server
// WHY: To provide GitHub issue operations
// HOW: By registering issue-related tools
// EXTENT: All issue operations
func registerIssueTools(s *server.MCPServer, getClient common.GetClientFn, t translations.TranslationHelperFunc, readOnly bool) {
	// Import GetIssue from issues.go instead of using the local function
	s.AddTool(GetIssue(getClient, t))
	s.AddTool(SearchIssues(getClient, t))
	s.AddTool(ListIssues(getClient, t))
	s.AddTool(GetIssueComments(getClient, t))

	if !readOnly {
		s.AddTool(CreateIssue(getClient, t))
		// // s.AddTool(AddIssueComment(getClient, t))
		s.AddTool(UpdateIssue(getClient, t))
	}
}

// WHO: PRToolRegistrar
// WHAT: Pull Request Tool Registration
// WHEN: Server initialization
// WHERE: GitHub MCP Server
// WHY: To provide GitHub PR operations
// HOW: By registering PR-related tools
// EXTENT: All pull request operations
func registerPRTools(s *server.MCPServer, getClient common.GetClientFn, t translations.TranslationHelperFunc, readOnly bool) {
	s.AddTool(GetPullRequest(getClient, t))
	s.AddTool(ListPullRequests(getClient, t))
	s.AddTool(GetPullRequestFiles(getClient, t))
	s.AddTool(GetPullRequestStatus(getClient, t))
	s.AddTool(GetPullRequestComments(getClient, t))
	s.AddTool(GetPullRequestReviews(getClient, t))

	if !readOnly {
		s.AddTool(MergePullRequest(getClient, t))
		s.AddTool(UpdatePullRequestBranch(getClient, t))
		s.AddTool(CreatePullRequestReview(getClient, t))
		s.AddTool(CreatePullRequest(getClient, t))
		s.AddTool(UpdatePullRequest(getClient, t))
	}
}

// WHO: RepositoryToolRegistrar
// WHAT: Repository Tool Registration
// WHEN: Server initialization
// WHERE: GitHub MCP Server
// WHY: To provide GitHub repository operations
// HOW: By registering repository-related tools
// EXTENT: All repository operations
func registerRepositoryTools(s *server.MCPServer, getClient common.GetClientFn, t translations.TranslationHelperFunc, readOnly bool) {
	s.AddTool(SearchRepositories(getClient, t))
	s.AddTool(GetFileContents(getClient, t))
	s.AddTool(GetCommit(getClient, t))
	s.AddTool(ListCommits(getClient, t))
	s.AddTool(ListBranches(getClient, t))

	if !readOnly {
		s.AddTool(CreateOrUpdateFile(getClient, t))
		s.AddTool(CreateRepository(getClient, t))
		s.AddTool(ForkRepository(getClient, t))
		s.AddTool(CreateBranch(getClient, t))
		s.AddTool(PushFiles(getClient, t))
	}
}

// WHO: SearchToolRegistrar
// WHAT: Search Tool Registration
// WHEN: Server initialization
// WHERE: GitHub MCP Server
// WHY: To provide GitHub search operations
// HOW: By registering search-related tools
// EXTENT: All search operations
func registerSearchTools(s *server.MCPServer, getClient common.GetClientFn, t translations.TranslationHelperFunc) {
	s.AddTool(SearchCode(getClient, t))
	s.AddTool(SearchUsers(getClient, t))
}

// WHO: UserToolRegistrar
// WHAT: User Tool Registration
// WHEN: Server initialization
// WHERE: GitHub MCP Server
// WHY: To provide GitHub user operations
// HOW: By registering user-related tools
// EXTENT: All user operations
func registerUserTools(s *server.MCPServer, getClient common.GetClientFn, t translations.TranslationHelperFunc) {
	s.AddTool(GetMe(getClient, t))
}

// WHO: CodeScanningToolRegistrar
// WHAT: Code Scanning Tool Registration
// WHEN: Server initialization
// WHERE: GitHub MCP Server
// WHY: To provide GitHub code scanning operations
// HOW: By registering code scanning-related tools
// EXTENT: All code scanning operations
func registerCodeScanningTools(s *server.MCPServer, getClient common.GetClientFn, t translations.TranslationHelperFunc) {
	s.AddTool(GetCodeScanningAlert(getClient, t))
	s.AddTool(ListCodeScanningAlerts(getClient, t))
}

// WHO: UserInfoTool
// WHAT: Authenticated User Tool
// WHEN: During user info request
// WHERE: GitHub MCP Server
// WHY: To provide user identity information
// HOW: By fetching authenticated user details from GitHub
// EXTENT: User identification
func GetMe(getClient common.GetClientFn, t translations.TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_me",
			mcp.WithDescription("Get details of the authenticated GitHub user. Use this when a request include \"me\", \"my\"..."),
			mcp.WithString("reason",
				mcp.Description("Optional: reason the session was created"),
			),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf(ErrGetGitHubClient, err)
			}
			user, resp, err := client.Users.Get(ctx, "")
			if err != nil {
				return nil, fmt.Errorf("failed to get user: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf(ErrReadResponseBody, err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to get user: %s", string(body))), nil
			}

			r, err := json.Marshal(user)
			if err != nil {
				return nil, fmt.Errorf(ErrMarshalUser, err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// Canonical server logic for GitHub MCP server
// Remove all stubs, placeholders, and incomplete logic
// All types and methods must be robust, DRY, and reference only canonical helpers from /pkg/common
// All server and event logic must be fully implemented

// The following functions have been moved to common.go to avoid redeclaration:
// - OptionalParamOK
// - isAcceptedError
// - requiredParam
// - RequiredInt
// - OptionalParam
// - OptionalIntParam
// - OptionalIntParamWithDefault
// - OptionalStringArrayParam
// - WithPagination
// - PaginationParams (type definition)
// - OptionalPaginationParams
// - extractPRParams
// - prepareClientAndPRParams

// The following PR-related functions have been moved to pullrequests.go:
// - GetPullRequest
// - ListPullRequests
// - GetPullRequestFiles
// - GetPullRequestStatus
// - GetPullRequestComments
// - GetPullRequestReviews
// - MergePullRequest
// - CreatePullRequestReview
// - CreatePullRequest
// - UpdatePullRequest
// - UpdatePullRequestBranch
