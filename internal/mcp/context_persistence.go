// WHO: ContextManager
// WHAT: Context persistence implementation for MCP Bridge
// WHEN: During MCP context operations
// WHERE: Memory and storage interface between GitHub MCP and TNOS
// WHY: To maintain context across sessions and requests
// HOW: Using persistent storage with compression
// EXTENT: All 7D context vectors used in the MCP Bridge

package mcp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ContextPersistenceConfig defines configuration for context persistence
type ContextPersistenceConfig struct {
	// WHO: Configuration component
	// WHAT: Context persistence parameters
	StoragePath        string        // WHERE: Path to store context files
	PersistenceEnabled bool          // HOW: Enable/disable persistence
	SyncInterval       time.Duration // WHEN: Time between context syncs
	CompressionEnabled bool          // HOW: Enable Möbius compression for storage
}

// ContextEntry represents a stored context with metadata
type ContextEntry struct {
	// WHO: Context storage component
	// WHAT: Context entry representation
	Vector     ContextVector7D      `json:"vector"`      // WHAT: The 7D context vector
	Metadata   map[string]string    `json:"metadata"`    // WHAT: Associated metadata
	CreateTime time.Time            `json:"create_time"` // WHEN: Creation timestamp
	UpdateTime time.Time            `json:"update_time"` // WHEN: Last update timestamp
	Source     string               `json:"source"`      // WHERE: Context origin
	Usage      int                  `json:"usage"`       // EXTENT: Number of times used
	Tags       []string             `json:"tags"`        // HOW: Classification tags
	Compressed bool                 `json:"compressed"`  // HOW: Compression status
	Variables  map[string]float64   `json:"variables"`   // HOW: Compression variables
}

// ContextPersistence manages the storage and retrieval of context vectors
type ContextPersistence struct {
	// WHO: Context manager component
	// WHAT: Context persistence implementation
	config        ContextPersistenceConfig // HOW: Configuration parameters
	contextCache  map[string]*ContextEntry // WHERE: In-memory context cache
	mutex         sync.RWMutex             // HOW: Thread safety mechanism
	syncTimer     *time.Timer              // WHEN: Timer for synchronization
	contextStats  ContextStats             // WHAT: Usage statistics tracking
	sessionActive bool                     // WHEN: Active session indicator
}

// ContextStats tracks usage metrics for contexts
type ContextStats struct {
	// WHO: Statistics component
	// WHAT: Context usage metrics
	TotalContexts    int            // EXTENT: Total number of contexts
	AccessCount      int            // EXTENT: Number of context accesses
	UpdateCount      int            // EXTENT: Number of context updates
	CompressionRatio float64        // EXTENT: Average compression ratio
	LastSyncTime     time.Time      // WHEN: Last synchronization time
	DimensionCount   map[string]int // WHAT: Count of dimension access
}

// NewContextPersistence creates a new context persistence manager
func NewContextPersistence(config ContextPersistenceConfig) (*ContextPersistence, error) {
	// WHO: Context persistence factory
	// WHAT: Manager instance creation
	// WHEN: During system initialization
	// WHERE: Bridge system startup
	// WHY: To establish context storage
	// HOW: Using provided configuration
	// EXTENT: Single persistence manager

	// Set default values if needed
	if config.SyncInterval == 0 {
		config.SyncInterval = 5 * time.Minute
	}

	// If persistence is enabled, ensure storage path exists
	if config.PersistenceEnabled && config.StoragePath != "" {
		if err := os.MkdirAll(config.StoragePath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create storage directory: %w", err)
		}
	}

	// Initialize the persistence manager
	cp := &ContextPersistence{
		config:       config,
		contextCache: make(map[string]*ContextEntry),
		contextStats: ContextStats{
			DimensionCount: make(map[string]int),
			LastSyncTime:   time.Now(),
		},
		sessionActive: true,
	}

	// Load existing contexts if enabled
	if config.PersistenceEnabled {
		if err := cp.loadContexts(); err != nil {
			return nil, fmt.Errorf("failed to load contexts: %w", err)
		}
	}

	// Start synchronization timer if enabled
	if config.PersistenceEnabled && config.SyncInterval > 0 {
		cp.startSyncTimer()
	}

	return cp, nil
}

