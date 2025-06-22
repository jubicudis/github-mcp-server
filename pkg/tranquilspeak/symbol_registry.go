// Package tranquilspeak provides TranquilSpeak symbol and cluster lookup for TNOS logging/event systems.
// Auto-generated symbol registry parser for Go integration.
//
// This utility loads the symbolic_mapping_registry_autogen_20250603.tsq file and provides lookup functions.
// Update: Now loads from circulatory/github-mcp-server/symbolic_mapping_registry_autogen_20250603.tsq
package tranquilspeak

import (
	"bufio"
	"log"
	"os"
	"strings"
	"sync"
)

type SymbolEntry struct {
	Component string
	LangSym string
	OrganSym string
	ComponentSym string
	QuantumSym string
	FormulaSym string
	HieroglyphicSym string
	RegistrySym string
	Cluster string // Composite Symbol
}

var (
	symbolRegistry map[string]SymbolEntry
	once sync.Once
)

// LoadSymbolRegistry parses the .tsq registry file and populates the symbolRegistry map.
func LoadSymbolRegistry(path string) error {
	var err error
	once.Do(func() {
		symbolRegistry = make(map[string]SymbolEntry)
		f, e := os.Open(path)
		if e != nil {
			err = e
			return
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "|") { continue }
			fields := strings.Split(line, "|")
			if len(fields) < 22 { continue }
			component := strings.TrimSpace(fields[1])
			entry := SymbolEntry{
				Component:      component,
				LangSym:        strings.TrimSpace(fields[2]),
				OrganSym:       strings.TrimSpace(fields[3]),
				ComponentSym:   strings.TrimSpace(fields[4]),
				QuantumSym:     strings.TrimSpace(fields[5]),
				FormulaSym:     strings.TrimSpace(fields[6]),
				HieroglyphicSym: strings.TrimSpace(fields[7]),
				RegistrySym:    strings.TrimSpace(fields[8]),
				Cluster:        strings.TrimSpace(fields[21]),
			}
			symbolRegistry[component] = entry
		}
		// Manual aliasing for ATM AI deduplication and metadata logging clusters
		// Use actual ATM entries from the registry
		if atmEntry, ok := symbolRegistry["atm/trigger_matrix.tranquilspeak"]; ok {
			symbolRegistry["atm/ai.dedup"] = atmEntry
			symbolRegistry["atm/ai.meta"] = atmEntry
			symbolRegistry["advanced_trigger_matrix"] = atmEntry
		}
		// Ensure circulatory blood entry exists, create manual alias if missing
		if _, ok := symbolRegistry["circulatory/blood.tranquilspeak"]; !ok {
			if atmEntry, ok := symbolRegistry["atm/trigger_matrix.tranquilspeak"]; ok {
				symbolRegistry["circulatory/blood.tranquilspeak"] = atmEntry
			}
		}
		
		// Manual aliases for Quantum Handshake Protocol and Tesla-Aether-Mobius-Goldbach operations
		// QHP uses operational ATM entry
		if atmOperationalEntry, ok := symbolRegistry["atm/operational.tranquilspeak"]; ok {
			qhpEntry := atmOperationalEntry
			qhpEntry.Cluster = "🔐⚛️🤝♦Q"
			symbolRegistry["quantum_handshake_protocol"] = qhpEntry
		}
		
		// Tesla-Aether-Mobius-Goldbach uses helical store ATM entry  
		if atmHelicalEntry, ok := symbolRegistry["atm/helical_store.tranquilspeak"]; ok {
			teslaEntry := atmHelicalEntry
			teslaEntry.Cluster = "⚡🌀⧖𓂀♦φ∞"
			symbolRegistry["tesla_aether_mobius_goldbach"] = teslaEntry
		}
	})
	return err
}

// GetSymbolCluster returns the composite symbol cluster for a component/file.
func GetSymbolCluster(component string) string {
	if entry, ok := symbolRegistry[component]; ok {
		return entry.Cluster
	}
	return ""
}

// Example: GetSymbolCluster("circulatory/blood.tranquilspeak")

// Example: Practical usage of TranquilSpeak symbol cluster lookup in logging
//
// This function demonstrates how to use the symbol registry to tag log messages
// or events with the correct TranquilSpeak symbol cluster for a given component/file.
//
// Usage:
//   // LogWithSymbolCluster("circulatory/blood.tranquilspeak", "Blood event occurred")
//
// This will prepend the correct symbol cluster to the log message.

// LogWithSymbolCluster logs a message with the TranquilSpeak symbol cluster for the given component.
func LogWithSymbolCluster(component string, message string) {
	cluster := GetSymbolCluster(component)
	if cluster == "" {
		log.Printf("[NO_SYMBOL] %s", message)
	} else {
		log.Printf("[%s] %s", cluster, message)
	}
}
