/*
 * WHO: ContextTranslator
 * WHAT: Context translation for GitHub MCP server
 * WHEN: During context exchange between systems
 * WHERE: System Layer 6 (Integration)
 * WHY: To facilitate interoperability between GitHub and TNOS
 * HOW: Using 7D context mapping with compression-first logic
 * EXTENT: All cross-system context translation
 */

package translations

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log"
)

// ContextVector7D represents a 7D context vector
type ContextVector7D struct {
	// WHO: ContextManager
	// WHAT: 7D context structure
	// WHEN: During context operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide contextual awareness
	// HOW: Using 7D framework dimensions
	// EXTENT: All contextual data

	Who    string                 `json:"who"`
	What   string                 `json:"what"`
	When   int64                  `json:"when"`
	Where  string                 `json:"where"`
	Why    string                 `json:"why"`
	How    string                 `json:"how"`
	Extent float64                `json:"extent"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
	Source string                 `json:"source,omitempty"`
}

// GitHubContext represents a GitHub MCP context
type GitHubContext struct {
	// WHO: GitHubContextManager
	// WHAT: GitHub context structure
	// WHEN: During GitHub context operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To store GitHub context
	// HOW: Using GitHub format
	// EXTENT: GitHub context dimensions

	User      string                 `json:"user,omitempty"`
	Identity  string                 `json:"identity,omitempty"`
	Operation string                 `json:"operation,omitempty"`
	Type      string                 `json:"type,omitempty"`
	Purpose   string                 `json:"purpose,omitempty"`
	Scope     float64                `json:"scope,omitempty"`
	Timestamp int64                  `json:"timestamp,omitempty"`
	Source    string                 `json:"source,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// CompressionFactors represents the factors used in Möbius compression
type CompressionFactors struct {
	// WHO: CompressionManager
	// WHAT: Compression parameter structure
	// WHEN: During compression operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To configure compression
	// HOW: Using formula parameters
	// EXTENT: Compression parameter space

	B float64 // Base factor
	V float64 // Value factor
	I float64 // Intent factor
	G float64 // Growth factor
	F float64 // Flexibility factor
	E float64 // Energy factor
	T float64 // Time factor
}

// DefaultFactors provides default compression factors
var DefaultFactors = CompressionFactors{
	B: 0.8, // Base factor
	V: 0.7, // Value factor
	I: 0.9, // Intent factor
	G: 1.2, // Growth factor
	F: 0.6, // Flexibility factor
	E: 0.5, // Energy factor
	T: 0.0, // Time factor (will be calculated)
}

// NewContext creates a new 7D context vector
func NewContext(who, what, where, why, how string, extent float64) ContextVector7D {
	// WHO: ContextCreator
	// WHAT: Create 7D context
	// WHEN: During context initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To establish context
	// HOW: Using provided dimensions
	// EXTENT: New context instance

	now := time.Now().Unix()
	return ContextVector7D{
		Who:    who,
		What:   what,
		When:   now,
		Where:  where,
		Why:    why,
		How:    how,
		Extent: extent,
		Meta: map[string]interface{}{
			"B":         0.8, // Base factor
			"V":         0.7, // Value factor
			"I":         0.9, // Intent factor
			"G":         1.2, // Growth factor
			"F":         0.6, // Flexibility factor
			"createdAt": now,
		},
		Source: "github_mcp",
	}
}

// TranslateToMCP converts 7D context to MCP format
func (c *ContextVector7D) TranslateToMCP() map[string]interface{} {
	// WHO: FormatTranslator
	// WHAT: Convert to MCP format
	// WHEN: During message exchange
	// WHERE: System Layer 6 (Integration)
	// WHY: For protocol compatibility
	// HOW: Using format mapping
	// EXTENT: Single context

	return map[string]interface{}{
		"identity":  c.Who,
		"operation": c.What,
		"timestamp": c.When,
		"location":  c.Where,
		"purpose":   c.Why,
		"method":    c.How,
		"scope":     c.Extent,
		"metadata":  c.Meta,
		"source":    c.Source,
	}
}

// TranslateFromMCP converts MCP context to 7D context
func TranslateFromMCP(mcpContext map[string]interface{}) ContextVector7D {
	// WHO: FormatImporter
	// WHAT: Convert from MCP format
	// WHEN: During message reception
	// WHERE: System Layer 6 (Integration)
	// WHY: For internal compatibility
	// HOW: Using format mapping
	// EXTENT: Single context

	// Extract values with fallbacks
	who, _ := mcpContext["identity"].(string)
	what, _ := mcpContext["operation"].(string)
	where, _ := mcpContext["location"].(string)
	why, _ := mcpContext["purpose"].(string)
	how, _ := mcpContext["method"].(string)
	source, _ := mcpContext["source"].(string)

	// Handle numeric values
	var when int64
	var extent float64

	if whenVal, ok := mcpContext["timestamp"]; ok {
		switch v := whenVal.(type) {
		case int64:
			when = v
		case float64:
			when = int64(v)
		case int:
			when = int64(v)
		default:
			when = time.Now().Unix()
		}
	} else {
		when = time.Now().Unix()
	}

	if extentVal, ok := mcpContext["scope"]; ok {
		switch v := extentVal.(type) {
		case float64:
			extent = v
		case int:
			extent = float64(v)
		case int64:
			extent = float64(v)
		default:
			extent = 1.0
		}
	} else {
		extent = 1.0
	}

	// Extract metadata
	meta := make(map[string]interface{})
	if metaVal, ok := mcpContext["metadata"].(map[string]interface{}); ok {
		meta = metaVal
	}

	// Ensure compression factors exist
	ensureCompressionFactors(meta)

	return ContextVector7D{
		Who:    who,
		What:   what,
		When:   when,
		Where:  where,
		Why:    why,
		How:    how,
		Extent: extent,
		Meta:   meta,
		Source: source,
	}
}

// ensureCompressionFactors makes sure necessary compression factors exist
func ensureCompressionFactors(meta map[string]interface{}) {
	// WHO: CompressionSupport
	// WHAT: Ensure compression factors
	// WHEN: During context processing
	// WHERE: System Layer 6 (Integration)
	// WHY: For compression support
	// HOW: Using default values
	// EXTENT: Compression parameters

	// Define default factors if missing
	factors := map[string]float64{
		"B": 0.8, // Base factor
		"V": 0.7, // Value factor
		"I": 0.9, // Intent factor
		"G": 1.2, // Growth factor
		"F": 0.6, // Flexibility factor
	}

	// Ensure each factor exists
	for key, defaultVal := range factors {
		if _, exists := meta[key]; !exists {
			meta[key] = defaultVal
		}
	}
}

// Merge combines two context vectors with priority given to the newer one
func (c ContextVector7D) Merge(other ContextVector7D) ContextVector7D {
	// WHO: ContextMerger
	// WHAT: Combine contexts
	// WHEN: During context integration
	// WHERE: System Layer 6 (Integration)
	// WHY: For context synchronization
	// HOW: Using prioritized merging
	// EXTENT: Multiple contexts

	// Create a new context based on the more recent one
	var base, update ContextVector7D
	if c.When >= other.When {
		base = other
		update = c
	} else {
		base = c
		update = other
	}

	// Merge metadata
	mergedMeta := make(map[string]interface{})
	for k, v := range base.Meta {
		mergedMeta[k] = v
	}
	for k, v := range update.Meta {
		mergedMeta[k] = v
	}

	// Create merged context
	return ContextVector7D{
		Who:    update.Who,
		What:   update.What,
		When:   update.When,
		Where:  update.Where,
		Why:    update.Why,
		How:    update.How,
		Extent: update.Extent,
		Meta:   mergedMeta,
		Source: update.Source,
	}
}

// Compress applies the Möbius Compression Formula to the context
func (c *ContextVector7D) Compress() *ContextVector7D {
	// WHO: ContextCompressor
	// WHAT: Compress context
	// WHEN: During data optimization
	// WHERE: System Layer 6 (Integration)
	// WHY: For efficient data transfer
	// HOW: Using Möbius Compression Formula
	// EXTENT: Single context

	// Extract compression factors from meta
	var B, V, I, G, F float64 = 0.8, 0.7, 0.9, 1.2, 0.6

	if val, exists := c.Meta["B"].(float64); exists {
		B = val
	}
	if val, exists := c.Meta["V"].(float64); exists {
		V = val
	}
	if val, exists := c.Meta["I"].(float64); exists {
		I = val
	}
	if val, exists := c.Meta["G"].(float64); exists {
		G = val
	}
	if val, exists := c.Meta["F"].(float64); exists {
		F = val
	}

	// Calculate entropy based on context complexity
	entropy := calculateContextEntropy(c)

	// Calculate time factor (normalized seconds since context creation)
	var t float64 = 1.0
	if createdAt, exists := c.Meta["createdAt"].(int64); exists {
		elapsed := float64(time.Now().Unix() - createdAt)
		t = math.Min(elapsed/86400, 1.0) // Normalize to max 1.0 (1 day)
	}

	// Calculate energy factor
	E := 0.5 // Base energy cost

	// Calculate sum of contextual connections
	C_sum := 0.3 // Base connection value

	// Calculate alignment
	alignment := (B + V*I) * math.Exp(-t*E)

	// Apply Möbius Compression Formula
	compressed := (c.Extent * B * I * (1 - (entropy / math.Log2(1+V))) * (G + F)) /
		(E*t + C_sum*entropy + alignment)

	// Create new compressed context
	compressedContext := *c
	compressedContext.Extent = compressed

	// Store compression variables
	compressedContext.Meta["compressed"] = true
	compressedContext.Meta["originalExtent"] = c.Extent
	compressedContext.Meta["compressionRatio"] = c.Extent / compressed
	compressedContext.Meta["entropy"] = entropy
	compressedContext.Meta["alignment"] = alignment

	return &compressedContext
}

// Decompress restores a compressed context
func (c *ContextVector7D) Decompress() *ContextVector7D {
	// WHO: ContextDecompressor
	// WHAT: Decompress context
	// WHEN: During data restoration
	// WHERE: System Layer 6 (Integration)
	// WHY: For complete data access
	// HOW: Using stored parameters
	// EXTENT: Single context

	// Check if context is compressed
	compressed, ok := c.Meta["compressed"].(bool)
	if !ok || !compressed {
		return c
	}

	// Get original value
	original, ok := c.Meta["originalExtent"].(float64)
	if !ok {
		return c
	}

	// Create decompressed context
	decompressedContext := *c
	decompressedContext.Extent = original

	// Clean up compression metadata
	delete(decompressedContext.Meta, "compressed")
	delete(decompressedContext.Meta, "originalExtent")
	delete(decompressedContext.Meta, "compressionRatio")
	delete(decompressedContext.Meta, "entropy")
	delete(decompressedContext.Meta, "alignment")

	return &decompressedContext
}

// ToMap converts context to a map representation
func (c *ContextVector7D) ToMap() map[string]interface{} {
	// WHO: FormatConverter
	// WHAT: Convert to map
	// WHEN: During serialization
	// WHERE: System Layer 6 (Integration)
	// WHY: For general compatibility
	// HOW: Using map conversion
	// EXTENT: Single context

	return map[string]interface{}{
		"who":    c.Who,
		"what":   c.What,
		"when":   c.When,
		"where":  c.Where,
		"why":    c.Why,
		"how":    c.How,
		"extent": c.Extent,
		"meta":   c.Meta,
		"source": c.Source,
	}
}

// FromMap converts a map to a context vector
func FromMap(m map[string]interface{}) ContextVector7D {
	// WHO: MapImporter
	// WHAT: Convert from map
	// WHEN: During deserialization
	// WHERE: System Layer 6 (Integration)
	// WHY: For general compatibility
	// HOW: Using map extraction
	// EXTENT: Single context

	// Extract values with fallbacks
	who, _ := m["who"].(string)
	what, _ := m["what"].(string)
	where, _ := m["where"].(string)
	why, _ := m["why"].(string)
	how, _ := m["how"].(string)
	source, _ := m["source"].(string)

	// Handle numeric values
	var when int64
	var extent float64

	if whenVal, ok := m["when"]; ok {
		switch v := whenVal.(type) {
		case int64:
			when = v
		case float64:
			when = int64(v)
		case int:
			when = int64(v)
		default:
			when = time.Now().Unix()
		}
	} else {
		when = time.Now().Unix()
	}

	if extentVal, ok := m["extent"]; ok {
		switch v := extentVal.(type) {
		case float64:
			extent = v
		case int:
			extent = float64(v)
		case int64:
			extent = float64(v)
		default:
			extent = 1.0
		}
	} else {
		extent = 1.0
	}

	// Extract metadata
	meta := make(map[string]interface{})
	if metaVal, ok := m["meta"].(map[string]interface{}); ok {
		meta = metaVal
	}

	// Ensure compression factors exist
	ensureCompressionFactors(meta)

	return ContextVector7D{
		Who:    who,
		What:   what,
		When:   when,
		Where:  where,
		Why:    why,
		How:    how,
		Extent: extent,
		Meta:   meta,
		Source: source,
	}
}

// calculateContextEntropy computes context entropy for compression
func calculateContextEntropy(c *ContextVector7D) float64 {
	// WHO: EntropyCalculator
	// WHAT: Calculate context entropy
	// WHEN: During compression
	// WHERE: System Layer 6 (Integration)
	// WHY: For compression efficiency
	// HOW: Using information theory
	// EXTENT: Context complexity

	// Base entropy starts at 0.1
	entropy := 0.1

	// Add entropy for each dimension that has content
	if c.Who != "" {
		entropy += 0.1
	}
	if c.What != "" {
		entropy += 0.1
	}
	if c.Where != "" {
		entropy += 0.1
	}
	if c.Why != "" {
		entropy += 0.1
	}
	if c.How != "" {
		entropy += 0.1
	}

	// Add entropy based on meta complexity
	entropy += float64(len(c.Meta)) * 0.01

	// Normalize to range [0.1, 0.9]
	entropy = math.Min(0.9, math.Max(0.1, entropy))

	return entropy
}

// NewContextVector7D creates a new 7D context vector with default values
func NewContextVector7D(params map[string]interface{}) ContextVector7D {
	// WHO: ContextCreator
	// WHAT: Create 7D context vector
	// WHEN: During context initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To initialize context vector
	// HOW: Using provided parameters with defaults
	// EXTENT: 7D context creation

	now := time.Now().Unix()

	// Extract values with defaults
	who := getStringParam(params, "who", "System")
	what := getStringParam(params, "what", "Transform")
	when := getInt64Param(params, "when", now)
	where := getStringParam(params, "where", "MCP_Bridge")
	why := getStringParam(params, "why", "Protocol_Compliance")
	how := getStringParam(params, "how", "Context_Translation")
	extent := getFloat64Param(params, "extent", 1.0)
	source := getStringParam(params, "source", "github_mcp")

	// Create context vector
	return ContextVector7D{
		Who:    who,
		What:   what,
		When:   when,
		Where:  where,
		Why:    why,
		How:    how,
		Extent: extent,
		Source: source,
		Meta: map[string]interface{}{
			"B":         0.8, // Base factor
			"V":         0.7, // Value factor
			"I":         0.9, // Intent factor
			"G":         1.2, // Growth factor
			"F":         0.6, // Flexibility factor
			"createdAt": now,
		},
	}
}

// Helpers for parameter extraction
func getStringParam(params map[string]interface{}, key, defaultValue string) string {
	// WHO: ParamExtractor
	// WHAT: Extract string parameter
	// WHEN: During parameter processing
	// WHERE: System Layer 6 (Integration)
	// WHY: To safely extract parameters
	// HOW: Using type assertion
	// EXTENT: Single parameter extraction

	if val, ok := params[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getInt64Param(params map[string]interface{}, key string, defaultValue int64) int64 {
	// WHO: ParamExtractor
	// WHAT: Extract int64 parameter
	// WHEN: During parameter processing
	// WHERE: System Layer 6 (Integration)
	// WHY: To safely extract parameters
	// HOW: Using type assertion
	// EXTENT: Single parameter extraction

	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case int:
			return int64(v)
		case int32:
			return int64(v)
		case int64:
			return v
		case float32:
			return int64(v)
		case float64:
			return int64(v)
		}
	}
	return defaultValue
}

func getFloat64Param(params map[string]interface{}, key string, defaultValue float64) float64 {
	// WHO: ParamExtractor
	// WHAT: Extract float64 parameter
	// WHEN: During parameter processing
	// WHERE: System Layer 6 (Integration)
	// WHY: To safely extract parameters
	// HOW: Using type assertion
	// EXTENT: Single parameter extraction

	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case int:
			return float64(v)
		case int32:
			return float64(v)
		case int64:
			return float64(v)
		case float32:
			return float64(v)
		case float64:
			return v
		}
	}
	return defaultValue
}

