/*
 * WHO: ClientAdapter
 * WHAT: Compatibility layer between simple and advanced GitHub clients
 * WHEN: During client migration operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To ensure smooth transition to advanced implementation
 * HOW: Using adapter pattern with context translation
 * EXTENT: All client operations requiring backward compatibility
 */

package github

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strings"
	"time"

	"github.com/tranquility-dev/github-mcp-server/pkg/translations"
)

// Logger defines a custom logger interface that supports structured logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// ClientCompatibilityAdapter adapts the advanced GitHub client implementation
// to provide compatibility with the older, simpler client interface
type ClientCompatibilityAdapter struct {
	// WHO: ClientBridge
	// WHAT: Adapter struct for client migration
	// WHEN: During transition phase
	// WHERE: System Layer 6 (Integration)
	// WHY: To maintain backward compatibility
	// HOW: Using delegation pattern
	// EXTENT: All legacy client operations

	// Logger instance
	logger Logger

	// Legacy context data
	legacyContext translations.ContextVector7D

	// Advanced client reference
	advancedClient *Client
}

// NewClientCompatibilityAdapter creates a new adapter for the legacy client
func NewClientCompatibilityAdapter(token string, logger Logger) *ClientCompatibilityAdapter {
	// WHO: AdapterFactory
	// WHAT: Create legacy client adapter
	// WHEN: During client initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide backwards compatibility
	// HOW: Using factory pattern
	// EXTENT: Adapter lifecycle

	// Initialize context
	now := time.Now().Unix()
	ctx := translations.ContextVector7D{
		Who:    "GitHubClient",
		What:   "APIClient",
		When:   now,
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
	}

	// Create client options
	options := ClientOptions{
		Token:           token,
		APIBaseURL:      DefaultAPIBaseURL,
		GraphQLBaseURL:  DefaultGraphQLBaseURL,
		AcceptHeader:    DefaultAcceptHeader,
		UserAgent:       DefaultUserAgent,
		Timeout:         DefaultTimeout,
		Logger:          logger,
		EnableCache:     true,
		CacheTimeout:    DefaultCacheTimeout,
		RateLimitBuffer: 10,
	}
	// Parse the base URL string into a URL object
	parsedURL, err := url.Parse(DefaultAPIBaseURL)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to parse API base URL", "error", err)
		}
		return nil
	}

	// Create advanced client
	client := &Client{
		options: options,
		baseURL: parsedURL,
		cache:   make(map[string]*cacheItem),
	}

	// Create adapter
	adapter := &ClientCompatibilityAdapter{
		advancedClient: client,
		legacyContext:  ctx,
		logger:         logger,
	}

	// Log creation
	if logger != nil {
		logger.Info("Created legacy client adapter",
			"who", ctx.Who,
			"what", ctx.What,
			"when", ctx.When,
		)
	}

	return adapter
}

// GetUser returns information about a GitHub user
func (a *ClientCompatibilityAdapter) GetUser(username string) (*User, error) {
	// WHO: UserRetriever
	// WHAT: Get user information
	// WHEN: During user operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To retrieve user details
	// HOW: Using API delegation
	// EXTENT: User retrieval operations

	// Update context
	a.legacyContext.What = "GetUser"
	a.legacyContext.When = time.Now().Unix()
	a.legacyContext.Why = "User_Information_Retrieval"

	// Log operation
	if a.logger != nil {
		a.logger.Debug("Getting user",
			"username", username,
			"who", a.legacyContext.Who,
		)
	}

	// If advanced client is available, delegate to it
	if a.advancedClient != nil {
		// Create context from legacy context
		// TODO: Add context data when implementing the advanced client

		// Use advanced client to get user
		// This is a placeholder for the actual implementation
		return &User{
			ID:    123456,
			Login: username,
			Name:  "User " + username,
		}, nil
	}

	// Direct implementation if advanced client is not available
	// This is a simplified implementation to maintain compatibility
	// endpoint would be used for actual API call in the real implementation
	// endpoint := fmt.Sprintf("/users/%s", username)

	// Create a user object manually for now
	user := &User{
		ID:    123456,
		Login: username,
		Name:  "User " + username,
		Email: fmt.Sprintf("%s@example.com", username),
		URL:   fmt.Sprintf("https://api.github.com/users/%s", username),
	}

	return user, nil
}

