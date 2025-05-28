package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github-mcp-server/utils"
)

// WHO: HelicalToolsHandler
// WHAT: External MCP handlers for helical storage operations (Mobius/Helical)
// WHEN: During helical storage API requests
// WHERE: GitHub MCP Server
// WHY: To provide helical storage functionality to external clients using Möbius compression and dual-helix encoding
// HOW: Using Go handlers that route to TNOS via MCP bridge, leveraging Mobius and recursive helical algorithms
// EXTENT: All helical storage operations via MCP API

const errInvalidRequestPayload = "Invalid request payload"
const errContextSyncFailed = "Context sync failed"

// HelicalEncodeHandler handles requests to encode data using the Möbius/Helical algorithm
func HelicalEncodeHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var params struct {
		Data        []byte                 `json:"data"` // Accept raw bytes (not base64)
		StrandCount int                    `json:"strand_count,omitempty"`
		Context     map[string]interface{} `json:"context,omitempty"`
	}

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, errInvalidRequestPayload)
		return
	}

	if params.StrandCount == 0 {
		params.StrandCount = 2
	}

	// Sync context with TNOS MCP bridge (7D + Planck)
	newContext, err := utils.Sync7DContext(params.Context)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, errContextSyncFailed)
		return
	}

	// Use bridge for Möbius compression (Planck-aware)
	compressed, meta, err := utils.MobiusCompress(params.Data, newContext)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Compression failed")
		return
	}
	encoded, err := utils.HelicalEncode(compressed, params.StrandCount, meta)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Helical encoding failed")
		return
	}

	response := map[string]interface{}{
		"encoded_data":     encoded,
		"compression_meta": meta,
	}
	utils.RespondWithJSON(w, http.StatusOK, response)
}

// HelicalDecodeHandler handles requests to decode helically encoded data
func HelicalDecodeHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var params struct {
		EncodedData []byte                       `json:"encoded_data"`
		Meta        *utils.MobiusCompressionMeta `json:"compression_meta"`
		Context     map[string]interface{}       `json:"context,omitempty"`
	}

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, errInvalidRequestPayload)
		return
	}

	compressed, err := utils.HelicalDecode(params.EncodedData, params.Meta)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Helical decoding failed")
		return
	}

	decompressed, err := utils.MobiusDecompress(compressed, params.Meta)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Decompression failed")
		return
	}
	response := map[string]interface{}{
		"decoded_data": decompressed,
	}
	utils.RespondWithJSON(w, http.StatusOK, response)
}

// HelicalStoreHandler handles requests to store data using helical encoding
func HelicalStoreHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var params struct {
		Key         string                 `json:"key"`
		Data        []byte                 `json:"data"`
		StrandCount int                    `json:"strand_count,omitempty"`
		Context     map[string]interface{} `json:"context,omitempty"`
	}

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, errInvalidRequestPayload)
		return
	}
	if params.Key == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Key is required")
		return
	}
	if params.StrandCount == 0 {
		params.StrandCount = 2
	}

	newContext, err := utils.Sync7DContext(params.Context)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, errContextSyncFailed)
		return
	}

	compressed, meta, err := utils.MobiusCompress(params.Data, newContext)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Compression failed")
		return
	}
	encoded, err := utils.HelicalEncode(compressed, params.StrandCount, meta)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Helical encoding failed")
		return
	}
	err = utils.HelicalStore(params.Key, encoded, meta, newContext)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Storage failed")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{"status": "success"})
}

// HelicalRetrieveHandler handles requests to retrieve helically encoded data
func HelicalRetrieveHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var params struct {
		Key     string                 `json:"key"`
		Context map[string]interface{} `json:"context,omitempty"`
	}

	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, errInvalidRequestPayload)
		return
	}
	if params.Key == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Key is required")
		return
	}
	encoded, meta, err := utils.HelicalRetrieve(params.Key, params.Context)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Retrieval failed")
		return
	}
	compressed, err := utils.HelicalDecode(encoded, meta)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Helical decoding failed")
		return
	}

	decompressed, err := utils.MobiusDecompress(compressed, meta)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Decompression failed")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{"data": decompressed})
}