// TranslateGitHubToTNOS7D translates GitHub context to TNOS 7D
func TranslateGitHubToTNOS7D(githubContext GitHubContext, logger *log.Logger) ContextVector7D {
	// WHO: GitHubToTNOSTranslator
	// WHAT: Translate GitHub to TNOS context
	// WHEN: During GitHub-to-TNOS operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To convert context formats
	// HOW: Using context mapping
	// EXTENT: Context translation operation

	params := map[string]interface{}{
		"who":    githubContext.User,
		"what":   githubContext.Type,
		"when":   githubContext.Timestamp,
		"where":  "MCP_Bridge",
		"why":    githubContext.Purpose,
		"how":    githubContext.Operation,
		"extent": githubContext.Scope,
		"source": "github_mcp",
	}

	// Create TNOS 7D context
	context := NewContextVector7D(params)

	// Transfer metadata
	if githubContext.Metadata != nil {
		for k, v := range githubContext.Metadata {
			context.Meta[k] = v
		}
	}

	if logger != nil {
		logger.Debug("Translated GitHub context to TNOS 7D",
			"github_user", githubContext.User,
			"tnos_who", context.Who)
	}

	return context
}

// TranslateTNOS7DToGitHub translates TNOS 7D context to GitHub
func TranslateTNOS7DToGitHub(tnos7D ContextVector7D, logger *log.Logger) GitHubContext {
	// WHO: TNOSToGitHubTranslator
	// WHAT: Translate TNOS to GitHub context
	// WHEN: During TNOS-to-GitHub operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To convert context formats
	// HOW: Using context mapping
	// EXTENT: Context translation operation

	// Create GitHub context
	githubContext := GitHubContext{
		User:      tnos7D.Who,
		Identity:  tnos7D.Who,
		Operation: tnos7D.How,
		Type:      tnos7D.What,
		Purpose:   tnos7D.Why,
		Scope:     tnos7D.Extent,
		Timestamp: tnos7D.When,
		Source:    "tnos_mcp",
		Metadata:  make(map[string]interface{}),
	}

	// Transfer metadata
	if tnos7D.Meta != nil {
		for k, v := range tnos7D.Meta {
			githubContext.Metadata[k] = v
		}
	}

	// Add translation metadata
	githubContext.Metadata["translatedAt"] = time.Now().Unix()
	githubContext.Metadata["translationMethod"] = "direct_mapping"

	if logger != nil {
		logger.Debug("Translated TNOS 7D context to GitHub",
			"tnos_who", tnos7D.Who,
			"github_user", githubContext.User)
	}

	return githubContext
}

