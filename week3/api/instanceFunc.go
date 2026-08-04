package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

// return a list of all database instances for GET/instances
func getInstances(w http.ResponseWriter, r *http.Request) {
	// Convert map -> slice
	instanceList := make([]DBInstance, 0, len(instances))

	// ignore id
	for _, instance := range instances {
		instanceList = append(instanceList, instance)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// slice -> JSON array
	err := json.NewEncoder(w).Encode(instanceList)
	if err != nil {
		log.Println("failed to encode get instances response:", err)
	}
}

// POST/instances
func createInstances(w http.ResponseWriter, r *http.Request) {
	// check json file
	if r.Header.Get("Content-Type") != "application/json" {
		printError(w, http.StatusUnsupportedMediaType, "Content-Type not JSON")
		return
	}

	//read json amd convert into go request struct
	var request CreateInstanceRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println("failed to decode request body:", err)
		printError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	//validate Name and Instance
	if request.Name == "" {
		printError(w, http.StatusBadRequest, "missing instance name")
		return
	}

	if request.Instances < 1 {
		printError(w, http.StatusBadRequest, "instance number must be positive")
		return
	}

	// prepare id and create instance
	id := strconv.Itoa(nextID)
	nextID++
	instance := DBInstance{
		ID:        id,
		Name:      request.Name,
		Instances: request.Instances,
		Status:    "pending",
		CreatedAt: time.Now().Format("2006-01-02T15:04:05"),
	}

	//save to map and return status
	instances[id] = instance
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(instance)
	if err != nil {
		log.Println("failed to encode create instance response:", err)
	}
}


// Get /instances/{id}
func getInstanceByID(w http.ResponseWriter, id string) {
	instance, exist := instances[id]
	if !exist {
		printError(w, http.StatusNotFound, "instance not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(instance)
	if err != nil {
		log.Println("fialed to encode instance response:", err)
	}

}

// DELETE /instances/{id}
func deleteInstance(w http.ResponseWriter, id string) {
	_, exist := instances[id]
	if !exist {
		printError(w, http.StatusNotFound, "instance not found")
		return
	}

	delete(instances, id) // go function for delet map element

	w.WriteHeader(http.StatusNoContent)
}

// PUT /instances/{id}
func updateInstance(w http.ResponseWriter, r *http.Request, id string) {
	instance, exists := instances[id]
	if !exists {
		printError(w, http.StatusNotFound, "instance not found")
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		printError(w, http.StatusUnsupportedMediaType, "Content-Type not JSON")
		return
	}

	var request UpdateInstanceRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println("failed to decode update request:", err)
		printError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if request.Name == "" {
		printError(w, http.StatusBadRequest, "missing instance name")
		return
	}

	if request.Instances < 1 {
		printError(
			w,
			http.StatusBadRequest,
			"instance number must be positive",
		)
		return
	}

	//update
	instance.Name = request.Name
	instance.Instances = request.Instances

	instances[id] = instance

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(instance)
	if err != nil {
		log.Println("failed to encode updated instance:", err)
	}

}