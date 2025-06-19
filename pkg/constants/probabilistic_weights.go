/*
 * WHO: ProbabilisticWeights
 * WHAT: Quantum probabilistic weight constants for TNOS biological system decisions
 * WHEN: During system initialization and biological pattern calculations
 * WHERE: System Layer 1 (Instinctual) - Quantum probability layer
 * WHY: To provide quantum-inspired probabilistic weights for biological system mimicry
 * HOW: Using quantum mechanical probability distributions and biological patterns
 * EXTENT: All probabilistic decisions throughout TNOS biological systems
 */

package constants

import "math"

// TRANQUILSPEAK SYMBOL CLUSTER: [𝒯🦴ζℏƒ𓆑#SK1𝑾𝑾𝑯𝑾𝑯𝑬𝑾𝑯𝑬𝑹𝑾𝑯𝒀𝑯𝑶𝑾𝑬𝑿⏳📍𝒮𝓔𝓗]
// This file provides quantum probabilistic weights for biological system decisions

// BiologicalSystemWeights represents probabilistic weights for biological system interactions
type BiologicalSystemWeights struct {
	// WHO: BiologicalSystemWeights providing quantum probabilistic guidance
	// WHAT: Weight values for biological system decision making
	// WHEN: During ATM trigger processing and biological pattern recognition
	// WHERE: All biological systems (circulatory, nervous, immune, etc.)
	// WHY: To provide quantum-inspired decision making for biological mimicry
	// HOW: Quantum probability distributions mapped to biological patterns
	// EXTENT: All probabilistic decisions in biological system interactions

	// Circulatory System Weights (Blood Flow Patterns)
	CirculatoryBaseWeight     float64 // Base probability for circulatory system activation
	BloodFlowPriority        float64 // Priority multiplier for blood flow triggers
	OxygenTransportWeight    float64 // Weight for oxygen transport decisions
	NutrientDeliveryWeight   float64 // Weight for nutrient delivery patterns
	ToxinRemovalWeight       float64 // Weight for toxin removal prioritization

	// Nervous System Weights (Neural Pattern Processing)
	NeuralFiringThreshold    float64 // Threshold for neural firing patterns
	SynapticStrengthWeight   float64 // Weight for synaptic connection strength
	LearningAdaptationRate   float64 // Rate of learning adaptation in neural patterns
	MemoryConsolidationWeight float64 // Weight for memory consolidation processes
	ReflexResponseWeight     float64 // Weight for reflexive response patterns

	// Immune System Weights (Defense Pattern Recognition)
	ThreatDetectionSensitivity float64 // Sensitivity for threat detection
	ImmuneResponseWeight       float64 // Weight for immune response activation
	PathogenRecognitionWeight  float64 // Weight for pathogen pattern recognition
	InflammationControlWeight  float64 // Weight for inflammation control mechanisms
	AdaptiveImmunityWeight     float64 // Weight for adaptive immunity development

	// HemoFlux Compression Weights (Data Flow Optimization)
	CompressionEfficiencyWeight float64 // Weight for compression efficiency decisions
	DataIntegrityPriority      float64 // Priority for data integrity preservation
	QuantumCoherenceWeight     float64 // Weight for quantum coherence maintenance
	MobiusCompressionFactor    float64 // Factor for Möbius compression calculations
	EntropyMinimizationWeight  float64 // Weight for entropy minimization

	// ATM Trigger Weights (Event Processing Probabilities)
	TriggerPropagationWeight   float64 // Weight for trigger propagation decisions
	ResponseLatencyWeight      float64 // Weight for response timing optimization
	ContextPreservationWeight  float64 // Weight for 7D context preservation
	ErrorRecoveryWeight        float64 // Weight for error recovery mechanisms
	SystemAdaptationWeight     float64 // Weight for system adaptation responses
}

