/*
 * WHO: UtilsPackage
 * WHAT: Utility functions for GitHub MCP server
 * WHEN: Throughout system operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide common helper functions
 * HOW: Using Go functions with 7D context
 * EXTENT: All utility needs
 */

package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/tranquility-dev/github-mcp-server/models"
)

// GenerateID generates a unique ID for requests
func GenerateID(prefix string) string {
	// WHO: IDGenerator
	// WHAT: Generate unique ID
	// WHEN: During request creation
	// WHERE: System Layer 6 (Integration)
	// WHY: For unique identification
	// HOW: Using random bytes and encoding
	// EXTENT: Single ID generation

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}

	return fmt.Sprintf("%s-%s", prefix, base64.RawURLEncoding.EncodeToString(b))
}

// ParseURI parses an MCP URI into its components
func ParseURI(uri string) (resourceType string, owner string, repo string, path string, err error) {
	// WHO: URIParser
	// WHAT: Parse URI
	// WHEN: During request processing
	// WHERE: System Layer 6 (Integration)
	// WHY: To extract URI components
	// HOW: Using string parsing
	// EXTENT: Single URI parsing

	// Example URIs:
	// repo://{owner}/{repo}/contents{/path*}
	// repo://{owner}/{repo}/refs/heads/{branch}/contents{/path*}
	// repo://{owner}/{repo}/sha/{sha}/contents{/path*}

	if !strings.HasPrefix(uri, "repo://") {
		return "", "", "", "", fmt.Errorf("invalid URI format, must start with repo://")
	}

	// Remove the scheme
	uri = strings.TrimPrefix(uri, "repo://")

	// Split the path parts
	parts := strings.Split(uri, "/")
	if len(parts) < 2 {
		return "", "", "", "", fmt.Errorf("invalid URI format, must contain owner and repo")
	}

	owner = parts[0]
	repo = parts[1]

	// Join the remaining parts back into path
	if len(parts) > 2 {
		path = strings.Join(parts[2:], "/")
	}

	return "repo", owner, repo, path, nil
}

// CompressWithMobius applies the Möbius Compression Formula to a value
func CompressWithMobius(value float64, context *models.ContextVector7D) (float64, map[string]float64) {
	// WHO: CompressionEngine
	// WHAT: Compress value using Möbius formula
	// WHEN: During data compression
	// WHERE: System Layer 6 (Integration)
	// WHY: For efficient data storage
	// HOW: Using Möbius Compression Formula
	// EXTENT: Single value compression

	// Extract compression factors from context
	var B, V, I, G, F float64

	if context != nil && context.Meta != nil {
		if b, ok := context.Meta["B"].(float64); ok {
			B = b
		} else {
			B = 0.9 // Default Base factor
		}

		if v, ok := context.Meta["V"].(float64); ok {
			V = v
		} else {
			V = 0.8 // Default Value factor
		}

		if i, ok := context.Meta["I"].(float64); ok {
			I = i
		} else {
			I = 0.7 // Default Intent factor
		}

		if g, ok := context.Meta["G"].(float64); ok {
			G = g
		} else {
			G = 0.6 // Default Growth factor
		}

		if f, ok := context.Meta["F"].(float64); ok {
			F = f
		} else {
			F = 0.5 // Default Flexibility factor
		}
	} else {
		B = 0.9 // Default Base factor
		V = 0.8 // Default Value factor
		I = 0.7 // Default Intent factor
		G = 0.6 // Default Growth factor
		F = 0.5 // Default Flexibility factor
	}

	// Calculate entropy (simplified)
	entropy := math.Log2(math.Abs(value) + 1)

	// Calculate time factor
	t := 0.1 // Default time factor
	if context != nil {
		t = float64(time.Now().Unix()-context.When) / 86400.0 // Days since context creation
		if t < 0.1 {
			t = 0.1 // Minimum time factor
		}
	}

	// Calculate energy factors
	E := 0.5     // Default energy factor
	C_sum := 0.1 // Default C_sum
	alignment := (B + V*I) * math.Exp(-t*E)

	// Apply the Möbius Compression Formula
	compressed := (value * B * I * (1 - (entropy / math.Log2(1+V))) * (G + F)) /
		(E*t + C_sum*entropy + alignment)

	// Store compression variables for decompression
	variables := map[string]float64{
		"B":         B,
		"V":         V,
		"I":         I,
		"G":         G,
		"F":         F,
		"t":         t,
		"E":         E,
		"C_sum":     C_sum,
		"entropy":   entropy,
		"alignment": alignment,
		"original":  value,
	}

	return compressed, variables
}

