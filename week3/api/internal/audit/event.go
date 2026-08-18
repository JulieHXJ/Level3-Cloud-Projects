package audit

import "time"

type Event struct {
	Timestamp     time.Time
	ActorID       string
	ActorRole     string
	Action        string
	ResourceType string
	ResourceID   string
	Result        string
	RequestID    string
	FailureReason string
}