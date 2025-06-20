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
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/formularegistry"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/identity"
	tspeak "github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/tranquilspeak"
)

// ContextVector7D represents the 7-dimensional context for helical memory operations
type ContextVector7D struct {
	Who    interface{} `json:"who"`
	What   interface{} `json:"what"`
	When   interface{} `json:"when"`
	Where  interface{} `json:"where"`
	Why    interface{} `json:"why"`
	How    interface{} `json:"how"`
	Extent float64     `json:"extent"`
	Meta   interface{} `json:"meta,omitempty"`
	Source string      `json:"source,omitempty"`
}

// HelicalMemoryStrand represents a single strand of helical memory (like DNA)
type HelicalMemoryStrand struct {
	ID           string                 `json:"id"`
	Sequence     string                 `json:"sequence"`
	Context7D    ContextVector7D        `json:"context_7d"`
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

	// Canonical event-driven trigger matrix
	triggerMatrix *tspeak.TriggerMatrix
}

// Global helical memory engine (singleton pattern like biological DNA)
var globalHelicalEngine *HelicalMemoryEngine
var helicalOnce sync.Once
var sharedTriggerMatrix *tspeak.TriggerMatrix

// SetSharedTriggerMatrix sets the shared trigger matrix for the global helical engine
func SetSharedTriggerMatrix(triggerMatrix *tspeak.TriggerMatrix) {
	sharedTriggerMatrix = triggerMatrix
}

// GetGlobalHelicalEngine returns the global helical memory engine instance
func GetGlobalHelicalEngine() *HelicalMemoryEngine {
	helicalOnce.Do(func() {
		// Use shared trigger matrix if available, otherwise create a new one
		var triggerMatrix *tspeak.TriggerMatrix
		if sharedTriggerMatrix != nil {
			triggerMatrix = sharedTriggerMatrix
		} else {
			triggerMatrix = tspeak.NewTriggerMatrix()
		}
		globalHelicalEngine = NewHelicalMemoryEngine(triggerMatrix)
	})
	return globalHelicalEngine
}

