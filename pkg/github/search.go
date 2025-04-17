package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v69/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// SearchRepositories creates a tool to search for GitHub repositories.
func SearchRepositories(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return githubgithubgithubMCP.NewTool("search_repositories",
			githubgithubgithubMCP.WithDescription(t("TOOL_SEARCH_REPOSITORIES_DESCRIPTION", "Search for GitHub repositories")),
			githubgithubgithubgithubgithubgithubgithubMCP.WithString("query",
				githubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubMCP.Description("Search query"),
			),
			WithPagination(),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			query, err := requiredParam[string](request, "query")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}

			opts := &github.SearchOptions{
				ListOptions: github.ListOptions{
					Page:    pagination.page,
					PerPage: pagination.perPage,
				},
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			result, resp, err := client.Search.Repositories(ctx, query, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to search repositories: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != 200 {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(fmt.Sprintf("failed to search repositories: %s", string(body))), nil
			}

			r, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return githubgithubgithubgithubgithubgithubMCP.NewToolResultText(string(r)), nil
		}
}

// SearchCode creates a tool to search for code across GitHub repositories.
func SearchCode(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return githubgithubgithubMCP.NewTool("search_code",
			githubgithubgithubMCP.WithDescription(t("TOOL_SEARCH_CODE_DESCRIPTION", "Search for code across GitHub repositories")),
			githubgithubgithubgithubgithubgithubgithubMCP.WithString("q",
				githubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubMCP.Description("Search query using GitHub code search syntax"),
			),
			githubgithubgithubgithubgithubgithubgithubMCP.WithString("sort",
				githubgithubgithubgithubgithubgithubgithubMCP.Description("Sort field ('indexed' only)"),
			),
			githubgithubgithubgithubgithubgithubgithubMCP.WithString("order",
				githubgithubgithubgithubgithubgithubgithubMCP.Description("Sort order ('asc' or 'desc')"),
				githubgithubgithubMCP.Enum("asc", "desc"),
			),
			WithPagination(),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			query, err := requiredParam[string](request, "q")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			sort, err := OptionalParam[string](request, "sort")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			order, err := OptionalParam[string](request, "order")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
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

			result, resp, err := client.Search.Code(ctx, query, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to search code: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != 200 {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(fmt.Sprintf("failed to search code: %s", string(body))), nil
			}

			r, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return githubgithubgithubgithubgithubgithubMCP.NewToolResultText(string(r)), nil
		}
}

// SearchUsers creates a tool to search for GitHub users.
func SearchUsers(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return githubgithubgithubMCP.NewTool("search_users",
			githubgithubgithubMCP.WithDescription(t("TOOL_SEARCH_USERS_DESCRIPTION", "Search for GitHub users")),
			githubgithubgithubgithubgithubgithubgithubMCP.WithString("q",
				githubgithubgithubMCP.Required(),
				githubgithubgithubgithubgithubgithubgithubMCP.Description("Search query using GitHub users search syntax"),
			),
			githubgithubgithubgithubgithubgithubgithubMCP.WithString("sort",
				githubgithubgithubgithubgithubgithubgithubMCP.Description("Sort field (followers, repositories, joined)"),
				githubgithubgithubMCP.Enum("followers", "repositories", "joined"),
			),
			githubgithubgithubgithubgithubgithubgithubMCP.WithString("order",
				githubgithubgithubgithubgithubgithubgithubMCP.Description("Sort order ('asc' or 'desc')"),
				githubgithubgithubMCP.Enum("asc", "desc"),
			),
			WithPagination(),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			query, err := requiredParam[string](request, "q")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			sort, err := OptionalParam[string](request, "sort")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			order, err := OptionalParam[string](request, "order")
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
			}
			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(err.Error()), nil
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

			result, resp, err := client.Search.getUser(ctx, query, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to search users: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != 200 {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return githubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubgithubMCP.NewToolResultError(fmt.Sprintf("failed to search users: %s", string(body))), nil
			}

			r, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return githubgithubgithubgithubgithubgithubMCP.NewToolResultText(string(r)), nil
		}
}

