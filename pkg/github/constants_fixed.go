/*
 * WHO: GitHubConstants
 * WHAT: Centralized constants for GitHub MCP server
 * WHEN: During GitHub API operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide consistent API endpoint definitions
 * HOW: Using Go constants with 7D documentation
 * EXTENT: All GitHub API operations
 */

package github

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/go-github/v71/github"
	"github.com/mark3labs/mcp-go/mcp"
)

// API endpoint constants
const (
	// WHO: EndpointManager
	// WHAT: GitHub API endpoints
	// WHEN: During API requests
	// WHERE: System Layer 6 (Integration)
	// WHY: To define API locations
	// HOW: Using string constants
	// EXTENT: All GitHub API locations

	BaseURL               = "https://api.github.com"
	ReposEndpoint         = "/repos"
	IssuesEndpoint        = "/issues"
	PullsEndpoint         = "/pulls"
	SearchEndpoint        = "/search"
	CodeScanEndpoint      = "/code-scanning"
	UserEndpoint          = "/user"
	OrganizationsEndpoint = "/organizations"
	TeamsEndpoint         = "/teams"
)

// Version information
const (
	// WHO: VersionManager
	// WHAT: Version constants
	// WHEN: During protocol operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To ensure protocol compatibility
	// HOW: Using string constants
	// EXTENT: All versioned components

	APIVersion      = "2022-11-28"
	MCPVersion      = "3.0"
	ProtocolVersion = "3.0"
)

// MCP bridge modes
const (
	// WHO: BridgeManager
	// WHAT: Bridge operation modes
	// WHEN: During bridge configuration
	// WHERE: System Layer 6 (Integration)
	// WHY: To define bridge behaviors
	// HOW: Using string constants
	// EXTENT: All bridge operations

	BridgeModeDirect  = "direct"
	BridgeModeProxied = "proxied"
	BridgeModeAsync   = "async"
)

// Resource URI templates
const (
	// WHO: ResourceManager
	// WHAT: URI templates for GitHub resources
	// WHEN: During resource access
	// WHERE: System Layer 6 (Integration)
	// WHY: To standardize resource addressing
	// HOW: Using string templates
	// EXTENT: All resource references

	RepoContentsURITemplate     = "repo://%s/%s/contents%s"
	RepoBranchURITemplate       = "repo://%s/%s/refs/heads/%s/contents%s"
	RepoCommitURITemplate       = "repo://%s/%s/sha/%s/contents%s"
	RepoIssueURITemplate        = "repo://%s/%s/issues/%d"
	RepoPullRequestURITemplate  = "repo://%s/%s/pulls/%d"
	RepoCodeScanningURITemplate = "repo://%s/%s/code-scanning/alerts"
)

// Error codes
const (
	// WHO: ErrorManager
	// WHAT: Error classification codes
	// WHEN: During error handling
	// WHERE: System Layer 6 (Integration)
	// WHY: To classify errors
	// HOW: Using string constants
	// EXTENT: All error scenarios

	ErrorAuthFailed       = "auth_failed"
	ErrorNotFound         = "not_found"
	ErrorPermissionDenied = "permission_denied"
	ErrorRateLimit        = "rate_limit"
	ErrorServerError      = "server_error"
	ErrorBadRequest       = "bad_request"
	ErrorBridgeFailed     = "bridge_failed"
	ErrorContextInvalid   = "context_invalid"
)

// Parameter descriptions
const (
	// WHO: ParameterDescriptionManager
	// WHAT: Parameter description constants
	// WHEN: During tool creation
	// WHERE: GitHub MCP Server
	// WHY: To standardize parameter descriptions
	// HOW: Using string constants
	// EXTENT: All tool parameter descriptions

	ConstRepoOwnerDesc = "Repository owner"
	ConstRepoNameDesc  = "Repository name"
)

// Error message templates
const (
	// WHO: ErrorMessageManager
	// WHAT: Error message templates
	// WHEN: During error handling
	// WHERE: GitHub MCP Server
	// WHY: To standardize error messages
	// HOW: Using string templates
	// EXTENT: All error reporting

	ErrMsgGetGitHubClient  = "failed to get GitHub client: %w"
	ErrMsgReadResponseBody = "failed to read response body: %w"
	ErrMsgMarshalIssue     = "failed to marshal issue: %w"
	ErrMsgMarshalComment   = "failed to marshal comment: %w"
	ErrMsgMarshalPR        = "failed to marshal pull request: %w"
	ErrMsgMarshalSearchRes = "failed to marshal search results: %w"
	ErrMsgMarshalUser      = "failed to marshal user: %w"
	ErrMsgMarshalComments  = "failed to marshal comments: %w"
	ErrMsgMarshalIssues    = "failed to marshal issues: %w"
)

// Event type constants
const (
	// WHO: EventTypeManager
	// WHAT: Event Type Constants
	// WHEN: During event processing
	// WHERE: GitHub MCP Server
	// WHY: To define event categories
	// HOW: By declaring string constants
	// EXTENT: All supported event types

	EventTypeRepository = "repository"
	EventTypeIssue      = "issue"
	EventTypePR         = "pull_request"
	EventTypeRelease    = "release"
	EventTypeWorkflow   = "workflow"
)

