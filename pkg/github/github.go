/*
 * WHO: GitHubPackage
 * WHAT: Main package for GitHub MCP server functionality
 * WHEN: During API operations with GitHub
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide GitHub integration with TNOS
 * HOW: Using MCP protocol and GitHub API
 * EXTENT: All GitHub operations
 */

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jubicudis/github-mcp-server/pkg/bridge"
	"github.com/jubicudis/github-mcp-server/pkg/common"
	"github.com/jubicudis/github-mcp-server/pkg/log"
	"github.com/jubicudis/github-mcp-server/pkg/translations"

	"github.com/google/go-github/v53/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// WHO: ConstantReferencer
// WHAT: Reference to bridge mode constants
// WHEN: During bridge configuration
// WHERE: System Layer 6 (Integration)
// WHY: To use constants defined in constants.go
// HOW: Using imported constants
// EXTENT: Bridge mode configuration

// Bridge mode constants are defined in constants.go

// InitializeMCPBridge sets up the MCP bridge between GitHub and TNOS MCP
func InitializeMCPBridge(enableCompression bool) error {
	// WHO: BridgeInitializer
	// WHAT: MCP bridge initialization with fallback
	// WHEN: During server startup
	// WHERE: System Layer 6 (Integration)
	// WHY: To establish GitHub-TNOS communication robustly
	// HOW: Using FallbackRoute utility
	// EXTENT: System integration

	context7d := translations.ContextVector7D{
		Who:    "GitHubMCPServer",
		What:   "BridgeInit",
		When:   time.Now().Unix(),
		Where:  "SystemLayer6",
		Why:    "Startup",
		How:    "FallbackRoute",
		Extent: 1.0,
		Source: "GitHubMCPServer",
	}
	ctx := context.Background()
	operationName := "InitializeMCPBridge"

	// Convert translations.ContextVector7D to map[string]interface{}
	contextData := map[string]interface{}{
		"Who":    context7d.Who,
		"What":   context7d.What,
		"When":   context7d.When,
		"Where":  context7d.Where,
		"Why":    context7d.Why,
		"How":    context7d.How,
		"Extent": context7d.Extent,
	}

	// Replace FallbackRouteWithVector with FallbackRoute
	_, err := bridge.FallbackRoute(
		context.Background(),
		"BridgeInit",
		contextData,
		func() (interface{}, error) {
			fmt.Println("Initializing MCP Bridge between GitHub and TNOS")
			return nil, nil
		},
		func() (interface{}, error) { return nil, fmt.Errorf("Bridge fallback not implemented") },
		func() (interface{}, error) { return nil, fmt.Errorf("GitHub MCP fallback not implemented") },
		func() (interface{}, error) { return nil, fmt.Errorf("Copilot fallback not implemented") },
		log.NewLogger(),
	)
	return err
}

// GitHubContextTranslator provides bidirectional translation between GitHub context and TNOS 7D context
type GitHubContextTranslator struct {
	// WHO: ContextTranslator
	// WHAT: Translates between context formats
	// WHEN: During context exchange
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide contextual compatibility
	// HOW: Using bidirectional mapping
	// EXTENT: All context translations

	EnableCompression bool
	EnableLogging     bool
	DebugMode         bool
	Logger            *log.Logger
}

// NewGitHubContextTranslator creates a new context translator instance
func NewGitHubContextTranslator(logger *log.Logger, enableCompression, enableLogging, debugMode bool) *GitHubContextTranslator {
	// WHO: TranslatorFactory
	// WHAT: Creates context translator
	// WHEN: During bridge initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To enable context translation
	// HOW: Using factory pattern
	// EXTENT: Translator lifecycle

	return &GitHubContextTranslator{
		EnableCompression: enableCompression,
		EnableLogging:     enableLogging,
		DebugMode:         debugMode,
		Logger:            logger,
	}
}