// DecompressWithMobius reverses the Möbius Compression Formula
func DecompressWithMobius(compressed float64, variables map[string]float64) float64 {
	// WHO: DecompressionEngine
	// WHAT: Decompress value using Möbius formula
	// WHEN: During data decompression
	// WHERE: System Layer 6 (Integration)
	// WHY: To restore original values
	// HOW: Using reversed Möbius formula
	// EXTENT: Single value decompression

	// If we have the original value stored, return it
	if original, ok := variables["original"]; ok {
		return original
	}

	// Extract compression variables
	B := variables["B"]
	V := variables["V"]
	I := variables["I"]
	G := variables["G"]
	F := variables["F"]
	t := variables["t"]
	E := variables["E"]
	C_sum := variables["C_sum"]
	entropy := variables["entropy"]
	alignment := variables["alignment"]

	// Reverse the Möbius Compression Formula
	value := compressed * (E*t + C_sum*entropy + alignment) / (B * I * (1 - (entropy / math.Log2(1+V))) * (G + F))

	return value
}

// MergeContexts merges multiple 7D contexts with appropriate weighting
func MergeContexts(contexts []*models.ContextVector7D) *models.ContextVector7D {
	// WHO: ContextMerger
	// WHAT: Merge multiple contexts
	// WHEN: During context aggregation
	// WHERE: System Layer 6 (Integration)
	// WHY: To combine context information
	// HOW: Using weighted averaging
	// EXTENT: Multiple context dimensions

	if len(contexts) == 0 {
		return nil
	}

	if len(contexts) == 1 {
		return contexts[0]
	}

	// Initialize with values from the first context
	result := &models.ContextVector7D{
		Who:    contexts[0].Who,
		What:   contexts[0].What,
		When:   contexts[0].When,
		Where:  contexts[0].Where,
		Why:    contexts[0].Why,
		How:    contexts[0].How,
		Extent: contexts[0].Extent,
		Meta:   make(map[string]interface{}),
		Source: "merged",
	}

	// For non-numeric fields, use the most common value
	whoCount := make(map[string]int)
	whatCount := make(map[string]int)
	whereCount := make(map[string]int)
	whyCount := make(map[string]int)
	howCount := make(map[string]int)

	for _, ctx := range contexts {
		whoCount[ctx.Who]++
		whatCount[ctx.What]++
		whereCount[ctx.Where]++
		whyCount[ctx.Why]++
		howCount[ctx.How]++
	}

	// Find the most common values
	result.Who = findMostCommon(whoCount)
	result.What = findMostCommon(whatCount)
	result.Where = findMostCommon(whereCount)
	result.Why = findMostCommon(whyCount)
	result.How = findMostCommon(howCount)

	// For numeric fields, use weighted average
	var weightedWhen float64
	var weightedExtent float64
	var totalWeight float64

	// Use the extent as the weight
	for _, ctx := range contexts {
		weight := ctx.Extent
		if weight <= 0 {
			weight = 1.0
		}

		weightedWhen += float64(ctx.When) * weight
		weightedExtent += ctx.Extent * weight
		totalWeight += weight

		// Merge metadata
		for k, v := range ctx.Meta {
			// For numeric values, use weighted average
			if fv, ok := v.(float64); ok {
				if existing, exists := result.Meta[k]; exists {
					if efv, eok := existing.(float64); eok {
						result.Meta[k] = efv + (fv-efv)*weight/totalWeight
					} else {
						result.Meta[k] = fv
					}
				} else {
					result.Meta[k] = fv
				}
			} else {
				// For non-numeric values, prefer the first occurrence
				if _, exists := result.Meta[k]; !exists {
					result.Meta[k] = v
				}
			}
		}
	}

	// Apply the weighted averages
	if totalWeight > 0 {
		result.When = int64(weightedWhen / totalWeight)
		result.Extent = weightedExtent / totalWeight
	}

	return result
}

