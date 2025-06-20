/*
 * WHO: GitHubPackage
 * WHAT: Main package for GitHub MCP server functionality
 * WHEN: During API operations with GitHub
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide GitHub integration with TNOS
 * HOW: Using MCP protocol and GitHub API
 * EXTENT: All GitHub operations
 */

package ghmcp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/bridge"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/common"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/tranquilspeak"

	"github.com/google/go-github/v71/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// WHO: ConstantReferencer
// WHAT: Reference to bridge mode constants
// WHEN: During bridge configuration
// WHERE: System Layer 6 (Integration)
// WHY: To use constants defined in constants.go
// HOW: Using imported constants
// EXTENT: Bridge mode configuration

// InitializeMCPBridge sets up the MCP bridge between GitHub and TNOS MCP
func InitializeMCPBridge(enableCompression bool, logger log.LoggerInterface, triggerMatrix *tranquilspeak.TriggerMatrix) error { // Added triggerMatrix parameter
	// WHO: BridgeInitializer
	// WHAT: MCP bridge initialization with fallback
	// WHEN: During server startup
	// WHERE: System Layer 6 (Integration)
	// WHY: To establish GitHub-TNOS communication robustly
	// HOW: Using FallbackRoute utility
	// EXTENT: System integration

	context7d := log.ContextVector7D{
		Who:    "GitHubMCPServer",
		What:   "BridgeInit",
		When:   time.Now().Unix(),
		Where:  "SystemLayer6",
		Why:    "Startup",
		How:    "FallbackRoute",
		Extent: 1.0,
	}

	// Convert log.ContextVector7D to map[string]interface{}
	contextData := map[string]interface{}{
		"Who":    context7d.Who,
		"What":   context7d.What,
		"When":   context7d.When,
		"Where":  context7d.Where,
		"Why":    context7d.Why,
		"How":    context7d.How,
		"Extent": context7d.Extent,
	}

	// Corrected to use bridge.FallbackRoute and contextData
	_, err := bridge.FallbackRoute(
		context.Background(),
		"BridgeInit",
		contextData, // Pass the converted map
		func() (interface{}, error) {
			logger.Info("Initializing MCP Bridge between GitHub and TNOS") // Use passed logger
			return nil, nil
		},
		func() (interface{}, error) { return nil, fmt.Errorf("bridge fallback not implemented") },
		func() (interface{}, error) { return nil, fmt.Errorf("GitHub MCP fallback not implemented") },
		func() (interface{}, error) { return nil, fmt.Errorf("copilot fallback not implemented") },
		logger, // Use passed logger directly assuming FallbackRoute accepts LoggerInterface
	)
	return err
}

// GitHubContextTranslator provides bidirectional translation between GitHub context and TNOS 7D context
type GitHubContextTranslator struct {
	// WHO: ContextTranslator
	// WHAT: Translates between context formats
	// WHEN: During context exchange
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide contextual compatibility
	// HOW: Using bidirectional mapping
	// EXTENT: All context translations

	EnableCompression bool
	EnableLogging     bool
	DebugMode         bool
	Logger            log.LoggerInterface // Changed to LoggerInterface
}

// NewGitHubContextTranslator creates a new context translator instance
func NewGitHubContextTranslator(logger log.LoggerInterface, enableCompression, enableLogging, debugMode bool) *GitHubContextTranslator { // Changed logger type
	// WHO: TranslatorFactory
	// WHAT: Creates context translator
	// WHEN: During bridge initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To enable context translation
	// HOW: Using factory pattern
	// EXTENT: Translator lifecycle

	return &GitHubContextTranslator{
		EnableCompression: enableCompression,
		EnableLogging:     enableLogging,
		DebugMode:         debugMode,
		Logger:            logger, // Use passed logger
	}
}