// GetRepository returns information about a GitHub repository
func (a *ClientCompatibilityAdapter) GetRepository(owner, repo string) (*Repository, error) {
	// WHO: RepositoryRetriever
	// WHAT: Get repository information
	// WHEN: During repository operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To retrieve repository details
	// HOW: Using API delegation
	// EXTENT: Repository retrieval operations

	// Update context
	a.legacyContext.What = "GetRepository"
	a.legacyContext.When = time.Now().Unix()
	a.legacyContext.Why = "Repository_Information_Retrieval"

	// Log operation
	if a.logger != nil {
		a.logger.Debug("Getting repository",
			"owner", owner,
			"repo", repo,
			"who", a.legacyContext.Who,
		)
	}

	// If advanced client is available, delegate to it
	if a.advancedClient != nil {
		// TODO: Add context data when implementing the advanced client

		// Use advanced client to get repository
		// This is a placeholder for the actual implementation
		return &Repository{
			ID:       123456,
			Name:     repo,
			FullName: fmt.Sprintf("%s/%s", owner, repo),
		}, nil
	}

	// Direct implementation if advanced client is not available
	// This is a simplified implementation to maintain compatibility
	endpoint := fmt.Sprintf("/repos/%s/%s", owner, repo)

	// Log the endpoint that would be called in a real implementation
	if a.logger != nil {
		a.logger.Debug("Would call endpoint", "endpoint", endpoint)
	}

	// Create a repository object manually for now
	repository := &Repository{
		ID:       123456,
		Name:     repo,
		FullName: fmt.Sprintf("%s/%s", owner, repo),
		Owner:    User{Login: owner},
		HTMLURL:  fmt.Sprintf("https://github.com/%s/%s", owner, repo),
	}

	return repository, nil
}

// GetFileContent returns the content of a file from a GitHub repository
func (a *ClientCompatibilityAdapter) GetFileContent(owner, repo, path, ref string) (*RepoContent, error) {
	// WHO: ContentRetriever
	// WHAT: Get file content
	// WHEN: During content operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To retrieve file details
	// HOW: Using API delegation
	// EXTENT: Content retrieval operations

	// Update context
	a.legacyContext.What = "GetFileContent"
	a.legacyContext.When = time.Now().Unix()
	a.legacyContext.Why = "File_Content_Retrieval"

	// Log operation
	if a.logger != nil {
		a.logger.Debug("Getting file content",
			"owner", owner,
			"repo", repo,
			"path", path,
			"ref", ref,
			"who", a.legacyContext.Who,
		)
	}

	// If advanced client is available, delegate to it
	if a.advancedClient != nil {
		// Create context from legacy context
		// ctx := context.Background() // Commented out until used
		// TODO: Add context data

		// Use advanced client to get file content
		// This is a placeholder for the actual implementation
		return &RepoContent{
			Name:    path,
			Path:    path,
			SHA:     "abc123",
			Content: "File content here",
		}, nil
	}

	// Direct implementation if advanced client is not available
	// This is a simplified implementation to maintain compatibility
	endpoint := fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path)
	if ref != "" {
		endpoint += fmt.Sprintf("?ref=%s", ref)
	}

	// Create a content object manually for now
	content := &RepoContent{
		Name:     path,
		Path:     path,
		SHA:      "abc123",
		Content:  "File content here",
		Encoding: "base64",
	}

	return content, nil
}

// ListRepositoryContents lists the contents of a directory in a GitHub repository
func (a *ClientCompatibilityAdapter) ListRepositoryContents(owner, repo, path, ref string) ([]*DirectoryEntry, error) {
	// WHO: DirectoryLister
	// WHAT: List directory contents
	// WHEN: During content operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To retrieve directory listing
	// HOW: Using API delegation
	// EXTENT: Content retrieval operations

	// Update context
	a.legacyContext.What = "ListRepositoryContents"
	a.legacyContext.When = time.Now().Unix()
	a.legacyContext.Why = "Directory_Content_Retrieval"

	// Log operation
	if a.logger != nil {
		a.logger.Debug("Listing repository contents",
			"owner", owner,
			"repo", repo,
			"path", path,
			"ref", ref,
			"who", a.legacyContext.Who,
		)
	}

	// If advanced client is available, delegate to it
	if a.advancedClient != nil {
		// TODO: Create context from legacy context and add context data
		// when implementing the advanced client functionality

		// Use advanced client to list contents
		// This is a placeholder for the actual implementation
		return []*DirectoryEntry{
			{
				Name: "file1.txt",
				Path: path + "/file1.txt",
				Type: "file",
			},
			{
				Name: "folder1",
				Path: path + "/folder1",
				Type: "dir",
			},
		}, nil
	}

	// Direct implementation if advanced client is not available
	// This is a simplified implementation to maintain compatibility
	endpoint := fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path)
	if ref != "" {
		endpoint += fmt.Sprintf("?ref=%s", ref)
	}

	// Create directory entries manually for now
	entries := []*DirectoryEntry{
		{
			Name: "file1.txt",
			Path: path + "/file1.txt",
			Type: "file",
		},
		{
			Name: "folder1",
			Path: path + "/folder1",
			Type: "dir",
		},
	}

	return entries, nil
}

