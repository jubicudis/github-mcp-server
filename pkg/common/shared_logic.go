package common

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/translations"
)

// CheckTNOSConnection attempts to connect to the TNOS MCP server using a QHP handshake.
func CheckTNOSConnection(ctx context.Context) bool {
	logger := log.NewLogger()
	// Build the WebSocket URL using the canonical TNOS MCP port
	url := fmt.Sprintf("ws://localhost:%d/ws", DefaultTNOSMCPPort)

	// Prepare dialer with handshake timeout
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 5 * time.Second

	// QHP headers for handshake
	headers := make(map[string][]string)
	headers["X-QHP-Version"] = []string{MCPVersion30}
	headers["X-QHP-Source"] = []string{"github-mcp-server"}
	headers["X-QHP-Intent"] = []string{"TNOS MCP Connection Test"}

	conn, _, err := dialer.Dial(url, headers)
	if err != nil {
		logger.Error("TNOS MCP connection failed", "error", err)
		return false
	}
	// Close connection immediately after successful handshake
	conn.Close()
	logger.Info("TNOS MCP connection successful", "url", url)
	return true
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