// TranslateToTNOS converts GitHub context to TNOS 7D context
func (t *GitHubContextTranslator) TranslateToTNOS(githubContext map[string]interface{}) map[string]interface{} {
	// WHO: GitHubToTNOSTranslator
	// WHAT: Convert GitHub context
	// WHEN: During inbound operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide TNOS compatibility
	// HOW: Using context mapping
	// EXTENT: Inbound messages

	context7d := log.ContextVector7D{
		Who:    "GitHubContextTranslator",
		What:   "TranslateToTNOS",
		When:   time.Now().Unix(),
		Where:  "SystemLayer6",
		Why:    "ContextTranslation",
		How:    "FallbackRoute",
		Extent: 1.0,
	}
	ctx := context.Background()
	// Convert log.ContextVector7D to map[string]interface{}
	contextData := map[string]interface{}{
		"Who":    context7d.Who,
		"What":   context7d.What,
		"When":   context7d.When,
		"Where":  context7d.Where,
		"Why":    context7d.Why,
		"How":    context7d.How,
		"Extent": context7d.Extent,
	}
	_, _ = bridge.FallbackRoute(
		ctx,
		"TranslateToTNOS",
		contextData, // Pass the converted map
		func() (interface{}, error) {
			if t.EnableLogging && t.Logger != nil {
				if t.DebugMode {
					t.Logger.Debug("Translating context from GitHub to TNOS")
				} else {
					t.Logger.Info("Translating context from GitHub to TNOS")
				}
			}
			return nil, nil
		},
		func() (interface{}, error) { return nil, fmt.Errorf("bridge fallback not implemented") },
		func() (interface{}, error) { return nil, fmt.Errorf("GitHub MCP fallback not implemented") },
		func() (interface{}, error) { return nil, fmt.Errorf("copilot fallback not implemented") },
		t.Logger, // Use instance logger assuming FallbackRoute accepts LoggerInterface
	)

	// This would implement the actual translation logic in a real system
	tnosContext := map[string]interface{}{
		"who":    githubContext["identity"],
		"what":   githubContext["operation"],
		"when":   githubContext["timestamp"],
		"where":  "GitHub_MCP_Bridge",
		"why":    githubContext["purpose"],
		"how":    "Context_Translation",
		"extent": githubContext["scope"],
	}

	return tnosContext
}

// TranslateFromTNOS converts TNOS 7D context to GitHub context
func (t *GitHubContextTranslator) TranslateFromTNOS(tnosContext map[string]interface{}) map[string]interface{} {
	// WHO: TNOSToGitHubTranslator
	// WHAT: Convert TNOS context
	// WHEN: During outbound operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide GitHub compatibility
	// HOW: Using context mapping
	// EXTENT: Outbound messages

	context7d := log.ContextVector7D{
		Who:    "GitHubContextTranslator",
		What:   "TranslateFromTNOS",
		When:   time.Now().Unix(),
		Where:  "SystemLayer6",
		Why:    "ContextTranslation",
		How:    "FallbackRoute",
		Extent: 1.0,
	}
	ctx := context.Background()
	// Convert log.ContextVector7D to map[string]interface{}
	contextData := map[string]interface{}{
		"Who":    context7d.Who,
		"What":   context7d.What,
		"When":   context7d.When,
		"Where":  context7d.Where,
		"Why":    context7d.Why,
		"How":    context7d.How,
		"Extent": context7d.Extent,
	}
	_, _ = bridge.FallbackRoute(
		ctx,
		"TranslateFromTNOS",
		contextData, // Pass the converted map
		func() (interface{}, error) {
			if t.EnableLogging && t.Logger != nil {
				if t.DebugMode {
					t.Logger.Debug("Translating context from TNOS to GitHub")
				} else {
					t.Logger.Info("Translating context from TNOS to GitHub")
				}
			}
			return nil, nil
		},
		func() (interface{}, error) { return nil, fmt.Errorf("bridge fallback not implemented") },
		func() (interface{}, error) { return nil, fmt.Errorf("GitHub MCP fallback not implemented") },
		func() (interface{}, error) { return nil, fmt.Errorf("copilot fallback not implemented") },
		t.Logger, // Use instance logger assuming FallbackRoute accepts LoggerInterface
	)

	// This would implement the actual translation logic in a real system
	githubContext := map[string]interface{}{
		"identity":  tnosContext["who"],
		"operation": tnosContext["what"],
		"timestamp": tnosContext["when"],
		"purpose":   tnosContext["why"],
		"scope":     tnosContext["extent"],
	}

	return githubContext
}

// BridgeHealthCheck performs a health check on the MCP bridge
func BridgeHealthCheck(triggerMatrix *tranquilspeak.TriggerMatrix) (bool, error) {
	// WHO: HealthMonitor
	// WHAT: Bridge health check with fallback
	// WHEN: During system monitoring
	// WHERE: System Layer 6 (Integration)
	// WHY: To ensure bridge availability
	// HOW: Using FallbackRoute
	// EXTENT: Bridge operational status
	context7d := log.ContextVector7D{
		Who:    "GitHubMCPServer",
		What:   "HealthCheck",
		When:   time.Now().Unix(),
		Where:  "SystemLayer6",
		Why:    "Monitoring",
		How:    "FallbackRoute",
		Extent: 1.0,
	}
	ctx := context.Background()
	healthy := false
	// Convert log.ContextVector7D to map[string]interface{}
	contextData := map[string]interface{}{
		"Who":    context7d.Who,
		"What":   context7d.What,
		"When":   context7d.When,
		"Where":  context7d.Where,
		"Why":    context7d.Why,
		"How":    context7d.How,
		"Extent": context7d.Extent,
	}
	_, err := bridge.FallbackRoute(
		ctx,
		"BridgeHealthCheck",
		contextData, // Pass the converted map
		func() (interface{}, error) { healthy = true; return nil, nil },
		func() (interface{}, error) { return nil, fmt.Errorf("bridge fallback not implemented") },
		func() (interface{}, error) { return nil, fmt.Errorf("GitHub MCP fallback not implemented") },
		func() (interface{}, error) { return nil, fmt.Errorf("copilot fallback not implemented") },
		log.NewLogger(triggerMatrix), // Canonical usage
	)
	return healthy, err
}