// BridgeMCPContext bridges MCP contexts between systems with context awareness
func BridgeMCPContext(githubContext GitHubContext, existingContext *ContextVector7D, logger *log.Logger) ContextVector7D {
	// WHO: ContextBridger
	// WHAT: Bridge contexts between systems
	// WHEN: During cross-system operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide unified context
	// HOW: Using context bridging algorithm
	// EXTENT: Complete context bridge operation

	// Convert to TNOS 7D
	context := TranslateGitHubToTNOS7D(githubContext, logger)

	// If existing context is provided, merge them
	if existingContext != nil {
		context = MergeContexts(context, *existingContext, logger)
	}

	// Apply compression
	compressedContext := CompressContext(context, logger)

	if logger != nil {
		logger.Debug("Bridged MCP contexts",
			"source", githubContext.Source,
			"destination", "tnos_7d")
	}

	return compressedContext
}

// MergeContexts merges two 7D contexts with weighting
func MergeContexts(primary ContextVector7D, secondary ContextVector7D, logger *log.Logger) ContextVector7D {
	// WHO: ContextMerger
	// WHAT: Merge multiple contexts
	// WHEN: During context combination
	// WHERE: System Layer 6 (Integration)
	// WHY: To combine context information
	// HOW: Using weighted merging
	// EXTENT: Context merge operation

	// Default to primary values but use most recent timestamp
	result := primary
	result.When = max(primary.When, secondary.When)

	// Record merge in metadata
	if result.Meta == nil {
		result.Meta = make(map[string]interface{})
	}

	result.Meta["mergedAt"] = time.Now().Unix()
	result.Meta["mergedFrom"] = []string{primary.Source, secondary.Source}

	if logger != nil {
		logger.Debug("Merged contexts",
			"primary_source", primary.Source,
			"secondary_source", secondary.Source)
	}

	return result
}

