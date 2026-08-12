package main

import (
	"net/http"

	api "cloud3-api/internal/api"

	"github.com/labstack/echo/v4"
)

// APIServer implements api.ServerInterface.
//
// Echo and the generated OpenAPI routes handle routing.
// The existing net/http mux continues handling business logic.
type APIServer struct {
	legacy http.Handler
	auth   *authService
}

func (s *APIServer) forward(ctx echo.Context) error {
	s.legacy.ServeHTTP(ctx.Response(), ctx.Request())
	return nil
}

func (s *APIServer) forwardAs(ctx echo.Context, roles ...string) error {
	if err := requireRoles(ctx, roles...); err != nil {
		return err
	}

	return s.forward(ctx)
}

func (s *APIServer) Login(ctx echo.Context) error {
	return s.auth.login(ctx)
}

func (s *APIServer) GetHealth(ctx echo.Context) error {
	return s.forward(ctx)
}

// only admin and viewer can access.
func (s *APIServer) ListInstances(ctx echo.Context) error {
	return s.forwardAs(ctx, "admin", "viewer")
}

func (s *APIServer) CreateInstance(ctx echo.Context) error {
	return s.forwardAs(ctx, "admin")
}

func (s *APIServer) GetInstance(ctx echo.Context, _ api.InstanceID) error {
	return s.forwardAs(ctx, "admin", "viewer")
}

func (s *APIServer) PatchInstance(ctx echo.Context, _ api.InstanceID) error {
	return s.forwardAs(ctx, "admin")
}

func (s *APIServer) DeleteInstance(ctx echo.Context, _ api.InstanceID) error {
	return s.forwardAs(ctx, "admin")
}

func (s *APIServer) GetInstanceConnection(ctx echo.Context, _ api.InstanceID) error {
	return s.forwardAs(ctx, "admin")
}

// func (s *APIServer) RotateInstanceCredentials(
// 	ctx echo.Context,
// 	_ api.InstanceID,
// ) error {
// 	return s.forwardAs(ctx, "admin")
// }
