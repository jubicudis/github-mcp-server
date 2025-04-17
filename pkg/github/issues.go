package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v69/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// GetIssue creates a tool to get details of a specific issue in a GitHub repository.
func GetIssue(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return githubgithubgithubgithubgithubgithubgithubMCP.NewTool("get_issue",
			githubgithubgithubgithubgithubgithubgithubMCP.WithDescription(t("TOOL_GET_ISSUE_DESCRIPTION", "Get details of a specific issue in a GitHub repository")),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("owner",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("The owner of the repository"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("repo",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("The name of the repository"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubMCP.WithNumber("issue_number",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("The number of the issue"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			issueNumber, err := RequiredInt(request, "issue_number")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
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
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(fmt.Sprintf("failed to get issue: %s", string(body))), nil
			}

			r, err := json.Marshal(issue)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal issue: %w", err)
			}

			return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultText(string(r)), nil
		}
}

// AddIssueComment creates a tool to add a comment to an issue.
func AddIssueComment(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return githubgithubgithubgithubgithubgithubgithubMCP.NewTool("add_issue_comment",
			githubgithubgithubgithubgithubgithubgithubMCP.WithDescription(t("TOOL_ADD_ISSUE_COMMENT_DESCRIPTION", "Add a comment to an existing issue")),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("owner",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Repository owner"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("repo",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Repository name"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubMCP.WithNumber("issue_number",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Issue number to comment on"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("body",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Comment text"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			issueNumber, err := RequiredInt(request, "issue_number")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			body, err := requiredParam[string](request, "body")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}

			comment := &github.IssueComment{
				Body: githubMCP.Ptr(body),
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
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(fmt.Sprintf("failed to create comment: %s", string(body))), nil
			}

			r, err := json.Marshal(createdComment)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultText(string(r)), nil
		}
}

// SearchIssues creates a tool to search for issues and pull requests.
func SearchIssues(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return githubgithubgithubgithubgithubgithubgithubMCP.NewTool("search_issues",
			githubgithubgithubgithubgithubgithubgithubMCP.WithDescription(t("TOOL_SEARCH_ISSUES_DESCRIPTION", "Search for issues and pull requests across GitHub repositories")),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("q",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Search query using GitHub issues search syntax"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("sort",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Sort field (comments, reactions, created, etc.)"),
				githubgithubgithubgithubgithubgithubMCP.Enum(
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
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("order",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Sort order ('asc' or 'desc')"),
				githubgithubgithubgithubgithubgithubMCP.Enum("asc", "desc"),
			),
			WithPagination(),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			query, err := requiredParam[string](request, "q")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			sort, err := OptionalParam[string](request, "sort")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			order, err := OptionalParam[string](request, "order")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
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
			result, resp, err := client.Search.getIssue(ctx, query, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to search issues: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(fmt.Sprintf("failed to search issues: %s", string(body))), nil
			}

			r, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultText(string(r)), nil
		}
}

// CreateIssue creates a tool to create a new issue in a GitHub repository.
func CreateIssue(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return githubgithubgithubgithubgithubgithubgithubMCP.NewTool("create_issue",
			githubgithubgithubgithubgithubgithubgithubMCP.WithDescription(t("TOOL_CREATE_ISSUE_DESCRIPTION", "Create a new issue in a GitHub repository")),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("owner",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Repository owner"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("repo",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Repository name"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("title",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Issue title"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("body",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Issue body content"),
			),
			githubgithubgithubgithubgithubMCP.WithArray("assignees",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Usernames to assign to this issue"),
				githubgithubgithubgithubgithubMCP.Items(
					map[string]interface{}{
						"type": "string",
					},
				),
			),
			githubgithubgithubgithubgithubMCP.WithArray("labels",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Labels to apply to this issue"),
				githubgithubgithubgithubgithubMCP.Items(
					map[string]interface{}{
						"type": "string",
					},
				),
			),
			githubgithubgithubgithubgithubgithubgithubgithubMCP.WithNumber("milestone",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Milestone number"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			title, err := requiredParam[string](request, "title")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}

			// Optional parameters
			body, err := OptionalParam[string](request, "body")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}

			// Get assignees
			assignees, err := OptionalStringArrayParam(request, "assignees")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}

			// Get labels
			labels, err := OptionalStringArrayParam(request, "labels")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}

			// Get optional milestone
			milestone, err := OptionalIntParam(request, "milestone")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}

			var milestoneNum *int
			if milestone != 0 {
				milestoneNum = &milestone
			}

			// Create the issue request
			issueRequest := &github.IssueRequest{
				Title:     githubMCP.Ptr(title),
				Body:      githubMCP.Ptr(body),
				Assignees: &assignees,
				Labels:    &labels,
				Milestone: milestoneNum,
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
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(fmt.Sprintf("failed to create issue: %s", string(body))), nil
			}

			r, err := json.Marshal(issue)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultText(string(r)), nil
		}
}

// ListIssues creates a tool to list and filter repository issues
func ListIssues(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return githubgithubgithubgithubgithubgithubgithubMCP.NewTool("list_issues",
			githubgithubgithubgithubgithubgithubgithubMCP.WithDescription(t("TOOL_LIST_ISSUES_DESCRIPTION", "List issues in a GitHub repository with filtering options")),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("owner",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Repository owner"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("repo",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Repository name"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("state",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Filter by state ('open', 'closed', 'all')"),
				githubgithubgithubgithubgithubgithubMCP.Enum("open", "closed", "all"),
			),
			githubgithubgithubgithubgithubMCP.WithArray("labels",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Filter by labels"),
				githubgithubgithubgithubgithubMCP.Items(
					map[string]interface{}{
						"type": "string",
					},
				),
			),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("sort",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Sort by ('created', 'updated', 'comments')"),
				githubgithubgithubgithubgithubgithubMCP.Enum("created", "updated", "comments"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("direction",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Sort direction ('asc', 'desc')"),
				githubgithubgithubgithubgithubgithubMCP.Enum("asc", "desc"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("since",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Filter by date (ISO 8601 timestamp)"),
			),
			WithPagination(),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}

			opts := &github.IssueListByRepoOptions{}

			// Set optional parameters if provided
			opts.State, err = OptionalParam[string](request, "state")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}

			// Get labels
			opts.Labels, err = OptionalStringArrayParam(request, "labels")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}

			opts.Sort, err = OptionalParam[string](request, "sort")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}

			opts.Direction, err = OptionalParam[string](request, "direction")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}

			since, err := OptionalParam[string](request, "since")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			if since != "" {
				timestamp, err := parseISOTimestamp(since)
				if err != nil {
					return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(fmt.Sprintf("failed to list issues: %s", err.Error())), nil
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
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(fmt.Sprintf("failed to list issues: %s", string(body))), nil
			}

			r, err := json.Marshal(issues)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal issues: %w", err)
			}

			return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultText(string(r)), nil
		}
}

// UpdateIssue creates a tool to update an existing issue in a GitHub repository.
func UpdateIssue(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return githubgithubgithubgithubgithubgithubgithubMCP.NewTool("update_issue",
			githubgithubgithubgithubgithubgithubgithubMCP.WithDescription(t("TOOL_UPDATE_ISSUE_DESCRIPTION", "Update an existing issue in a GitHub repository")),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("owner",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Repository owner"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("repo",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Repository name"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubMCP.WithNumber("issue_number",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Issue number to update"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("title",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("New title"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("body",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("New description"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("state",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("New state ('open' or 'closed')"),
				githubgithubgithubgithubgithubgithubMCP.Enum("open", "closed"),
			),
			githubgithubgithubgithubgithubMCP.WithArray("labels",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("New labels"),
				githubgithubgithubgithubgithubMCP.Items(
					map[string]interface{}{
						"type": "string",
					},
				),
			),
			githubgithubgithubgithubgithubMCP.WithArray("assignees",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("New assignees"),
				githubgithubgithubgithubgithubMCP.Items(
					map[string]interface{}{
						"type": "string",
					},
				),
			),
			githubgithubgithubgithubgithubgithubgithubgithubMCP.WithNumber("milestone",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("New milestone number"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			issueNumber, err := RequiredInt(request, "issue_number")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}

			// Create the issue request with only provided fields
			issueRequest := &github.IssueRequest{}

			// Set optional parameters if provided
			title, err := OptionalParam[string](request, "title")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			if title != "" {
				issueRequest.Title = githubMCP.Ptr(title)
			}

			body, err := OptionalParam[string](request, "body")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			if body != "" {
				issueRequest.Body = githubMCP.Ptr(body)
			}

			state, err := OptionalParam[string](request, "state")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			if state != "" {
				issueRequest.State = githubMCP.Ptr(state)
			}

			// Get labels
			labels, err := OptionalStringArrayParam(request, "labels")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			if len(labels) > 0 {
				issueRequest.Labels = &labels
			}

			// Get assignees
			assignees, err := OptionalStringArrayParam(request, "assignees")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			if len(assignees) > 0 {
				issueRequest.Assignees = &assignees
			}

			milestone, err := OptionalIntParam(request, "milestone")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
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
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(fmt.Sprintf("failed to update issue: %s", string(body))), nil
			}

			r, err := json.Marshal(updatedIssue)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultText(string(r)), nil
		}
}

// GetIssueComments creates a tool to get comments for a GitHub issue.
func GetIssueComments(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return githubgithubgithubgithubgithubgithubgithubMCP.NewTool("get_issue_comments",
			githubgithubgithubgithubgithubgithubgithubMCP.WithDescription(t("TOOL_GET_ISSUE_COMMENTS_DESCRIPTION", "Get comments for a GitHub issue")),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("owner",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Repository owner"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.WithString("repo",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Repository name"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubMCP.WithNumber("issue_number",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Issue number"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubMCP.WithNumber("page",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Page number"),
			),
			githubgithubgithubgithubgithubgithubgithubgithubMCP.WithNumber("per_page",
				githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.Description("Number of records per page"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			repo, err := requiredParam[string](request, "repo")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			issueNumber, err := RequiredInt(request, "issue_number")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			page, err := OptionalIntParamWithDefault(request, "page", 1)
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			perPage, err := OptionalIntParamWithDefault(request, "per_page", 30)
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
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
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(fmt.Sprintf("failed to get issue comments: %s", string(body))), nil
			}

			r, err := json.Marshal(comments)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultText(string(r)), nil
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