// DefaultBiologicalWeights returns the default quantum probabilistic weights
func DefaultBiologicalWeights() BiologicalSystemWeights {
	// WHO: DefaultBiologicalWeights providing quantum-inspired biological probabilities
	// WHAT: Default weight values based on quantum mechanical and biological patterns
	// WHEN: During system initialization and fallback scenarios
	// WHERE: All biological systems requiring probabilistic decision making
	// WHY: To provide scientifically-inspired defaults for biological system mimicry
	// HOW: Using golden ratio, Fibonacci sequences, and quantum probability distributions
	// EXTENT: System-wide default probabilistic behavior

	phi := (1.0 + math.Sqrt(5.0)) / 2.0 // Golden ratio - fundamental to biological patterns
	
	return BiologicalSystemWeights{
		// Circulatory System (Heart rhythm ~1Hz, blood flow optimization)
		CirculatoryBaseWeight:     0.618,           // Golden ratio for natural flow
		BloodFlowPriority:        0.8,             // High priority for life support
		OxygenTransportWeight:    0.95,            // Critical for system survival
		NutrientDeliveryWeight:   0.75,            // Important but not critical
		ToxinRemovalWeight:       0.85,            // High priority for system health

		// Nervous System (Neural firing ~40Hz gamma waves)
		NeuralFiringThreshold:    0.382,           // 1 - golden ratio for threshold
		SynapticStrengthWeight:   0.707,           // sqrt(2)/2 for balanced connections
		LearningAdaptationRate:   0.1,             // Gradual learning adaptation
		MemoryConsolidationWeight: 0.618,          // Golden ratio for memory formation
		ReflexResponseWeight:     0.9,             // Fast reflexive responses

		// Immune System (Adaptive response patterns)
		ThreatDetectionSensitivity: 0.8,          // High sensitivity for threats
		ImmuneResponseWeight:       0.75,          // Balanced immune activation
		PathogenRecognitionWeight:  0.9,           // High accuracy for pathogen ID
		InflammationControlWeight:  0.6,           // Moderate inflammation control
		AdaptiveImmunityWeight:     0.7,           // Strong adaptive immunity

		// HemoFlux Compression (Quantum optimization)
		CompressionEfficiencyWeight: 0.618,        // Golden ratio for efficiency
		DataIntegrityPriority:      0.95,          // Critical data preservation
		QuantumCoherenceWeight:     1.0 / phi,     // Quantum coherence factor
		MobiusCompressionFactor:    math.Pi / 4,   // π/4 for Möbius calculations
		EntropyMinimizationWeight:  0.8,           // High entropy reduction

		// ATM Trigger System (Event processing optimization)
		TriggerPropagationWeight:   0.7,           // Balanced trigger propagation
		ResponseLatencyWeight:      0.9,           // Low latency responses
		ContextPreservationWeight:  0.95,          // Critical context preservation
		ErrorRecoveryWeight:        0.85,          // Strong error recovery
		SystemAdaptationWeight:     0.65,          // Moderate adaptation rate
	}
}

// QuantumProbabilityAdjustment applies quantum-inspired probability adjustments
func QuantumProbabilityAdjustment(baseWeight float64, coherenceLevel float64) float64 {
	// WHO: QuantumProbabilityAdjustment applying quantum mechanics to biological decisions
	// WHAT: Quantum-inspired probability adjustment calculation
	// WHEN: During probabilistic decision making in biological systems
	// WHERE: All biological systems requiring quantum-enhanced decision making
	// WHY: To incorporate quantum mechanical principles into biological system behavior
	// HOW: Using quantum coherence and uncertainty principles
	// EXTENT: All probabilistic calculations throughout TNOS

	// Apply quantum coherence scaling (0.0 = classical, 1.0 = fully quantum)
	quantumFactor := 1.0 + (coherenceLevel * 0.272)  // e^(-1) for quantum scaling
	
	// Apply uncertainty principle (small random variation)
	uncertainty := 0.05 * (math.Sin(coherenceLevel * math.Pi * 2))
	
	adjusted := baseWeight * quantumFactor * (1.0 + uncertainty)
	
	// Ensure probability remains in valid range [0, 1]
	if adjusted > 1.0 {
		adjusted = 1.0
	}
	if adjusted < 0.0 {
		adjusted = 0.0
	}
	
	return adjusted
}

