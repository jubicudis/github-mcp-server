// TRANQUILSPEAK SYMBOL CLUSTER: [𝒯❤️🔄∞φ𓂀♦SK5𝑾𝑾𝑯𝑾𝑯𝑬𝑾𝑯𝑬𝑹𝑾𝑯𝒀𝑯𝑶𝑾𝑬𝑿💉🌊𝒮𝓒𝓲𝓻𝓬]
// This file is part of the 'circulatory' biosystem. See circulatory/github-mcp-server/symbolic_mapping_registry_autogen_20250603.tsq for details.
//
/*WHO: BloodCirculatorySystem
* WHAT: Core circulatory system managing ATM trigger flow through biological components
* WHEN: Throughout system runtime - continuous circulation
* WHERE: System Layer 5 (Quantum/Circulatory) - central nervous system integration
* WHY: To provide universal communication medium between all biological systems via ATM
* HOW: Using blood cells (red/white/platelet) carrying ATM triggers with 7D context
* EXTENT: All biological systems - universal connector and communication fabric
 */

package blood

import (
	"fmt"
	"sync"
	"time"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/tranquilspeak"
)

// BloodCirculation represents the main circulatory system engine
// Mimics biological blood circulation - transports oxygen (data), nutrients (context),
// immune cells (security), and waste removal (error handling) throughout the organism
type BloodCirculation struct {
	// WHO: BloodCirculation managing system-wide communication
	// WHAT: ATM trigger transport and routing between biological systems  
	// WHEN: Continuous circulation during system runtime
	// WHERE: Quantum/Circulatory layer - connects all biological systems
	// WHY: Universal communication medium following biological patterns
	// HOW: Blood cells carrying ATM triggers through circulation pathways
	// EXTENT: All biological system communications
	
	// Circulation components
	heart          *Heart                    // Central pumping mechanism
	arteries       map[string]*Artery        // Outbound pathways to biological systems
	veins          map[string]*Vein          // Return pathways from biological systems
	capillaries    map[string]*Capillary     // Fine-grained system connections
	
	// Blood cell management
	redCells       chan *RedBloodCell        // Data transport cells
	whiteCells     chan *WhiteBloodCell      // Control signal cells
	platelets      chan *Platelet            // System recovery cells
	
	// ATM integration
	triggerMatrix  *tranquilspeak.TriggerMatrix  // ATM trigger management
	logger         log.LoggerInterface        // Logging for circulation events
	
	// Circulation state
	isCirculating  bool                       // Whether circulation is active
	circulation    sync.Mutex                // Thread safety for circulation
	heartbeat      *time.Ticker              // Circulation rhythm
	
	// 7D Context
	context7D      tranquilspeak.Context7D   // Current circulation context
}

// Heart represents the central pumping mechanism for blood circulation
// Mimics biological heart - maintains rhythm, pressure, and flow distribution
type Heart struct {
	// WHO: Heart coordinating circulation rhythm and pressure
	// WHAT: Central pumping and flow control for ATM triggers  
	// WHEN: Continuous pumping during active circulation
	// WHERE: Core of circulatory system
	// WHY: Maintain steady flow and pressure for optimal ATM trigger transport
	// HOW: Rhythmic pumping with pressure regulation and flow distribution
	// EXTENT: Entire circulatory system pressure and flow
	
	beatInterval   time.Duration             // Heart beat frequency
	pressure       int                       // Circulation pressure (priority handling)
	chambers       map[string]*HeartChamber  // Heart chambers for different blood types
	isBeating      bool                      // Whether heart is actively beating
	beatCount      int64                     // Total heartbeats since start
	
	context7D      tranquilspeak.Context7D   // Heart context
}

// HeartChamber represents different chambers handling different blood cell types
type HeartChamber struct {
	cellType       string                    // RED, WHITE, or PLATELET
	capacity       int                       // Maximum cells per beat
	currentLoad    int                       // Current cells in chamber
	pressure       int                       // Chamber pressure
}

// Artery represents outbound pathways carrying blood to biological systems
type Artery struct {
	// WHO: Artery transporting ATM triggers to target biological systems
	// WHAT: High-pressure outbound pathway for trigger delivery
	// WHEN: During trigger transport to target systems
	// WHERE: Between circulatory system and target biological systems
	// WHY: Efficient delivery of ATM triggers to their destinations
	// HOW: Pressurized flow with priority-based routing
	// EXTENT: Specific biological system connections
	
	targetSystem   string                    // Target biological system
	pathway        chan BloodCell            // Blood cell transport channel
	pressure       int                       // Arterial pressure
	isOpen         bool                      // Whether artery is open for flow
	flowRate       int                       // Cells per heartbeat
	
	context7D      tranquilspeak.Context7D   // Artery context
}

