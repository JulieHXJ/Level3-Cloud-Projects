package httpresponse

import (
	"encoding/json"
	"log"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

// print JSON error with status code
func PrintError(w http.ResponseWriter, status int, m string) {
	WriteJSON(w, status, errorResponse{
		Error: m,
	})
}

// any -> interface
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Println("failed to encode instance response:", err)
	}
}