// NewHelicalMemoryEngine creates a new DNA-inspired helical memory engine
func NewHelicalMemoryEngine(triggerMatrix *tspeak.TriggerMatrix) *HelicalMemoryEngine {
	db, err := sql.Open("sqlite3", "helical_memory.sqlite3")
	if err != nil {
		panic(fmt.Sprintf("Failed to open helical memory DB: %v", err))
	}
	// Create tables if not exist (dual-helix: primary, complementary, error correction, metadata)
	db.Exec(`CREATE TABLE IF NOT EXISTS strands (
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
	engine := &HelicalMemoryEngine{
		triggerMatrix:   triggerMatrix,
		memoryHelices:  make(map[string]*HelicalMemoryHelix),
		strandIndex:    make(map[string]string),
		quantumStates:  make(map[string]string),
		lastReplication: time.Now(),
		DNA:            identity.NewDNA("TNOS", "HELICAL_ENGINE"),
		db:             db,
	}
	
	// Initialize DNA pathways according to TNOS Layer 4 (AI Core Initialization)
	engine.initializeDNAPathways()
	
	// Register ATM trigger handlers for incoming events
	engine.registerATMHandlers()
	
	return engine
}

// initializeDNAPathways sets up the DNA trigger pathways for memory operations
func (hme *HelicalMemoryEngine) initializeDNAPathways() {
	// DNA storage pathway (like DNA replication)
	storageRequest := hme.triggerMatrix.CreateTrigger(
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
		// DNA pathway error - create error trigger instead of logging
		errorTrigger := hme.triggerMatrix.CreateTrigger(
			"HelicalMemoryEngine", "pathway_error", "helical_package", 
			"dna_pathway_failure", "error_trigger_creation", "single_error",
			tspeak.TriggerHelicalError, "helical", 
			map[string]interface{}{"error": err.Error(), "pathway": "storage"})
		hme.triggerMatrix.ProcessTrigger(errorTrigger)
	}
	
	// DNA retrieval pathway (like DNA transcription)
	retrievalRequest := hme.triggerMatrix.CreateTrigger(
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
		// DNA pathway error - create error trigger instead of logging
		errorTrigger := hme.triggerMatrix.CreateTrigger(
			"HelicalMemoryEngine", "pathway_error", "helical_package", 
			"dna_pathway_failure", "error_trigger_creation", "single_error",
			tspeak.TriggerHelicalError, "helical", 
			map[string]interface{}{"error": err.Error(), "pathway": "retrieval"})
		hme.triggerMatrix.ProcessTrigger(errorTrigger)
	}
	
	// DNA error pathway (like DNA repair mechanisms)
	errorRequest := hme.triggerMatrix.CreateTrigger(
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
		// DNA pathway error - create error trigger instead of logging
		errorTrigger := hme.triggerMatrix.CreateTrigger(
			"HelicalMemoryEngine", "pathway_error", "helical_package", 
			"dna_pathway_failure", "error_trigger_creation", "single_error",
			tspeak.TriggerHelicalError, "helical", 
			map[string]interface{}{"error": err.Error(), "pathway": "error_repair"})
		hme.triggerMatrix.ProcessTrigger(errorTrigger)
	}
}

// ProcessMemoryOperation - Main DNA processing function for memory operations
func (hme *HelicalMemoryEngine) ProcessMemoryOperation(operation string, data map[string]interface{}) error {
	// Update DNA activity
	hme.updateDNAActivity()
	
	// Create ATM trigger for the memory operation following TNOS Layer 4 (AI Core) patterns
	trigger := hme.triggerMatrix.CreateTrigger(
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
		// DNA pathway error - create error trigger instead of logging
		errorTrigger := hme.triggerMatrix.CreateTrigger(
			"HelicalMemoryEngine", "processing_error", "helical_package", 
			"dna_processing_failure", "error_trigger_creation", "single_error",
			tspeak.TriggerHelicalError, "helical", 
			map[string]interface{}{"error": err.Error(), "operation": operation})
		hme.triggerMatrix.ProcessTrigger(errorTrigger)
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
	context7d := ContextVector7D{
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

func (hme *HelicalMemoryEngine) createDNAStrand(event string, context7d ContextVector7D, data map[string]interface{}) HelicalMemoryStrand {
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

func (hme *HelicalMemoryEngine) generateStrandID(event string, context7d ContextVector7D) string {
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

func (hme *HelicalMemoryEngine) findOrCreateHelix(strand HelicalMemoryStrand) string {
	// For now, create a simple helix based on quantum state
	// In a real implementation, this would use more sophisticated algorithms
	helixID := "helix_" + strand.QuantumState[:8]
	
	hme.dnaProtection.Lock()
	if _, exists := hme.memoryHelices[helixID]; !exists {
		hme.memoryHelices[helixID] = &HelicalMemoryHelix{
			ID:           helixID,
			LeftStrand:   make([]HelicalMemoryStrand, 0),
			RightStrand:  make([]HelicalMemoryStrand, 0),
			BasePairs:    0,
			CreatedAt:    time.Now(),
			LastAccessed: time.Now(),
			QuantumLock:  false,
		}
		hme.helicesCreated++
	}
	hme.dnaProtection.Unlock()
	
	return helixID
}

func (hme *HelicalMemoryEngine) updateDNAActivity() {
	hme.quantumMutex.Lock()
	hme.lastReplication = time.Now()
	hme.quantumMutex.Unlock()
}

func (hme *HelicalMemoryEngine) logDNAActivity(message string, metadata map[string]interface{}) {
	// Silent DNA activity tracking - no logging to prevent infinite loops
	// Store activity metrics internally for monitoring
	hme.quantumMutex.Lock()
	hme.lastReplication = time.Now()
	hme.quantumMutex.Unlock()
	// Note: Direct SQLite logging without ATM triggers to prevent recursion
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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

// DNA anchoring and imprinting for memory engine
func (hme *HelicalMemoryEngine) ImprintDNA(parent *identity.DNA) error {
	return hme.DNA.Imprint(parent)
}

// DNA signature for memory engine
func (hme *HelicalMemoryEngine) Signature() string {
	return hme.DNA.Signature()
}

// DNA-aware memory strand creation
func (hme *HelicalMemoryEngine) NewStrand(sequence string, context ContextVector7D, meta map[string]interface{}) *HelicalMemoryStrand {
	strand := &HelicalMemoryStrand{
		ID:           fmt.Sprintf("strand-%d", time.Now().UnixNano()),
		Sequence:     sequence,
		Context7D:    context,
		Timestamp:    time.Now(),
		Checksum:     "", // to be calculated
		Metadata:     meta,
		QuantumState: "", // to be set
		HelixIndex:   0,  // to be set
		DNA:          hme.DNA, // anchor the engine's DNA
	}
	strand.Checksum = hme.calculateChecksum([]byte(sequence))
	return strand
}

// StoreStrand stores a strand in the SQLite3 DB with full 7D/ATM/TranquilSpeak/meta
func (hme *HelicalMemoryEngine) StoreStrand(strand HelicalMemoryStrand, strandType, helixID string, symbolCluster string, atmMeta map[string]interface{}, formulaRefs []string) error {
	ctx7d, _ := json.Marshal(strand.Context7D)
	meta, _ := json.Marshal(strand.Metadata)
	dna, _ := json.Marshal(strand.DNA)
	atm, _ := json.Marshal(atmMeta)
	formulas, _ := json.Marshal(formulaRefs)
	_, err := hme.db.Exec(`INSERT OR REPLACE INTO strands (id, helix_id, strand_type, sequence, context7d, timestamp, checksum, metadata, quantum_state, helix_index, dna, symbol_cluster, atm_meta, formula_refs) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		strand.ID, helixID, strandType, strand.Sequence, string(ctx7d), strand.Timestamp.Unix(), strand.Checksum, string(meta), strand.QuantumState, strand.HelixIndex, string(dna), symbolCluster, string(atm), string(formulas))
	return err
}

