/*
 * WHO: BloodBridgeTypes
 * WHAT: Type definitions for blood integration
 * WHEN: During bridge operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To define blood integration types
 * HOW: Using Go type system
 * EXTENT: All blood integration operations
 */

package bridge

// Deprecated: This file is retained only for BloodCirculationState. BloodCell definition moved to blood.go

// BloodCirculationState represents the current state of the blood circulation system
// WHO: BloodBridgeTypes
// WHAT: State definitions for blood integration
// WHEN: During bridge operations
// WHERE: System Layer 6 (Integration)
// WHY: To define blood integration state
// HOW: Using Go type system
// EXTENT: All blood integration operations

type BloodCirculationState struct {
	CellCount        int      `json:"cell_count"`
	CirculationRate  float64  `json:"circulation_rate"`
	LastCirculation  int64    `json:"last_circulation"`
	SystemHealth     float64  `json:"system_health"`
	QHPFingerprints  []string `json:"qhp_fingerprints"`
	SecureChannels   int      `json:"secure_channels"`
	ActiveCells      int      `json:"active_cells"`
	ThreatDetections int      `json:"threat_detections"`
}
