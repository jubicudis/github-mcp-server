/*
 * WHO: GitHub MCP Client Adapter (Factory)
 * WHAT: Thin factory for canonical GitHub MCP compatibility adapter
 * WHEN: During server/client initialization
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide a single, canonical entry point for GitHub MCP client usage
 * HOW: Delegates to pkg/github/client_adapter.go (ClientCompatibilityAdapter)
 * EXTENT: Adapter instantiation only; no logic or handler implementation here
 * NOTE: All parameter extraction, request construction, and handler invocation must use canonical helpers from pkg/common and pkg/github. This file is fully aligned with the 7D documentation and MCP architecture vision. All logic is in pkg/github/client_adapter.go.
 */
package main

import (
	"github-mcp-server/pkg/common"
	"github-mcp-server/pkg/github"
)

// NewClient returns the canonical GitHub MCP compatibility adapter.
// All logic is implemented in pkg/github/client_adapter.go (ClientCompatibilityAdapter).
// This function is the only entry point for client instantiation in the server layer.
func NewClient(opts common.ConnectionOptions) *github.ClientCompatibilityAdapter {
	return github.NewClientCompatibilityAdapter(opts)
}

// TODO: If additional adapter methods are needed, add thin wrappers here that delegate to the canonical adapter.
//       Do not implement any logic or parameter extraction in this file.
//       All test and integration coverage should use the canonical adapter and helpers.
