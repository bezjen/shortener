// Package model provides data models and structures for the URL shortening service.
package model

import "github.com/google/uuid"

// ShortURLFileDto represents the data structure for URL storage in file-based repository.
// Used for JSON serialization/deserialization in file storage operations.
//
// Example JSON:
//
//	{
//	  "uuid": "123e4567-e89b-12d3-a456-426614174000",
//	  "short_url": "abc123",
//	  "original_url": "https://example.com",
//	  "user_id": "user-123"
//	}
type ShortURLFileDto struct {
	// ID is the unique identifier for the URL record.
	// Generated using UUID v1 for file storage.
	// Example: "123e4567-e89b-12d3-a456-426614174000"
	ID uuid.UUID `json:"uuid"`

	// ShortURL is the shortened URL identifier.
	// Example: "abc123"
	ShortURL string `json:"short_url"`

	// OriginalURL is the original URL that was shortened.
	// Example: "https://example.com"
	OriginalURL string `json:"original_url"`

	// UserID is the identifier of the user who created the short URL.
	// Example: "user-123"
	UserID string `json:"user_id"`
}

// URL represents the core URL entity in the URL shortening service.
// Used throughout the application for URL operations and business logic.
type URL struct {
	// ShortURL is the shortened URL identifier.
	// Example: "abc123"
	ShortURL string

	// OriginalURL is the original URL that was shortened.
	// Example: "https://example.com"
	OriginalURL string

	// IsDeleted indicates whether the URL has been soft-deleted.
	// Deleted URLs return 410 Gone status instead of redirecting.
	// Default: false
	IsDeleted bool
}

// NewURL creates a new URL instance.
// Constructor function for URL entities with default values.
//
// Parameters:
//   - shortURL: shortened URL identifier
//   - originalURL: original URL that was shortened
//
// Returns:
//   - *URL: initialized URL entity with IsDeleted set to false
func NewURL(shortURL string, originalURL string) *URL {
	return &URL{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		IsDeleted:   false,
	}
}
