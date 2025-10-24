// Package service provides business logic for URL shortening service.
package service

import (
	"context"
	"crypto/rand"
	"errors"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/model"
	"github.com/bezjen/shortener/internal/repository"
	"go.uber.org/zap"
	"math/big"
	"sync"
)

const (
	shortURLLength   = 8
	charset          = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	maxAttemptsCount = 10
)

// ErrGenerate is returned when short URL generation fails after maximum attempts.
var ErrGenerate = errors.New("failed to generate short url")

// Shortener defines the main interface for URL shortening operations.
// It provides methods for creating, retrieving, and managing short URLs.
type Shortener interface {
	// GenerateShortURLPart creates a short URL identifier for the given original URL.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeouts
	//   - userID: identifier of the user creating the short URL
	//   - url: original URL to be shortened
	//
	// Returns:
	//   - string: generated short URL identifier
	//   - error: error if generation fails
	GenerateShortURLPart(ctx context.Context, userID string, url string) (string, error)

	// GenerateShortURLPartBatch creates multiple short URLs in a single batch operation.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeouts
	//   - userID: identifier of the user creating the short URLs
	//   - urls: slice of URLs to be shortened with correlation IDs
	//
	// Returns:
	//   - []model.ShortenBatchResponseItem: slice of generated short URLs with correlation IDs
	//   - error: error if batch generation fails
	GenerateShortURLPartBatch(ctx context.Context, userID string,
		urls []model.ShortenBatchRequestItem) ([]model.ShortenBatchResponseItem, error)

	// DeleteUserShortURLsBatch marks user's short URLs as deleted using async processing.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeouts
	//   - userID: identifier of the user owning the short URLs
	//   - shortURLs: slice of short URL identifiers to mark as deleted
	//
	// Returns:
	//   - error: error if the deletion queue is full or processing fails
	DeleteUserShortURLsBatch(ctx context.Context, userID string, shortURLs []string) error

	// GetURLByShortURLPart retrieves the original URL by its short identifier.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeouts
	//   - shortURLPart: short URL identifier to look up
	//
	// Returns:
	//   - *model.URL: found URL object containing original URL and metadata
	//   - error: error if URL is not found or lookup fails
	GetURLByShortURLPart(ctx context.Context, shortURLPart string) (*model.URL, error)

	// GetURLsByUserID retrieves all URLs created by a specific user.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeouts
	//   - userID: user identifier to look up URLs for
	//
	// Returns:
	//   - []model.URL: slice of URLs created by the user
	//   - error: error if lookup fails
	GetURLsByUserID(ctx context.Context, userID string) ([]model.URL, error)

	// PingRepository checks the connectivity to the underlying data storage.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeouts
	//
	// Returns:
	//   - error: error if storage is unreachable
	PingRepository(ctx context.Context) error
}

// URLShortener implements the Shortener interface with background deletion workers.
// It provides URL shortening functionality with async batch deletion support.
type URLShortener struct {
	storage     repository.Repository
	logger      *logger.Logger
	deleteQueue chan deleteTask
	wg          sync.WaitGroup
}

type deleteTask struct {
	userID    string
	shortURLs []string
}

// NewURLShortener creates a new URLShortener instance with background workers.
//
// Parameters:
//   - storage: repository implementation for data persistence
//   - logger: logger instance for application logging
//
// Returns:
//   - *URLShortener: initialized URL shortener service
func NewURLShortener(storage repository.Repository, logger *logger.Logger) *URLShortener {
	shortener := &URLShortener{
		storage:     storage,
		logger:      logger,
		deleteQueue: make(chan deleteTask, 1000),
	}
	for i := 0; i < 5; i++ {
		shortener.wg.Add(1)
		go shortener.deleteWorker()
	}
	return shortener
}

func (u *URLShortener) GenerateShortURLPart(ctx context.Context, userID string, url string) (string, error) {
	for i := 0; i < maxAttemptsCount; i++ {
		shortURL, err := generateRandomString(shortURLLength)
		if err != nil {
			return "", err
		}
		err = u.storage.Save(ctx, userID, *model.NewURL(shortURL, url))
		if err != nil {
			if errors.Is(err, repository.ErrShortURLConflict) {
				continue
			}
			return "", err
		}
		return shortURL, nil
	}
	return "", ErrGenerate
}

func (u *URLShortener) GenerateShortURLPartBatch(ctx context.Context,
	userID string,
	urls []model.ShortenBatchRequestItem,
) ([]model.ShortenBatchResponseItem, error) {
	for i := 0; i < maxAttemptsCount; i++ {
		var generatedURLs []model.URL
		var response []model.ShortenBatchResponseItem
		for _, url := range urls {
			shortURL, err := generateRandomString(shortURLLength)
			if err != nil {
				return nil, err
			}
			generatedURLs = append(generatedURLs, *model.NewURL(shortURL, url.OriginalURL))
			response = append(response, *model.NewShortenBatchResponseItem(url.CorrelationID, shortURL))
		}
		err := u.storage.SaveBatch(ctx, userID, generatedURLs)
		if err != nil {
			if errors.Is(err, repository.ErrShortURLConflict) {
				continue
			}
			return nil, err
		}
		return response, nil
	}
	return nil, ErrGenerate
}

func (u *URLShortener) DeleteUserShortURLsBatch(_ context.Context, userID string, shortURLs []string) error {
	select {
	case u.deleteQueue <- deleteTask{userID: userID, shortURLs: shortURLs}:
		return nil
	default:
		return errors.New("delete queue is full")
	}
}

func (u *URLShortener) GetURLByShortURLPart(ctx context.Context, shortURLPart string) (*model.URL, error) {
	resultURL, err := u.storage.GetByShortURL(ctx, shortURLPart)
	if err != nil {
		return nil, err
	}
	return resultURL, nil
}

func (u *URLShortener) GetURLsByUserID(ctx context.Context, userID string) ([]model.URL, error) {
	return u.storage.GetByUserID(ctx, userID)
}

func (u *URLShortener) PingRepository(ctx context.Context) error {
	return u.storage.Ping(ctx)
}

func (u *URLShortener) Close() {
	close(u.deleteQueue)
	u.wg.Wait()
}

func (u *URLShortener) deleteWorker() {
	defer u.wg.Done()

	for task := range u.deleteQueue {
		err := u.storage.DeleteBatch(context.Background(), task.userID, task.shortURLs)
		if err != nil {
			u.logger.Error("Failed to delete short urls for user",
				zap.Error(err),
				zap.Strings("shortURLs", task.shortURLs),
				zap.String("userID", task.userID))
			continue
		}
	}
}

func generateRandomString(length int) (string, error) {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}
