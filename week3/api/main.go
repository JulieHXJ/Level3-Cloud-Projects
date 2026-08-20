package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	api "cloud3-api/internal/api"
	"cloud3-api/internal/auth"
	"cloud3-api/internal/instance"
	"cloud3-api/internal/user"

	cloudmonitor "cloud3-api/internal/monitor"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	apiV1Prefix   = "/api/v1"
	instancesPath = apiV1Prefix + "/instances"
	usersPath     = apiV1Prefix + "/users"
	petsPath      = apiV1Prefix + "/pets"
)

func main() {
	cloudmonitor.Init()

	// ctx := context.Background()
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// user db related
	platformDatabaseURL := os.Getenv("PLATFORM_DATABASE_URL")
	if platformDatabaseURL == "" {
		slog.Error(
			"platform_database_configuration_failed",
			"error", "platform_database_url_missing",
		)
		return
	}

	userStore, err := user.NewPostgresStorage(ctx, platformDatabaseURL)
	if err != nil {
		slog.Error(
			"platform_database_connection_failed",
			"error", "database_connection_failed",
		)
		return
	}
	defer userStore.Close()
	slog.Info("platform_database_connected")

	// create Kubernetes storage
	namespace := os.Getenv("DB_NAMESPACE")
	if namespace == "" {
		namespace = "postgres-demo"
	}

	store, err := instance.NewKubeStorage(namespace)
	if err != nil {
		slog.Error(
			"kubernetes_storage_initialization_failed",
			"error", "kubernetes_storage_initialization_failed",
		)
		return
	}

	slog.Info(
		"kubernetes_storage_initialized",
		"namespace", namespace,
	)

	// authService
	authService, err := auth.NewAuthService(userStore)
	if err != nil {
		slog.Error(
			"authentication_configuration_failed",
			"error", "authentication_configuration_failed",
		)
		return
	}

	//wrapper for Instance count
	instanceHandler := instance.NewHandler(store)

	if err := instanceHandler.SyncInstanceCount(ctx); err != nil {
		slog.Warn("instance_count_metric_sync_failed")
	}

	// ths real business starts!
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
	limiter := cloudmonitor.NewRateLimiter(60, time.Minute)

	apiHandler := cloudmonitor.Middleware(
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

	slog.Info("server_started", "address", server.Addr)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		slog.Error("server_failed", "error", "http_server_failed")
	}
}