// Pull request action constants
const (
	// WHO: PRActionTypeManager
	// WHAT: PR Action Constants
	// WHEN: During PR event processing
	// WHERE: GitHub MCP Server
	// WHY: To define PR action types
	// HOW: By declaring string constants
	// EXTENT: All supported PR actions

	PRActionOpened      = "opened"
	PRActionClosed      = "closed"
	PRActionMerged      = "merged"
	PRActionReopened    = "reopened"
	PRActionSynchronize = "synchronize"
)

// Search type constants
const (
	// WHO: SearchTypeManager
	// WHAT: Search Type Constants
	// WHEN: During search operations
	// WHERE: GitHub MCP Server
	// WHY: To define search categories
	// HOW: By declaring string constants
	// EXTENT: All supported search types

	SearchTypeCode         = "code"
	SearchTypeCommits      = "commits"
	SearchTypeIssues       = "issues"
	SearchTypePRs          = "pull_requests"
	SearchTypeRepositories = "repositories"
	SearchTypeUsers        = "users"
)

// Common Function Types
// WHO: TypeManager
// WHAT: Function Type Definitions
// WHEN: During package initialization
// WHERE: GitHub MCP Server
// WHY: To standardize function signatures
// HOW: Using type definitions
// EXTENT: All GitHub MCP operations

// GetClientFn is a function type for getting GitHub clients
type GetClientFn func(ctx context.Context) (*github.Client, error)

// Ptr returns a pointer to any value - utility for creating pointers to values
func Ptr[T any](v T) *T { return &v }

// OptionalParamOK retrieves an optional parameter with type checking
func OptionalParamOK[T any](r mcp.CallToolRequest, p string) (value T, ok bool, err error) {
    val, exists := r.Params.Arguments[p]
    if !exists { return }
    value, ok = val.(T)
    if !ok { err = fmt.Errorf("parameter %s is not of type %T, is %T", p, value, val); ok = true }
    return
}

// OptionalParam retrieves an optional parameter
func OptionalParam[T any](r mcp.CallToolRequest, p string) (value T, _ error) {
    v, _, err := OptionalParamOK[T](r, p)
    return v, err
}

// RequiredParam retrieves a required parameter with error handling
func RequiredParam[T comparable](r mcp.CallToolRequest, p string) (T, error) { 
    var zero T
    v, ok := r.Params.Arguments[p]
    if !ok {
        return zero, fmt.Errorf("missing required parameter: %s", p)
    }
    value, ok2 := v.(T)
    if !ok2 || value == zero {
        return zero, fmt.Errorf("invalid parameter: %s", p)
    }
    return value, nil
}

// RequiredInt retrieves a required integer parameter directly
func RequiredInt(r mcp.CallToolRequest, p string) (int, error) {
    v, err := RequiredParam[float64](r, p)
    if err != nil {
        return 0, err
    }
    return int(v), nil
}

// RequiredIntParam retrieves a required integer parameter
func RequiredIntParam(r mcp.CallToolRequest, p string) (int, error) { 
    v, err := RequiredParam[float64](r, p) 
    if err != nil { return 0, err }
    return int(v), nil 
}

// OptionalInt retrieves an optional integer parameter
func OptionalInt(r mcp.CallToolRequest, p string) (int, bool, error) { 
    v, ok, err := OptionalParamOK[float64](r, p) 
    if err != nil || !ok { return 0, ok, err }
    return int(v), ok, nil 
}

// OptionalIntParam gets an optional integer parameter
func OptionalIntParam(r mcp.CallToolRequest, p string) (int, error) { 
    v, _, err := OptionalParamOK[float64](r, p)
    return int(v), err 
}

// OptionalIntParamWithDefault gets an optional integer parameter with default
func OptionalIntParamWithDefault(r mcp.CallToolRequest, p string, d int) (int, error) { 
    v, err := OptionalIntParam(r, p)
    if err != nil { return 0, err }
    if v == 0 { return d, nil }
    return v, nil 
}

// WithPagination adds pagination parameters to a tool
func WithPagination(t mcp.Tool) mcp.Tool {
    return t.WithNumber("page", 
            mcp.Description("Page number for pagination")).
        WithNumber("per_page", 
            mcp.Description("Number of results per page"))
}

// OptionalPaginationParams gets pagination parameters
func OptionalPaginationParams(r mcp.CallToolRequest) (*github.ListOptions, error) {
    opts := &github.ListOptions{}
    
    page, hasPage, err := OptionalInt(r, "page")
    if err != nil {
        return nil, fmt.Errorf("invalid page parameter: %w", err)
    }
    
    perPage, hasPerPage, err := OptionalInt(r, "per_page")
    if err != nil {
        return nil, fmt.Errorf("invalid per_page parameter: %w", err)
    }
    
    if hasPage {
        opts.Page = page
    } else {
        opts.Page = 1 // Default
    }
    
    if hasPerPage {
        opts.PerPage = perPage
    } else {
        opts.PerPage = 30 // Default
    }
    
    return opts, nil
}

// isAcceptedError checks if error is an AcceptedError
func isAcceptedError(err error) bool { 
    var acceptedError *github.AcceptedError
    return errors.As(err, &acceptedError) 
}
