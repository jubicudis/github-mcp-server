// TRANQUILSPEAK SYMBOL CLUSTER: [𝒯🔴⚪🟡∞φ𓂀♦SK5𝑾𝑾𝑯𝑾𝑯𝑬𝑾𝑯𝑬𝑹𝑾𝑯𝒀𝑯𝑶𝑾𝑬𝑿🩸💊𝒮𝓒𝓮𝓵𝓵]
// This file is part of the 'circulatory' biosystem. See circulatory/github-mcp-server/symbolic_mapping_registry_autogen_20250603.tsq for details.

/*
 * WHO: BloodCells
 * WHAT: Blood cell types for ATM trigger transport through circulatory system
 * WHEN: During ATM trigger circulation and transport
 * WHERE: Circulatory system pathways (arteries, veins, capillaries)
 * WHY: To provide specialized transport mechanisms for different ATM trigger types
 * HOW: Using red cells (data), white cells (control), platelets (recovery)
 * EXTENT: All ATM trigger transport through blood circulation
 */

package blood

import (
	"time"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/tranquilspeak"
)

// BloodCell represents the base interface for all blood cell types
type BloodCell interface {
	// GetCellType returns the type of blood cell (RED, WHITE, PLATELET)
	GetCellType() string
	
	// GetTargetSystem returns the destination biological system
	GetTargetSystem() string
	
	// GetATMTrigger returns the ATM trigger carried by this blood cell
	GetATMTrigger() tranquilspeak.ATMTrigger
	
	// IsExpired checks if the blood cell has expired and should be removed
	IsExpired() bool
	
	// GetPriority returns the circulation priority of this blood cell
	GetPriority() int
	
	// GetLifespan returns the maximum lifespan of this blood cell
	GetLifespan() time.Duration
}

// RedBloodCell represents data transport blood cells
// Mimics biological red blood cells - carry oxygen (data) throughout the organism
type RedBloodCell struct {
	// WHO: RedBloodCell transporting data through circulation
	// WHAT: Data transport cell carrying ATM triggers with standard priority
	// WHEN: During normal data communication between biological systems
	// WHERE: Throughout arterial and venous circulation pathways
	// WHY: Transport data efficiently with standard priority and flow
	// HOW: Encapsulating ATM triggers in specialized transport cells
	// EXTENT: All standard data communication ATM triggers
	
	cellType      string                        // Always "RED"
	targetSystem  string                        // Destination biological system
	atmTrigger    tranquilspeak.ATMTrigger      // Carried ATM trigger
	birthTime     time.Time                     // When cell was created
	lifespan      time.Duration                 // Maximum cell lifetime
	oxygenLevel   int                           // Data payload "oxygen" level
	hemoglobin    map[string]interface{}        // Additional data transport capacity
	
	// 7D Context
	context7D     tranquilspeak.Context7D       // Cell context
}

// WhiteBloodCell represents control signal blood cells
// Mimics biological white blood cells - provide immune response and system control
type WhiteBloodCell struct {
	// WHO: WhiteBloodCell managing control signals and security
	// WHAT: Control signal cell carrying high-priority ATM triggers
	// WHEN: During system control, security alerts, and priority communications
	// WHERE: Priority circulation pathways with enhanced flow
	// WHY: Provide system defense, control, and priority signal transport
	// HOW: High-priority circulation with security and control capabilities
	// EXTENT: All control signals, security alerts, and priority ATM triggers
	
	cellType      string                        // Always "WHITE"
	targetSystem  string                        // Destination biological system
	atmTrigger    tranquilspeak.ATMTrigger      // Carried ATM trigger
	birthTime     time.Time                     // When cell was created
	lifespan      time.Duration                 // Maximum cell lifetime
	antibodyType  string                        // Type of immune/control response
	priority      int                           // Higher priority than red cells
	securityLevel int                           // Security clearance level
	
	// 7D Context
	context7D     tranquilspeak.Context7D       // Cell context
}

