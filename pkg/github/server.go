// WHO: GitHubMCPServer
// WHAT: GitHub API Server Implementation
// WHEN: During request processing
// WHERE: MCP Bridge Layer
// WHY: To handle GitHub API integration
// HOW: Using MCP protocol handlers
// EXTENT: All GitHub API operations
package githubapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/go-github/v69/github"
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

// TranslationHelperFunc defines a function type for translations
type TranslationHelperFunc func(key string, defaultValue string) string

// GetClientFn defines a function type to obtain GitHub clients
type GetClientFn func(context.Context) (*github.Client, error)

// WHO: ServerInitializer
// WHAT: MCP Server Configuration
// WHEN: System startup
// WHERE: GitHub Bridge
// WHY: To configure all available GitHub MCP tools
// HOW: By registering resource templates and tools
// EXTENT: All GitHub API functionality
func NewServer(getClient GetClientFn, version string, readOnly bool, t TranslationHelperFunc, opts ...server.ServerOption) *server.MCPServer {
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
func registerResourceTemplates(s *server.MCPServer, getClient GetClientFn, t TranslationHelperFunc) {
	s.AddResourceTemplate(GetRepositoryResourceContent(getClient, t))
	s.AddResourceTemplate(GetRepositoryResourceBranchContent(getClient, t))
	s.AddResourceTemplate(GetRepositoryResourceCommitContent(getClient, t))
	s.AddResourceTemplate(GetRepositoryResourceTagContent(getClient, t))
	s.AddResourceTemplate(GetRepositoryResourcePrContent(getClient, t))
}

// WHO: IssueToolRegistrar
// WHAT: Issue Tool Registration
// WHEN: Server initialization
// WHERE: GitHub MCP Server
// WHY: To provide GitHub issue operations
// HOW: By registering issue-related tools
// EXTENT: All issue operations
func registerIssueTools(s *server.MCPServer, getClient GetClientFn, t TranslationHelperFunc, readOnly bool) {
	s.AddTool(GetIssue(getClient, t))
	s.AddTool(SearchIssues(getClient, t))
	s.AddTool(ListIssues(getClient, t))
	s.AddTool(GetIssueComments(getClient, t))

	if !readOnly {
		s.AddTool(CreateIssue(getClient, t))
		s.AddTool(AddIssueComment(getClient, t))
		s.AddTool(UpdateIssue(getClient, t))
	}
}

// WHO: IssueOperationTool
// WHAT: Single Issue Retrieval Tool
// WHEN: During tool invocation
// WHERE: GitHub MCP Server
// WHY: To provide access to a specific issue
// HOW: By fetching issue details from GitHub
// EXTENT: Single issue access
func GetIssue(getClient GetClientFn, t TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_issue",
			mcp.WithDescription(t("TOOL_GET_ISSUE_DESCRIPTION", "Get details of a specific GitHub issue")),
			mcp.WithString("owner",
				mcp.Description(RepositoryOwnerDesc),
				mcp.Required(),
			),
			mcp.WithString("repo",
				mcp.Description(RepositoryNameDesc),
				mcp.Required(),
			),
			mcp.WithNumber("number",
				mcp.Description("Issue number"),
				mcp.Required(),
			),
		),
		func(ctx context.Context, r mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf(ErrGetGitHubClient, err)
			}

			// Extract required parameters
			owner, err := requiredParam[string](r, "owner")
			if err != nil {
				return nil, err
			}

			repo, err := requiredParam[string](r, "repo")
			if err != nil {
				return nil, err
			}

			num, err := RequiredInt(r, "number")
			if err != nil {
				return nil, err
			}

			// Call GitHub API
			issue, resp, err := client.Issues.Get(ctx, owner, repo, num)
			if err != nil {
				return nil, fmt.Errorf("failed to get issue: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf(ErrReadResponseBody, err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to get issue: %s", string(body))), nil
			}

			result, err := json.Marshal(issue)
			if err != nil {
				return nil, fmt.Errorf(ErrMarshalIssue, err)
			}

			return mcp.NewToolResultText(string(result)), nil
		}
}

// WHO: IssueCommentsTool
// WHAT: Issue Comments Retrieval Tool
// WHEN: During tool invocation
// WHERE: GitHub MCP Server
// WHY: To access comments on a GitHub issue
// HOW: By querying GitHub API with pagination support
// EXTENT: All comments on a specific issue
func GetIssueComments(getClient GetClientFn, t TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_issue_comments",
			mcp.WithDescription(t("TOOL_GET_ISSUE_COMMENTS_DESCRIPTION", "Get comments for a GitHub issue")),
			mcp.WithString("owner",
				mcp.Description(RepositoryOwnerDesc),
				mcp.Required(),
			),
			mcp.WithString("repo",
				mcp.Description(RepositoryNameDesc),
				mcp.Required(),
			),
			mcp.WithNumber("issue_number",
				mcp.Description("Issue number"),
				mcp.Required(),
			),
			WithPagination(),
		),
		func(ctx context.Context, r mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf(ErrGetGitHubClient, err)
			}

			// Extract required parameters
			owner, err := requiredParam[string](r, "owner")
			if err != nil {
				return nil, err
			}

			repo, err := requiredParam[string](r, "repo")
			if err != nil {
				return nil, err
			}

			issueNumber, err := RequiredInt(r, "issue_number")
			if err != nil {
				return nil, err
			}

			// Get pagination parameters
			pagination, err := OptionalPaginationParams(r)
			if err != nil {
				return nil, err
			}

			// Create list options
			opts := &github.IssueListCommentsOptions{
				ListOptions: github.ListOptions{
					Page:    pagination.page,
					PerPage: pagination.perPage,
				},
			}

			// Call GitHub API
			comments, resp, err := client.Issues.ListComments(ctx, owner, repo, issueNumber, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to list issue comments: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf(ErrReadResponseBody, err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to list issue comments: %s", string(body))), nil
			}

			result, err := json.Marshal(comments)
			if err != nil {
				return nil, fmt.Errorf(ErrMarshalComments, err)
			}

			return mcp.NewToolResultText(string(result)), nil
		}
}

