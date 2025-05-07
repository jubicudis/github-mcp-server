package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"tranquility-neuro-os/github-mcp-server/models"
	"tranquility-neuro-os/github-mcp-server/utils"
)

// WHO: HelicalToolsHandler
// WHAT: External MCP handlers for helical storage operations
// WHEN: During helical storage API requests
// WHERE: GitHub MCP Server
// WHY: To provide helical storage functionality to external clients
// HOW: Using Go handlers that route to TNOS via MCP bridge
// EXTENT: All helical storage operations via MCP API

// HelicalEncodeHandler handles requests to encode data using helical algorithm
func HelicalEncodeHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	// Extract request parameters
	var params struct {
		Data        string                 `json:"data"`
		StrandCount int                    `json:"strand_count,omitempty"`
		Context     map[string]interface{} `json:"context,omitempty"`
	}

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Decode base64 data
	rawData, err := base64.StdEncoding.DecodeString(params.Data)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid base64 data encoding")
		return
	}

	// Set default strand count if not provided
	if params.StrandCount == 0 {
		params.StrandCount = 2
	}

	// Prepare MCP message for TNOS
	message := models.MCPMessage{
		Tool: "helical_encode",
		Parameters: map[string]interface{}{
			"data":         rawData,
			"strand_count": params.StrandCount,
		},
		Context: params.Context,
	}

	// Send message to TNOS MCP bridge
	response, err := utils.SendToTNOSMCP(message)
	if err != nil {
		log.Printf("Error sending to TNOS MCP: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Error processing request")
		return
	}

	// Process and return response
	encodedData, exists := response.Result["encoded_data"]
	if !exists {
		utils.RespondWithError(w, http.StatusInternalServerError, "Invalid response from TNOS")
		return
	}

	// Convert binary data to base64 for JSON response
	if binData, ok := encodedData.([]byte); ok {
		response.Result["encoded_data"] = base64.StdEncoding.EncodeToString(binData)
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

// HelicalDecodeHandler handles requests to decode helically encoded data
func HelicalDecodeHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	// Extract request parameters
	var params struct {
		EncodedData string                 `json:"encoded_data"`
		Metadata    map[string]interface{} `json:"metadata"`
		Context     map[string]interface{} `json:"context,omitempty"`
	}

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Decode base64 data
	rawData, err := base64.StdEncoding.DecodeString(params.EncodedData)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid base64 data encoding")
		return
	}

	// Prepare MCP message for TNOS
	message := models.MCPMessage{
		Tool: "helical_decode",
		Parameters: map[string]interface{}{
			"encoded_data": rawData,
			"metadata":     params.Metadata,
		},
		Context: params.Context,
	}

	// Send message to TNOS MCP bridge
	response, err := utils.SendToTNOSMCP(message)
	if err != nil {
		log.Printf("Error sending to TNOS MCP: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Error processing request")
		return
	}

	// Process and return response
	decodedData, exists := response.Result["decoded_data"]
	if !exists {
		utils.RespondWithError(w, http.StatusInternalServerError, "Invalid response from TNOS")
		return
	}

	// Convert binary data to base64 for JSON response
	if binData, ok := decodedData.([]byte); ok {
		response.Result["decoded_data"] = base64.StdEncoding.EncodeToString(binData)
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

// HelicalStoreHandler handles requests to store data using helical encoding
func HelicalStoreHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	// Extract request parameters
	var params struct {
		Key         string                 `json:"key"`
		Data        string                 `json:"data"`
		StrandCount int                    `json:"strand_count,omitempty"`
		Context     map[string]interface{} `json:"context,omitempty"`
	}

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate required parameters
	if params.Key == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Key is required")
		return
	}

	// Decode base64 data
	rawData, err := base64.StdEncoding.DecodeString(params.Data)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid base64 data encoding")
		return
	}

	// Set default strand count if not provided
	if params.StrandCount == 0 {
		params.StrandCount = 2
	}

	// Prepare MCP message for TNOS
	message := models.MCPMessage{
		Tool: "helical_store",
		Parameters: map[string]interface{}{
			"key":          params.Key,
			"data":         rawData,
			"strand_count": params.StrandCount,
		},
		Context: params.Context,
	}

	// Send message to TNOS MCP bridge
	response, err := utils.SendToTNOSMCP(message)
	if err != nil {
		log.Printf("Error sending to TNOS MCP: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Error processing request")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

// HelicalRetrieveHandler handles requests to retrieve helically encoded data
func HelicalRetrieveHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	// Extract request parameters
	var params struct {
		Key     string                 `json:"key"`
		Context map[string]interface{} `json:"context,omitempty"`
	}

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate required parameters
	if params.Key == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Key is required")
		return
	}

	// Prepare MCP message for TNOS
	message := models.MCPMessage{
		Tool: "helical_retrieve",
		Parameters: map[string]interface{}{
			"key": params.Key,
		},
		Context: params.Context,
	}

	// Send message to TNOS MCP bridge
	response, err := utils.SendToTNOSMCP(message)
	if err != nil {
		log.Printf("Error sending to TNOS MCP: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error processing request: %v", err))
		return
	}

	// Process and return response
	data, exists := response.Result["data"]
	if !exists {
		utils.RespondWithError(w, http.StatusInternalServerError, "Invalid response from TNOS")
		return
	}

	// Convert binary data to base64 for JSON response
	if binData, ok := data.([]byte); ok {
		response.Result["data"] = base64.StdEncoding.EncodeToString(binData)
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}