// CompressContext compresses a 7D context using Möbius formula
func CompressContext(context ContextVector7D, logger *log.Logger) ContextVector7D {
	// WHO: ContextCompressor
	// WHAT: Compress context data
	// WHEN: During context optimization
	// WHERE: System Layer 6 (Integration)
	// WHY: To optimize context storage
	// HOW: Using Möbius compression
	// EXTENT: Context compression operation

	if context.Meta == nil {
		context.Meta = make(map[string]interface{})
	}

	// Extract contextual factors
	B := getMetaFloat(context.Meta, "B", 0.8) // Base factor
	V := getMetaFloat(context.Meta, "V", 0.7) // Value factor
	I := getMetaFloat(context.Meta, "I", 0.9) // Intent factor
	G := getMetaFloat(context.Meta, "G", 1.2) // Growth factor
	F := getMetaFloat(context.Meta, "F", 0.6) // Flexibility factor

	// Calculate time factor (how "fresh" the context is)
	now := time.Now().Unix()
	t := float64(now-context.When) / 86400.0 // Days
	if t < 0 {
		t = 0
	}

	// Calculate energy factor (computational cost)
	E := 0.5

	// Calculate context entropy (simplified)
	contextBytes, _ := json.Marshal(context)
	entropy := float64(len(contextBytes)) / 100.0

	// Apply Möbius compression formula
	alignment := (B + V*I) * math.Exp(-t*E)
	compressionFactor := (B * I * (1.0 - (entropy / math.Log2(1.0+V))) * (G + F)) /
		(E*t + entropy + alignment)

	// Create compressed context (same structure but with compression metadata)
	result := context

	// Add compression metadata
	result.Meta["compressionFactor"] = compressionFactor
	result.Meta["compressedAt"] = now
	result.Meta["entropy"] = entropy
	result.Meta["alignment"] = alignment
	result.Meta["t"] = t
	result.Meta["E"] = E

	if logger != nil {
		logger.Debug("Compressed context",
			"entropy", entropy,
			"compressionFactor", compressionFactor)
	}

	return result
}