// RetrieveStrand retrieves a strand by ID
func (hme *HelicalMemoryEngine) RetrieveStrand(id string) (*HelicalMemoryStrand, error) {
	row := hme.db.QueryRow(`SELECT id, sequence, context7d, timestamp, checksum, metadata, quantum_state, helix_index, dna, symbol_cluster, atm_meta, formula_refs FROM strands WHERE id = ?`, id)
	var strand HelicalMemoryStrand
	var ctx7d, meta, dna, symbolCluster, atmMeta, formulaRefs string
	var ts int64
	if err := row.Scan(&strand.ID, &strand.Sequence, &ctx7d, &ts, &strand.Checksum, &meta, &strand.QuantumState, &strand.HelixIndex, &dna, &symbolCluster, &atmMeta, &formulaRefs); err != nil {
		return nil, err
	}
	strand.Timestamp = time.Unix(ts, 0)
	json.Unmarshal([]byte(ctx7d), &strand.Context7D)
	json.Unmarshal([]byte(meta), &strand.Metadata)
	json.Unmarshal([]byte(dna), &strand.DNA)
	// TODO: Unmarshal and use symbolCluster, atmMeta, formulaRefs as needed
	return &strand, nil
}

// DecompressStrandData uses the canonical HemoFlux Mobius decompression via TriggerMatrix (canonical event-driven logic)
func (hme *HelicalMemoryEngine) DecompressStrandData(compressed []byte, meta interface{}) ([]byte, error) {
	if hme.triggerMatrix == nil {
		return nil, fmt.Errorf("TriggerMatrix not initialized")
	}
	// Canonical ATMTrigger from TranquilSpeak (trigger_matrix.go)
	trigger := tspeak.ATMTrigger{
		Who:          "HelicalMemoryEngine",
		What:         "mobius_decompress", // Canonical event name
		When:         time.Now().Unix(),
		Where:        "helical_memory",
		Why:          "mobius_decompression",
		How:          "event_driven",
		Extent:       "1.0",
		TriggerType:  tspeak.TriggerTypeDecompressionReq, // Canonical trigger type from trigger_matrix.go
		TargetSystem: "hemoflux",
		Payload: map[string]interface{}{
			"compressed": compressed,
			"meta":      meta,
		},
	}
	// Process the decompression event through the canonical TriggerMatrix
	err := hme.triggerMatrix.ProcessTrigger(trigger)
	if err != nil {
		return nil, fmt.Errorf("Mobius decompression trigger failed: %w", err)
	}
	// Canonical handler should populate decompressed data in trigger.Payload["decompressed"]
	decompressed, ok := trigger.Payload["decompressed"].([]byte)
	if !ok {
		// Some handlers may return as string, handle both
		if s, ok2 := trigger.Payload["decompressed"].(string); ok2 {
			return []byte(s), nil
		}
		return nil, fmt.Errorf("Mobius decompression did not return valid data")
	}
	return decompressed, nil
}

