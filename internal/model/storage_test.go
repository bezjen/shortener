package model

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewURL(t *testing.T) {
	shortURL := "abc123"
	originalURL := "https://example.com"

	url := NewURL(shortURL, originalURL)

	if url.ShortURL != shortURL {
		t.Errorf("Expected ShortURL %s, got %s", shortURL, url.ShortURL)
	}

	if url.OriginalURL != originalURL {
		t.Errorf("Expected OriginalURL %s, got %s", originalURL, url.OriginalURL)
	}

	if url.IsDeleted {
		t.Error("New URL should not be deleted")
	}
}

func TestShortURLFileDto(t *testing.T) {
	id := uuid.New()
	dto := ShortURLFileDto{
		ID:          id,
		ShortURL:    "abc123",
		OriginalURL: "https://example.com",
		UserID:      "user-123",
	}

	if dto.ID != id {
		t.Error("UUID mismatch")
	}

	if dto.ShortURL == "" {
		t.Error("ShortURL should not be empty")
	}

	if dto.OriginalURL == "" {
		t.Error("OriginalURL should not be empty")
	}

	if dto.UserID == "" {
		t.Error("UserID should not be empty")
	}
}
