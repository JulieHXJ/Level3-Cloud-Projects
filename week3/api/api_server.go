package main

import (
	"errors"
	"net/http"
	"time"

	api "cloud3-api/internal/api"
	"cloud3-api/internal/user"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// APIServer implements api.ServerInterface.
//
// Echo and the generated OpenAPI routes handle routing.
// The existing net/http mux continues handling business logic.
type APIServer struct {
	legacy    http.Handler
	auth      *authService
	userStore user.UserStore
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

func (s *APIServer) RegisterUser(ctx echo.Context) error {
	var req api.RegisterRequest

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "invalid request body",
		})
	}

	if req.Username == "" || req.Password == "" {
		return ctx.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "username and password are required",
		})
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Error: "failed to hash password",
		})
	}

	id := uuid.New()
	now := time.Now().UTC()

	newUser := user.User{
		ID:           id.String(),
		Username:     req.Username,
		PasswordHash: string(passwordHash),
		Role:         user.RoleUser,
		CreatedAt:    now,
	}

	if err := s.userStore.Create(ctx.Request().Context(), newUser); err != nil {
		if errors.Is(err, user.ErrUsernameExists) {
			return ctx.NoContent(http.StatusConflict)
		}

		return ctx.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Error: "failed to create user",
		})
	}

	return ctx.JSON(http.StatusCreated, api.UserResponse{
		Id:        id,
		Username:  newUser.Username,
		Role:      api.UserRole(newUser.Role),
		CreatedAt: now,
	})
}

func (s *APIServer) Login(ctx echo.Context) error {
	return s.auth.login(ctx)
}

func (s *APIServer) GetHealth(ctx echo.Context) error {
	return s.forward(ctx)
}

func (s *APIServer) ListInstances(ctx echo.Context) error {
	return s.forwardAs(ctx, "admin")
}

func (s *APIServer) CreateInstance(ctx echo.Context) error {
	return s.forwardAs(ctx, "admin")
}

func (s *APIServer) GetInstance(ctx echo.Context, _ api.InstanceID) error {
	return s.forwardAs(ctx, "admin")
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
