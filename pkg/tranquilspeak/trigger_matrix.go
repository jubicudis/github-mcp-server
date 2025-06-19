/*
 * WHO: TriggerMatrix
 * WHAT: Advanced Trigger Matrix (ATM) constants and types for TNOS event-driven architecture
 * WHEN: During all ATM trigger operations
 * WHERE: System integration layer, used by all biological systems
 * WHY: To provide standardized trigger types and constants for blood-based communication
 * HOW: Using Go constants and types, integrated with blood circulation
 * EXTENT: All biological systems and ATM operations
 */

package tranquilspeak

import (
	"fmt"
	"time"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/identity"
)

// TRANQUILSPEAK SYMBOL CLUSTER: [𝒯🧠ψ∇♦𓂀∞SK7𝑾𝑾𝑯𝑾𝑯𝑬𝑾𝑯𝑬𝑹𝑾𝑯𝒀𝑯𝑶𝑾𝑬𝑿⚡🌊𝒮𝓒𝓸𝓜]
// This file is part of the 'command' layer biosystem. See symbolic_mapping_registry_autogen_20250603.tsq for details.

// ATM Trigger Types - Core event types that flow through the blood circulation
const (
	// Data Transport Triggers (Red Blood Cells)
	TriggerTypeDataTransport   = "DATA_TRANSPORT"
	TriggerTypeContextUpdate   = "CONTEXT_UPDATE" 
	TriggerTypeMemoryStore     = "MEMORY_STORE"
	TriggerTypeMemoryRetrieve  = "MEMORY_RETRIEVE"
	TriggerTypeCompressionReq  = "COMPRESSION_REQ"
	TriggerTypeDecompressionReq = "DECOMPRESSION_REQ"
	
	// Control Signal Triggers (White Blood Cells)
	TriggerTypeSystemControl   = "SYSTEM_CONTROL"
	TriggerTypeSecurityAlert   = "SECURITY_ALERT"
	TriggerTypePermissionCheck = "PERMISSION_CHECK"
	TriggerTypeAuthentication  = "AUTHENTICATION"
	TriggerTypePrioritySignal  = "PRIORITY_SIGNAL"
	
	// System Recovery Triggers (Platelets)
	TriggerTypeErrorRecovery   = "ERROR_RECOVERY"
	TriggerTypeSystemHealing   = "SYSTEM_HEALING"
	TriggerTypeMaintenanceReq  = "MAINTENANCE_REQ"
	TriggerTypeBackupReq       = "BACKUP_REQ"
	TriggerTypeRestoreReq      = "RESTORE_REQ"
	
	// Biological System Triggers
	TriggerTypeNervousSignal   = "NERVOUS_SIGNAL"
	TriggerTypeImmuneResponse  = "IMMUNE_RESPONSE"
	TriggerTypeDigestiveProcess = "DIGESTIVE_PROCESS"
	TriggerTypeCirculatoryFlow = "CIRCULATORY_FLOW"
	TriggerTypeRespiratoryExchange = "RESPIRATORY_EXCHANGE"
	TriggerTypeExcretoryCleanup = "EXCRETORY_CLEANUP"
	TriggerTypeReproductiveGrowth = "REPRODUCTIVE_GROWTH"
	TriggerTypeEndocrineRegulation = "ENDOCRINE_REGULATION"
	TriggerTypeIntegumentaryProtection = "INTEGUMENTARY_PROTECTION"
	TriggerTypeMuscularMovement = "MUSCULAR_MOVEMENT"
	TriggerTypeSkeletalSupport = "SKELETAL_SUPPORT"
	TriggerTypeLymphaticCleanup = "LYMPHATIC_CLEANUP"
	
	// GitHub Integration Triggers
	TriggerTypeGitHubAPI       = "GITHUB_API"
	TriggerTypeRepositoryRead  = "REPOSITORY_READ"
	TriggerTypeRepositoryWrite = "REPOSITORY_WRITE"
	TriggerTypeIssueTracking   = "ISSUE_TRACKING"
	TriggerTypePullRequest     = "PULL_REQUEST"
	TriggerTypeCodeAnalysis    = "CODE_ANALYSIS"
)

