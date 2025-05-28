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
	"fmt"
	"github-mcp-server/models"
	"github-mcp-server/pkg/bridge"
	"log"
	"time"
)

// Global bridge connection - initialized during server startup
var utilsMcpBridge *bridge.Bridge

// InitializeUtilsMCPBridge sets up the MCP bridge connection for utils package
func InitializeUtilsMCPBridge(options bridge.BridgeOptions) error {
	// WHO: BridgeInitializer
	// WHAT: Bridge setup
	// WHEN: During server initialization
	// WHERE: Server startup
	// WHY: To establish MCP connection
	// HOW: Using bridge creation
	// EXTENT: Single bridge instance

	if utilsMcpBridge != nil {
		return nil // Already initialized
	}

	utilsMcpBridge = bridge.NewBridge(options)
	if utilsMcpBridge == nil {
		return fmt.Errorf("failed to create MCP bridge")
	}

	log.Printf("Successfully initialized Utils MCP bridge")
	return nil
}

// GetUtilsMCPBridge returns the global bridge instance
func GetUtilsMCPBridge() *bridge.Bridge {
	// WHO: BridgeProvider
	// WHAT: Bridge access
	// WHEN: During tool execution
	// WHERE: MCP operations
	// WHY: To access bridge
	// HOW: Using singleton pattern
	// EXTENT: Bridge instance

	return utilsMcpBridge
}

// SendToUtilsTNOSMCP sends a message to TNOS MCP and returns the response
func SendToUtilsTNOSMCP(message models.MCPMessage) (*models.MCPMessage, error) {
	if utilsMcpBridge == nil {
		return nil, fmt.Errorf("utils MCP bridge not initialized")
	}
	startTime := time.Now()
	// Use SendCommand for tool execution
	respMsg, err := utilsMcpBridge.SendCommand(message.Tool, message.Parameters)
	if err != nil {
		log.Printf("Error sending request to MCP bridge: %v", err)
		return nil, fmt.Errorf("bridge communication error: %w", err)
	}
	elapsed := time.Since(startTime)
	log.Printf("MCP request for tool %s completed in %s", message.Tool, elapsed)
	response := &models.MCPMessage{
		Tool:       message.Tool,
		Parameters: message.Parameters,
		Context:    message.Context,
		Result:     make(map[string]interface{}),
	}
	// Extract result from response
	if respMsg.Payload != nil {
		response.Result = respMsg.Payload
	}
	if respMsg.Type == "error" {
		if respMsg.Payload != nil {
			if errMap, ok := respMsg.Payload["error"].(map[string]interface{}); ok {
				response.Error = &models.MCPError{
					Code:    fmt.Sprintf("%v", errMap["code"]),
					Message: fmt.Sprintf("%v", errMap["message"]),
					Details: fmt.Sprintf("%v", errMap["details"]),
				}
			}
		}
	}
	return response, nil
}