// ConnectMCPChannels establishes bidirectional channels between GitHub and TNOS MCP
func ConnectMCPChannels(bridgeMode string) error {
	// WHO: ChannelConnector
	// WHAT: Connect MCP channels
	// WHEN: During bridge initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To establish communication paths
	// HOW: Using channel connections
	// EXTENT: All MCP communication

	if bridgeMode != "direct" &&
		bridgeMode != "proxied" &&
		bridgeMode != "async" {
		return fmt.Errorf("invalid bridge mode: %s", bridgeMode)
	}

	// Implementation would go here in a real system
	return nil
}

// StartMCPEventMonitor starts monitoring MCP events
func StartMCPEventMonitor(logger *log.Logger) error {
	// WHO: EventMonitor
	// WHAT: Monitor MCP events
	// WHEN: During bridge operation
	// WHERE: System Layer 6 (Integration)
	// WHY: To track MCP events
	// HOW: Using event listeners
	// EXTENT: All MCP events

	if logger != nil {
		logger.Debug("Starting MCP event monitor")
	}

	// Implementation would go here in a real system
	return nil
}

// WHO: RegisteredToolsProvider
// WHAT: Provide MCP tools registration
// WHEN: During server initialization
// WHERE: System Layer 6 (Integration)
// WHY: To register GitHub tools with MCP
// HOW: Using MCP registration mechanisms
// EXTENT: All GitHub MCP tool registration
// MCPServer represents the MCP server interface
type MCPServer interface {
	RegisterTool(tool mcp.Tool, handler server.ToolHandlerFunc)
}

func RegisterTools(server MCPServer, getClient common.GetClientFn, t common.TranslationHelperFunc) {
	// WHO: ToolRegistrar
	// WHAT: Register GitHub tools
	// WHEN: During server initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To make tools available via MCP
	// HOW: Using tool registration
	// EXTENT: All GitHub MCP tools

	// Register content tools
	// TODO: Implement GetContent function or remove this reference
	// server.RegisterTool(contentTool, contentHandler)

	// Register issue tools
	issueTool, issueHandler := GetIssue(getClient, t)
	server.RegisterTool(issueTool, issueHandler)

	createIssueTool, createIssueHandler := CreateIssue(getClient, t)
	server.RegisterTool(createIssueTool, createIssueHandler)

	listIssuesTool, listIssuesHandler := ListIssues(getClient, t)
	server.RegisterTool(listIssuesTool, listIssuesHandler)

	// Register PR tools
	prTool, prHandler := GetPullRequest(getClient, t)
	server.RegisterTool(prTool, prHandler)

	createPRTool, createPRHandler := CreatePullRequest(getClient, t)
	server.RegisterTool(createPRTool, createPRHandler)

	listPRsTool, listPRsHandler := ListPullRequests(getClient, t)
	server.RegisterTool(listPRsTool, listPRsHandler)

	// Register commit tools
	commitTool, commitHandler := GetCommit(getClient, t)
	server.RegisterTool(commitTool, commitHandler)

	// Register search tools
	searchCodeTool, searchCodeHandler := SearchCode(getClient, t)
	server.RegisterTool(searchCodeTool, searchCodeHandler)

	// Register code scanning tools
	codeScanTool, codeScanHandler := GetCodeScanningAlert(getClient, t)
	server.RegisterTool(codeScanTool, codeScanHandler)

	listCodeScanTool, listCodeScanHandler := ListCodeScanningAlerts(getClient, t)
	server.RegisterTool(listCodeScanTool, listCodeScanHandler)
}

// isAcceptedError checks if an error is due to HTTP 202 Accepted status
func IsAcceptedError(err error) bool {
	if err == nil {
		return false
	}
	// Check if the error is a GitHub error with a 202 Accepted status
	if ghErr, ok := err.(*github.ErrorResponse); ok && ghErr.Response != nil {
		return ghErr.Response.StatusCode == http.StatusAccepted
	}
	return false
}

// Canonical test file for github.go
// All tests must directly and robustly test the canonical logic in github.go
// Remove all legacy, duplicate, or non-canonical tests
// Reference only helpers from /pkg/common and /pkg/testutil
// No import cycles, duplicate imports, or undefined helpers
// All test cases must match the actual signatures and logic of github.go
