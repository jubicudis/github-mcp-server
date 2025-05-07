/*
 * WHO: GitHubModelsExtended
 * WHAT: Extended GitHub API data model definitions
 * WHEN: During GitHub API operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide additional structured data types for GitHub interactions
 * HOW: Using Go structs with JSON mapping
 * EXTENT: Extended GitHub API data structures
 */

package github

import (
	"fmt"
)

// Repository represents a GitHub repository
type Repository struct {
	// WHO: RepoManager
	// WHAT: Base repository data structure
	// WHEN: During repo operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To store core repository information
	// HOW: Using structured data
	// EXTENT: Single repository data

	ID            int64  `json:"id"`
	NodeID        string `json:"node_id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Description   string `json:"description"`
	Private       bool   `json:"private"`
	Owner         User   `json:"owner"`
	HTMLURL       string `json:"html_url"`
	URL           string `json:"url"`
	DefaultBranch string `json:"default_branch"`
}

// User represents a GitHub user
type User struct {
	// WHO: UserManager
	// WHAT: Base user data structure
	// WHEN: During user operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To store core user information
	// HOW: Using structured data
	// EXTENT: Single user data

	Login     string `json:"login"`
	ID        int64  `json:"id"`
	NodeID    string `json:"node_id"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	URL       string `json:"url"`
	Type      string `json:"type"`
}

// Issue represents a GitHub issue
type Issue struct {
	// WHO: IssueManager
	// WHAT: Base issue data structure
	// WHEN: During issue operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To store core issue information
	// HOW: Using structured data
	// EXTENT: Single issue data

	ID        int64  `json:"id"`
	NodeID    string `json:"node_id"`
	URL       string `json:"url"`
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	Locked    bool   `json:"locked"`
	User      User   `json:"user"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	HTMLURL   string `json:"html_url"`
}

// DetailedAPIError represents a detailed error returned from the GitHub API
// This is separate from the APIError in models.go
type DetailedAPIError struct {
	// WHO: ExtendedErrorManager
	// WHAT: Detailed API error structure
	// WHEN: During error handling
	// WHERE: System Layer 6 (Integration)
	// WHY: To standardize detailed error responses
	// HOW: Using structured error data
	// EXTENT: All API errors

	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
	RequestID  string `json:"request_id"`
	URL        string `json:"url"`
}

// Error implements the error interface for DetailedAPIError
func (e *DetailedAPIError) Error() string {
	// WHO: ErrorFormatter
	// WHAT: Format error message
	// WHEN: During error display
	// WHERE: System Layer 6 (Integration)
	// WHY: To standardize error output
	// HOW: Using string formatting
	// EXTENT: Single error message

	return fmt.Sprintf("GitHub API error: %s (status: %d, request: %s)", e.Message, e.StatusCode, e.RequestID)
}

// User represents a GitHub user
// Note: This extends the base User model from models.go with additional fields
type ExtendedUser struct {
	// WHO: UserManager
	// WHAT: Extended user data structure
	// WHEN: During user operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To store enhanced user information
	// HOW: Using structured data
	// EXTENT: Single user data

	User           // Embed base User struct from models.go
	Following int  `json:"following"`
	Followers int  `json:"followers"`
	SiteAdmin bool `json:"site_admin"`
}

// ExtendedRepository represents a GitHub repository with extended information
type ExtendedRepository struct {
	// WHO: RepoManager
	// WHAT: Extended repository data structure
	// WHEN: During repo operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To store enhanced repo information
	// HOW: Using structured data
	// EXTENT: Single repository data

	Repository               // Embed base Repository struct from models.go
	ForksCount      int      `json:"forks_count"`
	StargazersCount int      `json:"stargazers_count"`
	WatchersCount   int      `json:"watchers_count"`
	OpenIssuesCount int      `json:"open_issues_count"`
	Topics          []string `json:"topics"`
	Language        string   `json:"language"`
	HasIssues       bool     `json:"has_issues"`
	HasProjects     bool     `json:"has_projects"`
	HasWiki         bool     `json:"has_wiki"`
}

// ExtendedIssue represents a GitHub issue with extended information
type ExtendedIssue struct {
	// WHO: IssueManager
	// WHAT: Extended issue data structure
	// WHEN: During issue operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To store enhanced issue information
	// HOW: Using structured data
	// EXTENT: Single issue data

	Issue            // Embed base Issue struct from models.go
	Comments  int    `json:"comments"`
	Assignees []User `json:"assignees"`
	Labels    []struct {
		Name        string `json:"name"`
		Color       string `json:"color"`
		Description string `json:"description"`
	} `json:"labels"`
	Milestone struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		DueOn       string `json:"due_on"`
	} `json:"milestone"`
	ClosedAt string `json:"closed_at"`
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	// WHO: PRManager
	// WHAT: Base pull request data structure
	// WHEN: During PR operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To store core pull request information
	// HOW: Using structured data
	// EXTENT: Single PR data

	ID        int64  `json:"id"`
	NodeID    string `json:"node_id"`
	Number    int    `json:"number"`
	State     string `json:"state"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	User      User   `json:"user"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	HTMLURL   string `json:"html_url"`
	URL       string `json:"url"`
	Head      struct {
		Ref  string     `json:"ref"`
		SHA  string     `json:"sha"`
		Repo Repository `json:"repo"`
	} `json:"head"`
	Base struct {
		Ref  string     `json:"ref"`
		SHA  string     `json:"sha"`
		Repo Repository `json:"repo"`
	} `json:"base"`
}

// ExtendedPullRequest represents a GitHub PR with extended information
type ExtendedPullRequest struct {
	// WHO: PRManager
	// WHAT: Extended pull request data structure
	// WHEN: During PR operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To store enhanced PR information
	// HOW: Using structured data
	// EXTENT: Single PR data

	PullRequest                // Embed base PullRequest struct
	Merged              bool   `json:"merged"`
	MergedAt            string `json:"merged_at"`
	MergedBy            User   `json:"merged_by"`
	Comments            int    `json:"comments"`
	ReviewComments      int    `json:"review_comments"`
	Commits             int    `json:"commits"`
	Additions           int    `json:"additions"`
	Deletions           int    `json:"deletions"`
	ChangedFiles        int    `json:"changed_files"`
	MaintainerCanModify bool   `json:"maintainer_can_modify"`
}

// WorkflowJob represents a job within a GitHub Actions workflow
type WorkflowJob struct {
	// WHO: JobManager
	// WHAT: Workflow job structure
	// WHEN: During workflow operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To track job execution
	// HOW: Using structured data
	// EXTENT: Single workflow job

	ID          int64  `json:"id"`
	RunID       int64  `json:"run_id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Conclusion  string `json:"conclusion"`
	StartedAt   string `json:"started_at"`
	CompletedAt string `json:"completed_at"`
	Steps       []struct {
		Name       string `json:"name"`
		Status     string `json:"status"`
		Conclusion string `json:"conclusion"`
		Number     int    `json:"number"`
	} `json:"steps"`
}

