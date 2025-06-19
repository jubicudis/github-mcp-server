// WHO: IssuesModule
// WHAT: GitHub Issues functionality
// WHEN: During issue operations
// WHERE: System Layer 6 (Integration)
// WHY: To provide GitHub issues access
// HOW: Using GitHub API
// EXTENT: All issue operations

package ghmcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/common"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/tranquilspeak"

	"github.com/google/go-github/v71/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Canonical TNOS 7D context and Helical Memory logging integration
// All handlers below are canonical, DRY, and reference only helpers from /pkg/common, /pkg/testutil, and /pkg/translations.
// All logging is routed through Helical Memory with 7D context (Who, What, When, Where, Why, How, To What Extent).
// No placeholders, stubs, or duplicate code allowed. All logic is modular and polyglot-compliant.

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
func GetIssue(getClient common.GetClientFn, t common.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return getIssueInternal(getClient, t, nil)
}

// AddIssueComment creates a tool to add a comment to an issue.
func AddIssueComment(getClient common.GetClientFn, t common.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return addIssueCommentInternal(getClient, t, nil)
}

// SearchIssues creates a tool to search for issues and pull requests.
func SearchIssues(getClient common.GetClientFn, t common.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return searchIssuesInternal(getClient, t, nil)
}

// CreateIssue creates a tool to create a new issue in a GitHub repository.
func CreateIssue(getClient common.GetClientFn, t common.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	tool = mcp.NewTool(
		"create_issue",
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
	)
	handler = func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// WHO: IssueCreateHandler
		// WHAT: Handle create issue request
		// WHEN: During tool invocation
		// WHERE: System Layer 6 (Integration)
		// WHY: To create a new issue in GitHub
		// HOW: Using GitHub API client
		// EXTENT: Single issue creation

		// Log entry
		owner, err := common.RequiredParam[string](request, "owner")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		repo, err := common.RequiredParam[string](request, "repo")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		title, err := common.RequiredParam[string](request, "title")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Optional parameters
		body, _, err := common.OptionalParamOK[string](request, "body")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get assignees
		assignees, _, err := common.OptionalParamOK[[]string](request, "assignees")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get labels
		labels, _, err := common.OptionalParamOK[[]string](request, "labels")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get optional milestone
		milestone, err := common.OptionalIntParamWithDefault(request, "milestone", 0)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Create the issue request
		issueRequest := &github.IssueRequest{
			Title: common.Ptr(title),
		}
		if body != "" {
			issueRequest.Body = common.Ptr(body)
		}
		if len(assignees) > 0 {
			issueRequest.Assignees = &assignees
		}
		if len(labels) > 0 {
			issueRequest.Labels = &labels
		}
		if milestone > 0 {
			issueRequest.Milestone = common.Ptr(milestone)
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
	return tool, handler
}

// ListIssues creates a tool to list and filter repository issues
func ListIssues(getClient common.GetClientFn, t common.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	tool = mcp.NewTool(
		"list_issues",
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
		mcp.WithNumber("per_page",
			mcp.Description("Results per page (max 100)"),
		),
		mcp.WithNumber("page",
			mcp.Description("Page number"),
		),
	)
	handler = func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// WHO: IssueListHandler
		// WHAT: Handle list issues request
		// WHEN: During tool invocation
		// WHERE: System Layer 6 (Integration)
		// WHY: To list and filter issues in a GitHub repository
		// HOW: Using GitHub API client
		// EXTENT: Multiple issue retrieval

		owner, err := common.RequiredParam[string](request, "owner")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		repo, err := common.RequiredParam[string](request, "repo")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		opts := &github.IssueListByRepoOptions{}

		// Set optional parameters if provided
		state, _, err := common.OptionalParamOK[string](request, "state")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if state != "" {
			opts.State = state
		}

		labels, err := common.OptionalStringArrayParam(request, "labels")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if len(labels) > 0 {
			opts.Labels = labels
		}

		sort, _, err := common.OptionalParamOK[string](request, "sort")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if sort != "" {
			opts.Sort = sort
		}

		direction, _, err := common.OptionalParamOK[string](request, "direction")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if direction != "" {
			opts.Direction = direction
		}

		sinceStr, _, err := common.OptionalParamOK[string](request, "since")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if sinceStr != "" {
			timestamp, err := parseISOTimestamp(sinceStr)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to list issues: %s", err.Error())), nil
			}
			opts.Since = timestamp
		}

		page, err := common.OptionalIntParamWithDefault(request, "page", 1)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		opts.Page = page

		perPage, err := common.OptionalIntParamWithDefault(request, "per_page", 30)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		opts.PerPage = perPage

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
	return tool, handler
}