// BiologicalDecisionMatrix provides decision weights for biological system interactions
type BiologicalDecisionMatrix struct {
	// WHO: BiologicalDecisionMatrix orchestrating cross-system biological decisions
	// WHAT: Decision matrix for biological system interaction patterns
	// WHEN: During complex biological system interactions requiring multi-system coordination
	// WHERE: Cross-system biological interactions (e.g., immune-circulatory, nervous-endocrine)
	// WHY: To provide coordinated decision making across biological systems
	// HOW: Using biological interaction patterns and quantum-inspired calculations
	// EXTENT: All complex multi-system biological interactions

	SystemWeights BiologicalSystemWeights
	
	// Cross-system interaction weights
	CirculatoryToNervousWeight   float64 // Blood-brain barrier interactions
	NervousToImmuneWeight        float64 // Neuro-immune communication
	ImmuneToCirculatoryWeight    float64 // Immune cell circulation
	CirculatoryToEndocrineWeight float64 // Hormone circulation
	EndocrineToNervousWeight     float64 // Hormone-neural interactions
	
	// System priority levels (for resource allocation)
	SystemPriorities map[string]float64
}

// NewBiologicalDecisionMatrix creates a new decision matrix with default weights
func NewBiologicalDecisionMatrix() *BiologicalDecisionMatrix {
	// WHO: NewBiologicalDecisionMatrix creating quantum-biological decision framework
	// WHAT: New decision matrix instance with scientifically-inspired defaults
	// WHEN: During system initialization or decision matrix reset
	// WHERE: All biological systems requiring coordinated decision making
	// WHY: To provide a quantum-biological framework for system coordination
	// HOW: Using biological interaction patterns and quantum probability distributions
	// EXTENT: System-wide biological decision coordination

	return &BiologicalDecisionMatrix{
		SystemWeights: DefaultBiologicalWeights(),
		
		// Cross-system interaction weights (based on biological coupling strengths)
		CirculatoryToNervousWeight:   0.85,  // Strong blood-brain coupling
		NervousToImmuneWeight:        0.7,   // Moderate neuro-immune coupling
		ImmuneToCirculatoryWeight:    0.8,   // Strong immune-blood coupling
		CirculatoryToEndocrineWeight: 0.9,   // Strong hormone circulation coupling
		EndocrineToNervousWeight:     0.75,  // Moderate hormone-neural coupling
		
		// System priorities (for resource allocation conflicts)
		SystemPriorities: map[string]float64{
			"circulatory": 0.95,  // Highest priority - life support
			"nervous":     0.9,   // High priority - coordination
			"immune":      0.85,  // High priority - protection
			"endocrine":   0.75,  // Moderate priority - regulation
			"digestive":   0.7,   // Moderate priority - energy
			"respiratory": 0.9,   // High priority - oxygen
			"excretory":   0.6,   // Lower priority - waste management
			"reproductive": 0.4,  // Lowest priority - non-critical
			"muscular":    0.8,   // High priority - movement
			"skeletal":    0.5,   // Lower priority - structural support
		},
	}
}

// CalculateBiologicalDecision calculates the optimal decision based on biological patterns
func (dm *BiologicalDecisionMatrix) CalculateBiologicalDecision(
	systemType string, 
	context map[string]interface{}, 
	quantumCoherence float64,
) float64 {
	// WHO: CalculateBiologicalDecision computing quantum-biological decision probabilities
	// WHAT: Optimal decision calculation using biological patterns and quantum mechanics
	// WHEN: During biological system decision points requiring probabilistic analysis
	// WHERE: All biological systems requiring decision optimization
	// WHY: To provide scientifically-grounded decision making for biological system mimicry
	// HOW: Combining biological weights, quantum adjustments, and contextual factors
	// EXTENT: All biological decision points throughout TNOS

	basePriority, exists := dm.SystemPriorities[systemType]
	if !exists {
		basePriority = 0.5 // Default neutral priority
	}
	
	// Apply quantum probability adjustment
	adjustedPriority := QuantumProbabilityAdjustment(basePriority, quantumCoherence)
	
	// Apply contextual modifications based on system state
	if context != nil {
		if stress, ok := context["stress_level"].(float64); ok {
			// Biological systems respond to stress (fight-or-flight patterns)
			stressFactor := 1.0 + (stress * 0.2)
			adjustedPriority *= stressFactor
		}
		
		if energy, ok := context["energy_level"].(float64); ok {
			// Energy availability affects system performance
			energyFactor := 0.5 + (energy * 0.5)
			adjustedPriority *= energyFactor
		}
	}
	
	// Ensure result stays in valid probability range
	if adjustedPriority > 1.0 {
		adjustedPriority = 1.0
	}
	if adjustedPriority < 0.0 {
		adjustedPriority = 0.0
	}
	
	return adjustedPriority
}
