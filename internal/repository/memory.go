package repository

import (
	"context"
	"github.com/bezjen/shortener/internal/model"
	"sync"
)

type InMemoryRepository struct {
	storage map[string]string
	mu      sync.RWMutex
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		storage: make(map[string]string),
	}
}

func (m *InMemoryRepository) Save(_ context.Context, url model.URL) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.storage[url.ShortURL]; exists {
		return ErrShortURLConflict
	}
	m.storage[url.ShortURL] = url.OriginalURL
	return nil
}

func (m *InMemoryRepository) SaveBatch(_ context.Context, urls []model.URL) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, url := range urls {
		if _, exists := m.storage[url.ShortURL]; exists {
			return ErrShortURLConflict
		}
	}
	for _, url := range urls {
		m.storage[url.ShortURL] = url.OriginalURL
	}
	return nil
}

func (m *InMemoryRepository) GetByShortURL(_ context.Context, shortURL string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	storedURL, exists := m.storage[shortURL]
	if !exists {
		return "", ErrNotFound
	}
	return storedURL, nil
}

func (m *InMemoryRepository) Ping(_ context.Context) error {
	return nil
}

func (m *InMemoryRepository) Close() error {
	return nil
}
