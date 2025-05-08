/*
 * WHO: GitHubRepositoryHandler
 * WHAT: Repository operations for GitHub MCP server
 * WHEN: During repository API operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide repository access via MCP
 * HOW: Using GitHub API with MCP protocol adapters
 * EXTENT: All repository operations
 */

package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v69/github"
	"github.com/mark3labs/mcp-go/log"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// WHO: RepositoryToolProvider
// WHAT: Get repository tool definition
// WHEN: During tool registration
// WHERE: System Layer 6 (Integration)
// WHY: To provide repository information via MCP
// HOW: Using MCP tool definition mechanism
func GetRepository(getClient GetClientFn, t TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	// WHO: RepositoryToolDefiner
	// WHAT: Define repository tool
	// WHEN: During server initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To register tool with MCP
	// HOW: Using MCP tool definition
	// EXTENT: Repository tool interface

	return mcp.NewTool("get_repository",
		mcp.WithDescription("Gets detailed information about a GitHub repository"),
		mcp.WithString("owner", 
			mcp.Required(),
			mcp.Description(ConstRepoOwnerDesc),
		),
		mcp.WithString("repo",
			mcp.Required(),
			mcp.Description(ConstRepoNameDesc),
		),
	),
		ReturnSchema: map[string]interface{}{
			"name":        "string",
			"full_name":   "string",
			"description": "string",
			"owner": map[string]interface{}{
				"login": "string",
				"id":    "number",
			},
			"html_url":       "string",
			"language":       "string",
			"fork":           "boolean",
			"stargazers":     "number",
			"watchers":       "number",
			"forks":          "number",
			"open_issues":    "number",
			"default_branch": "string",
			"created_at":     "string",
			"updated_at":     "string",
			"archived":       "boolean",
			"private":        "boolean",
		},
	}

	// Handler function
	handler := func(ctx context.Context, request mcp.CallToolRequest) (interface{}, error) {
		// WHO: RepositoryRequestHandler
		// WHAT: Handle repository request
		// WHEN: During tool invocation
		// WHERE: System Layer 6 (Integration)
	// Handler function
	func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// EXTENT: Repository API operations

		log.Debug("Handling get_repository request")

		// Extract parameters
		owner, err := ExtractRequiredParam[string](request, "owner")
		if err != nil {
			return nil, fmt.Errorf("invalid owner parameter: %w", err)
		}

		repo, err := ExtractRequiredParam[string](request, "repo")
		if err != nil {
			return nil, fmt.Errorf("invalid repo parameter: %w", err)
		}

		// Get GitHub client
		client, err := getClient(ctx)
		if err != nil {
			return nil, fmt.Errorf(ErrMsgGetGitHubClient, err)
		}

		// Apply context translation if provided
		if t != nil {
			contextData := map[string]interface{}{
				"who":    "RepositoryHandler",
				"what":   "GetRepository",
				"when":   "APIRequest",
				"where":  "GitHub_API",
				"why":    "FetchRepositoryData",
				"how":    "HTTPRequest",
				"extent": "SingleRepository",
			}

			translatedContext, err := t(ctx, contextData)
			if err != nil {
				log.Warn("Context translation failed:", err)
			} else if translatedContext != nil {
				log.Debug("Using translated context for repository request")
				// Context would be used here in a real implementation
			}
		}

		// Call GitHub API
		repository, _, err := client.Repositories.Get(ctx, owner, repo)
		if err != nil {
			return nil, fmt.Errorf("failed to get repository: %w", err)
		}

		// Transform to response format
		response := map[string]interface{}{
			"name":        repository.GetName(),
			"full_name":   repository.GetFullName(),
			"description": repository.GetDescription(),
			"owner": map[string]interface{}{
				"login": repository.Owner.GetLogin(),
				"id":    repository.Owner.GetID(),
			},
			"html_url":       repository.GetHTMLURL(),
			"language":       repository.GetLanguage(),
			"fork":           repository.GetFork(),
			"stargazers":     repository.GetStargazersCount(),
			"watchers":       repository.GetWatchersCount(),
			"forks":          repository.GetForksCount(),
			"open_issues":    repository.GetOpenIssuesCount(),
			"default_branch": repository.GetDefaultBranch(),
			"created_at":     repository.GetCreatedAt().String(),
			"updated_at":     repository.GetUpdatedAt().String(),
			"archived":       repository.GetArchived(),
			"private":        repository.GetPrivate(),
		}

		return response, nil
	}

	return tool, handler
}

		return mcp.NewToolResult(response), nil
	}
}
// HOW: Using MCP tool definition mechanism
// EXTENT: Repository listing operations
func ListRepositories(getClient GetClientFn, t TranslationHelperFunc) (mcp.ToolDefinition, mcp.CallToolHandlerFunc) {
	// WHO: RepositoryListToolDefiner
	// WHAT: Define list repositories tool
	// WHEN: During server initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To register tool with MCP
	// HOW: Using MCP tool definition
	// EXTENT: Repository list tool interface
func ListRepositories(getClient GetClientFn, t TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	// WHO: RepositoryListToolDefiner
	// WHAT: Define list repositories tool
	// WHEN: During server initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To register tool with MCP
	// HOW: Using MCP tool definition
	// EXTENT: Repository list tool interface

	return mcp.NewTool("list_repositories",
		mcp.WithDescription("Lists GitHub repositories for a user or organization"),
		mcp.WithString("owner",
			mcp.Required(),
			mcp.Description("Username or organization name"),
		),
		mcp.WithString("type",
			mcp.Description("Type of repositories to list (all, owner, member, public, private)"),
		),
		mcp.WithString("sort",
			mcp.Description("Sorting criteria (created, updated, pushed, full_name)"),
		),
		mcp.WithString("direction",
			mcp.Description("Sort direction (asc, desc)"),
		),
		mcp.WithInt("per_page",
			mcp.Description("Number of results per page"),
		),
		mcp.WithInt("page",
			mcp.Description("Page number for pagination"),
		),
	),
				Description: "Type of repositories to list (all, owner, member, public, private)",
				Type:        "string",
				Required:    false,
			},
			{
				Name:        "sort",
				Description: "Sorting criteria (created, updated, pushed, full_name)",
				Type:        "string",
				Required:    false,
			},
			{
				Name:        "direction",
				Description: "Sort direction (asc, desc)",
				Type:        "string",
				Required:    false,
			},
			{
				Name:        "per_page",
				Description: "Number of results per page",
				Type:        "number",
				Required:    false,
			},
			{
				Name:        "page",
				Description: "Page number for pagination",
				Type:        "number",
				Required:    false,
			},
		},
		ReturnSchema: map[string]interface{}{
			"repositories": []map[string]interface{}{
				{
					"name":        "string",
					"full_name":   "string",
					"description": "string",
					"html_url":    "string",
					"language":    "string",
					"private":     "boolean",
				},
			},
			"total_count": "number",
		},
	}

	// Handler function
	handler := func(ctx context.Context, request mcp.CallToolRequest) (interface{}, error) {
		// WHO: RepositoryListRequestHandler
		// WHAT: Handle repository list request
		// WHEN: During tool invocation
		// WHERE: System Layer 6 (Integration)
		// WHY: To process repository listing
		// HOW: Using GitHub API client
		// EXTENT: Repository listing operations

		log.Debug("Handling list_repositories request")

		// Extract parameters
		owner, err := ExtractRequiredParam[string](request, "owner")
		if err != nil {
			return nil, fmt.Errorf("invalid owner parameter: %w", err)
		}

		// Optional parameters
		repoType, _ := OptionalParam[string](request, "type")
		sort, _ := OptionalParam[string](request, "sort")
		direction, _ := OptionalParam[string](request, "direction")
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
			return nil, fmt.Errorf(ErrMsgGetGitHubClient, err)
		}

		// Prepare list options
		opts := &github.RepositoryListOptions{}
		if repoType != "" {
			opts.Type = repoType
		}
		if sort != "" {
			opts.Sort = sort
		}
		if direction != "" {
			opts.Direction = direction
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

		// Apply context translation if provided
		if t != nil {
			contextData := map[string]interface{}{
				"who":    "RepositoryListHandler",
				"what":   "ListRepositories",
				"when":   "APIRequest",
				"where":  "GitHub_API",
				"why":    "FetchRepositoryList",
				"how":    "HTTPRequest",
				"extent": "MultipleRepositories",
			}

			translatedContext, err := t(ctx, contextData)
			if err != nil {
				log.Warn("Context translation failed:", err)
			} else if translatedContext != nil {
				log.Debug("Using translated context for repository listing")
				// Context would be used here in a real implementation
			}
		}

		// Call GitHub API
		repositories, _, err := client.Repositories.List(ctx, owner, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}

		// Transform to response format
		repoList := make([]map[string]interface{}, 0, len(repositories))
		for _, repo := range repositories {
			repoList = append(repoList, map[string]interface{}{
				"name":        repo.GetName(),
				"full_name":   repo.GetFullName(),
				"description": repo.GetDescription(),
				"html_url":    repo.GetHTMLURL(),
				"language":    repo.GetLanguage(),
				"private":     repo.GetPrivate(),
			})
		}

		response := map[string]interface{}{
			"repositories": repoList,
			"total_count":  len(repoList),
		}

		return response, nil
	}

	return tool, handler
}