// WHO: IssueSearchTool
// WHAT: Issue Search Tool
// WHEN: During tool invocation
// WHERE: GitHub MCP Server
// WHY: To allow searching for issues across repositories
// HOW: By querying GitHub search API with filters
// EXTENT: Cross-repository issue searching
func SearchIssues(getClient GetClientFn, t TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("search_issues",
			mcp.WithDescription(t("TOOL_SEARCH_ISSUES_DESCRIPTION", "Search for issues and pull requests across GitHub repositories")),
			mcp.WithString("q",
				mcp.Description("Search query using GitHub issues search syntax"),
				mcp.Required(),
			),
			mcp.WithString("sort",
				mcp.Description("Sort field (comments, reactions, created, etc.)"),
				mcp.Enum("comments", "reactions", "reactions-+1", "reactions--1", "reactions-smile",
					"reactions-thinking_face", "reactions-heart", "reactions-tada", "interactions",
					"created", "updated"),
			),
			mcp.WithString("order",
				mcp.Description("Sort order ('asc' or 'desc')"),
				mcp.Enum("asc", "desc"),
			),
			WithPagination(),
		),
		func(ctx context.Context, r mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf(ErrGetGitHubClient, err)
			}

			// Extract required parameters
			query, err := requiredParam[string](r, "q")
			if err != nil {
				return nil, err
			}

			// Get optional parameters
			sort, _ := OptionalParam[string](r, "sort")
			order, _ := OptionalParam[string](r, "order")

			// Get pagination parameters
			pagination, err := OptionalPaginationParams(r)
			if err != nil {
				return nil, err
			}

			// Create search options
			opts := &github.SearchOptions{
				Sort:  sort,
				Order: order,
				ListOptions: github.ListOptions{
					Page:    pagination.page,
					PerPage: pagination.perPage,
				},
			}

			// Call GitHub API
			issues, resp, err := client.Search.Issues(ctx, query, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to search issues: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf(ErrReadResponseBody, err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to search issues: %s", string(body))), nil
			}

			result, err := json.Marshal(issues)
			if err != nil {
				return nil, fmt.Errorf(ErrMarshalSearchRes, err)
			}

			return mcp.NewToolResultText(string(result)), nil
		}
}

// WHO: ListIssuesTool
// WHAT: Issue Listing Tool
// WHEN: During tool invocation
// WHERE: GitHub MCP Server
// WHY: To list repository issues with filtering options
// HOW: By querying GitHub Issues API with parameters
// EXTENT: All repository issues
func ListIssues(getClient GetClientFn, t TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("list_issues",
			mcp.WithDescription(t("TOOL_LIST_ISSUES_DESCRIPTION", "List issues in a GitHub repository with filtering options")),
			mcp.WithString("owner",
				mcp.Description(RepositoryOwnerDesc),
				mcp.Required(),
			),
			mcp.WithString("repo",
				mcp.Description(RepositoryNameDesc),
				mcp.Required(),
			),
			mcp.WithString("state",
				mcp.Description("Filter by state ('open', 'closed', 'all')"),
				mcp.Enum("open", "closed", "all"),
			),
			mcp.WithString("sort",
				mcp.Description("Sort by ('created', 'updated', 'comments')"),
				mcp.Enum("created", "updated", "comments"),
			),
			mcp.WithString("direction",
				mcp.Description("Sort direction ('asc', 'desc')"),
				mcp.Enum("asc", "desc"),
			),
			mcp.WithString("since",
				mcp.Description("Filter by date (ISO 8601 timestamp)"),
			),
			mcp.WithArray("labels",
				mcp.Description("Filter by labels"),
				mcp.Items(mcp.String()),
			),
			WithPagination(),
		),
		func(ctx context.Context, r mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf(ErrGetGitHubClient, err)
			}

			// Extract required parameters
			owner, err := requiredParam[string](r, "owner")
			if err != nil {
				return nil, err
			}

			repo, err := requiredParam[string](r, "repo")
			if err != nil {
				return nil, err
			}

			// Extract optional parameters
			state, _ := OptionalParam[string](r, "state")
			if state == "" {
				state = "open"
			}

			sort, _ := OptionalParam[string](r, "sort")
			direction, _ := OptionalParam[string](r, "direction")
			since, _ := OptionalParam[string](r, "since")

			// Get labels
			labels, err := OptionalStringArrayParam(r, "labels")
			if err != nil {
				return nil, err
			}

			// Get pagination parameters
			pagination, err := OptionalPaginationParams(r)
			if err != nil {
				return nil, err
			}

			// Create list options
			opts := &github.IssueListByRepoOptions{
				State:     state,
				Sort:      sort,
				Direction: direction,
				Since:     github.Timestamp{},
				Labels:    labels,
				ListOptions: github.ListOptions{
					Page:    pagination.page,
					PerPage: pagination.perPage,
				},
			}

			// Set timestamp if provided
			if since != "" {
				t, err := time.Parse(time.RFC3339, since)
				if err != nil {
					return nil, fmt.Errorf("invalid since timestamp format (expected RFC3339): %w", err)
				}
				opts.Since = github.Timestamp{Time: t}
			}

			// Call GitHub API
			issues, resp, err := client.Issues.ListByRepo(ctx, owner, repo, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to list issues: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf(ErrReadResponseBody, err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to list issues: %s", string(body))), nil
			}

			result, err := json.Marshal(issues)
			if err != nil {
				return nil, fmt.Errorf(ErrMarshalIssues, err)
			}

			return mcp.NewToolResultText(string(result)), nil
		}
}

// WHO: PRToolRegistrar
// WHAT: Pull Request Tool Registration
// WHEN: Server initialization
// WHERE: GitHub MCP Server
// WHY: To provide GitHub PR operations
// HOW: By registering PR-related tools
// EXTENT: All pull request operations
func registerPRTools(s *server.MCPServer, getClient GetClientFn, t TranslationHelperFunc, readOnly bool) {
	s.AddTool(GetPullRequest(getClient, t))
	s.AddTool(ListPullRequests(getClient, t))
	s.AddTool(getPullRequestFiles(getClient, t))
	s.AddTool(getPullRequestStatus(getClient, t))
	s.AddTool(getPullRequestComments(getClient, t))
	s.AddTool(getPullRequestReviews(getClient, t))

	if !readOnly {
		s.AddTool(mergePullRequest(getClient, t))
		s.AddTool(updatePullRequestBranch(getClient, t))
		s.AddTool(createPullRequestReview(getClient, t))
		s.AddTool(createPullRequest(getClient, t))
		s.AddTool(updatePullRequest(getClient, t))
	}
}

