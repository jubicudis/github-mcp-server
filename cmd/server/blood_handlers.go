/*
 * WHO: BloodHandlers
 * WHAT: Blood circulation message processing for GitHub MCP server
 * WHEN: During blood bridge operation
 * WHERE: System Layer 6 (Integration)
 * WHY: To handle blood cell messages from TNOS MCP
 * HOW: Using blood circulation with filter callbacks
 * EXTENT: All blood circulation interactions
 */

package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/bridge"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/translations"
)

// processBloodCirculationMessages sets up handlers for blood circulation messages
func processBloodCirculationMessages(bloodCirculation *bridge.BloodCirculation) {
	// WHO: BloodProcessor
	// WHAT: Process blood circulation messages
	// WHEN: During blood bridge operation
	// WHERE: System Layer 6 (Integration)
	// WHY: To handle bidirectional communication
	// HOW: Using blood cell filtering and processing
	// EXTENT: All blood circulation message handling

	logger.Info("Setting up blood circulation message processing")

	// Register callback to handle blood cells
	bloodCirculation.SetBloodFilter(func(cell *bridge.BloodCell) bool {
		// Process the cell based on its type
		switch cell.Type {
		case "Red":
			// Red blood cells carry data (payload)
			logger.Debug("Received red blood cell", "id", cell.ID, "source", cell.Source)

			// Extract message type from payload
			var messageType string
			if typeVal, ok := cell.Payload["type"].(string); ok {
				messageType = typeVal
			} else {
				logger.Warn("Red blood cell missing message type", "id", cell.ID)
				return true // Continue processing other cells
			}

			// Process based on message type
			handleBloodCellMessage(bloodCirculation, cell, messageType)

		case "White":
			// White blood cells handle control messages
			logger.Debug("Received white blood cell", "id", cell.ID, "source", cell.Source)

			// Extract action from payload
			var action string
			if actionVal, ok := cell.Payload["action"].(string); ok {
				action = actionVal
			} else {
				logger.Warn("White blood cell missing action type", "id", cell.ID)
				return true // Continue processing other cells
			}

			// Process based on action
			handleBloodCellControl(bloodCirculation, cell, action)

		case "Platelet":
			// Platelets handle system health and recovery
			logger.Debug("Received platelet", "id", cell.ID, "source", cell.Source)

			// For heartbeats, just respond to keep the connection alive
			if action, ok := cell.Payload["action"].(string); ok && action == "pulse" {
				// Send heartbeat response
				response := map[string]interface{}{
					"type":      "control",
					"action":    "pulse_ack",
					"source":    "github-mcp-server",
					"timestamp": time.Now().Unix(),
				}
				// Create and queue a response cell
				responseCell := &bridge.BloodCell{
					ID:          "pulse-ack-" + time.Now().Format(time.RFC3339Nano),
					Type:        "Platelet",
					Payload:     response,
					Timestamp:   time.Now().Unix(),
					Source:      "github-mcp-server",
					Destination: cell.Source,
					Priority:    1, // Low priority for heartbeats
					OxygenLevel: 1.0,
					Context7D:   cell.Context7D,
				}
				bloodCirculation.QueueBloodCell(responseCell)
			}
		}

		// Return true to continue processing other cells
		return true
	})

	// Send an initialization message to the TNOS MCP server
	initContext := translations.ContextVector7D{
		Who:    "GitHubMCPServer",
		What:   "Initialize",
		When:   time.Now().Unix(),
		Where:  "github-mcp-server",
		Why:    "Establish Connection",
		How:    "Blood Bridge",
		Extent: 1.0,
		Source: "github-mcp-server",
	}

	initCell := &bridge.BloodCell{
		ID:          "init-" + time.Now().Format(time.RFC3339Nano),
		Type:        "White",
		Payload: map[string]interface{}{
			"type":   "control",
			"action": "initialize",
			"source": "github-mcp-server",
			"capabilities": []string{
				"github_api",
				"code_completion",
				"repo_management",
				"issue_tracking",
			},
		},
		Timestamp:   time.Now().Unix(),
		Source:      "github-mcp-server",
		Destination: "tnos-mcp-server",
		Priority:    10,
		OxygenLevel: 1.0,
		Context7D:   initContext,
	}

	// Queue the initialization cell
	bloodCirculation.QueueBloodCell(initCell)

	logger.Info("Blood circulation message processing initialized")

	// Start a goroutine to periodically send system metrics
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Get system metrics
				metrics := map[string]interface{}{
					"uptime":        time.Since(startTime).String(),
					"client_count":  len(clients),
					"memory_usage":  "N/A", // TODO: implement memory usage tracking
					"cpu_usage":     "N/A", // TODO: implement CPU usage tracking
					"github_errors": 0,     // TODO: track GitHub API errors
					"timestamp":     time.Now().Unix(),
				}

				// Create and queue a metrics cell
				metricsCell := &bridge.BloodCell{
					ID:          "metrics-" + time.Now().Format(time.RFC3339Nano),
					Type:        "Platelet",
					Payload: map[string]interface{}{
						"type":    "metrics",
						"metrics": metrics,
						"source":  "github-mcp-server",
					},
					Timestamp:   time.Now().Unix(),
					Source:      "github-mcp-server",
					Destination: "tnos-mcp-server",
					Priority:    3,
					OxygenLevel: 1.0,
					Context7D: translations.ContextVector7D{
						Who:    "SystemMonitor",
						What:   "Metrics",
						When:   time.Now().Unix(),
						Where:  "github-mcp-server",
						Why:    "Health Monitoring",
						How:    "Blood Bridge",
						Extent: 0.5,
						Source: "github-mcp-server",
					},
				}

				bloodCirculation.QueueBloodCell(metricsCell)
			}
		}
	}()
}

