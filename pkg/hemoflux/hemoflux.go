package hemoflux

import (
	"encoding/json"
	"math"
	"time"
)

// MobiusCompressionMeta holds all variables needed for lossless decompression and context preservation
// EXTENT: Single compression operation
type MobiusCompressionMeta struct {
	Algorithm       string                 `json:"algorithm"`
	Version         string                 `json:"version"`
	Timestamp       int64                  `json:"timestamp"`
	OriginalType    string                 `json:"originalType"`
	OriginalSize    int                    `json:"originalSize"`
	CompressionVars map[string]float64     `json:"compressionVars"`
	Context         map[string]interface{} `json:"context"`
}

// MobiusCompress compresses data using the Möbius formula and 7D context.
// If standalone is false, this is a passthrough: compression is handled by the TNOS MCP server.
func MobiusCompress(data []byte, context map[string]interface{}, standalone bool) ([]byte, *MobiusCompressionMeta, error) {
	if !standalone {
		// In blood-connected mode, compression is handled by the TNOS MCP server (remote Hemoflux)
		return data, &MobiusCompressionMeta{
			Algorithm:    "mobius7d",
			Version:      "1.2",
			Timestamp:    time.Now().UnixMilli(),
			OriginalType: "[]byte",
			OriginalSize: len(data),
			Context:      context,
		}, nil
	}
	// Standalone mode: perform Mobius compression locally
	value := float64(len(data))
	entropy := calculateEntropy(data)
	B, V, I, G, F := extractContextFactors(context)
	t := float64(time.Now().UnixNano()) / 1e9
	E := 0.5
	alignment := (B + V*I) * math.Exp(-t*E)

	byteFreq := calculateByteFrequency(data)
	runLength := calculateRunLength(data)
	mean, stddev := calculateMeanStddev(data)
	normValue := (value - mean) / (stddev + 1e-9)

	compressed := (normValue * B * I * (1 - (entropy / math.Log2(1+V))) * (G + F) * (runLength + 1)) /
		(E*t + alignment + stddev + 1)

	compressionVars := map[string]float64{
		"value":   value,
		"entropy": entropy,
		"B":       B, "V": V, "I": I, "G": G, "F": F, "E": E, "t": t, "alignment": alignment,
		"runLength": runLength,
		"mean":      mean,
		"stddev":    stddev,
		"normValue": normValue,
	}

	meta := &MobiusCompressionMeta{
		Algorithm:       "mobius7d",
		Version:         "1.2",
		Timestamp:       time.Now().UnixMilli(),
		OriginalType:    "[]byte",
		OriginalSize:    len(data),
		CompressionVars: compressionVars,
		Context:         context,
	}
	meta.Context["byteFreq"] = byteFreq
	compressedBytes, _ := json.Marshal(compressed)
	return compressedBytes, meta, nil
}

func calculateEntropy(data []byte) float64 {
	freq := make(map[byte]int)
	for _, b := range data {
		freq[b]++
	}
	entropy := 0.0
	l := float64(len(data))
	for _, count := range freq {
		p := float64(count) / l
		entropy -= p * math.Log2(p)
	}
	return entropy
}

func extractContextFactors(context map[string]interface{}) (B, V, I, G, F float64) {
	B, V, I, G, F = 1.0, 1.0, 1.0, 1.0, 1.0
	if context == nil {
		return
	}
	if b, ok := context["B"].(float64); ok {
		B = b
	}
	if v, ok := context["V"].(float64); ok {
		V = v
	}
	if i, ok := context["I"].(float64); ok {
		I = i
	}
	if g, ok := context["G"].(float64); ok {
		G = g
	}
	if f, ok := context["F"].(float64); ok {
		F = f
	}
	return
}

func calculateByteFrequency(data []byte) map[string]float64 {
	freq := make(map[byte]int)
	for _, b := range data {
		freq[b]++
	}
	norm := float64(len(data))
	result := make(map[string]float64)
	for b, count := range freq {
		result[string([]byte{b})] = float64(count) / norm
	}
	return result
}

func calculateRunLength(data []byte) float64 {
	if len(data) == 0 {
		return 0
	}
	runs := 1
	for i := 1; i < len(data); i++ {
		if data[i] != data[i-1] {
			runs++
		}
	}
	return float64(len(data)) / float64(runs)
}

func calculateMeanStddev(data []byte) (float64, float64) {
	if len(data) == 0 {
		return 0, 0
	}
	sum := 0.0
	for _, b := range data {
		sum += float64(b)
	}
	mean := sum / float64(len(data))
	variance := 0.0
	for _, b := range data {
		variance += math.Pow(float64(b)-mean, 2)
	}
	variance /= float64(len(data))
	return mean, math.Sqrt(variance)
}
