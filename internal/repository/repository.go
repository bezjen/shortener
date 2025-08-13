package repository

import "errors"

var (
	ErrNotFound = errors.New("record not found")
	ErrConflict = errors.New("record already exists")
)

type Repository interface {
	Save(id string, url string) error
	GetByShortURL(id string) (string, error)
	Ping() error
	Close() error
}
