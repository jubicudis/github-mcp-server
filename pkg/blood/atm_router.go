/*
 * WHO: BloodCirculationManager
 * WHAT: ATM trigger routing and biological system integration for blood circulation
 * WHEN: During ATM trigger routing and biological system communication
 * WHERE: Circulatory system interfaces with all biological systems
 * WHY: To provide ATM trigger routing, system integration, and circulation management
 * HOW: Through artery/vein management, capillary connections, and ATM trigger routing
 * EXTENT: All biological system connections and ATM trigger routing
 */

package blood

import (
	"fmt"
	"time"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/identity"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/tranquilspeak"
)

// TRANQUILSPEAK SYMBOL CLUSTER: [𝒯🔄🌐∞φ𓂀♦SK5𝑾𝑾𝑯𝑾𝑯𝑬𝑾𝑯𝑬𝑹𝑾𝑯𝒀𝑯𝑶𝑾𝑬𝑿🩸🔗𝒮𝓡𝓸𝓾𝓽]
// This file is part of the 'circulatory' biosystem. See symbolic_mapping_registry_autogen_20250603.tsq for details.

// ConnectBiologicalSystem creates artery and vein connections to a biological system
func (bc *BloodCirculation) ConnectBiologicalSystem(systemName string, capacity int) error {
	// Create artery for outbound triggers to the system
	artery := &Artery{
		targetSystem: systemName,
		pathway:      make(chan BloodCell, capacity),
		pressure:     bc.heart.pressure,
		isOpen:       true,
		flowRate:     capacity / 10, // 10% of capacity per heartbeat
		context7D: tranquilspeak.Context7D{
			Who:    "Artery_" + systemName,
			What:   "Outbound_ATM_Trigger_Pathway",
			When:   time.Now().Unix(),
			Where:  "Arterial_Circulation_To_" + systemName,
			Why:    "Transport_ATM_Triggers_To_Biological_System",
			How:    "Pressurized_Blood_Flow",
			Extent: "System_" + systemName,
		},
	}
	
	// Create vein for return flow from the system
	vein := &Vein{
		sourceSystem: systemName,
		pathway:      make(chan BloodCell, capacity/2), // Smaller return capacity
		pressure:     bc.heart.pressure / 2,           // Lower venous pressure
		isOpen:       true,
		flowRate:     capacity / 20, // 5% of capacity per heartbeat
		context7D: tranquilspeak.Context7D{
			Who:    "Vein_" + systemName,
			What:   "Return_ATM_Response_Pathway",
			When:   time.Now().Unix(),
			Where:  "Venous_Circulation_From_" + systemName,
			Why:    "Return_Processed_Blood_From_Biological_System",
			How:    "Low_Pressure_Return_Flow",
			Extent: "System_" + systemName,
		},
	}
	
	// Store connections
	bc.arteries[systemName] = artery
	bc.veins[systemName] = vein
	
	tranquilspeak.LogWithSymbolCluster("circulatory/blood",
		fmt.Sprintf("Connected biological system: %s (arterial capacity: %d, venous capacity: %d)",
			systemName, capacity, capacity/2))
	
	return nil
}

// DisconnectBiologicalSystem removes artery and vein connections to a biological system
func (bc *BloodCirculation) DisconnectBiologicalSystem(systemName string) error {
	// Close and remove artery
	if artery, exists := bc.arteries[systemName]; exists {
		artery.isOpen = false
		close(artery.pathway)
		delete(bc.arteries, systemName)
	}
	
	// Close and remove vein
	if vein, exists := bc.veins[systemName]; exists {
		vein.isOpen = false
		close(vein.pathway)
		delete(bc.veins, systemName)
	}
	
	tranquilspeak.LogWithSymbolCluster("circulatory/blood",
		fmt.Sprintf("Disconnected biological system: %s", systemName))
	
	return nil
}

