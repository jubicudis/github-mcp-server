/*
 * WHO: MCPModels
 * WHAT: Model definitions for MCP data structures
 * WHEN: During MCP message processing
 * WHERE: System Layer 6 (Integration)
 * WHY: To define structured data for MCP interactions
 * HOW: Using Go struct definitions with proper tagging
 * EXTENT: All MCP message types
 */

package models

// MCPMessage represents a message sent to or received from TNOS MCP
type MCPMessage struct {
	// WHO: MessageDefinition
	// WHAT: Message structure
	Tool       string                 `json:"tool"`       // WHAT: Tool to execute
	Parameters map[string]interface{} `json:"parameters"` // HOW: Tool parameters
	Context    map[string]interface{} `json:"context"`    // WHERE: 7D context for execution
	Result     map[string]interface{} `json:"result"`     // WHY: Operation result
	Error      *MCPError              `json:"error"`      // EXTENT: Error details if unsuccessful
}

// MCPError represents an error response from MCP
type MCPError struct {
	// WHO: ErrorDefinition
	// WHAT: Error details
	Code    string `json:"code"`    // WHAT: Error classification
	Message string `json:"message"` // WHY: Error description
	Details string `json:"details"` // EXTENT: Additional context
}

// NewMCPMessage creates a new MCP message with specified tool and parameters
func NewMCPMessage(tool string, params map[string]interface{}) *MCPMessage {
	// WHO: MessageFactory
	// WHAT: Message creation
	// WHEN: Before sending to TNOS
	// WHERE: Message preparation
	// WHY: To standardize message creation
	// HOW: Using struct initialization
	// EXTENT: New message instance

	return &MCPMessage{
		Tool:       tool,
		Parameters: params,
		Context:    make(map[string]interface{}),
		Result:     make(map[string]interface{}),
	}
}

// WithContext adds context to an MCP message
func (m *MCPMessage) WithContext(context map[string]interface{}) *MCPMessage {
	// WHO: ContextEnricher
	// WHAT: Context addition
	// WHEN: Before sending to TNOS
	// WHERE: Message preparation
	// WHY: To provide execution context
	// HOW: Using context assignment
	// EXTENT: Updated message

	if context != nil {
		m.Context = context
	}
	return m
}

// IsError checks if the message contains an error
func (m *MCPMessage) IsError() bool {
	// WHO: ErrorDetector
	// WHAT: Error state check
	// WHEN: After receiving from TNOS
	// WHERE: Response processing
	// WHY: To determine success/failure
	// HOW: Using error presence check
	// EXTENT: Boolean status

	return m.Error != nil
}
