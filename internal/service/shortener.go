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

type UrlShortener struct {
	storage repository.Repository
}

func NewUrlShortener(storage repository.Repository) *UrlShortener {
	return &UrlShortener{
		storage: storage,
	}
}

func (u *UrlShortener) GenerateShortUrlPart(url string) (string, error) {
	shortUrl := generateShort(u.storage)
	if shortUrl == nil {
		return "", errors.New("failed to generate short url")
	}
	u.storage.Save(*shortUrl, url)
	return *shortUrl, nil
}

func (u *UrlShortener) GetUrlByShortUrlPart(shortUrlPart string) (string, error) {
	resultUrl := u.storage.GetByShortUrl(shortUrlPart)
	if resultUrl == "" {
		return "", errors.New("url not found")
	}
	return resultUrl, nil
}

func generateShort(rep repository.Repository) *string {
	for i := 0; i < maxAttemptsCount; i++ {
		shortUrl := generateRandomString(shortUrlLength)
		if rep.GetByShortUrl(shortUrl) == "" {
			return &shortUrl
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
