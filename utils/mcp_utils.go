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

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/internal/mcp"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/models"
)

var (
	// Global bridge connection - would be initialized during server startup
	mcpBridge *mcp.TNOSMCPBridge
)

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

	// Create a tool request with the provided parameters
	params := make(map[string]interface{})
	params["tool"] = message.Tool
	for k, v := range message.Parameters {
		params[k] = v
	}

	// Build the required request structure for mcp.CallToolRequest
	toolRequest := struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}{
		Name:      message.Tool,
		Arguments: message.Parameters,
	}

	// Create JSON string from the request
	requestJSON, err := json.Marshal(toolRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tool request: %w", err)
	}

	// Send the request via the bridge
	startTime := time.Now()
	result, err := mcpBridge.SendRequest(struct {
		Params struct {
			Name      string
			Arguments map[string]interface{}
		}
	}{
		Params: struct {
			Name      string
			Arguments map[string]interface{}
		}{
			Name:      message.Tool,
			Arguments: message.Parameters,
		},
	})

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
	if result != nil && len(result.Contents) > 0 {
		for _, content := range result.Contents {
			if content.Type == "text" {
				// Try to parse the text as JSON first
				var resultData map[string]interface{}
				if err := json.Unmarshal([]byte(content.Text), &resultData); err == nil {
					response.Result = resultData
				} else {
					// If not JSON, use as plain text result
					response.Result["text"] = content.Text
				}
			} else if content.Type == "error" {
				response.Error = &models.MCPError{
					Message: content.Text,
				}
			}
		}
	}

	return response, nil
}