// UpdateIssue creates a tool to update an existing issue in a GitHub repository.
func UpdateIssue(getClient common.GetClientFn, t common.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	tool = mcp.NewTool(
		"update_issue",
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
	)
	handler = func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// WHO: IssueUpdateHandler
		// WHAT: Handle update issue request
		// WHEN: During tool invocation
		// WHERE: System Layer 6 (Integration)
		// WHY: To update an existing issue in GitHub
		// HOW: Using GitHub API client
		// EXTENT: Single issue update

		owner, err := common.RequiredParam[string](request, "owner")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		repo, err := common.RequiredParam[string](request, "repo")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		issueNumber, err := common.RequiredIntParam(request, "issue_number")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Optional parameters
		title, _, err := common.OptionalParamOK[string](request, "title")
		hasTitle := title != "" && err == nil

		body, _, err := common.OptionalParamOK[string](request, "body")
		hasBody := body != "" && err == nil

		state, _, err := common.OptionalParamOK[string](request, "state")
		hasState := state != "" && err == nil

		labels, err := common.OptionalStringArrayParam(request, "labels")
		hasLabels := len(labels) > 0 && err == nil

		assignees, err := common.OptionalStringArrayParam(request, "assignees")
		hasAssignees := len(assignees) > 0 && err == nil

		milestone, err := common.OptionalIntParam(request, "milestone")
		hasMilestone := err == nil && milestone != 0

		if !hasTitle && !hasBody && !hasState && !hasLabels && !hasAssignees && !hasMilestone {
			return mcp.NewToolResultError("at least one update parameter (title, body, state, labels, assignees) must be provided"), nil
		}

		client, err := getClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get GitHub client: %w", err)
		}

		issueRequest := &github.IssueRequest{}
		if hasTitle {
			issueRequest.Title = common.Ptr(title)
		}
		if hasBody {
			issueRequest.Body = common.Ptr(body)
		}
		if hasState {
			issueRequest.State = common.Ptr(state)
		}
		if hasLabels {
			issueRequest.Labels = &labels
		}
		if hasAssignees {
			issueRequest.Assignees = &assignees
		}
		if hasMilestone {
			issueRequest.Milestone = common.Ptr(milestone)
		}

		issue, resp, err := client.Issues.Edit(ctx, owner, repo, issueNumber, issueRequest)
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

// GetIssueComments creates a tool to get comments for a GitHub issue.
func GetIssueComments(getClient common.GetClientFn, t common.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return getIssueCommentsInternal(getClient, t, nil)
}

