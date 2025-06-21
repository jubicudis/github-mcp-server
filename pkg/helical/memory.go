// TRANQUILSPEAK SYMBOL CLUSTER: [𝒯🦴ζℏƒ𓆑#SK1𝑾𝑾𝑯𝑾𝑯𝑬𝑾𝑯𝑬𝑹𝑾𝑯𝒀𝑯𝑶𝑾𝑬𝑿⏳📍𝒮𝓔𝓗]
// HEMOFLUX_FILE_ID: "_USERS_JUBICUDIS_TRANQUILITY-NEURO-OS_GITHUB-MCP-SERVER_PKG_HELICAL_MEMORY.GO"
// HEMOFLUX_FORMULA: helical.dna.memory
/*
 * WHO: HelicalMemoryEngine
 * WHAT: Biological DNA-inspired helical memory storage system with quantum resonance
 * WHEN: During all memory operations and data persistence across the system
 * WHERE: System Layer 4 (Quantum Layer) - DNA/Genetic System mimicry
 * WHY: To provide biological DNA-like memory storage with helical structure and quantum properties
 * HOW: Using helical data structures, quantum entanglement patterns, and ATM triggers
 * EXTENT: All memory operations throughout the TNOS ecosystem with biological fidelity
 */

package helical

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/identity"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log"
	tspeak "github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/tranquilspeak"
	_ "github.com/mattn/go-sqlite3"
)

// HelicalMemoryStrand represents a single strand of helical memory (like DNA)
type HelicalMemoryStrand struct {
	ID           string                 `json:"id"`
	Sequence     string                 `json:"sequence"`
	Context7D    log.ContextVector7D    `json:"context_7d"`
	Timestamp    time.Time              `json:"timestamp"`
	Checksum     string                 `json:"checksum"`
	Metadata     map[string]interface{} `json:"metadata"`
	QuantumState string                 `json:"quantum_state"`
	HelixIndex   int                    `json:"helix_index"`
	DNA          *identity.DNA          `json:"dna"` // AI-DNA identity anchor
}

// HelicalMemoryHelix represents a double helix structure containing memory strands
type HelicalMemoryHelix struct {
	ID            string                `json:"id"`
	LeftStrand    []HelicalMemoryStrand `json:"left_strand"`   // Left DNA strand
	RightStrand   []HelicalMemoryStrand `json:"right_strand"`  // Right DNA strand (complementary)
	BasePairs     int                   `json:"base_pairs"`    // Number of base pairs
	CreatedAt     time.Time             `json:"created_at"`    // Helix creation time
	LastAccessed  time.Time             `json:"last_accessed"` // Last access time
	QuantumLock   bool                  `json:"quantum_lock"`  // Quantum entanglement lock
}

// HelicalMemoryEngine - The biological DNA-inspired memory system
// WHO: Helical Memory Engine
// WHAT: DNA-like helical memory storage with quantum properties
// WHY: To provide biological memory storage that mimics DNA structure
// HOW: Through helical data structures, quantum mechanics, and ATM triggers
type HelicalMemoryEngine struct {
	// Biological DNA system components
	triggerMatrix    *tspeak.TriggerMatrix          // ATM trigger system
	logger          log.LoggerInterface             // Biological logging
	
	// DNA helix storage
	memoryHelices   map[string]*HelicalMemoryHelix  // All DNA helices
	strandIndex     map[string]string               // Strand ID to Helix ID mapping
	quantumStates   map[string]string               // Quantum entanglement states
	
	// Biological protection mechanisms
	dnaProtection   sync.RWMutex                    // DNA strand protection
	quantumMutex    sync.RWMutex                    // Quantum state protection
	
	// DNA metrics (biological monitoring)
	strandsStored   int64                           // Total strands stored
	helicesCreated  int64                           // Total helices created
	quantumOps      int64                           // Quantum operations count
	lastReplication time.Time                       // Last DNA replication
	
	DNA *identity.DNA // AI-DNA identity for the memory engine

	// SQLite3 database for helical memory storage
	db *sql.DB
}