// Vein represents return pathways carrying blood back from biological systems  
type Vein struct {
	// WHO: Vein returning processed blood from biological systems
	// WHAT: Low-pressure return pathway for circulation completion
	// WHEN: After biological systems process ATM triggers
	// WHERE: Between biological systems and circulatory system
	// WHY: Complete circulation cycle and enable continuous flow
	// HOW: Low-pressure collection with waste removal
	// EXTENT: Return flow from all connected biological systems
	
	sourceSystem   string                    // Source biological system
	pathway        chan BloodCell            // Return blood cell channel
	pressure       int                       // Venous pressure (typically lower)
	isOpen         bool                      // Whether vein is open for return flow
	flowRate       int                       // Return cells per heartbeat
	
	context7D      tranquilspeak.Context7D   // Vein context
}

// Capillary represents fine-grained connections for detailed system interaction
type Capillary struct {
	// WHO: Capillary providing detailed biological system interface
	// WHAT: Fine-grained connection for precise ATM trigger exchange
	// WHEN: During detailed biological system interactions
	// WHERE: At the interface between circulation and biological system components
	// WHY: Enable precise, controlled exchange of triggers and responses
	// HOW: Thin-walled interface allowing selective permeability
	// EXTENT: Component-level connections within biological systems
	
	systemComponent string                   // Specific biological system component
	pathway        chan BloodCell            // Bidirectional exchange channel
	permeability   map[string]bool           // Which trigger types can pass
	isOpen         bool                      // Whether capillary is open
	
	context7D      tranquilspeak.Context7D   // Capillary context
}

// NewBloodCirculation creates a new circulatory system instance
func NewBloodCirculation(logger log.LoggerInterface) *BloodCirculation {
	triggerMatrix := tranquilspeak.NewTriggerMatrix()
	
	circulation := &BloodCirculation{
		heart:         NewHeart(50 * time.Millisecond), // 50ms heartbeat
		arteries:      make(map[string]*Artery),
		veins:         make(map[string]*Vein),
		capillaries:   make(map[string]*Capillary),
		redCells:      make(chan *RedBloodCell, 1000),
		whiteCells:    make(chan *WhiteBloodCell, 100),
		platelets:     make(chan *Platelet, 50),
		triggerMatrix: triggerMatrix,
		logger:        logger,
		isCirculating: false,
		context7D: tranquilspeak.Context7D{
			Who:    "BloodCirculation",
			What:   "Circulatory_System_Initialization",
			When:   time.Now().Unix(),
			Where:  "SystemLayer5_Quantum_Circulatory",
			Why:    "Initialize_Universal_ATM_Communication_Medium",
			How:    "Blood_Circulation_Engine_Creation",
			Extent: "All_Biological_System_Communications",
		},
	}
	
	// Register ATM trigger handlers for circulation management
	circulation.registerATMHandlers()
	
	// Log circulatory system initialization
	tranquilspeak.LogWithSymbolCluster("circulatory/blood.tranquilspeak", 
		"Blood circulation system initialized - ready for ATM trigger transport")
	
	return circulation
}

// NewHeart creates a new heart with specified beat interval
func NewHeart(beatInterval time.Duration) *Heart {
	return &Heart{
		beatInterval: beatInterval,
		pressure:     80, // Normal circulation pressure
		chambers: map[string]*HeartChamber{
			tranquilspeak.BloodCellRed: {
				cellType:    tranquilspeak.BloodCellRed,
				capacity:    100,
				currentLoad: 0,
				pressure:    80,
			},
			tranquilspeak.BloodCellWhite: {
				cellType:    tranquilspeak.BloodCellWhite,
				capacity:    20,
				currentLoad: 0,
				pressure:    90, // Higher pressure for control signals
			},
			tranquilspeak.BloodCellPlatelet: {
				cellType:    tranquilspeak.BloodCellPlatelet,
				capacity:    10,
				currentLoad: 0,
				pressure:    85, // Medium pressure for recovery
			},
		},
		isBeating: false,
		beatCount: 0,
		context7D: tranquilspeak.Context7D{
			Who:    "Heart",
			What:   "Circulation_Pumping_Engine",
			When:   time.Now().Unix(),
			Where:  "Core_Circulatory_System",
			Why:    "Maintain_Rhythmic_ATM_Trigger_Flow",
			How:    "Rhythmic_Pumping_With_Pressure_Control",
			Extent: "Entire_Circulation_Pressure_And_Flow",
		},
	}
}

