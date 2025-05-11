package testutil

import (
	"context"
	"testing"

	"github.com/google/go-github/v49/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// StubGetClientFn returns a function that returns the provided client, useful for testing
func StubGetClientFn(client *github.Client) func(context.Context) (*github.Client, error) {
	return func(context.Context) (*github.Client, error) {
		return client, nil
	}
}

// CreateMCPRequest creates an MCP request with the provided arguments
func CreateMCPRequest(args map[string]interface{}) *mcp.CallToolRequest {
	return &mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

// GetTextResult extracts text content from an MCP result
func GetTextResult(t *testing.T, result *mcp.CallToolResult) string {
	require.NotNil(t, result)

	// In the updated API, text results use the Value field
	text, ok := result.Value.(string)
	require.True(t, ok, "Result value is not a string")
	return text
}

// Ptr returns a pointer to the provided value
func Ptr[T any](v T) *T {
	return &v
}