// Platelet represents system recovery blood cells
// Mimics biological platelets - provide clotting (error recovery) and healing
type Platelet struct {
	// WHO: Platelet providing system recovery and healing
	// WHAT: Recovery cell carrying system healing and error recovery triggers
	// WHEN: During system errors, maintenance, and recovery operations
	// WHERE: Error sites and system maintenance areas
	// WHY: Provide system healing, error recovery, and maintenance support
	// HOW: Specialized recovery cells with healing and repair capabilities
	// EXTENT: All error recovery, system healing, and maintenance ATM triggers
	
	cellType      string                        // Always "PLATELET"
	targetSystem  string                        // Destination biological system
	atmTrigger    tranquilspeak.ATMTrigger      // Carried ATM trigger
	birthTime     time.Time                     // When cell was created
	lifespan      time.Duration                 // Maximum cell lifetime
	clottingFactor int                          // Recovery/healing strength
	healingType   string                        // Type of recovery provided
	
	// 7D Context
	context7D     tranquilspeak.Context7D       // Cell context
}

// NewRedBloodCell creates a new red blood cell for data transport
func NewRedBloodCell(targetSystem string, atmTrigger tranquilspeak.ATMTrigger) *RedBloodCell {
	return &RedBloodCell{
		cellType:     tranquilspeak.BloodCellRed,
		targetSystem: targetSystem,
		atmTrigger:   atmTrigger,
		birthTime:    time.Now(),
		lifespan:     120 * time.Second, // 120 seconds lifetime like biological red cells
		oxygenLevel:  100,               // Full oxygen capacity
		hemoglobin:   make(map[string]interface{}),
		context7D: tranquilspeak.Context7D{
			Who:    "RedBloodCell",
			What:   "Data_Transport_Cell",
			When:   time.Now().Unix(),
			Where:  "Circulatory_System",
			Why:    "Transport_ATM_Trigger_Data",
			How:    "Hemoglobin_Based_Transport",
			Extent: "Target_" + targetSystem,
		},
	}
}

// NewWhiteBloodCell creates a new white blood cell for control signals
func NewWhiteBloodCell(targetSystem string, atmTrigger tranquilspeak.ATMTrigger, antibodyType string) *WhiteBloodCell {
	priority := tranquilspeak.GetTriggerPriority(atmTrigger.TriggerType)
	
	return &WhiteBloodCell{
		cellType:      tranquilspeak.BloodCellWhite,
		targetSystem:  targetSystem,
		atmTrigger:    atmTrigger,
		birthTime:     time.Now(),
		lifespan:      30 * time.Second, // Shorter lifetime for quick response
		antibodyType:  antibodyType,
		priority:      priority,
		securityLevel: 5, // High security clearance
		context7D: tranquilspeak.Context7D{
			Who:    "WhiteBloodCell",
			What:   "Control_Signal_Cell",
			When:   time.Now().Unix(),
			Where:  "Priority_Circulation_Pathway",
			Why:    "Transport_Control_Signals_And_Security",
			How:    "High_Priority_Immune_Transport",
			Extent: "Target_" + targetSystem,
		},
	}
}

// NewPlatelet creates a new platelet for system recovery
func NewPlatelet(targetSystem string, atmTrigger tranquilspeak.ATMTrigger, healingData map[string]interface{}) *Platelet {
	// Determine healing type from trigger
	healingType := "general_recovery"
	if ht, exists := healingData["healing_type"]; exists {
		if htStr, ok := ht.(string); ok {
			healingType = htStr
		}
	}
	
	return &Platelet{
		cellType:       tranquilspeak.BloodCellPlatelet,
		targetSystem:   targetSystem,
		atmTrigger:     atmTrigger,
		birthTime:      time.Now(),
		lifespan:       60 * time.Second, // Medium lifetime for recovery
		clottingFactor: 75,               // Strong clotting/recovery ability
		healingType:    healingType,
		context7D: tranquilspeak.Context7D{
			Who:    "Platelet",
			What:   "System_Recovery_Cell",
			When:   time.Now().Unix(),
			Where:  "Recovery_Circulation_Pathway",
			Why:    "Provide_System_Healing_And_Recovery",
			How:    "Clotting_And_Healing_Mechanisms",
			Extent: "Target_" + targetSystem,
		},
	}
}

