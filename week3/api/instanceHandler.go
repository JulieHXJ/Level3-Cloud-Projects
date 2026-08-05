package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
)

// GET/instances
// return a list of all database instances
func (h *Handler) getInstances(w http.ResponseWriter, r *http.Request) {
	// get list from memoryStorage
	instanceList, err := h.store.List(r.Context())
	if err != nil {
		log.Println("failed to list instances:", err)
		printError(w, http.StatusInternalServerError, "failed to list instances")
		return
	}

	writeJSON(w, http.StatusOK, instanceList)
}

// POST/instances
func (h *Handler) createInstances(w http.ResponseWriter, r *http.Request) {
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

	instance, err := h.store.Create(r.Context(), request)
	if err != nil {
		log.Println("failed to create instance:", err)
		printError(w, http.StatusInternalServerError, "failed to create instance")
		return
	}

	writeJSON(w, http.StatusCreated, instance)

}

// Get /instances/{id}
func (h *Handler) getInstanceByID(w http.ResponseWriter, r *http.Request, id string) {
	instance, err := h.store.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			printError(w, http.StatusNotFound, "instance not fount")
			return
		}

		log.Println("failed to get instance:", err)
		printError(w, http.StatusInternalServerError, "failed to get instance")
		return
	}

	writeJSON(w, http.StatusOK, instance)
}

// DELETE /instances/{id}
func (h *Handler) deleteInstance(w http.ResponseWriter, r *http.Request, id string) {
	err := h.store.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			printError(w, http.StatusNotFound, "instance not fount")
			return
		}
		log.Println("failed to delete instance:", err)
		printError(w, http.StatusInternalServerError, "failed to delete instance")
		return

	}

	w.WriteHeader(http.StatusNoContent)
}

// PUT /instances/{id}
func (h *Handler) updateInstance(w http.ResponseWriter, r *http.Request, id string) {
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
		printError(w, http.StatusBadRequest, "instance number must be positive")
		return
	}

	instance, err := h.store.Update(r.Context(), id, request)
	if err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			printError(w, http.StatusNotFound, "instance not fount")
			return
		}
		log.Println("failed to update instance:", err)
		printError(w, http.StatusInternalServerError, "failed to update instance")
		return
	}

	writeJSON(w, http.StatusOK, instance)

}


// GET /instacnes/{id}/connection
func (h *Handler) getConnection(w http.ResponseWriter, r *http.Request, id string) {
	connectionStore, ok := h.store.(ConnectionStore)
	if !ok {
		printError(w, http.StatusNotImplemented, "connection endpoint is not supported")
		return
	}

	connection, err := connectionStore.GetConnection(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			printError(w, http.StatusNotFound, "instance not found")
			return
		}

		if errors.Is(err, ErrConnectionNotReady) {
			printError(w, http.StatusConflict, "connection information is not ready")
			return
		}

		log.Println("failed to get connection information:", err)
		printError(w, http.StatusInternalServerError, "failed to get connection information")
		return
	}

	writeJSON(w, http.StatusOK, connection)
}

// GET & POST wrapper
func (h *Handler) instancesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getInstances(w, r)
	case http.MethodPost:
		h.createInstances(w, r)
	default:
		printError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) instancesIDHandler(w http.ResponseWriter, r *http.Request) {
	//get id (and connction) from url
	path := strings.TrimPrefix(r.URL.Path, "/instances/")
	path = strings.Trim(path, "/")
	if path == "" {
		printError(w, http.StatusBadRequest, "missing instance id")
		return
	}

	arr := strings.Split(path, "/")
	id := arr[0]
	if id == "" {
		printError(w, http.StatusBadRequest, "missing instance id")
		return
	}

	// GET /instances/{id}/connection
	if len(arr) == 2 && arr[1] == "connection" {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			printError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.getConnection(w, r, id)
		return
	}

	//router
	switch r.Method {
	case http.MethodGet:
		h.getInstanceByID(w, r, id)

	case http.MethodPut:
		h.updateInstance(w, r, id)

	case http.MethodDelete:
		h.deleteInstance(w, r, id)

	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		printError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
