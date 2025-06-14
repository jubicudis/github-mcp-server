/*
 * WHO: BloodBridgeManager
 * WHAT: Blood integration for GitHub MCP server
 * WHEN: During API operations with GitHub
 * WHERE: System Layer 6 (Integration)
 * WHY: To provide Helical Memory and QHP integration
 * HOW: Using biomimetic bridge architecture
 * EXTENT: All blood-related operations
 */

package ghmcp

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gorilla/websocket"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/bridge"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/common"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/translations"

	"github.com/google/go-github/v71/github"
)

// OperationalMode defines the operational state of the GitHub MCP server
type OperationalMode int

const (
	// ModeStandalone indicates the server is running without a connection to TNOS MCP
	ModeStandalone OperationalMode = iota
	// ModeBloodConnected indicates the server is connected to TNOS MCP and blood components are active
	ModeBloodConnected
)

var currentOperationalMode OperationalMode = ModeStandalone // Default to standalone

// logToHelicalMemory routes events through the provided LoggerInterface
func logToHelicalMemory(ctx context.Context, event string, details map[string]interface{}, context7d translations.ContextVector7D, logger log.LoggerInterface) error {
	// Standalone: log locally
	if currentOperationalMode != ModeBloodConnected {
		logger.Info(event+"_standalone", "details", details, "context", context7d)
		return nil
	}
	// Blood-connected: route through TNOS MCP logger
	logger.Info(event, "details", details, "context", context7d)
	return nil
}

// logWarningToFile logs a warning message to the GitHub MCP server logs
func logWarningToFile(ctx context.Context, message string) {
	// TODO: Implement logging to github-mcp-server/logs/
	fmt.Fprintf(os.Stderr, "WARNING: %s\n", message)
}

// Canonical QHP ports for GitHub MCP server
var canonicalQHPPorts = []int{9001, 10617, 10619, 8083}

// InitializeBloodSystem checks for TNOS MCP connection and sets the operational mode.
// This should be called at server startup.
func InitializeBloodSystem(ctx context.Context, logger log.LoggerInterface) {
	if common.CheckTNOSConnection(ctx) {
		currentOperationalMode = ModeBloodConnected
		logger.Info("BloodSystemInit", "status", "connected", "mode", currentOperationalMode)
	} else {
		currentOperationalMode = ModeStandalone
		logger.Info("BloodSystemInit", "status", "standalone", "mode", currentOperationalMode)
	}
}

// GetOperationalMode returns the current operational mode.
func GetOperationalMode() OperationalMode {
	return currentOperationalMode
}

// QuantumBridgeMode represents the operational mode of the QHP bridge
type QuantumBridgeMode int

// QHP bridge modes
const (
	QHPBridgeModeNormal QuantumBridgeMode = iota
	QHPBridgeModeFallback
	QHPBridgeModeSecure
	QHPBridgeModeDebug
)

// HelicalMemoryLog logs an event to the TNOS Helical Memory system
func HelicalMemoryLog(ctx context.Context, event string, details map[string]interface{}, context7d translations.ContextVector7D, logger log.LoggerInterface) error {
	if GetOperationalMode() != ModeBloodConnected {
		_ = logToHelicalMemory(ctx, event+"_standalone", details, context7d, logger)
		return nil
	}
	return logToHelicalMemory(ctx, event, details, context7d, logger)
}

// PerformQuantumHandshake executes the QHP handshake protocol with TNOS components using canonical port list
func PerformQuantumHandshake(ctx context.Context, metadata map[string]interface{}, logger log.LoggerInterface) (map[string]interface{}, error) {
	context7d := translations.ContextVector7D{
		Who:    "GitHubMCPServer",
		What:   "QHPHandshake",
		When:   time.Now().Unix(),
		Where:  "SystemLayer6",
		Why:    "SecureComms",
		How:    "QHP",
		Extent: 1.0,
		Source: "github-mcp-server",
	}
	if GetOperationalMode() != ModeBloodConnected {
		logger.Info("QHPHandshake_skipped", "mode", "standalone")
		return map[string]interface{}{"status": "skipped_standalone_mode"}, nil
	}
	for _, port := range canonicalQHPPorts {
		endpoint := fmt.Sprintf("ws://localhost:%d/ws", port)
		logger.Info("QHPHandshake_connect", "endpoint", endpoint)
		conn, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
		if err != nil {
			logger.Info("QHPHandshake_dial_failed", "endpoint", endpoint, "error", err)
			continue
		}
		defer conn.Close()
		// Send handshake challenge
		challenge := fmt.Sprintf("QHP-Challenge-%d", time.Now().UnixNano())
		msg := map[string]interface{}{"event": "QHPHandshake", "challenge": challenge, "metadata": metadata}
		if err := conn.WriteJSON(msg); err != nil {
			logger.Info("QHPHandshake_write_failed", "endpoint", endpoint, "error", err)
			continue
		}
		// Await response
		var resp map[string]interface{}
		if err := conn.ReadJSON(&resp); err != nil {
			logger.Info("QHPHandshake_read_failed", "endpoint", endpoint, "error", err)
			continue
		}
		// Validate response status
		if status, ok := resp["status"].(string); ok && status == "ok" {
			logger.Info("QHPHandshake_success", "endpoint", endpoint)
			_ = HelicalMemoryLog(ctx, "QHPHandshake_success", resp, context7d, logger)
			return resp, nil
		}
		logger.Info("QHPHandshake_rejected", "endpoint", endpoint, "resp", resp)
	}
	logger.Info("QHPHandshake_all_failed", "attempts", len(canonicalQHPPorts))
	return nil, fmt.Errorf("all QHP handshake attempts failed")
}