// WHO: PRRetrievalTool
// WHAT: Pull Request Details Tool
// WHEN: During tool invocation
// WHERE: GitHub MCP Server
// WHY: To access details of a specific pull request
// HOW: By querying GitHub API with PR number
// EXTENT: Single PR details
func GetPullRequest(getClient GetClientFn, t TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_pull_request",
			mcp.WithDescription(t("TOOL_GET_PULL_REQUEST_DESCRIPTION", "Get details of a specific pull request")),
			mcp.WithString("owner",
				mcp.Description(RepositoryOwnerDesc),
				mcp.Required(),
			),
			mcp.WithString("repo",
				mcp.Description(RepositoryNameDesc),
				mcp.Required(),
			),
			mcp.WithNumber("pullNumber",
				mcp.Description("Pull request number"),
				mcp.Required(),
			),
		),
		func(ctx context.Context, r mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf(ErrGetGitHubClient, err)
			}

			// Extract required parameters
			owner, err := requiredParam[string](r, "owner")
			if err != nil {
				return nil, err
			}

			repo, err := requiredParam[string](r, "repo")
			if err != nil {
				return nil, err
			}

			pullNumber, err := RequiredInt(r, "pullNumber")
			if err != nil {
				return nil, err
			}

			// Call GitHub API
			pr, resp, err := client.PullRequests.Get(ctx, owner, repo, pullNumber)
			if err != nil {
				return nil, fmt.Errorf("failed to get pull request: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf(ErrReadResponseBody, err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to get pull request: %s", string(body))), nil
			}

			result, err := json.Marshal(pr)
			if err != nil {
				return nil, fmt.Errorf(ErrMarshalPR, err)
			}

			return mcp.NewToolResultText(string(result)), nil
		}
}

