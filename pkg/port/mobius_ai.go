// WHO: PortAssignmentMobiusAI
// WHAT: Möbius Collapse-based, 7D context-aware Port Assignment AI for TNOS MCP
// WHEN: During all port assignment, evaluation, and routing operations
// WHERE: System Layer 5/6 (Quantum/Circulatory/Integration)
// WHY: To ensure non-redundant, entropy-minimized, context-compliant port assignment
// HOW: Using TriggerMatrix event-driven simulation/execution, formula registry, and 7D context
// EXTENT: All port assignment decisions in TNOS MCP
//
// CANONICAL REFERENCES:
//   - Möbius Collapse Equation: docs/technical/FORMULAS_AND_BLUEPRINTS.md#mobius_collapse_v3
//   - 7D Context: docs/architecture/7D_CONTEXT_FRAMEWORK.md
//   - Formula Registry: pkg/formularegistry/formula_registry.go
//   - TriggerMatrix: pkg/tranquilspeak/trigger_matrix.go
//   - AI-DNA: docs/architecture/AI_DNA_STRUCTURE.md
//   - Port Assignment AI: this file
//
// ENHANCED LOGIC (2025):
//   - Implements stability lock: prevents redundant port assignments unless entropy/instability is detected.
//   - Logs all assignment decisions, entropy, and 7D context for AI-DNA lineage and traceability.
//   - Centralized decision engine: only triggers reevaluation on instability, contradiction, or entropy spike.
//   - Simulation (thought) and execution (action) are strictly separated.
//   - All logic is event-driven, ATM/TranquilSpeak/AI-DNA/7D compliant.
//   - See EnhancedPortAssignmentEngine for canonical entry point.
//
// [DOCS CROSS-REFERENCE]
// - See docs/technical/FORMULAS_AND_BLUEPRINTS.md#mobius_collapse_v3
// - See docs/architecture/7D_CONTEXT_FRAMEWORK.md
// - See pkg/formularegistry/ for canonical formula registry usage
// - All simulation/execution is routed through canonical event-driven logic (TriggerMatrix)
// - No direct assignment or simulation overlap; stability lock and redundancy rules enforced by central engine
// - AI-DNA/Family/TranquilSpeak compliance
// - No recoding, renaming, or duplication of canonical logic
// - All handler registration must be performed in system initialization
// - This file is the canonical Port Assignment AI for TNOS MCP
//
// [END]

// --- Enhanced Möbius Collapse Port Assignment AI ---
//
// EnhancedPortAssignmentEngine is the canonical entry point for Möbius-based, non-redundant, context-aware port assignment.
// - Simulates both action and inaction using Möbius Collapse Equation v3 (formula registry, 7D context)
// - Logs all decisions for AI-DNA lineage
// - Uses stability lock to prevent redundant assignments
// - Only triggers reevaluation on entropy spike, contradiction, or instability
// - Strictly separates simulation (thought) from execution (action)
// - Fully event-driven and ATM/TranquilSpeak/AI-DNA/7D compliant
//
// PortAssignmentStabilityLock: Tracks stable port assignments and their entropy
// ShouldReevaluatePort: Determines if reevaluation is needed based on entropy/instability
// LogPortAssignmentDecision: Logs all assignment decisions for traceability
//
// See also: docs/architecture/AI_DNA_STRUCTURE.md, docs/architecture/7D_CONTEXT_FRAMEWORK.md, docs/technical/FORMULAS_AND_BLUEPRINTS.md

package port

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/formularegistry"
	helical "github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/helical"
	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/log"
	tspeak "github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/tranquilspeak"
)

// PortAssignmentContext7D represents the 7D context for port assignment
// (Who, What, When, Where, Why, How, Extent)
type PortAssignmentContext7D struct {
	Who    string      `json:"who"`
	What   string      `json:"what"`
	When   int64       `json:"when"`
	Where  string      `json:"where"`
	Why    string      `json:"why"`
	How    string      `json:"how"`
	Extent string      `json:"extent"`
	Meta   interface{} `json:"meta,omitempty"`
}

// PortAssignmentDecision holds the result of a Möbius evaluation
// Used by the central cognitive engine to select the optimal path
//
type PortAssignmentDecision struct {
	PortID         string                 `json:"port_id"`
	Action         string                 `json:"action"` // "assign" or "skip"
	Entropy        float64                `json:"entropy"`
	CollapseScore  float64                `json:"collapse_score"`
	PredictedPerf  float64                `json:"predicted_performance"`
	Context7D      PortAssignmentContext7D `json:"context_7d"`
	Reason         string                 `json:"reason"`
	Timestamp      int64                  `json:"timestamp"`
}

