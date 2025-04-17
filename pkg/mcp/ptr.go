// WHO: MCP.Ptr
// WHAT: Generic pointer creator helper function
// WHEN: During API interactions
// WHERE: GitHub MCP Server / Utility Functions
// WHY: To create pointers to values for GitHub API
// HOW: Using Go generics to provide type safety
// EXTENT: Used throughout the MCP server for GitHub API interactions

package mcp

// Ptr returns a pointer to the provided value.
// This is a utility function to help create pointers for use with GitHub API functions
// that require pointers to primitive types.
func Ptr[T any](v T) *T {
	return &v
}