// Helper to extract float from metadata
func getMetaFloat(meta map[string]interface{}, key string, defaultValue float64) float64 {
	if meta == nil {
		return defaultValue
	}

	if val, ok := meta[key]; ok {
		switch v := val.(type) {
		case int:
			return float64(v)
		case int32:
			return float64(v)
		case int64:
			return float64(v)
		case float32:
			return float64(v)
		case float64:
			return v
		}
	}

	return defaultValue
}

// Helper to get max of two int64s
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// DecompressContext decompresses a compressed context
func DecompressContext(compressedContext ContextVector7D, logger *log.Logger) ContextVector7D {
	// WHO: ContextDecompressor
	// WHAT: Decompress context data
	// WHEN: During context restoration
	// WHERE: System Layer 6 (Integration)
	// WHY: To restore full context
	// HOW: Using compression metadata
	// EXTENT: Context decompression operation

	// In a real implementation, this would use the compression metadata
	// to restore the original context. For this placeholder, we simply
	// return the original context with a note that it was "decompressed".

	result := compressedContext
	if result.Meta == nil {
		result.Meta = make(map[string]interface{})
	}
	result.Meta["decompressedAt"] = time.Now().Unix()

	if logger != nil {
		logger.Debug("Decompressed context",
			"who", result.Who,
			"what", result.What)
	}

	return result
}