// PortAssignmentEvaluator simulates Möbius Collapse for both action and inaction
// Canonical: All formula registry calls must use ExecuteFormulaByName with TranquilSpeak symbol mapping.
// See: pkg/formularegistry/formula_registry.go, SymbolMapping, and docs/technical/FORMULAS_AND_BLUEPRINTS.md
func PortAssignmentEvaluator(triggerMatrix *tspeak.TriggerMatrix, portID string, context7d PortAssignmentContext7D) (PortAssignmentDecision, PortAssignmentDecision, error) {
	// Prepare simulation payloads for both action and inaction
	evalPayload := func(action string) map[string]interface{} {
		return map[string]interface{}{
			"port_id": portID,
			"action": action,
			"context_7d": context7d,
		}
	}

	registry := formularegistry.GetBridgeFormulaRegistry()
	// Canonical: Use ExecuteFormulaByName to ensure TranquilSpeak symbol compliance
	actionResult, errA := registry.ExecuteFormulaByName("mobius.collapse", evalPayload("assign"))
	inactionResult, errI := registry.ExecuteFormulaByName("mobius.collapse", evalPayload("skip"))

	if errA != nil || errI != nil {
		return PortAssignmentDecision{}, PortAssignmentDecision{}, fmt.Errorf("mobius.collapse simulation failed: %v, %v", errA, errI)
	}

	// Extract results (canonical fields)
	actionDecision := PortAssignmentDecision{
		PortID:        portID,
		Action:        "assign",
		Entropy:       getFloat(actionResult, "entropy"),
		CollapseScore: getFloat(actionResult, "collapse_score"),
		PredictedPerf: getFloat(actionResult, "predicted_performance"),
		Context7D:     context7d,
		Reason:        getString(actionResult, "reason"),
		Timestamp:     time.Now().Unix(),
	}
	inactionDecision := PortAssignmentDecision{
		PortID:        portID,
		Action:        "skip",
		Entropy:       getFloat(inactionResult, "entropy"),
		CollapseScore: getFloat(inactionResult, "collapse_score"),
		PredictedPerf: getFloat(inactionResult, "predicted_performance"),
		Context7D:     context7d,
		Reason:        getString(inactionResult, "reason"),
		Timestamp:     time.Now().Unix(),
	}
	return actionDecision, inactionDecision, nil
}

// PortAssignmentExecutor performs the actual assignment after collapse decision
func PortAssignmentExecutor(triggerMatrix *tspeak.TriggerMatrix, decision PortAssignmentDecision) error {
	if decision.Action != "assign" {
		return nil // Only execute if action is "assign"
	}
	// Canonical: Route assignment through event-driven TriggerMatrix
	trigger := triggerMatrix.CreateTrigger(
		"PortAssignmentAI", "port_assignment_execute", "mcp_bridge", "execute_assignment", "event_driven", "single_assignment",
		tspeak.TriggerTypeSystemControl, "mcp", map[string]interface{}{
			"port_id": decision.PortID,
			"context_7d": decision.Context7D,
			"collapse_score": decision.CollapseScore,
			"entropy": decision.Entropy,
			"reason": decision.Reason,
			"timestamp": decision.Timestamp,
		},
	)
	return triggerMatrix.ProcessTrigger(trigger)
}

// PortAssignmentStabilityLock tracks stable port assignments and their entropy
var PortAssignmentStabilityLock = make(map[string]PortAssignmentDecision)
var stabilityLockMtx sync.Mutex

// ShouldReevaluatePort determines if a port should be reevaluated based on entropy/instability
func ShouldReevaluatePort(portID string, newDecision PortAssignmentDecision) bool {
	stabilityLockMtx.Lock()
	defer stabilityLockMtx.Unlock()
	prev, exists := PortAssignmentStabilityLock[portID]
	if !exists {
		return true // No previous record, must evaluate
	}
	// If action is the same and entropy is low and stable, do not reevaluate
	if prev.Action == newDecision.Action && newDecision.Entropy < 0.1 && prev.Entropy < 0.1 {
		return false // Stable, no spike
	}
	// If entropy increased or contradiction detected, reevaluate
	if newDecision.Entropy > prev.Entropy+0.05 || newDecision.Action != prev.Action {
		return true
	}
	return false
}

