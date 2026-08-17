package instance

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"cloud3-api/internal/httpresponse"
)

type Handler struct {
	store InstanceStore
}

func NewHandler(storage InstanceStore) *Handler {
	return &Handler{
		store: storage,
	}
}

// GET/instances
// return a list of all database instances
func (h *Handler) ListInstances(w http.ResponseWriter, r *http.Request) {
	principal, ok := requirePrincipal(w, r)
	if !ok {
		return
	}

	instanceList, err := h.store.List(r.Context())
	if err != nil {
		log.Println("failed to list instances:", err)
		httpresponse.PrintError(
			w,
			http.StatusInternalServerError,
			"failed to list instances",
		)
		return
	}

	// Admin can see every managed instance.
	if principal.Role == "admin" {
		httpresponse.WriteJSON(w, http.StatusOK, instanceList)
		return
	}

	if principal.Role != "user" {
		httpresponse.PrintError(w, http.StatusForbidden, "forbidden")
		return
	}

	// check normal user id matched with instance ownerID to list 
	ownedInstances := make([]DBInstance, 0)
	for _, instance := range instanceList {
		if instance.OwnerID == principal.UserID {
			ownedInstances = append(
				ownedInstances,
				instance,
			)
		}
	}

	httpresponse.WriteJSON(
		w,
		http.StatusOK,
		ownedInstances,
	)
}

// POST/instances
func (h *Handler) CreateInstance(w http.ResponseWriter, r *http.Request) {
	// check json file
	if r.Header.Get("Content-Type") != "application/json" {
		httpresponse.PrintError(w, http.StatusUnsupportedMediaType, "Content-Type not JSON")
		return
	}

	//read json amd convert into go request struct
	var request CreateInstanceRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println("failed to decode request body:", err)
		httpresponse.PrintError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	//validate Name and Instance
	if request.Name == "" {
		httpresponse.PrintError(w, http.StatusBadRequest, "missing instance name")
		return
	}

	if request.Instances < 1 {
		httpresponse.PrintError(w, http.StatusBadRequest, "instance number must be positive")
		return
	}

	principal, ok := requirePrincipal(w, r)
	if !ok {
		return
	}

	instance, err := h.store.Create(r.Context(), request, principal.UserID)
	if err != nil {
		log.Println("failed to create instance:", err)
		httpresponse.PrintError(w, http.StatusInternalServerError, "failed to create instance")
		return
	}

	httpresponse.WriteJSON(w, http.StatusCreated, instance)

}

// Get /instances/{id}
func (h *Handler) GetInstance(w http.ResponseWriter, r *http.Request, id string) {
	instance, ok := h.getAuthorizedInstance(w, r, id)
	if !ok {
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, instance)
}

// DELETE /instances/{id}
func (h *Handler) DeleteInstance(w http.ResponseWriter, r *http.Request, id string) {
	_, ok := h.getAuthorizedInstance(w, r, id)
	if !ok {
		return
	}

	err := h.store.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			httpresponse.PrintError(
				w,
				http.StatusNotFound,
				"instance not found",
			)
			return
		}

		log.Println("failed to delete instance:", err)

		httpresponse.PrintError(
			w,
			http.StatusInternalServerError,
			"failed to delete instance",
		)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// PUT /instances/{id}
func (h *Handler) PatchInstance(w http.ResponseWriter, r *http.Request, id string) {
	_, ok := h.getAuthorizedInstance(w, r, id)
	if !ok {
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		httpresponse.PrintError(w, http.StatusUnsupportedMediaType, "Content-Type not JSON")
		return
	}

	var request PatchInstanceRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println("failed to decode update request:", err)
		httpresponse.PrintError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if request.Name == nil &&
		request.Instances == nil &&
		request.Storage == nil &&
		request.CPU == nil {
		httpresponse.PrintError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	instance, err := h.store.Patch(r.Context(), id, request)
	if err != nil {
		switch {
		case errors.Is(err, ErrInstanceNotFound):
			httpresponse.PrintError(w, http.StatusNotFound, "instance not found")
		case errors.Is(err, ErrInvalidInstance):
			httpresponse.PrintError(w, http.StatusBadRequest, err.Error())
		default:
			log.Println("failed to patch instance:", err)
			httpresponse.PrintError(
				w,
				http.StatusInternalServerError,
				"failed to patch instance",
			)
		}
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, instance)

}

// GET /instacnes/{id}/connection
func (h *Handler) GetInstanceConnection(w http.ResponseWriter, r *http.Request, id string) {
	_, ok := h.getAuthorizedInstance(w, r, id)
	if !ok {
		return
	}

	connectionStore, ok := h.store.(ConnectionStore)
	if !ok {
		httpresponse.PrintError(w, http.StatusNotImplemented, "connection endpoint is not supported")
		return
	}

	connection, err := connectionStore.GetConnection(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			httpresponse.PrintError(w, http.StatusNotFound, "instance not found")
			return
		}

		if errors.Is(err, ErrConnectionNotReady) {
			httpresponse.PrintError(w, http.StatusConflict, "connection information is not ready")
			return
		}

		log.Println("failed to get connection information:", err)
		httpresponse.PrintError(w, http.StatusInternalServerError, "failed to get connection information")
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, connection)
}


func requirePrincipal(
	w http.ResponseWriter,
	r *http.Request,
) (Principal, bool) {
	principal, ok := PrincipalFromContext(r.Context())
	if !ok || strings.TrimSpace(principal.UserID) == "" {
		httpresponse.PrintError(
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
			httpresponse.PrintError(
				w,
				http.StatusNotFound,
				"instance not found",
			)
			return DBInstance{}, false
		}

		log.Println("failed to get instance:", err)
		httpresponse.PrintError(
			w,
			http.StatusInternalServerError,
			"failed to get instance",
		)
		return DBInstance{}, false
	}

	if !canAccessInstance(principal, instance) {
		// Do not reveal that another user's instance exists.
		httpresponse.PrintError(
			w,
			http.StatusNotFound,
			"instance not found",
		)
		return DBInstance{}, false
	}

	return instance, true
}
