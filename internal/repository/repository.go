//go:generate mockery --name=Repository --output=../mocks --outpkg=mocks --case=underscore
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
	Save(ctx context.Context, userID string, url model.URL) error
	SaveBatch(ctx context.Context, userID string, urls []model.URL) error
	DeleteBatch(ctx context.Context, userID string, shortURLs []string) error
	GetByShortURL(ctx context.Context, id string) (string, error)
	GetByUserID(ctx context.Context, userID string) ([]model.URL, error)
	Ping(ctx context.Context) error
	Close() error
}
