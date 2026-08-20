package monitor

import (
	"context"
	"log/slog"
	"os"
)

type requestInfoKey struct{}

type RequestInfo struct {
	RequestID  string
	ActorID    string
	ActorRole  string
	ResourceID string
	Error      string
}

func Init() {
	handler := slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: slog.LevelInfo,

			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					a.Key = "timestamp"
				}

				return a
			},
		},
	)

	slog.SetDefault(slog.New(handler))
}

// getter
func InfoFromContext(ctx context.Context) *RequestInfo {
	info, _ := ctx.Value(requestInfoKey{}).(*RequestInfo)
	return info
}

// setter
func SetActor(ctx context.Context, actorID string, actorRole string) {
	info := InfoFromContext(ctx)
	if info == nil {
		return
	}

	info.ActorID = actorID
	info.ActorRole = actorRole
}

func SetResourceID(ctx context.Context, resourceID string) {
	info := InfoFromContext(ctx)
	if info == nil {
		return
	}

	info.ResourceID = resourceID
}

func SetError(ctx context.Context, err string) {
	info := InfoFromContext(ctx)
	if info == nil {
		return
	}

	info.Error = err
}

// logger
func Logger(ctx context.Context) *slog.Logger {
	logger := slog.Default()

	info := InfoFromContext(ctx)
	if info == nil {
		return logger
	}

	attributes := []any{
		"request_id", info.RequestID,
	}

	if info.ActorID != "" {
		attributes = append(
			attributes,
			"actor_id", info.ActorID,
			"actor_role", info.ActorRole,
		)
	}

	if info.ResourceID != "" {
		attributes = append(
			attributes,
			"resource_id", info.ResourceID,
		)
	}

	return logger.With(attributes...)
}
