// WHO: IssuesModule
// WHAT: GitHub Issues functionality
// WHEN: During issue operations
// WHERE: System Layer 6 (Integration)
// WHY: To provide GitHub issues access
// HOW: Using GitHub API
// EXTENT: All issue operations

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github-mcp-server/pkg/common"
	"github-mcp-server/pkg/translations"

	"github.com/google/go-github/v71/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// The following imports are available from other package files:
// - GetClientFn from github.go
// - ContextTranslationFunc from common.go
// - RequiredParam, RequiredIntParam from common.go
// - OptionalParam from github.go
// - OptionalStringArrayParam from common.go
// - OptionalIntParamWithDefault from common.go
// - OptionalInt from github.go
// - Ptr from common.go
// - WithPagination from common.go

/*
 * WHO: IssueToolProvider
 * WHAT: Get issue tool definition
 * WHEN: During tool registration
 * WHERE: System Layer 6 (Integration)
 * WHY: To retrieve specific issue details via MCP
 * HOW: Using GitHub API with MCP protocol adapters
 * EXTENT: Single issue retrieval operation
 */
// GetIssue creates a tool to get details of a specific issue in a GitHub repository.
func GetIssue(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_issue",
			mcp.WithDescription(t("TOOL_GET_ISSUE_DESCRIPTION", "Get details of a specific issue in a GitHub repository")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("The owner of the repository"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("The name of the repository"),
			),
			mcp.WithNumber("issue_number",
				mcp.Required(),
				mcp.Description("The number of the issue"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// WHO: IssueRequestHandler
			// WHAT: Process issue retrieval request
			// WHEN: During tool invocation
			// WHERE: System Layer 6 (Integration)
			// WHY: To fetch issue data from GitHub
			// HOW: Using GitHub API client
			// EXTENT: Single issue data retrieval

			owner, err := common.OptionalStringArrayParam(request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			repo, err := common.OptionalStringArrayParam(request, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			issueNumber, err := common.OptionalIntParamWithDefault(request, "issue_number", 0)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			issue, resp, err := client.Issues.Get(ctx, owner[0], repo[0], issueNumber)
			if err != nil {
				return nil, fmt.Errorf("failed to get issue: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to get issue: %s", string(body))), nil
			}

			r, err := json.Marshal(issue)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal issue: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		})
}

// AddIssueComment creates a tool to add a comment to an issue.
func AddIssueComment(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("add_issue_comment",
			mcp.WithDescription(t("TOOL_ADD_ISSUE_COMMENT_DESCRIPTION", "Add a comment to an existing issue")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithNumber("issue_number",
				mcp.Required(),
				mcp.Description("Issue number to comment on"),
			),
			mcp.WithString("body",
				mcp.Required(),
				mcp.Description("Comment text"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, repo, issueNumber, err := extractIssueParams(request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			body, err := common.RequiredParam(request, "body")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			comment := &github.IssueComment{
				Body: Ptr(body),
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			createdComment, resp, err := client.Issues.CreateComment(ctx, owner, repo, issueNumber, comment)
			if err != nil {
				return nil, fmt.Errorf("failed to create comment: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusCreated {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to create comment: %s", string(body))), nil
			}

			r, err := json.Marshal(createdComment)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// SearchIssues creates a tool to search for issues and pull requests.
func SearchIssues(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("search_issues",
			mcp.WithDescription(t("TOOL_SEARCH_ISSUES_DESCRIPTION", "Search for issues and pull requests across GitHub repositories")),
			mcp.WithString("q",
				mcp.Required(),
				mcp.Description("Search query using GitHub issues search syntax"),
			),
			mcp.WithString("sort",
				mcp.Description("Sort field (comments, reactions, created, etc.)"),
				mcp.Enum(
					"comments",
					"reactions",
					"reactions-+1",
					"reactions--1",
					"reactions-smile",
					"reactions-thinking_face",
					"reactions-heart",
					"reactions-tada",
					"interactions",
					"created",
					"updated",
				),
			),
			mcp.WithString("order",
				mcp.Description("Sort order ('asc' or 'desc')"),
				mcp.Enum("asc", "desc"),
			),
			WithPagination(),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			query, err := common.RequiredParam(request, "q")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			sort, err := OptionalParam[string](request, "sort")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			order, err := OptionalParam[string](request, "order")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			pagination, err := common.OptionalPaginationParams(request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			opts := &github.SearchOptions{
				Sort:  sort,
				Order: order,
				ListOptions: github.ListOptions{
					PerPage: pagination.perPage,
					Page:    pagination.page,
				},
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			result, resp, err := client.Search.Issues(ctx, query, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to search issues: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to search issues: %s", string(body))), nil
			}

			r, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// CreateIssue creates a tool to create a new issue in a GitHub repository.
func CreateIssue(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("create_issue",
			mcp.WithDescription(t("TOOL_CREATE_ISSUE_DESCRIPTION", "Create a new issue in a GitHub repository")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithString("title",
				mcp.Required(),
				mcp.Description("Issue title"),
			),
			mcp.WithString("body",
				mcp.Description("Issue body content"),
			),
			mcp.WithArray("assignees",
				mcp.Description("Usernames to assign to this issue"),
				mcp.Items(
					map[string]interface{}{
						"type": "string",
					},
				),
			),
			mcp.WithArray("labels",
				mcp.Description("Labels to apply to this issue"),
				mcp.Items(
					map[string]interface{}{
						"type": "string",
					},
				),
			),
			mcp.WithNumber("milestone",
				mcp.Description("Milestone number"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := common.RequiredParam(request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repo, err := common.RequiredParam(request, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			title, err := common.RequiredParam(request, "title")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Optional parameters
			body, err := OptionalParam[string](request, "body")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Get assignees
			assignees, err := OptionalStringArrayParam(request, "assignees")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Get labels
			labels, err := OptionalStringArrayParam(request, "labels")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Get optional milestone
			milestone, err := OptionalIntParam(request, "milestone")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Create the issue request
			issueRequest := &github.IssueRequest{
				Title:     Ptr(title),
				Body:      Ptr(body),
				Assignees: &assignees,
				Labels:    &labels,
				Milestone: Ptr(milestone),
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			issue, resp, err := client.Issues.Create(ctx, owner, repo, issueRequest)
			if err != nil {
				return nil, fmt.Errorf("failed to create issue: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusCreated {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to create issue: %s", string(body))), nil
			}

			r, err := json.Marshal(issue)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// ListIssues creates a tool to list and filter repository issues
func ListIssues(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("list_issues",
			mcp.WithDescription(t("TOOL_LIST_ISSUES_DESCRIPTION", "List issues in a GitHub repository with filtering options")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithString("state",
				mcp.Description("Filter by state ('open', 'closed', 'all')"),
				mcp.Enum("open", "closed", "all"),
			),
			mcp.WithArray("labels",
				mcp.Description("Filter by labels"),
				mcp.Items(
					map[string]interface{}{
						"type": "string",
					},
				),
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
			WithPagination(),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := common.RequiredParam(request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repo, err := common.RequiredParam(request, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			opts := &github.IssueListByRepoOptions{}

			// Set optional parameters if provided
			opts.State, err = OptionalParam[string](request, "state")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Get labels
			opts.Labels, err = OptionalStringArrayParam(request, "labels")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			opts.Sort, err = OptionalParam[string](request, "sort")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			opts.Direction, err = OptionalParam[string](request, "direction")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			since, err := OptionalParam[string](request, "since")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if since != "" {
				timestamp, err := parseISOTimestamp(since)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("failed to list issues: %s", err.Error())), nil
				}
				opts.Since = timestamp
			}

			if page, ok := request.Params.Arguments["page"].(float64); ok {
				opts.Page = int(page)
			}

			if perPage, ok := request.Params.Arguments["perPage"].(float64); ok {
				opts.PerPage = int(perPage)
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			issues, resp, err := client.Issues.ListByRepo(ctx, owner, repo, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to list issues: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to list issues: %s", string(body))), nil
			}

			r, err := json.Marshal(issues)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal issues: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// UpdateIssue creates a tool to update an existing issue in a GitHub repository.
func UpdateIssue(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("update_issue",
			mcp.WithDescription(t("TOOL_UPDATE_ISSUE_DESCRIPTION", "Update an existing issue in a GitHub repository")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithNumber("issue_number",
				mcp.Required(),
				mcp.Description("Issue number to update"),
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
			mcp.WithArray("labels",
				mcp.Description("New labels"),
				mcp.Items(
					map[string]interface{}{
						"type": "string",
					},
				),
			),
			mcp.WithArray("assignees",
				mcp.Description("New assignees"),
				mcp.Items(
					map[string]interface{}{
						"type": "string",
					},
				),
			),
			mcp.WithNumber("milestone",
				mcp.Description("New milestone number"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := common.RequiredParam(request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repo, err := common.RequiredParam(request, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			issueNumber, err := common.RequiredInt(request, "issue_number")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Create the issue request with only provided fields
			issueRequest := &github.IssueRequest{}

			title, err := OptionalParam[string](request, "title")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if title != "" {
				issueRequest.Title = Ptr(title)
			}

			body, err := OptionalParam[string](request, "body")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if body != "" {
				issueRequest.Body = Ptr(body)
			}

			state, err := OptionalParam[string](request, "state")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if state != "" {
				issueRequest.State = Ptr(state)
			}

			labels, err := OptionalStringArrayParam(request, "labels")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if len(labels) > 0 {
				issueRequest.Labels = &labels
			}
			assignees, err := OptionalStringArrayParam(request, "assignees")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if len(assignees) > 0 {
				issueRequest.Assignees = &assignees
			}

			milestone, err := OptionalIntParam(request, "milestone")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if milestone != 0 {
				milestoneNum := milestone
				issueRequest.Milestone = &milestoneNum
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			updatedIssue, resp, err := client.Issues.Edit(ctx, owner, repo, issueNumber, issueRequest)
			if err != nil {
				return nil, fmt.Errorf("failed to update issue: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to update issue: %s", string(body))), nil
			}

			r, err := json.Marshal(updatedIssue)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// GetIssueComments creates a tool to get comments for a GitHub issue.
func GetIssueComments(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_issue_comments",
			mcp.WithDescription(t("TOOL_GET_ISSUE_COMMENTS_DESCRIPTION", "Get comments for a GitHub issue")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("Repository name"),
			),
			mcp.WithNumber("issue_number",
				mcp.Required(),
				mcp.Description("Issue number"),
			),
			mcp.WithNumber("page",
				mcp.Description("Page number"),
			),
			mcp.WithNumber("per_page",
				mcp.Description("Number of records per page"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := common.RequiredParam(request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			repo, err := common.RequiredParam(request, "repo")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			issueNumber, err := common.RequiredInt(request, "issue_number")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			page, err := OptionalIntParamWithDefault(request, "page", 1)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			perPage, err := OptionalIntParamWithDefault(request, "per_page", 30)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			opts := &github.IssueListCommentsOptions{
				ListOptions: github.ListOptions{
					Page:    page,
					PerPage: perPage,
				},
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			comments, resp, err := client.Issues.ListComments(ctx, owner, repo, issueNumber, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to get issue comments: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to get issue comments: %s", string(body))), nil
			}

			r, err := json.Marshal(comments)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// parseISOTimestamp parses an ISO 8601 timestamp string into a time.Time object.
// Returns the parsed time or an error if parsing fails.
// Example formats supported: "2023-01-15T14:30:00Z", "2023-01-15"
func parseISOTimestamp(timestamp string) (time.Time, error) {
	if timestamp == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}

	// Try RFC3339 format (standard ISO 8601 with time)
	t, err := time.Parse(time.RFC3339, timestamp)
	if err == nil {
		return t, nil
	}

	// Try simple date format (YYYY-MM-DD)
	t, err = time.Parse("2006-01-02", timestamp)
	if err == nil {
		return t, nil
	}

	// Return error with supported formats
	return time.Time{}, fmt.Errorf("invalid ISO 8601 timestamp: %s (supported formats: YYYY-MM-DDThh:mm:ssZ or YYYY-MM-DD)", timestamp)
}

/*
 * WHO: GitHubIssueHandler
 * WHAT: Issue operations for GitHub MCP server
 * WHEN: During issue API operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide issue management via MCP
 * HOW: Using GitHub API with MCP protocol adapters
 * EXTENT: All issue operations
 *
 * Example implementation following 7D Context Framework:
 */

/*
 * WHO: GitHubIssueHandler
 * WHAT: Issue operations for GitHub MCP server
 * WHEN: During issue API operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide issue management via MCP
 * HOW: Using GitHub API with MCP protocol adapters
 * EXTENT: All issue operations
 *
 * Example implementation following 7D Context Framework:
 */
// Additional implementations below

// WHO: IssueToolProvider
// WHAT: Get issues tool definition
// WHEN: During tool registration
// WHERE: System Layer 6 (Integration)
// WHY: To retrieve issues via MCP
// HOW: Using MCP tool definition mechanism
// EXTENT: Issue retrieval operations
func GetIssues(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	// We'll use the translation function directly without creating an adapter
	// since we're using it only for string translation
	tool = mcp.NewTool("get_issues",
		mcp.WithDescription(t("TOOL_GET_ISSUES_DESCRIPTION", "Gets issues from a GitHub repository")),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("Repository owner"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("Repository name"),
		),
		mcp.WithString("state",
			mcp.Description("Issue state (open, closed, all)"),
		),
		mcp.WithString("labels",
			mcp.Description("Comma-separated list of label names"),
		),
		// State already defined above, removing duplicate definition
		mcp.WithString("sort",
			mcp.Description("Sort field (created, updated, comments)"),
		),
		mcp.WithString("direction",
			mcp.Description("Sort direction (asc or desc)"),
		),
		mcp.WithString("since",
			mcp.Description("Only issues updated after this time (ISO 8601 format)"),
		),
		mcp.WithNumber("per_page",
			mcp.Description("Results per page (max 100)"),
		),
		mcp.WithNumber("page",
			mcp.Description("Page number"),
		),
	)

	// Handler function
	handler = func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// WHO: IssueRequestHandler
		// WHAT: Handle issue request
		// WHEN: During tool invocation
		// WHERE: System Layer 6 (Integration)
		// WHY: To process issue requests
		// HOW: Using GitHub API client
		// EXTENT: Issue API operations

		// Debug logs would go here if we had a logger
		// Extract parameters
		owner, err := ExtractRequiredParam[string](request, "owner")
		if err != nil {
			return nil, fmt.Errorf("invalid owner parameter: %w", err)
		}

		repo, err := ExtractRequiredParam[string](request, "repo")
		if err != nil {
			return nil, fmt.Errorf("invalid repo parameter: %w", err)
		}

		// Optional parameters
		state, _ := OptionalParam[string](request, "state")
		labels, _ := OptionalParam[string](request, "labels")
		sort, _ := OptionalParam[string](request, "sort")
		direction, _ := OptionalParam[string](request, "direction")
		sinceStr, _ := OptionalParam[string](request, "since")

		perPage, hasPerPage, err := OptionalInt(request, "per_page")
		if err != nil {
			return nil, fmt.Errorf("invalid per_page parameter: %w", err)
		}

		page, hasPage, err := OptionalInt(request, "page")
		if err != nil {
			return nil, fmt.Errorf("invalid page parameter: %w", err)
		}

		// Get GitHub client
		client, err := getClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get GitHub client: %w", err)
		}

		// Apply string translation if provided
		if t != nil {
			_ = t("LOG_HANDLING_ISSUES_REQUEST", "Handling issues request")
			// This just translates the log message but doesn't log it since we're not using the logger
		}

		// Prepare list options
		opts := &github.IssueListByRepoOptions{}

		if state != "" {
			opts.State = state
		}

		if labels != "" {
			opts.Labels = []string{labels}
		}

		if sort != "" {
			opts.Sort = sort
		}

		if direction != "" {
			opts.Direction = direction
		}

		if sinceStr != "" {
			since, err := time.Parse(time.RFC3339, sinceStr)
			if err != nil {
				return nil, fmt.Errorf("invalid since parameter: %w", err)
			}
			opts.Since = since
		}

		if hasPerPage {
			opts.PerPage = perPage
		} else {
			opts.PerPage = 30 // Default
		}

		if hasPage {
			opts.Page = page
		} else {
			opts.Page = 1 // Default
		}

		// Call GitHub API
		issues, _, err := client.Issues.ListByRepo(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to get issues: %w", err)
		}

		// Transform to response format
		issuesList := make([]map[string]interface{}, 0, len(issues))
		for _, issue := range issues {
			// Skip pull requests (which GitHub API returns as issues)
			if issue.IsPullRequest() {
				continue
			}

			// Map labels
			labels := make([]map[string]interface{}, 0, len(issue.Labels))
			for _, label := range issue.Labels {
				labels = append(labels, map[string]interface{}{
					"name":  label.GetName(),
					"color": label.GetColor(),
				})
			}

			// Map assignees
			assignees := make([]map[string]interface{}, 0, len(issue.Assignees))
			for _, assignee := range issue.Assignees {
				assignees = append(assignees, map[string]interface{}{
					"login": assignee.GetLogin(),
				})
			}

			issuesList = append(issuesList, map[string]interface{}{
				"number":     issue.GetNumber(),
				"title":      issue.GetTitle(),
				"state":      issue.GetState(),
				"html_url":   issue.GetHTMLURL(),
				"body":       issue.GetBody(),
				"user":       map[string]interface{}{"login": issue.User.GetLogin()},
				"labels":     labels,
				"created_at": issue.GetCreatedAt().Format(time.RFC3339),
				"updated_at": issue.GetUpdatedAt().Format(time.RFC3339),
				"comments":   issue.GetComments(),
				"assignees":  assignees,
			})
		}

		response := map[string]interface{}{
			"issues":      issuesList,
			"total_count": len(issuesList),
		}

		responseJSON, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response: %w", err)
		}
		return mcp.NewToolResultText(string(responseJSON)), nil
	}

	return tool, handler
}

// WHO: IssueCreateToolProvider
// WHAT: Create issue tool definition
// WHEN: During tool registration
// WHERE: System Layer 6 (Integration)
// WHY: To create issues via MCP
// HOW: Using MCP tool definition mechanism
// EXTENT: Issue creation operations
func CreateIssueEnhanced(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	// WHO: IssueCreateToolDefiner
	// WHAT: Define create issue tool
	// WHEN: During server initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To register tool with MCP
	// HOW: Using MCP tool definition
	// EXTENT: Issue creation tool interface

	tool = mcp.NewTool("create_issue",
		mcp.WithDescription("Creates a new issue in a GitHub repository"),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("Repository owner"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("Repository name"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Issue title"),
		),
		mcp.WithString("body",
			mcp.Description("Issue body content in markdown format"),
		),
		mcp.WithArray("labels",
			mcp.Description("Array of label names to apply"),
			mcp.Items(
				map[string]interface{}{
					"type": "string",
				},
			),
		),
		mcp.WithArray("assignees",
			mcp.Description("Array of usernames to assign"),
			mcp.Items(
				map[string]interface{}{
					"type": "string",
				},
			),
		),
	)

	// Handler function
	handler = func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// WHO: IssueCreateHandler
		// WHAT: Handle issue creation
		// WHEN: During tool invocation
		// WHERE: System Layer 6 (Integration)
		// WHY: To process issue creation
		// HOW: Using GitHub API client
		// EXTENT: Issue creation operations

		// Debug logs would go here if we had a logger

		// Extract parameters
		owner, err := common.RequiredParam(request, "owner")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		repo, err := common.RequiredParam(request, "repo")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		title, err := common.RequiredParam(request, "title")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Optional parameters
		body, _ := OptionalParam[string](request, "body")
		labels, err := OptionalStringArrayParam(request, "labels")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		assignees, err := OptionalStringArrayParam(request, "assignees")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get GitHub client
		client, err := getClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get GitHub client: %w", err)
		}

		// Apply string translation if provided
		if t != nil {
			_ = t("ISSUE_CREATE", "Creating issue")
			// In a real implementation, we would do more with this
		}

		// Prepare request
		issueRequest := &github.IssueRequest{
			Title: Ptr(title),
		}

		if body != "" {
			issueRequest.Body = Ptr(body)
		}

		if len(labels) > 0 {
			issueRequest.Labels = &labels
		}

		if len(assignees) > 0 {
			issueRequest.Assignees = &assignees
		}

		// Call GitHub API
		issue, _, err := client.Issues.Create(ctx, owner, repo, issueRequest)
		if err != nil {
			return nil, fmt.Errorf("failed to create issue: %w", err)
		}

		// Transform to response format
		response := map[string]interface{}{
			"number":     issue.GetNumber(),
			"title":      issue.GetTitle(),
			"state":      issue.GetState(),
			"html_url":   issue.GetHTMLURL(),
			"created_at": issue.GetCreatedAt().Format(time.RFC3339),
		}

		responseJSON, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response: %w", err)
		}

		return mcp.NewToolResultText(string(responseJSON)), nil
	}

	return tool, handler
}

// WHO: IssueUpdateToolProvider
// WHAT: Update issue tool definition
// WHEN: During tool registration
// WHERE: System Layer 6 (Integration)
// WHY: To update issues via MCP
// HOW: Using MCP tool definition mechanism
// EXTENT: Issue update operations
// WHO: IssueUpdateToolProvider
// WHAT: Update issue tool definition
// WHEN: During tool registration
// WHERE: System Layer 6 (Integration)
// WHY: To update issues via MCP
// HOW: Using MCP tool definition mechanism
// EXTENT: Issue update operations
func UpdateIssueEnhanced(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	// WHO: IssueUpdateToolDefiner
	// WHAT: Define update issue tool
	// WHEN: During server initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To register tool with MCP
	// HOW: Using MCP tool definition
	// EXTENT: Issue update tool interface

	tool = mcp.NewTool("update_issue",
		mcp.WithDescription("Updates an existing issue in a GitHub repository"),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("Repository owner"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("Repository name"),
		),
		mcp.WithNumber("number",
			mcp.Required(),
			mcp.Description("Issue number"),
		),
		mcp.WithString("title",
			mcp.Description("New issue title"),
		),
		mcp.WithString("body",
			mcp.Description("New issue body content in markdown format"),
		),
		mcp.WithString("state",
			mcp.Description("Issue state (open, closed)"),
		),
		mcp.WithArray("labels",
			mcp.Description("Array of label names to apply (replaces existing labels)"),
			mcp.Items(
				map[string]interface{}{
					"type": "string",
				},
			),
		),
		mcp.WithArray("assignees",
			mcp.Description("Array of usernames to assign (replaces existing assignees)"),
			mcp.Items(
				map[string]interface{}{
					"type": "string",
				},
			),
		),
	)

	// Handler function
	handler = func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract parameters
		owner, err := common.RequiredParam(request, "owner")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		repo, err := common.RequiredParam(request, "repo")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		number, err := common.RequiredInt(request, "number")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Optional parameters
		title, err := OptionalParam[string](request, "title")
		hasTitle := title != "" && err == nil

		body, err := OptionalParam[string](request, "body")
		hasBody := body != "" && err == nil

		state, err := OptionalParam[string](request, "state")
		hasState := state != "" && err == nil

		labels, err := OptionalStringArrayParam(request, "labels")
		hasLabels := len(labels) > 0 && err == nil

		assignees, err := OptionalStringArrayParam(request, "assignees")
		hasAssignees := len(assignees) > 0 && err == nil

		// Verify we have at least one update parameter
		if !hasTitle && !hasBody && !hasState && !hasLabels && !hasAssignees {
			return mcp.NewToolResultError("at least one update parameter (title, body, state, labels, assignees) must be provided"), nil
		}

		// Get GitHub client
		client, err := getClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get GitHub client: %w", err)
		}

		// Apply string translation if provided
		if t != nil {
			_ = t("ISSUE_UPDATE", "Updating issue")
			// In a real implementation, we would do more with this
		}

		// Prepare request
		issueRequest := &github.IssueRequest{}

		if hasTitle {
			issueRequest.Title = github.String(title)
		}

		if hasBody {
			issueRequest.Body = github.String(body)
		}

		if hasState {
			issueRequest.State = github.String(state)
		}

		if hasLabels {
			issueRequest.Labels = &labels
		}

		if hasAssignees {
			issueRequest.Assignees = &assignees
		}

		// Call GitHub API
		issue, _, err := client.Issues.Edit(ctx, owner, repo, number, issueRequest)
		if err != nil {
			return nil, fmt.Errorf("failed to update issue: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response body: %w", err)
			}
			return mcp.NewToolResultError(fmt.Sprintf("failed to update issue: %s", string(body))), nil
		}

		r, err := json.Marshal(issue)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response: %w", err)
		}

		return mcp.NewToolResultText(string(r)), nil
	}

	return tool, handler
}

// CreateContextAdapter creates a context translation adapter for use in issue handlers
func CreateContextAdapter(t translations.TranslationHelperFunc) ContextTranslationFunc {
	return func(ctx context.Context, contextData map[string]interface{}) (map[string]interface{}, error) {
		// This is a simple adapter that doesn't modify the context
		return contextData, nil
	}
}

// WHO: TypeDefinitions
// WHAT: Common type definitions for GitHub MCP server
// WHEN: During component initialization
// WHERE: System Layer 6 (Integration)
// WHY: To provide consistent type signatures
// HOW: Using Go type definitions
// EXTENT: All GitHub MCP operations

// Use GetClientFn from the top of the file
// GetClientFn is already defined at the top of this file and in github.go

// Deprecated: use common.RequiredParam, common.OptionalParamOK, common.OptionalIntParamWithDefault, etc.
// func OptionalIntParam(request mcp.CallToolRequest, name string) (int, error) { ... }
// func RequiredIntParam(request mcp.CallToolRequest, name string) (int, error) { ... }
// func OptionalBoolParam(request mcp.CallToolRequest, name string) (bool, bool, error) { ... }
