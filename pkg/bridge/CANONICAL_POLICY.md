// Canonical Go MCP Bridge Implementation
// --------------------------------------
// This directory (`pkg/bridge`) contains the ONLY canonical Go implementation of the MCP bridge for the GitHub MCP Server.
//
// - All bridge logic, protocol, and shared types MUST reside here.
// - Do NOT duplicate bridge logic in `pkg/github` or any other subfolder.
// - If bridge functionality is needed elsewhere, always import from this package.
// - For GitHub-specific API logic, use `pkg/github` and call into the bridge as needed.
//
// This ensures a single source of truth for bridge operations and avoids redundancy, confusion, and maintenance issues.
//
// [AUTOMATED POLICY: Any detected duplication of bridge logic outside this directory will be flagged for removal.]
