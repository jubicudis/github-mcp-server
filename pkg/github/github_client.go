/*
 * WHO: GitHubClient
 * WHAT: GitHub API integration for MCP
 * WHEN: During GitHub operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To interface with GitHub API
 * HOW: Using REST and GraphQL APIs
 * EXTENT: All GitHub interactions
 */

package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"strings"
	"sync"
	"time"

	"github-mcp-server/pkg/common"
	"github-mcp-server/pkg/log"
	"github-mcp-server/pkg/translations"

	"github.com/mark3labs/mcp-go/mcp"

	"encoding/base64"

	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
)

// We use translations.ContextVector7D defined in the translations package
// Use translations.ToMap() function to convert the context vector to a map

// We're using the Logger interface defined in client_adapter.go
// Additional methods can be handled through interface composition if needed

// DefaultAPI constants
const (
	// WHO: ConfigManager
	// WHAT: GitHub API constants
	// WHEN: During API operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To define API parameters
	// HOW: Using constant definitions
	// EXTENT: All API operations

	DefaultAPIBaseURL     = "https://api.github.com"
	DefaultGraphQLBaseURL = "https://api.github.com/graphql"
	DefaultAcceptHeader   = "application/vnd.github.v3+json"
	DefaultUserAgent      = "TNOS-GitHub-MCP-Client"
	DefaultTimeout        = 30 * time.Second
	DefaultCacheTimeout   = 5 * time.Minute
)

// Resource type constants
const (
	// WHO: ResourceManager
	// WHAT: Resource type constants
	// WHEN: During resource operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To categorize resources
	// HOW: Using type definitions
	// EXTENT: All resource types

	ResourceTypeUser       = "user"
	ResourceTypeRepo       = "repo"
	ResourceTypeIssue      = "issue"
	ResourceTypePR         = "pr"
	ResourceTypeComment    = "comment"
	ResourceTypeContent    = "content"
	ResourceTypeRef        = "ref"
	ResourceTypeCommit     = "commit"
	ResourceTypeBranch     = "branch"
	ResourceTypeWorkflow   = "workflow"
	ResourceTypeRelease    = "release"
	ResourceTypeTag        = "tag"
	ResourceTypeCodeScan   = "codescan"
	ResourceTypeDependency = "dependency"
	ResourceTypeSecret     = "secret"
	ResourceTypeAction     = "action"
)

// ClientOptions configures the GitHub client
type ClientOptions struct {
	// WHO: ConfigManager
	// WHAT: Client configuration
	// WHEN: During client initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To configure client behavior
	// HOW: Using structured options
	// EXTENT: All client parameters

	Token           string
	APIBaseURL      string
	GraphQLBaseURL  string
	AcceptHeader    string
	UserAgent       string
	Timeout         time.Duration
	Logger          Logger
	EnableCache     bool
	CacheTimeout    time.Duration
	RateLimitBuffer int
}

// Client represents a GitHub API client
type Client struct {
	// WHO: ClientManager
	// WHAT: GitHub client
	// WHEN: During API operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To communicate with GitHub
	// HOW: Using HTTP and GraphQL
	// EXTENT: All GitHub operations

	options     ClientOptions
	httpClient  *http.Client
	graphClient *graphql.Client
	baseURL     *url.URL
	graphQLURL  *url.URL

	// Context for requests
	context *translations.ContextVector7D

	// Cache for common operations
	cache         map[string]*cacheItem
	cacheMutex    sync.RWMutex
	cacheDisabled bool

	// Rate limiting
	rateLimitMutex  sync.RWMutex
	rateLimitReset  time.Time
	rateLimitRemain int
}

// cacheItem represents a cached API response
type cacheItem struct {
	// WHO: CacheManager
	// WHAT: Cache entry
	// WHEN: During caching operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To store API responses
	// HOW: Using cached data structure
	// EXTENT: Single cache entry

	data      interface{}
	timestamp time.Time
	expires   time.Time
}