// Global helical memory engine (singleton pattern like biological DNA)
var globalHelicalEngine *HelicalMemoryEngine
var helicalOnce sync.Once

// GetGlobalHelicalEngine returns the global helical memory engine instance
func GetGlobalHelicalEngine() *HelicalMemoryEngine {
	helicalOnce.Do(func() {
		globalHelicalEngine = NewHelicalMemoryEngine(nil)
	})
	return globalHelicalEngine
}

// NewHelicalMemoryEngine creates a new DNA-inspired helical memory engine
func NewHelicalMemoryEngine(logger log.LoggerInterface) *HelicalMemoryEngine {
	// Log the working directory for debugging
	wd, _ := os.Getwd()
	if logger != nil {
		logger.Info("[HelicalMemoryEngine] Working directory: %s", wd)
	}
	// Attempt to open DB and log absolute path
	// Use absolute path for DB to ensure correct location
	dbPath := "/Users/Jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/helical/helical_memory.sqlite3"
	absDBPath := dbPath
	if abs, err := filepath.Abs(dbPath); err == nil {
		absDBPath = abs
	}
	if logger != nil {
		logger.Info("[HelicalMemoryEngine] Attempting to open DB at: %s", absDBPath)
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		if logger != nil {
			logger.Error("Failed to open helical memory DB: %v", err)
		}
		panic(fmt.Sprintf("Failed to open helical memory DB: %v", err))
	}
	// Create tables if not exist (dual-helix: primary, complementary, error correction, metadata)
	_, tableErr := db.Exec(`CREATE TABLE IF NOT EXISTS strands (
		id TEXT PRIMARY KEY,
		helix_id TEXT,
		strand_type TEXT,
		sequence TEXT,
		context7d TEXT,
		timestamp INTEGER,
		checksum TEXT,
		metadata TEXT,
		quantum_state TEXT,
		helix_index INTEGER,
		dna TEXT,
		symbol_cluster TEXT,
		atm_meta TEXT,
		formula_refs TEXT
	);`)
	if tableErr != nil && logger != nil {
		logger.Error("Failed to create helical memory table: %v", tableErr)
	}
	engine := &HelicalMemoryEngine{
		triggerMatrix:   tspeak.NewTriggerMatrix(),
		logger:         logger,
		memoryHelices:  make(map[string]*HelicalMemoryHelix),
		strandIndex:    make(map[string]string),
		quantumStates:  make(map[string]string),
		lastReplication: time.Now(),
		DNA:            identity.NewDNA("TNOS", "HELICAL_ENGINE"),
		db:             db,
	}
	
	// Initialize DNA pathways (register ATM triggers)
	engine.initializeDNAPathways()
	
	// Log DNA system initialization
	engine.logDNAActivity("DNA Helical Memory Engine initialized (SQLite3 backend)", map[string]interface{}{
		"who":     "HelicalMemoryEngine",
		"what":    "dna_initialization",
		"db":      "sqlite3",
		"db_path": absDBPath,
	})
	
	return engine
}

