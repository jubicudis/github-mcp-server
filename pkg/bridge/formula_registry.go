// filepath: pkg/bridge/formula_registry.go
// WHO: FormulaRegistry
// WHAT: Minimal formula registry for TNOS MCP/bridge
// WHEN: Server startup
// WHERE: System Layer 6 (Integration)
// WHY: To provide formula lookup and metadata
// HOW: Loads formulas from JSON file at startup
// EXTENT: All formula execution

package bridge

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type BridgeFormula struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type BridgeFormulaRegistry struct {
	formulas map[string]BridgeFormula
	mu       sync.RWMutex
}

var bridgeRegistryInstance *BridgeFormulaRegistry
var bridgeOnce sync.Once

// LoadBridgeFormulaRegistry loads formulas from a JSON file
func LoadBridgeFormulaRegistry(path string) error {
	var err error
	bridgeOnce.Do(func() {
		f, e := os.Open(path)
		if e != nil {
			err = fmt.Errorf("failed to open formula registry: %w", e)
			return
		}
		defer f.Close()
		var data struct {
			Formulas []BridgeFormula `json:"formulas"`
		}
		if e := json.NewDecoder(f).Decode(&data); e != nil {
			err = fmt.Errorf("failed to decode formula registry: %w", e)
			return
		}
		m := make(map[string]BridgeFormula)
		for _, formula := range data.Formulas {
			m[formula.ID] = formula
		}
		bridgeRegistryInstance = &BridgeFormulaRegistry{formulas: m}
	})
	return err
}

// GetBridgeFormulaRegistry returns the singleton registry
func GetBridgeFormulaRegistry() *BridgeFormulaRegistry {
	return bridgeRegistryInstance
}

// GetFormula returns a formula by ID
func (r *BridgeFormulaRegistry) GetFormula(id string) (BridgeFormula, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.formulas[id]
	return f, ok
}

// ListFormulas returns all formulas
func (r *BridgeFormulaRegistry) ListFormulas() []BridgeFormula {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]BridgeFormula, 0, len(r.formulas))
	for _, f := range r.formulas {
		out = append(out, f)
	}
	return out
}
