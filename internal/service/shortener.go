package service

import (
	"errors"
	"github.com/bezjen/shortener/internal/repository"
	"math/rand"
)

const (
	shortUrlLength   = 8
	charset          = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	maxAttemptsCount = 10
)

func GenerateShort(rep repository.Repository) (*string, error) {
	for i := 0; i < maxAttemptsCount; i++ {
		shortUrl := generateRandomString(shortUrlLength)
		if rep.GetByShortUrl(shortUrl) == "" {
			return &shortUrl, nil
		}
	}
	return nil, errors.New("failed to generate short url")
}

func generateRandomString(length int) string {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		randomNumber := rand.Intn(len(charset))
		result[i] = charset[randomNumber]
	}
	return string(result)
}
