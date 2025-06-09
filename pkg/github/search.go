// WHO: GitHubMCPBridge
// WHAT: GitHub API Integration Package
// WHEN: MCP bridge initialization
// WHERE: MCP Bridge Layer
// WHY: To provide API access to GitHub
// HOW: By implementing MCP protocol handlers
// EXTENT: All GitHub API operations
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/jubicudis/github-mcp-server/pkg/common"
	"github.com/jubicudis/github-mcp-server/pkg/translations"

	"github.com/google/go-github/v71/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Canonical search logic for GitHub MCP server
// Remove all stubs, placeholders, and incomplete logic
// All types and methods must be robust, DRY, and reference only canonical helpers from /pkg/common
// All search and event logic must be fully implemented

// WHO: SearchRepositoriesTool
// WHAT: GitHub Repository Search
// WHEN: Tool invocation
// WHERE: GitHub MCP Server
// WHY: To find repositories matching search criteria
// HOW: By querying GitHub Search API
// EXTENT: All public and authorized GitHub repositories
func SearchRepositories(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("search_repositories",
			mcp.WithDescription(t("TOOL_SEARCH_REPOSITORIES_DESCRIPTION", "Search for GitHub repositories")),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("Search query"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			query, err := common.RequiredParam[string](request, "query")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			page, err := common.OptionalIntParamWithDefault(request, "page", 1)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			perPage, err := common.OptionalIntParamWithDefault(request, "perPage", 30)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			opts := &github.SearchOptions{
				ListOptions: github.ListOptions{Page: page, PerPage: perPage},
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
				return mcp.NewToolResultError(fmt.Sprintf("failed to search repositories: %s", string(body))), nil
			}

			r, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// SearchCode creates a tool to search for code across GitHub repositories.
func SearchCode(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("search_code",
			mcp.WithDescription(t("TOOL_SEARCH_CODE_DESCRIPTION", "Search for code across GitHub repositories")),
			mcp.WithString("q",
				mcp.Required(),
				mcp.Description("Search query using GitHub code search syntax"),
			),
			mcp.WithString("sort",
				mcp.Description("Sort field ('indexed' only)"),
			),
			mcp.WithString("order",
				mcp.Description("Sort order ('asc' or 'desc')"),
				mcp.Enum("asc", "desc"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
			perPage, err := common.OptionalIntParamWithDefault(request, "perPage", 30)
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
				return mcp.NewToolResultError(fmt.Sprintf("failed to search code: %s", string(body))), nil
			}

			r, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// SearchUsers creates a tool to search for GitHub users.
func SearchUsers(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("search_users",
			mcp.WithDescription(t("TOOL_SEARCH_USERS_DESCRIPTION", "Search for GitHub users")),
			mcp.WithString("q",
				mcp.Required(),
				mcp.Description("Search query using GitHub users search syntax"),
			),
			mcp.WithString("sort",
				mcp.Description("Sort field (followers, repositories, joined)"),
				mcp.Enum("followers", "repositories", "joined"),
			),
			mcp.WithString("order",
				mcp.Description("Sort order ('asc' or 'desc')"),
				mcp.Enum("asc", "desc"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
			perPage, err := common.OptionalIntParamWithDefault(request, "perPage", 30)
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

			result, resp, err := client.Search.Users(ctx, query, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to search users: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != 200 {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to search users: %s", string(body))), nil
			}

			r, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}
