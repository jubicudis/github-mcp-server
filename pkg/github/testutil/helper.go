/*
 * WHO: TestHelpers
 * WHAT: Helper utilities for GitHub MCP server tests
 * WHEN: During test execution
 * WHERE: System Layer 6 (Integration) Tests
 * WHY: To simplify common test operations
 * HOW: Using utility functions
 * EXTENT: All test cases
 */

package testutil

// NullTranslationHelperFunc returns a null translation helper function
// This bridges between the TranslationHelper interface and TranslationHelperFunc type
func NullTranslationHelperFunc(key string, defaultValue string) string {
	return defaultValue
}
