/*
 * WHO: MCPBridgeDemonstrator
 * WHAT: Example of properly importing and using MCP packages
 * WHEN: During compilation and execution of the bridge
 * WHERE: System Layer 6 (Integration)
 * WHY: To demonstrate correct import paths and usage patterns
 * HOW: Using proper Go module import paths and initialization
 * EXTENT: Module import demonstration for GitHub MCP Bridge
 */

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	// Import using the module name as defined in go.mod
	"tranquility-neuro-os/github-mcp-server/pkg/log"
	"tranquility-neuro-os/github-mcp-server/pkg/translations"
)

// WHO: ExecutionEntrypoint
// WHAT: Main function demonstrating imports
// WHEN: During program execution
// WHERE: System Layer 6 (Integration)
// WHY: To showcase proper imports
// HOW: Using standard Go initialization
// EXTENT: Entire demonstration program
func importExampleMain() {
	// WHO: MCPBridgeInitializer
	// WHAT: Initialize MCP Bridge components
	// WHEN: At program start
	// WHERE: System Layer 6 (Integration)
	// WHY: To establish proper logging and context
	// HOW: Using the log package with debug configuration
	// EXTENT: MCP Bridge initialization
	logger := log.NewLogger(log.Config{
		Level:      log.LevelDebug,
		ConsoleOut: true,
		FileOut:    false,
	})

	logger.Info("Starting MCP Bridge Import Test",
		"timestamp", time.Now().Format(time.RFC3339),
		"process", os.Getpid())

	// WHO: ContextVectorGenerator
	// WHAT: Create a 7D context vector
	// WHEN: During demonstration
	// WHERE: System Layer 6 (Integration)
	// WHY: To establish proper context for the bridge operation
	// HOW: Using the translations package
	// EXTENT: Context initialization
	contextVector := translations.NewContextVector7D(map[string]interface{}{
		"who":    "MCPBridgeDemonstrator",
		"what":   "ImportVerification",
		"when":   time.Now().Unix(),
		"where":  "System_Layer_6",
		"why":    "Module_Resolution",
		"how":    "Direct_Instantiation",
		"extent": 1.0,
		"source": "import_example.go",
		"meta": map[string]interface{}{
			"version": "1.0.0",
			"purpose": "demonstrate_imports",
		},
	})

	logger.Info("Created context vector",
		"who", contextVector.Who,
		"what", contextVector.What,
		"when", contextVector.When,
		"where", contextVector.Where)

	// WHO: CompressionDemonstrator
	// WHAT: Demonstrate compression-first approach
	// WHEN: During example execution
	// WHERE: System Layer 6 (Integration)
	// WHY: To showcase compression capabilities
	// HOW: Using compression-first approach
	// EXTENT: Compression demonstration
	logger.Info("Compressing context vector...")
	compressedVector := contextVector.Compress()

	// WHO: SerializationDemonstrator
	// WHAT: Demonstrate context serialization
	// WHEN: During example execution
	// WHERE: System Layer 6 (Integration)
	// WHY: To showcase serialization capabilities
	// HOW: Using standard encoding with context preservation
	// EXTENT: Serialization demonstration
	contextMap := compressedVector.ToMap()

	jsonData, err := json.MarshalIndent(contextMap, "", "  ")
	if err != nil {
		logger.Error("Failed to serialize context", "error", err.Error())
		os.Exit(1)
	}

	// WHO: ResultDisplayer
	// WHAT: Display test results
	// WHEN: During results presentation
	// WHERE: System Layer 6 (Integration)
	// WHY: To show successful import and usage
	// HOW: Using formatted console output
	// EXTENT: Results presentation
	fmt.Println("\n=== TNOS GitHub MCP Server Import Test ===")
	fmt.Println("✅ Successfully imported and used translations package")
	fmt.Println("✅ Successfully imported and used log package")
	fmt.Println("✅ Module path resolution is working correctly")

	// WHO: ContextVectorDisplayer
	// WHAT: Display compressed context vector
	// WHEN: During results presentation
	// WHERE: System Layer 6 (Integration)
	// WHY: To demonstrate proper context handling
	// HOW: Using JSON output
	// EXTENT: Context presentation
	fmt.Println("\n=== Compressed 7D Context Vector ===")
	fmt.Println(string(jsonData))

	// WHO: GitHubToTNOSTranslator
	// WHAT: Demonstrate GitHub to TNOS context translation
	// WHEN: During example execution
	// WHERE: System Layer 6 (Integration)
	// WHY: To showcase translation capabilities
	// HOW: Using translation functions with compression-first approach
	// EXTENT: Translation demonstration
	githubContext := map[string]interface{}{
		"identity":   "github-user",
		"operation":  "create-issue",
		"resource":   "repo://owner/repo/issues",
		"timestamp":  time.Now().Unix(),
		"data":       map[string]interface{}{"title": "Example Issue", "body": "This is an example"},
		"metadata":   map[string]interface{}{"client": "web", "version": "1.0"},
		"token_type": "oauth",
	}

	logger.Info("Translating GitHub context to TNOS 7D...")
	tnosContext := translations.MCPContextToTNOS(githubContext)

	// Apply compression (compression-first approach)
	tnosContext = translations.CompressTranslationContext(tnosContext)

	jsonTNOS, _ := json.MarshalIndent(tnosContext.ToMap(), "", "  ")

	fmt.Println("\n=== Translated Context (GitHub → TNOS) ===")
	fmt.Println(string(jsonTNOS))

	logger.Info("Import test completed successfully")
}
