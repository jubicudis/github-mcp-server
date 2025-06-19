/*
 * WHO: HemoFluxNeuralBridgeHelpers
 * WHAT: Helper functions for AI-enhanced HemoFlux neural bridge operations
 * WHEN: During neural network processing and AI-enhanced compression operations
 * WHERE: System Layer 5 (Quantum/Circulatory) - HemoFlux AI Integration Support
 * WHY: To provide utility functions for neural network operations and AI integration
 * HOW: Using mathematical functions, data processing utilities, and biological algorithms
 * EXTENT: Supporting functions for all HemoFlux neural bridge operations
 */

package hemoflux

import (
	"math"
	"time"
)

// TRANQUILSPEAK SYMBOL CLUSTER: [𝒯❤️🧠🔧⚡∞φ𓂀♦SK5𝑾𝑾𝑯𝑾𝑯𝑬𝑾𝑯𝑬𝑹𝑾𝑯𝒀𝑯𝑶𝑾𝑬𝑿🌊⚛️𝒮𝓗𝓮𝓵𝓹]
// This file is part of the 'circulatory' biosystem with neural AI helpers. See symbolic_mapping_registry_autogen_20250603.tsq for details.

// ActivationFunction represents different neural activation functions
type ActivationFunction int

const (
	ActivationSigmoid ActivationFunction = iota
	ActivationTanh
	ActivationReLU
	ActivationLeakyReLU
	ActivationSwish    // Biologically inspired activation
	ActivationSoftmax
)

// BiologicalConstants holds constants that mimic biological neural properties
var BiologicalConstants = struct {
	RestingPotential    float64 // Typical neuron resting potential (-70mV normalized)
	ActionThreshold     float64 // Action potential threshold (-55mV normalized)
	RefractoryPeriod    time.Duration // Refractory period duration
	SynapticDelay       time.Duration // Synaptic transmission delay
	PlasticityDecay     float64 // Synaptic plasticity decay rate
	InhibitoryRatio     float64 // Ratio of inhibitory to excitatory neurons
	MaxSynapticWeight   float64 // Maximum synaptic weight
	MinSynapticWeight   float64 // Minimum synaptic weight
}{
	RestingPotential:    -0.2,  // Normalized from -70mV
	ActionThreshold:     0.1,   // Normalized from -55mV
	RefractoryPeriod:    2 * time.Millisecond,
	SynapticDelay:       500 * time.Microsecond, // 0.5ms as microseconds
	PlasticityDecay:     0.95,
	InhibitoryRatio:     0.2,   // 20% inhibitory neurons (biological ratio)
	MaxSynapticWeight:   1.0,
	MinSynapticWeight:   -1.0,
}

// Activate applies the specified activation function
func Activate(input float64, function ActivationFunction) float64 {
	switch function {
	case ActivationSigmoid:
		return 1.0 / (1.0 + math.Exp(-input))
		
	case ActivationTanh:
		return math.Tanh(input)
		
	case ActivationReLU:
		if input > 0 {
			return input
		}
		return 0
		
	case ActivationLeakyReLU:
		if input > 0 {
			return input
		}
		return 0.01 * input
		
	case ActivationSwish:
		// Swish activation: x * sigmoid(x) - biologically inspired
		return input / (1.0 + math.Exp(-input))
		
	case ActivationSoftmax:
		// For softmax, this should be applied to a vector, but for single value:
		return math.Exp(input)
		
	default:
		return Activate(input, ActivationSigmoid)
	}
}

// ActivateDerivative computes the derivative of the activation function for backpropagation
func ActivateDerivative(output float64, function ActivationFunction) float64 {
	switch function {
	case ActivationSigmoid:
		return output * (1.0 - output)
		
	case ActivationTanh:
		return 1.0 - output*output
		
	case ActivationReLU:
		if output > 0 {
			return 1.0
		}
		return 0
		
	case ActivationLeakyReLU:
		if output > 0 {
			return 1.0
		}
		return 0.01
		
	case ActivationSwish:
		sigmoid := 1.0 / (1.0 + math.Exp(-output))
		return sigmoid + output*sigmoid*(1.0-sigmoid)
		
	case ActivationSoftmax:
		return output * (1.0 - output)
		
	default:
		return ActivateDerivative(output, ActivationSigmoid)
	}
}

