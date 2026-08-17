package instance

import "context"

type Principal struct {
	UserID string
	Role   string
}

type principalContextKey struct{}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(
		ctx,
		principalContextKey{},
		principal,
	)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(
		principalContextKey{},
	).(Principal)

	return principal, ok
}

// take jwt claims to principal and then to handler
