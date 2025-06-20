// protocol.go
// WHO: ProtocolConstants
// WHAT: Protocol, Bridge, and Intent constants for TNOS MCP system
// WHEN: During protocol, bridge, and intent operations
// WHERE: System Layer 6 (Integration)
// WHY: To provide canonical, centralized, and cross-referenced protocol/bridge/intent constants
// HOW: Using Go constants with 7D documentation and TranquilSpeak compliance
// EXTENT: All protocol, bridge, and intent operations

package constants

// Protocol versioning
const (
	ProtocolVersion = "3.0"
)

// Bridge modes
const (
	BridgeModeDirect  = "direct"
	BridgeModeProxied = "proxied"
	BridgeModeAsync   = "async"
)

// QHP Intents (Quantum Handshake Protocol)
const (
	QHPIntentGitHubToTNOS = "github_to_tnos"
	QHPIntentTNOSToGitHub = "tnos_to_github"
)