// WHO: RepositoryToolRegistrar
// WHAT: Repository Tool Registration
// WHEN: Server initialization
// WHERE: GitHub MCP Server
// WHY: To provide GitHub repository operations
// HOW: By registering repository-related tools
// EXTENT: All repository operations
func registerRepositoryTools(s *server.MCPServer, getClient GetClientFn, t TranslationHelperFunc, readOnly bool) {
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
func registerSearchTools(s *server.MCPServer, getClient GetClientFn, t TranslationHelperFunc) {
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
func registerUserTools(s *server.MCPServer, getClient GetClientFn, t TranslationHelperFunc) {
	s.AddTool(GetMe(getClient, t))
}

// WHO: CodeScanningToolRegistrar
// WHAT: Code Scanning Tool Registration
// WHEN: Server initialization
// WHERE: GitHub MCP Server
// WHY: To provide GitHub code scanning operations
// HOW: By registering code scanning-related tools
// EXTENT: All code scanning operations
func registerCodeScanningTools(s *server.MCPServer, getClient GetClientFn, t TranslationHelperFunc) {
	s.AddTool(GetCodeScanningAlert(getClient, t))
	s.AddTool(ListCodeScanningAlerts(getClient, t))
}

// WHO: PRListingTool
// WHAT: Pull Request Listing Tool
// WHEN: During tool invocation
// WHERE: GitHub MCP Server
// WHY: To provide a list of repository pull requests
// HOW: By calling GitHub API with proper pagination
// EXTENT: All repository pull requests
func ListPullRequests(getClient GetClientFn, t TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
	return createPRListTool(t), handlePRListRequest(getClient)
}

// WHO: PRListToolCreator
// WHAT: Pull Request List Tool Definition
// WHEN: Tool registration
// WHERE: GitHub MCP Server
// WHY: To define the PR listing tool parameters
// HOW: By configuring tool options with proper validation
// EXTENT: PR list tool interface
func createPRListTool(t TranslationHelperFunc) mcp.Tool {
	return mcp.NewTool("list_pull_requests",
		mcp.WithDescription(t("TOOL_LIST_PULL_REQUESTS_DESCRIPTION", "List and filter repository pull requests")),
		mcp.WithString("owner",
			mcp.Description(RepositoryOwnerDesc),
			mcp.Required(),
		),
		mcp.WithString("repo",
			mcp.Description(RepositoryNameDesc),
			mcp.Required(),
		),
		mcp.WithString("state",
			mcp.Description("Filter by state ('open', 'closed', 'all')"),
		),
		mcp.WithString("sort",
			mcp.Description("Sort by ('created', 'updated', 'popularity', 'long-running')"),
		),
		mcp.WithString("direction",
			mcp.Description("Sort direction ('asc', 'desc')"),
		),
		mcp.WithString("head",
			mcp.Description("Filter by head user/org and branch"),
		),
		mcp.WithString("base",
			mcp.Description("Filter by base branch"),
		),
		WithPagination(),
	)
}

// WHO: PRListRequestHandler
// WHAT: Pull Request List Request Handler
// WHEN: Tool invocation
// WHERE: GitHub MCP Server
// WHY: To process PR list requests
// HOW: By extracting parameters and calling GitHub API
// EXTENT: PR list request processing
func handlePRListRequest(getClient GetClientFn) server.ToolHandlerFunc {
	return func(ctx context.Context, r mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := getClient(ctx)
		if err != nil {
			return nil, fmt.Errorf(ErrGetGitHubClient, err)
		}

		// Extract required parameters
		owner, err := requiredParam[string](r, "owner")
		if err != nil {
			return nil, err
		}

		repo, err := requiredParam[string](r, "repo")
		if err != nil {
			return nil, err
		}

		// Extract optional parameters
		state, _ := OptionalParam[string](r, "state")
		if state == "" {
			state = "open"
		}

		// Get pagination parameters
		pagination, err := OptionalPaginationParams(r)
		if err != nil {
			return nil, err
		}

		// Extract additional optional parameters
		sort, _ := OptionalParam[string](r, "sort")
		direction, _ := OptionalParam[string](r, "direction")
		head, _ := OptionalParam[string](r, "head")
		base, _ := OptionalParam[string](r, "base")

		// Create list options
		opts := &github.PullRequestListOptions{
			State:     state,
			Sort:      sort,
			Direction: direction,
			Head:      head,
			Base:      base,
			ListOptions: github.ListOptions{
				Page:    pagination.page,
				PerPage: pagination.perPage,
			},
		}

		// Call GitHub API
		return fetchPullRequests(ctx, client, owner, repo, opts)
	}
}

// WHO: PRListFetcher
// WHAT: Pull Request List Fetcher
// WHEN: During PR list request handling
// WHERE: GitHub MCP Server
// WHY: To fetch pull requests from GitHub
// HOW: By making a GitHub API request and processing response
// EXTENT: GitHub API interaction
func fetchPullRequests(ctx context.Context, client *github.Client, owner string, repo string, opts *github.PullRequestListOptions) (*mcp.CallToolResult, error) {
	prs, resp, err := client.PullRequests.List(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf(ErrReadResponseBody, err)
		}
		return mcp.NewToolResultError(fmt.Sprintf("failed to list pull requests: %s", string(body))), nil
	}

	result, err := json.Marshal(prs)
	if err != nil {
		return nil, fmt.Errorf(ErrMarshalPR, err)
	}

	return mcp.NewToolResultText(string(result)), nil
}

// WHO: UserInfoTool
// WHAT: Authenticated User Tool
// WHEN: During user info request
// WHERE: GitHub MCP Server
// WHY: To provide user identity information
// HOW: By fetching authenticated user details from GitHub
// EXTENT: User identification
func GetMe(getClient GetClientFn, t TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_me",
			mcp.WithDescription(t("TOOL_GET_ME_DESCRIPTION", "Get details of the authenticated GitHub user. Use this when a request include \"me\", \"my\"...")),
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

// WHO: ParameterHelper
// WHAT: Optional Parameter Extraction
// WHEN: During parameter processing
// WHERE: GitHub MCP Server
// WHY: To safely extract optional parameters
// HOW: By checking parameter existence and type
// EXTENT: All tool parameter processing
func OptionalParamOK[T any](r mcp.CallToolRequest, p string) (value T, ok bool, err error) {
	// Check if the parameter is present in the request
	val, exists := r.Params.Arguments[p]
	if !exists {
		// Not present, return zero value, false, no error
		return
	}

	// Check if the parameter is of the expected type
	value, ok = val.(T)
	if !ok {
		// Present but wrong type
		err = fmt.Errorf("parameter %s is not of type %T, is %T", p, value, val)
		ok = true // Set ok to true because the parameter *was* present, even if wrong type
		return
	}

	// Present and correct type
	ok = true
	return
}

// WHO: ErrorChecker
// WHAT: Error Acceptance Checking
// WHEN: During error handling
// WHERE: GitHub MCP Server
// WHY: To identify acceptable errors
// HOW: By checking error type
// EXTENT: All error processing
func isAcceptedError(err error) bool {
	var acceptedError *github.AcceptedError
	return errors.As(err, &acceptedError)
}

// WHO: ParameterHelper
// WHAT: Required Parameter Extraction
// WHEN: During parameter processing
// WHERE: GitHub MCP Server
// WHY: To safely extract required parameters
// HOW: By checking parameter existence, type, and value
// EXTENT: All tool parameter processing
func requiredParam[T comparable](r mcp.CallToolRequest, p string) (T, error) {
	var zero T

	// Check if the parameter is present in the request
	if _, ok := r.Params.Arguments[p]; !ok {
		return zero, fmt.Errorf("missing required parameter: %s", p)
	}

	// Check if the parameter is of the expected type
	if _, ok := r.Params.Arguments[p].(T); !ok {
		return zero, fmt.Errorf("parameter %s is not of type %T", p, zero)
	}

	if r.Params.Arguments[p].(T) == zero {
		return zero, fmt.Errorf("missing required parameter: %s", p)
	}

	return r.Params.Arguments[p].(T), nil
}

// WHO: ParameterHelper
// WHAT: Required Integer Extraction
// WHEN: During parameter processing
// WHERE: GitHub MCP Server
// WHY: To safely extract required integer parameters
// HOW: By checking parameter existence, type, and value
// EXTENT: All integer parameter processing
func RequiredInt(r mcp.CallToolRequest, p string) (int, error) {
	v, err := requiredParam[float64](r, p)
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

// WHO: ParameterHelper
// WHAT: Optional Parameter Extraction
// WHEN: During parameter processing
// WHERE: GitHub MCP Server
// WHY: To safely extract optional parameters
// HOW: By checking parameter existence and type
// EXTENT: All tool parameter processing
func OptionalParam[T any](r mcp.CallToolRequest, p string) (T, error) {
	var zero T

	// Check if the parameter is present in the request
	if _, ok := r.Params.Arguments[p]; !ok {
		return zero, nil
	}

	// Check if the parameter is of the expected type
	if _, ok := r.Params.Arguments[p].(T); !ok {
		return zero, fmt.Errorf("parameter %s is not of type %T, is %T", p, zero, r.Params.Arguments[p])
	}

	return r.Params.Arguments[p].(T), nil
}

// WHO: ParameterHelper
// WHAT: Optional Integer Extraction
// WHEN: During parameter processing
// WHERE: GitHub MCP Server
// WHY: To safely extract optional integer parameters
// HOW: By checking parameter existence and type
// EXTENT: All integer parameter processing
func OptionalIntParam(r mcp.CallToolRequest, p string) (int, error) {
	v, err := OptionalParam[float64](r, p)
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

// WHO: ParameterHelper
// WHAT: Optional Integer with Default
// WHEN: During parameter processing
// WHERE: GitHub MCP Server
// WHY: To extract integer parameters with defaults
// HOW: By checking parameter and providing default if needed
// EXTENT: All integer parameter processing
func OptionalIntParamWithDefault(r mcp.CallToolRequest, p string, d int) (int, error) {
	v, err := OptionalIntParam(r, p)
	if err != nil {
		return 0, err
	}
	if v == 0 {
		return d, nil
	}
	return v, nil
}

// WHO: ParameterHelper
// WHAT: Optional String Array Extraction
// WHEN: During parameter processing
// WHERE: GitHub MCP Server
// WHY: To safely extract optional string array parameters
// HOW: By checking parameter existence and converting types
// EXTENT: All string array parameter processing
func OptionalStringArrayParam(r mcp.CallToolRequest, p string) ([]string, error) {
	// Check if the parameter is present in the request
	if _, ok := r.Params.Arguments[p]; !ok {
		return []string{}, nil
	}

	switch v := r.Params.Arguments[p].(type) {
	case nil:
		return []string{}, nil
	case []string:
		return v, nil
	case []any:
		strSlice := make([]string, len(v))
		for i, v := range v {
			s, ok := v.(string)
			if !ok {
				return []string{}, fmt.Errorf("parameter %s is not of type string, is %T", p, v)
			}
			strSlice[i] = s
		}
		return strSlice, nil
	default:
		return []string{}, fmt.Errorf("parameter %s could not be coerced to []string, is %T", p, r.Params.Arguments[p])
	}
}

// WHO: ToolOptionProvider
// WHAT: Pagination Option Provider
// WHEN: During tool definition
// WHERE: GitHub MCP Server
// WHY: To add standardized pagination parameters
// HOW: By adding page and perPage parameters to tool
// EXTENT: All paginated tools
func WithPagination() mcp.ToolOption {
	return func(tool *mcp.Tool) {
		mcp.WithNumber("page",
			mcp.Description("Page number for pagination (min 1)"),
			mcp.Min(1),
		)(tool)

		mcp.WithNumber("perPage",
			mcp.Description("Results per page for pagination (min 1, max 100)"),
			mcp.Min(1),
			mcp.Max(100),
		)(tool)
	}
}

// WHO: PaginationManager
// WHAT: Pagination Parameter Structure
// WHEN: During parameter processing
// WHERE: GitHub MCP Server
// WHY: To standardize pagination parameter handling
// HOW: By grouping pagination parameters in a struct
// EXTENT: All paginated requests
type PaginationParams struct {
	page    int
	perPage int
}

// WHO: PaginationHelper
// WHAT: Pagination Parameter Extraction
// WHEN: During parameter processing
// WHERE: GitHub MCP Server
// WHY: To safely extract pagination parameters
// HOW: By checking parameters and providing defaults
// EXTENT: All paginated requests
func OptionalPaginationParams(r mcp.CallToolRequest) (PaginationParams, error) {
	page, err := OptionalIntParamWithDefault(r, "page", 1)
	if err != nil {
		return PaginationParams{}, err
	}
	perPage, err := OptionalIntParamWithDefault(r, "perPage", 30)
	if err != nil {
		return PaginationParams{}, err
	}
	return PaginationParams{
		page:    page,
		perPage: perPage,
	}, nil
}

// Placeholder functions for PR operations - implementations would be in separate files
func getPullRequestFiles(getClient GetClientFn, t TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_pull_request_files",
			mcp.WithDescription(t("TOOL_GET_PR_FILES_DESCRIPTION", "Get the list of files changed in a pull request")),
			mcp.WithString("owner",
				mcp.Description(RepositoryOwnerDesc),
				mcp.Required(),
			),
			mcp.WithString("repo",
				mcp.Description(RepositoryNameDesc),
				mcp.Required(),
			),
			mcp.WithNumber("pullNumber",
				mcp.Description("Pull request number"),
				mcp.Required(),
			),
		),
		func(ctx context.Context, r mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Get GitHub client
			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf(ErrGetGitHubClient, err)
			}

			// Process request parameters
			params, err := extractPRParams(r)
			if err != nil {
				return nil, err
			}

			// Call GitHub API
			return fetchPRFiles(ctx, client, params.owner, params.repo, params.number)
		}
}

// WHO: PRParamsExtractor
// WHAT: Pull Request Parameter Extraction
// WHEN: During PR tool invocation
// WHERE: GitHub MCP Server
// WHY: To standardize PR parameter handling
// HOW: By grouping common parameter extraction logic
// EXTENT: All PR operation tools
type prParams struct {
	owner  string
	repo   string
	number int
}

// Extract common PR parameters from a request
func extractPRParams(r mcp.CallToolRequest) (prParams, error) {
	// Extract required parameters
	owner, err := requiredParam[string](r, "owner")
	if err != nil {
		return prParams{}, err
	}

	repo, err := requiredParam[string](r, "repo")
	if err != nil {
		return prParams{}, err
	}

	number, err := RequiredInt(r, "pullNumber")
	if err != nil {
		return prParams{}, err
	}

	return prParams{
		owner:  owner,
		repo:   repo,
		number: number,
	}, nil
}

// WHO: PRFileFetcher
// WHAT: Pull Request Files Fetcher
// WHEN: During PR files tool invocation
// WHERE: GitHub MCP Server
// WHY: To fetch changed files in a PR
// HOW: By calling GitHub API with proper PR details
// EXTENT: All changed files in a PR
func fetchPRFiles(ctx context.Context, client *github.Client, owner, repo string, number int) (*mcp.CallToolResult, error) {
	// Call GitHub API
	files, resp, err := client.PullRequests.ListFiles(ctx, owner, repo, number, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR files: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf(ErrReadResponseBody, err)
		}
		return mcp.NewToolResultError(fmt.Sprintf("failed to get PR files: %s", string(body))), nil
	}

	// Return result
	result, err := json.Marshal(files)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PR files: %w", err)
	}

	return mcp.NewToolResultText(string(result)), nil
}

