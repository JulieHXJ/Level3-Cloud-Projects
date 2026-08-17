package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	api "cloud3-api/internal/api"
	"cloud3-api/internal/httpresponse"
	"cloud3-api/internal/instance"
	"cloud3-api/internal/user"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// APIServer implements api.ServerInterface.
// APIServer = instanceHandler + Health + Login + Register
type APIServer struct {
	auth      *authService
	instances *instance.Handler
	userStore user.UserStore
}

// interface check
var _ api.ServerInterface = (*APIServer)(nil)

func (s *APIServer) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req api.LoginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		httpresponse.PrintError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	if strings.TrimSpace(req.Username) == "" ||
		req.Password == "" {
		httpresponse.PrintError(
			w,
			http.StatusBadRequest,
			"username and password are required",
		)
		return
	}

	dbUser, err := s.auth.authenticate(
		r.Context(),
		req.Username,
		req.Password,
	)
	if err != nil {
		if errors.Is(err, errInvalidCredentials) {
			httpresponse.PrintError(
				w,
				http.StatusUnauthorized,
				"invalid username or password",
			)
			return
		}

		log.Printf("failed to authenticate user: %v", err)

		httpresponse.PrintError(
			w,
			http.StatusInternalServerError,
			"failed to authenticate user",
		)
		return
	}

	signedToken, err := s.auth.issueToken(dbUser)
	if err != nil {
		log.Printf("failed to issue token: %v", err)

		httpresponse.PrintError(
			w,
			http.StatusInternalServerError,
			"failed to issue token",
		)
		return
	}

	httpresponse.WriteJSON(
		w,
		http.StatusOK,
		api.LoginResponse{
			AccessToken: signedToken,
			TokenType:   "Bearer",
			ExpiresIn:   int(tokenTTL.Seconds()),
			Role:        api.UserRole(dbUser.Role),
		},
	)

}

func (s *APIServer) RegisterUser(w http.ResponseWriter, r *http.Request) {
	// Prevent an excessively large registration request body.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req api.RegisterRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		httpresponse.PrintError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	// if empoty
	if req.Username == "" || req.Password == "" {
		httpresponse.PrintError(
			w,
			http.StatusBadRequest,
			"username and password are required",
		)
		return
	}

	// abay the password policy
	if err := user.ValidatePassword(req.Password); err != nil {
		httpresponse.PrintError(
			w,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		log.Printf("failed to hash password: %v", err)
		httpresponse.PrintError(
			w,
			http.StatusInternalServerError,
			"failed to hash password",
		)
		return
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

	if err := s.userStore.Create(r.Context(), newUser); err != nil {
		if errors.Is(err, user.ErrUsernameExists) {
			httpresponse.PrintError(
				w,
				http.StatusConflict,
				"username already exists",
			)
			return
		}
		log.Printf("failed to create user: %v", err)

		httpresponse.WriteJSON(
			w,
			http.StatusInternalServerError,
			"failed to create user",
		)
		return
	}

	httpresponse.WriteJSON(w, http.StatusCreated, api.UserResponse{
		Id:        id,
		Username:  newUser.Username,
		Role:      api.UserRole(newUser.Role),
		CreatedAt: now,
	})
}

func (s *APIServer) GetHealth(w http.ResponseWriter, r *http.Request) {
	httpresponse.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})

}

func (s *APIServer) ListInstances(w http.ResponseWriter, r *http.Request) {
	s.instances.ListInstances(w, r)
}

func (s *APIServer) CreateInstance(w http.ResponseWriter, r *http.Request) {
	s.instances.CreateInstance(w, r)
}

func (s *APIServer) GetInstance(w http.ResponseWriter, r *http.Request, id api.InstanceID) {
	s.instances.GetInstance(w, r, id)
}

func (s *APIServer) PatchInstance(w http.ResponseWriter, r *http.Request, id api.InstanceID) {
	s.instances.PatchInstance(w, r, id)
}

func (s *APIServer) DeleteInstance(w http.ResponseWriter, r *http.Request, id api.InstanceID) {
	s.instances.DeleteInstance(w, r, id)
}

func (s *APIServer) GetInstanceConnection(w http.ResponseWriter, r *http.Request, id api.InstanceID) {
	s.instances.GetInstanceConnection(w, r, id)
}