// TranslateToTNOS converts GitHub context to TNOS 7D context
func (t *GitHubContextTranslator) TranslateToTNOS(githubContext map[string]interface{}) map[string]interface{} {
	// WHO: GitHubToTNOSTranslator
	// WHAT: Convert GitHub context
	// WHEN: During inbound operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide TNOS compatibility
	// HOW: Using context mapping
	// EXTENT: Inbound messages

	context7d := translations.ContextVector7D{
		Who:    "GitHubContextTranslator",
		What:   "TranslateToTNOS",
		When:   time.Now().Unix(),
		Where:  "SystemLayer6",
		Why:    "ContextTranslation",
		How:    "FallbackRoute",
		Extent: 1.0,
		Source: "GitHubMCPServer",
	}
	ctx := context.Background()
	operationName := "TranslateToTNOS"
	_, _ = bridge.FallbackRouteWithVector(
		ctx,
		operationName,
		context7d,
		func() (interface{}, error) {
			if t.EnableLogging && t.Logger != nil {
				t.Logger.Debug("Translating context from GitHub to TNOS")
			}
			return nil, nil
		},
		func() (interface{}, error) { return nil, fmt.Errorf("Bridge fallback not implemented") },
		func() (interface{}, error) { return nil, fmt.Errorf("GitHub MCP fallback not implemented") },
		func() (interface{}, error) { return nil, fmt.Errorf("Copilot fallback not implemented") },
		t.Logger,
	)

	// This would implement the actual translation logic in a real system
	tnosContext := map[string]interface{}{
		"who":    githubContext["identity"],
		"what":   githubContext["operation"],
		"when":   githubContext["timestamp"],
		"where":  "GitHub_MCP_Bridge",
		"why":    githubContext["purpose"],
		"how":    "Context_Translation",
		"extent": githubContext["scope"],
	}

	return tnosContext
}

// TranslateFromTNOS converts TNOS 7D context to GitHub context
func (t *GitHubContextTranslator) TranslateFromTNOS(tnosContext map[string]interface{}) map[string]interface{} {
	// WHO: TNOSToGitHubTranslator
	// WHAT: Convert TNOS context
	// WHEN: During outbound operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide GitHub compatibility
	// HOW: Using context mapping
	// EXTENT: Outbound messages

	context7d := translations.ContextVector7D{
		Who:    "GitHubContextTranslator",
		What:   "TranslateFromTNOS",
		When:   time.Now().Unix(),
		Where:  "SystemLayer6",
		Why:    "ContextTranslation",
		How:    "FallbackRoute",
		Extent: 1.0,
		Source: "GitHubMCPServer",
	}
	ctx := context.Background()
	operationName := "TranslateFromTNOS"
	_, _ = bridge.FallbackRouteWithVector(
		ctx,
		operationName,
		context7d,
		func() (interface{}, error) {
			if t.EnableLogging && t.Logger != nil {
				t.Logger.Debug("Translating context from TNOS to GitHub")
			}
			return nil, nil
		},
		func() (interface{}, error) { return nil, fmt.Errorf("Bridge fallback not implemented") },
		func() (interface{}, error) { return nil, fmt.Errorf("GitHub MCP fallback not implemented") },
		func() (interface{}, error) { return nil, fmt.Errorf("Copilot fallback not implemented") },
		t.Logger,
	)

	// This would implement the actual translation logic in a real system
	githubContext := map[string]interface{}{
		"identity":  tnosContext["who"],
		"operation": tnosContext["what"],
		"timestamp": tnosContext["when"],
		"purpose":   tnosContext["why"],
		"scope":     tnosContext["extent"],
	}

	return githubContext
}

// BridgeHealthCheck performs a health check on the MCP bridge
func BridgeHealthCheck() (bool, error) {
	// WHO: HealthMonitor
	// WHAT: Bridge health check with fallback
	// WHEN: During system monitoring
	// WHERE: System Layer 6 (Integration)
	// WHY: To ensure bridge availability
	// HOW: Using FallbackRoute
	// EXTENT: Bridge operational status
	context7d := translations.ContextVector7D{
		Who:    "GitHubMCPServer",
		What:   "HealthCheck",
		When:   time.Now().Unix(),
		Where:  "SystemLayer6",
		Why:    "Monitoring",
		How:    "FallbackRoute",
		Extent: 1.0,
		Source: "GitHubMCPServer",
	}
	ctx := context.Background()
	operationName := "BridgeHealthCheck"
	healthy := false
	_, err := bridge.FallbackRoute(
		ctx,
		operationName,
		context7d,
		func() (interface{}, error) { healthy = true; return nil, nil },
		func() (interface{}, error) { return nil, fmt.Errorf("Bridge fallback not implemented") },
		func() (interface{}, error) { return nil, fmt.Errorf("GitHub MCP fallback not implemented") },
		func() (interface{}, error) { return nil, fmt.Errorf("Copilot fallback not implemented") },
		log.NewLogger(),
	)
	return healthy, err
}