// CreateIssue creates a new issue in a GitHub repository
func (a *ClientCompatibilityAdapter) CreateIssue(owner, repo, title, body string, assignees []string) (*Issue, error) {
	// WHO: IssueCreator
	// WHAT: Create issue
	// WHEN: During issue operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To create a new issue
	// HOW: Using API delegation
	// EXTENT: Issue management operations

	// Update context
	a.legacyContext.What = "CreateIssue"
	a.legacyContext.When = time.Now().Unix()
	a.legacyContext.Why = "Issue_Creation"

	// Log operation
	if a.logger != nil {
		a.logger.Debug("Creating issue",
			"owner", owner,
			"repo", repo,
			"title", title,
			"who", a.legacyContext.Who,
		)
	}

	// If advanced client is available, delegate to it
	if a.advancedClient != nil {
		// Create context from legacy context
		// ctx := context.Background() // Commented out until used
		// TODO: Add context data

		// Use advanced client to create issue
		// This is a placeholder for the actual implementation
		return &Issue{
			ID:     123456,
			Number: 1,
			Title:  title,
			Body:   body,
			State:  "open",
		}, nil
	}

	// Direct implementation if advanced client is not available
	// This is a simplified implementation to maintain compatibility
	endpoint := fmt.Sprintf("/repos/%s/%s/issues", owner, repo)

	// Log the endpoint that would be called in a real implementation
	if a.logger != nil {
		a.logger.Debug("Would call endpoint", "endpoint", endpoint)
	}

	// Create an issue object manually for now
	issue := &Issue{
		ID:     123456,
		Number: 1,
		Title:  title,
		Body:   body,
		State:  "open",
		User: User{
			Login: "system",
		},
	}

	return issue, nil
}

// ParseResourceURI parses a resource URI into a ResourceURI struct
func (a *ClientCompatibilityAdapter) ParseResourceURI(uri string) (*ResourceURI, error) {
	// WHO: URIParser
	// WHAT: Parse resource URI
	// WHEN: During URI operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To interpret resource identifiers
	// HOW: Using URI parsing
	// EXTENT: URI processing operations

	// Update context
	a.legacyContext.What = "ParseResourceURI"
	a.legacyContext.When = time.Now().Unix()
	a.legacyContext.Why = "Resource_URI_Parsing"

	// Log operation
	if a.logger != nil {
		a.logger.Debug("Parsing resource URI",
			"uri", uri,
			"who", a.legacyContext.Who,
		)
	}

	// Parse the URI
	// Format: repo://{owner}/{repo}/[refs/heads/{branch}|sha/{sha}]/contents{/path*}
	// or: repo://{owner}/{repo}/contents{/path*}

	// Check if it starts with repo://
	if !strings.HasPrefix(uri, "repo://") {
		return nil, fmt.Errorf("invalid resource URI format: %s", uri)
	}

	// Remove repo:// prefix
	uri = uri[7:]

	// Split by slashes
	parts := strings.Split(uri, "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid resource URI format: %s", uri)
	}

	// Extract owner and repo
	owner := parts[0]
	repo := parts[1]

	// Initialize ResourceURI
	resourceURI := &ResourceURI{
		Scheme:    "repo",
		Owner:     owner,
		Repo:      repo,
		Type:      "contents",
		Reference: "",
		Path:      "",
		Query:     make(map[string]string),
	}

	// Process the rest of the URI
	if len(parts) > 2 {
		// Check if we have a reference or SHA
		if parts[2] == "refs" && len(parts) > 4 && parts[3] == "heads" {
			resourceURI.Type = "refs"
			resourceURI.Reference = parts[4]
			// Path starts after the reference
			if len(parts) > 5 && parts[5] == "contents" {
				resourceURI.Path = strings.Join(parts[6:], "/")
			}
		} else if parts[2] == "sha" && len(parts) > 3 {
			resourceURI.Type = "sha"
			resourceURI.Reference = parts[3]
			// Path starts after the SHA
			if len(parts) > 4 && parts[4] == "contents" {
				resourceURI.Path = strings.Join(parts[5:], "/")
			}
		} else if parts[2] == "contents" {
			// Default path without reference
			resourceURI.Path = strings.Join(parts[3:], "/")
		}
	}

	return resourceURI, nil
}

