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
	"time"

	"log"
)

// ContextVector7D is already defined in context_translator.go
// Using that definition for all client operations

// DefaultContextVector7D creates a default context vector with standard values
func DefaultContextVector7D() ContextVector7D {
	// WHO: ContextFactory
	// WHAT: Create default context vector
	// WHEN: During initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide standard context values
	// HOW: Using 7D framework defaults
	// EXTENT: All client operations requiring context

	return ContextVector7D{
		Who:    "System",
		What:   "DefaultOperation",
		When:   time.Now().Unix(), // Convert time.Time to Unix timestamp (int64)
		Where:  "GitHub_MCP_Server",
		Why:    "StandardProcessing",
		How:    "DefaultMethod",
		Extent: 1.0,
	}
}

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

// Legacy client uses the ContextVector7D defined in context_translator.go

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

// NewLegacyClient creates a new legacy client adapter
func NewLegacyClient(token string, logger *log.Logger) *LegacyClientAdapter {
	// WHO: ClientFactory
	// WHAT: Create legacy client
	// WHEN: During client initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To maintain compatibility
	// HOW: Using adapter pattern
	// EXTENT: Client lifecycle management

	return NewLegacyClientAdapter(token, logger)
}

// NewLegacyClientAdapter creates a new legacy client adapter with default context
func NewLegacyClientAdapter(token string, logger *log.Logger) *LegacyClientAdapter {
	// WHO: AdapterFactory
	// WHAT: Create legacy client adapter
	// WHEN: During client initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To wrap advanced client with legacy interface
	// HOW: Using adapter pattern
	// EXTENT: Client lifecycle management

	options := ClientOptions{
		Token:       token,
		EnableCache: true,
	}
	advancedClient, _ := NewAdvancedClient(options)

	return &LegacyClientAdapter{
		advancedClient: advancedClient,
		legacyContext:  DefaultContextVector7D(),
		logger:         logger,
	}
}