// Internal implementations that accept triggerMatrix for event-driven logging (used only by event-driven code)
func getIssueInternal(getClient common.GetClientFn, t common.TranslationHelperFunc, triggerMatrix *tranquilspeak.TriggerMatrix) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	tool = mcp.NewTool(
		"get_issue",
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
	)
	handler = func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		logger := log.NewLogger(triggerMatrix)
		logger.Info("Handling get_issue request: who=%s what=%s when=%s where=%s why=%s how=%s extent=%s", "IssueRequestHandler", "Process issue retrieval request", time.Now().Format(time.RFC3339), "System Layer 6 (Integration)", "To fetch issue data from GitHub", "Using GitHub API client", "Single issue data retrieval")

		// Extract required parameters
		owner, err := common.RequiredParam[string](request, "owner")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		repo, err := common.RequiredParam[string](request, "repo")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		issueNumber, err := common.RequiredIntParam(request, "issue_number")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		client, err := getClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get GitHub client: %w", err)
		}
		issue, resp, err := client.Issues.Get(ctx, owner, repo, issueNumber)
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
	}
	return tool, handler
}
func addIssueCommentInternal(getClient common.GetClientFn, t common.TranslationHelperFunc, triggerMatrix *tranquilspeak.TriggerMatrix) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	tool = mcp.NewTool(
		"add_issue_comment",
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
	)
	handler = func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		logger := log.NewLogger(triggerMatrix)
		logger.Info("Handling add_issue_comment request: who=%s what=%s when=%s where=%s why=%s how=%s extent=%s", "IssueRequestHandler", "Process issue comment addition", time.Now().Format(time.RFC3339), "System Layer 6 (Integration)", "To add a comment to an issue in GitHub", "Using GitHub API client", "Single comment addition")

		// Extract required parameters
		owner, err := common.RequiredParam[string](request, "owner")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		repo, err := common.RequiredParam[string](request, "repo")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		issueNumber, err := common.RequiredIntParam(request, "issue_number")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		body, err := common.RequiredParam[string](request, "body")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		comment := &github.IssueComment{Body: common.Ptr(body)}
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
	return tool, handler
}
func searchIssuesInternal(getClient common.GetClientFn, t common.TranslationHelperFunc, triggerMatrix *tranquilspeak.TriggerMatrix) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	tool = mcp.NewTool(
		"search_issues",
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
			),
		),
		mcp.WithString("order",
			mcp.Description("Sort order ('asc' or 'desc')"),
			mcp.Enum("asc", "desc"),
		),
		mcp.WithNumber("page",
			mcp.Description("Page number"),
		),
		mcp.WithNumber("per_page",
			mcp.Description("Results per page (max 100)"),
		),
	)
	handler = func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		logger := log.NewLogger(triggerMatrix)
		logger.Info("Handling search_issues request: who=%s what=%s when=%s where=%s why=%s how=%s extent=%s", "IssueRequestHandler", "Process issue search request", time.Now().Format(time.RFC3339), "System Layer 6 (Integration)", "To search for issues in GitHub repositories", "Using GitHub API client", "Multiple issue search")

		query, err := common.RequiredParam[string](request, "q")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		sortVal, _, err := common.OptionalParamOK[string](request, "sort")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		orderVal, _, err := common.OptionalParamOK[string](request, "order")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		page, err := common.OptionalIntParamWithDefault(request, "page", 1)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		perPage, err := common.OptionalIntParamWithDefault(request, "per_page", 30)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		opts := &github.SearchOptions{
			Sort:  sortVal,
			Order: orderVal,
			ListOptions: github.ListOptions{PerPage: perPage, Page: page},
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
	return tool, handler
}
func getIssueCommentsInternal(getClient common.GetClientFn, t common.TranslationHelperFunc, triggerMatrix *tranquilspeak.TriggerMatrix) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	tool = mcp.NewTool(
		"get_issue_comments",
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
	)
	handler = func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		logger := log.NewLogger(triggerMatrix)
		logger.Info("Handling get_issue_comments request: who=%s what=%s when=%s where=%s why=%s how=%s extent=%s", "IssueCommentsHandler", "Process get issue comments request", time.Now().Format(time.RFC3339), "System Layer 6 (Integration)", "To retrieve comments for a specific issue in GitHub", "Using GitHub API client", "Multiple comment retrieval")

		owner, err := common.RequiredParam[string](request, "owner")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		repo, err := common.RequiredParam[string](request, "repo")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		issueNumber, err := common.RequiredIntParam(request, "issue_number")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		page, err := common.OptionalIntParamWithDefault(request, "page", 1)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		perPage, err := common.OptionalIntParamWithDefault(request, "per_page", 30)
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
	return tool, handler
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
func GetIssues(getClient common.GetClientFn, t common.TranslationHelperFunc, triggerMatrix *tranquilspeak.TriggerMatrix) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	tool = mcp.NewTool(
		"get_issues",
		mcp.WithDescription(t("TOOL_GET_ISSUES_DESCRIPTION", "Get issues for a GitHub repository")),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("Repository owner"),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description("Repository name"),
		),
		mcp.WithString("state",
			mcp.Description("State of the issues (open, closed, all)"),
		),
		mcp.WithString("labels",
			mcp.Description("Comma-separated list of labels"),
		),
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
	handler = func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		logger := log.NewLogger(triggerMatrix)
		logger.Info("Handling get_issues request: who=%s what=%s when=%s where=%s why=%s how=%s extent=%s", "IssueRequestHandler", "Process get issues request", time.Now().Format(time.RFC3339), "System Layer 6 (Integration)", "To retrieve issues from GitHub", "Using GitHub API client", "Multiple issue retrieval")

		// Extract parameters
		owner, err := common.RequiredParam[string](request, "owner")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		repo, err := common.RequiredParam[string](request, "repo")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Optional parameters
		state, _, _ := common.OptionalParamOK[string](request, "state")
		labels, _, _ := common.OptionalParamOK[string](request, "labels")
		sort, _, _ := common.OptionalParamOK[string](request, "sort")
		direction, _, _ := common.OptionalParamOK[string](request, "direction")
		sinceStr, _, _ := common.OptionalParamOK[string](request, "since")

		perPage, err := common.OptionalIntParamWithDefault(request, "per_page", 30)
		if err != nil {
			return nil, fmt.Errorf("invalid per_page parameter: %w", err)
		}

		page, err := common.OptionalIntParamWithDefault(request, "page", 1)
		if err != nil {
			return nil, fmt.Errorf("invalid page parameter: %w", err)
		}

		// Get GitHub client
		client, err := getClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get GitHub client: %w", err)
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

		opts.PerPage = perPage
		opts.Page = page

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