// SendATMTrigger sends an ATM trigger through the blood circulation to a target system
func (bc *BloodCirculation) SendATMTrigger(trigger tranquilspeak.ATMTrigger) error {
	// Validate circulation is active
	if !bc.isCirculating {
		return fmt.Errorf("blood circulation not active - cannot send ATM trigger")
	}
	
	// Validate target system connection
	if _, exists := bc.arteries[trigger.TargetSystem]; !exists {
		return fmt.Errorf("no arterial connection to target system: %s", trigger.TargetSystem)
	}
	
	// Create appropriate blood cell type based on trigger
	var bloodCell BloodCell
	var err error
	
	switch trigger.TriggerType {
	case tranquilspeak.TriggerTypeDataTransport, tranquilspeak.TriggerTypeContextUpdate,
		 tranquilspeak.TriggerTypeMemoryStore, tranquilspeak.TriggerTypeMemoryRetrieve:
		// Use red blood cell for data transport
		bloodCell = NewRedBloodCell(trigger.TargetSystem, trigger)
		
	case tranquilspeak.TriggerTypeSystemControl, tranquilspeak.TriggerTypeSecurityAlert,
		 tranquilspeak.TriggerTypePrioritySignal, tranquilspeak.TriggerTypeAuthentication:
		// Use white blood cell for control signals
		antibodyType := "control_signal"
		if trigger.TriggerType == tranquilspeak.TriggerTypeSecurityAlert {
			antibodyType = "security_threat"
		}
		bloodCell = NewWhiteBloodCell(trigger.TargetSystem, trigger, antibodyType)
		
	case tranquilspeak.TriggerTypeErrorRecovery, tranquilspeak.TriggerTypeSystemHealing,
		 tranquilspeak.TriggerTypeMaintenanceReq, tranquilspeak.TriggerTypeBackupReq:
		// Use platelet for recovery operations
		healingData := map[string]interface{}{
			"healing_type": "system_recovery",
			"priority":     trigger.Priority,
		}
		bloodCell = NewPlatelet(trigger.TargetSystem, trigger, healingData)
		
	default:
		// Default to red blood cell
		bloodCell = NewRedBloodCell(trigger.TargetSystem, trigger)
	}
	
	// Inject blood cell into appropriate circulation channel
	switch bloodCell.GetCellType() {
	case tranquilspeak.BloodCellRed:
		select {
		case bc.redCells <- bloodCell.(*RedBloodCell):
			// Successfully injected into red cell circulation
		default:
			return fmt.Errorf("red blood cell circulation channel full")
		}
		
	case tranquilspeak.BloodCellWhite:
		select {
		case bc.whiteCells <- bloodCell.(*WhiteBloodCell):
			// Successfully injected into white cell circulation
		default:
			return fmt.Errorf("white blood cell circulation channel full")
		}
		
	case tranquilspeak.BloodCellPlatelet:
		select {
		case bc.platelets <- bloodCell.(*Platelet):
			// Successfully injected into platelet circulation
		default:
			return fmt.Errorf("platelet circulation channel full")
		}
	}
	
	tranquilspeak.LogWithSymbolCluster("circulatory/blood",
		fmt.Sprintf("Injected %s cell carrying %s trigger to %s",
			bloodCell.GetCellType(), trigger.TriggerType, trigger.TargetSystem))
	
	return err
}

// ReceiveATMResponse receives an ATM response from a biological system through venous return
func (bc *BloodCirculation) ReceiveATMResponse(sourceSystem string) (tranquilspeak.ATMTrigger, error) {
	// Check if we have a venous connection from the source system
	vein, exists := bc.veins[sourceSystem]
	if !exists {
		return tranquilspeak.ATMTrigger{}, fmt.Errorf("no venous connection from source system: %s", sourceSystem)
	}
	
	// Check if vein is open
	if !vein.isOpen {
		return tranquilspeak.ATMTrigger{}, fmt.Errorf("venous connection from %s is closed", sourceSystem)
	}
	
	// Receive blood cell from venous return
	select {
	case bloodCell := <-vein.pathway:
		return bloodCell.GetATMTrigger(), nil
	default:
		return tranquilspeak.ATMTrigger{}, fmt.Errorf("no ATM response available from %s", sourceSystem)
	}
}

// CreateCapillaryConnection creates a fine-grained connection to a specific biological system component
func (bc *BloodCirculation) CreateCapillaryConnection(systemComponent string, permeability map[string]bool) error {
	capillary := &Capillary{
		systemComponent: systemComponent,
		pathway:         make(chan BloodCell, 10), // Small capacity for fine-grained exchange
		permeability:    permeability,
		isOpen:          true,
		context7D: tranquilspeak.Context7D{
			Who:    "Capillary_" + systemComponent,
			What:   "Fine_Grained_ATM_Exchange",
			When:   time.Now().Unix(),
			Where:  "Capillary_Network_" + systemComponent,
			Why:    "Precise_Component_Level_ATM_Communication",
			How:    "Selective_Permeability_Exchange",
			Extent: "Component_" + systemComponent,
		},
	}
	
	bc.capillaries[systemComponent] = capillary
	
	tranquilspeak.LogWithSymbolCluster("circulatory/blood",
		fmt.Sprintf("Created capillary connection to component: %s", systemComponent))
	
	return nil
}

