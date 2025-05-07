/*
 * WHO: ModelManager
 * WHAT: Shared model definitions for GitHub API
 * WHEN: During API operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide consistent data models
 * HOW: Using Go struct definitions
 * EXTENT: All GitHub API data models
 */

package github

// RepoContent represents content retrieved from a repository
type RepoContent struct {
	// WHO: ContentStructureManager
	// WHAT: Repository content structure
	// WHEN: During content operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To model content data
	// HOW: Using GitHub API schema
	// EXTENT: Content representation

	Name        string `json:"name"`
	Path        string `json:"path"`
	SHA         string `json:"sha"`
	Size        int    `json:"size"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	GitURL      string `json:"git_url"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
}

// Branch represents a Git branch
type Branch struct {
	// WHO: BranchStructureManager
	// WHAT: Branch data structure
	// WHEN: During branch operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To model branch data
	// HOW: Using GitHub API schema
	// EXTENT: Branch representation

	Name      string `json:"name"`
	Commit    Commit `json:"commit"`
	Protected bool   `json:"protected"`
}

// Commit represents a Git commit
type Commit struct {
	// WHO: CommitStructureManager
	// WHAT: Commit data structure
	// WHEN: During commit operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To model commit data
	// HOW: Using GitHub API schema
	// EXTENT: Commit representation

	SHA       string `json:"sha"`
	URL       string `json:"url"`
	Author    User   `json:"author,omitempty"`
	Committer User   `json:"committer,omitempty"`
	Message   string `json:"message,omitempty"`
}

