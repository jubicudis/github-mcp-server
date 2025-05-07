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
	"encoding/json"
)

// TranslationHelper provides methods for transforming data between formats
// WHO: TranslationHelper
// WHAT: Interface for translation operations
// WHEN: During data format conversion
// WHERE: MCP Bridge / ATL Layer
// WHY: To standardize translation operations
// HOW: Through defined interfaces
// EXTENT: All format conversions
type TranslationHelper interface {
	// ToJSON converts a value to its JSON representation
	ToJSON(v interface{}) (string, error)

	// FromJSON converts JSON to a value
	FromJSON(data string, v interface{}) error

	// ApplyContextVector applies 7D context to a value
	ApplyContextVector(v interface{}, context map[string]string) (interface{}, error)
}

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
	data, err := json.Marshal(v)
	return string(data), err
}

// FromJSON implements the TranslationHelper interface
func (n *nullTranslationHelper) FromJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}

// ApplyContextVector implements the TranslationHelper interface
func (n *nullTranslationHelper) ApplyContextVector(v interface{}, context map[string]string) (interface{}, error) {
	// No-op implementation for testing
	return v, nil
}