// Handle BloodCell messages
func handleBloodCellMessage(bloodCirculation *bridge.BloodCirculation, cell *bridge.BloodCell, messageType string) {
	// Get the formula registry for translations
	registry := bridge.GetBridgeFormulaRegistry()

	// Create context map for compression/decompression operations
	contextMap := map[string]interface{}{
		"who":    cell.Context7D.Who,
		"what":   cell.Context7D.What,
		"when":   cell.Context7D.When,
		"where":  cell.Context7D.Where,
		"why":    cell.Context7D.Why,
		"how":    cell.Context7D.How,
		"extent": cell.Context7D.Extent,
		"source": cell.Source,
	}

	// Check if payload is compressed and decompress if needed
	decompressed := false
	if compressed, ok := cell.Payload["_compressed"].(bool); ok && compressed {
		logger.Debug("Received compressed blood cell", "id", cell.ID, "algorithm", cell.Payload["_compression_algorithm"])

		// Attempt decompression
		if registry != nil {
			decompressedPayload, err := decompressPayload(cell.Payload, registry, contextMap)
			if err != nil {
				logger.Error("Failed to decompress payload", "error", err.Error())
			} else {
				cell.Payload = decompressedPayload
				decompressed = true
				logger.Debug("Successfully decompressed payload")
			}
		}
	}

	// Process based on message type
	switch messageType {
	case "github_event":
		// Handle GitHub event
		logger.Info("Blood cell carrying GitHub event", "id", cell.ID, "compressed", !decompressed && cell.Payload["_compressed"] == true)

		// Extract event data
		if eventData, ok := cell.Payload["data"].(map[string]interface{}); ok {
			// Process GitHub event
			// This is where we'd integrate with the GitHub API
			logger.Debug("Processing GitHub event data", "event_type", eventData["type"])

			// If we have a formula registry, translate the event
			if registry != nil {
				translationParams := map[string]interface{}{
					"data":        eventData,
					"formula_key": "github_event",
					"context":     contextMap,
				}

				result, err := registry.ExecuteFormula("github.translate", translationParams)
				if err != nil {
					logger.Error("Failed to translate GitHub event", "error", err.Error())
				} else {
					logger.Debug("Translated GitHub event successfully")

					// Broadcast event to WebSocket clients if applicable
					if translatedData, ok := result["translated_data"].(map[string]interface{}); ok {
						event := map[string]interface{}{
							"type":      "github_event",
							"source":    "blood_bridge",
							"timestamp": time.Now().Unix(),
							"data":      translatedData,
						}

						eventData, _ := json.Marshal(event)
						broadcast <- eventData
					}
				}
			}
		}

	case "query":
		// Handle query message
		logger.Info("Blood cell carrying query", "id", cell.ID)

		// Extract query
		query, ok := cell.Payload["query"].(string)
		if !ok {
			logger.Warn("Query cell missing query string", "id", cell.ID)
			return
		}

		// Process query based on type
		queryType, _ := cell.Payload["query_type"].(string)
		var responseData map[string]interface{}

		switch queryType {
		case "github_repo":
			// Query for GitHub repository info
			owner, _ := cell.Payload["owner"].(string)
			repo, _ := cell.Payload["repo"].(string)

			if owner != "" && repo != "" && gitHubClient != nil {
				// Call GitHub API
				repoInfo, err := gitHubClient.GetRepository(owner, repo)
				if err != nil {
					responseData = map[string]interface{}{
						"status": "error",
						"error":  err.Error(),
					}
				} else {
					responseData = map[string]interface{}{
						"status": "success",
						"data":   repoInfo,
					}
				}
			} else {
				responseData = map[string]interface{}{
					"status": "error",
					"error":  "Missing owner or repo parameters",
				}
			}

		case "system_status":
			// Return system status information
			responseData = map[string]interface{}{
				"status":        "success",
				"uptime":        time.Since(startTime).String(),
				"client_count":  len(clients),
				"capabilities":  []string{"github_api", "code_completion", "repo_management"},
				"timestamp":     time.Now().Unix(),
			}

		default:
			// Unknown query type
			responseData = map[string]interface{}{
				"status": "error",
				"error":  "Unknown query type: " + queryType,
			}
		}

		// Prepare response payload
		response := map[string]interface{}{
			"type":      "query_response",
			"query":     query,
			"query_id":  cell.ID, // Reference original query ID
			"timestamp": time.Now().Unix(),
			"response":  responseData,
		}

		// Check if we should compress large responses
		var payload map[string]interface{} = response
		if registry != nil && bloodCirculation.HasCompressionEnabled() {
			// Try to compress the response if it's large
			compressedPayload, err := compressPayload(
				response,
				registry,
				contextMap,
				bloodCirculation.GetCompressionThreshold(),
				[]string{"type", "query_id"},
			)

			if err != nil {
				logger.Warn("Failed to compress response payload", "error", err.Error())
			} else if compressedPayload["_compressed"] == true {
				payload = compressedPayload
				logger.Debug("Response payload compressed successfully",
					"original_size", compressedPayload["_original_size"],
					"compression_ratio", compressedPayload["_compression_ratio"])
			}
		}

		// Create and queue response cell
		responseCell := &bridge.BloodCell{
			ID:          "response-" + time.Now().Format(time.RFC3339Nano),
			Type:        "Red",
			Payload:     payload,
			Timestamp:   time.Now().Unix(),
			Source:      "github-mcp-server",
			Destination: cell.Source,
			Priority:    5,
			OxygenLevel: 1.0,
			Context7D:   cell.Context7D,
		}

		bloodCirculation.QueueBloodCell(responseCell)

	case "tnos_data":
		// Handle TNOS data
		logger.Info("Blood cell carrying TNOS data", "id", cell.ID)

		// Extract TNOS data
		if tnosData, ok := cell.Payload["data"].(map[string]interface{}); ok {
			// Process TNOS data
			logger.Debug("Processing TNOS data", "data_type", tnosData["type"])

			// Add specialized processing for different TNOS data types here
		}
	}
}

