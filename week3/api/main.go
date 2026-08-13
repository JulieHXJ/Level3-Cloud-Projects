package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	api "cloud3-api/internal/api"
	"cloud3-api/internal/instance"
	"cloud3-api/internal/user"

	"github.com/labstack/echo/v4"
)

const (
	apiV1Prefix   = "/api/v1"
	instancesPath = apiV1Prefix + "/instances"
	usersPath     = apiV1Prefix + "/users"
	petsPath      = apiV1Prefix + "/pets"
)

func main() {

	// create user db
	ctx := context.Background()

	platformDatabaseURL := os.Getenv("PLATFORM_DATABASE_URL")
	if platformDatabaseURL == "" {
		log.Fatal("PLATFORM_DATABASE_URL is not set")
	}

	userStore, err := user.NewPostgresStorage(ctx, platformDatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to platform database: ", err)
	}
	defer userStore.Close()

	log.Println("connected to platform database")

	// CloudNativePG Cluster CR namespace。
	namespace := os.Getenv("DB_NAMESPACE")
	if namespace == "" {
		namespace = "postgres-demo"
	}

	// create Kubernetes client and save into InstanceStorage
	store, err := instance.NewKubeStorage(namespace)
	if err != nil {
		log.Fatal("failed to create Kubernetes storage: ", err)
	}

	auth, err := newAuthService(userStore)
	if err != nil {
		log.Fatal("failed to configure authentication: ", err)
	}

	handler := instance.NewHandler(store)

	// use servermux, register router
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", healthHandler)
	handler.RegisterRoutes(mux)

	// mux.HandleFunc(usersPath, handler.userHandler)
	// mux.HandleFunc(usersPath+"/", handler.userByIDHandler)

	// mux.HandleFunc(petsPath, handler.petHandler)
	// mux.HandleFunc(petsPath+"/", handler.petByIDHandler)

	// Echo is now the externally exposed router.
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Authenticate every request through middleware
	e.Use(auth.jwtMiddleware)
	generatedServer := &APIServer{
		legacy:    mux,
		auth:      auth,
		userStore: userStore,
	}

	// Register every route generated from openapi.yaml.
	api.RegisterHandlers(e, generatedServer)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           corsMiddleware(e),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Server is listening to http://localhost:8080")
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// ---------------- Handlers---------------
// r = Request
// w = Response
func healthHandler(w http.ResponseWriter, r *http.Request) {
	// check request name
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	}); err != nil {
		log.Printf("failed to encode health response: %v", err)
	}
}