// CalculateContextEntropy calculates entropy of a context vector
func CalculateContextEntropy(context ContextVector7D) float64 {
	// WHO: EntropyCalculator
	// WHAT: Calculate context entropy
	// WHEN: During compression operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To measure context complexity
	// HOW: Using Shannon entropy formula
	// EXTENT: Context entropy calculation

	// Convert context to bytes
	data, err := json.Marshal(context)
	if err != nil {
		return 1.0 // Default on error
	}

	// Count character frequencies
	freqMap := make(map[byte]int)
	for _, b := range data {
		freqMap[b]++
	}

	// Calculate Shannon entropy
	var entropy float64
	length := float64(len(data))

	for _, count := range freqMap {
		prob := float64(count) / length
		entropy -= prob * math.Log2(prob)
	}

	return entropy
}

// ConvertJSONToContext converts a JSON representation to a 7D context
func ConvertJSONToContext(jsonData []byte, logger *log.Logger) (ContextVector7D, error) {
	// WHO: JSONConverter
	// WHAT: Convert JSON to context
	// WHEN: During data parsing
	// WHERE: System Layer 6 (Integration)
	// WHY: To transform external data
	// HOW: Using JSON unmarshaling
	// EXTENT: JSON conversion operation

	var context ContextVector7D
	err := json.Unmarshal(jsonData, &context)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to convert JSON to context", "error", err.Error())
		}
		return ContextVector7D{}, fmt.Errorf("failed to parse context: %w", err)
	}

	// Ensure default values for missing fields
	if context.Who == "" {
		context.Who = "System"
	}

	if context.What == "" {
		context.What = "Transform"
	}

	if context.When == 0 {
		context.When = time.Now().Unix()
	}

	if context.Where == "" {
		context.Where = "MCP_Bridge"
	}

	if context.Why == "" {
		context.Why = "Protocol_Compliance"
	}

	if context.How == "" {
		context.How = "Context_Translation"
	}

	if context.Meta == nil {
		context.Meta = make(map[string]interface{})
	}

	if logger != nil {
		logger.Debug("Converted JSON to context", "who", context.Who)
	}

	return context, nil
}

// WHO: ContextTranslator
// WHAT: Context translation for GitHub MCP server
// WHEN: During context exchange between systems
// WHERE: System Layer 6 (Integration)
// WHY: To facilitate interoperability between GitHub and TNOS
// HOW: Using 7D context mapping with compression-first logic
// EXTENT: All cross-system context translation

// ContextTranslatorFunc defines a standard interface for context translation functions
// WHO: TranslationFunctionDefiner
// WHAT: Define standard translation function signature
// WHEN: During type declarations
// WHERE: System Layer 6 (Integration)
// WHY: To standardize context translation interfaces
// HOW: Using function type definition
// EXTENT: All context translation operations
type ContextTranslatorFunc func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)

// GitHubContextTranslator translates GitHub context to TNOS context
// WHO: GitHubToTNOSTranslator
// WHAT: Translate GitHub context to TNOS format
// WHEN: During inbound context translation
// WHERE: System Layer 6 (Integration)
// WHY: To provide TNOS compatibility
// HOW: Using structured context mapping
// EXTENT: All inbound context operations
func GitHubContextTranslator(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Extract the 7D context from the GitHub context
	cv := MCPContextToTNOS(input)

	// Apply compression to optimize the context
	cv = CompressContext(cv, nil) // Pass nil logger as it's optional

	// Store the 7D context in the Go context
	ctx = ContextWithVector(ctx, cv)

	// Convert back to map for further processing
	result := TNOSContextToMCP(cv)

	return result, nil
}

