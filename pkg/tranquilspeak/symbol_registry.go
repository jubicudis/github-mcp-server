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
//   LogWithSymbolCluster("circulatory/blood.tranquilspeak", "Blood event occurred")
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
