/*
 * WHO: TranslationHelper
 * WHAT: 7D Context aware translation functions
 * WHEN: During MCP communication
 * WHERE: ATL translation layer
 * WHY: To convert between GitHub API and TNOS formats
 * HOW: Using standardized mapping functions
 * EXTENT: All external API communications
 */
package translations

import (
	"context"
)

// NullTranslationHelper is a no-op implementation for testing
// WHO: TranslationProvider
// WHAT: NullObject pattern implementation
// WHEN: During test execution
// WHERE: Testing environment
// WHY: To provide a non-functional placeholder
// HOW: Using nil implementations
// EXTENT: All test cases
var NullTranslationHelper TranslationHelper = &nullTranslationHelper{}

// nullTranslationHelper provides a no-op implementation
type nullTranslationHelper struct{}

// ToJSON implements the TranslationHelper interface
func (n *nullTranslationHelper) ToJSON(v interface{}) (string, error) {
	return ToJSON(v)
}

// FromJSON implements the TranslationHelper interface
func (n *nullTranslationHelper) FromJSON(data string, v interface{}) error {
	// Use helper function that handles the unmarshaling correctly
	_, err := FromJSON(data)
	return err
}

// ApplyContextVector implements the TranslationHelper interface
func (n *nullTranslationHelper) ApplyContextVector(v interface{}, context map[string]string) (interface{}, error) {
	// No-op implementation for testing
	return v, nil
}

// TranslateMessageToTNOS implements MessageTranslationHelper interface
func (n *nullTranslationHelper) TranslateMessageToTNOS(ctx context.Context, message interface{}) (interface{}, error) {
	return message, nil
}

// TranslateMessageFromTNOS implements MessageTranslationHelper interface
func (n *nullTranslationHelper) TranslateMessageFromTNOS(ctx context.Context, message interface{}) (interface{}, error) {
	return message, nil
}

// Translate implements a helper method for translation functions
func (n *nullTranslationHelper) Translate(ctx context.Context, key string, args ...interface{}) string {
	// Return the key as is for testing purposes
	return key
}