// startSyncTimer initiates periodic synchronization of contexts to storage
func (cp *ContextPersistence) startSyncTimer() {
	// WHO: Synchronization component
	// WHAT: Timer initialization 
	// WHEN: During manager creation
	// WHERE: Context persistence
	// WHY: To ensure regular persistence
	// HOW: Using timer-based triggers
	// EXTENT: Periodic synchronization

	if cp.syncTimer != nil {
		cp.syncTimer.Stop()
	}

	cp.syncTimer = time.AfterFunc(cp.config.SyncInterval, func() {
		if err := cp.SyncToStorage(); err != nil {
			fmt.Printf("Context sync failed: %v\n", err)
		}
		// Update stats
		cp.mutex.Lock()
		cp.contextStats.LastSyncTime = time.Now()
		cp.mutex.Unlock()
		
		// Schedule next sync if session is active
		if cp.sessionActive {
			cp.startSyncTimer()
		}
	})
}

// StoreContext saves a context vector with associated metadata
func (cp *ContextPersistence) StoreContext(id string, vector *ContextVector7D, source string, tags []string) error {
	// WHO: Storage component
	// WHAT: Context storage
	// WHEN: During context creation/update
	// WHERE: Memory cache and persistent storage
	// WHY: To preserve context for future use
	// HOW: Using structured storage format
	// EXTENT: Single context vector

	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	// Create new entry or update existing one
	now := time.Now()
	entry, exists := cp.contextCache[id]
	
	if !exists {
		// Create new entry
		entry = &ContextEntry{
			Vector:     *vector,
			Metadata:   make(map[string]string),
			CreateTime: now,
			UpdateTime: now,
			Source:     source,
			Usage:      0,
			Tags:       tags,
			Compressed: false,
			Variables:  make(map[string]float64),
		}
		cp.contextCache[id] = entry
		cp.contextStats.TotalContexts++
	} else {
		// Update existing entry
		entry.Vector = *vector
		entry.UpdateTime = now
		entry.Source = source
		entry.Tags = tags
		cp.contextStats.UpdateCount++
	}

	// Compress if enabled
	if cp.config.CompressionEnabled && !entry.Compressed {
		// Apply compression using Möbius Compression Formula
		// This would actually call the compression implementation
		entry.Compressed = true
		// Store compression variables for decompression
		entry.Variables = map[string]float64{
			"B": 1.5,
			"I": 0.8,
			"V": 2.0,
			"G": 1.0,
			"F": 0.5,
			"E": 0.3,
			"t": 0.1,
		}
	}

	// If persistence is enabled, immediately sync this context
	if cp.config.PersistenceEnabled {
		return cp.syncContextToStorage(id, entry)
	}

	return nil
}

// GetContext retrieves a stored context by ID
func (cp *ContextPersistence) GetContext(id string) (*ContextVector7D, error) {
	// WHO: Retrieval component
	// WHAT: Context retrieval
	// WHEN: During request processing
	// WHERE: From memory cache or storage
	// WHY: To access previously stored context
	// HOW: Using ID-based lookup
	// EXTENT: Single context vector

	cp.mutex.RLock()
	defer cp.mutex.RUnlock()

	// Look up in memory cache
	entry, exists := cp.contextCache[id]
	if !exists {
		return nil, fmt.Errorf("context not found: %s", id)
	}

	// Update usage stats
	entry.Usage++
	cp.contextStats.AccessCount++
	
	// Update dimension access statistics
	cp.trackDimensionAccess(entry.Vector)

	// Decompress if necessary
	if entry.Compressed {
		// Apply Möbius decompression using stored variables
		// This would actually call the decompression implementation
	}

	// Return a copy to avoid concurrent modification
	vectorCopy := entry.Vector
	return &vectorCopy, nil
}

// trackDimensionAccess updates statistics on which context dimensions are accessed
func (cp *ContextPersistence) trackDimensionAccess(vector ContextVector7D) {
	// WHO: Statistics component
	// WHAT: Dimension access tracking
	// WHEN: During context retrieval
	// WHERE: Internal statistics
	// WHY: To analyze usage patterns
	// HOW: Using counter increments
	// EXTENT: Dimension-level tracking

	// Increment counters for each dimension
	cp.contextStats.DimensionCount["Who"]++
	cp.contextStats.DimensionCount["What"]++
	cp.contextStats.DimensionCount["When"]++
	cp.contextStats.DimensionCount["Where"]++
	cp.contextStats.DimensionCount["Why"]++
	cp.contextStats.DimensionCount["How"]++
	cp.contextStats.DimensionCount["Extent"]++
}

