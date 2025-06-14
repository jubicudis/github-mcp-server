// WHO: BridgeHelpers
// WHAT: Helper functions for bridge operations
// WHEN: During all bridge operations
// WHERE: System Layer 6 (Integration)
// WHY: To provide helper utilities for bridge
// HOW: Using standard Go functions
// EXTENT: All bridge operations

// TRANQUILSPEAK SYMBOL CLUSTER: [𝒯🩹βℏƒ𓆑#IN1𝑾𝑾𝑯𝑾𝑯𝑬𝑾𝑯𝑬𝑹𝑾𝑯𝒀𝑯𝑶𝑾𝑬𝑿⏳📍𝒮𝓔𝓗]
// This file is part of the 'integumentary' biosystem. See symbolic_mapping_registry_autogen_20250603.tsq for details.

package bridge

import (
	"time"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/common"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/translations"
)

// Constants from common.go made accessible in the bridge package
var (
	WriteTimeout   = common.WriteTimeout
	ReadTimeout    = common.ReadTimeout
	MaxMessageSize int64 = int64(common.MaxMessageSize) // Cast to int64 for SetReadLimit
)

// Error constants from common.go made accessible in the bridge package
var (
	ErrConnectionClosed = common.ErrConnectionClosed
	ErrBridgeNotConnected = common.ErrBridgeNotConnected
)

// Use common package helper functions
var (
	getString  = common.GetString
	getInt64   = common.GetInt64
	getFloat64 = common.GetFloat64
)

// ConvertToContextVector7D converts a map[string]interface{} to a ContextVector7D
func ConvertToContextVector7D(m map[string]interface{}) translations.ContextVector7D {
	if m == nil {
		m = make(map[string]interface{})
	}
	
	return translations.ContextVector7D{
		Who:    getString(m, "Who", getString(m, "who", "BridgeClient")),
		What:   getString(m, "What", getString(m, "what", "Communication")),
		When:   getInt64(m, "When", getInt64(m, "when", time.Now().Unix())),
		Where:  getString(m, "Where", getString(m, "where", "SystemLayer6")),
		Why:    getString(m, "Why", getString(m, "why", "Communication")),
		How:    getString(m, "How", getString(m, "how", "WebSocket")),
		Extent: getFloat64(m, "Extent", getFloat64(m, "extent", 1.0)),
		Source: getString(m, "Source", getString(m, "source", "GitHubMCPServer")),
	}
}

// HasLoggerInContext checks if the given map has a logger
func HasLoggerInContext(m map[string]interface{}) (*log.Logger, bool) {
	if m == nil {
		return nil, false
	}
	
	if val, ok := m["logger"]; ok {
		if logger, ok := val.(*log.Logger); ok {
			return logger, true
		}
	}
	return nil, false
}
