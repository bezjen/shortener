package model

import "github.com/google/uuid"

type ShortURLDto struct {
	Id          uuid.UUID `json:"uuid"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
}
