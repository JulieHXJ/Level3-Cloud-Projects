package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// Local memory storage for testing.
// Key: instance ID
// Value: database instance
var instances = make(map[string]DBInstance)

// generate IDs
var nextID int = 1

func main() {

	// register router
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/instances", instancesHandler)    //GET POST /instance
	http.HandleFunc("/instances/", instancesIDHandler) //Get DELETE PUT /instances/{id}

	log.Println("Server is listening to http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// ---------------- Handlers---------------
// r = Request
// w = Response
func healthHandler(w http.ResponseWriter, r *http.Request) {
	// check request name
	if r.Method != http.MethodGet {
		printError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) //status code

	_, err := w.Write([]byte(`{"status":"ok"}`)) // write into response body, returns written bytes and error
	if err != nil {
		log.Println("failed to write health response:", err)
	}

}

// GET & POST wrapper
func instancesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getInstances(w, r)
	case http.MethodPost:
		createInstances(w, r)
	default:
		printError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func instancesIDHandler(w http.ResponseWriter, r *http.Request) {
	//get id from url
	id := strings.TrimPrefix(r.URL.Path, "/instances/")
	if id == "" {
		printError(w, http.StatusBadRequest, "missing instance id")
		return
	}

	//router
	switch r.Method {
	case http.MethodGet:
		getInstanceByID(w, id)

	case http.MethodPut:
		updateInstance(w, r, id)

	case http.MethodDelete:
		deleteInstance(w, id)

	default:
		printError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

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
