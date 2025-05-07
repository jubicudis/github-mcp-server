/*
 * WHO: ModelsPackage
 * WHAT: Core data models for GitHub MCP server
 * WHEN: Throughout system operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide structured data representations
 * HOW: Using Go structs with 7D context
 * EXTENT: All data modeling needs
 */

package models

import (
	"time"
)

// ContextVector7D represents the 7D context framework
type ContextVector7D struct {
	// WHO: Actor & Identity Context
	Who string `json:"who"`

	// WHAT: Function & Content Context
	What string `json:"what"`

	// WHEN: Temporal Context
	When int64 `json:"when"`

	// WHERE: Location Context
	Where string `json:"where"`

	// WHY: Intent & Purpose Context
	Why string `json:"why"`

	// HOW: Method & Process Context
	How string `json:"how"`

	// TO WHAT EXTENT: Scope & Impact Context
	Extent float64 `json:"extent"`

	// Additional metadata
	Meta   map[string]interface{} `json:"meta,omitempty"`
	Source string                 `json:"source,omitempty"`
}

// Repository represents a GitHub repository
type Repository struct {
	// WHO: ContentManager
	// WHAT: Repository data structure
	// WHEN: During repository operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To model repository data
	// HOW: Using Go struct
	// EXTENT: All repository attributes

	Owner         string                 `json:"owner"`
	Name          string                 `json:"name"`
	FullName      string                 `json:"full_name"`
	Description   string                 `json:"description"`
	URL           string                 `json:"url"`
	DefaultBranch string                 `json:"default_branch"`
	Private       bool                   `json:"private"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Context       *ContextVector7D       `json:"context,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// Issue represents a GitHub issue
type Issue struct {
	// WHO: IssueManager
	// WHAT: Issue data structure
	// WHEN: During issue operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To model issue data
	// HOW: Using Go struct
	// EXTENT: All issue attributes

	Number     int                    `json:"number"`
	Title      string                 `json:"title"`
	Body       string                 `json:"body"`
	State      string                 `json:"state"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	ClosedAt   *time.Time             `json:"closed_at,omitempty"`
	User       string                 `json:"user"`
	Labels     []string               `json:"labels,omitempty"`
	Assignees  []string               `json:"assignees,omitempty"`
	URL        string                 `json:"url"`
	HTMLURL    string                 `json:"html_url"`
	Context    *ContextVector7D       `json:"context,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Repository string                 `json:"repository"`
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	// WHO: PRManager
	// WHAT: PR data structure
	// WHEN: During PR operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To model PR data
	// HOW: Using Go struct
	// EXTENT: All PR attributes

	Number     int                    `json:"number"`
	Title      string                 `json:"title"`
	Body       string                 `json:"body"`
	State      string                 `json:"state"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	ClosedAt   *time.Time             `json:"closed_at,omitempty"`
	MergedAt   *time.Time             `json:"merged_at,omitempty"`
	User       string                 `json:"user"`
	Base       string                 `json:"base"`
	Head       string                 `json:"head"`
	Draft      bool                   `json:"draft"`
	Merged     bool                   `json:"merged"`
	Mergeable  *bool                  `json:"mergeable,omitempty"`
	URL        string                 `json:"url"`
	HTMLURL    string                 `json:"html_url"`
	Context    *ContextVector7D       `json:"context,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Repository string                 `json:"repository"`
}

// FileContent represents a file in a GitHub repository
type FileContent struct {
	// WHO: ContentManager
	// WHAT: File content data structure
	// WHEN: During content operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To model file data
	// HOW: Using Go struct
	// EXTENT: All file attributes

	Path        string                 `json:"path"`
	Name        string                 `json:"name"`
	SHA         string                 `json:"sha"`
	Size        int                    `json:"size"`
	Type        string                 `json:"type"`
	Content     string                 `json:"content,omitempty"`
	Encoding    string                 `json:"encoding,omitempty"`
	URL         string                 `json:"url"`
	HTMLURL     string                 `json:"html_url"`
	DownloadURL string                 `json:"download_url,omitempty"`
	Context     *ContextVector7D       `json:"context,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Branch represents a GitHub branch
type Branch struct {
	// WHO: BranchManager
	// WHAT: Branch data structure
	// WHEN: During branch operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To model branch data
	// HOW: Using Go struct
	// EXTENT: All branch attributes

	Name       string                 `json:"name"`
	SHA        string                 `json:"sha"`
	Protected  bool                   `json:"protected"`
	URL        string                 `json:"url"`
	Context    *ContextVector7D       `json:"context,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Repository string                 `json:"repository"`
}

// Commit represents a GitHub commit
type Commit struct {
	// WHO: CommitManager
	// WHAT: Commit data structure
	// WHEN: During commit operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To model commit data
	// HOW: Using Go struct
	// EXTENT: All commit attributes

	SHA         string                 `json:"sha"`
	Message     string                 `json:"message"`
	Author      string                 `json:"author"`
	AuthorEmail string                 `json:"author_email"`
	AuthorDate  time.Time              `json:"author_date"`
	Committer   string                 `json:"committer"`
	CommitDate  time.Time              `json:"commit_date"`
	URL         string                 `json:"url"`
	HTMLURL     string                 `json:"html_url"`
	Context     *ContextVector7D       `json:"context,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Repository  string                 `json:"repository"`
}

// MCPRequest represents an MCP (Model Context Protocol) request
type MCPRequest struct {
	// WHO: MCPRequestManager
	// WHAT: MCP request data structure
	// WHEN: During MCP operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To model MCP requests
	// HOW: Using Go struct
	// EXTENT: All MCP request attributes

	ID        string                 `json:"id"`
	Method    string                 `json:"method"`
	URI       string                 `json:"uri"`
	Params    map[string]interface{} `json:"params,omitempty"`
	Context   *ContextVector7D       `json:"context,omitempty"`
	Timestamp int64                  `json:"timestamp"`
	Source    string                 `json:"source,omitempty"`
}

// MCPResponse represents an MCP (Model Context Protocol) response
type MCPResponse struct {
	// WHO: MCPResponseManager
	// WHAT: MCP response data structure
	// WHEN: During MCP operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To model MCP responses
	// HOW: Using Go struct
	// EXTENT: All MCP response attributes

	ID        string                 `json:"id"`
	Status    int                    `json:"status"`
	Data      interface{}            `json:"data,omitempty"`
	Error     *MCPError              `json:"error,omitempty"`
	Context   *ContextVector7D       `json:"context,omitempty"`
	Timestamp int64                  `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MCPError represents an MCP (Model Context Protocol) error
type MCPError struct {
	// WHO: ErrorManager
	// WHAT: Error data structure
	// WHEN: During error handling
	// WHERE: System Layer 6 (Integration)
	// WHY: To model error data
	// HOW: Using Go struct
	// EXTENT: All error attributes

	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Details interface{}            `json:"details,omitempty"`
	Context *ContextVector7D       `json:"context,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// ToMap converts a ContextVector7D to a map
func (cv *ContextVector7D) ToMap() map[string]interface{} {
	// WHO: ContextConverter
	// WHAT: Convert context to map
	// WHEN: During serialization
	// WHERE: System Layer 6 (Integration)
	// WHY: For JSON conversion
	// HOW: Using struct to map conversion
	// EXTENT: All context dimensions

	return map[string]interface{}{
		"who":    cv.Who,
		"what":   cv.What,
		"when":   cv.When,
		"where":  cv.Where,
		"why":    cv.Why,
		"how":    cv.How,
		"extent": cv.Extent,
		"meta":   cv.Meta,
		"source": cv.Source,
	}
}

// NewContextVector7D creates a new 7D context vector
func NewContextVector7D(who, what, where, why, how string, extent float64) *ContextVector7D {
	// WHO: ContextCreator
	// WHAT: Create context vector
	// WHEN: During context initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To initialize context
	// HOW: Using struct constructor
	// EXTENT: All context dimensions

	return &ContextVector7D{
		Who:    who,
		What:   what,
		When:   time.Now().Unix(),
		Where:  where,
		Why:    why,
		How:    how,
		Extent: extent,
		Meta:   make(map[string]interface{}),
		Source: "github_mcp",
	}
}