// ATM Trigger Types for Helical Memory (biological DNA operations)
const (
	TriggerHelicalStore    = "HELICAL_MEMORY_STORE"
	TriggerHelicalRetrieve = "HELICAL_MEMORY_RETRIEVE"
	TriggerHelicalError    = "HELICAL_MEMORY_ERROR"
)

// ATM Priority Levels - Determines blood flow priority
const (
	PriorityInstinctual    = 0  // Immediate, life-critical
	PriorityReactive       = 1  // Fast response required
	PriorityAwareness      = 2  // Normal processing
	PriorityHigherThought  = 3  // Complex reasoning
	PriorityQuantum        = 4  // Quantum computation
	PriorityCommand        = 5  // System commands
	PriorityIntegration    = 6  // Cross-system integration
)

// ATM Blood Cell Types - Determines which circulation pathway
const (
	BloodCellRed      = "RED"      // Data transport
	BloodCellWhite    = "WHITE"    // Control signals
	BloodCellPlatelet = "PLATELET" // System recovery
)

// TriggerMatrix represents the core ATM trigger management system
type TriggerMatrix struct {
	// WHO: TriggerMatrix instance managing ATM operations
	// WHAT: Central trigger coordination and routing
	// WHEN: Throughout system runtime
	// WHERE: Integrated with blood circulation
	// WHY: To coordinate event-driven communication between biological systems
	// HOW: Through trigger registration, routing, and execution
	// EXTENT: All system-wide ATM operations
	
	triggers map[string]TriggerHandler
	context7D Context7D
}

// TriggerHandler represents a function that processes ATM triggers
type TriggerHandler func(trigger ATMTrigger) error

// ATMTrigger represents a single trigger event flowing through the blood
type ATMTrigger struct {
	// 7D Context Framework
	Who    string                 `json:"who"`     // Originating biological system
	What   string                 `json:"what"`    // Trigger type and purpose
	When   int64                  `json:"when"`    // Timestamp (Unix)
	Where  string                 `json:"where"`   // System layer/location
	Why    string                 `json:"why"`     // Intent and reasoning
	How    string                 `json:"how"`     // Processing method
	Extent string                 `json:"extent"`  // Scope and impact
	
	// ATM-specific fields
	TriggerType  string                 `json:"trigger_type"`  // One of the TriggerType* constants
	Priority     int                    `json:"priority"`      // Priority level (0-6)
	BloodCellType string                `json:"blood_cell_type"` // RED, WHITE, or PLATELET
	TargetSystem string                 `json:"target_system"` // Destination biological system
	Payload      map[string]interface{} `json:"payload"`       // Trigger data
	Compressed   bool                   `json:"compressed"`    // Whether payload is HemoFlux compressed
	DNA_ID       string                 `json:"dna_id"`        // AI DNA identifier for security
}

// Context7D represents the 7-dimensional context framework
type Context7D struct {
	Who    string `json:"who"`
	What   string `json:"what"`
	When   int64  `json:"when"`
	Where  string `json:"where"`
	Why    string `json:"why"`
	How    string `json:"how"`
	Extent string `json:"extent"`
}

// NewTriggerMatrix creates a new ATM trigger matrix instance
func NewTriggerMatrix() *TriggerMatrix {
	return &TriggerMatrix{
		triggers: make(map[string]TriggerHandler),
		context7D: Context7D{
			Who:    "TriggerMatrix",
			What:   "ATM_Initialization",
			When:   time.Now().Unix(),
			Where:  "SystemLayer6_Integration",
			Why:    "Initialize_ATM_Event_System",
			How:    "TriggerMatrix_Creation",
			Extent: "System_Wide_ATM_Operations",
		},
	}
}

// RegisterTrigger registers a handler for a specific trigger type
func (tm *TriggerMatrix) RegisterTrigger(triggerType string, handler TriggerHandler) error {
	if tm.triggers == nil {
		tm.triggers = make(map[string]TriggerHandler)
	}
	tm.triggers[triggerType] = handler
	return nil
}