// GetCirculationStats returns current circulation statistics
func (bc *BloodCirculation) GetCirculationStats() map[string]interface{} {
	stats := map[string]interface{}{
		"circulation_active": bc.isCirculating,
		"heart_rate":         bc.heart.beatInterval.String(),
		"heart_pressure":     bc.heart.pressure,
		"heartbeat_count":    bc.heart.beatCount,
		"red_cells_queue":    len(bc.redCells),
		"white_cells_queue":  len(bc.whiteCells),
		"platelets_queue":    len(bc.platelets),
		"arterial_connections": len(bc.arteries),
		"venous_connections": len(bc.veins),
		"capillary_connections": len(bc.capillaries),
	}
	
	// Add chamber statistics
	for cellType, chamber := range bc.heart.chambers {
		stats["chamber_"+cellType+"_load"] = chamber.currentLoad
		stats["chamber_"+cellType+"_capacity"] = chamber.capacity
		stats["chamber_"+cellType+"_pressure"] = chamber.pressure
	}
	
	return stats
}

// AdjustCirculationPressure adjusts the circulation pressure
func (bc *BloodCirculation) AdjustCirculationPressure(newPressure int) error {
	if newPressure < 10 || newPressure > 200 {
		return fmt.Errorf("circulation pressure must be between 10 and 200, got: %d", newPressure)
	}
	
	oldPressure := bc.heart.pressure
	bc.heart.pressure = newPressure
	
	// Update arterial pressures
	for _, artery := range bc.arteries {
		artery.pressure = newPressure
	}
	
	// Update venous pressures (typically half of arterial)
	for _, vein := range bc.veins {
		vein.pressure = newPressure / 2
	}
	
	// Update chamber pressures
	for _, chamber := range bc.heart.chambers {
		chamber.pressure = newPressure
		// White cells get higher pressure for priority
		if chamber.cellType == tranquilspeak.BloodCellWhite {
			chamber.pressure = newPressure + 10
		}
	}
	
	tranquilspeak.LogWithSymbolCluster("circulatory/blood",
		fmt.Sprintf("Adjusted circulation pressure from %d to %d", oldPressure, newPressure))
	
	return nil
}

// CleanupExpiredCells removes expired blood cells from circulation
func (bc *BloodCirculation) CleanupExpiredCells() {
	// This would typically be called by the spleen/lymphatic system
	// For now, we'll implement basic cleanup logic
	
	cleanupCount := 0
	
	// Cleanup expired red cells
	for len(bc.redCells) > 0 {
		var redCell *RedBloodCell
		select {
		case redCell = <-bc.redCells:
			if !redCell.IsExpired() {
				select {
				case bc.redCells <- redCell:
				default:
					// Channel full, cell lost
				}
			} else {
				cleanupCount++
			}
		default:
			// No more cells to process
			break
		}
		if len(bc.redCells) == 0 {
			break
		}
	}

	// Cleanup expired white cells
	for len(bc.whiteCells) > 0 {
		var whiteCell *WhiteBloodCell
		select {
		case whiteCell = <-bc.whiteCells:
			if !whiteCell.IsExpired() {
				select {
				case bc.whiteCells <- whiteCell:
				default:
					// Channel full, cell lost
				}
			} else {
				cleanupCount++
			}
		default:
			// No more cells to process
			break
		}
		if len(bc.whiteCells) == 0 {
			break
		}
	}

	// Cleanup expired platelets
	for len(bc.platelets) > 0 {
		var platelet *Platelet
		select {
		case platelet = <-bc.platelets:
			if !platelet.IsExpired() {
				select {
				case bc.platelets <- platelet:
				default:
					// Channel full, platelet lost
				}
			} else {
				cleanupCount++
			}
		default:
			// No more platelets to process
			break
		}
		if len(bc.platelets) == 0 {
			break
		}
	}
	
	if cleanupCount > 0 {
		tranquilspeak.LogWithSymbolCluster("circulatory/blood",
			fmt.Sprintf("Cleaned up %d expired blood cells", cleanupCount))
	}
}