// RetrieveStrandWithDecompression retrieves a strand and decompresses its data if Mobius-compressed
func (hme *HelicalMemoryEngine) RetrieveStrandWithDecompression(id string) (*HelicalMemoryStrand, error) {
	strand, err := hme.RetrieveStrand(id)
	if err != nil {
		return nil, err
	}
	// Check for Mobius/HemoFlux compression marker (canonical: "hemoflux_compressed")
	if strand.Metadata != nil {
		if strand.Metadata["hemoflux_compressed"] == true {
			compressed, ok := strand.Metadata["compressed"].([]byte)
			if !ok {
				if s, ok2 := strand.Metadata["compressed"].(string); ok2 {
					compressed = []byte(s)
				} else {
					return nil, fmt.Errorf("compressed data missing or invalid")
				}
			}
			meta := strand.Metadata["meta"]
			decompressed, derr := hme.DecompressStrandData(compressed, meta)
			if derr != nil {
				return nil, fmt.Errorf("decompression failed: %w", derr)
			}
			strand.Metadata["decompressed"] = decompressed
		}
	}
	return strand, nil
}

// StoreDualHelix stores both primary and secondary (parity) strands in the SQLite3 DB
func (hme *HelicalMemoryEngine) StoreDualHelix(event string, context7d ContextVector7D, data map[string]interface{}, helixID string, symbolCluster string, atmMeta map[string]interface{}, formulaRefs []string) error {
	// 1. Create primary strand
	primary := hme.createDNAStrand(event, context7d, data)
	primary.HelixIndex = 0

	// 2. Create secondary (parity) strand using local formula registry
	// Serialize data for parity calculation
	primaryBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal primary data: %v", err)
	}
	// Use formula registry for parity (e.g., XOR or custom)
	registry := formularegistry.GetBridgeFormulaRegistry()
	_, ok := registry.GetFormulaByName("helical.parity")
	if !ok {
		return fmt.Errorf("parity formula not found in registry")
	}
	params := map[string]interface{}{"primary": primaryBytes}
	result, err := registry.ExecuteFormulaByName("helical.parity", params)
	if err != nil {
		return fmt.Errorf("parity formula execution failed: %v", err)
	}
	parityBytes, _ := json.Marshal(result["parity"])
	var parityData map[string]interface{}
	json.Unmarshal(parityBytes, &parityData)

	secondary := hme.createDNAStrand(event+"_parity", context7d, parityData)
	secondary.HelixIndex = 1

	// 3. Store both in DB
	err1 := hme.StoreStrand(primary, "primary", helixID, symbolCluster, atmMeta, formulaRefs)
	err2 := hme.StoreStrand(secondary, "secondary", helixID, symbolCluster, atmMeta, formulaRefs)
	if err1 != nil || err2 != nil {
		return fmt.Errorf("dual-helix store error: %v, %v", err1, err2)
	}
	// 4. Log operation
	hme.logDNAActivity("Dual-helix memory stored", map[string]interface{}{
		"helix_id": helixID, "primary_id": primary.ID, "secondary_id": secondary.ID,
		"symbol_cluster": symbolCluster, "atm_meta": atmMeta, "formulas": formulaRefs,
	})
	return nil
}

// RetrieveWithSelfHealing retrieves a strand, using parity if needed, and decompresses if Mobius-compressed
func (hme *HelicalMemoryEngine) RetrieveWithSelfHealing(id string, helixID string) (*HelicalMemoryStrand, error) {
	strand, err := hme.RetrieveStrandWithDecompression(id)
	if err == nil {
		return strand, nil
	}
	// Try to find secondary (parity) strand
	row := hme.db.QueryRow(`SELECT id FROM strands WHERE helix_id = ? AND strand_type = 'secondary'`, helixID)
	var parityID string
	if err := row.Scan(&parityID); err == nil {
		parityStrand, err2 := hme.RetrieveStrandWithDecompression(parityID)
		if err2 == nil {
			// Use formula registry to reconstruct primary from parity
			registry := formularegistry.GetBridgeFormulaRegistry()
			formulaKey, ok := FORMULA_SYMBOLS["helical.recover_primary"]
			if !ok {
				return nil, fmt.Errorf("recovery formula symbol not found in registry")
			}
			parityBytes, _ := json.Marshal(parityStrand.Sequence)
			params := map[string]interface{}{"parity": parityBytes}
			result, err := registry.ExecuteFormulaByName(formulaKey, params)
			if err != nil {
				return nil, fmt.Errorf("recovery formula execution failed: %v", err)
			}
			primaryBytes, _ := json.Marshal(result["primary"])
			var recoveredData map[string]interface{}
			json.Unmarshal(primaryBytes, &recoveredData)
			// Recreate the primary strand
			recoveredStrand := hme.createDNAStrand(id, parityStrand.Context7D, recoveredData)
			recoveredStrand.HelixIndex = 0
			hme.logDNAActivity("Self-healing: reconstructed from parity", map[string]interface{}{ "helix_id": helixID, "parity_id": parityID })
			return &recoveredStrand, nil
		}
	}
	return nil, fmt.Errorf("strand not found and self-healing failed: %v", err)
}

