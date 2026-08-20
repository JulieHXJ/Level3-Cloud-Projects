package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud3-api/internal/httpresponse"
	"cloud3-api/internal/instance"
	"cloud3-api/internal/monitor"
	"cloud3-api/internal/user"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	jwtIssuer  = "Cloud3-api"
	TokenTTL   = time.Hour
	bcryptCost = 12
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type claimsContextKeyType struct{}

var claimsContextKey claimsContextKeyType

// raised by server, stores info about token identity
type Claims struct {
	Role     string `json:"role"`
	ClinicID string `json:"clinic_id,omitempty"`
	jwt.RegisteredClaims
}

// link the authentication data to user group
type AuthService struct {
	secret        []byte
	users         user.UserStore
	dummyPassHash []byte //preserve time for non-existing user
}

// get jwt secrect and create authService
// currently read from env
func NewAuthService(userStore user.UserStore) (*AuthService, error) {
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

	return &AuthService{
		secret:        []byte(secret),
		users:         userStore,
		dummyPassHash: dummyPasswordHash,
	}, nil
}

func (a *AuthService) Authenticate(ctx context.Context, username string, password string) (user.User, error) {
	username = strings.TrimSpace(username)
	dbUser, err := a.users.GetByUsername(ctx, username)

	userExists := true
	passwordHash := a.dummyPassHash

	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			userExists = false
		} else {
			return user.User{}, fmt.Errorf(
				"query user: %w",
				err,
			)
		}
	} else {
		passwordHash = []byte(dbUser.PasswordHash)
	}

	// compare a bcrypt hash, even when the user does not exist.
	passwordMatches := bcrypt.CompareHashAndPassword(
		passwordHash,
		[]byte(password),
	) == nil

	// same response when the username or password is incorrect.
	if !userExists || !passwordMatches {
		return user.User{}, ErrInvalidCredentials
	}

	return *dbUser, nil
}

func (a *AuthService) IssueToken(dbUser user.User) (string, error) {
	now := time.Now().UTC()

	claims := Claims{
		Role: string(dbUser.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtIssuer,
			Subject:   dbUser.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(TokenTTL)),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	signedToken, err := token.SignedString(a.secret)
	if err != nil {
		return "", fmt.Errorf(
			"sign JWT: %w",
			err,
		)
	}

	return signedToken, nil
}

// jwtMiddleware authenticates protected requests.
func (a *AuthService) JwtMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Browser CORS preflight do not require a JWT.
		// if r.Method == http.MethodOptions {
		// 	next.ServeHTTP(w, r)
		// 	return
		// }

		// Public endpoints.
		if r.Method == http.MethodGet &&
			r.URL.Path == "/api/v1/health" {
			next.ServeHTTP(w, r)
			return
		}

		if r.Method == http.MethodPost &&
			(r.URL.Path == "/api/v1/auth/login" ||
				r.URL.Path == "/api/v1/auth/register") {
			next.ServeHTTP(w, r)
			return
		}

		//  check
		tokenString, ok := readBearerToken(
			r.Header.Get("Authorization"),
		)
		if !ok {
			monitor.SetError(r.Context(), "missing_bearer_token")
			monitor.Logger(r.Context()).Warn(
				"authentication_failed",
				"reason", "missing_bearer_token",
			)
			unauthorized(w)
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(
			tokenString,
			claims,
			func(token *jwt.Token) (any, error) {
				if token.Method.Alg() !=
					jwt.SigningMethodHS256.Alg() {
					return nil, fmt.Errorf(
						"unexpected signing method: %s",
						token.Method.Alg(),
					)
				}

				return a.secret, nil
			},
			jwt.WithValidMethods([]string{
				jwt.SigningMethodHS256.Alg(),
			}),
			jwt.WithIssuer(jwtIssuer),
			jwt.WithExpirationRequired(),
			jwt.WithIssuedAt(),
		)

		if err != nil || !token.Valid {
			monitor.SetError(r.Context(), "invalid_token")
			monitor.Logger(r.Context()).Warn(
				"authentication_failed",
				"reason", "invalid_token",
			)
			unauthorized(w)
			return
		}

		// These fields are required by my application,
		// Only supported platform roles are accepted.
		if claims.Subject == "" {
			monitor.SetError(
				r.Context(),
				"invalid_token_subject",
			)

			monitor.Logger(r.Context()).Warn(
				"authentication_failed",
				"reason", "invalid_token_subject",
			)
			unauthorized(w)
			return
		}

		if claims.Role != "admin" &&
			claims.Role != "user" {
			monitor.SetError(
				r.Context(),
				"invalid_token_role",
			)

			monitor.Logger(r.Context()).Warn(
				"authentication_failed",
				"reason", "invalid_token_role",
			)

			unauthorized(w)
			return
		}

		monitor.SetActor(r.Context(), claims.Subject, claims.Role)

		// pass claim to later handlers.
		requestContext := context.WithValue(
			r.Context(),
			claimsContextKey,
			claims,
		)

		// convert to instance principal for instace handler
		requestContext = instance.WithPrincipal(
			requestContext,
			instance.Principal{
				UserID: claims.Subject,
				Role:   claims.Role,
			},
		)

		next.ServeHTTP(
			w,
			r.WithContext(requestContext),
		)
	})
}

// ----------helper-----------------
func readBearerToken(header string) (string, bool) {
	parts := strings.Fields(header)

	if len(parts) != 2 {
		return "", false
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	if parts[1] == "" {
		return "", false
	}

	return parts[1], true
}

func claimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	if !ok || claims == nil {
		return nil, false
	}

	return claims, true
}

// read middleware claim and check role is allowed
func requireRoles(w http.ResponseWriter, r *http.Request, allowedRoles ...string) bool {
	claims, ok := claimsFromContext(r.Context())
	if !ok {
		unauthorized(w)
		return false
	}

	for _, allowedRole := range allowedRoles {
		if claims.Role == allowedRole {
			return true
		}
	}

	httpresponse.PrintError(
		w,
		http.StatusForbidden,
		"forbidden",
	)

	return false
}

func unauthorized(w http.ResponseWriter) {
	httpresponse.PrintError(
		w,
		http.StatusUnauthorized,
		"invalid or missing bearer token",
	)
}