// EmergencyCirculationStop immediately stops circulation for critical system protection
func (bc *BloodCirculation) EmergencyCirculationStop(reason string) error {
	tranquilspeak.LogWithSymbolCluster("circulatory/blood",
		fmt.Sprintf("EMERGENCY CIRCULATION STOP: %s", reason))
	
	// Create emergency stop trigger
	emergencyTrigger := tranquilspeak.CreateTrigger(
		"BloodCirculation",
		"Emergency_Circulation_Stop",
		"SystemLayer5_Quantum_Circulatory",
		"Critical_System_Protection",
		"Emergency_Shutdown_Protocol",
		"All_Biological_Systems",
		tranquilspeak.TriggerTypeSystemControl,
		"All_Systems",
		map[string]interface{}{
			"emergency_reason": reason,
			"shutdown_time":    time.Now().Unix(),
		},
	)
	emergencyTrigger.DNA_ID = identity.DNAInstance.Signature()
	
	// Process emergency trigger
	if err := bc.triggerMatrix.ProcessTrigger(emergencyTrigger); err != nil {
		bc.logger.Error("Failed to process emergency stop trigger", map[string]interface{}{
			"error": err.Error(),
		})
	}
	
	// Stop circulation
	return bc.StopCirculation()
}

// RestoreCirculationAfterEmergency restores circulation after emergency stop
func (bc *BloodCirculation) RestoreCirculationAfterEmergency() error {
	tranquilspeak.LogWithSymbolCluster("circulatory/blood",
		"Restoring circulation after emergency stop")
	
	// Create restoration trigger
	restorationTrigger := tranquilspeak.CreateTrigger(
		"BloodCirculation",
		"Circulation_Restoration",
		"SystemLayer5_Quantum_Circulatory",
		"System_Recovery_After_Emergency",
		"Emergency_Recovery_Protocol",
		"All_Biological_Systems",
		tranquilspeak.TriggerTypeSystemHealing,
		"All_Systems",
		map[string]interface{}{
			"restoration_time": time.Now().Unix(),
			"recovery_status":  "INITIATED",
		},
	)
	restorationTrigger.DNA_ID = identity.DNAInstance.Signature()
	
	// Process restoration trigger
	if err := bc.triggerMatrix.ProcessTrigger(restorationTrigger); err != nil {
		bc.logger.Error("Failed to process restoration trigger", map[string]interface{}{
			"error": err.Error(),
		})
	}
	
	// Start circulation
	return bc.StartCirculation()
}

// ATMRouter manages the routing of ATM triggers through the blood circulation system
type ATMRouter struct {
	// WHO: ATMRouter orchestrating ATM trigger flow through blood circulation
	// WHAT: ATM trigger routing and biological system coordination
	// WHEN: During all ATM trigger events requiring inter-system communication
	// WHERE: Blood circulation system - central routing layer
	// WHY: To provide centralized ATM trigger routing through biological circulation patterns
	// HOW: Using biological routing algorithms and quantum decision matrices
	// EXTENT: All ATM triggers flowing through the TNOS biological systems

	circulation    *BloodCirculation                     // Blood circulation system
	triggerMatrix  *tranquilspeak.TriggerMatrix         // ATM trigger system
	systemRoutes   map[string]string                    // System routing table
	routingContext string                               // Routing context identifier
}

// NewATMRouter creates a new ATMRouter instance
func NewATMRouter(circulation *BloodCirculation) *ATMRouter {
	// WHO: NewATMRouter creating ATM routing system for blood circulation
	// WHAT: New ATMRouter instance with biological routing capabilities
	// WHEN: During blood circulation system initialization
	// WHERE: Blood circulation routing layer initialization
	// WHY: To provide ATM trigger routing through biological circulation patterns
	// HOW: Using biological decision matrices and quantum routing algorithms
	// EXTENT: ATM routing system initialization

	router := &ATMRouter{
		circulation:    circulation,
		triggerMatrix:  tranquilspeak.NewTriggerMatrix(),
		systemRoutes:   make(map[string]string),
		routingContext: "atm_router",
	}

	// Initialize biological system routing table
	router.initializeSystemRoutes()

	// Register ATM trigger handlers for routing operations
	router.registerATMHandlers()

	return router
}

