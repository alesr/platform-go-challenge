package httputil

import (
	"encoding/json"
	"log"
	"net/http"
)

// Response represents a standard API response with generic data type.
type Response[T any] struct {
	Status  string `json:"status"`
	Data    T      `json:"data,omitzero"` // Go 1.24 FTW!
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ResponseWriter wraps http.ResponseWriter to capture status code.
type ResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w, http.StatusOK} // default
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RespondWithJSON writes JSON response with automatic success wrapper
func RespondWithJSON[T any](w http.ResponseWriter, code int, data T) {
	response := Response[T]{
		Status: "success",
		Data:   data,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		log.Printf("Could not marshal JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(jsonBytes)
}

// RespondWithError writes JSON error response
func RespondWithError[T any](w http.ResponseWriter, code int, message string) {
	response := Response[T]{
		Status:  "error",
		Message: message,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		log.Printf("Could not marshal JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(jsonBytes)
}
