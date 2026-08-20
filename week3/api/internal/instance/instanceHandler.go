package instance

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"cloud3-api/internal/httpresponse"
	"cloud3-api/internal/monitor"
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
		// log.Println("failed to list instances:", err)
		monitor.SetError(r.Context(), "instance_list_failed")
		monitor.Logger(r.Context()).Error(
			"instance_operation_failed",
			"operation", "list",
			"cause", err,
		)
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
		// log.Println("failed to decode request body:", err)
		monitor.SetError(
			r.Context(),
			"invalid_request_body",
		)

		monitor.Logger(r.Context()).Warn(
			"request_validation_failed",
			"reason", "invalid_request_body",
		)
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
		monitor.SetError(
			r.Context(),
			"instance_create_failed",
		)

		monitor.Logger(r.Context()).Error(
			"instance_operation_failed",
			"operation", "create",
			"cause", err,
		)

		// count failure in operations cout
		monitor.InstanceOperationsTotal.WithLabelValues("create", "failure").Inc()

		httpresponse.PrintError(w, http.StatusInternalServerError, "failed to create instance")
		return
	}

	// count success into operations total
	monitor.InstanceOperationsTotal.WithLabelValues("create", "success").Inc()
	monitor.SetResourceID(r.Context(), instance.ID)

	// instacne total +1
	if err := h.SyncInstanceCount(r.Context()); err != nil {
		monitor.Logger(r.Context()).Warn("instance_count_metric_sync_failed")
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
			monitor.SetError(r.Context(), "instance_not_found")
			httpresponse.PrintError(
				w,
				http.StatusNotFound,
				"instance not found",
			)
			return
		}

		monitor.SetError(r.Context(), "instance_deleted_failed")
		monitor.Logger(r.Context()).Error(
			"instance_operation_failed",
			"operation", "delete",
			"cause", err,
		)
		monitor.InstanceOperationsTotal.WithLabelValues("delete", "failure").Inc()

		httpresponse.PrintError(
			w,
			http.StatusInternalServerError,
			"failed to delete instance",
		)
		return
	}

	monitor.InstanceOperationsTotal.WithLabelValues("delete", "success").Inc()

	// instance total -1
	if err := h.SyncInstanceCount(r.Context()); err != nil {
		monitor.Logger(r.Context()).Warn("instance_count_metric_sync_failed")
	}

	w.WriteHeader(http.StatusAccepted)
}

// PATCH /instances/{id}
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
		monitor.SetError(r.Context(), "invalid_request_body")

		monitor.Logger(r.Context()).Warn(
			"request_validation_failed",
			"reason", "invalid_request_body",
		)
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
			monitor.SetError(r.Context(), "instance_not_found")
			httpresponse.PrintError(w, http.StatusNotFound, "instance not found")
		case errors.Is(err, ErrInvalidInstance):
			monitor.SetError(r.Context(), "invalid_instance")
			httpresponse.PrintError(w, http.StatusBadRequest, err.Error())
		default:
			monitor.SetError(r.Context(), "instance_patch_failed")
			monitor.Logger(r.Context()).Error(
				"instance_operation_failed",
				"operation", "patch",
				"cause", err,
			)
			monitor.InstanceOperationsTotal.WithLabelValues("update", "failure").Inc()
			httpresponse.PrintError(
				w,
				http.StatusInternalServerError,
				"failed to patch instance",
			)
		}
		return
	}

	monitor.InstanceOperationsTotal.WithLabelValues("update", "success").Inc()
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
		monitor.SetError(r.Context(), "connection_not_supported")
		httpresponse.PrintError(w, http.StatusNotImplemented, "connection endpoint is not supported")
		return
	}

	connection, err := connectionStore.GetConnection(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			monitor.SetError(r.Context(), "instance_not_found")
			httpresponse.PrintError(w, http.StatusNotFound, "instance not found")
			return
		}

		if errors.Is(err, ErrConnectionNotReady) {
			monitor.SetError(r.Context(), "connection_not_ready")
			httpresponse.PrintError(w, http.StatusConflict, "connection information is not ready")
			return
		}

		monitor.SetError(r.Context(), "connection_lookup_failed")
		monitor.Logger(r.Context()).Error(
			"instance_connection_lookup_failed",
			"cause", err,
		)
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
		monitor.SetError(r.Context(), "missing_principal")

		monitor.Logger(r.Context()).Error("request_principal_missing")

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

func (h *Handler) getAuthorizedInstance(w http.ResponseWriter, r *http.Request, id string) (DBInstance, bool) {
	monitor.SetResourceID(r.Context(), id)

	principal, ok := requirePrincipal(w, r)
	if !ok {
		monitor.SetError(r.Context(), "missing_principal")
		return DBInstance{}, false
	}

	instance, err := h.store.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			monitor.SetError(r.Context(), "instance_not_found")
			httpresponse.PrintError(
				w,
				http.StatusNotFound,
				"instance not found",
			)
			return DBInstance{}, false
		}

		monitor.SetError(r.Context(), "instance_get_failed")
		monitor.Logger(r.Context()).Error(
			"instance_operation_failed",
			"operation", "get",
			"cause", err,
		)
		// log.Println("failed to get instance:", err)
		httpresponse.PrintError(
			w,
			http.StatusInternalServerError,
			"failed to get instance",
		)
		return DBInstance{}, false
	}

	if !canAccessInstance(principal, instance) {
		monitor.SetError(
			r.Context(),
			"instance_access_denied",
		)

		monitor.Logger(r.Context()).Warn(
			"authorization_failed",
			"reason", "instance_access_denied",
		)

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

// helper for InstacneCurrent metrics
func (h *Handler) SyncInstanceCount(ctx context.Context) error {
	instanceList, err := h.store.List(ctx)
	if err != nil {
		return err
	}

	monitor.InstancesCurrent.Set(float64(len(instanceList)))
	return nil
}
