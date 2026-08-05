package main

import (
	"context"
	"errors"
)	

var ErrConnectionNotReady = errors.New(("connection information not ready"))

// interface defination
type InstanceStore interface {
	List(ctx context.Context) ([]DBInstance, error) //func, param, return type succeed and fail
	Get(ctx context.Context, id string) (DBInstance, error)
	Create(ctx context.Context, r CreateInstanceRequest) (DBInstance, error)
	Update(ctx context.Context, id string, r UpdateInstanceRequest) (DBInstance, error)
	Delete(ctx context.Context, id string) error
}


type ConnectionStore interface {
	GetConnection (ctx context.Context, id string) (ConnectionInfo, error)
}