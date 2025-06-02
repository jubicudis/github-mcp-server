package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	logpkg "github-mcp-server/pkg/log"
)

// CompressionRequest represents a request to the TNOS MCP server for compression or decompression.
type CompressionRequest struct {
	Type    string                 `json:"type"` // "compress" or "decompress"
	Data    string                 `json:"data"`
	Context map[string]interface{} `json:"context,omitempty"` // 7D context: who, what, when, where, why, how, intent
	Params  map[string]interface{} `json:"params,omitempty"`  // includes formula_key, geo, file_path, runtime_path, transfer_distance, etc.
}

// CompressionResponse represents a response from the TNOS MCP server.
type CompressionResponse struct {
	Success          bool                   `json:"success"`
	CompressedData   string                 `json:"compressed_data,omitempty"`
	DecompressedData string                 `json:"decompressed_data,omitempty"`
	CompressionVars  map[string]interface{} `json:"compression_vars,omitempty"`
	Entropy          float64                `json:"entropy,omitempty"`
	CompressionRatio float64                `json:"compression_ratio,omitempty"`
	LocationMetadata map[string]interface{} `json:"location_metadata,omitempty"`
	EnergyMetrics    map[string]interface{} `json:"energy_metrics,omitempty"`
	Error            string                 `json:"error,omitempty"`
}

// CompressionBridge manages communication with the TNOS MCP server for compression.
type CompressionBridge struct {
	ServerURL string
	Client    *http.Client
}

// NewCompressionBridge creates a new CompressionBridge.
func NewCompressionBridge(serverURL string) *CompressionBridge {
	return &CompressionBridge{
		ServerURL: serverURL,
		Client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// Compress sends data to the TNOS MCP server for compression.
func (cb *CompressionBridge) Compress(ctx context.Context, data string, contextMap map[string]interface{}, params map[string]interface{}) (*CompressionResponse, error) {
	if params == nil {
		params = make(map[string]interface{})
	}
	// Ensure 7D context and metadata fields are present
	for _, k := range []string{"who", "what", "when", "where", "why", "how", "intent", "geo", "file_path", "runtime_path", "transfer_distance", "formula_key"} {
		if _, ok := params[k]; !ok {
			params[k] = ""
		}
	}
	request := CompressionRequest{
		Type:    "compress",
		Data:    data,
		Context: contextMap,
		Params:  params,
	}
	resp, err := cb.sendRequest(ctx, request)
	logCompressionOperation("compress", data, request, resp, err)
	return resp, err
}

// Decompress sends data to the TNOS MCP server for decompression.
func (cb *CompressionBridge) Decompress(ctx context.Context, compressedData string, compressionVars map[string]interface{}) (*CompressionResponse, error) {
	params := map[string]interface{}{"compression_vars": compressionVars}
	// Pass through 7D context and metadata if present in compressionVars
	for _, k := range []string{"who", "what", "when", "where", "why", "how", "intent", "geo", "file_path", "runtime_path", "transfer_distance", "formula_key"} {
		if v, ok := compressionVars[k]; ok {
			params[k] = v
		}
	}
	request := CompressionRequest{
		Type:   "decompress",
		Data:   compressedData,
		Params: params,
	}
	resp, err := cb.sendRequest(ctx, request)
	logCompressionOperation("decompress", compressedData, request, resp, err)
	return resp, err
}

// sendRequest marshals the request and sends it to the TNOS MCP server.
func (cb *CompressionBridge) sendRequest(ctx context.Context, req CompressionRequest) (*CompressionResponse, error) {
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	url := cb.ServerURL + "/api/compression"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := cb.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var compResp CompressionResponse
	if err := json.NewDecoder(resp.Body).Decode(&compResp); err != nil {
		return nil, err
	}
	if !compResp.Success {
		return &compResp, errors.New(compResp.Error)
	}
	return &compResp, nil
}

// logCompressionOperation logs all compression/decompression operations to Helical Memory (TNOS-compliant)
func logCompressionOperation(opType, data string, req CompressionRequest, resp *CompressionResponse, err error) {
	meta := map[string]interface{}{
		"request_params": req.Params,
	}
	if resp != nil {
		meta["success"] = resp.Success
		meta["entropy"] = resp.Entropy
		meta["compression_ratio"] = resp.CompressionRatio
		meta["location_metadata"] = resp.LocationMetadata
		meta["energy_metrics"] = resp.EnergyMetrics
		meta["error"] = resp.Error
	}
	if err != nil {
		meta["go_error"] = err.Error()
	}
	context := req.Context
	who := "CompressionBridge"
	what := opType
	where := "github-mcp-server"
	why := "compression"
	how := "Go"
	extent := "full"
	if context != nil {
		if v, ok := context["who"].(string); ok && v != "" {
			who = v
		}
		if v, ok := context["what"].(string); ok && v != "" {
			what = v
		}
		if v, ok := context["where"].(string); ok && v != "" {
			where = v
		}
		if v, ok := context["why"].(string); ok && v != "" {
			why = v
		}
		if v, ok := context["how"].(string); ok && v != "" {
			how = v
		}
		if v, ok := context["extent"].(string); ok && v != "" {
			extent = v
		}
	}
	msg := fmt.Sprintf("%s operation, data_length=%d", opType, len(data))
	event := logpkg.NewHelicalEvent(who, what, where, why, how, extent, msg)
	event.Meta = meta
	_ = logpkg.LogHelicalEvent(event)
}

// Example usage:
// cb := NewCompressionBridge("http://localhost:9001")
// resp, err := cb.Compress(context.Background(), "my data", nil, nil)
// if err != nil { ... }
// fmt.Println(resp.CompressedData)
