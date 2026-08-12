package instance

import (
	"encoding/json"
	"log"
	"net/http"
)

// print JSON error with status code
func printError(w http.ResponseWriter, status int, m string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := map[string]string{
		"error": m,
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println("failed to encode error response:", err)
	}
}

// any -> interface
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Println("failed to encode instance response:", err)
	}
}
