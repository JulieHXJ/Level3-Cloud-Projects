package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud3-api/internal/user"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	jwtIssuer  = "Cloud3-api"
	tokenTTL   = time.Hour
	bcryptCost = 12
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"tokenType"`
	ExpiresIn int    `json:"expiresIn"`
	Role      string `json:"role"`
}

// raised by server, stores info about token identity
type Claims struct {
	Role     string `json:"role"`
	ClinicID string `json:"clinic_id,omitempty"`
	jwt.RegisteredClaims
}

// link the authentication data to user group
type authService struct {
	secret        []byte
	users         user.UserStore
	dummyPassHash []byte //preserve time for non-existing user
}

// get jwt secrect and create authService
// currently read from env
func newAuthService(userStore user.UserStore) (*authService, error) {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return nil, fmt.Errorf(
			"JWT_SECRET must contain at least 32 characters",
		)
	}

	dummyPasswordHash, err := bcrypt.GenerateFromPassword(
		[]byte("this-is-not-a-real-user-password"),
		bcryptCost,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create dummy password hash: %w",
			err,
		)
	}

	return &authService{
		secret:        []byte(secret),
		users:         userStore,
		dummyPassHash: dummyPasswordHash,
	}, nil
}

// func loginHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusNotImplemented)

// 	if err := json.NewEncoder(w).Encode(map[string]string{
// 		"error": "login not implemented yet",
// 	}); err != nil {
// 		log.Printf("failed to encode login response: %v", err)
// 	}
// }

func (a *authService) login(ctx echo.Context) error {
	var request loginRequest

	decoder := json.NewDecoder(ctx.Request().Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return ctx.JSON(
			http.StatusBadRequest,
			map[string]string{
				"error": "invalid request body",
			},
		)
	}

	request.Username = strings.TrimSpace(request.Username)

	dbUser, err := a.users.GetByUsername(
		ctx.Request().Context(),
		request.Username,
	)

	userExists := true
	passwordHash := a.dummyPassHash

	// keep comparing password, to provent attacker guessing by charactor(r.pass == user.pass)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			userExists = false
		} else {
			return ctx.JSON(
				http.StatusInternalServerError,
				map[string]string{
					"error": "failed to query user",
				},
			)
		}
	} else {
		passwordHash = []byte(dbUser.PasswordHash)
	}

	passwordMatches := bcrypt.CompareHashAndPassword(
		passwordHash,
		[]byte(request.Password),
	) == nil

	// user not exist / wrong password (don't return seperately! risky!)
	if !userExists || !passwordMatches {
		return ctx.JSON(
			http.StatusUnauthorized,
			map[string]string{
				"error": "invalid username or password",
			},
		)
	}

	now := time.Now().UTC()

	claims := Claims{
		Role: string(dbUser.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtIssuer,
			Subject:   dbUser.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
	}

	// builde Token
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	signedToken, err := token.SignedString(a.secret)
	if err != nil {
		return ctx.JSON(
			http.StatusInternalServerError,
			map[string]string{
				"error": "failed to issue token",
			},
		)
	}

	return ctx.JSON(
		http.StatusOK,
		loginResponse{
			Token:     signedToken,
			TokenType: "Bearer",
			ExpiresIn: int(tokenTTL.Seconds()),
			Role:      string(dbUser.Role),
		},
	)
}