// TNOSContextTranslator translates TNOS context to GitHub context
// WHO: TNOSToGitHubTranslator
// WHAT: Translate TNOS context to GitHub format
// WHEN: During outbound context translation
// WHERE: System Layer 6 (Integration)
// WHY: To provide GitHub compatibility
// HOW: Using structured context mapping
// EXTENT: All outbound context operations
func TNOSContextTranslator(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Extract existing context if available
	var cv ContextVector7D
	var exists bool

	if cv, exists = VectorFromContext(ctx); !exists {
		// Create a new context from the input
		ifaceMap := make(map[string]interface{})
		for k, v := range input {
			ifaceMap[k] = v
		}
		cv = FromMap(ifaceMap)
	}

	// Apply compression per TNOS guidelines
	cv = CompressContext(cv, nil) // Pass nil logger as it's optional

	// Convert to GitHub context format
	result := map[string]interface{}{
		"identity":  cv.Who,
		"operation": cv.What,
		"timestamp": cv.When,
		"purpose":   cv.Why,
		"scope":     cv.Extent,
		// Store full TNOS context in a special field
		"_tnos_context": cv.ToMap(),
	}

	return result, nil
}

// ContextParseError represents an error parsing context
// WHO: ErrorDefiner
// WHAT: Define context parsing error type
// WHEN: During error handling
// WHERE: System Layer 6 (Integration)
// WHY: To standardize error reporting
// HOW: Using error type definition
// EXTENT: All context parsing operations
type ContextParseError struct {
	Dimension string
	Value     interface{}
	Reason    string
}

// Error returns the error message string
func (e ContextParseError) Error() string {
	return fmt.Sprintf("unable to parse context dimension %s with value %v: %s",
		e.Dimension, e.Value, e.Reason)
}

// ValidateContextDimension checks if a context dimension is valid
// WHO: ContextValidator
// WHAT: Validate context dimension
// WHEN: During context validation
// WHERE: System Layer 6 (Integration)
// WHY: To ensure context integrity
// HOW: Using type checking and validation
// EXTENT: All context validation operations
func ValidateContextDimension(dim string, value interface{}) error {
	// Validate based on dimension
	switch dim {
	case "who", "what", "where", "why", "how", "extent":
		// These should be strings
		if _, ok := value.(string); !ok {
			return ContextParseError{
				Dimension: dim,
				Value:     value,
				Reason:    "expected string value",
			}
		}
	case "when":
		// Should be a valid time string
		if strVal, ok := value.(string); !ok {
			return ContextParseError{
				Dimension: dim,
				Value:     value,
				Reason:    "expected string value for timestamp",
			}
		} else if _, err := time.Parse(time.RFC3339, strVal); err != nil {
			return ContextParseError{
				Dimension: dim,
				Value:     value,
				Reason:    "invalid timestamp format, expected RFC3339",
			}
		}
	}

	return nil
}

// CreateContextHelperFunc creates a context helper function
// WHO: HelperFunctionFactory
// WHAT: Create context helper function
// WHEN: During system initialization
// WHERE: System Layer 6 (Integration)
// WHY: To provide context assistance
// HOW: Using function factory pattern
// EXTENT: All context helper operations
func CreateContextHelperFunc(translator ContextTranslatorFunc) func(context.Context, map[string]interface{}) (context.Context, map[string]interface{}, error) {
	return func(ctx context.Context, input map[string]interface{}) (context.Context, map[string]interface{}, error) {
		// Translate the context
		output, err := translator(ctx, input)
		if err != nil {
			return ctx, nil, err
		}

		// Extract 7D context if available
		var cv ContextVector7D
		var exists bool

		if cv, exists = VectorFromContext(ctx); !exists {
			// Create a new context from the output
			ifaceMap := make(map[string]interface{})
			for k, v := range output {
				ifaceMap[k] = v
			}
			cv = FromMap(ifaceMap)

			// Store it in the context
			ctx = ContextWithVector(ctx, cv)
		}

		return ctx, output, nil
	}
}

// Key type for storing ContextVector7D in context.Context
// WHO: ContextKeyDefiner
// WHAT: Define context key type
// WHEN: During context operations
// WHERE: System Layer 6 (Integration)
// WHY: To safely store and retrieve context
// HOW: Using context key pattern
// EXTENT: All context storage/retrieval operations
type contextKey int

// Define the key for storing ContextVector7D
// WHO: ContextKeyAssigner
// WHAT: Assign context key value
// WHEN: During context operations
// WHERE: System Layer 6 (Integration)
// WHY: To uniquely identify context data
// HOW: Using key constant
// EXTENT: All context storage/retrieval operations
const (
	contextVector7DKey contextKey = iota
)

// ContextWithVector adds a 7D Context Vector to a Go context
// WHO: ContextEnricher
// WHAT: Add context vector to Go context
// WHEN: During context enrichment
// WHERE: System Layer 6 (Integration)
// WHY: To propagate context information
// HOW: Using context.WithValue
// EXTENT: All context propagation operations
func ContextWithVector(ctx context.Context, vector ContextVector7D) context.Context {
	return context.WithValue(ctx, contextVector7DKey, vector)
}

