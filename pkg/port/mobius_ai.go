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

package port

import (
	"fmt"
	"time"

	"github.com/jubicudis/Tranquility-Neuro-OS/github-mcp-server/pkg/formularegistry"
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
	// Use Möbius Collapse formula (canonical)
	mobiusSymbol, ok := formularegistry.FORMULA_SYMBOLS["mobius.collapse"]
	if !ok {
		return PortAssignmentDecision{}, PortAssignmentDecision{}, fmt.Errorf("mobius.collapse formula symbol not found")
	}

	// Simulate action path
	actionResult, errA := registry.ExecuteFormula(mobiusSymbol, evalPayload("assign"))
	// Simulate inaction path
	inactionResult, errI := registry.ExecuteFormula(mobiusSymbol, evalPayload("skip"))

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