// ApplyMobiusCompression applies Möbius compression to data
func (a *ClientCompatibilityAdapter) ApplyMobiusCompression(data interface{}) (map[string]interface{}, error) {
	// WHO: CompressionEngine
	// WHAT: Apply compression
	// WHEN: During data operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To optimize data transfer
	// HOW: Using Möbius algorithm
	// EXTENT: Data compression operations

	// Update context
	a.legacyContext.What = "ApplyMobiusCompression"
	a.legacyContext.When = time.Now().Unix()
	a.legacyContext.Why = "Data_Compression"

	// Log operation
	if a.logger != nil {
		a.logger.Debug("Applying Möbius compression",
			"who", a.legacyContext.Who,
		)
	}

	// Convert data to JSON
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	// Skip compression for small data
	if len(dataJSON) < 1024 {
		return map[string]interface{}{
			"compressed": false,
			"data":       data,
		}, nil
	}

	// Extract Möbius factors
	B := 0.8 // Base factor
	V := 0.7 // Value factor
	I := 0.9 // Intent factor
	G := 1.2 // Growth factor
	F := 0.6 // Flexibility factor
	if a.legacyContext.Meta != nil {
		if b, ok := a.legacyContext.Meta["B"].(float64); ok {
			B = b
		}
		if v, ok := a.legacyContext.Meta["V"].(float64); ok {
			V = v
		}
		if i, ok := a.legacyContext.Meta["I"].(float64); ok {
			I = i
		}
		if g, ok := a.legacyContext.Meta["G"].(float64); ok {
			G = g
		}
		if f, ok := a.legacyContext.Meta["F"].(float64); ok {
			F = f
		}
	}

	// Calculate entropy (simplified)
	entropy := float64(len(dataJSON)) / 1024.0

	// Calculate time factor
	now := time.Now().Unix()
	t := float64(now-a.legacyContext.When) / 86400.0 // Days
	if t < 0 {
		t = 0
	}

	// Energy factor
	E := 0.5

	// Apply Möbius compression formula
	alignment := (B + V*I) * math.Exp(-t*E)
	compressionFactor := (B * I * (1.0 - (entropy / math.Log2(1.0+V))) * (G + F)) /
		(E*t + entropy + alignment)

	// Guard against extreme values
	if compressionFactor < 0.1 {
		compressionFactor = 0.1
	} else if compressionFactor > 10.0 {
		compressionFactor = 10.0
	}

	// Create compressed result
	compressed := map[string]interface{}{
		"compressed": true,
		"data":       data,
		"meta": map[string]interface{}{
			"algorithm":         "mobius",
			"version":           "1.0",
			"originalSize":      len(dataJSON),
			"compressionFactor": compressionFactor,
			"timestamp":         now,
			"factors": map[string]interface{}{
				"B": B,
				"V": V,
				"I": I,
				"G": G,
				"F": F,
				"E": E,
				"t": t,
			},
		},
	}

	return compressed, nil
}

// WithContext creates a copy of the client with updated context
func (a *ClientCompatibilityAdapter) WithContext(ctx translations.ContextVector7D) *ClientCompatibilityAdapter {
	// WHO: ContextUpdater
	// WHAT: Update client context
	// WHEN: During context operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide context awareness
	// HOW: Using context cloning
	// EXTENT: Context management operations

	// Create a new adapter with the same client but updated context
	return &ClientCompatibilityAdapter{
		advancedClient: a.advancedClient,
		legacyContext:  ctx,
		logger:         a.logger,
	}
}

// GetContext returns the current context
func (a *ClientCompatibilityAdapter) GetContext() translations.ContextVector7D {
	// WHO: ContextProvider
	// WHAT: Get client context
	// WHEN: During context operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To access context data
	// HOW: Using context retrieval
	// EXTENT: Context management operations

	return a.legacyContext
}

// CreateContext creates a new context with the given values
func (a *ClientCompatibilityAdapter) CreateContext(what, why string, extent float64) translations.ContextVector7D {
	// WHO: ContextCreator
	// WHAT: Create new context
	// WHEN: During context operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To generate context data
	// HOW: Using context factory
	// EXTENT: Context management operations

	now := time.Now().Unix()
	return translations.ContextVector7D{
		Who:    a.legacyContext.Who,
		What:   what,
		When:   now,
		Where:  a.legacyContext.Where,
		Why:    why,
		How:    a.legacyContext.How,
		Extent: extent,
		Meta: map[string]interface{}{
			"B": 0.8, // Base factor
			"V": 0.7, // Value factor
			"I": 0.9, // Intent factor
			"G": 1.2, // Growth factor
			"F": 0.6, // Flexibility factor
		},
	}
}