// NewAdvancedClient creates a new GitHub client with advanced options
// This is renamed from NewClient to avoid conflicts with the legacy client
func NewAdvancedClient(options ClientOptions) (*Client, error) {
	// WHO: ClientCreator
	// WHAT: Client creation
	// WHEN: During system initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To establish GitHub connection
	// HOW: Using provided options
	// EXTENT: Client creation

	// Set defaults
	if options.APIBaseURL == "" {
		options.APIBaseURL = DefaultAPIBaseURL
	}

	if options.GraphQLBaseURL == "" {
		options.GraphQLBaseURL = DefaultGraphQLBaseURL
	}

	if options.AcceptHeader == "" {
		options.AcceptHeader = DefaultAcceptHeader
	}

	if options.UserAgent == "" {
		options.UserAgent = DefaultUserAgent
	}

	if options.Timeout == 0 {
		options.Timeout = DefaultTimeout
	}

	if options.CacheTimeout == 0 {
		options.CacheTimeout = DefaultCacheTimeout
	}

	// Parse URLs
	baseURL, err := url.Parse(options.APIBaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid API URL: %w", err)
	}

	graphQLURL, err := url.Parse(options.GraphQLBaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid GraphQL URL: %w", err)
	}

	// Configure OAuth2 client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: options.Token},
	)
	httpClient := oauth2.NewClient(context.Background(), ts)
	httpClient.Timeout = options.Timeout

	// Create GraphQL client
	graphClient := graphql.NewClient(options.GraphQLBaseURL, httpClient)

	// Create 7D context for the client
	now := time.Now().Unix()
	contextVector := &translations.ContextVector7D{
		Who:    "GitHubClient",
		What:   "GitHubIntegration",
		When:   now,
		Where:  "System Layer 6 (Integration)",
		Why:    "GitHub API Communication",
		How:    "REST and GraphQL APIs",
		Extent: 1.0,
		Meta: map[string]interface{}{
			"B":         0.9, // Base factor
			"V":         0.8, // Value factor
			"I":         0.7, // Intent factor
			"G":         0.6, // Growth factor
			"F":         0.5, // Flexibility factor
			"createdAt": now,
			"Source":    "github_mcp",
		},
	}

	return &Client{
		options:     options,
		httpClient:  httpClient,
		graphClient: graphClient,
		baseURL:     baseURL,
		graphQLURL:  graphQLURL,
		context:     contextVector,
		cache:       make(map[string]*cacheItem),
	}, nil
}

// SetContext updates the client context
func (c *Client) SetContext(context *translations.ContextVector7D) {
	// WHO: ContextManager
	// WHAT: Update context
	// WHEN: During context changes
	// WHERE: System Layer 6 (Integration)
	// WHY: To update operation context
	// HOW: Using context replacement
	// EXTENT: All client operations

	c.context = context
}

// GetRepositoryByName fetches repository information
func (c *Client) GetRepositoryByName(ctx context.Context, owner, repo string) (map[string]interface{}, error) {
	// WHO: RepositoryFetcher
	// WHAT: Fetch repository
	// WHEN: During repository operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To retrieve repository data
	// HOW: Using REST API
	// EXTENT: Single repository

	// Check cache first
	cacheKey := fmt.Sprintf("repo:%s/%s", owner, repo)
	cached, found := c.getCachedItem(cacheKey)
	if found {
		if c.options.Logger != nil {
			c.options.Logger.Debug("Cache hit", "key", cacheKey)
		}
		return cached.(map[string]interface{}), nil
	}

	if c.options.Logger != nil {
		c.options.Logger.Info("Fetching repository",
			"owner", owner,
			"repo", repo)
	}

	// Make API request
	path := fmt.Sprintf("/repos/%s/%s", owner, repo)
	data, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository: %w", err)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse repository data: %w", err)
	}

	// Cache the result
	c.cacheItem(cacheKey, result)

	return result, nil
}