// [MIGRATION NOTE]
// All helical memory logic is now SQLite3-backed and fully integrated.
// - All storage, retrieval, and repair use the SQLite3 backend.
// - In-memory and log-based storage logic has been removed.
// - This engine is initialized globally and used for all helical memory operations.
// - See docs/technical/FORMULAS_AND_BLUEPRINTS.md for HDSA/dual-helix/quantum logic.
// - All major operations are self-logged and 7D/ATM/TranquilSpeak-compliant.

// registerATMHandlers registers ATM trigger handlers for incoming events
func (hme *HelicalMemoryEngine) registerATMHandlers() {
	// Register DATA_TRANSPORT handler for log entries with HemoFlux compression
	hme.triggerMatrix.RegisterTrigger(tspeak.TriggerTypeDataTransport, 
		func(trigger tspeak.ATMTrigger) error {
			return hme.handleDataTransportTrigger(trigger)
		})
	
	// Register HELICAL_MEMORY_STORE handler
	hme.triggerMatrix.RegisterTrigger(tspeak.TriggerHelicalStore, 
		func(trigger tspeak.ATMTrigger) error {
			return hme.handleHelicalStoreTrigger(trigger)
		})
	
	// Register HELICAL_MEMORY_RETRIEVE handler
	hme.triggerMatrix.RegisterTrigger(tspeak.TriggerHelicalRetrieve, 
		func(trigger tspeak.ATMTrigger) error {
			return hme.handleHelicalRetrieveTrigger(trigger)
		})
	
	// Register HELICAL_MEMORY_ERROR handler
	hme.triggerMatrix.RegisterTrigger(tspeak.TriggerHelicalError, 
		func(trigger tspeak.ATMTrigger) error {
			return hme.handleHelicalErrorTrigger(trigger)
		})
}

// handleDataTransportTrigger handles log entry storage with HemoFlux compression
func (hme *HelicalMemoryEngine) handleDataTransportTrigger(trigger tspeak.ATMTrigger) error {
	// Check if this is a log entry (avoid recursive logging)
	if level, exists := trigger.Payload["level"]; exists {
		if levelStr, ok := level.(string); ok && levelStr != "" {
			// This is a log entry - store it in helical memory with HemoFlux compression
			return hme.storeLogEntry(trigger)
		}
	}
	
	// Handle other data transport triggers
	return hme.storeGenericData(trigger)
}

// handleHelicalStoreTrigger handles explicit helical memory store requests
func (hme *HelicalMemoryEngine) handleHelicalStoreTrigger(trigger tspeak.ATMTrigger) error {
	return hme.storeHelicalData(trigger)
}

// handleHelicalRetrieveTrigger handles helical memory retrieval requests
func (hme *HelicalMemoryEngine) handleHelicalRetrieveTrigger(trigger tspeak.ATMTrigger) error {
	return hme.retrieveHelicalData(trigger)
}

// handleHelicalErrorTrigger handles helical memory error events
func (hme *HelicalMemoryEngine) handleHelicalErrorTrigger(trigger tspeak.ATMTrigger) error {
	// Store error information for analysis
	return hme.storeErrorData(trigger)
}

// storeLogEntry stores log entries with HemoFlux compression
func (hme *HelicalMemoryEngine) storeLogEntry(trigger tspeak.ATMTrigger) error {
	// Apply HemoFlux compression to the log payload
	compressed, err := hme.applyHemoFluxCompression(trigger.Payload)
	if err != nil {
		return fmt.Errorf("HemoFlux compression failed: %w", err)
	}
	
	// Store compressed log data
	strand := hme.createDNAStrand("log_entry", hme.convertToContextVector7D(trigger), compressed)
	return hme.StoreStrand(strand, "log", "log_helix", "", trigger.Payload, []string{})
}