// initializeSystemRoutes initializes the biological system routing table
func (ar *ATMRouter) initializeSystemRoutes() {
	// WHO: initializeSystemRoutes setting up biological system routing patterns
	// WHAT: Initialization of routing table for biological systems
	// WHEN: During ATMRouter initialization
	// WHERE: ATM routing layer setup
	// WHY: To establish routing patterns based on biological system interactions
	// HOW: Using biological system mapping and circulation patterns
	// EXTENT: All biological system routing initialization

	// Biological system routing addresses (mimicking biological pathways)
	ar.systemRoutes = map[string]string{
		// Primary biological systems
		"circulatory":   "blood_circulation_core",
		"nervous":       "neural_pathway_network", 
		"immune":        "lymphatic_immune_network",
		"endocrine":     "hormone_circulation_network",
		"digestive":     "nutrient_absorption_network",
		"respiratory":   "oxygen_exchange_network",
		"excretory":     "waste_elimination_network",
		"reproductive":  "reproductive_hormone_network",
		"muscular":      "muscle_activation_network",
		"skeletal":      "structural_support_network",
		
		// Specialized subsystems
		"helical_memory": "memory_storage_helix",
		"context":        "context_orchestration_network",
		"hemoflux":       "compression_flow_network",
		"github":         "external_integration_network",
		
		// Cross-system integration points
		"blood_brain_barrier":    "circulatory_nervous_bridge",
		"neuro_immune":          "nervous_immune_bridge",
		"hormone_neural":        "endocrine_nervous_bridge",
		"immune_circulation":    "immune_circulatory_bridge",
	}
}

// registerATMHandlers registers ATM trigger handlers for routing operations
func (ar *ATMRouter) registerATMHandlers() {
	// WHO: registerATMHandlers setting up ATM trigger system for routing
	// WHAT: Registration of ATM trigger handlers for routing operations
	// WHEN: During ATMRouter initialization
	// WHERE: ATM routing layer setup
	// WHY: To enable event-driven routing operations through blood circulation
	// HOW: Using ATM trigger registration for circulation-based routing
	// EXTENT: All routing ATM trigger operations

	// Register routing trigger handler
	ar.triggerMatrix.RegisterTrigger(
		tranquilspeak.TriggerTypeDataTransport,
		ar.handleRoutingTrigger,
	)

	// Register system broadcast trigger handler
	ar.triggerMatrix.RegisterTrigger(
		tranquilspeak.TriggerTypePrioritySignal,
		ar.handleBroadcastTrigger,
	)
}

// handleRoutingTrigger handles standard ATM trigger routing operations
func (ar *ATMRouter) handleRoutingTrigger(trigger tranquilspeak.ATMTrigger) error {
	// WHO: handleRoutingTrigger processing standard ATM routing triggers
	// WHAT: Standard ATM trigger routing through blood circulation
	// WHEN: When standard ATM routing triggers are fired
	// WHERE: ATM routing layer trigger processing
	// WHY: To provide standard routing of triggers through biological circulation
	// HOW: Using biological routing patterns and blood cell transport
	// EXTENT: All standard ATM trigger routing operations

	tranquilspeak.LogWithSymbolCluster("circulatory/blood", "Processing standard ATM routing trigger")

	// Route through blood circulation
	return ar.circulation.SendATMTrigger(trigger)
}

// handleBroadcastTrigger handles ATM trigger broadcast operations
func (ar *ATMRouter) handleBroadcastTrigger(trigger tranquilspeak.ATMTrigger) error {
	// WHO: handleBroadcastTrigger processing ATM broadcast triggers
	// WHAT: ATM trigger broadcast to multiple biological systems
	// WHEN: When ATM broadcast triggers are fired
	// WHERE: ATM routing layer trigger processing
	// WHY: To provide system-wide broadcast of triggers through biological circulation
	// HOW: Using blood circulation to deliver triggers to all biological systems
	// EXTENT: All ATM trigger broadcast operations

	tranquilspeak.LogWithSymbolCluster("circulatory/blood", "Processing ATM broadcast trigger")

	// Broadcast to all biological systems through blood circulation
	for targetSystem := range ar.systemRoutes {
		if targetSystem != trigger.Who {
			broadcastTrigger := tranquilspeak.ATMTrigger{
				Who:           trigger.Who,
				What:          "BroadcastTrigger",
				When:          time.Now().Unix(),
				Where:         ar.routingContext,
				Why:           fmt.Sprintf("BroadcastTo%s", targetSystem),
				How:           "BloodCirculationBroadcast",
				Extent:        "SystemWideBroadcast",
				TriggerType:   tranquilspeak.TriggerTypeDataTransport,
				Priority:      trigger.Priority,
				TargetSystem:  targetSystem,
				Payload:       trigger.Payload,
				Compressed:    trigger.Compressed,
				DNA_ID:        trigger.DNA_ID,
			}

			err := ar.circulation.SendATMTrigger(broadcastTrigger)
			if err != nil {
				tranquilspeak.LogWithSymbolCluster("circulatory/blood",
					fmt.Sprintf("Failed to broadcast to %s: %v", targetSystem, err))
			}
		}
	}

	return nil
}

