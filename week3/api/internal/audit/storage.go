package audit

import "context"

type EventStore interface {
	Record(ctx context.Context, event Event) error
}