// WHO: IssueCreationTool
// WHAT: Issue Creation Tool
// WHEN: During tool invocation
// WHERE: GitHub MCP Server
// WHY: To create new issues in repositories
// HOW: By submitting issue data to GitHub API
// EXTENT: New issue creation
func CreateIssue(getClient GetClientFn, t TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("create_issue",
			mcp.WithDescription(t("TOOL_CREATE_ISSUE_DESCRIPTION", "Create a new issue in a GitHub repository")),
			mcp.WithString("owner",
				mcp.Description(RepositoryOwnerDesc),
				mcp.Required(),
			),
			mcp.WithString("repo",
				mcp.Description(RepositoryNameDesc),
				mcp.Required(),
			),
			mcp.WithString("title",
				mcp.Description("Issue title"),
				mcp.Required(),
			),
			mcp.WithString("body",
				mcp.Description("Issue body content"),
			),
			mcp.WithArray("assignees",
				mcp.Description("Usernames to assign to this issue"),
				mcp.Items(mcp.String()),
			),
			mcp.WithArray("labels",
				mcp.Description("Labels to apply to this issue"),
				mcp.Items(mcp.String()),
			),
			mcp.WithNumber("milestone",
				mcp.Description("Milestone number"),
			),
		),
		func(ctx context.Context, r mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf(ErrGetGitHubClient, err)
			}

			// Extract required parameters
			owner, err := requiredParam[string](r, "owner")
			if err != nil {
				return nil, err
			}

			repo, err := requiredParam[string](r, "repo")
			if err != nil {
				return nil, err
			}

			title, err := requiredParam[string](r, "title")
			if err != nil {
				return nil, err
			}

			// Extract optional parameters
			body, _ := OptionalParam[string](r, "body")
			milestone, err := OptionalIntParam(r, "milestone")
			if err != nil {
				return nil, err
			}

			// Extract assignees and labels
			assignees, err := OptionalStringArrayParam(r, "assignees")
			if err != nil {
				return nil, err
			}

			labels, err := OptionalStringArrayParam(r, "labels")
			if err != nil {
				return nil, err
			}

			// Create issue request
			issueRequest := &github.IssueRequest{
				Title:     github.String(title),
				Body:      github.String(body),
				Assignees: &assignees,
				Labels:    &labels,
			}

			// Set milestone if provided
			if milestone > 0 {
				issueRequest.Milestone = github.Int(milestone)
			}

			// Call GitHub API
			issue, resp, err := client.Issues.Create(ctx, owner, repo, issueRequest)
			if err != nil {
				return nil, fmt.Errorf("failed to create issue: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusCreated {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf(ErrReadResponseBody, err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to create issue: %s", string(body))), nil
			}

			result, err := json.Marshal(issue)
			if err != nil {
				return nil, fmt.Errorf(ErrMarshalIssue, err)
			}

			return mcp.NewToolResultText(string(result)), nil
		}
}

