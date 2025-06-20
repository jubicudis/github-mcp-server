// WHO: TranslationInterface
// WHAT: Defines translation helper interfaces
// WHEN: During MCP translation operations
// WHERE: System Layer 6 (Integration)
// WHY: To provide consistent translation interfaces
// HOW: Using Go interfaces
// EXTENT: All translation needs

package translations

import (
	"context"
	"encoding/json"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log"
)

// TranslationHelperFunc is a function type for simple string translations
type TranslationHelperFunc func(key string, defaultValue string) string

// NullTranslationHelperFunc is a no-op implementation that returns the default value
var NullTranslationHelperFunc TranslationHelperFunc = func(key string, defaultValue string) string {
	return defaultValue
}

// TranslationHelper defines the interface for translation helpers
type TranslationHelper interface {
	ToJSON(v interface{}) (string, error)
	FromJSON(data string, v interface{}) error
	ApplyContextVector(v interface{}, context map[string]string) (interface{}, error)
	TranslateMessageToTNOS(ctx context.Context, message interface{}) (interface{}, error)
	TranslateMessageFromTNOS(ctx context.Context, message interface{}) (interface{}, error)
}

// FromJSON parses JSON into a value, used when TranslationHelper is not available
func FromJSON(jsonStr string) (log.ContextVector7D, error) {
	var cv log.ContextVector7D
	if jsonStr == "" {
		return log.FromMap(map[string]interface{}{}), nil
	}
	err := json.Unmarshal([]byte(jsonStr), &cv)
	if err != nil {
		return log.FromMap(map[string]interface{}{}), err
	}
	return cv, nil
}
