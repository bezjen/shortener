package model

import "github.com/google/uuid"

type ShortURLFileDto struct {
	ID          uuid.UUID `json:"uuid"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
}

type URL struct {
	ShortURL    string
	OriginalURL string
}

func NewURL(shortURL string, originalURL string) *URL {
	return &URL{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
}
