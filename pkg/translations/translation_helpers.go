/*
 * WHO: TranslationProvider
 * WHAT: Translation helpers for MCP protocol
 * WHEN: During message translation
 * WHERE: System Layer 6 (Integration)
 * WHY: To adapt messages between TNOS and GitHub contexts
 * HOW: Using Go interface implementations
 * EXTENT: All translation needs for GitHub MCP
 */

package translations

import (
	"context"
)

// MessageTranslationHelper provides methods to translate between different contexts
// WHO: TranslationProvider
// WHAT: Interface for translation operations
// WHEN: During communication
// WHERE: System Layer 6 (Integration)
// WHY: To standardize context translation
// HOW: Using Go interface definition
// EXTENT: All translation implementations
type MessageTranslationHelper interface {
	// TranslateMessageToTNOS converts GitHub messages to TNOS format
	// WHO: MessageTranslator
	// WHAT: Message conversion to TNOS
	// WHEN: During inbound processing
	// WHERE: MCP Bridge
	// WHY: To ensure context compatibility
	// HOW: Using 7D context framework
	// EXTENT: All inbound GitHub messages
	TranslateMessageToTNOS(ctx context.Context, message interface{}) (interface{}, error)

	// TranslateMessageFromTNOS converts TNOS messages to GitHub format
	// WHO: MessageTranslator
	// WHAT: Message conversion from TNOS
	// WHEN: During outbound processing
	// WHERE: MCP Bridge
	// WHY: To ensure GitHub API compatibility
	// HOW: Using standard GitHub formats
	// EXTENT: All outbound TNOS messages
	TranslateMessageFromTNOS(ctx context.Context, message interface{}) (interface{}, error)
}

// nullMessageTranslationHelper provides a no-op implementation of MessageTranslationHelper
// WHO: NullObjectProvider
// WHAT: Empty translation implementation
// WHEN: During testing or when translation is not needed
// WHERE: System Layer 6 (Integration Tests)
// WHY: To provide a default implementation
// HOW: Using the Null Object pattern
// EXTENT: All test scenarios
type nullMessageTranslationHelper struct{}

// TranslateMessageToTNOS is a no-op implementation that returns the message unchanged
func (n *nullMessageTranslationHelper) TranslateMessageToTNOS(ctx context.Context, message interface{}) (interface{}, error) {
	return message, nil
}

// TranslateMessageFromTNOS is a no-op implementation that returns the message unchanged
func (n *nullMessageTranslationHelper) TranslateMessageFromTNOS(ctx context.Context, message interface{}) (interface{}, error) {
	return message, nil
}

// NullMessageTranslationHelper is a singleton instance of nullMessageTranslationHelper
// WHO: SingletonProvider
// WHAT: Global null translation helper
// WHEN: During system initialization
// WHERE: System Layer 6 (Integration)
// WHY: To provide a consistent default implementation
// HOW: Using Go variable initialization
// EXTENT: All test and default scenarios
var NullMessageTranslationHelper MessageTranslationHelper = &nullMessageTranslationHelper{}
