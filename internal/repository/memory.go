package repository

import "sync"

type InMemoryRepository struct {
	storage map[string]string
	mu      sync.RWMutex
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		storage: make(map[string]string),
	}
}

func (m *InMemoryRepository) Save(shortURL string, url string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.storage[shortURL]; exists {
		return ErrConflict
	}
	m.storage[shortURL] = url
	return nil
}

func (m *InMemoryRepository) GetByShortURL(shortURL string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	storedURL, exists := m.storage[shortURL]
	if !exists {
		return "", ErrNotFound
	}
	return storedURL, nil
}

func (m *InMemoryRepository) Close() error {
	return nil
}

func (m *InMemoryRepository) Ping() error {
	return nil
}
