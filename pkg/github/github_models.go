// filepath: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/github-mcp-server/pkg/github/github.go

/*
 * WHO: GitHubClient
 * WHAT: GitHub API client for MCP server
 * WHEN: During GitHub API operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide GitHub API access
 * HOW: Using REST API with 7D context awareness
 * EXTENT: All GitHub interactions
 */

package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log"
)

// Constants for API endpoints
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

// Client represents a GitHub API client
type Client struct {
	// WHO: ClientManager
	// WHAT: GitHub API client
	// WHEN: During API operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To manage GitHub connections
	// HOW: Using HTTP client with auth
	// EXTENT: All API interactions

	BaseURL    string
	HTTPClient *http.Client
	Token      string
	Logger     *log.Logger
	Context    *log.ContextVector7D
}

// ContextVector7D represents a 7D context vector
type ContextVector7D struct {
	// WHO: ContextManager
	// WHAT: 7D context for GitHub operations
	// WHEN: During API operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide contextual awareness
	// HOW: Using 7D framework dimensions
	// EXTENT: All contextual data

	Who    string                 `json:"who"`
	What   string                 `json:"what"`
	When   int64                  `json:"when"`
	Where  string                 `json:"where"`
	Why    string                 `json:"why"`
	How    string                 `json:"how"`
	Extent float64                `json:"extent"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
}

// NewClient creates a new GitHub API client
func NewClient(token string, logger *log.Logger) *Client {
	// WHO: ClientCreator
	// WHAT: Initialize GitHub client
	// WHEN: During server startup
	// WHERE: System Layer 6 (Integration)
	// WHY: To create API client
	// HOW: Using provided credentials
	// EXTENT: Client lifecycle

	// Create 7D context for the client
	now := time.Now().Unix()
	context := &log.ContextVector7D{
		Who:    "GitHubClient",
		What:   "APIClient",
		When:   now,
		Where:  "System Layer 6 (Integration)",
		Why:    "GitHub API Access",
		How:    "REST API",
		Extent: 1.0,
		Meta: map[string]interface{}{
			"B":         0.8, // Base factor
			"V":         0.7, // Value factor
			"I":         0.9, // Intent factor
			"G":         1.2, // Growth factor
			"F":         0.6, // Flexibility factor
			"createdAt": now,
		},
		Source: "github_mcp",
	}

	return &Client{
		BaseURL: BaseURL,
		HTTPClient: &http.Client{
			Timeout: time.Second * 30,
		},
		Token:   token,
		Logger:  logger,
		Context: context,
	}
}

// Request makes an authenticated request to the GitHub API
func (c *Client) Request(method, path string, body interface{}, target interface{}) error {
	// WHO: RequestManager
	// WHAT: Make API request
	// WHEN: During API operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To send requests to GitHub
	// HOW: Using HTTP with 7D context
	// EXTENT: Single API request

	url := c.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	// Create the HTTP request
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	// Add context
	ctx := context.WithValue(req.Context(), "7d-context", c.Context)
	req = req.WithContext(ctx)

	// Execute request with timing
	c.Logger.Debug("Making GitHub API request", "method", method, "path", path)
	startTime := time.Now()

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Log response time
	duration := time.Since(startTime).Milliseconds()
	c.Logger.Debug("GitHub API response received",
		"method", method,
		"path", path,
		"status", resp.StatusCode,
		"duration_ms", duration)

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error status codes
	if resp.StatusCode >= 400 {
		var apiError APIError
		if err := json.Unmarshal(respBody, &apiError); err != nil {
			// If can't parse as API error, create a generic one
			return &APIError{
				Message:    fmt.Sprintf("API error with status %d", resp.StatusCode),
				StatusCode: resp.StatusCode,
			}
		}

		// Add status code to the error
		apiError.StatusCode = resp.StatusCode
		return &apiError
	}

	// Parse response if target is provided
	if target != nil {
		if err := json.Unmarshal(respBody, target); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// GetRepository gets a GitHub repository
func (c *Client) GetRepository(owner, repo string) (*Repository, error) {
	// WHO: RepositoryFetcher
	// WHAT: Get repository details
	// WHEN: During repo operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To fetch repo information
	// HOW: Using GitHub API
	// EXTENT: Single repository

	c.Logger.Debug("Fetching repository", "owner", owner, "repo", repo)

	var repository Repository
	path := fmt.Sprintf("%s/%s/%s", ReposEndpoint, owner, repo)

	err := c.Request("GET", path, nil, &repository)
	return &repository, err
}

// ListRepositories lists repositories for the authenticated user
func (c *Client) ListRepositories(visibility string, affiliation string, page int) ([]Repository, error) {
	// WHO: RepositoryLister
	// WHAT: List repositories
	// WHEN: During repo listing operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To fetch multiple repositories
	// HOW: Using GitHub API
	// EXTENT: Multiple repositories

	c.Logger.Debug("Listing repositories", "visibility", visibility, "affiliation", affiliation, "page", page)

	var repos []Repository

	// Build query parameters
	query := "?"
	if visibility != "" {
		query += "visibility=" + visibility + "&"
	}
	if affiliation != "" {
		query += "affiliation=" + affiliation + "&"
	}
	if page > 0 {
		query += fmt.Sprintf("page=%d&per_page=100&", page)
	} else {
		query += "per_page=100&"
	}

	// Trim trailing &
	if query[len(query)-1] == '&' {
		query = query[:len(query)-1]
	}

	path := ReposEndpoint + query
	err := c.Request("GET", path, nil, &repos)
	return repos, err
}

// GetIssue gets a specific issue from a repository
func (c *Client) GetIssue(owner, repo string, issueNumber int) (*Issue, error) {
	// WHO: IssueFetcher
	// WHAT: Get issue details
	// WHEN: During issue operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To fetch issue information
	// HOW: Using GitHub API
	// EXTENT: Single issue

	c.Logger.Debug("Fetching issue", "owner", owner, "repo", repo, "issue", issueNumber)

	var issue Issue
	path := fmt.Sprintf("%s/%s/%s/issues/%d", ReposEndpoint, owner, repo, issueNumber)

	err := c.Request("GET", path, nil, &issue)
	return &issue, err
}

// ListIssues lists issues for a repository
func (c *Client) ListIssues(owner, repo string, state string, page int) ([]Issue, error) {
	// WHO: IssueLister
	// WHAT: List repository issues
	// WHEN: During issues listing
	// WHERE: System Layer 6 (Integration)
	// WHY: To fetch multiple issues
	// HOW: Using GitHub API
	// EXTENT: Multiple issues

	c.Logger.Debug("Listing issues", "owner", owner, "repo", "state", state, "page", page)

	var issues []Issue

	// Build query parameters
	query := "?"
	if state != "" {
		query += "state=" + state + "&"
	}
	if page > 0 {
		query += fmt.Sprintf("page=%d&per_page=100&", page)
	} else {
		query += "per_page=100&"
	}

	// Trim trailing &
	if query[len(query)-1] == '&' {
		query = query[:len(query)-1]
	}

	path := fmt.Sprintf("%s/%s/%s/issues%s", ReposEndpoint, owner, repo, query)
	err := c.Request("GET", path, nil, &issues)
	return issues, err
}

// GetPullRequest gets a specific pull request
func (c *Client) GetPullRequest(owner, repo string, prNumber int) (*PullRequest, error) {
	// WHO: PRFetcher
	// WHAT: Get pull request details
	// WHEN: During PR operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To fetch PR information
	// HOW: Using GitHub API
	// EXTENT: Single pull request

	c.Logger.Debug("Fetching pull request", "owner", owner, "repo", "pr", prNumber)

	var pr PullRequest
	path := fmt.Sprintf("%s/%s/%s/pulls/%d", ReposEndpoint, owner, repo, prNumber)

	err := c.Request("GET", path, nil, &pr)
	return &pr, err
}

// ListPullRequests lists pull requests for a repository
func (c *Client) ListPullRequests(owner, repo string, state string, page int) ([]PullRequest, error) {
	// WHO: PRLister
	// WHAT: List repository pull requests
	// WHEN: During PR listing
	// WHERE: System Layer 6 (Integration)
	// WHY: To fetch multiple PRs
	// HOW: Using GitHub API
	// EXTENT: Multiple pull requests

	c.Logger.Debug("Listing pull requests", "owner", owner, "repo", "state", state, "page", page)

	var prs []PullRequest

	// Build query parameters
	query := "?"
	if state != "" {
		query += "state=" + state + "&"
	}
	if page > 0 {
		query += fmt.Sprintf("page=%d&per_page=100&", page)
	} else {
		query += "per_page=100&"
	}

	// Trim trailing &
	if query[len(query)-1] == '&' {
		query = query[:len(query)-1]
	}

	path := fmt.Sprintf("%s/%s/%s/pulls%s", ReposEndpoint, owner, repo, query)
	err := c.Request("GET", path, nil, &prs)
	return prs, err
}

// SearchCode searches for code on GitHub
func (c *Client) SearchCode(query string, page int) (*SearchResult, error) {
	// WHO: CodeSearcher
	// WHAT: Search GitHub code
	// WHEN: During search operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To find code matches
	// HOW: Using GitHub search API
	// EXTENT: Multiple code files

	c.Logger.Debug("Searching code", "query", query, "page", page)

	var result SearchResult

	// Build query parameters
	params := "?q=" + query
	if page > 0 {
		params += fmt.Sprintf("&page=%d&per_page=100", page)
	} else {
		params += "&per_page=100"
	}

	path := SearchEndpoint + "/code" + params
	err := c.Request("GET", path, nil, &result)
	return &result, err
}

// GetCodeScanningAlerts gets code scanning alerts for a repository
func (c *Client) GetCodeScanningAlerts(owner, repo string) ([]CodeScanningAlert, error) {
	// WHO: SecurityScanner
	// WHAT: Get code scanning alerts
	// WHEN: During security operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To check for vulnerabilities
	// HOW: Using GitHub security API
	// EXTENT: Repository security issues

	c.Logger.Debug("Fetching code scanning alerts", "owner", owner, "repo", repo)

	var alerts []CodeScanningAlert
	path := fmt.Sprintf("%s/%s/%s/code-scanning/alerts", ReposEndpoint, owner, repo)

	err := c.Request("GET", path, nil, &alerts)
	return alerts, err
}

// GetUser gets a GitHub user
func (c *Client) GetUser(username string) (*User, error) {
	// WHO: UserFetcher
	// WHAT: Get user details
	// WHEN: During user operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To fetch user information
	// HOW: Using GitHub API
	// EXTENT: Single user

	c.Logger.Debug("Fetching user", "username", username)

	var user User
	path := "/users/" + username

	err := c.Request("GET", path, nil, &user)
	return &user, err
}

// GetAuthenticatedUser gets the currently authenticated user
func (c *Client) GetAuthenticatedUser() (*User, error) {
	// WHO: AuthManager
	// WHAT: Get authenticated user
	// WHEN: During auth operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To fetch current user
	// HOW: Using GitHub API
	// EXTENT: Current user

	c.Logger.Debug("Fetching authenticated user")

	var user User
	err := c.Request("GET", UserEndpoint, nil, &user)
	return &user, err
}

// Repository represents a GitHub repository
type Repository struct {
	// WHO: RepoManager
	// WHAT: Repository data structure
	// WHEN: During repo operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To store repo information
	// HOW: Using structured data
	// EXTENT: Single repository data

	ID          int64  `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
	Owner       struct {
		Login string `json:"login"`
		ID    int64  `json:"id"`
	} `json:"owner"`
	DefaultBranch string   `json:"default_branch"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
	URL           string   `json:"url"`
	HTMLURL       string   `json:"html_url"`
	Topics        []string `json:"topics"`
}

// Issue represents a GitHub issue
type Issue struct {
	// WHO: IssueManager
	// WHAT: Issue data structure
	// WHEN: During issue operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To store issue information
	// HOW: Using structured data
	// EXTENT: Single issue data

	ID        int64  `json:"id"`
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	User      struct {
		Login string `json:"login"`
		ID    int64  `json:"id"`
	} `json:"user"`
	Labels []struct {
		Name        string `json:"name"`
		Color       string `json:"color"`
		Description string `json:"description"`
	} `json:"labels"`
	Assignees []struct {
		Login string `json:"login"`
		ID    int64  `json:"id"`
	} `json:"assignees"`
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	// WHO: PRManager
	// WHAT: Pull request data structure
	// WHEN: During PR operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To store PR information
	// HOW: Using structured data
	// EXTENT: Single PR data

	ID        int64  `json:"id"`
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	User      struct {
		Login string `json:"login"`
		ID    int64  `json:"id"`
	} `json:"user"`
	Base struct {
		Ref  string `json:"ref"`
		Repo struct {
			FullName string `json:"full_name"`
			ID       int64  `json:"id"`
		} `json:"repo"`
	} `json:"base"`
	Head struct {
		Ref  string `json:"ref"`
		Repo struct {
			FullName string `json:"full_name"`
			ID       int64  `json:"id"`
		} `json:"repo"`
	} `json:"head"`
	Draft  bool `json:"draft"`
	Merged bool `json:"merged"`
}

// User represents a GitHub user
type User struct {
	// WHO: UserManager
	// WHAT: User data structure
	// WHEN: During user operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To store user information
	// HOW: Using structured data
	// EXTENT: Single user profile

	Login       string `json:"login"`
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	AvatarURL   string `json:"avatar_url"`
	Bio         string `json:"bio"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	Followers   int    `json:"followers"`
	Following   int    `json:"following"`
	PublicRepos int    `json:"public_repos"`
}

