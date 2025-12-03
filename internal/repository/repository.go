// Package repository provides data storage implementations for the URL shortening service.
// It defines the Repository interface and common error types used by all storage implementations.

//go:generate mockery --name=Repository --output=../mocks --outpkg=mocks --case=underscore
package repository

import (
	"context"
	"errors"
	"github.com/bezjen/shortener/internal/model"
)

// Common repository error types used by all storage implementations.
var (
	// ErrNotFound is returned when a requested record is not found in storage.
	ErrNotFound = errors.New("record not found")

	// ErrShortURLConflict is returned when attempting to save a short URL that already exists.
	ErrShortURLConflict = errors.New("record with short url already exists")
)

// Repository defines the interface for URL storage operations.
// Implementations provide persistence for URL mappings with support for
// basic CRUD operations, batch processing, and user-specific queries.
type Repository interface {
	// Save stores a single URL mapping in the repository.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeouts
	//   - userID: identifier of the user creating the URL
	//   - url: URL object containing short and original URLs
	//
	// Returns:
	//   - error: error if storage operation fails or URL conflict occurs
	Save(ctx context.Context, userID string, url model.URL) error

	// SaveBatch stores multiple URL mappings in a single atomic operation.
	// Implementations should ensure that either all URLs are saved or none.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeouts
	//   - userID: identifier of the user creating the URLs
	//   - urls: slice of URL objects to store
	//
	// Returns:
	//   - error: error if any storage operation fails
	SaveBatch(ctx context.Context, userID string, urls []model.URL) error

	// DeleteBatch marks multiple short URLs as deleted.
	// The deletion should be performed asynchronously for better performance.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeouts
	//   - userID: identifier of the user owning the URLs
	//   - shortURLs: slice of short URL identifiers to delete
	//
	// Returns:
	//   - error: error if deletion request cannot be processed
	DeleteBatch(ctx context.Context, userID string, shortURLs []string) error

	// GetByShortURL retrieves the original URL by its short identifier.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeouts
	//   - id: short URL identifier to look up
	//
	// Returns:
	//   - *model.URL: found URL object
	//   - error: error if URL is not found or lookup fails
	GetByShortURL(ctx context.Context, id string) (*model.URL, error)

	// GetByUserID retrieves all URLs created by a specific user.
	// Should only return non-deleted URLs.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeouts
	//   - userID: user identifier to look up URLs for
	//
	// Returns:
	//   - []model.URL: slice of URLs created by the user
	//   - error: error if lookup fails
	GetByUserID(ctx context.Context, userID string) ([]model.URL, error)

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
	GetStats(ctx context.Context) (urlsCount int, usersCount int, err error)

	// Ping checks the connectivity to the underlying storage.
	// Used for health checks and monitoring.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeouts
	//
	// Returns:
	//   - error: error if storage is unreachable
	Ping(ctx context.Context) error

	// Close releases resources used by the repository.
	// Should be called when the repository is no longer needed.
	//
	// Returns:
	//   - error: error if resource cleanup fails
	Close() error
}
