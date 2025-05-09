#!/bin/bash
# WHO: MCPModulePathResolver
# WHAT: Script to properly resolve Go module imports for GitHub MCP Server
# WHEN: During development and build processes
# WHERE: System Layer 6 (Integration)
# WHY: To enable proper module resolution for the MCP Bridge
# HOW: Using Go module replacement directives
# EXTENT: All Go module import operations in GitHub MCP Server

# Set proper Go environment variables following 7D principles
# WHO: EnvironmentConfigurator
# WHAT: Environment setup
# WHEN: Script initialization
# WHERE: System Layer 6 (Integration)
# WHY: To configure proper Go module environment
# HOW: Using environment variables
# EXTENT: Current script execution
export GO111MODULE=on
export GOWORK=off
export TNOS_CONTEXT_MODE="debug"
export MCP_LOG_LEVEL="debug"

# Define colors for better output visibility
# WHO: OutputFormatter
# WHAT: Terminal color configuration
# WHEN: During output presentation
# WHERE: System Layer 6 (Integration)
# WHY: To improve readability of diagnostic output
# HOW: Using ANSI color codes
# EXTENT: All terminal output in this script
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Get the absolute path to the GitHub MCP Server directory
# WHO: PathResolver
# WHAT: Directory path resolution
# WHEN: Script initialization
# WHERE: System Layer 6 (Integration)
# WHY: To establish proper path references
# HOW: Using shell path resolution
# EXTENT: All path operations in this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MCP_SERVER_DIR="$(cd "${SCRIPT_DIR}/../../" && pwd)"
echo -e "${BLUE}=== TNOS GitHub MCP Server Module Resolution ===\n${NC}"
echo -e "${CYAN}Working Directory:${NC} ${MCP_SERVER_DIR}"

# Create a go.mod.fixed file with the proper replace directive
# WHO: ModuleFixGenerator
# WHAT: Go module fix configuration
# WHEN: During module resolution
# WHERE: System Layer 6 (Integration)
# WHY: To provide correct import paths
# HOW: Using Go module replace directives
# EXTENT: Temporary module configuration
GO_MOD_FIXED="${MCP_SERVER_DIR}/go.mod.fixed"
cat >"${GO_MOD_FIXED}" <<EOG
// WHO: ModuleDefinition
// WHAT: Fixed Go module configuration for GitHub MCP Server
// WHEN: During module resolution
// WHERE: System Layer 6 (Integration)
// WHY: To support proper import resolution
// HOW: Using module declarations and replace directives
// EXTENT: All module imports

module tranquility-neuro-os/github-mcp-server

go 1.24

// Preserve original requirements from go.mod
$(grep -v "^module\|^go" "${MCP_SERVER_DIR}/go.mod")

// Add replace directive to map module to local directory
replace tranquility-neuro-os/github-mcp-server => ./
EOG

echo -e "${GREEN}✓${NC} Created fixed module configuration"

# Create a test file that properly imports from the module
# WHO: ImportExampleGenerator
# WHAT: Import example generation
# WHEN: During import testing
# WHERE: System Layer 6 (Integration)
# WHY: To verify module resolution
# HOW: Using proper import statements
# EXTENT: Import verification
EXAMPLE_FILE="${MCP_SERVER_DIR}/cmd/bridge/import_example.go"
cat >"${EXAMPLE_FILE}" <<EOT
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
func main() {
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
EOT

echo -e "${GREEN}✓${NC} Created import example file"

# Create the execution script to test the imports
# WHO: TestExecutor
# WHAT: Import test executor
# WHEN: During import testing
# WHERE: System Layer 6 (Integration)
# WHY: To verify module resolution
# HOW: Using temporary module configuration
# EXTENT: Import verification
cat >"${MCP_SERVER_DIR}/cmd/bridge/run_import_example.sh" <<EOS
#!/bin/bash
# WHO: ImportExampleRunner
# WHAT: Script to test Go module imports
# WHEN: During development and testing
# WHERE: System Layer 6 (Integration)
# WHY: To verify module resolution
# HOW: Using temporary module configuration
# EXTENT: Import example execution

# Set proper Go environment variables
export GO111MODULE=on
export GOWORK=off
export GOMOD="${MCP_SERVER_DIR}/go.mod.fixed"

# Navigate to the MCP server directory
cd "${MCP_SERVER_DIR}"

# Run the example with the fixed module configuration
echo "Running import example with fixed module configuration..."
env GO111MODULE=on GOMOD="\${GOMOD}" GOWORK=off go run cmd/bridge/import_example.go

# Return status
exit \$?
EOS

chmod +x "${MCP_SERVER_DIR}/cmd/bridge/run_import_example.sh"
echo -e "${GREEN}✓${NC} Created example execution script"

# Create a persistent module fix script
# WHO: ModuleFixApplier
# WHAT: Module fix application
# WHEN: During development setup
# WHERE: System Layer 6 (Integration)
# WHY: To apply permanent module fix
# HOW: Using module configuration
# EXTENT: Repository-wide fix
cat >"${MCP_SERVER_DIR}/cmd/bridge/apply_module_fix.sh" <<EOA
#!/bin/bash
# WHO: ModuleFixApplier
# WHAT: Script to apply Go module fixes
# WHEN: During development setup
# WHERE: System Layer 6 (Integration)
# WHY: To permanently fix module resolution
# HOW: Using module configuration replacement
# EXTENT: Repository-wide fix

# Set proper Go environment variables
export GO111MODULE=on

# Navigate to the MCP server directory
cd "${MCP_SERVER_DIR}"

# Backup the original go.mod file
cp go.mod go.mod.bak

# Replace the go.mod file with the fixed version
cp go.mod.fixed go.mod

# Run go mod tidy to update dependencies
go mod tidy

echo "Module fix applied successfully. Original go.mod saved as go.mod.bak"
echo "You can now import packages from this module using: \"tranquility-neuro-os/github-mcp-server/pkg/...\""
EOA

chmod +x "${MCP_SERVER_DIR}/cmd/bridge/apply_module_fix.sh"
echo -e "${GREEN}✓${NC} Created module fix application script"

# Now run the example to see if it works
echo -e "\n${YELLOW}Testing module resolution...${NC}"
"${MCP_SERVER_DIR}/cmd/bridge/run_import_example.sh"

# Provide instructions
echo -e "\n${BLUE}=== Module Resolution Instructions ===\n${NC}"
echo -e "1. To test imports: ${CYAN}./cmd/bridge/run_import_example.sh${NC}"
echo -e "2. To apply the fix permanently: ${CYAN}./cmd/bridge/apply_module_fix.sh${NC}"
echo -e "3. To revert to original configuration: ${CYAN}cp go.mod.bak go.mod${NC} (after applying fix)"

echo -e "\n${BLUE}=== Import Pattern ===\n${NC}"
echo -e "Use the following import pattern in your Go files:"
echo -e "${CYAN}import \"tranquility-neuro-os/github-mcp-server/pkg/translations\"${NC}"
echo -e "${CYAN}import \"tranquility-neuro-os/github-mcp-server/pkg/log\"${NC}"

echo -e "\n${GREEN}Module resolution setup complete!${NC}"