// initializeDNAPathways sets up the DNA trigger pathways for memory operations
func (hme *HelicalMemoryEngine) initializeDNAPathways() {
	// DNA storage pathway (like DNA replication)
	storageRequest := tspeak.CreateTrigger(
		"HelicalMemoryEngine",                // WHO
		"dna_pathway_initialization",         // WHAT  
		"helical_package",                   // WHERE
		"setup_biological_memory_triggers",  // WHY
		"atm_trigger_registration",          // HOW
		"dna_storage_operations",            // EXTENT
		tspeak.TriggerHelicalStore,          // TriggerType
		"helical",                           // TargetSystem
		map[string]interface{}{
			"action": "register_pathway",
			"pathway_type": "dna_storage",
			"biological_process": "replication",
		}, // Payload
	)
	
	err := hme.triggerMatrix.ProcessTrigger(storageRequest)
	if err != nil {
		hme.logger.Error("Failed to initialize DNA storage pathway: %v", err)
	}
	
	// DNA retrieval pathway (like DNA transcription)
	retrievalRequest := tspeak.CreateTrigger(
		"HelicalMemoryEngine",                // WHO
		"dna_pathway_initialization",         // WHAT
		"helical_package",                   // WHERE
		"setup_biological_memory_triggers",  // WHY
		"atm_trigger_registration",          // HOW
		"dna_retrieval_operations",          // EXTENT
		tspeak.TriggerHelicalRetrieve,       // TriggerType
		"helical",                           // TargetSystem
		map[string]interface{}{
			"action": "register_pathway", 
			"pathway_type": "dna_retrieval",
			"biological_process": "transcription",
		}, // Payload
	)
	
	err = hme.triggerMatrix.ProcessTrigger(retrievalRequest)
	if err != nil {
		hme.logger.Error("Failed to initialize DNA retrieval pathway: %v", err)
	}
	
	// DNA error pathway (like DNA repair mechanisms)
	errorRequest := tspeak.CreateTrigger(
		"HelicalMemoryEngine",                // WHO
		"dna_pathway_initialization",         // WHAT
		"helical_package",                   // WHERE
		"setup_biological_memory_triggers",  // WHY
		"atm_trigger_registration",          // HOW
		"dna_error_repair_operations",       // EXTENT
		tspeak.TriggerHelicalError,          // TriggerType
		"helical",                           // TargetSystem
		map[string]interface{}{
			"action": "register_pathway",
			"pathway_type": "dna_error_repair", 
			"biological_process": "repair_mechanisms",
		}, // Payload
	)
	
	err = hme.triggerMatrix.ProcessTrigger(errorRequest)
	if err != nil {
		hme.logger.Error("Failed to initialize DNA error repair pathway: %v", err)
	}
}

// ProcessMemoryOperation - Main DNA processing function for memory operations
func (hme *HelicalMemoryEngine) ProcessMemoryOperation(operation string, data map[string]interface{}) error {
	// Update DNA activity
	hme.updateDNAActivity()
	
	// Create ATM trigger for the memory operation
	trigger := tspeak.CreateTrigger(
		"HelicalMemoryEngine",               // WHO
		"memory_operation",                  // WHAT
		"helical_package",                  // WHERE
		"process_biological_memory_request", // WHY
		"atm_trigger_processing",           // HOW
		"single_memory_operation",          // EXTENT
		operation,                          // TriggerType
		"helical",                          // TargetSystem
		data,                               // Payload
	)
	
	// Process through ATM trigger matrix
	err := hme.triggerMatrix.ProcessTrigger(trigger)
	
	// Update biological metrics
	hme.quantumMutex.Lock()
	hme.quantumOps++
	hme.quantumMutex.Unlock()
	
	if err != nil {
		hme.logger.Error("DNA pathway processing failed: %v", err)
		return fmt.Errorf("DNA pathway processing failed: %w", err)
	}
	
	return nil
}

// RecordMemory - Public interface for recording memory (like DNA synthesis)
func RecordMemory(event string, data map[string]interface{}) error {
	engine := GetGlobalHelicalEngine()
	if engine == nil {
		return fmt.Errorf("helical memory engine not initialized")
	}
	
	// Create 7D context for this memory
	context7d := log.ContextVector7D{
		Who:    "HelicalMemoryEngine",
		What:   event,
		When:   time.Now().Unix(),
		Where:  "helical_memory",
		Why:    "memory_storage",
		How:    "dna_synthesis",
		Extent: 1.0,
		Meta:   data,
		Source: "helical_memory_system",
	}
	
	// Store via DNA pathway
	err := engine.ProcessMemoryOperation(tspeak.TriggerHelicalStore, map[string]interface{}{
		"event":      event,
		"context7d":  context7d,
		"data":       data,
		"strand_type": "memory_record",
	})
	
	return err
}