// SearchResult represents GitHub search results
type SearchResult struct {
	// WHO: SearchManager
	// WHAT: Search result structure
	// WHEN: During search operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To store search results
	// HOW: Using structured data
	// EXTENT: Multiple search items

	TotalCount        int           `json:"total_count"`
	IncompleteResults bool          `json:"incomplete_results"`
	Items             []interface{} `json:"items"`
}

// CodeScanningAlert represents a GitHub code scanning alert
type CodeScanningAlert struct {
	// WHO: SecurityManager
	// WHAT: Security alert structure
	// WHEN: During security operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To store vulnerability data
	// HOW: Using structured data
	// EXTENT: Single security issue

	Number                int    `json:"number"`
	State                 string `json:"state"`
	Severity              string `json:"severity"`
	SecurityVulnerability struct {
		Package struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"package"`
		Severity string `json:"severity"`
		CVSS     struct {
			Score float64 `json:"score"`
		} `json:"cvss"`
	} `json:"security_vulnerability"`
	CreatedAt   string `json:"created_at"`
	DismissedAt string `json:"dismissed_at"`
	DismissedBy struct {
		Login string `json:"login"`
	} `json:"dismissed_by"`
	DismissedReason string `json:"dismissed_reason"`
}

// APIError represents a GitHub API error
type APIError struct {
	// WHO: ErrorManager
	// WHAT: API error structure
	// WHEN: During API error handling
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide structured error data
	// HOW: Using error details
	// EXTENT: API error information

	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url"`
	Errors           []struct {
		Resource string `json:"resource"`
		Field    string `json:"field"`
		Code     string `json:"code"`
	} `json:"errors"`
	StatusCode int `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("GitHub API error: %s (Status: %d)", e.Message, e.StatusCode)
}

// CompressContext compresses the client context using Möbius formula
func (c *Client) CompressContext() map[string]interface{} {
	// WHO: ContextCompressor
	// WHAT: Compress client context
	// WHEN: During context operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To optimize context storage
	// HOW: Using Möbius compression
	// EXTENT: Client context data

	if c.Context == nil || c.Context.Meta == nil {
		return nil
	}

	// Extract contextual factors
	B := getMetaFloat(c.Context.Meta, "B", 0.8) // Base factor
	V := getMetaFloat(c.Context.Meta, "V", 0.7) // Value factor
	I := getMetaFloat(c.Context.Meta, "I", 0.9) // Intent factor
	G := getMetaFloat(c.Context.Meta, "G", 1.2) // Growth factor
	F := getMetaFloat(c.Context.Meta, "F", 0.6) // Flexibility factor

	// Calculate time factor (how "fresh" the context is)
	now := time.Now().Unix()
	t := float64(now-c.Context.When) / 86400.0 // Days
	if t < 0 {
		t = 0
	}

	// Calculate energy factor (computational cost)
	E := 0.5

	// Calculate context entropy (simplified)
	contextBytes, _ := json.Marshal(c.Context)
	entropy := float64(len(contextBytes)) / 100.0

	// Apply Möbius compression formula
	alignment := (B + V*I) * math.Exp(-t*E)
	compressionFactor := (B * I * (1.0 - (entropy / math.Log2(1.0+V))) * (G + F)) /
		(E*t + entropy + alignment)

	// Create compressed context
	compressed := map[string]interface{}{
		"who":    c.Context.Who,
		"what":   c.Context.What,
		"when":   c.Context.When,
		"where":  c.Context.Where,
		"why":    c.Context.Why,
		"how":    c.Context.How,
		"extent": c.Context.Extent,
		"meta": map[string]interface{}{
			"B":                 B,
			"V":                 V,
			"I":                 I,
			"G":                 G,
			"F":                 F,
			"t":                 t,
			"E":                 E,
			"entropy":           entropy,
			"compressionFactor": compressionFactor,
			"compressedAt":      now,
		},
	}

	return compressed
}

// Helper to extract float from metadata
func getMetaFloat(meta map[string]interface{}, key string, defaultValue float64) float64 {
	if meta == nil {
		return defaultValue
	}

	if val, ok := meta[key]; ok {
		switch v := val.(type) {
		case int:
			return float64(v)
		case int32:
			return float64(v)
		case int64:
			return float64(v)
		case float32:
			return float64(v)
		case float64:
			return v
		}
	}

	return defaultValue
}
