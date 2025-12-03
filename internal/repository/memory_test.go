package repository

import (
	"context"
	"errors"
	"github.com/bezjen/shortener/internal/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInMemoryRepositorySuccess(t *testing.T) {
	repo := NewInMemoryRepository()

	tests := []struct {
		name string
		url  *model.URL
	}{
		{
			name: "Simple positive case",
			url:  model.NewURL("qwerty12", "https://practicum.yandex.ru/"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Save(context.TODO(), "", *tt.url)
			if err != nil {
				t.Fatalf("Save failed: %v", err)
			}
			result, err := repo.GetByShortURL(context.TODO(), tt.url.ShortURL)
			if err != nil {
				t.Fatalf("GetByShortURL failed: %v", err)
			}
			assert.Equal(t, tt.url, result)
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
				result, errRepo := repo.GetByShortURL(context.TODO(), url.ShortURL)
				if errRepo != nil {
					t.Fatalf("GetByShortURL failed: %v", errRepo)
				}
				assert.Equal(t, url, *result)
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

func TestInMemoryRepositoryGetByUserID(t *testing.T) {
	repo := NewInMemoryRepository()

	urls, err := repo.GetByUserID(context.TODO(), "user1")
	assert.Error(t, err)
	assert.Nil(t, urls)
	assert.Equal(t, "method not implemented", err.Error())
}

func TestInMemoryRepositoryDeleteBatch(t *testing.T) {
	repo := NewInMemoryRepository()

	err := repo.DeleteBatch(context.TODO(), "user1", []string{"qwerty12"})
	assert.Error(t, err)
	assert.Equal(t, "method not implemented", err.Error())
}

func TestInMemoryRepositoryGetStats(t *testing.T) {
	repo, cleanup := setupFileRepository(t)
	defer cleanup()

	_, _, err := repo.GetStats(context.TODO())
	assert.Error(t, err)
	assert.Equal(t, "method not implemented", err.Error())
}

func TestInMemoryRepositoryPing(t *testing.T) {
	repo := NewInMemoryRepository()

	err := repo.Ping(context.TODO())
	assert.NoError(t, err)
}

func TestInMemoryRepositoryClose(t *testing.T) {
	repo := NewInMemoryRepository()

	err := repo.Close()
	assert.NoError(t, err)
}
