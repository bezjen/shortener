package repository

import (
	"context"
	"errors"
)

var (
	ErrNotFound = errors.New("record not found")
	ErrConflict = errors.New("record already exists")
)

type Repository interface {
	Save(ctx context.Context, id string, url string) error
	GetByShortURL(ctx context.Context, id string) (string, error)
	Ping(ctx context.Context) error
	Close() error
}