// WHO: IssueCommentTool
// WHAT: Issue Comment Creation Tool
// WHEN: During tool invocation
// WHERE: GitHub MCP Server
// WHY: To add comments to existing issues
// HOW: By submitting comment data to GitHub API
// EXTENT: New comment creation
func AddIssueComment(getClient GetClientFn, t TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("add_issue_comment",
			mcp.WithDescription(t("TOOL_ADD_ISSUE_COMMENT_DESCRIPTION", "Add a comment to an existing issue")),
			mcp.WithString("owner",
				mcp.Description(RepositoryOwnerDesc),
				mcp.Required(),
			),
			mcp.WithString("repo",
				mcp.Description(RepositoryNameDesc),
				mcp.Required(),
			),
			mcp.WithNumber("issue_number",
				mcp.Description("Issue number to comment on"),
				mcp.Required(),
			),
			mcp.WithString("body",
				mcp.Description("Comment text"),
				mcp.Required(),
			),
		),
		func(ctx context.Context, r mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf(ErrGetGitHubClient, err)
			}

			// Extract required parameters
			owner, err := requiredParam[string](r, "owner")
			if err != nil {
				return nil, err
			}

			repo, err := requiredParam[string](r, "repo")
			if err != nil {
				return nil, err
			}

			issueNumber, err := RequiredInt(r, "issue_number")
			if err != nil {
				return nil, err
			}

			body, err := requiredParam[string](r, "body")
			if err != nil {
				return nil, err
			}

			// Create comment request
			commentRequest := &github.IssueComment{
				Body: github.String(body),
			}

			// Call GitHub API
			comment, resp, err := client.Issues.CreateComment(ctx, owner, repo, issueNumber, commentRequest)
			if err != nil {
				return nil, fmt.Errorf("failed to create comment: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusCreated {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf(ErrReadResponseBody, err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to create comment: %s", string(body))), nil
			}

			result, err := json.Marshal(comment)
			if err != nil {
				return nil, fmt.Errorf(ErrMarshalComment, err)
			}

			return mcp.NewToolResultText(string(result)), nil
		}
}