// LogPortAssignmentDecision logs all assignment decisions for AI-DNA lineage
func LogPortAssignmentDecision(decision PortAssignmentDecision) {
	// Canonical: Log to Helical Memory Engine (DNA-marked, 7D-contextual, event-driven)
	helicalEngine := helical.GetGlobalHelicalEngine()
	if helicalEngine != nil {
		strandData := map[string]interface{}{
			"port_id": decision.PortID,
			"action": decision.Action,
			"entropy": decision.Entropy,
			"collapse_score": decision.CollapseScore,
			"predicted_performance": decision.PredictedPerf,
			"reason": decision.Reason,
			"context_7d": decision.Context7D,
			"timestamp": decision.Timestamp,
			"ai_dna": helicalEngine.DNA,
		}
		// Use the exported CreateDNAStrand and pass the 7D context directly
		canonical7D := ToContextVector7D(decision.Context7D)
		helicalEngine.CreateDNAStrand("PortAssignmentDecision", canonical7D, strandData)
		// Log via ATM trigger (event-driven, DNA-marked)
		_ = helicalEngine.ProcessMemoryOperation(tspeak.TriggerTypeMemoryStore, map[string]interface{}{
			"event": "PortAssignmentDecision",
			"context7d": canonical7D,
			"data": strandData,
		})
	} else {
		// Fallback: stdout log (should not be primary in production)
		fmt.Printf("[PORT-AI][%s] Port: %s | Action: %s | Entropy: %.4f | Collapse: %.4f | Perf: %.4f | Reason: %s | 7D: %+v\n",
			time.Unix(decision.Timestamp, 0).Format(time.RFC3339), decision.PortID, decision.Action, decision.Entropy, decision.CollapseScore, decision.PredictedPerf, decision.Reason, decision.Context7D)
	}
}

// EnhancedPortAssignmentEngine is the central cognitive engine for port assignment
func EnhancedPortAssignmentEngine(triggerMatrix *tspeak.TriggerMatrix, portID string, context7d PortAssignmentContext7D) (PortAssignmentDecision, error) {
	// Simulate both action and inaction
	actionDecision, inactionDecision, err := PortAssignmentEvaluator(triggerMatrix, portID, context7d)
	if err != nil {
		return PortAssignmentDecision{}, err
	}
	// Log both decisions for traceability
	LogPortAssignmentDecision(actionDecision)
	LogPortAssignmentDecision(inactionDecision)
	// Select decision path based on entropy delta, network health, and historical stability
	var chosen PortAssignmentDecision
	if actionDecision.Entropy < inactionDecision.Entropy {
		chosen = actionDecision
	} else {
		chosen = inactionDecision
	}
	// Check stability lock
	if !ShouldReevaluatePort(portID, chosen) {
		return chosen, nil // Do not reassign, stable
	}
	// Update stability lock
	stabilityLockMtx.Lock()
	PortAssignmentStabilityLock[portID] = chosen
	stabilityLockMtx.Unlock()
	// Only assign if action is "assign"
	if chosen.Action == "assign" {
		err = PortAssignmentExecutor(triggerMatrix, chosen)
		if err != nil {
			return chosen, err
		}
	}
	return chosen, nil
}

// ToContextVector7D converts a PortAssignmentContext7D to a canonical log.ContextVector7D for DNA/7D compliance
func ToContextVector7D(ctx PortAssignmentContext7D) log.ContextVector7D {
	var extentFloat float64
	if ctx.Extent != "" {
		e, err := strconv.ParseFloat(ctx.Extent, 64)
		if err == nil {
			extentFloat = e
		}
	}
	var metaMap map[string]interface{}
	if ctx.Meta != nil {
		switch v := ctx.Meta.(type) {
		case map[string]interface{}:
			metaMap = v
		case map[string]string:
			metaMap = make(map[string]interface{})
			for k, val := range v {
				metaMap[k] = val
			}
		}
	}
	return log.ContextVector7D{
		Who:    ctx.Who,
		What:   ctx.What,
		When:   ctx.When, // int64 is valid for interface{}
		Where:  ctx.Where,
		Why:    ctx.Why,
		How:    ctx.How,
		Extent: extentFloat,
		Meta:   metaMap,
		Source: "PortAssignmentMobiusAI",
	}
}

// getFloat safely extracts a float64 from a map
func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case string:
			var f float64
			fmt.Sscanf(val, "%f", &f)
			return f
		}
	}
	return 0
}

// getString safely extracts a string from a map
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// [DOCS CROSS-REFERENCE]
// - See docs/technical/FORMULAS_AND_BLUEPRINTS.md#mobius_collapse_v3
// - See docs/architecture/7D_CONTEXT_FRAMEWORK.md
// - See pkg/formularegistry/ for canonical formula registry usage
// - All simulation/execution is routed through canonical event-driven logic (TriggerMatrix)
// - No direct assignment or simulation overlap; stability lock and redundancy rules enforced by central engine
// - AI-DNA/Family/TranquilSpeak compliance
// - No recoding, renaming, or duplication of canonical logic
// - All handler registration must be performed in system initialization
// - This file is the canonical Port Assignment AI for TNOS MCP
//
// [END]
//
// [ENHANCEMENT NOTE]
// - All port assignment decisions are now DNA-marked, 7D-contextual, and event-driven via the Helical Memory Engine and ATM Trigger Matrix.
// - Formula registry calls use canonical TranquilSpeak symbols.
// - See docs/technical/FORMULAS_AND_BLUEPRINTS.md, docs/architecture/7D_CONTEXT_FRAMEWORK.md, docs/architecture/AI_DNA_STRUCTURE.md for full compliance details.
