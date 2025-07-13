package repository

import "errors"

var ErrNotFound = errors.New("record not found")

type Repository interface {
	Save(id string, url string)
	GetByShortURL(id string) (string, error)
}