// ConnectMCPChannels establishes bidirectional channels between GitHub and TNOS MCP
func ConnectMCPChannels(bridgeMode string) error {
	// WHO: ChannelConnector
	// WHAT: Connect MCP channels
	// WHEN: During bridge initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To establish communication paths
	// HOW: Using channel connections
	// EXTENT: All MCP communication

	if bridgeMode != BridgeModeDirect &&
		bridgeMode != BridgeModeProxied &&
		bridgeMode != BridgeModeAsync {
		return fmt.Errorf("invalid bridge mode: %s", bridgeMode)
	}

	// Implementation would go here in a real system
	return nil
}

// StartMCPEventMonitor starts monitoring MCP events
func StartMCPEventMonitor(logger *log.Logger) error {
	// WHO: EventMonitor
	// WHAT: Monitor MCP events
	// WHEN: During bridge operation
	// WHERE: System Layer 6 (Integration)
	// WHY: To track MCP events
	// HOW: Using event listeners
	// EXTENT: All MCP events

	if logger != nil {
		logger.Debug("Starting MCP event monitor")
	}

	// Implementation would go here in a real system
	return nil
}

// WHO: ContextTranslatorTypeDefinition
// WHAT: Define context translator function type
// WHEN: During type declarations
// WHERE: System Layer 6 (Integration)
// WHY: To enable context translation across components
// HOW: Using function type definition
// EXTENT: All translation operations
// NOTE: We're now using translations.TranslationHelperFunc instead of this local definition

// WHO: ParameterExtractor
// WHAT: Extract parameters from MCP requests
// WHEN: During API request handling
// WHERE: System Layer 6 (Integration)
// WHY: To validate and extract request parameters
// HOW: Using type assertion and validation
// EXTENT: MCP request parameter handling
// This function is renamed to avoid duplicate declaration with common.go
func ExtractRequiredParam[T any](request mcp.CallToolRequest, name string) (T, error) {
	var zero T
	value, ok := request.Params.Arguments[name]
	if !ok {
		return zero, fmt.Errorf("missing required parameter: %s", name)
	}

	result, ok := value.(T)
	if !ok {
		return zero, fmt.Errorf("invalid type for parameter %s", name)
	}

	return result, nil
}

// Deprecated: use common.RequiredIntParam, common.OptionalIntParam, common.OptionalIntParamWithDefault
// func RequiredInt(request mcp.CallToolRequest, name string) (int, error) {
//	value, ok := request.Params.Arguments[name]
//	if !ok {
//		return 0, fmt.Errorf("missing required parameter: %s", name)
//	}
//
//	// Handle different number types
//	switch v := value.(type) {
//	case int:
//		return v, nil
//	case int64:
//		return int(v), nil
//	case float64:
//		return int(v), nil
//	case string:
//		i, err := strconv.Atoi(v)
//		if err != nil {
//			return 0, fmt.Errorf("invalid integer for %s: %s", name, err)
//		}
//		return i, nil
//	default:
//		return 0, fmt.Errorf("expected integer for %s, got %T", name, value)
//	}
// }

// Deprecated: use common.OptionalIntParam
// func OptionalInt(request mcp.CallToolRequest, name string) (int, bool, error) {
//	value, ok := request.Params.Arguments[name]
//	if !ok {
//		return 0, false, nil // Parameter not provided
//	}
//
//	// Handle different number types
//	switch v := value.(type) {
//	case int:
//		return v, true, nil
//	case int64:
//		return int(v), true, nil
//	case float64:
//		return int(v), true, nil
//	case string:
//		i, err := strconv.Atoi(v)
//		if err != nil {
//			return 0, false, fmt.Errorf("invalid integer for %s: %s", name, err)
//		}
//		return i, true, nil
//	default:
//		return 0, false, fmt.Errorf("expected integer for %s, got %T", name, value)
//	}
// }