// VectorFromContext extracts a 7D Context Vector from a Go context
// WHO: ContextExtractor
// WHAT: Extract context vector from Go context
// WHEN: During context retrieval
// WHERE: System Layer 6 (Integration)
// WHY: To access context information
// HOW: Using context.Value with type assertion
// EXTENT: All context access operations
func VectorFromContext(ctx context.Context) (ContextVector7D, bool) {
	value := ctx.Value(contextVector7DKey)
	if value == nil {
		return ContextVector7D{}, false
	}

	if vector, ok := value.(ContextVector7D); ok {
		return vector, true
	}

	return ContextVector7D{}, false
}

// MCPContextToTNOS converts MCP context format to TNOS 7D context
// WHO: FormatConverter
// WHAT: Convert MCP to TNOS format
// WHEN: During inbound translation
// WHERE: System Layer 6 (Integration)
// WHY: For internal processing
// HOW: Using standardized mapping
// EXTENT: All inbound context translations
func MCPContextToTNOS(mcpContext map[string]interface{}) ContextVector7D {
	// Extract values with fallbacks for MCP format
	who, _ := mcpContext["identity"].(string)
	what, _ := mcpContext["operation"].(string)
	where, _ := mcpContext["location"].(string)
	why, _ := mcpContext["purpose"].(string)
	how, _ := mcpContext["method"].(string)
	source, _ := mcpContext["source"].(string)

	// Handle numeric values
	var when int64
	var extent float64

	if whenVal, ok := mcpContext["timestamp"]; ok {
		switch v := whenVal.(type) {
		case int64:
			when = v
		case float64:
			when = int64(v)
		case int:
			when = int64(v)
		default:
			when = time.Now().Unix()
		}
	} else {
		when = time.Now().Unix()
	}

	if extentVal, ok := mcpContext["scope"]; ok {
		switch v := extentVal.(type) {
		case float64:
			extent = v
		case int:
			extent = float64(v)
		case int64:
			extent = float64(v)
		default:
			extent = 1.0
		}
	} else {
		extent = 1.0
	}

	// Extract metadata
	meta := make(map[string]interface{})
	if metaVal, ok := mcpContext["metadata"].(map[string]interface{}); ok {
		meta = metaVal
	}

	// Ensure compression factors exist
	ensureCompressionFactors(meta)

	// Check if there's a stored _tnos_context for complete roundtrip
	if tnosCtx, ok := mcpContext["_tnos_context"].(map[string]interface{}); ok {
		// We have a full TNOS context stored, use that directly
		return FromMap(tnosCtx)
	}

	return ContextVector7D{
		Who:    who,
		What:   what,
		When:   when,
		Where:  where,
		Why:    why,
		How:    how,
		Extent: extent,
		Meta:   meta,
		Source: source,
	}
}

// TNOSContextToMCP converts TNOS 7D context to MCP format
// WHO: FormatConverter
// WHAT: Convert TNOS to MCP format
// WHEN: During outbound translation
// WHERE: System Layer 6 (Integration)
// WHY: For external compatibility
// HOW: Using standardized mapping
// EXTENT: All outbound context translations
func TNOSContextToMCP(tnos7D ContextVector7D) map[string]interface{} {
	// Convert to the standard MCP format
	mcpContext := map[string]interface{}{
		"identity":  tnos7D.Who,
		"operation": tnos7D.What,
		"timestamp": tnos7D.When,
		"location":  tnos7D.Where,
		"purpose":   tnos7D.Why,
		"method":    tnos7D.How,
		"scope":     tnos7D.Extent,
		"metadata":  tnos7D.Meta,
		"source":    tnos7D.Source,
		// Store the complete 7D context for roundtrip preservation
		"_tnos_context": tnos7D.ToMap(),
	}

	return mcpContext
}

// WHO: ContextManager
// WHAT: Create context with full 7D awareness
// WHEN: During system operations
// WHERE: System Layer 6 (Integration)
// WHY: To provide comprehensive context
// HOW: Using Go context with values
// EXTENT: All context operations
func CreateContext7D(baseCtx context.Context, who, what, where, why, how string, extent float64) context.Context {
	// Create a new context vector
	cv := NewContext(who, what, where, why, how, extent)

	// Apply compression as per TNOS guidelines
	cv = CompressContext(cv, nil)

	// Store in Go context
	return ContextWithVector(baseCtx, cv)
}

// WHO: ContextMigrator
// WHAT: Migrate context between systems
// WHEN: During system interactions
// WHERE: System Layer 6 (Integration)
// WHY: To preserve context across boundaries
// HOW: Using serialization/deserialization
// EXTENT: All cross-system operations
func MigrateContext(baseCtx context.Context, serializedContext string, logger *log.Logger) (context.Context, error) {
	// Parse JSON context
	var contextMap map[string]interface{}
	err := json.Unmarshal([]byte(serializedContext), &contextMap)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to unmarshal context", "error", err.Error())
		}
		return baseCtx, err
	}

	// Convert to 7D context
	cv := MCPContextToTNOS(contextMap)

	// Store in Go context
	return ContextWithVector(baseCtx, cv), nil
}