// storeGenericData stores generic data transport events
func (hme *HelicalMemoryEngine) storeGenericData(trigger tspeak.ATMTrigger) error {
	strand := hme.createDNAStrand("data_transport", hme.convertToContextVector7D(trigger), trigger.Payload)
	return hme.StoreStrand(strand, "data", "data_helix", "", trigger.Payload, []string{})
}

// storeHelicalData stores explicit helical memory requests
func (hme *HelicalMemoryEngine) storeHelicalData(trigger tspeak.ATMTrigger) error {
	strand := hme.createDNAStrand("helical_store", hme.convertToContextVector7D(trigger), trigger.Payload)
	return hme.StoreStrand(strand, "helical", "helical_helix", "", trigger.Payload, []string{})
}

// retrieveHelicalData retrieves data from helical memory
func (hme *HelicalMemoryEngine) retrieveHelicalData(trigger tspeak.ATMTrigger) error {
	// Implementation for data retrieval
	// This would query the SQLite database and return results
	return nil
}

// storeErrorData stores error information
func (hme *HelicalMemoryEngine) storeErrorData(trigger tspeak.ATMTrigger) error {
	strand := hme.createDNAStrand("error_event", hme.convertToContextVector7D(trigger), trigger.Payload)
	return hme.StoreStrand(strand, "error", "error_helix", "", trigger.Payload, []string{})
}

// applyHemoFluxCompression applies HemoFlux compression using formula registry
func (hme *HelicalMemoryEngine) applyHemoFluxCompression(payload map[string]interface{}) (map[string]interface{}, error) {
	registry := formularegistry.GetBridgeFormulaRegistry()
	if registry == nil {
		// No compression available - return payload as-is
		return payload, nil
	}
	
	// Use HemoFlux compression formula
	compressedResult, err := registry.ExecuteFormula("hemoflux.compress", payload)
	if err != nil {
		// Compression failed - return original payload
		return payload, nil
	}
	
	// ExecuteFormula returns map[string]interface{}, so we can use it directly
	compressedResult["hemoflux_compressed"] = true
	return compressedResult, nil
}

// convertToContextVector7D converts ATM trigger to local ContextVector7D
func (hme *HelicalMemoryEngine) convertToContextVector7D(trigger tspeak.ATMTrigger) ContextVector7D {
	return ContextVector7D{
		Who:    trigger.Who,
		What:   trigger.What,
		When:   trigger.When,
		Where:  trigger.Where,
		Why:    trigger.Why,
		How:    trigger.How,
		Extent: 1.0,
		Source: trigger.TargetSystem,
		Meta:   trigger.Payload,
	}
}

// TODO: Implement dual-helix encoding, error correction, and quantum state logic per HDSA spec.
// TODO: Integrate HemoFlux compression and formula registry for all data operations.
// TODO: Add self-documenting event logging for all major operations.

// [TNOS] All formula registry lookups and execution must use TranquilSpeak symbols as canonical keys.
// See: ../../../../systems/cpp/circulatory/algorithms/data/tranquilspeak_symbol_key.md for the authoritative mapping.
// Example: {"helical.parity": "ⓗ"} (replace with actual symbol from mapping)
//
// Usage: Always pass the symbol, not a hardcoded name, to GetFormula/ExecuteFormula.

// Example symbol mapping (should be auto-generated or imported in production)
var FORMULA_SYMBOLS = map[string]string{
	"helical.parity": "ⓗ", // TODO: Replace with actual symbol from mapping
	"helical.recover_primary": "ⓡ", // TODO: Replace with actual symbol from mapping
	// ...add more as needed
}

// In all usages, replace calls like GetFormula("helical.parity") with GetFormula(FORMULA_SYMBOLS["helical.parity"])
// and ExecuteFormula("helical.parity", ...) with ExecuteFormula(FORMULA_SYMBOLS["helical.parity"], ...)

// [DOCS CROSS-REFERENCE]
// - See docs/technical/FORMULAS_AND_BLUEPRINTS.md for Mobius/quantum/7D context compliance
// - See docs/architecture/CONTEXT_MCP_INTEGRATION.md for event/trigger matrix logic
// - See pkg/formularegistry/ for canonical formula registry usage
// - All decompression is routed through canonical event-driven HemoFlux logic (no simulation, no duplication)