// CodeScanAlert represents a security alert from code scanning
type CodeScanAlert struct {
	// WHO: SecurityStructureManager
	// WHAT: Code scan alert data structure
	// WHEN: During security operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To model security data
	// HOW: Using GitHub API schema
	// EXTENT: Security alert representation

	Number      int    `json:"number"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	State       string `json:"state"`
	DismissedAt string `json:"dismissed_at,omitempty"`
	DismissedBy User   `json:"dismissed_by,omitempty"`
	Rule        struct {
		ID          string `json:"id"`
		Severity    string `json:"severity"`
		Security    string `json:"security_severity_level,omitempty"`
		Description string `json:"description"`
	} `json:"rule"`
	Tool struct {
		Name    string `json:"name"`
		Version string `json:"version,omitempty"`
	} `json:"tool"`
	MostRecentInstance struct {
		Ref         string `json:"ref"`
		AnalysisKey string `json:"analysis_key"`
		Location    struct {
			Path        string `json:"path"`
			StartLine   int    `json:"start_line"`
			EndLine     int    `json:"end_line"`
			StartColumn int    `json:"start_column"`
			EndColumn   int    `json:"end_column"`
		} `json:"location"`
		Message struct {
			Text string `json:"text"`
		} `json:"message"`
	} `json:"most_recent_instance"`
}

// WorkflowRun represents a GitHub Actions workflow run
type WorkflowRun struct {
	// WHO: WorkflowStructureManager
	// WHAT: Workflow run data structure
	// WHEN: During workflow operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To model workflow data
	// HOW: Using GitHub API schema
	// EXTENT: Workflow representation

	ID         int64  `json:"id"`
	Name       string `json:"name"`
	HeadBranch string `json:"head_branch"`
	HeadSHA    string `json:"head_sha"`
	RunNumber  int    `json:"run_number"`
	Event      string `json:"event"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	WorkflowID int64  `json:"workflow_id"`
	URL        string `json:"url"`
	HTMLURL    string `json:"html_url"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// Release represents a GitHub release
type Release struct {
	// WHO: ReleaseStructureManager
	// WHAT: Release data structure
	// WHEN: During release operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To model release data
	// HOW: Using GitHub API schema
	// EXTENT: Release representation

	ID          int64  `json:"id"`
	TagName     string `json:"tag_name"`
	Target      string `json:"target_commitish"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	Draft       bool   `json:"draft"`
	Prerelease  bool   `json:"prerelease"`
	CreatedAt   string `json:"created_at"`
	PublishedAt string `json:"published_at"`
	Author      User   `json:"author"`
	Assets      []struct {
		ID                 int64  `json:"id"`
		Name               string `json:"name"`
		Label              string `json:"label"`
		ContentType        string `json:"content_type"`
		State              string `json:"state"`
		Size               int    `json:"size"`
		DownloadCount      int    `json:"download_count"`
		CreatedAt          string `json:"created_at"`
		UpdatedAt          string `json:"updated_at"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// ResourceURI represents a parsed GitHub resource URI
type ResourceURI struct {
	// WHO: URIStructureManager
	// WHAT: Resource URI structure
	// WHEN: During URI parsing
	// WHERE: System Layer 6 (Integration)
	// WHY: To model resource identifiers
	// HOW: Using structured parsing
	// EXTENT: URI representation

	Scheme    string            // e.g., 'repo'
	Owner     string            // Repository owner
	Repo      string            // Repository name
	Type      string            // Resource type (e.g., 'contents', 'refs', 'sha')
	Reference string            // Branch name, SHA, etc.
	Path      string            // File/directory path
	Query     map[string]string // Query parameters
	Fragment  string            // Fragment identifier
}

// BinaryFile represents a binary file from the repository
type BinaryFile struct {
	// WHO: BinaryStructureManager
	// WHAT: Binary file structure
	// WHEN: During file operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To model binary data
	// HOW: Using binary representation
	// EXTENT: Binary data handling

	Name     string // File name
	Path     string // File path
	Data     []byte // Raw binary data
	Size     int    // File size
	Encoding string // File encoding
	SHA      string // File SHA
}

// DirectoryEntry represents an entry in a directory listing
type DirectoryEntry struct {
	// WHO: DirectoryStructureManager
	// WHAT: Directory entry structure
	// WHEN: During directory operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To model directory data
	// HOW: Using entry representation
	// EXTENT: Directory navigation

	Name        string // Entry name
	Path        string // Entry path
	Type        string // Entry type (file, dir, symlink)
	Size        int    // Entry size (for files)
	SHA         string // Entry SHA
	URL         string // API URL
	HTMLURL     string // Web URL
	DownloadURL string // Download URL (for files)
}

// APIError represents an error from the GitHub API
type APIError struct {
	// WHO: ErrorStructureManager
	// WHAT: API error structure
	// WHEN: During error handling
	// WHERE: System Layer 6 (Integration)
	// WHY: To model error responses
	// HOW: Using error representation
	// EXTENT: Error handling

	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url"`
	Errors           []struct {
		Resource string `json:"resource"`
		Field    string `json:"field"`
		Code     string `json:"code"`
	} `json:"errors,omitempty"`
}

// Error implements the error interface for APIError
func (e *APIError) Error() string {
	// WHO: ErrorFormatter
	// WHAT: Format API error
	// WHEN: During error formatting
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide error message
	// HOW: Using string formatting
	// EXTENT: Error presentation

	return e.Message
}

// RateLimit represents GitHub API rate limit information
type RateLimit struct {
	// WHO: RateLimitStructureManager
	// WHAT: Rate limit structure
	// WHEN: During API operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To model rate limits
	// HOW: Using limit representation
	// EXTENT: Rate management

	Limit     int   `json:"limit"`
	Remaining int   `json:"remaining"`
	Reset     int64 `json:"reset"`
	Used      int   `json:"used"`
}

// SearchResult represents a GitHub search result
type SearchResult struct {
	// WHO: SearchStructureManager
	// WHAT: Search result structure
	// WHEN: During search operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To model search results
	// HOW: Using result representation
	// EXTENT: Search operations

	TotalCount        int                      `json:"total_count"`
	IncompleteResults bool                     `json:"incomplete_results"`
	Items             []interface{}            `json:"items"`
	ItemsTyped        map[string][]interface{} // Typed items by resource type
}