// TeamMember represents a member of a GitHub team
type TeamMember struct {
	// WHO: TeamManager
	// WHAT: Team member structure
	// WHEN: During team operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To track team membership
	// HOW: Using structured data
	// EXTENT: Single team member

	Login     string `json:"login"`
	ID        int64  `json:"id"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
	Role      string `json:"role"`
}

// OrganizationMember represents a member of a GitHub organization
type OrganizationMember struct {
	// WHO: OrgManager
	// WHAT: Organization member structure
	// WHEN: During org operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To track org membership
	// HOW: Using structured data
	// EXTENT: Single org member

	Login     string `json:"login"`
	ID        int64  `json:"id"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
	Role      string `json:"role"`
}

// ExtendedCodeScanningAlert represents a code scanning alert with detailed information
type ExtendedCodeScanningAlert struct {
	// WHO: SecurityManager
	// WHAT: Extended security alert structure
	// WHEN: During security operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide detailed vulnerability data
	// HOW: Using structured data
	// EXTENT: Single security issue

	CodeScanAlert         // Embed base CodeScanAlert struct from models.go
	InstancesCount int    `json:"instances_count"`
	FixedAt        string `json:"fixed_at,omitempty"`
}

// SearchQueryBuilder helps construct GitHub search queries
type SearchQueryBuilder struct {
	// WHO: SearchQueryManager
	// WHAT: Search query construction helper
	// WHEN: During search preparation
	// WHERE: System Layer 6 (Integration)
	// WHY: To simplify query construction
	// HOW: Using builder pattern
	// EXTENT: Single search query

	queryParts []string
	type_      string
}

// NewSearchQueryBuilder creates a new search query builder for the specified type
func NewSearchQueryBuilder(type_ string) *SearchQueryBuilder {
	// WHO: SearchQueryFactory
	// WHAT: Create search query builder
	// WHEN: During search preparation
	// WHERE: System Layer 6 (Integration)
	// WHY: To initialize builder
	// HOW: Using factory pattern
	// EXTENT: Builder creation

	return &SearchQueryBuilder{
		queryParts: []string{},
		type_:      type_,
	}
}

// AddTerm adds a search term to the builder
func (b *SearchQueryBuilder) AddTerm(term string) *SearchQueryBuilder {
	// WHO: SearchTermAdder
	// WHAT: Add search term
	// WHEN: During query construction
	// WHERE: System Layer 6 (Integration)
	// WHY: To build search query
	// HOW: Using builder pattern
	// EXTENT: Query modification

	if term != "" {
		b.queryParts = append(b.queryParts, term)
	}
	return b
}

// AddRepoFilter adds a repository filter to the search
func (b *SearchQueryBuilder) AddRepoFilter(owner, repo string) *SearchQueryBuilder {
	// WHO: RepoFilterAdder
	// WHAT: Add repo filter
	// WHEN: During query construction
	// WHERE: System Layer 6 (Integration)
	// WHY: To scope search to repo
	// HOW: Using builder pattern
	// EXTENT: Query modification

	if owner != "" && repo != "" {
		b.queryParts = append(b.queryParts, fmt.Sprintf("repo:%s/%s", owner, repo))
	}
	return b
}

// AddLanguage adds a language filter to the search
func (b *SearchQueryBuilder) AddLanguage(language string) *SearchQueryBuilder {
	// WHO: LanguageFilterAdder
	// WHAT: Add language filter
	// WHEN: During query construction
	// WHERE: System Layer 6 (Integration)
	// WHY: To filter by language
	// HOW: Using builder pattern
	// EXTENT: Query modification

	if language != "" {
		b.queryParts = append(b.queryParts, fmt.Sprintf("language:%s", language))
	}
	return b
}

// Build builds the final search query string
func (b *SearchQueryBuilder) Build() string {
	// WHO: QueryBuilder
	// WHAT: Build search query
	// WHEN: During query finalization
	// WHERE: System Layer 6 (Integration)
	// WHY: To create final query
	// HOW: Using string joining
	// EXTENT: Query construction

	query := ""
	for i, part := range b.queryParts {
		if i > 0 {
			query += " "
		}
		query += part
	}
	if b.type_ != "" {
		query = fmt.Sprintf("type:%s %s", b.type_, query)
	}
	return query
}
