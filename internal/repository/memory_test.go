package repository

import (
	"errors"
	"testing"
)

func TestInMemoryRepositorySuccess(t *testing.T) {
	repo := NewInMemoryRepository()

	tests := []struct {
		name        string
		shortURL    string
		originalURL string
	}{
		{
			name:        "Simple positive case",
			shortURL:    "qwerty12",
			originalURL: "https://practicum.yandex.ru/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Save(tt.shortURL, tt.originalURL)
			if err != nil {
				t.Fatalf("Save failed: %v", err)
			}
			result, err := repo.GetByShortURL(tt.shortURL)
			if err != nil {
				t.Fatalf("GetByShortURL failed: %v", err)
			}
			if result != tt.originalURL {
				t.Errorf("got %q, want %q", result, tt.originalURL)
			}
		})
	}
}

func TestInMemoryRepositoryErrConflict(t *testing.T) {
	repo := NewInMemoryRepository()

	tests := []struct {
		name        string
		shortURL    string
		originalURL string
	}{
		{
			name:        "Save the same url twice (ErrConflict)",
			shortURL:    "qwerty12",
			originalURL: "https://practicum.yandex.ru/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Save(tt.shortURL, tt.originalURL)
			if err != nil {
				t.Fatalf("Save failed: %v", err)
			}
			err = repo.Save(tt.shortURL, tt.originalURL)
			if !errors.Is(err, ErrConflict) {
				t.Errorf("got %v, want %v", err, ErrConflict)
			}
		})
	}

	t.Run("Get not-existed URL (ErrNotFound)", func(t *testing.T) {
		_, err := repo.GetByShortURL("non-existent")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want %v", err, ErrNotFound)
		}
	})
}

func TestInMemoryRepositoryErrNotFound(t *testing.T) {
	repo := NewInMemoryRepository()

	t.Run("Get not-existed URL (ErrNotFound)", func(t *testing.T) {
		_, err := repo.GetByShortURL("non-existent")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want %v", err, ErrNotFound)
		}
	})
}
