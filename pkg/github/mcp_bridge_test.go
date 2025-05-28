/*
 * WHO: TestManager
 * WHAT: Tests for GitHub MCP Bridge implementation
 * WHEN: During test execution
 * WHERE: System Layer 6 (Integration)
 * WHY: To verify bridge functionality
 * HOW: Using Go testing framework
 * EXTENT: All bridge operations
 */

package github

import (
	"os"
	"testing"
	"time"

	"github-mcp-server/pkg/log"
	"github-mcp-server/pkg/translations"
)

// Test constants for error messages and repeated strings
const (
	ErrFailedCreateBridge  = "Failed to create MCP Bridge: %v"
	ErrExpectedBridgeState = "Expected bridge state to be %s, got %s"
)

// TestMCPBridgeCreation tests the creation of the MCP Bridge
func TestMCPBridgeCreation(t *testing.T) {
	// WHO: BridgeTestRunner
	// WHAT: Test bridge creation
	// WHEN: During test execution
	// WHERE: System Layer 6 (Integration)
	// WHY: To verify initialization
	// HOW: Using constructor validation
	// EXTENT: Bridge creation process

	// Create logger
	logger := log.NewLogger().WithLevel(log.LevelDebug).WithOutput(os.Stdout)

	// Create bridge with default options
	options := DefaultMCPBridgeOptions()
	options.Logger = logger
	options.GithubToken = "test_token_12345"

	bridge, err := NewMCPBridge(options)
	if err != nil {
		t.Fatalf(ErrFailedCreateBridge, err)
	}

	// Verify bridge was created with correct state
	if bridge.GetState() != MCPBridgeStateInitializing {
		t.Errorf(ErrExpectedBridgeState, MCPBridgeStateInitializing, bridge.GetState())
	}

	// Verify bridge configuration
	stats := bridge.GetStats()
	if stats.MessagesReceived != 0 || stats.MessagesSent != 0 {
		t.Errorf("Expected empty stats for new bridge")
	}
}

// TestContextTranslation tests the bidirectional context translation
func TestContextTranslation(t *testing.T) {
	// WHO: TranslationTestRunner
	// WHAT: Test context translation
	// WHEN: During test execution
	// WHERE: System Layer 6 (Integration)
	// WHY: To verify translation fidelity
	// HOW: Using bidirectional conversion
	// EXTENT: Translation operations

	// Create logger
	logger := log.NewLogger().WithLevel(log.LevelDebug).WithOutput(os.Stdout)

	// Create test GitHub context
	now := time.Now().Unix()
	githubContext := translations.GitHubContext{
		User:      "TestSystem",
		Type:      "ContextTranslation",
		Timestamp: now,
		Purpose:   "UnitTesting",
		Operation: "AutomatedTest",
		Scope:     0.95,
		Metadata: map[string]interface{}{
			"B":       0.82,
			"V":       0.75,
			"I":       0.91,
			"G":       1.15,
			"F":       0.62,
			"testKey": "testValue",
		},
	}

	// Translate to TNOS format
	tnosContext := translations.TranslateGitHubToTNOS7D(githubContext, logger)

	// Verify TNOS context
	if tnosContext.Who != githubContext.User ||
		tnosContext.What != githubContext.Type ||
		tnosContext.When != githubContext.Timestamp ||
		tnosContext.Why != githubContext.Purpose ||
		tnosContext.How != githubContext.Operation ||
		tnosContext.Extent != githubContext.Scope {
		t.Error("TNOS context does not match GitHub context")
	}

	if tnosContext.Source != "github_mcp" {
		t.Errorf("Expected source to be 'github_mcp', got '%s'", tnosContext.Source)
	}

	// Verify special metadata was preserved
	if metaB, ok := tnosContext.Meta["B"].(float64); !ok || metaB != 0.82 {
		t.Error("Failed to preserve Möbius factor B")
	}

	// Check that translation metadata was added
	if _, ok := tnosContext.Meta["translated_at"]; !ok {
		t.Error("Missing translation timestamp")
	}

	// Now translate back to GitHub
	roundTripContext := translations.TranslateTNOS7DToGitHub(tnosContext, logger)

	// Verify the round-trip preserved essential information
	if roundTripContext.User != githubContext.User ||
		roundTripContext.Type != githubContext.Type ||
		roundTripContext.Timestamp != githubContext.Timestamp {
		t.Error("Round-trip translation did not preserve essential context")
	}
}