// GetFileContent fetches file content from a repository
func (c *Client) GetFileContent(ctx context.Context, owner, repo, path, ref string) ([]byte, map[string]interface{}, error) {
	// WHO: ContentFetcher
	// WHAT: Fetch file content
	// WHEN: During content operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To retrieve file data
	// HOW: Using REST API
	// EXTENT: Single file

	if c.options.Logger != nil {
		c.options.Logger.Info("Fetching file content",
			"owner", owner,
			"repo", repo,
			"path", path,
			"ref", ref)
	}

	// Build API path
	apiPath := fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path)
	query := url.Values{}
	if ref != "" {
		query.Set("ref", ref)
	}

	// Make API request
	data, err := c.doRequest(ctx, http.MethodGet, apiPath+"?"+query.Encode(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch file content: %w", err)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, nil, fmt.Errorf("failed to parse file data: %w", err)
	}

	// Parse content
	content, ok := result["content"].(string)
	if !ok {
		return nil, result, fmt.Errorf("content field missing or invalid")
	}

	// Decode content from Base64
	decoded, err := decodeBase64Content(content)
	if err != nil {
		return nil, result, fmt.Errorf("failed to decode content: %w", err)
	}

	return decoded, result, nil
}

// GetDirectoryContent fetches directory content from a repository
func (c *Client) GetDirectoryContent(ctx context.Context, owner, repo, path, ref string) ([]map[string]interface{}, error) {
	// WHO: DirectoryFetcher
	// WHAT: Fetch directory content
	// WHEN: During content operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To list directory contents
	// HOW: Using REST API
	// EXTENT: Directory listing

	if c.options.Logger != nil {
		c.options.Logger.Info("Fetching directory content",
			"owner", owner,
			"repo", repo,
			"path", path,
			"ref", ref)
	}

	// Build API path
	apiPath := fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path)
	query := url.Values{}
	if ref != "" {
		query.Set("ref", ref)
	}

	// Make API request
	data, err := c.doRequest(ctx, http.MethodGet, apiPath+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch directory content: %w", err)
	}

	// Try to parse as array first (directory)
	var directoryItems []map[string]interface{}
	if err := json.Unmarshal(data, &directoryItems); err != nil {
		// Try to parse as single file
		var fileItem map[string]interface{}
		if err := json.Unmarshal(data, &fileItem); err != nil {
			return nil, fmt.Errorf("failed to parse directory data: %w", err)
		}

		// Return single file as array
		return []map[string]interface{}{fileItem}, nil
	}

	return directoryItems, nil
}

// CreateIssue creates a new issue in a repository
func (c *Client) CreateIssue(ctx context.Context, owner, repo string, issue map[string]interface{}) (map[string]interface{}, error) {
	// WHO: IssueCreator
	// WHAT: Create issue
	// WHEN: During issue operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To create new issues
	// HOW: Using REST API
	// EXTENT: Single issue creation

	title, _ := issue["title"].(string)
	if c.options.Logger != nil {
		c.options.Logger.Info("Creating issue",
			"owner", owner,
			"repo", repo,
			"title", title)
	}

	// Validate required fields
	if _, ok := issue["title"]; !ok {
		return nil, fmt.Errorf("title is required for creating an issue")
	}

	// Make API request
	apiPath := fmt.Sprintf("/repos/%s/%s/issues", owner, repo)
	data, err := c.doRequest(ctx, http.MethodPost, apiPath, issue)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse issue data: %w", err)
	}

	return result, nil
}

// UpdateIssue updates an existing issue
func (c *Client) UpdateIssue(ctx context.Context, owner, repo string, number int, issue map[string]interface{}) (map[string]interface{}, error) {
	// WHO: IssueUpdater
	// WHAT: Update issue
	// WHEN: During issue operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To modify existing issues
	// HOW: Using REST API
	// EXTENT: Single issue update

	if c.options.Logger != nil {
		c.options.Logger.Info("Updating issue",
			"owner", owner,
			"repo", repo,
			"number", number)
	}

	// Make API request
	apiPath := fmt.Sprintf("/repos/%s/%s/issues/%d", owner, repo, number)
	data, err := c.doRequest(ctx, http.MethodPatch, apiPath, issue)
	if err != nil {
		return nil, fmt.Errorf("failed to update issue: %w", err)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse issue data: %w", err)
	}

	return result, nil
}

