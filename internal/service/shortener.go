package service

import (
	"crypto/rand"
	"errors"
	"github.com/bezjen/shortener/internal/repository"
	"math/big"
)

const (
	shortURLLength   = 8
	charset          = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	maxAttemptsCount = 10
)

var ErrGenerate = errors.New("failed to generate short url")

type Shortener interface {
	GenerateShortURLPart(url string) (string, error)
	GetURLByShortURLPart(shortURLPart string) (string, error)
	PingRepository() error
}

type URLShortener struct {
	storage repository.Repository
}

func NewURLShortener(storage repository.Repository) *URLShortener {
	return &URLShortener{
		storage: storage,
	}
}

func (u *URLShortener) GenerateShortURLPart(url string) (string, error) {
	for i := 0; i < maxAttemptsCount; i++ {
		shortURL, err := generateRandomString(shortURLLength)
		if err != nil {
			return "", err
		}
		err = u.storage.Save(shortURL, url)
		if err != nil {
			if errors.Is(err, repository.ErrConflict) {
				continue
			}
			return "", err
		}
		return shortURL, nil
	}
	return "", ErrGenerate
}

func (u *URLShortener) GetURLByShortURLPart(shortURLPart string) (string, error) {
	resultURL, err := u.storage.GetByShortURL(shortURLPart)
	if err != nil {
		return "", err
	}
	return resultURL, nil
}

func (u *URLShortener) PingRepository() error {
	return u.storage.Ping()
}

func generateRandomString(length int) (string, error) {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}