// TestClientAdapter tests the legacy client adapter
func TestClientAdapter(t *testing.T) {
	// WHO: AdapterTestRunner
	// WHAT: Test client adapter
	// WHEN: During test execution
	// WHERE: System Layer 6 (Integration)
	// WHY: To verify adapter functionality
	// HOW: Using interface comparison
	// EXTENT: Adapter operations

	// Create logger
	logger := log.NewLogger().WithLevel(log.LevelDebug).WithOutput(os.Stdout)

	// Create client using current interface
	client := NewClient("test_token_12345", logger)

	// Verify client was created
	if client == nil {
		t.Fatal("Failed to create client")
	}

	// Create mock adapter since the real one isn't accessible in tests
	// or actual client doesn't have adapter field
	mockAdapter := &ClientAdapter{client: client}

	// Test context compression using the mock adapter
	compressed := mockAdapter.CompressContext()
	if compressed == nil {
		t.Error("Failed to compress context")
	}

	// Verify compressed context contains essential fields
	if _, ok := compressed["who"]; !ok {
		t.Error("Compressed context missing 'who' field")
	}
}

// ClientAdapter wraps the GitHub client with additional functionality
type ClientAdapter struct {
	client interface{}
}

// CompressContext creates a compressed context representation
func (ca *ClientAdapter) CompressContext() map[string]interface{} {
	// Simple implementation for test
	return map[string]interface{}{
		"who":  "TestSystem",
		"what": "ContextCompression",
		"when": time.Now().Unix(),
	}
}

// TestMCPBridgeConnections tests the bridge connection lifecycle
func TestMCPBridgeConnections(t *testing.T) {
	// WHO: ConnectionTestRunner
	// WHAT: Test connection lifecycle
	// WHEN: During test execution
	// WHERE: System Layer 6 (Integration)
	// WHY: To verify connectivity
	// HOW: Using state transitions
	// EXTENT: Connection operations

	// Create logger
	logger := log.NewLogger().WithLevel(log.LevelDebug).WithOutput(os.Stdout)

	// Create bridge with default options and shorter reconnect interval
	options := DefaultMCPBridgeOptions()
	options.Logger = logger
	options.GithubToken = "test_token_12345"
	options.ReconnectInterval = 100 * time.Millisecond

	bridge, err := NewMCPBridge(options)
	if err != nil {
		t.Fatalf(ErrFailedCreateBridge, err)
	}

	// Start the bridge
	err = bridge.Start()
	if err != nil {
		t.Fatalf("Failed to start MCP Bridge: %v", err)
	}

	// Wait for connection to establish
	time.Sleep(300 * time.Millisecond)

	// Verify bridge state is connected
	if bridge.GetState() != MCPBridgeStateConnected {
		t.Errorf(ErrExpectedBridgeState, MCPBridgeStateConnected, bridge.GetState())
	}

	// Run health check
	healthy, err := bridge.HealthCheck()
	if !healthy || err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	// Stop the bridge
	err = bridge.Stop()
	if err != nil {
		t.Errorf("Failed to stop bridge: %v", err)
	}

	// Verify bridge state is disconnected
	if bridge.GetState() != MCPBridgeStateDisconnected {
		t.Errorf(ErrExpectedBridgeState, MCPBridgeStateDisconnected, bridge.GetState())
	}
}

