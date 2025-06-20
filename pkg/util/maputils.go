/*
 * WHO: SafeValueExtractor
 * WHAT: Safe value extraction utilities for map[string]interface{}
 * WHEN: During data extraction operations
 * WHERE: System Layer 6 (Integration)
 * WHY: To prevent type assertion panics and provide safe defaults
 * HOW: Using type assertion with default fallback patterns
 * EXTENT: All map value extraction operations
 */

package util

// GetStringValue safely extracts string values from a map with default fallback
// WHO: SafeValueExtractor
// WHAT: Extract string value with type safety
// WHEN: During map value extraction
// WHERE: System Layer 6 (Integration)
// WHY: To prevent type assertion panics
// HOW: Using type assertion with default fallback
// EXTENT: All map string value extraction
func GetStringValue(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

// GetFloatValue safely extracts float64 values from a map with default fallback
func GetFloatValue(m map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := m[key]; ok {
		if floatVal, ok := val.(float64); ok {
			return floatVal
		}
		// Try to convert from int
		if intVal, ok := val.(int); ok {
			return float64(intVal)
		}
	}
	return defaultValue
}

// GetIntValue safely extracts int values from a map with default fallback  
func GetIntValue(m map[string]interface{}, key string, defaultValue int) int {
	if val, ok := m[key]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
		// Try to convert from float64
		if floatVal, ok := val.(float64); ok {
			return int(floatVal)
		}
	}
	return defaultValue
}

// GetBoolValue safely extracts bool values from a map with default fallback
func GetBoolValue(m map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := m[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}