// findMostCommon returns the most common string from a count map
func findMostCommon(counts map[string]int) string {
	var mostCommon string
	var highestCount int

	for value, count := range counts {
		if count > highestCount {
			mostCommon = value
			highestCount = count
		}
	}

	return mostCommon
}

// FormatJSON formats a JSON object with indentation
func FormatJSON(obj interface{}) string {
	// WHO: JSONFormatter
	// WHAT: Format JSON data
	// WHEN: During data presentation
	// WHERE: System Layer 6 (Integration)
	// WHY: For human-readable output
	// HOW: Using JSON marshaling
	// EXTENT: Single JSON object

	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting JSON: %v", err)
	}
	return string(bytes)
}

// ConvertToContextVector7D converts a map to a ContextVector7D
func ConvertToContextVector7D(data map[string]interface{}) *models.ContextVector7D {
	// WHO: ContextConverter
	// WHAT: Convert map to context vector
	// WHEN: During context reconstruction
	// WHERE: System Layer 6 (Integration)
	// WHY: To standardize context format
	// HOW: Using map to struct conversion
	// EXTENT: Single context vector

	result := &models.ContextVector7D{
		Meta: make(map[string]interface{}),
	}

	// Extract the main dimensions
	if who, ok := data["who"].(string); ok {
		result.Who = who
	}

	if what, ok := data["what"].(string); ok {
		result.What = what
	}

	if when, ok := data["when"].(float64); ok {
		result.When = int64(when)
	} else if whenStr, ok := data["when"].(string); ok {
		if whenInt, err := strconv.ParseInt(whenStr, 10, 64); err == nil {
			result.When = whenInt
		}
	}

	if where, ok := data["where"].(string); ok {
		result.Where = where
	}

	if why, ok := data["why"].(string); ok {
		result.Why = why
	}

	if how, ok := data["how"].(string); ok {
		result.How = how
	}

	if extent, ok := data["extent"].(float64); ok {
		result.Extent = extent
	} else if extStr, ok := data["extent"].(string); ok {
		if extFloat, err := strconv.ParseFloat(extStr, 64); err == nil {
			result.Extent = extFloat
		}
	}

	if source, ok := data["source"].(string); ok {
		result.Source = source
	}

	// Extract metadata
	if meta, ok := data["meta"].(map[string]interface{}); ok {
		result.Meta = meta
	}

	return result
}

// ValidateContext validates if a context object has all required dimensions
func ValidateContext(context *models.ContextVector7D) error {
	// WHO: ContextValidator
	// WHAT: Validate context completeness
	// WHEN: During context validation
	// WHERE: System Layer 6 (Integration)
	// WHY: To ensure context quality
	// HOW: Using field validation
	// EXTENT: All context dimensions

	if context == nil {
		return fmt.Errorf("context cannot be nil")
	}

	if context.Who == "" {
		return fmt.Errorf("'who' dimension is required")
	}

	if context.What == "" {
		return fmt.Errorf("'what' dimension is required")
	}

	if context.When == 0 {
		return fmt.Errorf("'when' dimension is required")
	}

	if context.Where == "" {
		return fmt.Errorf("'where' dimension is required")
	}

	if context.Why == "" {
		return fmt.Errorf("'why' dimension is required")
	}

	if context.How == "" {
		return fmt.Errorf("'how' dimension is required")
	}

	// Extent can be zero, so no validation needed

	return nil
}