// CreatePullRequest creates a new pull request
func (c *Client) CreatePullRequest(ctx context.Context, owner, repo string, pr map[string]interface{}) (map[string]interface{}, error) {
	// WHO: PRCreator
	// WHAT: Create pull request
	// WHEN: During PR operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To create new PRs
	// HOW: Using REST API
	// EXTENT: Single PR creation

	title, _ := pr["title"].(string)
	head, _ := pr["head"].(string)
	base, _ := pr["base"].(string)

	if c.options.Logger != nil {
		c.options.Logger.Info("Creating pull request",
			"owner", owner,
			"repo", repo,
			"title", title,
			"head", head,
			"base", base)
	}

	// Validate required fields
	required := []string{"title", "head", "base"}
	for _, field := range required {
		if _, ok := pr[field]; !ok {
			return nil, fmt.Errorf("%s is required for creating a pull request", field)
		}
	}

	// Make API request
	apiPath := fmt.Sprintf("/repos/%s/%s/pulls", owner, repo)
	data, err := c.doRequest(ctx, http.MethodPost, apiPath, pr)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse pull request data: %w", err)
	}

	return result, nil
}

// GetBranches fetches branches from a repository
func (c *Client) GetBranches(ctx context.Context, owner, repo string) ([]map[string]interface{}, error) {
	// WHO: BranchFetcher
	// WHAT: Fetch branches
	// WHEN: During branch operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To list repository branches
	// HOW: Using REST API
	// EXTENT: Branch listing

	if c.options.Logger != nil {
		c.options.Logger.Info("Fetching branches",
			"owner", owner,
			"repo", repo)
	}

	// Make API request
	apiPath := fmt.Sprintf("/repos/%s/%s/branches", owner, repo)
	data, err := c.doRequest(ctx, http.MethodGet, apiPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch branches: %w", err)
	}

	// Parse response
	var result []map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse branches data: %w", err)
	}

	return result, nil
}

// GetCommit fetches a specific commit
func (c *Client) GetCommit(ctx context.Context, owner, repo, sha string) (map[string]interface{}, error) {
	// WHO: CommitFetcher
	// WHAT: Fetch commit
	// WHEN: During commit operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To retrieve commit data
	// HOW: Using REST API
	// EXTENT: Single commit

	// Check cache first
	cacheKey := fmt.Sprintf("commit:%s/%s/%s", owner, repo, sha)
	cached, found := c.getCachedItem(cacheKey)
	if found {
		if c.options.Logger != nil {
			c.options.Logger.Debug("Cache hit", "key", cacheKey)
		}
		return cached.(map[string]interface{}), nil
	}

	if c.options.Logger != nil {
		c.options.Logger.Info("Fetching commit",
			"owner", owner,
			"repo", repo,
			"sha", sha)
	}

	// Make API request
	apiPath := fmt.Sprintf("/repos/%s/%s/commits/%s", owner, repo, sha)
	data, err := c.doRequest(ctx, http.MethodGet, apiPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch commit: %w", err)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse commit data: %w", err)
	}

	// Cache the result
	c.cacheItem(cacheKey, result)

	return result, nil
}

// SearchCode searches for code in repositories
func (c *Client) SearchCode(ctx context.Context, query string, options map[string]string) (map[string]interface{}, error) {
	// WHO: CodeSearcher
	// WHAT: Search code
	// WHEN: During search operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To find code in repositories
	// HOW: Using search API
	// EXTENT: Code search results

	if c.options.Logger != nil {
		c.options.Logger.Info("Searching code",
			"query", query)
	}

	// Build query parameters
	params := url.Values{}
	params.Set("q", query)

	// Add optional parameters
	for key, value := range options {
		params.Set(key, value)
	}

	// Make API request
	apiPath := "/search/code?" + params.Encode()
	data, err := c.doRequest(ctx, http.MethodGet, apiPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search code: %w", err)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	return result, nil
}

// GetCodeScanningAlerts fetches code scanning alerts
func (c *Client) GetCodeScanningAlerts(ctx context.Context, owner, repo string) ([]map[string]interface{}, error) {
	// WHO: SecurityAnalyzer
	// WHAT: Fetch security alerts
	// WHEN: During security operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To retrieve security data
	// HOW: Using security API
	// EXTENT: Security alerts

	if c.options.Logger != nil {
		c.options.Logger.Info("Fetching code scanning alerts",
			"owner", owner,
			"repo", repo)
	}

	// Make API request with preview header
	headers := map[string]string{
		"Accept": "application/vnd.github.v3+json",
	}

	apiPath := fmt.Sprintf("/repos/%s/%s/code-scanning/alerts", owner, repo)
	data, err := c.doRequestWithHeaders(ctx, http.MethodGet, apiPath, nil, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch code scanning alerts: %w", err)
	}

	// Parse response
	var result []map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse alerts data: %w", err)
	}

	return result, nil
}

