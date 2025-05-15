/*
 * WHO: GitHubClientAdapter
 * WHAT: Adapter for GitHub client
 * WHEN: During GitHub API operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide simplified client API
 * HOW: Using adapter pattern
 * EXTENT: Main server operations
 */
package main

import (
	"context"
	"encoding/json"
	"fmt"

	pkggithub "github.com/tranquility-neuro-os/github-mcp-server/pkg/github"
	"github.com/tranquility-neuro-os/github-mcp-server/pkg/log"

	githubapi "github.com/google/go-github/v49/github" // Import the go-github client
	"github.com/mark3labs/mcp-go/mcp"
)

// GitHubService defines the interface used by the main server
type GitHubService interface {
	// Repository operations
	GetRepository(owner, repo string) (interface{}, error)
	GetIssues(owner, repo string) (interface{}, error)
	GetPullRequests(owner, repo string) (interface{}, error)
	SearchCode(query string) (interface{}, error)
	GetCodeScanningAlerts(owner, repo string) (interface{}, error)
}

// GitHubClientAdapter adapts the GitHub client to the server's needs
type GitHubClientAdapter struct {
	client    *pkggithub.Client
	logger    *log.Logger
	getClient pkggithub.GetClientFn
}

// NewClient creates a new GitHub client adapter
func NewClient(token string, logger *log.Logger) *GitHubClientAdapter {
	// Create the standard client
	client := pkggithub.NewClient(token, logger)

	// Create a GetClientFn that returns a go-github client
	getClient := func(ctx context.Context) (*githubapi.Client, error) {
		// In a real implementation, this would return the go-github client
		// For now, we'll return nil as we're using our own client interfaces
		return nil, nil
	}

	return &GitHubClientAdapter{
		client:    client,
		logger:    logger,
		getClient: getClient,
	}
}

// extractTextContent extracts text from MCP content
func extractTextContent(content []mcp.Content) (string, error) {
	if len(content) == 0 {
		return "", fmt.Errorf("no content returned")
	}

	// Get text content from the first element
	textContent, ok := content[0].(mcp.TextContent)
	if !ok {
		return "", fmt.Errorf("content is not text")
	}

	return textContent.Text, nil
}

// GetRepository retrieves a GitHub repository
func (a *GitHubClientAdapter) GetRepository(owner, repo string) (interface{}, error) {
	a.logger.Debug("Getting repository", "owner", owner, "repo", repo)

	ctx := context.Background()
	// Create a simple request object with the required parameters
	request := createRequest(map[string]interface{}{
		"owner": owner,
		"repo":  repo,
	})

	// Create simple translation helper function
	translationHelper := func(key string, defaultValue string) string {
		return defaultValue
	}

	// Use the repository tool handler
	_, handler := pkggithub.GetRepository(a.getClient, translationHelper)
	result, err := handler(ctx, *request)
	if err != nil {
		return nil, err
	}

	// Extract and parse the result content
	text, err := extractTextContent(result.Content)
	if err != nil {
		return nil, err
	}

	var repoData interface{}
	err = json.Unmarshal([]byte(text), &repoData)
	return repoData, err
}

// GetIssues retrieves issues from a GitHub repository
func (a *GitHubClientAdapter) GetIssues(owner, repo string) (interface{}, error) {
	a.logger.Debug("Getting issues", "owner", owner, "repo", repo)

	ctx := context.Background()
	// Create a simple request object with the required parameters
	request := createRequest(map[string]interface{}{
		"owner": owner,
		"repo":  repo,
	})

	// Create simple translation helper function
	translationHelper := func(key string, defaultValue string) string {
		return defaultValue
	}

	// Use the issues tool handler
	_, handler := pkggithub.GetIssues(a.getClient, translationHelper)
	result, err := handler(ctx, *request)
	if err != nil {
		return nil, err
	}

	// Extract and parse the result content
	text, err := extractTextContent(result.Content)
	if err != nil {
		return nil, err
	}

	var issuesData interface{}
	err = json.Unmarshal([]byte(text), &issuesData)
	return issuesData, err
}

// GetPullRequests retrieves pull requests from a GitHub repository
func (a *GitHubClientAdapter) GetPullRequests(owner, repo string) (interface{}, error) {
	a.logger.Debug("Getting pull requests", "owner", owner, "repo", repo)

	ctx := context.Background()
	// Create a simple request object with the required parameters
	request := createRequest(map[string]interface{}{
		"owner": owner,
		"repo":  repo,
		"state": "all", // Get all PRs by default
	})

	// Create simple translation helper function
	translationHelper := func(key string, defaultValue string) string {
		return defaultValue
	}

	// Use the pull requests list tool handler
	_, handler := pkggithub.ListPullRequests(a.getClient, translationHelper)
	result, err := handler(ctx, *request)
	if err != nil {
		return nil, err
	}

	// Extract and parse the result content
	text, err := extractTextContent(result.Content)
	if err != nil {
		return nil, err
	}

	var prsData interface{}
	err = json.Unmarshal([]byte(text), &prsData)
	return prsData, err
}

// SearchCode searches for code in GitHub repositories
func (a *GitHubClientAdapter) SearchCode(query string) (interface{}, error) {
	a.logger.Debug("Searching code", "query", query)

	ctx := context.Background()
	// Create a simple request object with the required parameters
	request := createRequest(map[string]interface{}{
		"q": query,
	})

	// Create simple translation helper function
	translationHelper := func(key string, defaultValue string) string {
		return defaultValue
	}

	// Use the search code tool handler
	_, handler := pkggithub.SearchCode(a.getClient, translationHelper)
	result, err := handler(ctx, *request)
	if err != nil {
		return nil, err
	}

	// Extract and parse the result content
	text, err := extractTextContent(result.Content)
	if err != nil {
		return nil, err
	}

	var searchData interface{}
	err = json.Unmarshal([]byte(text), &searchData)
	return searchData, err
}

// GetCodeScanningAlerts retrieves code scanning alerts from a GitHub repository
func (a *GitHubClientAdapter) GetCodeScanningAlerts(owner, repo string) (interface{}, error) {
	a.logger.Debug("Getting code scanning alerts", "owner", owner, "repo", repo)

	ctx := context.Background()
	// Create a simple request object with the required parameters
	request := createRequest(map[string]interface{}{
		"owner": owner,
		"repo":  repo,
	})

	// Create simple translation helper function
	translationHelper := func(key string, defaultValue string) string {
		return defaultValue
	}

	// Use the code scanning alerts tool handler
	_, handler := pkggithub.ListCodeScanningAlerts(a.getClient, translationHelper)
	result, err := handler(ctx, *request)
	if err != nil {
		return nil, err
	}

	// Extract and parse the result content
	text, err := extractTextContent(result.Content)
	if err != nil {
		return nil, err
	}

	var alertsData interface{}
	err = json.Unmarshal([]byte(text), &alertsData)
	return alertsData, err
}

// createRequest creates an MCP request with the provided arguments
func createRequest(args map[string]interface{}) *mcp.CallToolRequest {
	req := &mcp.CallToolRequest{}
	req.Params.Arguments = args
	return req
}
