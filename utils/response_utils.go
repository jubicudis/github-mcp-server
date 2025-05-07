/*
 * WHO: ResponseUtilities
 * WHAT: HTTP response handling utilities
 * WHEN: During API request processing
 * WHERE: System Layer 6 (Integration)
 * WHY: To standardize API responses
 * HOW: Using Go HTTP response functions
 * EXTENT: All HTTP responses
 */

package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

// ResponseEnvelope wraps all API responses
type ResponseEnvelope struct {
	// WHO: ResponseStructure
	// WHAT: Response container
	Success bool        `json:"success"`           // WHY: Operation result
	Data    interface{} `json:"data,omitempty"`    // WHAT: Response payload
	Error   string      `json:"error,omitempty"`   // WHY: Error message if failed
	Context interface{} `json:"context,omitempty"` // WHERE: 7D context information
}

// RespondWithError sends an error response with specified status code
func RespondWithError(w http.ResponseWriter, code int, message string) {
	// WHO: ErrorResponder
	// WHAT: Error response generation
	// WHEN: During error handling
	// WHERE: HTTP response processing
	// WHY: To communicate error state
	// HOW: Using standardized format
	// EXTENT: Single error response

	RespondWithJSON(w, code, ResponseEnvelope{
		Success: false,
		Error:   message,
	})
}

// RespondWithJSON sends a JSON response with specified status code
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	// WHO: JSONResponder
	// WHAT: JSON response generation
	// WHEN: During response processing
	// WHERE: HTTP request completion
	// WHY: To return structured data
	// HOW: Using JSON serialization
	// EXTENT: Single JSON response

	// Set content type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	// For nil payload, just return empty JSON object
	if payload == nil {
		payload = struct{}{}
	}

	// Wrap non-envelope responses
	if _, isEnvelope := payload.(ResponseEnvelope); !isEnvelope {
		payload = ResponseEnvelope{
			Success: code >= 200 && code < 300,
			Data:    payload,
		}
	}

	// Serialize to JSON
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success":false,"error":"Failed to generate response"}`))
		return
	}

	w.Write(response)
}