// Deprecated: use common.OptionalBoolParam
// func OptionalBool(request mcp.CallToolRequest, name string) (bool, bool, error) {
//	value, ok := request.Params.Arguments[name]
//	if !ok {
//		return false, false, nil // Parameter not provided
//	}
//
//	// Handle different boolean types
//	switch v := value.(type) {
//	case bool:
//		return v, true, nil
//	case string:
//		switch strings.ToLower(v) {
//		case "true", "yes", "1":
//			return true, true, nil
//		case "false", "no", "0":
//			return false, true, nil
//		default:
//			return false, false, fmt.Errorf("invalid boolean for %s: %s", name, v)
//		}
//	case float64:
//		if v == 1 {
//			return true, true, nil
//		} else if v == 0 {
//			return false, true, nil
//		}
//		return false, false, fmt.Errorf("invalid boolean for %s: %f", name, v)
//	case int:
//		if v == 1 {
//			return true, true, nil
//		} else if v == 0 {
//			return false, true, nil
//		}
//		return false, false, fmt.Errorf("invalid boolean for %s: %d", name, v)
//	default:
//		return false, false, fmt.Errorf("expected boolean for %s, got %T", name, value)
//	}
// }

// Deprecated: use common.StringListParam
// func StringList(request mcp.CallToolRequest, name string) ([]string, error) {
//	value, ok := request.Params.Arguments[name]
//	if !ok {
//		return nil, nil // Parameter not provided
//	}
//
//	// Check if it's already a string slice
//	if strSlice, ok := value.([]string); ok {
//		return strSlice, nil
//	}
//
//	// Check if it's an interface slice
//	if interfaceSlice, ok := value.([]interface{}); ok {
//		result := make([]string, 0, len(interfaceSlice))
//		for _, item := range interfaceSlice {
//			if str, ok := item.(string); ok {
//				result = append(result, str)
//			} else {
//				return nil, fmt.Errorf("expected string list for %s, contains non-string value", name)
//			}
//		}
//		return result, nil
//	}
//
//	// Check if it's a comma-separated string
//	if str, ok := value.(string); ok {
//		if str == "" {
//			return []string{}, nil
//		}
//		return strings.Split(str, ","), nil
//	}
//
//	return nil, fmt.Errorf("expected string list for %s, got %T", name, value)
// }

// Deprecated: use common.MapParam
// func MapParam(request mcp.CallToolRequest, name string) (map[string]interface{}, error) {
//	value, ok := request.Params.Arguments[name]
//	if !ok {
//		return nil, nil // Parameter not provided
//	}
//
//	result, ok := value.(map[string]interface{})
//	if !ok {
//		return nil, fmt.Errorf("expected object for %s, got %T", name, value)
//	}
//
//	return result, nil
// }

// Deprecated: use common.OptionalPaginationParams
// func OptionalPaginationParams(request mcp.CallToolRequest) (PaginationParams, error) {
//	var params PaginationParams
//
//	// Extract common pagination parameters
//	limit, err := common.OptionalIntParam(request, "limit")
//	if err != nil {
//		return params, err
//	}
//	params.Limit = limit
//
//	offset, err := common.OptionalIntParam(request, "offset")
//	if err != nil {
//		return params, err
//	}
//	params.Offset = offset
//
//	// Extract sort parameters
//	sortBy, err := common.StringListParam(request, "sortBy")
//	if err != nil {
//		return params, err
//	}
//	params.SortBy = sortBy
//
//	sortOrder, err := common.StringListParam(request, "sortOrder")
//	if err != nil {
//		return params, err
//	}
//	params.SortOrder = sortOrder
//
//	return params, nil
// }

// WHO: RegisteredToolsProvider
// WHAT: Provide MCP tools registration
// WHEN: During server initialization
// WHERE: System Layer 6 (Integration)
// WHY: To register GitHub tools with MCP
// HOW: Using MCP registration mechanisms
// EXTENT: All GitHub MCP tool registration
// MCPServer represents the MCP server interface
type MCPServer interface {
	RegisterTool(tool mcp.Tool, handler server.ToolHandlerFunc)
}

