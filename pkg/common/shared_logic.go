package common

import (
	"context"
	"fmt"
	"github-mcp-server/pkg/log"
	"github-mcp-server/pkg/translations"

	"github.com/mark3labs/mcp-go/mcp"
)

// CheckTNOSConnection attempts to connect to the TNOS MCP server.
func CheckTNOSConnection(ctx context.Context) bool {
	logger := log.NewLogger()
	logger.Info("Attempting to connect to TNOS MCP Server...")
	logger.Warn("Simulated TNOS MCP connection check: FAILED (running in Standalone Mode)")
	return false
}

// GenerateContextMap creates a context map from a ContextVector7D.
func GenerateContextMap(context7d translations.ContextVector7D) map[string]interface{} {
	return map[string]interface{}{
		"who":    context7d.Who,
		"what":   context7d.What,
		"when":   context7d.When,
		"where":  context7d.Where,
		"why":    context7d.Why,
		"how":    context7d.How,
		"extent": context7d.Extent,
		"source": context7d.Source,
	}
}

// OptionalStringArrayParam extracts an optional string array parameter from the request.
func OptionalStringArrayParam(request mcp.CallToolRequest, key string) ([]string, error) {
	if value, ok := request.Params[key]; ok {
		if array, ok := value.([]string); ok {
			return array, nil
		}
		return nil, fmt.Errorf("parameter '%s' is not a string array", key)
	}
	return nil, nil
}

// OptionalIntParam extracts an optional integer parameter from the request.
func OptionalIntParam(request mcp.CallToolRequest, key string) (*int, error) {
	if value, ok := request.Params[key]; ok {
		if number, ok := value.(int); ok {
			return &number, nil
		}
		return nil, fmt.Errorf("parameter '%s' is not an integer", key)
	}
	return nil, nil
}

// OptionalIntParamWithDefault extracts an optional integer parameter with a default value.
func OptionalIntParamWithDefault(request mcp.CallToolRequest, key string, defaultValue int) (int, error) {
	if value, ok := request.Params[key]; ok {
		if number, ok := value.(int); ok {
			return number, nil
		}
		return 0, fmt.Errorf("parameter '%s' is not an integer", key)
	}
	return defaultValue, nil
}

// WithPagination adds pagination parameters to a request.
func WithPagination(request mcp.CallToolRequest) (page, perPage int, err error) {
	page, err = OptionalIntParamWithDefault(request, "page", 1)
	if err != nil {
		return 0, 0, err
	}
	perPage, err = OptionalIntParamWithDefault(request, "per_page", 30)
	if err != nil {
		return 0, 0, err
	}
	return page, perPage, nil
}