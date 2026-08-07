package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// // Local memory storage for testing.
// // Key: instance ID
// // Value: database instance
// var instances = make(map[string]DBInstance)

// // generate IDs
// var nextID int = 1

type Handler struct {
	store InstanceStore
}

const (
	apiV1Prefix   = "/api/v1"
	instancesPath = apiV1Prefix + "/instances"
	usersPath     = apiV1Prefix + "/users"
	petsPath      = apiV1Prefix + "/pets"
)

func NewHandler(storage InstanceStore) *Handler {
	return &Handler{
		store: storage,
	}
}

func main() {
	// CloudNativePG Cluster CR namespace。
	namespace := os.Getenv("DB_NAMESPACE")
	if namespace == "" {
		namespace = "postgres-demo"
	}

	// create Kubernetes client save into InstanceStorage
	store, err := NewKubeStorage(namespace)
	if err != nil {
		log.Fatal("failed to create Kubernetes storage: ", err)
	}

	// mstorage := NewMemoryStorage()
	handler := NewHandler(store)

	// use servermux, register router
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc(instancesPath, handler.instancesHandler)       //GET POST /instance
	mux.HandleFunc(instancesPath+"/", handler.instancesIDHandler) //Get DELETE PUT /instances/{id}

	// mux.HandleFunc(usersPath, handler.userHandler)
	// mux.HandleFunc(usersPath+"/", handler.userByIDHandler)

	// mux.HandleFunc(petsPath, handler.petHandler)
	// mux.HandleFunc(petsPath+"/", handler.petByIDHandler)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Server is listening to http://localhost:8080")
	err = server.ListenAndServe()
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

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	},
	)
}

// ---------------------Helper func ---------------
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