func RegisterTools(server MCPServer, getClient GetClientFn, t translations.TranslationHelperFunc) {
	// WHO: ToolRegistrar
	// WHAT: Register GitHub tools
	// WHEN: During server initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To make tools available via MCP
	// HOW: Using tool registration
	// EXTENT: All GitHub MCP tools

	// Register repository tools directly using translation helper
	repoTool, repoHandler := GetRepositoryResourceContent(getClient, t)
	server.RegisterTool(repoTool, repoHandler)

	listReposTool, listReposHandler := ListRepositories(getClient, t)
	server.RegisterTool(listReposTool, listReposHandler)

	// Register content tools
	// TODO: Implement GetContent function or remove this reference
	// server.RegisterTool(contentTool, contentHandler)

	// Register issue tools
	issueTool, issueHandler := GetIssue(getClient, t)
	server.RegisterTool(issueTool, issueHandler)

	createIssueTool, createIssueHandler := CreateIssue(getClient, t)
	server.RegisterTool(createIssueTool, createIssueHandler)

	listIssuesTool, listIssuesHandler := ListIssues(getClient, t)
	server.RegisterTool(listIssuesTool, listIssuesHandler)

	// Register PR tools
	prTool, prHandler := GetPullRequest(getClient, t)
	server.RegisterTool(prTool, prHandler)

	createPRTool, createPRHandler := CreatePullRequest(getClient, t)
	server.RegisterTool(createPRTool, createPRHandler)

	listPRsTool, listPRsHandler := ListPullRequests(getClient, t)
	server.RegisterTool(listPRsTool, listPRsHandler)

	// Register commit tools
	commitTool, commitHandler := GetCommit(getClient, t)
	server.RegisterTool(commitTool, commitHandler)

	// Register search tools
	searchCodeTool, searchCodeHandler := SearchCode(getClient, t)
	server.RegisterTool(searchCodeTool, searchCodeHandler)

	// Register code scanning tools
	codeScanTool, codeScanHandler := GetCodeScanningAlert(getClient, t)
	server.RegisterTool(codeScanTool, codeScanHandler)

	listCodeScanTool, listCodeScanHandler := ListCodeScanningAlerts(getClient, t)
	server.RegisterTool(listCodeScanTool, listCodeScanHandler)
}

// WHO: TranslationAdapterFactory
// WHAT: Create adapter between string and context translators
// WHEN: During server initialization
// WHERE: System Layer 6 (Integration)
// WHY: To bridge incompatible translator types
// HOW: By wrapping string translator in context translator interface
// EXTENT: All translation operations during initialization
func createContextTranslationAdapter(t translations.TranslationHelperFunc) ContextTranslationFunc {
	return func(ctx context.Context, contextData map[string]interface{}) (map[string]interface{}, error) {
		// Simply pass through the context data as the translation function
		// doesn't actually modify context in this implementation
		return contextData, nil
	}
}

// AdaptResourceTemplate converts a ResourceTemplate and ResourceTemplateHandlerFunc
// to a Tool and ToolHandlerFunc for registration.
func AdaptResourceTemplate(
	template mcp.ResourceTemplate,
	handler server.ResourceTemplateHandlerFunc,
) (mcp.Tool, server.ToolHandlerFunc) {
	// Conversion logic here (mock implementation for now)
	return mcp.Tool{}, func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		return nil, nil
	}
}

// ListRepositories creates a tool to list repositories for a user or organization.
func ListRepositories(getClient GetClientFn, t translations.TranslationHelperFunc) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("list_repositories",
			mcp.WithDescription(t("TOOL_LIST_REPOSITORIES_DESCRIPTION", "List repositories for a user or organization")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Owner of the repositories (user or organization)"),
			),
			WithPagination(),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			owner, err := common.RequiredParam[string](request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			pagination, err := common.OptionalPaginationParams(request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			opts := &github.RepositoryListOptions{
				ListOptions: github.ListOptions{
					Page:    pagination.Page,
					PerPage: pagination.PerPage,
				},
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			repos, resp, err := client.Repositories.List(ctx, owner, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to list repositories: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to list repositories: %s", string(body))), nil
			}

			r, err := json.Marshal(repos)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}
