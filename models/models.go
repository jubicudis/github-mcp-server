/*
 * WHO: ModelsPackage
 * WHAT: Core data models for GitHub MCP server
 * WHEN: Throughout system operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To define standard data structures
 * HOW: Using Go structures with 7D context
 * EXTENT: All data modeling needs
 */

package models

import (
	"time"
)

// ContextVector7D represents the 7-dimensional context framework
type ContextVector7D struct {
	// WHO: Identity and actor information
	Who string `json:"who"`

	// WHAT: Function and content information
	What string `json:"what"`

	// WHEN: Temporal context (Unix timestamp)
	When int64 `json:"when"`

	// WHERE: Location context within the system
	Where string `json:"where"`

	// WHY: Intent and purpose context
	Why string `json:"why"`

	// HOW: Method and process context
	How string `json:"how"`

	// TO WHAT EXTENT: Scope and impact context (0.0-1.0)
	Extent float64 `json:"extent"`

	// Additional metadata for the context
	Meta map[string]interface{} `json:"meta,omitempty"`

	// Source of the context
	Source string `json:"source,omitempty"`
}

// NewContextVector7D creates a new context vector with default values
func NewContextVector7D(who, what, where, why, how string) *ContextVector7D {
	// WHO: ContextFactory
	// WHAT: Create new context vector
	// WHEN: During context initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To standardize context creation
	// HOW: Using factory pattern
	// EXTENT: Single context creation

	return &ContextVector7D{
		Who:    who,
		What:   what,
		When:   time.Now().Unix(),
		Where:  where,
		Why:    why,
		How:    how,
		Extent: 1.0,
		Meta:   make(map[string]interface{}),
		Source: "system",
	}
}

// MCPRequest represents a generic MCP request structure
type MCPRequest struct {
	// Request ID
	ID string `json:"id"`

	// Request type
	Type string `json:"type"`

	// Request URI
	URI string `json:"uri"`

	// Request parameters
	Params map[string]interface{} `json:"params,omitempty"`

	// Request context
	Context *ContextVector7D `json:"context"`

	// Request timestamp
	Timestamp int64 `json:"timestamp"`
}

// MCPResponse represents a generic MCP response structure
type MCPResponse struct {
	// Request ID this response is for
	RequestID string `json:"requestId"`

	// Status code
	Status int `json:"status"`

	// Status message
	Message string `json:"message"`

	// Response data
	Data interface{} `json:"data,omitempty"`

	// Response context
	Context *ContextVector7D `json:"context"`

	// Response timestamp
	Timestamp int64 `json:"timestamp"`
}

// GitHubRepository represents a GitHub repository
type GitHubRepository struct {
	Owner       string           `json:"owner"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	URL         string           `json:"url"`
	Private     bool             `json:"private"`
	Fork        bool             `json:"fork"`
	Stars       int              `json:"stars"`
	Watchers    int              `json:"watchers"`
	Issues      int              `json:"issues"`
	Context     *ContextVector7D `json:"context,omitempty"`
}

// GitHubUser represents a GitHub user
type GitHubUser struct {
	Login     string           `json:"login"`
	ID        int              `json:"id"`
	Name      string           `json:"name"`
	Email     string           `json:"email"`
	AvatarURL string           `json:"avatarUrl"`
	URL       string           `json:"url"`
	Type      string           `json:"type"`
	Context   *ContextVector7D `json:"context,omitempty"`
}

// GitHubIssue represents a GitHub issue
type GitHubIssue struct {
	Number    int              `json:"number"`
	Title     string           `json:"title"`
	Body      string           `json:"body"`
	State     string           `json:"state"`
	CreatedAt int64            `json:"createdAt"`
	UpdatedAt int64            `json:"updatedAt"`
	ClosedAt  int64            `json:"closedAt,omitempty"`
	User      *GitHubUser      `json:"user"`
	Labels    []string         `json:"labels"`
	Context   *ContextVector7D `json:"context,omitempty"`
}

// GitHubPullRequest represents a GitHub pull request
type GitHubPullRequest struct {
	Number     int              `json:"number"`
	Title      string           `json:"title"`
	Body       string           `json:"body"`
	State      string           `json:"state"`
	CreatedAt  int64            `json:"createdAt"`
	UpdatedAt  int64            `json:"updatedAt"`
	ClosedAt   int64            `json:"closedAt,omitempty"`
	MergedAt   int64            `json:"mergedAt,omitempty"`
	BaseBranch string           `json:"baseBranch"`
	HeadBranch string           `json:"headBranch"`
	User       *GitHubUser      `json:"user"`
	Mergeable  bool             `json:"mergeable"`
	Context    *ContextVector7D `json:"context,omitempty"`
}

// GitHubContent represents content from a GitHub repository
type GitHubContent struct {
	Name        string           `json:"name"`
	Path        string           `json:"path"`
	SHA         string           `json:"sha"`
	Size        int              `json:"size"`
	Type        string           `json:"type"` // file, dir, symlink
	Content     string           `json:"content,omitempty"`
	Encoding    string           `json:"encoding,omitempty"`
	DownloadURL string           `json:"downloadUrl,omitempty"`
	Context     *ContextVector7D `json:"context,omitempty"`
}

// GitHubSearchResult represents a search result from GitHub
type GitHubSearchResult struct {
	TotalCount int              `json:"totalCount"`
	Items      []interface{}    `json:"items"`
	Context    *ContextVector7D `json:"context,omitempty"`
}

// GitHubCodeScanningAlert represents a code scanning alert from GitHub
type GitHubCodeScanningAlert struct {
	Number      int    `json:"number"`
	RuleID      string `json:"ruleId"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	State       string `json:"state"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
	FixedAt     int64  `json:"fixedAt,omitempty"`
	Location    struct {
		Path      string `json:"path"`
		StartLine int    `json:"startLine"`
		EndLine   int    `json:"endLine"`
	} `json:"location"`
	Context *ContextVector7D `json:"context,omitempty"`
}

// BridgeMessage represents a message exchanged between GitHub MCP and TNOS MCP
type BridgeMessage struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"`
	Source          string                 `json:"source"`
	Destination     string                 `json:"destination"`
	Payload         interface{}            `json:"payload"`
	Context         *ContextVector7D       `json:"context"`
	Version         string                 `json:"version"`
	Timestamp       int64                  `json:"timestamp"`
	Compressed      bool                   `json:"compressed"`
	CompressionData map[string]interface{} `json:"compressionData,omitempty"`
}

// TranslationResult represents the result of a context translation
type TranslationResult struct {
	Original   *ContextVector7D `json:"original"`
	Translated *ContextVector7D `json:"translated"`
	Success    bool             `json:"success"`
	Message    string           `json:"message,omitempty"`
	Timestamp  int64            `json:"timestamp"`
}

// SystemHealth represents the health status of the MCP system
type SystemHealth struct {
	Status      string                     `json:"status"` // "healthy", "degraded", "failing"
	Components  map[string]ComponentHealth `json:"components"`
	LastChecked int64                      `json:"lastChecked"`
	Context     *ContextVector7D           `json:"context"`
}

// ComponentHealth represents the health status of a specific component
type ComponentHealth struct {
	Status      string `json:"status"` // "healthy", "degraded", "failing"
	Message     string `json:"message,omitempty"`
	LastChecked int64  `json:"lastChecked"`
}
