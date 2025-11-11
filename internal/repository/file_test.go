package repository

import (
	"context"
	"errors"
	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/model"
	"github.com/stretchr/testify/assert"
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

func TestFileRepositoryErrConflict(t *testing.T) {
	repo, cleanup := setupFileRepository(t)
	defer cleanup()

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

func TestFileRepositoryErrNotFound(t *testing.T) {
	repo, cleanup := setupFileRepository(t)
	defer cleanup()

	t.Run("Get not-existed OriginalURL (ErrNotFound)", func(t *testing.T) {
		_, err := repo.GetByShortURL(context.TODO(), "non-existent")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want %v", err, ErrNotFound)
		}
	})
}

func TestFileRepositorySaveBatch(t *testing.T) {
	repo, cleanup := setupFileRepository(t)
	defer cleanup()

	tests := []struct {
		name  string
		batch []model.URL
	}{
		{
			name: "Save batch successfully",
			batch: []model.URL{
				*model.NewURL("qwerty12", "https://practicum.yandex.ru/"),
				*model.NewURL("qwerty13", "https://example.com/"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.SaveBatch(context.TODO(), "", tt.batch)
			assert.NoError(t, err)

			for _, url := range tt.batch {
				result, err := repo.GetByShortURL(context.TODO(), url.ShortURL)
				assert.NoError(t, err)
				assert.Equal(t, &url, result)
			}
		})
	}
}

func TestFileRepositorySaveBatchErrConflict(t *testing.T) {
	repo, cleanup := setupFileRepository(t)
	defer cleanup()

	existingURL := model.NewURL("qwerty12", "https://practicum.yandex.ru/")
	err := repo.Save(context.TODO(), "", *existingURL)
	assert.NoError(t, err)

	batch := []model.URL{
		*existingURL,
		*model.NewURL("qwerty13", "https://example.com/"),
	}

	err = repo.SaveBatch(context.TODO(), "", batch)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrShortURLConflict))
}

func TestFileRepositoryGetByUserID(t *testing.T) {
	repo, cleanup := setupFileRepository(t)
	defer cleanup()

	urls, err := repo.GetByUserID(context.TODO(), "user1")
	assert.Error(t, err)
	assert.Nil(t, urls)
	assert.Equal(t, "method not implemented", err.Error())
}

func TestFileRepositoryDeleteBatch(t *testing.T) {
	repo, cleanup := setupFileRepository(t)
	defer cleanup()

	err := repo.DeleteBatch(context.TODO(), "user1", []string{"qwerty12"})
	assert.Error(t, err)
	assert.Equal(t, "method not implemented", err.Error())
}

func TestFileRepositoryPing(t *testing.T) {
	repo, cleanup := setupFileRepository(t)
	defer cleanup()

	err := repo.Ping(context.TODO())
	assert.NoError(t, err)
}

func TestFileRepositoryClose(t *testing.T) {
	repo, _ := setupFileRepository(t)

	err := repo.Close()
	assert.NoError(t, err)
	err = os.Remove(testConfig().FileStoragePath)
	if err != nil {
		t.Fatalf("Failed to remove file storage: %v", err)
	}
}

func TestFileRepositorySaveShortURLDtoToStorage(t *testing.T) {
	repo, cleanup := setupFileRepository(t)
	defer cleanup()

	url := model.URL{
		ShortURL:    "test123",
		OriginalURL: "https://test.com",
	}

	dto, err := repo.saveShortURLDtoToStorage(url)
	assert.NoError(t, err)
	assert.Equal(t, url.ShortURL, dto.ShortURL)
	assert.Equal(t, url.OriginalURL, dto.OriginalURL)
	assert.NotEmpty(t, dto.ID)
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
