package main

import (
	"context"
	"errors"
)	

var (
	ErrInstanceNotFound  = errors.New("instance not found")
	ErrInvalidInstance   = errors.New("invalid instance")
	ErrConnectionNotReady = errors.New("connection information not ready")
)

const (
	defaultStorageSize = "1Gi"
	defaultCPURequest  = "250m"
)

// interface defination
type InstanceStore interface {
	List(ctx context.Context) ([]DBInstance, error) //func, param, return type succeed and fail
	Get(ctx context.Context, id string) (DBInstance, error)
	Create(ctx context.Context, r CreateInstanceRequest) (DBInstance, error)
	Patch(ctx context.Context, id string, r PatchInstanceRequest) (DBInstance, error)
	Delete(ctx context.Context, id string) error
}


type ConnectionStore interface {
	GetConnection (ctx context.Context, id string) (ConnectionInfo, error)
}