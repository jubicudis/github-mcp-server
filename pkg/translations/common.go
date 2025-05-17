/*
 * WHO: TranslationsCommon
 * WHAT: Centralized translation utilities
 * WHEN: Throughout all translation operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide consistent translation behavior
 * HOW: By defining common interfaces and functions
 * EXTENT: All translation components
 */

package translations

import (
	"context"
	"encoding/json"
	"errors"
)

// Common error definitions
var (
	ErrInvalidContext = errors.New("invalid context")
	ErrInvalidMessage = errors.New("invalid message format")
	ErrEmptyMessage   = errors.New("empty message")
	ErrUnsupported    = errors.New("unsupported translation")
)

// Context keys
const (
	ContextKeyWho    = "who"
	ContextKeyWhat   = "what"
	ContextKeyWhen   = "when"
	ContextKeyWhere  = "where"
	ContextKeyWhy    = "why" 
	ContextKeyHow    = "how"
	ContextKeyExtent = "extent"
)

// Default values
const (
	DefaultWho    = "System"
	DefaultWhat   = "Communication"
	DefaultWhere  = "Integration"
	DefaultWhy    = "Protocol"
	DefaultHow    = "MCP"
	DefaultExtent = 1.0
)

// TranslationHelper interface consolidation from both helper files
// WHO: TranslationProvider
// WHAT: Unified translation interface
// WHEN: During all translation operations
// WHERE: All system layers
// WHY: To standardize translation operations
// HOW: Using a comprehensive interface
// EXTENT: All translation implementations
type TranslationHelper interface {
	// Core serialization operations
	ToJSON(v interface{}) (string, error)
	FromJSON(data string, v interface{}) error // Interface still uses FromJSON for implementations
	
	// Context operations
	ApplyContextVector(v interface{}, context map[string]string) (interface{}, error)
	
	// Message translation operations
	TranslateMessageToTNOS(ctx context.Context, message interface{}) (interface{}, error)
	TranslateMessageFromTNOS(ctx context.Context, message interface{}) (interface{}, error)
}

// WHO: JSONProcessor
// WHAT: Generic JSON encoding
// WHEN: During message serialization
// WHERE: All translation components
// WHY: To provide consistent JSON handling
// HOW: Using Go's json package
// EXTENT: All JSON operations
func ToJSON(v interface{}) (string, error) {
	if v == nil {
		return "{}", nil
	}
	
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	
	return string(data), nil
}

// WHO: JSONProcessor
// WHAT: Generic JSON decoding
// WHEN: During message deserialization
// WHERE: All translation components
// WHY: To provide consistent JSON handling
// HOW: Using Go's json package
// EXTENT: All JSON operations
func DecodeJSON(data string, v interface{}) error {
	if data == "" {
		return ErrEmptyMessage
	}
	
	return json.Unmarshal([]byte(data), v)
}

// WHO: ContextProcessor 
// WHAT: Extract standard 7D context
// WHEN: During context processing
// WHERE: All translation components
// WHY: To standardize context extraction
// HOW: Using type assertions and defaults
// EXTENT: All context operations
func Extract7DContext(ctx map[string]interface{}) map[string]string {
	result := make(map[string]string)
	
	// Extract with defaults
	if who, ok := ctx[ContextKeyWho].(string); ok && who != "" {
		result[ContextKeyWho] = who
	} else {
		result[ContextKeyWho] = DefaultWho
	}
	
	if what, ok := ctx[ContextKeyWhat].(string); ok && what != "" {
		result[ContextKeyWhat] = what
	} else {
		result[ContextKeyWhat] = DefaultWhat
	}
	
	if when, ok := ctx[ContextKeyWhen].(string); ok && when != "" {
		result[ContextKeyWhen] = when
	} else {
		result[ContextKeyWhen] = ""
	}
	
	if where, ok := ctx[ContextKeyWhere].(string); ok && where != "" {
		result[ContextKeyWhere] = where
	} else {
		result[ContextKeyWhere] = DefaultWhere
	}
	
	if why, ok := ctx[ContextKeyWhy].(string); ok && why != "" {
		result[ContextKeyWhy] = why
	} else {
		result[ContextKeyWhy] = DefaultWhy
	}
	
	if how, ok := ctx[ContextKeyHow].(string); ok && how != "" {
		result[ContextKeyHow] = how
	} else {
		result[ContextKeyHow] = DefaultHow
	}
	
	if extent, ok := ctx[ContextKeyExtent].(string); ok && extent != "" {
		result[ContextKeyExtent] = extent
	} else {
		result[ContextKeyExtent] = "1.0"
	}
	
	return result
}
