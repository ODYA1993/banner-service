package user

import (
	"context"
)

type Storage interface {
	Create(ctx context.Context, user *User) error
	FindAll(ctx context.Context) ([]User, error)
	FindOne(ctx context.Context, id string) (User, error)
	FindOneByEmail(ctx context.Context, email string) (*User, error)
	Delete(ctx context.Context, id string) (int64, error)
}
