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

import "context"

// MessageTranslationHelper defines methods for translating messages between TNOS and GitHub
// WHO: InterfaceProvider
// WHAT: Message translation interface
// WHEN: During message processing
// WHERE: System Layer 6 (Integration)
// WHY: To define standard translation methods
// HOW: Using Go interface
// EXTENT: All translation needs for GitHub MCP
type MessageTranslationHelper interface {
	TranslateMessageToTNOS(ctx context.Context, message interface{}) (interface{}, error)
	TranslateMessageFromTNOS(ctx context.Context, message interface{}) (interface{}, error)
}

// This file is kept for backward compatibility.
// New code should use the interfaces and functions from common.go directly.
// The redundant MessageTranslationHelper interface has been consolidated into
// the TranslationHelper interface in common.go.
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