// global bridge instance registered by main
var bloodBridge *bridge.BloodCirculation

// RegisterBloodBridge sets the bridge instance for circulation
func RegisterBloodBridge(b *bridge.BloodCirculation) {
	bloodBridge = b
}

// TranslateGitHubEventToBloodCell converts and sends a GitHub event into TNOS circulation
func TranslateGitHubEventToBloodCell(ctx context.Context, event *github.Event, logger log.LoggerInterface) error {
	cell := &bridge.BloodCell{
		ID:          GenerateBloodCellID(),
		Type:        "Red",
		Source:      "github_mcp",
		Destination: "tnos-mcp-server",
		Timestamp:   time.Now().Unix(),
		Priority:    5,
		OxygenLevel: 1.0,
		Context7D: translations.ContextVector7D{
			Who:    "GitHubMCPServer",
			What:   "EventTranslation",
			When:   time.Now().Unix(),
			Where:  "SystemLayer6",
			Why:    "BloodCirculation",
			How:    "Biomimetics",
			Extent: 1.0,
			Source: "github-mcp-server",
		},
		Payload: map[string]interface{}{"event": event},
	}
	if bloodBridge == nil {
		logger.Info("Blood bridge not registered, event not sent", "cellID", cell.ID)
		return nil
	}
	// send over the bridge
	if err := bloodBridge.SendBloodCell(cell); err != nil {
		logger.Info("Failed to send blood cell", "error", err)
		return err
	}
	logger.Info("TranslateGitHubEventToBloodCell sent", "cellID", cell.ID)
	return nil
}

// Handler stub
func HandleIncomingBloodCell(ctx context.Context, cell *bridge.BloodCell, logger log.LoggerInterface) {
	logger.Info("Received blood cell", "cellID", cell.ID)
	// TODO: map cell.Payload back into GitHub MCP responses
}

// GenerateBloodCellID generates a unique ID for a blood cell
func GenerateBloodCellID() string {
	return fmt.Sprintf("blood-cell-%d", time.Now().UnixNano())
}

// CirculateBloodCell circulates a blood cell through the TNOS blood system
func CirculateBloodCell(ctx context.Context, cell *bridge.BloodCell, logger log.LoggerInterface) error {
	if GetOperationalMode() != ModeBloodConnected {
		logger.Info("CirculateBloodCell_standalone", "cellID", cell.ID)
		return nil
	}
	return HelicalMemoryLog(ctx, "CirculateBloodCell", map[string]interface{}{"cellId": cell.ID, "payload": cell.Payload}, cell.Context7D, logger)
}

// GetBloodCirculationState retrieves the current state of the blood circulation system
func GetBloodCirculationState(ctx context.Context, logger log.LoggerInterface) (*bridge.BloodCirculationState, error) {
	context7d := translations.ContextVector7D{
		Who:    "GitHubMCPServer",
		What:   "BloodCirculationState",
		When:   time.Now().Unix(),
		Where:  "SystemLayer6",
		Why:    "Monitoring",
		How:    "HelicalMemory",
		Extent: 1.0,
		Source: "github-mcp-server",
	}
	state := &bridge.BloodCirculationState{}
	if GetOperationalMode() != ModeBloodConnected {
		logger.Info("GetBloodCirculationState_standalone", "state", state)
		return state, nil
	}
	if err := HelicalMemoryLog(ctx, "GetBloodCirculationState", map[string]interface{}{"state": state}, context7d, logger); err != nil {
		return state, err
	}
	return state, nil
}

// Canonical test file for blood.go
// All tests must directly and robustly test the canonical logic in blood.go
// Remove all legacy, duplicate, or non-canonical tests
// Reference only helpers from /pkg/common and /pkg/testutil
// No import cycles, duplicate imports, or undefined helpers
// All test cases must match the actual signatures and logic of blood.go
