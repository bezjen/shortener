package repository

import (
	"context"
	"errors"
	"github.com/bezjen/shortener/internal/model"
	"testing"
)

func TestInMemoryRepositorySuccess(t *testing.T) {
	repo := NewInMemoryRepository()

	tests := []struct {
		name string
		url  model.URL
	}{
		{
			name: "Simple positive case",
			url:  *model.NewURL("qwerty12", "https://practicum.yandex.ru/"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Save(context.TODO(), "", tt.url)
			if err != nil {
				t.Fatalf("Save failed: %v", err)
			}
			result, err := repo.GetByShortURL(context.TODO(), tt.url.ShortURL)
			if err != nil {
				t.Fatalf("GetByShortURL failed: %v", err)
			}
			if result != tt.url.OriginalURL {
				t.Errorf("got %q, want %q", result, tt.url.OriginalURL)
			}
		})
	}
}

func TestInMemoryRepositorySaveBatchSuccess(t *testing.T) {
	repo := NewInMemoryRepository()

	tests := []struct {
		name  string
		batch []model.URL
	}{
		{
			name: "Simple positive case",
			batch: []model.URL{
				*model.NewURL("qwerty12", "https://practicum.yandex.ru/"),
				*model.NewURL("qwerty13", "https://practicum.yandex.ru/"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.SaveBatch(context.TODO(), "", tt.batch)
			if err != nil {
				t.Fatalf("Save failed: %v", err)
			}
			for _, url := range tt.batch {
				result, err := repo.GetByShortURL(context.TODO(), url.ShortURL)
				if err != nil {
					t.Fatalf("GetByShortURL failed: %v", err)
				}
				if result != url.OriginalURL {
					t.Errorf("got %q, want %q", result, url.OriginalURL)
				}
			}
		})
	}
}

func TestInMemoryRepositoryErrConflict(t *testing.T) {
	repo := NewInMemoryRepository()

	tests := []struct {
		name string
		url  model.URL
	}{
		{
			name: "Save the same url twice (ErrShortURLConflict)",
			url:  *model.NewURL("qwerty12", "https://practicum.yandex.ru/"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Save(context.TODO(), "", tt.url)
			if err != nil {
				t.Fatalf("Save failed: %v", err)
			}
			err = repo.Save(context.TODO(), "", tt.url)
			if !errors.Is(err, ErrShortURLConflict) {
				t.Errorf("got %v, want %v", err, ErrShortURLConflict)
			}
		})
	}

	t.Run("Get not-existed OriginalURL (ErrNotFound)", func(t *testing.T) {
		_, err := repo.GetByShortURL(context.TODO(), "non-existent")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want %v", err, ErrNotFound)
		}
	})
}

func TestInMemoryRepositorySaveBatchErrConflict(t *testing.T) {
	repo := NewInMemoryRepository()

	tests := []struct {
		name  string
		batch []model.URL
	}{
		{
			name: "Save the same url twice (ErrShortURLConflict)",
			batch: []model.URL{
				*model.NewURL("qwerty12", "https://practicum.yandex.ru/"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.SaveBatch(context.TODO(), "", tt.batch)
			if err != nil {
				t.Fatalf("Save failed: %v", err)
			}
			err = repo.SaveBatch(context.TODO(), "", tt.batch)
			if !errors.Is(err, ErrShortURLConflict) {
				t.Errorf("got %v, want %v", err, ErrShortURLConflict)
			}
		})
	}

	t.Run("Get not-existed OriginalURL (ErrNotFound)", func(t *testing.T) {
		_, err := repo.GetByShortURL(context.TODO(), "non-existent")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want %v", err, ErrNotFound)
		}
	})
}

func TestInMemoryRepositoryErrNotFound(t *testing.T) {
	repo := NewInMemoryRepository()

	t.Run("Get not-existed OriginalURL (ErrNotFound)", func(t *testing.T) {
		_, err := repo.GetByShortURL(context.TODO(), "non-existent")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want %v", err, ErrNotFound)
		}
	})
}