// RouteATMTrigger provides a high-level interface for routing ATM triggers
func (ar *ATMRouter) RouteATMTrigger(sourceSystem, targetSystem string, triggerContent interface{}) error {
	// WHO: RouteATMTrigger providing high-level ATM routing interface
	// WHAT: High-level ATM trigger routing operation
	// WHEN: When external systems need to route triggers through blood circulation
	// WHERE: ATM routing public interface
	// WHY: To provide a simple interface for ATM trigger routing operations
	// HOW: Using ATM triggers to initiate routing through blood circulation
	// EXTENT: All external ATM trigger routing requests

	if sourceSystem == "" || targetSystem == "" {
		return fmt.Errorf("source and target systems must be specified")
	}

	// Create ATM trigger for routing
	trigger := tranquilspeak.ATMTrigger{
		Who:           sourceSystem,
		What:          "RouteTrigger",
		When:          time.Now().Unix(),
		Where:         ar.routingContext,
		Why:           fmt.Sprintf("RouteFrom%sTo%s", sourceSystem, targetSystem),
		How:           "BloodCirculationRouting",
		Extent:        "TriggerTransport",
		TriggerType:   tranquilspeak.TriggerTypeDataTransport,
		Priority:      3, // Normal priority
		TargetSystem:  targetSystem,
		Payload:       map[string]interface{}{"content": triggerContent},
		Compressed:    false,
		DNA_ID:        "atm_router_001",
	}

	return ar.triggerMatrix.ProcessTrigger(trigger)
}

// BroadcastATMTrigger broadcasts an ATM trigger to all biological systems
func (ar *ATMRouter) BroadcastATMTrigger(sourceSystem string, triggerContent interface{}) error {
	// WHO: BroadcastATMTrigger providing high-level ATM broadcast interface
	// WHAT: High-level ATM trigger broadcast operation
	// WHEN: When external systems need to broadcast triggers to all biological systems
	// WHERE: ATM routing public interface
	// WHY: To provide a simple interface for system-wide ATM trigger broadcasting
	// HOW: Using ATM triggers to initiate broadcast through blood circulation
	// EXTENT: All external ATM trigger broadcast requests

	// Create ATM trigger for broadcast
	trigger := tranquilspeak.ATMTrigger{
		Who:           sourceSystem,
		What:          "BroadcastTrigger",
		When:          time.Now().Unix(),
		Where:         ar.routingContext,
		Why:           "SystemWideBroadcast",
		How:           "BloodCirculationBroadcast",
		Extent:        "AllBiologicalSystems",
		TriggerType:   tranquilspeak.TriggerTypePrioritySignal, // Use priority signal for broadcasts
		Priority:      5, // High priority for broadcasts
		TargetSystem:  "all_systems",
		Payload:       map[string]interface{}{"content": triggerContent},
		Compressed:    false,
		DNA_ID:        "atm_router_broadcast_001",
	}

	return ar.triggerMatrix.ProcessTrigger(trigger)
}

// GetSystemRoutes returns the current biological system routing table
func (ar *ATMRouter) GetSystemRoutes() map[string]string {
	// WHO: GetSystemRoutes providing routing table access
	// WHAT: Access to biological system routing table
	// WHEN: When external systems need to inspect routing configuration
	// WHERE: ATM routing public interface
	// WHY: To provide visibility into biological system routing patterns
	// HOW: Returning a copy of the current routing table
	// EXTENT: Routing table inspection operations

	// Return a copy to prevent external modification
	routes := make(map[string]string)
	for system, address := range ar.systemRoutes {
		routes[system] = address
	}
	return routes
}
