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

type URLShortener struct {
	storage repository.Repository
}

func NewURLShortener(storage repository.Repository) *URLShortener {
	return &URLShortener{
		storage: storage,
	}
}

func (u *URLShortener) GenerateShortURLPart(url string) (string, error) {
	shortURL := generateShort(u.storage)
	if shortURL == nil {
		return "", errors.New("failed to generate short url")
	}
	u.storage.Save(*shortURL, url)
	return *shortURL, nil
}

func (u *URLShortener) GetURLByShortURLPart(shortURLPart string) (string, error) {
	resultURL := u.storage.GetByShortURL(shortURLPart)
	if resultURL == "" {
		return "", errors.New("url not found")
	}
	return resultURL, nil
}

func generateShort(rep repository.Repository) *string {
	for i := 0; i < maxAttemptsCount; i++ {
		shortURL := generateRandomString(shortURLLength)
		if rep.GetByShortURL(shortURL) == "" {
			return &shortURL
		}
	}
	return nil
}

func generateRandomString(length int) string {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		randomNumber := rand.Intn(len(charset))
		result[i] = charset[randomNumber]
	}
	return string(result)
}