// handleBloodCellControl processes control messages in blood cells
func handleBloodCellControl(bloodCirculation *bridge.BloodCirculation, cell *bridge.BloodCell, action string) {
	// WHO: BloodControlHandler
	// WHAT: Process blood cell control messages
	// WHEN: During blood bridge operation
	// WHERE: System Layer 6 (Integration)
	// WHY: To handle system control messages
	// HOW: Using blood cell action handlers
	// EXTENT: All control messages

	registry := bridge.GetBridgeFormulaRegistry()

	// Create context map for operations
	contextMap := map[string]interface{}{
		"who":    cell.Context7D.Who,
		"what":   cell.Context7D.What,
		"when":   cell.Context7D.When,
		"where":  cell.Context7D.Where,
		"why":    cell.Context7D.Why,
		"how":    cell.Context7D.How,
		"extent": cell.Context7D.Extent,
		"source": cell.Source,
	}

	// Process based on action type
	switch action {
	case "initialize":
		// Handle initialization
		logger.Info("Blood circulation initialization message", "id", cell.ID, "source", cell.Source)

		// Send acknowledgment
		response := map[string]interface{}{
			"type":   "control",
			"action": "initialize_ack",
			"source": "github-mcp-server",
		}
		// Create and queue a response cell
		responseCell := &bridge.BloodCell{
			ID:          "init-ack-" + time.Now().Format(time.RFC3339Nano),
			Type:        "White",
			Payload:     response,
			Timestamp:   time.Now().Unix(),
			Source:      "github-mcp-server",
			Destination: cell.Source,
			Priority:    10,
			OxygenLevel: 1.0,
			Context7D:   cell.Context7D,
		}
		bloodCirculation.QueueBloodCell(responseCell)

	case "heartbeat":
		// Handle heartbeat
		logger.Debug("Received heartbeat from blood circulation", "id", cell.ID)

		// Send heartbeat response
		response := map[string]interface{}{
			"type":       "control",
			"action":     "heartbeat_ack",
			"source":     "github-mcp-server",
			"timestamp":  time.Now().Unix(),
		}
		// Create and queue a response cell
		responseCell := &bridge.BloodCell{
			ID:          "heartbeat-ack-" + time.Now().Format(time.RFC3339Nano),
			Type:        "White",
			Payload:     response,
			Timestamp:   time.Now().Unix(),
			Source:      "github-mcp-server",
			Destination: cell.Source,
			Priority:    5,
			OxygenLevel: 1.0,
			Context7D:   cell.Context7D,
		}
		bloodCirculation.QueueBloodCell(responseCell)

	case "shutdown":
		// Handle shutdown
		logger.Info("Blood circulation shutdown requested", "id", cell.ID, "source", cell.Source)

		// Send acknowledgment before shutting down
		response := map[string]interface{}{
			"type":   "control",
			"action": "shutdown_ack",
			"source": "github-mcp-server",
		}
		// Create and queue a response cell
		responseCell := &bridge.BloodCell{
			ID:          "shutdown-ack-" + time.Now().Format(time.RFC3339Nano),
			Type:        "White",
			Payload:     response,
			Timestamp:   time.Now().Unix(),
			Source:      "github-mcp-server",
			Destination: cell.Source,
			Priority:    10,
			OxygenLevel: 1.0,
			Context7D:   cell.Context7D,
		}
		bloodCirculation.QueueBloodCell(responseCell)
	case "ping":
		// Respond to ping with pong
		logger.Debug("Received ping", "id", cell.ID, "source", cell.Source)

		// Create and send pong response
		responseCell := &bridge.BloodCell{
			ID:          "pong-" + time.Now().Format(time.RFC3339Nano),
			Type:        "White",
			Payload: map[string]interface{}{
				"action": "pong",
				"reference_id": cell.ID,
				"timestamp": time.Now().Unix(),
				"source": "github-mcp-server",
			},
			Timestamp:   time.Now().Unix(),
			Source:      "github-mcp-server",
			Destination: cell.Source,
			Priority:    5, // Medium priority for control messages
			OxygenLevel: 1.0,
			Context7D:   cell.Context7D,
		}

		bloodCirculation.QueueBloodCell(responseCell)

	case "config_update":
		// Handle configuration update request
		logger.Info("Received configuration update request", "id", cell.ID, "source", cell.Source)

		// Extract configuration settings
		if config, ok := cell.Payload["config"].(map[string]interface{}); ok {
			// Update compression settings if provided
			if compressionConfig, ok := config["compression"].(map[string]interface{}); ok {
				if err := bloodCirculation.SetCompressionOptions(compressionConfig); err != nil {
					logger.Error("Failed to update compression settings", "error", err.Error())
				} else {
					logger.Info("Updated compression settings", "config", compressionConfig)
				}
			}

			// Send acknowledgement
			responseCell := &bridge.BloodCell{
				ID:          "config-ack-" + time.Now().Format(time.RFC3339Nano),
				Type:        "White",
				Payload: map[string]interface{}{
					"action": "config_ack",
					"reference_id": cell.ID,
					"status": "success",
					"timestamp": time.Now().Unix(),
					"source": "github-mcp-server",
				},
				Timestamp:   time.Now().Unix(),
				Source:      "github-mcp-server",
				Destination: cell.Source,
				Priority:    5,
				OxygenLevel: 1.0,
				Context7D:   cell.Context7D,
			}

			bloodCirculation.QueueBloodCell(responseCell)
		}

	case "formula_request":
		// Handle formula execution request
		logger.Debug("Received formula execution request", "id", cell.ID, "source", cell.Source)

		formulaID, _ := cell.Payload["formula_id"].(string)
		params, _ := cell.Payload["parameters"].(map[string]interface{})

		if registry == nil {
			logger.Error("Formula registry not available", "formula", formulaID)

			// Send error response
			responseCell := &bridge.BloodCell{
				ID:          "formula-error-" + time.Now().Format(time.RFC3339Nano),
				Type:        "White",
				Payload: map[string]interface{}{
					"action": "formula_error",
					"reference_id": cell.ID,
					"error": "Formula registry not available",
					"formula_id": formulaID,
				},
				Timestamp:   time.Now().Unix(),
				Source:      "github-mcp-server",
				Destination: cell.Source,
				Priority:    5,
				OxygenLevel: 1.0,
				Context7D:   cell.Context7D,
			}

			bloodCirculation.QueueBloodCell(responseCell)
			return
		}

		// Execute the formula
		result, err := registry.ExecuteFormula(formulaID, params)
		if err != nil {
			logger.Error("Formula execution failed", "formula", formulaID, "error", err.Error())

			// Send error response
			responseCell := &bridge.BloodCell{
				ID:          "formula-error-" + time.Now().Format(time.RFC3339Nano),
				Type:        "White",
				Payload: map[string]interface{}{
					"action": "formula_error",
					"reference_id": cell.ID,
					"error": err.Error(),
					"formula_id": formulaID,
				},
				Timestamp:   time.Now().Unix(),
				Source:      "github-mcp-server",
				Destination: cell.Source,
				Priority:    5,
				OxygenLevel: 1.0,
				Context7D:   cell.Context7D,
			}

			bloodCirculation.QueueBloodCell(responseCell)
		} else {
			logger.Debug("Formula executed successfully", "formula", formulaID)

			// Check if result needs compression
			resultPayload := map[string]interface{}{
				"action": "formula_result",
				"reference_id": cell.ID,
				"formula_id": formulaID,
				"result": result,
				"timestamp": time.Now().Unix(),
			}

			// Compress if needed
			var payload map[string]interface{}
			if bloodCirculation.HasCompressionEnabled() {
				compressedPayload, err := compressPayload(
					resultPayload,
					registry,
					contextMap,
					bloodCirculation.GetCompressionThreshold(),
					[]string{"action", "reference_id", "formula_id"},
				)

				if err != nil {
					logger.Warn("Failed to compress formula result", "error", err.Error())
					payload = resultPayload
				} else {
					payload = compressedPayload
				}
			} else {
				payload = resultPayload
			}

			// Send result response
			responseCell := &bridge.BloodCell{
				ID:          "formula-result-" + time.Now().Format(time.RFC3339Nano),
				Type:        "White",
				Payload:     payload,
				Timestamp:   time.Now().Unix(),
				Source:      "github-mcp-server",
				Destination: cell.Source,
				Priority:    5,
				OxygenLevel: 1.0,
				Context7D:   cell.Context7D,
			}

			bloodCirculation.QueueBloodCell(responseCell)
		}

	case "status_request":
		// Handle status request
		logger.Debug("Received status request", "id", cell.ID, "source", cell.Source)

		// Collect system status information
		status := map[string]interface{}{
			"uptime":         time.Since(startTime).String(),
			"client_count":   len(clients),
			"timestamp":      time.Now().Unix(),
			"blood_metrics":  bloodCirculation.GetCirculationMetrics(),
			"compression":    bloodCirculation.GetCompressionOptions(),
			"formula_count": 0,
		}

		// Add formula count if registry available
		if registry != nil {
			status["formula_count"] = len(registry.ListFormulas())
		}

		// Send status response
		responseCell := &bridge.BloodCell{
			ID:          "status-" + time.Now().Format(time.RFC3339Nano),
			Type:        "White",
			Payload: map[string]interface{}{
				"action": "status_response",
				"reference_id": cell.ID,
				"status": status,
			},
			Timestamp:   time.Now().Unix(),
			Source:      "github-mcp-server",
			Destination: cell.Source,
			Priority:    5,
			OxygenLevel: 1.0,
			Context7D:   cell.Context7D,
		}

		bloodCirculation.QueueBloodCell(responseCell)

	default:
		// Unknown action
		logger.Warn("Unknown blood cell action", "action", action, "id", cell.ID)
	}
}

