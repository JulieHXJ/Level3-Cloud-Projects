package user

import (
	"context"
	"errors"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrUsernameExists = errors.New("username already exists")
)



type UserStore interface {
    Create(ctx context.Context, user User) error
    GetByUsername(ctx context.Context, username string) (*User, error)
    GetByID(ctx context.Context, id string) (*User, error)
}