// Helper functions

// doRequest performs an HTTP request to the GitHub API
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	// WHO: APIRequester
	// WHAT: Perform API request
	// WHEN: During API operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To communicate with GitHub
	// HOW: Using HTTP protocol
	// EXTENT: Single API request

	return c.doRequestWithHeaders(ctx, method, path, body, nil)
}

// doRequestWithHeaders performs an HTTP request with custom headers
func (c *Client) doRequestWithHeaders(ctx context.Context, method, path string, body interface{}, headers map[string]string) ([]byte, error) {
	// WHO: HeadersManager
	// WHAT: Request with headers
	// WHEN: During API operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To set request headers
	// HOW: Using HTTP headers
	// EXTENT: Single API request

	// Rate limit check
	if err := c.checkRateLimit(); err != nil {
		return nil, err
	}

	// Prepare request
	req, err := c.prepareHTTPRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}

	// Set headers
	c.setHTTPRequestHeaders(req, headers, body != nil)

	// Execute request
	if c.options.Logger != nil {
		c.options.Logger.Debug("Making API request",
			"method", method,
			"path", path)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limit info
	c.updateRateLimit(resp)

	return c.processHTTPResponse(resp)
}

// prepareHTTPRequest creates and prepares an HTTP request
func (c *Client) prepareHTTPRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	// WHO: RequestPreparer
	// WHAT: Prepare HTTP request
	// WHEN: Before API operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To build proper request
	// HOW: Using HTTP request creation
	// EXTENT: Single API request

	// Create request URL
	reqURL, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid request path: %w", err)
	}

	// Join with base URL if path is not absolute
	if !reqURL.IsAbs() {
		reqURL = c.baseURL.ResolveReference(reqURL)
	}

	// Prepare request body
	var reqBody io.Reader
	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(bodyJSON)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}

