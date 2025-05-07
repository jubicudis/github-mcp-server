/*
 * WHO: LegacyClient
 * WHAT: Legacy GitHub client interface for backward compatibility
 * WHEN: During GitHub API operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide backward compatibility with advanced client
 * HOW: Using adapter pattern with context awareness
 * EXTENT: All GitHub API interactions requiring legacy support
 */

package github

import (
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log"
)

// This file serves as a compatibility layer.
// All actual functionality has been migrated to github_client.go and client_adapter.go.
// The client_adapter.go file provides a LegacyClientAdapter that allows older code
// to continue using the simpler client interface while leveraging the advanced
// implementation under the hood.

/*
 * Migration Guide:
 *
 * 1. For new code, directly use the advanced Client from github_client.go:
 *
 *    options := github.ClientOptions{
 *        Token:      "your_token",
 *        EnableCache: true,
 *    }
 *    client, err := github.NewAdvancedClient(options)
 *
 * 2. For existing code, you can continue to use:
 *
 *    client := github.NewClient("your_token", logger)
 *
 *    This will now return a LegacyClientAdapter that provides the same interface
 *    but uses the advanced client implementation internally.
 */

// ContextVector7D represents the 7D context used by the legacy client
type ContextVector7D struct {
	// WHO: ContextManager
	// WHAT: 7D Context Structure
	// WHEN: During contextual operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide contextual awareness
	// HOW: Using structured context data
	// EXTENT: All context-aware operations

	Who    string                 `json:"who"`
	What   string                 `json:"what"`
	When   int64                  `json:"when"`
	Where  string                 `json:"where"`
	Why    string                 `json:"why"`
	How    string                 `json:"how"`
	Extent float64                `json:"extent"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
}

// LegacyClientAdapter adapts the advanced client to the legacy interface
type LegacyClientAdapter struct {
	// WHO: AdapterManager
	// WHAT: Client Adapter Structure
	// WHEN: During client operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To bridge legacy and modern code
	// HOW: Using adapter pattern
	// EXTENT: All legacy client users

	advancedClient *Client
	legacyContext  ContextVector7D
	logger         *log.Logger
}

// NewClient creates a new legacy client adapter
func NewClient(token string, logger *log.Logger) *LegacyClientAdapter {
	// WHO: ClientFactory
	// WHAT: Create legacy client
	// WHEN: During client initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To maintain compatibility
	// HOW: Using adapter pattern
	// EXTENT: Client lifecycle management

	return NewLegacyClientAdapter(token, logger)
}
