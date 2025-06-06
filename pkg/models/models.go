// Package models contains shared data model definitions for GitHub API,
// formerly in pkg/github/models.go.
package models

// RepoContent represents content retrieved from a repository
type RepoContent struct {
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
    Name      string `json:"name"`
    Commit    Commit `json:"commit"`
    Protected bool   `json:"protected"`
}

// Commit represents a Git commit
type Commit struct {
    SHA       string `json:"sha"`
    URL       string `json:"url"`
    Author    User   `json:"author,omitempty"`
    Committer User   `json:"committer,omitempty"`
    Message   string `json:"message,omitempty"`
}

// CodeScanAlert represents a security alert from code scanning
type CodeScanAlert struct {
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
    Name     string // File name
    Path     string // File path
    Data     []byte // Raw binary data
    Size     int    // File size
    Encoding string // File encoding
    SHA      string // File SHA
}

// DirectoryEntry represents an entry in a directory listing
type DirectoryEntry struct {
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
    return e.Message
}

// RateLimit represents GitHub API rate limit information

type RateLimit struct {
    Limit     int   `json:"limit"`
    Remaining int   `json:"remaining"`
    Reset     int64 `json:"reset"`
    Used      int   `json:"used"`
}

// SearchResult represents a GitHub search result

type SearchResult struct {
    TotalCount        int                      `json:"total_count"`
    IncompleteResults bool                     `json:"incomplete_results"`
    Items             []interface{}            `json:"items"`
    ItemsTyped        map[string][]interface{} // Typed items by resource type
}