// RedBloodCell interface implementations

func (rbc *RedBloodCell) GetCellType() string {
	return rbc.cellType
}

func (rbc *RedBloodCell) GetTargetSystem() string {
	return rbc.targetSystem
}

func (rbc *RedBloodCell) GetATMTrigger() tranquilspeak.ATMTrigger {
	return rbc.atmTrigger
}

func (rbc *RedBloodCell) IsExpired() bool {
	return time.Since(rbc.birthTime) > rbc.lifespan
}

func (rbc *RedBloodCell) GetPriority() int {
	return tranquilspeak.PriorityAwareness // Normal priority for data transport
}

func (rbc *RedBloodCell) GetLifespan() time.Duration {
	return rbc.lifespan
}

// WhiteBloodCell interface implementations

func (wbc *WhiteBloodCell) GetCellType() string {
	return wbc.cellType
}

func (wbc *WhiteBloodCell) GetTargetSystem() string {
	return wbc.targetSystem
}

func (wbc *WhiteBloodCell) GetATMTrigger() tranquilspeak.ATMTrigger {
	return wbc.atmTrigger
}

func (wbc *WhiteBloodCell) IsExpired() bool {
	return time.Since(wbc.birthTime) > wbc.lifespan
}

func (wbc *WhiteBloodCell) GetPriority() int {
	return wbc.priority // Variable priority based on trigger type
}

func (wbc *WhiteBloodCell) GetLifespan() time.Duration {
	return wbc.lifespan
}

// Platelet interface implementations

func (p *Platelet) GetCellType() string {
	return p.cellType
}

func (p *Platelet) GetTargetSystem() string {
	return p.targetSystem
}

func (p *Platelet) GetATMTrigger() tranquilspeak.ATMTrigger {
	return p.atmTrigger
}

func (p *Platelet) IsExpired() bool {
	return time.Since(p.birthTime) > p.lifespan
}

func (p *Platelet) GetPriority() int {
	return tranquilspeak.PriorityInstinctual // High priority for recovery
}

func (p *Platelet) GetLifespan() time.Duration {
	return p.lifespan
}

// Blood cell utility functions

// CarryOxygen adds data "oxygen" to a red blood cell
func (rbc *RedBloodCell) CarryOxygen(data map[string]interface{}) {
	for key, value := range data {
		rbc.hemoglobin[key] = value
	}
	// Update oxygen level based on data load
	rbc.oxygenLevel = min(100, len(rbc.hemoglobin)*10)
}

// ReleaseOxygen releases data "oxygen" from a red blood cell
func (rbc *RedBloodCell) ReleaseOxygen() map[string]interface{} {
	oxygen := make(map[string]interface{})
	for key, value := range rbc.hemoglobin {
		oxygen[key] = value
	}
	// Clear hemoglobin after release
	rbc.hemoglobin = make(map[string]interface{})
	rbc.oxygenLevel = 0
	return oxygen
}

// ActivateAntibody activates immune response for white blood cell
func (wbc *WhiteBloodCell) ActivateAntibody(threat string) {
	wbc.antibodyType = threat
	wbc.priority = tranquilspeak.PriorityReactive // Increase priority for active threat
}

// StartClotting initiates recovery process for platelet
func (p *Platelet) StartClotting(errorSite string) {
	p.healingType = "error_clotting_" + errorSite
	p.clottingFactor = 100 // Maximum clotting strength
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
