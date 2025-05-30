/*
 * WHO: MCPClientCreator
 * WHAT: Simple GitHub client creation
 * WHEN: During MCP server initialization
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide GitHub API access
 * HOW: Using a simplified client creation function
 * EXTENT: All GitHub API operations
 */

package github

import (
	"context"
	"github-mcp-server/pkg/log"
)

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

// GitHubService defines the interface expected by the MCP server for GitHub operations
// Add all methods used in main.go and handlers
// These are thin wrappers around the real client methods

type GitHubService interface {
	GetRepository(owner, repo string) (map[string]interface{}, error)
	GetIssues(owner, repo string) ([]map[string]interface{}, error)
	GetPullRequests(owner, repo string) ([]map[string]interface{}, error)
	SearchCode(query string) (map[string]interface{}, error)
	GetCodeScanningAlerts(owner, repo string) ([]map[string]interface{}, error)
}

// Ensure *Client implements GitHubService
func (c *Client) GetRepository(owner, repo string) (map[string]interface{}, error) {
	return c.GetRepositoryByName(context.Background(), owner, repo)
}

func (c *Client) GetIssues(owner, repo string) ([]map[string]interface{}, error) {
	// TODO: Implement using c.doRequest or delegate to a real method
	return nil, nil
}

func (c *Client) GetPullRequests(owner, repo string) ([]map[string]interface{}, error) {
	// TODO: Implement using c.doRequest or delegate to a real method
	return nil, nil
}

func (c *Client) SearchCode(query string) (map[string]interface{}, error) {
	return c.SearchCode(query)
}

func (c *Client) GetCodeScanningAlerts(owner, repo string) ([]map[string]interface{}, error) {
	return c.GetCodeScanningAlerts(owner, repo)
}

// NewClient creates a new GitHub API client with the given token and logger.
// This is a simplified wrapper around NewAdvancedClient.
//
// WHO: MCPClientFactory
// WHAT: Create GitHub API client
// WHEN: During server initialization
// WHERE: System Layer 6 (Integration)
// WHY: To provide GitHub API access
// HOW: Using a simplified interface with sensible defaults
// EXTENT: All GitHub API operations
func NewClient(token string, logger *log.Logger) GitHubService {
	var logAdapter Logger
	if logger != nil {
		logAdapter = &loggerAdapter{logger: logger}
	}

	options := ClientOptions{
		Token:          token,
		APIBaseURL:     DefaultAPIBaseURL,
		GraphQLBaseURL: DefaultGraphQLBaseURL,
		AcceptHeader:   DefaultAcceptHeader,
		UserAgent:      DefaultUserAgent,
		Timeout:        DefaultTimeout,
		CacheTimeout:   DefaultCacheTimeout,
		Logger:         logAdapter,
	}

	client, err := NewAdvancedClient(options)
	if err != nil {
		// Log the error but return a client anyway to avoid nil pointer issues
		if logger != nil {
			logger.Error("Failed to create GitHub client", "error", err.Error())
		}
		// Create a minimal client with just the options
		return &Client{
			options: options,
		}
	}

	return client
}