// ApplySynapticPlasticity applies Hebbian learning to improve neural bridge performance
func ApplySynapticPlasticity(weights [][]float64, inputs, outputs []float64, learningRate float64) {
	// Apply Hebbian learning: weights that fire together wire together
	for i := 0; i < len(weights); i++ {
		for j := 0; j < len(weights[i]); j++ {
			if j < len(inputs) && i < len(outputs) {
				// Hebbian rule: Δw = η * x_i * y_j
				correlation := inputs[j] * outputs[i]
				adjustment := learningRate * correlation
				
				weights[i][j] += adjustment
				
				// Apply biological weight limits
				weights[i][j] = math.Max(BiologicalConstants.MinSynapticWeight, 
					math.Min(BiologicalConstants.MaxSynapticWeight, weights[i][j]))
			}
		}
	}
}

// CalculateCompressionEffectiveness measures how effective the neural-enhanced compression was
func CalculateCompressionEffectiveness(originalSize, compressedSize int, parameters map[string]float64) float64 {
	if originalSize <= 0 {
		return 0.0
	}
	
	// Basic compression ratio
	compressionRatio := 1.0 - (float64(compressedSize) / float64(originalSize))
	
	// Factor in parameter optimization (parameters closer to 1.0 are better)
	parameterOptimization := 0.0
	count := 0
	for _, param := range parameters {
		parameterOptimization += 1.0 - math.Abs(param-1.0)
		count++
	}
	if count > 0 {
		parameterOptimization /= float64(count)
	}
	
	// Combine compression ratio with parameter optimization
	effectiveness := (compressionRatio * 0.7) + (parameterOptimization * 0.3)
	
	return math.Max(0.0, math.Min(1.0, effectiveness))
}

// NormalizeContext7D normalizes 7D context values for neural processing
func NormalizeContext7D(context map[string]interface{}) map[string]float64 {
	normalized := make(map[string]float64)
	
	dimensions := []string{"who", "what", "when", "where", "why", "how", "extent"}
	
	for _, dim := range dimensions {
		if value, exists := context[dim]; exists {
			switch v := value.(type) {
			case string:
				normalized[dim] = StringToFloat(v)
			case int64:
				normalized[dim] = math.Mod(float64(v), 1000) / 1000.0
			case int:
				normalized[dim] = math.Mod(float64(v), 1000) / 1000.0  
			case float64:
				normalized[dim] = math.Mod(v, 1.0)
			default:
				normalized[dim] = 0.5 // Default middle value
			}
		} else {
			normalized[dim] = 0.0
		}
	}
	
	return normalized
}

// StringToFloat converts a string to a normalized float64 value using biological-inspired hashing
func StringToFloat(s string) float64 {
	if s == "" {
		return 0.0
	}
	
	// Biological-inspired string hashing using golden ratio
	hash := 0.0
	for i, char := range s {
		charVal := float64(char)
		hash += charVal * math.Pow(0.618033988749, float64(i)) // Golden ratio
	}
	
	// Normalize to 0-1 range using modulo and sigmoid
	return Activate(math.Mod(hash, 100)/100.0, ActivationSigmoid)
}

// CalculateNeuralComplexity estimates the computational complexity of neural processing
func CalculateNeuralComplexity(layerSizes []int) float64 {
	if len(layerSizes) < 2 {
		return 0.0
	}
	
	complexity := 0.0
	for i := 1; i < len(layerSizes); i++ {
		// Add complexity for connections between layers
		connections := layerSizes[i-1] * layerSizes[i]
		complexity += float64(connections)
	}
	
	// Normalize by total possible connections in largest layer
	maxLayer := 0
	for _, size := range layerSizes {
		if size > maxLayer {
			maxLayer = size
		}
	}
	
	if maxLayer > 0 {
		complexity = complexity / float64(maxLayer*maxLayer)
	}
	
	return complexity
}