// WHO: IssueUpdateTool
// WHAT: Issue Update Tool
// WHEN: During tool invocation
// WHERE: GitHub MCP Server
// WHY: To modify existing GitHub issues
// HOW: By submitting updated issue data to GitHub API
// EXTENT: Existing issue modification
func UpdateIssue(getClient GetClientFn, t TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("update_issue",
			mcp.WithDescription(t("TOOL_UPDATE_ISSUE_DESCRIPTION", "Update an existing issue in a GitHub repository")),
			mcp.WithString("owner",
				mcp.Description(RepositoryOwnerDesc),
				mcp.Required(),
			),
			mcp.WithString("repo",
				mcp.Description(RepositoryNameDesc),
				mcp.Required(),
			),
			mcp.WithNumber("issue_number",
				mcp.Description("Issue number to update"),
				mcp.Required(),
			),
			mcp.WithString("title",
				mcp.Description("New title"),
			),
			mcp.WithString("body",
				mcp.Description("New description"),
			),
			mcp.WithString("state",
				mcp.Description("New state ('open' or 'closed')"),
				mcp.Enum("open", "closed"),
			),
			mcp.WithArray("assignees",
				mcp.Description("New assignees"),
				mcp.Items(mcp.String()),
			),
			mcp.WithArray("labels",
				mcp.Description("New labels"),
				mcp.Items(mcp.String()),
			),
			mcp.WithNumber("milestone",
				mcp.Description("New milestone number"),
			),
		),
		func(ctx context.Context, r mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf(ErrGetGitHubClient, err)
			}

			// Extract required parameters
			owner, err := requiredParam[string](r, "owner")
			if err != nil {
				return nil, err
			}

			repo, err := requiredParam[string](r, "repo")
			if err != nil {
				return nil, err
			}

			issueNumber, err := RequiredInt(r, "issue_number")
			if err != nil {
				return nil, err
			}

			// Extract optional parameters
			title, _ := OptionalParam[string](r, "title")
			body, _ := OptionalParam[string](r, "body")
			state, _ := OptionalParam[string](r, "state")
			milestone, err := OptionalIntParam(r, "milestone")
			if err != nil {
				return nil, err
			}

			// Extract assignees and labels
			assignees, err := OptionalStringArrayParam(r, "assignees")
			if err != nil {
				return nil, err
			}

			labels, err := OptionalStringArrayParam(r, "labels")
			if err != nil {
				return nil, err
			}

			// Create issue update request
			issueRequest := &github.IssueRequest{}

			// Only add fields that were provided
			if title != "" {
				issueRequest.Title = github.String(title)
			}

			if body != "" {
				issueRequest.Body = github.String(body)
			}

			if state != "" {
				issueRequest.State = github.String(state)
			}

			if milestone > 0 {
				issueRequest.Milestone = github.Int(milestone)
			} else if r.Params.Arguments["milestone"] != nil {
				// If milestone is explicitly set to 0, remove the milestone
				issueRequest.Milestone = github.Int(0)
			}

			if r.Params.Arguments["assignees"] != nil {
				issueRequest.Assignees = &assignees
			}

			if r.Params.Arguments["labels"] != nil {
				issueRequest.Labels = &labels
			}

			// Call GitHub API
			issue, resp, err := client.Issues.Edit(ctx, owner, repo, issueNumber, issueRequest)
			if err != nil {
				return nil, fmt.Errorf("failed to update issue: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf(ErrReadResponseBody, err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to update issue: %s", string(body))), nil
			}

			result, err := json.Marshal(issue)
			if err != nil {
				return nil, fmt.Errorf(ErrMarshalIssue, err)
			}

			return mcp.NewToolResultText(string(result)), nil
		}
}

// WHO: PRStatusTool
// WHAT: Pull Request Status Tool
// WHEN: During tool invocation
// WHERE: GitHub MCP Server
// WHY: To get combined status of PR checks
// HOW: By querying GitHub API for status information
// EXTENT: All status checks for a PR
func getPullRequestStatus(getClient GetClientFn, t TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_pull_request_status",
			mcp.WithDescription(t("TOOL_GET_PR_STATUS_DESCRIPTION", "Get the combined status of all status checks for a pull request")),
			mcp.WithString("owner",
				mcp.Description(RepositoryOwnerDesc),
				mcp.Required(),
			),
			mcp.WithString("repo",
				mcp.Description(RepositoryNameDesc),
				mcp.Required(),
			),
			mcp.WithNumber("pullNumber",
				mcp.Description("Pull request number"),
				mcp.Required(),
			),
		),
		func(ctx context.Context, r mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Get GitHub client
			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf(ErrGetGitHubClient, err)
			}

			// Extract PR parameters
			params, err := extractPRParams(r)
			if err != nil {
				return nil, err
			}

			// Fetch the PR to get its head SHA
			pr, resp, err := client.PullRequests.Get(ctx, params.owner, params.repo, params.number)
			if err != nil {
				return nil, fmt.Errorf("failed to get pull request: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("failed to get pull request: status %d", resp.StatusCode)
			}

			// Get the combined status for the head SHA
			return fetchCombinedStatus(ctx, client, params.owner, params.repo, pr.GetHead().GetSHA())
		}
}

// WHO: StatusFetcher
// WHAT: PR Status Fetcher
// WHEN: During PR status check
// WHERE: GitHub MCP Server
// WHY: To fetch combined status from GitHub
// HOW: By calling GitHub API with commit SHA
// EXTENT: All status checks for a commit
func fetchCombinedStatus(ctx context.Context, client *github.Client, owner, repo, sha string) (*mcp.CallToolResult, error) {
	combinedStatus, resp, err := client.Repositories.GetCombinedStatus(ctx, owner, repo, sha, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get combined status: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf(ErrReadResponseBody, err)
		}
		return mcp.NewToolResultError(fmt.Sprintf("failed to get combined status: %s", string(body))), nil
	}

	result, err := json.Marshal(combinedStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal status: %w", err)
	}

	return mcp.NewToolResultText(string(result)), nil
}

// WHO: PRCommentsTool
// WHAT: Pull Request Comments Tool
// WHEN: During tool invocation
// WHERE: GitHub MCP Server
// WHY: To retrieve PR review comments
// HOW: By fetching comments from GitHub API
// EXTENT: All review comments on a PR
func getPullRequestComments(getClient GetClientFn, t TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_pull_request_comments",
			mcp.WithDescription(t("TOOL_GET_PR_COMMENTS_DESCRIPTION", "Get the review comments on a pull request")),
			mcp.WithString("owner",
				mcp.Description(RepositoryOwnerDesc),
				mcp.Required(),
			),
			mcp.WithString("repo",
				mcp.Description(RepositoryNameDesc),
				mcp.Required(),
			),
			mcp.WithNumber("pullNumber",
				mcp.Description("Pull request number"),
				mcp.Required(),
			),
		),
		func(ctx context.Context, r mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Get GitHub client and parameters
			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf(ErrGetGitHubClient, err)
			}

			params, err := extractPRParams(r)
			if err != nil {
				return nil, err
			}

			// Call GitHub API
			comments, resp, err := client.PullRequests.ListComments(ctx, params.owner, params.repo, params.number, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to get PR comments: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf(ErrReadResponseBody, err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to get PR comments: %s", string(body))), nil
			}

			result, err := json.Marshal(comments)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal PR comments: %w", err)
			}

			return mcp.NewToolResultText(string(result)), nil
		}
}