// setHTTPRequestHeaders sets headers for the HTTP request
func (c *Client) setHTTPRequestHeaders(req *http.Request, headers map[string]string, hasBody bool) {
	// WHO: HeaderSetter
	// WHAT: Set request headers
	// WHEN: Before API operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To configure request headers
	// HOW: Using header configuration
	// EXTENT: Single API request

	// Set default headers
	req.Header.Set("Accept", c.options.AcceptHeader)
	req.Header.Set("User-Agent", c.options.UserAgent)

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set content type for requests with body
	if hasBody {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add context as header
	if c.context != nil {
		contextJSON, _ := json.Marshal(c.context.ToMap())
		req.Header.Set("X-MCP-Context", string(contextJSON))
	}
}

// processHTTPResponse processes the HTTP response
func (c *Client) processHTTPResponse(resp *http.Response) ([]byte, error) {
	// WHO: ResponseProcessor
	// WHAT: Process HTTP response
	// WHEN: After API operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To handle API responses
	// HOW: Using response parsing
	// EXTENT: Single API response

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error response
	if resp.StatusCode >= 400 {
		return nil, c.createErrorFromResponse(resp.StatusCode, respBody)
	}

	return respBody, nil
}

// createErrorFromResponse creates an error from an error response
func (c *Client) createErrorFromResponse(statusCode int, respBody []byte) error {
	// WHO: ErrorProcessor
	// WHAT: Create error from response
	// WHEN: During error handling
	// WHERE: System Layer 6 (Integration)
	// WHY: To standardize error format
	// HOW: Using error parsing
	// EXTENT: Single API error

	var errResp struct {
		Message string `json:"message"`
		Errors  []struct {
			Resource string `json:"resource"`
			Field    string `json:"field"`
			Code     string `json:"code"`
		} `json:"errors,omitempty"`
	}

	if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Message != "" {
		return fmt.Errorf("GitHub API error (%d): %s", statusCode, errResp.Message)
	}

	return fmt.Errorf("GitHub API error: status %d", statusCode)
}

// Request is a compatibility method for legacy code
// It provides a simplified interface to perform HTTP requests
// WHO: CompatibilityAdapter
// WHAT: Legacy request adapter
// WHEN: During API operations
// WHERE: System Layer 6 (Integration)
// WHY: To maintain compatibility
// HOW: Using existing doRequest method
// EXTENT: All legacy request operations
func (c *Client) Request(ctx context.Context, method string, path string, body interface{}, target interface{}) error {
	// WHO: CompatibilityAdapter
	// WHAT: Use provided context for request
	// WHEN: During API operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To propagate deadlines/cancellations
	// HOW: Pass context to doRequest
	// EXTENT: All legacy request operations

	// Perform the actual request using the underlying implementation
	respBody, err := c.doRequest(ctx, method, path, body)
	if err != nil {
		return err
	}

	// If a target object was provided, unmarshal the response into it
	if target != nil {
		if err := json.Unmarshal(respBody, target); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}
	return nil
}

// checkRateLimit checks if we're rate limited and waits if needed
func (c *Client) checkRateLimit() error {
	// WHO: RateLimitManager
	// WHAT: Check rate limits
	// WHEN: Before API requests
	// WHERE: System Layer 6 (Integration)
	// WHY: To respect API limits
	// HOW: Using rate limiting
	// EXTENT: Rate limit compliance

	c.rateLimitMutex.RLock()
	defer c.rateLimitMutex.RUnlock()

	// Check if we're in rate limit and need to wait
	if c.rateLimitRemain <= c.options.RateLimitBuffer && time.Now().Before(c.rateLimitReset) {
		waitTime := time.Until(c.rateLimitReset)
		if waitTime > 0 {
			if c.options.Logger != nil {
				c.options.Logger.Info("Rate limit exceeded, waiting",
					"remaining", c.rateLimitRemain,
					"resetIn", waitTime.String())
			}

			time.Sleep(waitTime)
		}
	}

	return nil
}

// updateRateLimit updates rate limit information from response
func (c *Client) updateRateLimit(resp *http.Response) {
	// WHO: RateLimitTracker
	// WHAT: Update rate limit info
	// WHEN: After API requests
	// WHERE: System Layer 6 (Integration)
	// WHY: To track API limits
	// HOW: Using response headers
	// EXTENT: Rate limit awareness

	c.rateLimitMutex.Lock()
	defer c.rateLimitMutex.Unlock()

	// Get rate limit headers
	if remain := resp.Header.Get("X-RateLimit-Remaining"); remain != "" {
		if val, err := strconv.Atoi(remain); err == nil {
			c.rateLimitRemain = val
		}
	}

	if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
		if val, err := strconv.ParseInt(reset, 10, 64); err == nil {
			c.rateLimitReset = time.Unix(val, 0)
		}
	}
}

// cacheItem adds an item to the cache
func (c *Client) cacheItem(key string, value interface{}) {
	// WHO: CacheWriter
	// WHAT: Store cache item
	// WHEN: During caching operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To cache API results
	// HOW: Using memory cache
	// EXTENT: Single cache item

	if c.cacheDisabled || !c.options.EnableCache {
		return
	}

	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	now := time.Now()
	c.cache[key] = &cacheItem{
		data:      value,
		timestamp: now,
		expires:   now.Add(c.options.CacheTimeout),
	}
}

// getCachedItem retrieves an item from cache
func (c *Client) getCachedItem(key string) (interface{}, bool) {
	// WHO: CacheReader
	// WHAT: Retrieve cache item
	// WHEN: During API operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To use cached data
	// HOW: Using memory lookup
	// EXTENT: Single cache item

	if c.cacheDisabled || !c.options.EnableCache {
		return nil, false
	}

	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()

	item, found := c.cache[key]
	if !found {
		return nil, false
	}

	// Check if expired
	if time.Now().After(item.expires) {
		return nil, false
	}

	return item.data, true
}