// DNA utility functions - Core biological memory operations
// Note: ATM trigger processing now handled by TriggerMatrix.ProcessTrigger()
// These utility functions support the biological DNA operations

// CreateDNAStrand is the canonical, exported method for creating a DNA-marked helical memory strand (7D/AI-DNA/ATM compliant)
func (hme *HelicalMemoryEngine) CreateDNAStrand(event string, context7d log.ContextVector7D, data map[string]interface{}) HelicalMemoryStrand {
	// Create DNA sequence from data
	dataBytes, _ := json.Marshal(data)
	sequence := hme.generateDNASequence(dataBytes)
	
	// Create unique strand ID
	strandID := hme.generateStrandID(event, context7d)
	
	// Create checksum for integrity
	checksum := hme.calculateChecksum(dataBytes)
	
	// Determine quantum state
	quantumState := hme.generateQuantumState(strandID, sequence)
	
	return HelicalMemoryStrand{
		ID:           strandID,
		Sequence:     sequence,
		Context7D:    context7d,
		Timestamp:    time.Now(),
		Checksum:     checksum,
		Metadata:     map[string]interface{}{"event": event},
		QuantumState: quantumState,
		DNA:          hme.DNA, // anchor the engine's DNA
	}
}

func (hme *HelicalMemoryEngine) generateDNASequence(data []byte) string {
	// Convert data to DNA base pairs (A, T, G, C)
	sequence := ""
	for _, b := range data {
		switch b % 4 {
		case 0:
			sequence += "A"
		case 1:
			sequence += "T"
		case 2:
			sequence += "G"
		case 3:
			sequence += "C"
		}
	}
	return sequence
}

func (hme *HelicalMemoryEngine) generateStrandID(event string, context7d log.ContextVector7D) string {
	data := fmt.Sprintf("%s-%s-%v-%d", event, context7d.Who, context7d.What, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16] // 16 character ID
}

func (hme *HelicalMemoryEngine) calculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (hme *HelicalMemoryEngine) generateQuantumState(strandID, sequence string) string {
	// Simple quantum state based on strand properties
	quantumData := strandID + sequence
	hash := sha256.Sum256([]byte(quantumData))
	return hex.EncodeToString(hash[:])[:8] // 8 character quantum state
}

// updateDNAActivity updates the last DNA replication timestamp
func (hme *HelicalMemoryEngine) updateDNAActivity() {
	hme.quantumMutex.Lock()
	hme.lastReplication = time.Now()
	hme.quantumMutex.Unlock()
}

// logDNAActivity logs a DNA-related message with metadata
func (hme *HelicalMemoryEngine) logDNAActivity(message string, metadata map[string]interface{}) {
	if hme.logger != nil {
		metadataStr := fmt.Sprintf("metadata: %+v", metadata)
		hme.logger.Info("HelicalMemoryEngine: %s | %s", message, metadataStr)
	}
}

// GetDNAMetrics returns current DNA system metrics
func (hme *HelicalMemoryEngine) GetDNAMetrics() map[string]interface{} {
	hme.dnaProtection.RLock()
	helixCount := len(hme.memoryHelices)
	strandCount := len(hme.strandIndex)
	hme.dnaProtection.RUnlock()

	hme.quantumMutex.RLock()
	defer hme.quantumMutex.RUnlock()

	return map[string]interface{}{
		"strands_stored":      hme.strandsStored,
		"helices_created":     hme.helicesCreated,
		"active_helices":      helixCount,
		"active_strands":      strandCount,
		"quantum_operations":  hme.quantumOps,
		"last_replication":    hme.lastReplication.Unix(),
		"biological_system":   "dna_memory_system",
	}
}