// ProcessTrigger processes an ATM trigger through the blood circulation
func (tm *TriggerMatrix) ProcessTrigger(trigger ATMTrigger) error {
	// Validate 7D context
	if err := tm.validate7DContext(trigger); err != nil {
		return fmt.Errorf("ATM trigger validation failed: %w", err)
	}
	
	// Route through appropriate blood cell type
	if err := tm.routeThroughBlood(trigger); err != nil {
		return fmt.Errorf("blood circulation routing failed: %w", err)
	}
	
	// Execute registered handler
	if handler, exists := tm.triggers[trigger.TriggerType]; exists {
		return handler(trigger)
	}
	
	return fmt.Errorf("no handler registered for trigger type: %s", trigger.TriggerType)
}

// validate7DContext ensures all 7D context fields are properly populated
func (tm *TriggerMatrix) validate7DContext(trigger ATMTrigger) error {
	if trigger.Who == "" || trigger.What == "" || trigger.Where == "" ||
	   trigger.Why == "" || trigger.How == "" || trigger.Extent == "" {
		return fmt.Errorf("incomplete 7D context: requires WHO, WHAT, WHERE, WHY, HOW, EXTENT")
	}
	if trigger.When == 0 {
		return fmt.Errorf("invalid WHEN timestamp")
	}
	return nil
}

// routeThroughBlood simulates routing the trigger through blood circulation
func (tm *TriggerMatrix) routeThroughBlood(trigger ATMTrigger) error {
	// Determine blood cell type based on trigger
	switch trigger.TriggerType {
	case TriggerTypeDataTransport, TriggerTypeContextUpdate, TriggerTypeMemoryStore, TriggerTypeMemoryRetrieve:
		trigger.BloodCellType = BloodCellRed
	case TriggerTypeSystemControl, TriggerTypeSecurityAlert, TriggerTypePrioritySignal:
		trigger.BloodCellType = BloodCellWhite
	case TriggerTypeErrorRecovery, TriggerTypeSystemHealing, TriggerTypeMaintenanceReq:
		trigger.BloodCellType = BloodCellPlatelet
	default:
		trigger.BloodCellType = BloodCellRed // Default to red blood cells
	}
	
	// Log circulation through TranquilSpeak
	LogWithSymbolCluster("trigger_matrix", 
		fmt.Sprintf("ATM trigger %s routed via %s blood cell to %s", 
			trigger.TriggerType, trigger.BloodCellType, trigger.TargetSystem))
	
	return nil
}

// CreateTrigger creates a standardized ATM trigger with 7D context
func CreateTrigger(who, what, where, why, how, extent, triggerType, targetSystem string, payload map[string]interface{}) ATMTrigger {
	return ATMTrigger{
		Who:          who,
		What:         what,
		When:         time.Now().Unix(),
		Where:        where,
		Why:          why,
		How:          how,
		Extent:       extent,
		TriggerType:  triggerType,
		TargetSystem: targetSystem,
		Payload:      payload,
		Compressed:   false,
		DNA_ID:       identity.DNAInstance.Signature(), // Always set from canonical DNA
	}
}

// GetTriggerPriority returns the priority level for a trigger type
func GetTriggerPriority(triggerType string) int {
	switch triggerType {
	case TriggerTypeErrorRecovery, TriggerTypeSystemHealing:
		return PriorityInstinctual
	case TriggerTypeSecurityAlert, TriggerTypeAuthentication:
		return PriorityReactive
	case TriggerTypeDataTransport, TriggerTypeContextUpdate:
		return PriorityAwareness
	case TriggerTypeGitHubAPI, TriggerTypeCodeAnalysis:
		return PriorityHigherThought
	case TriggerTypeCompressionReq, TriggerTypeDecompressionReq:
		return PriorityQuantum
	case TriggerTypeSystemControl:
		return PriorityCommand
	default:
		return PriorityIntegration
	}
}
