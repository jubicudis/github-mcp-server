// ...existing code...
package tnosmcp

// ContextVector7D represents a 7D context vector in the TNOS system.
// Encodes Who, What, When, Where, Why, How, and Extent for every quantum-contextual operation.
type ContextVector7D struct {
	Who    string  `json:"who"`    // Identity of source, actor, or recipient
	What   string  `json:"what"`   // Task, function, or object of attention
	When   int64   `json:"when"`   // Temporal vector (initiation or time window)
	Where  string  `json:"where"`  // Spatial or domain-specific location
	Why    string  `json:"why"`    // Underlying intent, causality, or purpose
	How    string  `json:"how"`    // Methodology or mode of transformation
	Extent float64 `json:"extent"` // Degree of force, resolution, completeness, or priority
	Source string  `json:"source"` // Origin or context emitter
}

// Canonical port assignments for Quantum Handshake Protocol (QHP)
const (
	PortTNOSMCPServer   = 9001  // TNOS MCP server
	PortMCPBridge       = 10619 // MCP Bridge
	PortGitHubMCPServer = 10617 // GitHub MCP server
	PortCopilotLLM      = 8083  // Copilot LLM server
)

// MCPBridge defines the interface for an MCP bridge.
// Responsible for bridging and translating 7D context between systems.
type MCPBridge interface {
	// BridgeMCPContext bridges context between GitHub and TNOS using TS7D vectors.
	BridgeMCPContext(githubContext, tnosContext interface{}) *ContextVector7D
	// TranslateContext translates a source context to a target system using 7D collapse.
	TranslateContext(sourceContext interface{}, targetSystem string) *ContextVector7D
}

// QuantumHandshakeProtocol defines the QHP handshake and trust protocol for secure AI-to-AI/server comms.
type QuantumHandshakeProtocol interface {
	// GenerateQuantumFingerprint returns the unique cryptographic identity for this node/AI.
	GenerateQuantumFingerprint() string
	// PerformHandshake executes the QHP handshake with another entity using canonical ports.
	PerformHandshake(remoteFingerprint string, port int) (success bool, sessionKey string)
	// SyncTrustTable updates the trust table after a successful handshake.
	SyncTrustTable(remoteFingerprint string, timestamp int64) error
	// ExchangeSessionKeys exchanges ephemeral session keys post-verification.
	ExchangeSessionKeys(remoteFingerprint string) (sessionKey string, err error)
	// SetDecayTimeout sets the handshake/session expiration policy.
	SetDecayTimeout(timeoutSeconds int) error
	// DeveloperOverride allows a developer override with a verified token (auditable, restricted).
	DeveloperOverride(overrideToken string) (success bool, err error)
}

// NewContextVector7D creates a default context vector for TS7D-based operations.
func NewContextVector7D() *ContextVector7D {
	return &ContextVector7D{
		Who:    "System",
		What:   "Operation",
		When:   0,
		Where:  "Context_System",
		Why:    "Default_Operation",
		How:    "Standard_Method",
		Extent: 1.0,
		Source: "system",
	}
}

// Note:
// - All context transfer, handshake, and trust operations must use the 7D context model.
// - QHP handshake attempts and fallback routing must use the canonical ports in order of priority.
// - This file is canonical for TNOS MCP interface, 7D context, and QHP logic. Do not duplicate or move.
// ...existing code...