// WHO: PRReviewsTool
// WHAT: Pull Request Reviews Tool
// WHEN: During tool invocation
// WHERE: GitHub MCP Server
// WHY: To retrieve PR review information
// HOW: By fetching reviews from GitHub API
// EXTENT: All reviews on a PR
func getPullRequestReviews(getClient GetClientFn, t TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_pull_request_reviews",
			mcp.WithDescription(t("TOOL_GET_PR_REVIEWS_DESCRIPTION", "Get the reviews on a pull request")),
			mcp.WithString("owner",
				mcp.Description(RepositoryOwnerDesc),
				mcp.Required(),
			),
			mcp.WithString("repo",
				mcp.Description(RepositoryNameDesc),
				mcp.Required(),
			),
			mcp.WithNumber("pullNumber",
				mcp.Description("Pull request number"),
				mcp.Required(),
			),
		),
		func(ctx context.Context, r mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Get GitHub client and extract parameters
			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf(ErrGetGitHubClient, err)
			}

			params, err := extractPRParams(r)
			if err != nil {
				return nil, err
			}

			// Call GitHub API
			return fetchPRReviews(ctx, client, params.owner, params.repo, params.number)
		}
}

// WHO: PRReviewsFetcher
// WHAT: Pull Request Reviews Fetcher
// WHEN: During PR reviews tool invocation
// WHERE: GitHub MCP Server
// WHY: To fetch reviews from GitHub
// HOW: By calling GitHub API with proper PR details
// EXTENT: All reviews on a PR
func fetchPRReviews(ctx context.Context, client *github.Client, owner, repo string, number int) (*mcp.CallToolResult, error) {
	// Call GitHub API
	reviews, resp, err := client.PullRequests.ListReviews(ctx, owner, repo, number, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR reviews: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf(ErrReadResponseBody, err)
		}
		return mcp.NewToolResultError(fmt.Sprintf("failed to get PR reviews: %s", string(body))), nil
	}

	// Return result
	result, err := json.Marshal(reviews)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PR reviews: %w", err)
	}

	return mcp.NewToolResultText(string(result)), nil
}

// WHO: PRMergeTool
// WHAT: Pull Request Merge Tool
// WHEN: During tool invocation
// WHERE: GitHub MCP Server
// WHY: To merge open pull requests
// HOW: By calling GitHub merge API with appropriate options
// EXTENT: PR merge operations
func mergePullRequest(getClient GetClientFn, t TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("merge_pull_request",
			mcp.WithDescription(t("TOOL_MERGE_PR_DESCRIPTION", "Merge a pull request")),
			mcp.WithString("owner",
				mcp.Description(RepositoryOwnerDesc),
				mcp.Required(),
			),
			mcp.WithString("repo",
				mcp.Description(RepositoryNameDesc),
				mcp.Required(),
			),
			mcp.WithNumber("pullNumber",
				mcp.Description("Pull request number"),
				mcp.Required(),
			),
			mcp.WithString("commit_title",
				mcp.Description("Title for merge commit"),
			),
			mcp.WithString("commit_message",
				mcp.Description("Extra detail for merge commit"),
			),
			mcp.WithString("merge_method",
				mcp.Description("Merge method ('merge', 'squash', 'rebase')"),
			),
		),
		func(ctx context.Context, r mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Extract client and parameters
			client, params, err := prepareClientAndPRParams(ctx, getClient, r)
			if err != nil {
				return nil, err
			}

			// Extract merge options
			options, err := extractMergeOptions(r)
			if err != nil {
				return nil, err
			}

			// Perform the merge
			return performMerge(ctx, client, params, options)
		}
}

// WHO: MergeOptionsExtractor
// WHAT: PR Merge Options Extractor
// WHEN: During PR merge tool invocation
// WHERE: GitHub MCP Server
// WHY: To prepare merge options
// HOW: By extracting merge parameters from request
// EXTENT: PR merge configuration
func extractMergeOptions(r mcp.CallToolRequest) (*github.PullRequestOptions, error) {
	// Extract optional parameters
	commitTitle, _ := OptionalParam[string](r, "commit_title")
	commitMessage, _ := OptionalParam[string](r, "commit_message")
	mergeMethod, _ := OptionalParam[string](r, "merge_method")

	// Create merge options
	options := &github.PullRequestOptions{
		CommitTitle:   commitTitle,
		CommitMessage: commitMessage,
	}

	// Set merge method if provided
	if mergeMethod != "" {
		if mergeMethod != "merge" && mergeMethod != "squash" && mergeMethod != "rebase" {
			return nil, fmt.Errorf("invalid merge_method: %s (must be 'merge', 'squash', or 'rebase')", mergeMethod)
		}
		options.MergeMethod = mergeMethod
	}

	return options, nil
}

// WHO: MergeOperationPerformer
// WHAT: PR Merge Operation
// WHEN: During PR merge
// WHERE: GitHub MCP Server
// WHY: To merge a pull request
// HOW: By calling GitHub API with merge options
// EXTENT: Single PR merge
func performMerge(ctx context.Context, client *github.Client, params prParams, options *github.PullRequestOptions) (*mcp.CallToolResult, error) {
	// Call GitHub API
	result, resp, err := client.PullRequests.Merge(ctx, params.owner, params.repo, params.number, "", options)
	if err != nil {
		return nil, fmt.Errorf("failed to merge pull request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf(ErrReadResponseBody, err)
		}
		return mcp.NewToolResultError(fmt.Sprintf("failed to merge pull request: %s", string(body))), nil
	}

	// Marshal result
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal merge result: %w", err)
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

// WHO: ClientPRParamsPreparation
// WHAT: Client and PR Parameters Preparation
// WHEN: During PR operations
// WHERE: GitHub MCP Server
// WHY: To standardize client and parameter extraction
// HOW: By combining client and parameter extraction
// EXTENT: All PR operations
func prepareClientAndPRParams(ctx context.Context, getClient GetClientFn, r mcp.CallToolRequest) (*github.Client, prParams, error) {
	// Get GitHub client
	client, err := getClient(ctx)
	if err != nil {
		return nil, prParams{}, fmt.Errorf(ErrGetGitHubClient, err)
	}

	// Extract PR parameters
	params, err := extractPRParams(r)
	if err != nil {
		return nil, prParams{}, err
	}

	return client, params, nil
}
