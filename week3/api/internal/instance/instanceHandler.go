package instance

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
)

const instancesPath = "/api/v1/instances"

type Handler struct {
	store InstanceStore
}

func NewHandler(storage InstanceStore) *Handler {
	return &Handler{
		store: storage,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc(instancesPath, h.instancesHandler)
	mux.HandleFunc(instancesPath+"/", h.instancesIDHandler)
}

// GET/instances
// return a list of all database instances
func (h *Handler) getInstances(w http.ResponseWriter, r *http.Request) {
	principal, ok := requirePrincipal(w, r)
	if !ok {
		return
	}

	instanceList, err := h.store.List(r.Context())
	if err != nil {
		log.Println("failed to list instances:", err)
		printError(
			w,
			http.StatusInternalServerError,
			"failed to list instances",
		)
		return
	}

	// Admin can see every managed instance.
	if principal.Role == "admin" {
		writeJSON(w, http.StatusOK, instanceList)
		return
	}

	if principal.Role != "user" {
		printError(w, http.StatusForbidden, "forbidden")
		return
	}

	// Normal users only see their own instances.
	ownedInstances := make([]DBInstance, 0)

	for _, instance := range instanceList {
		if instance.OwnerID == principal.UserID {
			ownedInstances = append(
				ownedInstances,
				instance,
			)
		}
	}

	writeJSON(
		w,
		http.StatusOK,
		ownedInstances,
	)
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

	principal, ok := requirePrincipal(w, r)
	if !ok {
		return
	}

	instance, err := h.store.Create(r.Context(), request, principal.UserID)
	if err != nil {
		log.Println("failed to create instance:", err)
		printError(w, http.StatusInternalServerError, "failed to create instance")
		return
	}

	writeJSON(w, http.StatusCreated, instance)

}

// Get /instances/{id}
func (h *Handler) getInstanceByID(w http.ResponseWriter, r *http.Request, id string) {
	instance, ok := h.getAuthorizedInstance(w, r, id)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, instance)
}

// DELETE /instances/{id}
func (h *Handler) deleteInstance(w http.ResponseWriter, r *http.Request, id string) {
	_, ok := h.getAuthorizedInstance(w, r, id)
	if !ok {
		return
	}

	err := h.store.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			printError(
				w,
				http.StatusNotFound,
				"instance not found",
			)
			return
		}

		log.Println("failed to delete instance:", err)

		printError(
			w,
			http.StatusInternalServerError,
			"failed to delete instance",
		)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// PUT /instances/{id}
func (h *Handler) patchInstance(w http.ResponseWriter, r *http.Request, id string) {
	_, ok := h.getAuthorizedInstance(w, r, id)
	if !ok {
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		printError(w, http.StatusUnsupportedMediaType, "Content-Type not JSON")
		return
	}

	var request PatchInstanceRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println("failed to decode update request:", err)
		printError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if request.Name == nil &&
		request.Instances == nil &&
		request.Storage == nil &&
		request.CPU == nil {
		printError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	instance, err := h.store.Patch(r.Context(), id, request)
	if err != nil {
		switch {
		case errors.Is(err, ErrInstanceNotFound):
			printError(w, http.StatusNotFound, "instance not found")
		case errors.Is(err, ErrInvalidInstance):
			printError(w, http.StatusBadRequest, err.Error())
		default:
			log.Println("failed to patch instance:", err)
			printError(
				w,
				http.StatusInternalServerError,
				"failed to patch instance",
			)
		}
		return
	}

	writeJSON(w, http.StatusOK, instance)

}

// GET /instacnes/{id}/connection
func (h *Handler) getConnection(w http.ResponseWriter, r *http.Request, id string) {
	_, ok := h.getAuthorizedInstance(w, r, id)
	if !ok {
		return
	}

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
	path := strings.TrimPrefix(r.URL.Path, instancesPath+"/")
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
			w.Header().Set("Allow", http.MethodGet)
			printError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.getConnection(w, r, id)
		return
	}

	// 拒绝 /instances/{id}/unknown 等无效路径。
	if len(arr) != 1 {
		printError(
			w,
			http.StatusNotFound,
			"endpoint not found",
		)
		return
	}

	//router
	switch r.Method {
	case http.MethodGet:
		h.getInstanceByID(w, r, id)

	case http.MethodPatch:
		h.patchInstance(w, r, id)

	case http.MethodDelete:
		h.deleteInstance(w, r, id)

	default:
		w.Header().Set("Allow", "GET, PATCH, DELETE")
		printError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func requirePrincipal(
	w http.ResponseWriter,
	r *http.Request,
) (Principal, bool) {
	principal, ok := PrincipalFromContext(r.Context())
	if !ok || strings.TrimSpace(principal.UserID) == "" {
		printError(
			w,
			http.StatusUnauthorized,
			"authenticated user identity is missing",
		)
		return Principal{}, false
	}

	return principal, true
}

func canAccessInstance(
	principal Principal,
	instance DBInstance,
) bool {
	if principal.Role == "admin" {
		return true
	}

	return principal.Role == "user" &&
		instance.OwnerID != "" &&
		instance.OwnerID == principal.UserID
}

func (h *Handler) getAuthorizedInstance(
	w http.ResponseWriter,
	r *http.Request,
	id string,
) (DBInstance, bool) {
	principal, ok := requirePrincipal(w, r)
	if !ok {
		return DBInstance{}, false
	}

	instance, err := h.store.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			printError(
				w,
				http.StatusNotFound,
				"instance not found",
			)
			return DBInstance{}, false
		}

		log.Println("failed to get instance:", err)
		printError(
			w,
			http.StatusInternalServerError,
			"failed to get instance",
		)
		return DBInstance{}, false
	}

	if !canAccessInstance(principal, instance) {
		// Do not reveal that another user's instance exists.
		printError(
			w,
			http.StatusNotFound,
			"instance not found",
		)
		return DBInstance{}, false
	}

	return instance, true
}
