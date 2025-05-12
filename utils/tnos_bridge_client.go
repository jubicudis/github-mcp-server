// WHO: TNOSBridgeClient
// WHAT: Go HTTP client for TNOS MCP Bridge integration
// WHEN: During context sync, formula execution, or advanced compression
// WHERE: github-mcp-server/utils/tnos_bridge_client.go
// WHY: To enable GitHub MCP server to communicate with TNOS MCP bridge (Python)
// HOW: Provides functions to call bridge endpoints (e.g., /context-sync, /mobius-compress)
// EXTENT: All cross-system operations requiring bridge access

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// TNOSBridgeConfig holds configuration for the bridge client
var TNOSBridgeConfig = struct {
	BaseURL string
}{
	BaseURL: "http://127.0.0.1:8087", // Default bridge URL, update as needed
}

// Sync7DContext sends a 7D context to the TNOS MCP bridge and returns the translated context
func Sync7DContext(context map[string]interface{}) (map[string]interface{}, error) {
	url := TNOSBridgeConfig.BaseURL + "/context-sync"
	body, _ := json.Marshal(map[string]interface{}{"context": context})
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bridge returned status %d", resp.StatusCode)
	}
	respBody, _ := ioutil.ReadAll(resp.Body)
	var result struct {
		Context map[string]interface{} `json:"context"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return result.Context, nil
}

// MobiusCompressRemote calls the bridge to perform Möbius compression (Python-side)
func MobiusCompressRemote(data []byte, context map[string]interface{}) ([]byte, map[string]interface{}, error) {
	url := TNOSBridgeConfig.BaseURL + "/mobius-compress"
	body, _ := json.Marshal(map[string]interface{}{"data": data, "context": context})
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, nil, fmt.Errorf("bridge returned status %d", resp.StatusCode)
	}
	respBody, _ := ioutil.ReadAll(resp.Body)
	var result struct {
		Compressed []byte                 `json:"compressed"`
		Meta       map[string]interface{} `json:"meta"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, nil, err
	}
	return result.Compressed, result.Meta, nil
}

// --- Planck Dimension Integration ---
// Accept Planck dimension values in context (if present) and use in compressionVars and formula
// Planck dimensions: length, mass, time, charge, temperature, amount, luminous_intensity
// Use defaults if not provided
func getPlanckDimensions(context map[string]interface{}) (length, mass, timeP, charge, temp, amount, luminous float64) {
	length, mass, timeP, charge, temp, amount, luminous = 1, 1, 1, 1, 1, 1, 1
	if context == nil {
		return
	}
	if v, ok := context["planck_length"].(float64); ok {
		length = v
	}
	if v, ok := context["planck_mass"].(float64); ok {
		mass = v
	}
	if v, ok := context["planck_time"].(float64); ok {
		timeP = v
	}
	if v, ok := context["planck_charge"].(float64); ok {
		charge = v
	}
	if v, ok := context["planck_temperature"].(float64); ok {
		temp = v
	}
	if v, ok := context["planck_amount"].(float64); ok {
		amount = v
	}
	if v, ok := context["planck_luminous_intensity"].(float64); ok {
		luminous = v
	}
	return
}

// Ensure MobiusCompress, Sync7DContext, MobiusCompressRemote, getPlanckDimensions, etc. are only defined once in the utils package
// If any of these are duplicated in other files (e.g., helical_utils.go), remove the duplicate from tnos_bridge_client.go