// compressPayload applies Hemoflux compression to the payload if it exceeds the threshold
func compressPayload(payload map[string]interface{}, registry *bridge.BridgeFormulaRegistry, context map[string]interface{}, threshold int, preserveKeys []string) (map[string]interface{}, error) {
	// WHO: CompressionManager
	// WHAT: Apply Hemoflux compression to large payloads
	// WHEN: Before transmitting large blood cells
	// WHERE: System Layer 6 (Integration)
	// WHY: To optimize bandwidth usage
	// HOW: Using Hemoflux compression formulas
	// EXTENT: Large payloads in blood cells

	// Skip compression if registry is not available
	if registry == nil {
		return payload, fmt.Errorf("formula registry not available for compression")
	}

	// Make a copy of the payload to preserve original
	compressedPayload := make(map[string]interface{})
	for k, v := range payload {
		compressedPayload[k] = v
	}

	// Create a JSON string of the payload to determine size
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return payload, fmt.Errorf("failed to marshal payload for compression: %w", err)
	}

	// Skip if below threshold
	if len(jsonData) <= threshold {
		return payload, nil
	}

	// Preserve specific keys
	preservedValues := make(map[string]interface{})
	for _, key := range preserveKeys {
		if val, ok := payload[key]; ok {
			preservedValues[key] = val
		}
	}

	// Prepare payload for compression
	formulaParams := map[string]interface{}{
		"data":    payload,
		"context": context,
	}

	// Execute compression formula
	result, err := registry.ExecuteFormula("hemoflux.compress", formulaParams)
	if err != nil {
		return payload, fmt.Errorf("hemoflux compression failed: %w", err)
	}

	// Check if we got a valid result
	if result == nil {
		return payload, fmt.Errorf("compression returned nil result")
	}

	// Extract compressed data
	if compressedData, ok := result["compressed_data"].(map[string]interface{}); ok {
		// Mark as compressed
		compressedData["_compressed"] = true
		compressedData["_compression_algorithm"] = "hemoflux.mobius"
		compressedData["_original_size"] = len(jsonData)

		// Restore preserved keys
		for k, v := range preservedValues {
			compressedData[k] = v
		}

		return compressedData, nil
	}

	return payload, fmt.Errorf("invalid compression result format")
}