// SyncToStorage persists all context entries to storage
func (cp *ContextPersistence) SyncToStorage() error {
	// WHO: Synchronization component
	// WHAT: Full context synchronization
	// WHEN: During scheduled sync
	// WHERE: From memory to persistent storage
	// WHY: To ensure durability of contexts
	// HOW: Using file-based storage
	// EXTENT: All context vectors

	if !cp.config.PersistenceEnabled {
		return nil
	}

	cp.mutex.RLock()
	defer cp.mutex.RUnlock()

	// Create a copy of contexts to avoid holding lock during I/O
	contextsCopy := make(map[string]*ContextEntry, len(cp.contextCache))
	for id, entry := range cp.contextCache {
		entryCopy := *entry
		contextsCopy[id] = &entryCopy
	}

	// Write each context to storage
	for id, entry := range contextsCopy {
		if err := cp.syncContextToStorage(id, entry); err != nil {
			return fmt.Errorf("failed to sync context %s: %w", id, err)
		}
	}

	return nil
}

// syncContextToStorage writes a single context to persistent storage
func (cp *ContextPersistence) syncContextToStorage(id string, entry *ContextEntry) error {
	// WHO: Storage component
	// WHAT: Individual context persistence
	// WHEN: During context update or sync
	// WHERE: To persistent storage
	// WHY: To ensure context durability
	// HOW: Using JSON serialization
	// EXTENT: Single context entry

	if !cp.config.PersistenceEnabled || cp.config.StoragePath == "" {
		return nil
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(cp.config.StoragePath, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Create file path for this context
	filePath := filepath.Join(cp.config.StoragePath, fmt.Sprintf("context_%s.json", id))

	// Serialize context to JSON
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	// Write to file
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}

	return nil
}

// loadContexts loads all persisted contexts from storage
func (cp *ContextPersistence) loadContexts() error {
	// WHO: Loading component
	// WHAT: Context retrieval from storage
	// WHEN: During initialization
	// WHERE: From persistent storage to memory
	// WHY: To restore previous contexts
	// HOW: Using file system traversal
	// EXTENT: All stored contexts

	if !cp.config.PersistenceEnabled || cp.config.StoragePath == "" {
		return nil
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(cp.config.StoragePath, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Read all context files
	files, err := ioutil.ReadDir(cp.config.StoragePath)
	if err != nil {
		return fmt.Errorf("failed to read storage directory: %w", err)
	}

	// Process each context file
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			// Read file
			filePath := filepath.Join(cp.config.StoragePath, file.Name())
			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read context file %s: %w", file.Name(), err)
			}

			// Parse JSON
			var entry ContextEntry
			if err := json.Unmarshal(data, &entry); err != nil {
				return fmt.Errorf("failed to unmarshal context file %s: %w", file.Name(), err)
			}

			// Extract ID from filename (context_ID.json)
			id := file.Name()
			id = id[8 : len(id)-5] // Remove "context_" prefix and ".json" suffix

			// Add to cache
			cp.mutex.Lock()
			cp.contextCache[id] = &entry
			cp.contextStats.TotalContexts++
			cp.mutex.Unlock()
		}
	}

	return nil
}

// GetStats returns the current context statistics
func (cp *ContextPersistence) GetStats() ContextStats {
	// WHO: Statistics provider
	// WHAT: Statistics retrieval
	// WHEN: During monitoring
	// WHERE: From internal tracking
	// WHY: To monitor system behavior
	// HOW: Using safe concurrent access
	// EXTENT: All statistical dimensions

	cp.mutex.RLock()
	defer cp.mutex.RUnlock()

	// Create a copy of the stats
	statsCopy := cp.contextStats
	statsCopy.DimensionCount = make(map[string]int)
	for k, v := range cp.contextStats.DimensionCount {
		statsCopy.DimensionCount[k] = v
	}

	return statsCopy
}

// Close shuts down the context persistence manager
func (cp *ContextPersistence) Close() error {
	// WHO: Lifecycle manager
	// WHAT: Persistence shutdown
	// WHEN: During system termination
	// WHERE: Context persistence manager
	// WHY: To ensure clean shutdown
	// HOW: Using final synchronization
	// EXTENT: Complete manager shutdown

	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	// Stop synchronization timer
	if cp.syncTimer != nil {
		cp.syncTimer.Stop()
		cp.syncTimer = nil
	}

	cp.sessionActive = false

	// Perform final sync if persistence is enabled
	if cp.config.PersistenceEnabled {
		return cp.SyncToStorage()
	}

	return nil
}
