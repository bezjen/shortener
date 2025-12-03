// Package repository provides data storage implementations for the URL shortening service.
package repository

import (
	"context"
	"fmt"
	"github.com/bezjen/shortener/internal/model"
	"github.com/pkg/errors"
	"sync"
)

// InMemoryRepository implements Repository interface for in-memory storage.
// It stores URL mappings in a concurrent map without persistence.
// Suitable for testing and development environments.
type InMemoryRepository struct {
	storage map[string]string
	mu      *sync.RWMutex
}

// NewInMemoryRepository creates a new InMemoryRepository instance.
// Initializes an empty in-memory storage map.
//
// Returns:
//   - *InMemoryRepository: initialized in-memory repository
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		storage: make(map[string]string),
		mu:      &sync.RWMutex{},
	}
}

// Save stores a URL mapping in memory.
// Returns ErrShortURLConflict if the short URL already exists.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts (not used)
//   - userID: identifier of the user creating the URL (not used in memory storage)
//   - url: URL object containing short and original URLs
//
// Returns:
//   - error: error if URL conflict occurs
func (m *InMemoryRepository) Save(_ context.Context, _ string, url model.URL) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.storage[url.ShortURL]; exists {
		return ErrShortURLConflict
	}
	m.storage[url.ShortURL] = url.OriginalURL
	return nil
}

// SaveBatch stores multiple URL mappings in a single atomic operation.
// If any URL conflicts, no URLs are saved and ErrShortURLConflict is returned.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts (not used)
//   - userID: identifier of the user creating the URLs (not used in memory storage)
//   - urls: slice of URL objects to store
//
// Returns:
//   - error: error if any URL conflict occurs
func (m *InMemoryRepository) SaveBatch(_ context.Context, _ string, urls []model.URL) error {
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

// GetByShortURL retrieves the original URL by its short identifier.
// Uses read lock for concurrent access.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts (not used)
//   - shortURL: short URL identifier to look up
//
// Returns:
//   - *model.URL: found URL object
//   - error: error if URL is not found
func (m *InMemoryRepository) GetByShortURL(_ context.Context, shortURL string) (*model.URL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	storedURL, exists := m.storage[shortURL]
	if !exists {
		return nil, ErrNotFound
	}
	return model.NewURL(shortURL, storedURL), nil
}

// GetByUserID retrieves all URLs created by a specific user.
// Not implemented for in-memory storage as it doesn't track user ownership.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - userID: user identifier to look up URLs for
//
// Returns:
//   - []model.URL: slice of URLs (always empty for in-memory storage)
//   - error: always returns "method not implemented" error
func (m *InMemoryRepository) GetByUserID(_ context.Context, _ string) ([]model.URL, error) {
	return nil, fmt.Errorf("method not implemented")
}

// DeleteBatch marks multiple short URLs as deleted.
// Not implemented for in-memory storage.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - userID: identifier of the user owning the URLs
//   - shortURLs: slice of short URL identifiers to delete
//
// Returns:
//   - error: always returns "method not implemented" error
func (m *InMemoryRepository) DeleteBatch(_ context.Context, _ string, _ []string) error {
	return fmt.Errorf("method not implemented")
}

// GetStats retrieves service statistics including total URLs and unique users.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//
// Returns:
//   - int: total number of shortened URLs in the service
//   - int: total number of unique users in the service
//   - error: any error that occurred during statistics retrieval
//
// Note: Access to this method should be restricted to trusted networks.
func (m *InMemoryRepository) GetStats(_ context.Context) (urlsCount int, usersCount int, err error) {
	return 0, 0, errors.New("method not implemented")
}

// Ping checks the connectivity to in-memory storage.
// Always returns nil as in-memory storage is always available.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//
// Returns:
//   - error: always nil for in-memory storage
func (m *InMemoryRepository) Ping(_ context.Context) error {
	return nil
}

// Close releases resources used by the in-memory repository.
// For in-memory storage, this is a no-op that always returns nil.
//
// Returns:
//   - error: always nil
func (m *InMemoryRepository) Close() error {
	return nil
}
