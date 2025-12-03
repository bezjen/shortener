package model

import (
	"testing"
)

func TestNewShortenBatchRequestItem(t *testing.T) {
	correlationID := "test-id"
	originalURL := "https://example.com"

	item := NewShortenBatchRequestItem(correlationID, originalURL)

	if item.CorrelationID != correlationID {
		t.Errorf("Expected CorrelationID %s, got %s", correlationID, item.CorrelationID)
	}

	if item.OriginalURL != originalURL {
		t.Errorf("Expected OriginalURL %s, got %s", originalURL, item.OriginalURL)
	}
}

func TestNewShortenBatchResponseItem(t *testing.T) {
	correlationID := "test-id"
	shortURL := "http://localhost:8080/abc123"

	item := NewShortenBatchResponseItem(correlationID, shortURL)

	if item.CorrelationID != correlationID {
		t.Errorf("Expected CorrelationID %s, got %s", correlationID, item.CorrelationID)
	}

	if item.ShortURL != shortURL {
		t.Errorf("Expected ShortURL %s, got %s", shortURL, item.ShortURL)
	}
}

func TestNewUserURLResponseItem(t *testing.T) {
	shortURL := "http://localhost:8080/abc123"
	originalURL := "https://example.com"

	item := NewUserURLResponseItem(shortURL, originalURL)

	if item.ShortURL != shortURL {
		t.Errorf("Expected ShortURL %s, got %s", shortURL, item.ShortURL)
	}

	if item.OriginalURL != originalURL {
		t.Errorf("Expected OriginalURL %s, got %s", originalURL, item.OriginalURL)
	}
}

func TestShortenJSONResponse(t *testing.T) {
	t.Run("success response", func(t *testing.T) {
		resp := ShortenJSONResponse{ShortURL: "http://localhost:8080/abc123"}

		if resp.ShortURL == "" {
			t.Error("Expected ShortURL in success response")
		}
		if resp.Error != "" {
			t.Error("Error should be empty in success response")
		}
	})

	t.Run("error response", func(t *testing.T) {
		resp := ShortenJSONResponse{Error: "incorrect url"}

		if resp.ShortURL != "" {
			t.Error("ShortURL should be empty in error response")
		}
		if resp.Error == "" {
			t.Error("Expected Error in error response")
		}
	})
}

func TestNewStatsResponse(t *testing.T) {
	urls := 150
	users := 25

	response := NewStatsResponse(urls, users)

	if response.URLs != urls {
		t.Errorf("Expected URLs %d, got %d", urls, response.URLs)
	}

	if response.Users != users {
		t.Errorf("Expected Users %d, got %d", users, response.Users)
	}
}
