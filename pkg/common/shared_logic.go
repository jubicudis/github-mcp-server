package common

import (
	"context"
	"github.com/jubicudis/github-mcp-server/pkg/log"
	"github.com/jubicudis/github-mcp-server/pkg/translations"
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