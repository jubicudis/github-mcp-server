/*
 * WHO: GitHubConstants
 * WHAT: Centralized constants for GitHub MCP server
 * WHEN: During GitHub API operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide consistent API endpoint definitions
 * HOW: Using Go constants with 7D documentation
 * EXTENT: All GitHub API operations
 */

package common

import (
	"context"

	"github.com/google/go-github/v71/github"
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
