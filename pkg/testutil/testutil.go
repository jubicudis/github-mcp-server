package testutil

import (
	"context"
	"testing"

	"github.com/google/go-github/v71/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// CreateMCPRequest creates a canonical MCP CallToolRequest from arguments.
func CreateMCPRequest(args map[string]interface{}) mcp.CallToolRequest {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	return req
}

// GetTextResult extracts the text result from a CallToolResult, failing the test if not present.
func GetTextResult(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	require.NotNil(t, result)
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			return textContent.Text
		}
	}
	return ""
}

// NullTranslationHelperFunc is a no-op translation function for tests.
var NullTranslationHelperFunc = func(key, defaultValue string) string { return defaultValue }

// StubGetClientFnWithClient returns a GetClientFn that always returns the provided *github.Client.
func StubGetClientFnWithClient(client *github.Client) func(ctx context.Context) (*github.Client, error) {
	return func(ctx context.Context) (*github.Client, error) {
		return client, nil
	}
}
