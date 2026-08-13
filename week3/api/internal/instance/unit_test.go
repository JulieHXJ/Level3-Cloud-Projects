package instance

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testCreatedAt = "2026-08-05T11:36:16Z"
const testOwnerID = "11111111-1111-1111-1111-111111111111"

type testStore struct {
	*MemoryStorage
	connections map[string]ConnectionInfo
}

func newTestStore() *testStore {
	return &testStore{
		MemoryStorage: NewMemoryStorage(),
		connections:   make(map[string]ConnectionInfo),
	}
}

// get Service & Secret
func (s *testStore) GetConnection(ctx context.Context, id string) (ConnectionInfo, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return ConnectionInfo{}, err
	}

	connection, ok := s.connections[id]
	if !ok {
		return ConnectionInfo{}, ErrConnectionNotReady
	}

	return connection, nil
}

// router
func newTestMux(store InstanceStore) *http.ServeMux {
	handler := NewHandler(store)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	return mux
}

// simulate client request and get response
// t - text object, for error reporting
// method - GET、POST、PATCH、DELETE
// path - /health、/instances、/instances/1
// body - POST  JSON
// Content-Type: application/json
func performRequest(t *testing.T, handler http.Handler, method string, path string, body []byte, contentType string,
) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(
		method,
		path,
		bytes.NewReader(body),
	)

	// All instance API tests simulate an authenticated normal user.
	request = withTestPrincipal(request)

	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	return recorder
}

// convert JSON from response to GO struct
func decodeResponse[T any](t *testing.T, recorder *httptest.ResponseRecorder) T {
	t.Helper()

	var response T
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf(
			"failed to decode response: %v; body=%q",
			err,
			recorder.Body.String(),
		)
	}
	return response
}

func createTestInstance(t *testing.T, store InstanceStore, name string, instances int) DBInstance {
	t.Helper()

	instance, err := store.Create(
		context.Background(),
		CreateInstanceRequest{
			Name:      name,
			Instances: instances,
		},
		testOwnerID,
	)

	if err != nil {
		t.Fatalf("failed to create test instance: %v", err)
	}

	if instance.OwnerID != testOwnerID {
		t.Fatalf(
			"expected owner ID %q, got %q",
			testOwnerID,
			instance.OwnerID,
		)
	}

	return instance
}

// GET /instances
func TestListInstancesSuccess(t *testing.T) {
	store := newTestStore()
	createTestInstance(t, store, "first-database", 1)
	createTestInstance(t, store, "second-database", 2)

	recorder := performRequest(
		t,
		newTestMux(store),
		http.MethodGet,
		instancesPath,
		nil,
		"",
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	instances := decodeResponse[[]DBInstance](t, recorder)
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}
}

