package service

import (
	"errors"
	"github.com/bezjen/shortener/internal/repository"
	"math/rand"
)

const (
	shortURLLength   = 8
	charset          = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	maxAttemptsCount = 10
)

var GenerateError = errors.New("failed to generate short url")

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
		shortURL := generateRandomString(shortURLLength)
		err := u.storage.Save(shortURL, url)
		if err != nil {
			if errors.Is(err, repository.ErrConflict) {
				continue
			}
			return "", GenerateError
		}
		return shortURL, nil
	}
	return "", GenerateError
}

func (u *URLShortener) GetURLByShortURLPart(shortURLPart string) (string, error) {
	resultURL, err := u.storage.GetByShortURL(shortURLPart)
	if err != nil {
		return "", err
	}
	return resultURL, nil
}

func generateRandomString(length int) string {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		randomNumber := rand.Intn(len(charset))
		result[i] = charset[randomNumber]
	}
	return string(result)
}