// StartCirculation begins the blood circulation process
func (bc *BloodCirculation) StartCirculation() error {
	bc.circulation.Lock()
	defer bc.circulation.Unlock()
	
	if bc.isCirculating {
		return fmt.Errorf("circulation already active")
	}
	
	// Start the heart
	if err := bc.heart.StartBeating(); err != nil {
		return fmt.Errorf("failed to start heart: %w", err)
	}
	
	// Begin circulation rhythm
	bc.heartbeat = time.NewTicker(bc.heart.beatInterval)
	bc.isCirculating = true
	
	// Start circulation goroutine
	go bc.circulationLoop()
	
	// Create ATM trigger for circulation start
	trigger := tranquilspeak.CreateTrigger(
		"BloodCirculation",
		"Circulation_Started",
		"SystemLayer5_Quantum_Circulatory",
		"Initialize_System_Wide_ATM_Communication",
		"Blood_Flow_Activation",
		"All_Biological_Systems",
		tranquilspeak.TriggerTypeCirculatoryFlow,
		"All_Systems",
		map[string]interface{}{
			"circulation_status": "ACTIVE",
			"heart_rate":        bc.heart.beatInterval.String(),
			"pressure":          bc.heart.pressure,
		},
	)
	
	// Process the circulation start trigger
	if err := bc.triggerMatrix.ProcessTrigger(trigger); err != nil {
		bc.logger.Info("ATM trigger processing during circulation start", map[string]interface{}{
			"error": err.Error(),
		})
	}
	
	tranquilspeak.LogWithSymbolCluster("circulatory/blood.tranquilspeak", 
		fmt.Sprintf("Blood circulation started - heartbeat: %v, pressure: %d", 
			bc.heart.beatInterval, bc.heart.pressure))
	
	return nil
}

// StopCirculation stops the blood circulation process
func (bc *BloodCirculation) StopCirculation() error {
	bc.circulation.Lock()
	defer bc.circulation.Unlock()
	
	if !bc.isCirculating {
		return fmt.Errorf("circulation not active")
	}
	
	// Stop the heartbeat
	if bc.heartbeat != nil {
		bc.heartbeat.Stop()
	}
	
	// Stop the heart
	if err := bc.heart.StopBeating(); err != nil {
		return fmt.Errorf("failed to stop heart: %w", err)
	}
	
	bc.isCirculating = false
	
	tranquilspeak.LogWithSymbolCluster("circulatory/blood.tranquilspeak", "Blood circulation stopped")
	
	return nil
}

// StartBeating starts the heart pumping
func (h *Heart) StartBeating() error {
	if h.isBeating {
		return fmt.Errorf("heart already beating")
	}
	
	h.isBeating = true
	h.beatCount = 0
	
	return nil
}

// StopBeating stops the heart pumping  
func (h *Heart) StopBeating() error {
	if !h.isBeating {
		return fmt.Errorf("heart not beating")
	}
	
	h.isBeating = false
	
	return nil
}

// circulationLoop is the main circulation processing loop
func (bc *BloodCirculation) circulationLoop() {
	for bc.isCirculating {
		<-bc.heartbeat.C
		// Heartbeat - pump blood through circulation
		bc.heartbeatPump()
	}
}

// heartbeatPump performs one heartbeat circulation cycle
func (bc *BloodCirculation) heartbeatPump() {
	bc.heart.beatCount++
	
	// Process each type of blood cell
	bc.pumpRedCells()
	bc.pumpWhiteCells()
	bc.pumpPlatelets()
	
	// Log heartbeat (periodic logging to avoid spam)
	if bc.heart.beatCount%1000 == 0 {
		tranquilspeak.LogWithSymbolCluster("circulatory/blood.tranquilspeak", 
			fmt.Sprintf("Heartbeat #%d - circulation active", bc.heart.beatCount))
	}
}

// pumpRedCells pumps red blood cells (data transport) through circulation
func (bc *BloodCirculation) pumpRedCells() {
	// Pump red cells from chamber through arteries
redLoop:
	for {
		if len(bc.redCells) == 0 {
			break redLoop
		}
		select {
		case redCell := <-bc.redCells:
			if artery, exists := bc.arteries[redCell.targetSystem]; exists {
				select {
				case artery.pathway <- redCell:
					// Successfully pumped through artery
				default:
					// Artery blocked - recirculate
					select {
					case bc.redCells <- redCell:
					default:
						// Channel full - cell lost (simulate circulation overflow)
					}
				}
			}
		default:
			break redLoop
		}
	}
}

