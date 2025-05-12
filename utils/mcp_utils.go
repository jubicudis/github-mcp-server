/*
 * WHO: MCPUtilities
 * WHAT: MCP communication utilities
 * WHEN: During tool execution
 * WHERE: System Layer 6 (Integration)
 * WHY: To standardize MCP communication
 * HOW: Using structured messaging
 * EXTENT: All MCP bridge operations
 */

package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"tranquility-neuro-os/github-mcp-server/internal/mcp"
	"tranquility-neuro-os/github-mcp-server/models"

	mcpGo "github.com/mark3labs/mcp-go/mcp"
)

// Global bridge connection - would be initialized during server startup
var mcpBridge *mcp.TNOSMCPBridge

// InitializeMCPBridge sets up the MCP bridge connection
func InitializeMCPBridge(config mcp.BridgeConfig) error {
	// WHO: BridgeInitializer
	// WHAT: Bridge setup
	// WHEN: During server initialization
	// WHERE: Server startup
	// WHY: To establish MCP connection
	// HOW: Using bridge creation
	// EXTENT: Single bridge instance

	if mcpBridge != nil {
		return nil // Already initialized
	}

	mcpBridge = mcp.NewTNOSMCPBridge(config)
	if mcpBridge == nil {
		return fmt.Errorf("failed to create MCP bridge")
	}

	log.Printf("Successfully initialized MCP bridge")
	return nil
}

// GetMCPBridge returns the global bridge instance
func GetMCPBridge() *mcp.TNOSMCPBridge {
	// WHO: BridgeProvider
	// WHAT: Bridge access
	// WHEN: During tool execution
	// WHERE: MCP operations
	// WHY: To access bridge
	// HOW: Using singleton pattern
	// EXTENT: Bridge instance

	return mcpBridge
}

// SendToTNOSMCP sends a message to TNOS MCP and returns the response
func SendToTNOSMCP(message models.MCPMessage) (*models.MCPMessage, error) {
	// WHO: MessageSender
	// WHAT: MCP communication
	// WHEN: During tool execution
	// WHERE: Bridge operations
	// WHY: To execute remote tools
	// HOW: Using bridge protocol
	// EXTENT: Complete request-response cycle

	if mcpBridge == nil {
		return nil, fmt.Errorf("MCP bridge not initialized")
	}

	// Send the request via the bridge
	startTime := time.Now()

	// Create a CallToolRequest as required by the MCP bridge
	toolRequest := &mcp.CallToolRequest{}
	toolRequest.Params.Name = message.Tool
	toolRequest.Params.Arguments = message.Parameters

	result, err := mcpBridge.SendRequest(toolRequest)
	if err != nil {
		log.Printf("Error sending request to MCP bridge: %v", err)
		return nil, fmt.Errorf("bridge communication error: %w", err)
	}

	elapsed := time.Since(startTime)
	log.Printf("MCP request for tool %s completed in %s", message.Tool, elapsed)

	// Process the response - extract text content
	response := &models.MCPMessage{
		Tool:       message.Tool,
		Parameters: message.Parameters,
		Context:    message.Context,
		Result:     make(map[string]interface{}),
	}

	// Extract result from response
	if result != nil {
		// Extract text content from the first content item
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcpGo.TextContent); ok {
				// Try to parse the text as JSON
				var resultData map[string]interface{}
				if err := json.Unmarshal([]byte(textContent.Text), &resultData); err == nil {
					response.Result = resultData
				} else {
					// If parsing fails, store as raw text
					response.Result["raw"] = textContent.Text
				}
			}
		}

		// Check if the result was an error
		if result.IsError {
			errorMsg := "Error from tool execution"

			// Try to extract error message from content
			if len(result.Content) > 0 {
				if textContent, ok := result.Content[0].(mcpGo.TextContent); ok {
					errorMsg = textContent.Text
				}
			}

			response.Error = &models.MCPError{
				Message: errorMsg,
			}
		}
	}

	return response, nil
}
