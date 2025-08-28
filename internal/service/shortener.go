package service

import (
	"context"
	"crypto/rand"
	"errors"
	"github.com/bezjen/shortener/internal/model"
	"github.com/bezjen/shortener/internal/repository"
	"math/big"
	"sync"
)

const (
	shortURLLength   = 8
	charset          = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	maxAttemptsCount = 10
)

var ErrGenerate = errors.New("failed to generate short url")

type Shortener interface {
	GenerateShortURLPart(ctx context.Context, userID string, url string) (string, error)
	GenerateShortURLPartBatch(ctx context.Context, userID string,
		urls []model.ShortenBatchRequestItem) ([]model.ShortenBatchResponseItem, error)
	DeleteUserShortURLsBatch(ctx context.Context, userID string, shortURLs []string) error
	GetURLByShortURLPart(ctx context.Context, shortURLPart string) (*model.URL, error)
	GetURLsByUserID(ctx context.Context, userID string) ([]model.URL, error)
	PingRepository(ctx context.Context) error
}

type URLShortener struct {
	storage     repository.Repository
	deleteQueue chan deleteTask
	wg          sync.WaitGroup
}

type deleteTask struct {
	userID    string
	shortURLs []string
}

func NewURLShortener(storage repository.Repository) *URLShortener {
	shortener := &URLShortener{
		storage:     storage,
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
			// TODO: log err
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