// pumpWhiteCells pumps white blood cells (control signals) through circulation
func (bc *BloodCirculation) pumpWhiteCells() {
	// Pump white cells from chamber through arteries (higher priority)
whiteLoop:
	for {
		if len(bc.whiteCells) == 0 {
			break whiteLoop
		}
		select {
		case whiteCell := <-bc.whiteCells:
			if artery, exists := bc.arteries[whiteCell.targetSystem]; exists {
				select {
				case artery.pathway <- whiteCell:
					// Successfully pumped through artery
				default:
					// Artery blocked - recirculate with priority
					select {
					case bc.whiteCells <- whiteCell:
					default:
						// Channel full - critical cell lost
					}
				}
			}
		default:
			break whiteLoop
		}
	}
}

// pumpPlatelets pumps platelets (system recovery) through circulation
func (bc *BloodCirculation) pumpPlatelets() {
	// Pump platelets from chamber through arteries
plateletLoop:
	for {
		if len(bc.platelets) == 0 {
			break plateletLoop
		}
		select {
		case platelet := <-bc.platelets:
			if artery, exists := bc.arteries[platelet.targetSystem]; exists {
				select {
				case artery.pathway <- platelet:
					// Successfully pumped through artery
				default:
					// Artery blocked - recirculate
					select {
					case bc.platelets <- platelet:
					default:
						// Channel full - recovery cell lost
					}
				}
			}
		default:
			break plateletLoop
		}
	}
}

// registerATMHandlers registers ATM trigger handlers for circulation management
func (bc *BloodCirculation) registerATMHandlers() {
	// Register circulation flow handler
	bc.triggerMatrix.RegisterTrigger(tranquilspeak.TriggerTypeCirculatoryFlow, 
		func(trigger tranquilspeak.ATMTrigger) error {
			return bc.handleCirculatoryFlowTrigger(trigger)
		})
	
	// Register system control handler
	bc.triggerMatrix.RegisterTrigger(tranquilspeak.TriggerTypeSystemControl, 
		func(trigger tranquilspeak.ATMTrigger) error {
			return bc.handleSystemControlTrigger(trigger)
		})
	
	// Register error recovery handler
	bc.triggerMatrix.RegisterTrigger(tranquilspeak.TriggerTypeErrorRecovery, 
		func(trigger tranquilspeak.ATMTrigger) error {
			return bc.handleErrorRecoveryTrigger(trigger)
		})
}

// handleCirculatoryFlowTrigger handles circulation flow ATM triggers
func (bc *BloodCirculation) handleCirculatoryFlowTrigger(trigger tranquilspeak.ATMTrigger) error {
	tranquilspeak.LogWithSymbolCluster("circulatory/blood.tranquilspeak", 
		fmt.Sprintf("Processing circulatory flow trigger from %s", trigger.Who))
	
	// Update circulation parameters based on trigger
	if newPressure, exists := trigger.Payload["pressure"]; exists {
		if pressure, ok := newPressure.(int); ok {
			bc.heart.pressure = pressure
		}
	}
	
	return nil
}

// handleSystemControlTrigger handles system control ATM triggers
func (bc *BloodCirculation) handleSystemControlTrigger(trigger tranquilspeak.ATMTrigger) error {
	tranquilspeak.LogWithSymbolCluster("circulatory/blood.tranquilspeak", 
		fmt.Sprintf("Processing system control trigger from %s", trigger.Who))
	
	// Handle system control commands
	if command, exists := trigger.Payload["command"]; exists {
		switch command {
		case "increase_pressure":
			bc.heart.pressure += 10
		case "decrease_pressure":
			bc.heart.pressure -= 10
		case "emergency_stop":
			return bc.StopCirculation()
		case "emergency_start":
			return bc.StartCirculation()
		}
	}
	
	return nil
}

// handleErrorRecoveryTrigger handles error recovery ATM triggers
func (bc *BloodCirculation) handleErrorRecoveryTrigger(trigger tranquilspeak.ATMTrigger) error {
	tranquilspeak.LogWithSymbolCluster("circulatory/blood.tranquilspeak", 
		fmt.Sprintf("Processing error recovery trigger from %s", trigger.Who))
	
	// Create recovery platelet
	platelet := NewPlatelet(trigger.TargetSystem, trigger, trigger.Payload)
	
	// Inject recovery platelet into circulation
	select {
	case bc.platelets <- platelet:
		return nil
	default:
		return fmt.Errorf("platelet circulation channel full - cannot inject recovery platelet")
	}
}