// clearCache clears the entire cache
func (c *Client) clearCache() {
	// WHO: CacheCleaner
	// WHAT: Clear cache
	// WHEN: During cache management
	// WHERE: System Layer 6 (Integration)
	// WHY: To refresh cached data
	// HOW: Using cache invalidation
	// EXTENT: All cache items

	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	c.cache = make(map[string]*cacheItem)
}

// ParseResourceURI parses a GitHub resource URI
func ParseResourceURI(uri string) (map[string]string, error) {
	// WHO: URIParser
	// WHAT: Parse resource URI
	// WHEN: During resource operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To extract resource details
	// HOW: Using URI parsing
	// EXTENT: URI components

	// Supported URI formats:
	// - repo://{owner}/{repo}/contents{/path*}
	// - repo://{owner}/{repo}/refs/heads/{branch}/contents{/path*}
	// - repo://{owner}/{repo}/sha/{sha}/contents{/path*}
	// - issue://{owner}/{repo}/issues/{number}
	// - pr://{owner}/{repo}/pulls/{number}

	// Check if it's a valid URI
	if !strings.Contains(uri, "://") {
		return nil, fmt.Errorf("invalid URI format: %s", uri)
	}

	// Split scheme and path
	parts := strings.SplitN(uri, "://", 2)
	scheme := parts[0]
	path := parts[1]

	// Parse based on scheme
	switch scheme {
	case "repo":
		return parseRepoURI(path)
	case "issue":
		return parseIssueURI(path)
	case "pr":
		return parsePRURI(path)
	default:
		return nil, fmt.Errorf("unsupported URI scheme: %s", scheme)
	}
}

// parseRepoURI parses a repository URI
func parseRepoURI(path string) (map[string]string, error) {
	// WHO: RepoURIParser
	// WHAT: Parse repo URI
	// WHEN: During URI parsing
	// WHERE: System Layer 6 (Integration)
	// WHY: To extract repo details
	// HOW: Using path patterns
	// EXTENT: Repo URI components

	components := strings.Split(path, "/")
	if len(components) < 2 {
		return nil, fmt.Errorf("invalid repo URI format")
	}

	// Initialize with owner and repo
	result := map[string]string{
		"owner": components[0],
		"repo":  components[1],
	}

	// If we only have owner/repo, we're done
	if len(components) <= 2 {
		return result, nil
	}

	// Process path type
	processRepoPathType(components, result)
	return result, nil
}

// processRepoPathType handles different repository path types
func processRepoPathType(components []string, result map[string]string) {
	pathType := components[2]

	switch pathType {
	case "contents":
		processContentsPath(components, result)
	case "refs":
		processRefsPath(components, result)
	case "sha":
		processShaPath(components, result)
	}
}

// processContentsPath handles contents path format
func processContentsPath(components []string, result map[string]string) {
	result["type"] = "contents"
	if len(components) > 3 {
		result["path"] = strings.Join(components[3:], "/")
	}
}

// processRefsPath handles refs path format
func processRefsPath(components []string, result map[string]string) {
	if len(components) < 5 || components[3] != "heads" {
		return
	}

	result["type"] = "branch"
	result["branch"] = components[4]

	if len(components) > 6 && components[5] == "contents" {
		result["path"] = strings.Join(components[6:], "/")
	}
}

// processShaPath handles sha path format
func processShaPath(components []string, result map[string]string) {
	if len(components) < 4 {
		return
	}

	result["type"] = "commit"
	result["sha"] = components[3]

	if len(components) > 5 && components[4] == "contents" {
		result["path"] = strings.Join(components[5:], "/")
	}
}

