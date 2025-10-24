// Package repository provides data storage implementations for the URL shortening service.
// It includes file-based and PostgreSQL storage backends.
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/model"
	"github.com/google/uuid"
	"io"
	"os"
	"sync"
)

// FileRepository implements Repository interface for file-based storage.
// It stores URL mappings in a JSON file with in-memory caching for performance.
type FileRepository struct {
	fileStorage   os.File
	encoder       json.Encoder
	decoder       json.Decoder
	memoryStorage map[string]model.ShortURLFileDto
	mu            sync.RWMutex
}

// NewFileRepository creates a new FileRepository instance.
// It initializes file storage and loads existing data into memory.
//
// Parameters:
//   - cfg: application configuration containing file storage path
//
// Returns:
//   - *FileRepository: initialized file repository
//   - error: error if file operations fail during initialization
func NewFileRepository(cfg config.Config) (*FileRepository, error) {
	fileStorage, err := os.OpenFile(cfg.FileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	decoder := *json.NewDecoder(fileStorage)
	memoryStorage, err := loadFileData(decoder)
	if err != nil {
		return nil, err
	}
	return &FileRepository{
		fileStorage:   *fileStorage,
		memoryStorage: memoryStorage,
		encoder:       *json.NewEncoder(fileStorage),
		decoder:       decoder,
	}, nil
}

// Save stores a URL mapping in file storage and memory cache.
// Returns ErrShortURLConflict if the short URL already exists.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - userID: identifier of the user creating the URL (not used in file storage)
//   - url: URL object containing short and original URLs
//
// Returns:
//   - error: error if storage operation fails or URL conflict occurs
func (f *FileRepository) Save(_ context.Context, _ string, url model.URL) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, exists := f.memoryStorage[url.ShortURL]; exists {
		return ErrShortURLConflict
	}
	shortURLDto, err := f.saveShortURLDtoToStorage(url)
	if err != nil {
		return err
	}
	f.memoryStorage[url.ShortURL] = *shortURLDto
	return nil
}

// SaveBatch stores multiple URL mappings in a single operation.
// Implements atomic batch save - if any URL fails, all changes are rolled back.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - userID: identifier of the user creating the URLs (not used in file storage)
//   - urls: slice of URL objects to store
//
// Returns:
//   - error: error if any storage operation fails or URL conflict occurs
func (f *FileRepository) SaveBatch(_ context.Context, _ string, urls []model.URL) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, url := range urls {
		if _, exists := f.memoryStorage[url.ShortURL]; exists {
			return ErrShortURLConflict
		}
	}
	var savedKeys []string
	for _, url := range urls {
		shortURLDto, err := f.saveShortURLDtoToStorage(url)
		if err != nil {
			for _, savedKey := range savedKeys {
				delete(f.memoryStorage, savedKey)
			}
			return err
		}
		savedKeys = append(savedKeys, url.ShortURL)
		f.memoryStorage[url.ShortURL] = *shortURLDto
	}
	return nil
}

// GetByShortURL retrieves the original URL by its short identifier.
// Uses in-memory cache for fast lookups.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - shortURL: short URL identifier to look up
//
// Returns:
//   - *model.URL: found URL object
//   - error: error if URL is not found
func (f *FileRepository) GetByShortURL(_ context.Context, shortURL string) (*model.URL, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	storedShortURLDto, exists := f.memoryStorage[shortURL]
	if !exists {
		return nil, ErrNotFound
	}
	return model.NewURL(storedShortURLDto.ShortURL, storedShortURLDto.OriginalURL), nil
}

// GetByUserID retrieves all URLs created by a specific user.
// Not implemented for file storage.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - userID: user identifier to look up URLs for
//
// Returns:
//   - []model.URL: slice of URLs (always empty for file storage)
//   - error: always returns "method not implemented" error
func (f *FileRepository) GetByUserID(_ context.Context, _ string) ([]model.URL, error) {
	return nil, fmt.Errorf("method not implemented")
}

// DeleteBatch marks multiple short URLs as deleted.
// Not implemented for file storage.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//   - userID: identifier of the user owning the URLs
//   - shortURLs: slice of short URL identifiers to delete
//
// Returns:
//   - error: always returns "method not implemented" error
func (f *FileRepository) DeleteBatch(_ context.Context, _ string, _ []string) error {
	return fmt.Errorf("method not implemented")
}

// Ping checks the connectivity to file storage.
// Always returns nil for file storage as file operations are checked during initialization.
//
// Parameters:
//   - ctx: context for request cancellation and timeouts
//
// Returns:
//   - error: always nil for file storage
func (f *FileRepository) Ping(_ context.Context) error {
	return nil
}

// Close closes the file storage and releases resources.
// Should be called when the repository is no longer needed.
//
// Returns:
//   - error: error if file closing fails
func (f *FileRepository) Close() error {
	return f.fileStorage.Close()
}

// saveShortURLDtoToStorage writes a URL mapping to the file storage.
// Generates a UUID for each entry and appends it as JSON line to the file.
//
// Parameters:
//   - url: URL object to store
//
// Returns:
//   - *model.ShortURLFileDto: DTO containing stored URL data
//   - error: error if UUID generation or file writing fails
func (f *FileRepository) saveShortURLDtoToStorage(url model.URL) (*model.ShortURLFileDto, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	shortURLDto := model.ShortURLFileDto{
		ID:          id,
		ShortURL:    url.ShortURL,
		OriginalURL: url.OriginalURL,
	}
	err = f.encoder.Encode(&shortURLDto)
	if err != nil {
		return nil, err
	}
	return &shortURLDto, nil
}

// loadFileData reads existing URL mappings from file storage into memory.
// Parses JSON lines from the file and builds an in-memory map for fast access.
//
// Parameters:
//   - decoder: JSON decoder configured for the storage file
//
// Returns:
//   - map[string]model.ShortURLFileDto: map of short URL to URL DTO
//   - error: error if file reading or JSON parsing fails
func loadFileData(decoder json.Decoder) (map[string]model.ShortURLFileDto, error) {
	memoryStorage := make(map[string]model.ShortURLFileDto)
	for {
		var dto model.ShortURLFileDto
		err := decoder.Decode(&dto)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		memoryStorage[dto.ShortURL] = dto
	}
	return memoryStorage, nil
}
