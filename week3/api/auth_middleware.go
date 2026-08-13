package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const claimsContextKey = "jwtClaims"

// jwtMiddleware authenticates protected requests.
func (a *authService) jwtMiddleware(next echo.HandlerFunc) echo.HandlerFunc {

	return func(ctx echo.Context) error {
		request := ctx.Request()

		// Browser CORS preflight do not require a JWT.
		if request.Method == http.MethodOptions {
			return next(ctx)
		}

		// Public endpoints health.
		if request.Method == http.MethodGet &&
			request.URL.Path == "/api/v1/health" {
			return next(ctx)
		}

		if request.Method == http.MethodPost &&
			(request.URL.Path == "/api/v1/auth/login" ||
				request.URL.Path == "/api/v1/auth/register") {
			return next(ctx)
		}

		tokenString, ok := readBearerToken(
			request.Header.Get("Authorization"),
		)
		if !ok {
			return unauthorized(ctx)
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
			return unauthorized(ctx)
		}

		// These fields are required by my application,
		// Only supported platform roles are accepted.
		if claims.Subject == "" {
			return unauthorized(ctx)
		}

		if claims.Role != "admin" &&
			claims.Role != "user" {
			return unauthorized(ctx)
		}

		// pass claim to later handlers.
		ctx.Set(claimsContextKey, claims)

		return next(ctx)
	}
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

// read middleware claim and check role is allowed
func requireRoles(ctx echo.Context, allowedRoles ...string) error {
	claims, ok := ctx.Get(claimsContextKey).(*Claims)
	if !ok || claims == nil {
		return unauthorized(ctx)
	}

	for _, allowedRole := range allowedRoles {
		if claims.Role == allowedRole {
			return nil
		}
	}

	return ctx.JSON(
		http.StatusForbidden,
		map[string]string{
			"error": "forbidden",
		},
	)
}

func unauthorized(ctx echo.Context) error {
	return ctx.JSON(
		http.StatusUnauthorized,
		map[string]string{
			"error": "invalid or missing bearer token",
		},
	)
}
