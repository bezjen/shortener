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
	shortURL, err := u.generateShort()
	if err != nil {
		return "", err
	}
	u.storage.Save(shortURL, url)
	return shortURL, nil
}

func (u *URLShortener) GetURLByShortURLPart(shortURLPart string) (string, error) {
	resultURL, err := u.storage.GetByShortURL(shortURLPart)
	if err != nil {
		return "", err
	}
	return resultURL, nil
}

func (u *URLShortener) generateShort() (string, error) {
	for i := 0; i < maxAttemptsCount; i++ {
		shortURL := generateRandomString(shortURLLength)
		_, err := u.storage.GetByShortURL(shortURL)
		if errors.Is(err, repository.ErrNotFound) {
			return shortURL, nil
		}
	}
	return "", errors.New("failed to generate short url")
}

func generateRandomString(length int) string {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		randomNumber := rand.Intn(len(charset))
		result[i] = charset[randomNumber]
	}
	return string(result)
}
