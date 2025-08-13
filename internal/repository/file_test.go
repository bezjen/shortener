package repository

import (
	"context"
	"errors"
	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/model"
	"os"
	"testing"
)

func testConfig() config.Config {
	return config.Config{
		ServerAddr:      "localhost:8080",
		BaseURL:         "http://localhost:8080",
		LogLevel:        "info",
		FileStoragePath: "./storage.json",
	}
}

func TestFileRepositorySuccess(t *testing.T) {
	repo, cleanup := setupFileRepository(t)
	defer cleanup()

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
			err := repo.Save(context.TODO(), *model.NewURL(tt.shortURL, tt.originalURL))
			if err != nil {
				t.Fatalf("Save failed: %v", err)
			}
			result, err := repo.GetByShortURL(context.TODO(), tt.shortURL)
			if err != nil {
				t.Fatalf("GetByShortURL failed: %v", err)
			}
			if result != tt.originalURL {
				t.Errorf("got %q, want %q", result, tt.originalURL)
			}
		})
	}
}

func TestFileRepositoryErrConflict(t *testing.T) {
	repo, cleanup := setupFileRepository(t)
	defer cleanup()

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
			err := repo.Save(context.TODO(), *model.NewURL(tt.shortURL, tt.originalURL))
			if err != nil {
				t.Fatalf("Save failed: %v", err)
			}
			err = repo.Save(context.TODO(), *model.NewURL(tt.shortURL, tt.originalURL))
			if !errors.Is(err, ErrConflict) {
				t.Errorf("got %v, want %v", err, ErrConflict)
			}
		})
	}

	t.Run("Get not-existed URL (ErrNotFound)", func(t *testing.T) {
		_, err := repo.GetByShortURL(context.TODO(), "non-existent")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want %v", err, ErrNotFound)
		}
	})
}

func TestFileRepositoryErrNotFound(t *testing.T) {
	repo, cleanup := setupFileRepository(t)
	defer cleanup()

	t.Run("Get not-existed URL (ErrNotFound)", func(t *testing.T) {
		_, err := repo.GetByShortURL(context.TODO(), "non-existent")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want %v", err, ErrNotFound)
		}
	})
}

func setupFileRepository(t *testing.T) (*FileRepository, func()) {
	testCfg := testConfig()

	_ = os.Remove(testCfg.FileStoragePath)
	repo, err := NewFileRepository(testCfg)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	cleanup := func() {
		err := repo.fileStorage.Close()
		if err != nil {
			t.Fatalf("Failed to close file storage: %v", err)
		}
		err = os.Remove(testCfg.FileStoragePath)
		if err != nil {
			t.Fatalf("Failed to remove file storage: %v", err)
		}
	}

	return repo, cleanup
}