// POST /instance
func TestCreateInstanceSuccess(t *testing.T) {
	recorder := performRequest(
		t,
		newTestMux(newTestStore()),
		http.MethodPost,
		instancesPath,
		[]byte(`{"name":"pet-health-db","instances":1}`),
		"application/json",
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	instance := decodeResponse[DBInstance](t, recorder)
	if instance.ID == "" {
		t.Fatal("expected a generated instance ID")
	}
	if instance.Name != "pet-health-db" {
		t.Fatalf("expected name pet-health-db, got %q", instance.Name)
	}
	if instance.Instances != 1 {
		t.Fatalf("expected instances=1, got %d", instance.Instances)
	}

	if instance.Storage != defaultStorageSize {
		t.Fatalf(
			"expected storage=%q, got %q",
			defaultStorageSize,
			instance.Storage,
		)
	}

	if instance.CPU != defaultCPURequest {
		t.Fatalf(
			"expected cpu=%q, got %q",
			defaultCPURequest,
			instance.CPU,
		)
	}
}

func TestCreateInstanceWithResourcesSuccess(t *testing.T) {
	recorder := performRequest(
		t,
		newTestMux(newTestStore()),
		http.MethodPost,
		instancesPath,
		[]byte(`{
			"name": "pet-health-db",
			"instances": 2,
			"storage": "5Gi",
			"cpu": "500m"
		}`),
		"application/json",
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	instance := decodeResponse[DBInstance](t, recorder)

	if instance.Storage != "5Gi" {
		t.Fatalf("expected storage=5Gi, got %q", instance.Storage)
	}

	if instance.CPU != "500m" {
		t.Fatalf("expected cpu=500m, got %q", instance.CPU)
	}
}

// POST /instance missing body
func TestCreateInstanceInvalidJSON(t *testing.T) {
	recorder := performRequest(
		t,
		newTestMux(newTestStore()),
		http.MethodPost,
		instancesPath,
		[]byte(`{"name":`),
		"application/json",
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

// POST /instance not json
func TestCreateInstanceUnsupportedMediaType(t *testing.T) {
	recorder := performRequest(
		t,
		newTestMux(newTestStore()),
		http.MethodPost,
		instancesPath,
		[]byte(`{"name":"pet-health-db","instances":1}`),
		"text/plain",
	)

	if recorder.Code != http.StatusUnsupportedMediaType {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusUnsupportedMediaType,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

// GET /instance/{id}
func TestGetInstanceSuccess(t *testing.T) {
	store := newTestStore()

	// first create instance
	created := createTestInstance(t, store, "pet-health-db", 1)

	// then get info
	recorder := performRequest(
		t,
		newTestMux(store),
		http.MethodGet,
		instancesPath+"/"+created.ID,
		nil,
		"",
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	instance := decodeResponse[DBInstance](t, recorder)

	//check
	if instance.ID != created.ID {
		t.Fatalf("expected id %q, got %q", created.ID, instance.ID)
	}
}

// GET /instance/{id} with wrong id
func TestGetInstanceNotFound(t *testing.T) {
	recorder := performRequest(
		t,
		newTestMux(newTestStore()),
		http.MethodGet,
		instancesPath+"/missing",
		nil,
		"",
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestPatchInstancePartialUpdateSuccess(t *testing.T) {
	store := newTestStore()
	created := createTestInstance(
		t,
		store,
		"pet-health-db",
		1,
	)

	recorder := performRequest(
		t,
		newTestMux(store),
		http.MethodPatch,
		instancesPath+"/"+created.ID,
		[]byte(`{"storage":"5Gi"}`),
		"application/json",
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	instance := decodeResponse[DBInstance](t, recorder)

	if instance.Storage != "5Gi" {
		t.Fatalf("expected storage=5Gi, got %q", instance.Storage)
	}

	// 没有出现在 PATCH body 中的字段必须保持不变。
	if instance.Name != created.Name {
		t.Fatalf(
			"expected name to remain %q, got %q",
			created.Name,
			instance.Name,
		)
	}

	if instance.Instances != created.Instances {
		t.Fatalf(
			"expected instances to remain %d, got %d",
			created.Instances,
			instance.Instances,
		)
	}

	if instance.CPU != created.CPU {
		t.Fatalf(
			"expected cpu to remain %q, got %q",
			created.CPU,
			instance.CPU,
		)
	}
}

func TestPatchInstanceRejectsStorageDecrease(t *testing.T) {
	store := newTestStore()
	created := createTestInstance(
		t,
		store,
		"pet-health-db",
		1,
	)

	firstPatch := performRequest(
		t,
		newTestMux(store),
		http.MethodPatch,
		instancesPath+"/"+created.ID,
		[]byte(`{"storage":"10Gi"}`),
		"application/json",
	)

	if firstPatch.Code != http.StatusOK {
		t.Fatalf(
			"failed to prepare instance: status=%d body=%s",
			firstPatch.Code,
			firstPatch.Body.String(),
		)
	}

	recorder := performRequest(
		t,
		newTestMux(store),
		http.MethodPatch,
		instancesPath+"/"+created.ID,
		[]byte(`{"storage":"5Gi"}`),
		"application/json",
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

// DELETE /instance/{id}
func TestDeleteInstanceSuccess(t *testing.T) {
	store := newTestStore()
	created := createTestInstance(t, store, "pet-health-db", 1)

	recorder := performRequest(
		t,
		newTestMux(store),
		http.MethodDelete,
		instancesPath+"/"+created.ID,
		nil,
		"",
	)

	//check status code = 204
	if recorder.Code != http.StatusAccepted {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusAccepted,
			recorder.Code,
			recorder.Body.String(),
		)
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("expected an empty body, got %q", recorder.Body.String())
	}

	_, err := store.Get(context.Background(), created.ID)
	if err != ErrInstanceNotFound {
		t.Fatalf("expected instance to be deleted, got error %v", err)
	}
}

func TestDeleteInstanceNotFound(t *testing.T) {
	recorder := performRequest(
		t,
		newTestMux(newTestStore()),
		http.MethodDelete,
		instancesPath+"/missing",
		nil,
		"",
	)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusNotFound,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

// GET /instance/{id}/connection
func TestGetConnectionSuccess(t *testing.T) {
	store := newTestStore()
	created := createTestInstance(t, store, "pet-health-db", 1)

	store.connections[created.ID] = ConnectionInfo{
		Host:     created.ID + "-rw",
		Port:     "5432",
		Database: "app",
		Username: "app",
		Password: "test-password",
		URI: "postgresql://app:test-password@" +
			created.ID +
			"-rw.postgres-demo:5432/app",
	}

	recorder := performRequest(
		t,
		newTestMux(store),
		http.MethodGet,
		instancesPath+"/"+created.ID+"/connection",
		nil,
		"",
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	connection := decodeResponse[ConnectionInfo](t, recorder)
	if connection.Host != created.ID+"-rw" {
		t.Fatalf(
			"expected host %q, got %q",
			created.ID+"-rw",
			connection.Host,
		)
	}
	if connection.Port != "5432" {
		t.Fatalf("expected port 5432, got %q", connection.Port)
	}
}

func TestGetConnectionNotReady(t *testing.T) {
	store := newTestStore()
	created := createTestInstance(t, store, "pet-health-db", 1)

	recorder := performRequest(
		t,
		newTestMux(store),
		http.MethodGet,
		instancesPath+"/"+created.ID+"/connection",
		nil,
		"",
	)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusConflict,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func withTestPrincipal(r *http.Request) *http.Request {
	ctx := WithPrincipal(
		r.Context(),
		Principal{
			UserID: testOwnerID,
			Role:   "user",
		},
	)

	return r.WithContext(ctx)
}
