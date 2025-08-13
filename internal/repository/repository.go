package repository

import (
	"context"
	"errors"
	"github.com/bezjen/shortener/internal/model"
)

var (
	ErrNotFound         = errors.New("record not found")
	ErrShortURLConflict = errors.New("record with short url already exists")
)

type Repository interface {
	Save(ctx context.Context, url model.URL) error
	SaveBatch(ctx context.Context, urls []model.URL) error
	GetByShortURL(ctx context.Context, id string) (string, error)
	Ping(ctx context.Context) error
	Close() error
}