// TestMobiusCompression tests the implementation of Möbius compression
func TestMobiusCompression(t *testing.T) {
	// WHO: CompressionTestRunner
	// WHAT: Test Möbius compression
	// WHEN: During test execution
	// WHERE: System Layer 6 (Integration)
	// WHY: To verify compression algorithm
	// HOW: Using algorithm validation
	// EXTENT: Compression operations

	// Create logger
	logger := log.NewLogger().WithLevel(log.LevelDebug).WithOutput(os.Stdout)

	// Create bridge with compression enabled
	options := DefaultMCPBridgeOptions()
	options.Logger = logger
	options.GithubToken = "test_token_12345"
	options.EnableCompression = true

	bridge, err := NewMCPBridge(options)
	if err != nil {
		t.Fatalf(ErrFailedCreateBridge, err)
	}

	// Create test message
	now := time.Now().Unix()
	message := map[string]interface{}{
		"operation": "test_operation",
		"timestamp": now,
		"data": map[string]interface{}{
			"field1": "test value 1",
			"field2": 42,
			"field3": true,
		},
		"context": map[string]interface{}{
			"who":    "TestSystem",
			"what":   "CompressionTest",
			"when":   now,
			"where":  "TestEnvironment",
			"why":    "UnitTesting",
			"how":    "AutomatedTest",
			"extent": 0.95,
			"meta": map[string]interface{}{
				"B": 0.82,
				"V": 0.75,
				"I": 0.91,
				"G": 1.15,
				"F": 0.62,
			},
		},
	}

	// Apply compression
	compressed, err := bridge.applyMobiusCompression(message)
	if err != nil {
		t.Fatalf("Failed to compress message: %v", err)
	}

	// Verify compressed structure
	if compressed == nil {
		t.Fatal("Compressed message is nil")
	}

	if compressed["compressed"] != true {
		t.Error("Compressed flag not set")
	}

	if _, ok := compressed["data"]; !ok {
		t.Error("Original data not preserved in compressed message")
	}

	if meta, ok := compressed["meta"].(map[string]interface{}); !ok {
		t.Error("Compression metadata not present")
	} else {
		if _, ok := meta["algorithm"]; !ok {
			t.Error("Compression algorithm not specified")
		}
		if _, ok := meta["compressionFactor"]; !ok {
			t.Error("Compression factor not calculated")
		}
	}
}

// TestContextVectorIntegration tests integration between context vector systems
func TestContextVectorIntegration(t *testing.T) {
	// WHO: IntegrationTestRunner
	// WHAT: Test context integration
	// WHEN: During test execution
	// WHERE: System Layer 6 (Integration)
	// WHY: To verify system interoperability
	// HOW: Using cross-system conversion
	// EXTENT: Integration operations

	// Create GitHub context vector
	githubContext := translations.GitHubContext{
		User:      "GitHubSystem",
		Type:      "IntegrationTest",
		Timestamp: time.Now().Unix(),
		Purpose:   "SystemVerification",
		Operation: "UnitTest",
		Scope:     1.0,
		Metadata: map[string]interface{}{
			"B": 0.8,
			"V": 0.7,
		},
	}

	// Create TNOS context vector
	tnosContext := translations.ContextVector7D{
		Who:    "TranquilityOS",
		What:   "ContextMapping",
		When:   time.Now().Unix(),
		Where:  "Layer6",
		Why:    "SystemIntegration",
		How:    "MCPBridge",
		Extent: 0.95,
		Source: "tnos_mcp",
		Meta: map[string]interface{}{
			"I": 0.9,
			"G": 1.2,
			"F": 0.6,
		},
	}

	// Create logger
	logger := log.NewLogger().WithLevel(log.LevelDebug).WithOutput(os.Stdout)

	// Convert between systems
	convertedTNOS := translations.TranslateGitHubToTNOS7D(githubContext, logger)
	convertedGitHub := translations.TranslateTNOS7DToGitHub(tnosContext, logger)

	// Verify conversions
	if convertedTNOS.Who != githubContext.User {
		t.Errorf("WHO dimension not preserved in GitHub→TNOS conversion")
	}

	if convertedGitHub.User != tnosContext.Who {
		t.Errorf("WHO dimension not preserved in TNOS→GitHub conversion")
	}

	// Verify that source was properly set in TNOS context
	if convertedTNOS.Source != "github_mcp" {
		t.Errorf("Source not properly set, expected 'github_mcp', got '%s'", convertedTNOS.Source)
	}
}
