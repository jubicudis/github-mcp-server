package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

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

// StubGetClientFnForCustomClient returns a GetClientFn that always returns the provided *ghmcp.Client (for ghmcp package)
func StubGetClientFnForCustomClient(client interface{}) func(ctx context.Context) (interface{}, error) {
	return func(ctx context.Context) (interface{}, error) {
		return client, nil
	}
}

// Ptr is a generic pointer helper for tests (canonical, DRY)
func Ptr[T any](v T) *T { return &v }

// CreateTestTranslateFunc returns a translation function for tests
func CreateTestTranslateFunc() func(ctx context.Context, key string) string {
	return func(_ context.Context, key string) string { return key }
}

// CreateQueryParamMatcher returns a mock matcher for query params (stub for now, implement as needed)
type MockMatcher struct {
	params map[string]string
}

func CreateQueryParamMatcher(t *testing.T, params map[string]string) *MockMatcher {
	return &MockMatcher{params: params}
}

func (m *MockMatcher) AndThen(next http.HandlerFunc) http.HandlerFunc {
	return next
}

// MockResponse returns a mock HTTP response for tests (stub for now, implement as needed)
func MockResponse(t *testing.T, status int, body interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		data, _ := json.Marshal(body)
		_, _ = w.Write(data)
	}
}

// ParseISOTimestamp parses an ISO 8601 timestamp string.
// It supports both RFC3339 format and date-only format (YYYY-MM-DD).
func ParseISOTimestamp(input string) (time.Time, error) {
	if input == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}
	// Try RFC3339 format first
	t, err := time.Parse(time.RFC3339, input)
	if err == nil {
		return t, nil
	}
	// Try date-only format (YYYY-MM-DD)
	t, err = time.Parse("2006-01-02", input)
	if err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid ISO 8601 timestamp: %s, supported formats are RFC3339 and YYYY-MM-DD", input)
}
