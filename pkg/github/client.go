/*
 * WHO: GitHubClient
 * WHAT: GitHub API client for MCP server
 * WHEN: During GitHub API operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide GitHub API access with context awareness
 * HOW: Using REST API with 7D context and compression
 * EXTENT: All GitHub interactions
 */

package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/github/github-mcp-server/pkg/log"
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

	BaseURL          = "https://api.github.com"
	ReposEndpoint    = "/repos"
	IssuesEndpoint   = "/issues"
	PullsEndpoint    = "/pulls"
	SearchEndpoint   = "/search"
	CodeScanEndpoint = "/code-scanning"
	UserEndpoint     = "/user"
)

// Client represents a GitHub API client
type Client struct {
	// WHO: ClientManager
	// WHAT: GitHub API client
	// WHEN: During API operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To manage GitHub connections
	// HOW: Using HTTP client with auth and context
	// EXTENT: All API interactions

	BaseURL    string
	HTTPClient *http.Client
	Token      string
	Logger     *log.Logger
	Context    ContextVector7D
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

// NewClient creates a new GitHub API client with context
func NewClient(token string, logger *log.Logger) *Client {
	// WHO: ClientCreator
	// WHAT: Initialize GitHub client
	// WHEN: During server startup
	// WHERE: System Layer 6 (Integration)
	// WHY: To create API client
	// HOW: Using provided credentials
	// EXTENT: Client lifecycle

	return &Client{
		BaseURL: BaseURL,
		HTTPClient: &http.Client{
			Timeout: time.Second * 30,
		},
		Token:  token,
		Logger: logger,
		Context: ContextVector7D{
			Who:    "GitHubClient",
			What:   "APIClient",
			When:   time.Now().Unix(),
			Where:  "System Layer 6 (Integration)",
			Why:    "GitHub API Access",
			How:    "REST API",
			Extent: 1.0,
			Meta: map[string]interface{}{
				"B": 0.8, // Base factor
				"V": 0.7, // Value factor
				"I": 0.9, // Intent factor
				"G": 1.2, // Growth factor
				"F": 0.6, // Flexibility factor
			},
		},
	}
}

// Request makes an authenticated request to the GitHub API
func (c *Client) Request(method, path string, body interface{}, target interface{}) error {
	// WHO: RequestManager
	// WHAT: Make API request
	// WHEN: During API operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To send requests to GitHub
	// HOW: Using HTTP with context
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

	// Add compressed context to custom header for tracking
	contextBytes, _ := json.Marshal(c.CompressContext())
	req.Header.Set("X-TNOS-Context", string(contextBytes))

	// Execute request with timing for metrics
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

	// Check for error status codes
	if resp.StatusCode >= 400 {
		var errorResponse struct {
			Message string `json:"message"`
		}

		// Try to get error message
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			return fmt.Errorf("API error: %s (status %d)", errorResponse.Message, resp.StatusCode)
		}

		return fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	// Parse response if target is provided
	if target != nil {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	// Update context with latest operation
	c.Context.What = fmt.Sprintf("API_%s", method)
	c.Context.When = time.Now().Unix()

	return nil
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

// GetRepository gets a GitHub repository
func (c *Client) GetRepository(owner, repo string) (Repository, error) {
	// WHO: RepositoryFetcher
	// WHAT: Get repository details
	// WHEN: During repo operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To fetch repo information
	// HOW: Using GitHub API
	// EXTENT: Single repository

	var repository Repository
	path := fmt.Sprintf("%s/%s/%s", ReposEndpoint, owner, repo)

	err := c.Request("GET", path, nil, &repository)
	return repository, err
}

// SearchCode searches GitHub code
func (c *Client) SearchCode(query string) (SearchResult, error) {
	// WHO: CodeSearcher
	// WHAT: Search GitHub code
	// WHEN: During search operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To find code matches
	// HOW: Using GitHub search API
	// EXTENT: Multiple code files

	var result SearchResult
	path := fmt.Sprintf("%s/code?q=%s", SearchEndpoint, query)

	err := c.Request("GET", path, nil, &result)
	return result, err
}

// GetIssues gets issues for a repository
func (c *Client) GetIssues(owner, repo string) ([]Issue, error) {
	// WHO: IssueManager
	// WHAT: Get repository issues
	// WHEN: During issue operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To fetch issue data
	// HOW: Using GitHub API
	// EXTENT: Repository issues

	var issues []Issue
	path := fmt.Sprintf("%s/%s/%s/issues", ReposEndpoint, owner, repo)

	err := c.Request("GET", path, nil, &issues)
	return issues, err
}

// GetPullRequests gets pull requests for a repository
func (c *Client) GetPullRequests(owner, repo string) ([]PullRequest, error) {
	// WHO: PRManager
	// WHAT: Get pull requests
	// WHEN: During PR operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To fetch PR data
	// HOW: Using GitHub API
	// EXTENT: Repository PRs

	var prs []PullRequest
	path := fmt.Sprintf("%s/%s/%s/pulls", ReposEndpoint, owner, repo)

	err := c.Request("GET", path, nil, &prs)
	return prs, err
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

	var alerts []CodeScanningAlert
	path := fmt.Sprintf("%s/%s/%s%s/alerts", ReposEndpoint, owner, repo, CodeScanEndpoint)

	err := c.Request("GET", path, nil, &alerts)
	return alerts, err
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
	} `json:"owner"`
	DefaultBranch string `json:"default_branch"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	URL           string `json:"url"`
	HTMLURL       string `json:"html_url"`
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
	} `json:"user"`
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
	} `json:"user"`
	Base struct {
		Ref  string `json:"ref"`
		Repo struct {
			FullName string `json:"full_name"`
		} `json:"repo"`
	} `json:"base"`
	Head struct {
		Ref  string `json:"ref"`
		Repo struct {
			FullName string `json:"full_name"`
		} `json:"repo"`
	} `json:"head"`
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

	TotalCount int           `json:"total_count"`
	Items      []interface{} `json:"items"`
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
	CreatedAt string `json:"created_at"`
}
