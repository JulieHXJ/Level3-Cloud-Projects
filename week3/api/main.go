package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	api "cloud3-api/internal/api"
	"cloud3-api/internal/auth"
	"cloud3-api/internal/instance"
	"cloud3-api/internal/user"

	cloudmetrics "cloud3-api/internal/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	apiV1Prefix   = "/api/v1"
	instancesPath = apiV1Prefix + "/instances"
	usersPath     = apiV1Prefix + "/users"
	petsPath      = apiV1Prefix + "/pets"
)

func main() {

	ctx := context.Background()

	// user db related
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

	// create Kubernetes storage
	namespace := os.Getenv("DB_NAMESPACE")
	if namespace == "" {
		namespace = "postgres-demo"
	}

	store, err := instance.NewKubeStorage(namespace)
	if err != nil {
		log.Fatal("failed to create Kubernetes storage: ", err)
	}

	// authService
	authService, err := auth.NewAuthService(userStore)
	if err != nil {
		log.Fatal("failed to configure authentication: ", err)
	}


	//wrapper for Instance count
	instanceHandler := instance.NewHandler(store)

	if err := instanceHandler.SyncInstanceCount(ctx); err != nil {
		log.Printf(
			"failed to initialize instance count metric: %v",
			err,
		)
	}

	apiServer := &APIServer{
		auth:      authService,
		userStore: userStore,
		instances: instanceHandler,
	}

	// endpoint handler
	router := api.Handler(apiServer)

	// CORS & JWT authentication
	// corsHandler := corsMiddleware(auth.jwtMiddleware(routerHandler))


	//rate limiter
	limiter := cloudmetrics.NewRateLimiter(60, time.Minute)


	apiHandler := cloudmetrics.Middleware(
		corsMiddleware(
			limiter.Middleware(
				authService.JwtMiddleware(router),
			),
		),
	)

	rootMux := http.NewServeMux()
	rootMux.Handle("/metrics", promhttp.Handler()) // for GET /metrics
	rootMux.Handle("/", apiHandler)                // ui & api

	server := &http.Server{
		Addr:              ":8080",
		Handler:           rootMux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Server is listening to http://localhost:8080")
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