// parseIssueURI parses an issue URI
func parseIssueURI(path string) (map[string]string, error) {
	// WHO: IssueURIParser
	// WHAT: Parse issue URI
	// WHEN: During URI parsing
	// WHERE: System Layer 6 (Integration)
	// WHY: To extract issue details
	// HOW: Using path patterns
	// EXTENT: Issue URI components

	components := strings.Split(path, "/")
	result := make(map[string]string)

	if len(components) < 4 || components[2] != "issues" {
		return nil, fmt.Errorf("invalid issue URI format")
	}

	// Extract components
	result["owner"] = components[0]
	result["repo"] = components[1]
	result["type"] = "issue"
	result["number"] = components[3]

	return result, nil
}

// parsePRURI parses a PR URI
func parsePRURI(path string) (map[string]string, error) {
	// WHO: PRURIParser
	// WHAT: Parse PR URI
	// WHEN: During URI parsing
	// WHERE: System Layer 6 (Integration)
	// WHY: To extract PR details
	// HOW: Using path patterns
	// EXTENT: PR URI components

	components := strings.Split(path, "/")
	result := make(map[string]string)

	if len(components) < 4 || components[2] != "pulls" {
		return nil, fmt.Errorf("invalid PR URI format")
	}

	// Extract components
	result["owner"] = components[0]
	result["repo"] = components[1]
	result["type"] = "pr"
	result["number"] = components[3]

	return result, nil
}

// decodeBase64Content decodes base64-encoded content
func decodeBase64Content(content string) ([]byte, error) {
	// WHO: ContentDecoder
	// WHAT: Decode content
	// WHEN: During content operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To decode base64 content
	// HOW: Using base64 decoding
	// EXTENT: Content decoding

	// Remove newlines
	content = strings.ReplaceAll(content, "\n", "")

	// Decode from base64
	decoded, err := base64DecodeString(content)
	if err != nil {
		return nil, fmt.Errorf("base64 decode error: %w", err)
	}

	return decoded, nil
}

// base64DecodeString decodes a base64 string
func base64DecodeString(s string) ([]byte, error) {
	// WHO: Base64Decoder
	// WHAT: Decode string
	// WHEN: During decoding operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To convert base64 to bytes
	// HOW: Using Go standard library
	// EXTENT: String decoding

	return base64.StdEncoding.DecodeString(s)
}

// loggerAdapter adapts a log.Logger to the Logger interface
type loggerAdapter struct {
	logger *log.Logger
}

func (a *loggerAdapter) Debug(msg string, args ...interface{}) {
	if a.logger != nil {
		a.logger.Debug(msg, args...)
	}
}

func (a *loggerAdapter) Info(msg string, args ...interface{}) {
	if a.logger != nil {
		a.logger.Info(msg, args...)
	}
}

func (a *loggerAdapter) Error(msg string, args ...interface{}) {
	if a.logger != nil {
		a.logger.Error(msg, args...)
	}
}

// NewClient creates a new GitHub API client with the given token and logger.
// This is a simplified wrapper around NewAdvancedClient.
func NewClient(token string, logger *log.Logger) *Client {
	var logAdapter Logger
	if logger != nil {
		logAdapter = &loggerAdapter{logger: logger}
	}

	options := ClientOptions{
		APIBaseURL:     DefaultAPIBaseURL,
		GraphQLBaseURL: DefaultGraphQLBaseURL,
		AcceptHeader:   DefaultAcceptHeader,
		UserAgent:      DefaultUserAgent,
		Timeout:        DefaultTimeout,
		CacheTimeout:   DefaultCacheTimeout,
		Logger:         logAdapter,
		Token:          token,
	}

	client, err := NewAdvancedClient(options)
	if err != nil {
		panic(fmt.Sprintf("failed to create GitHub client: %v", err))
	}
	return client
}

// Updated parameter extraction to use shared utilities
func extractParams(request mcp.CallToolRequest) (string, string, error) {
	owner, err := common.RequiredParam[string](request, "owner")
	if err != nil {
		return "", "", err
	}
	repo, err := common.RequiredParam[string](request, "repo")
	if err != nil {
		return "", "", err
	}
	return owner, repo, nil
}

// Example usage in a function
func GetRepositoryDetails(request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	owner, repo, err := extractParams(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	// ...existing logic...
	return mcp.NewToolResultText(fmt.Sprintf("Owner: %s, Repo: %s", owner, repo)), nil
}