// decompressPayload decompresses a Hemoflux-compressed payload
func decompressPayload(payload map[string]interface{}, registry *bridge.BridgeFormulaRegistry, context map[string]interface{}) (map[string]interface{}, error) {
	// WHO: DecompressionManager
	// WHAT: Decompress Hemoflux-compressed payloads
	// WHEN: After receiving compressed blood cells
	// WHERE: System Layer 6 (Integration)
	// WHY: To restore original data
	// HOW: Using Hemoflux decompression formulas
	// EXTENT: Compressed payloads in blood cells

	// Check if payload is compressed
	if compressed, ok := payload["_compressed"].(bool); !ok || !compressed {
		return payload, nil // Not compressed
	}

	// Skip decompression if registry is not available
	if registry == nil {
		return payload, fmt.Errorf("formula registry not available for decompression")
	}

	// Prepare payload for decompression
	formulaParams := map[string]interface{}{
		"compressed_data": payload,
		"context":         context,
	}

	// Execute decompression formula
	result, err := registry.ExecuteFormula("hemoflux.decompress", formulaParams)
	if err != nil {
		return payload, fmt.Errorf("hemoflux decompression failed: %w", err)
	}

	// Check if we got a valid result
	if result == nil {
		return payload, fmt.Errorf("decompression returned nil result")
	}

	// Extract decompressed data
	if decompressedData, ok := result["decompressed_data"].(map[string]interface{}); ok {
		return decompressedData, nil
	}

	return payload, fmt.Errorf("invalid decompression result format")
}
