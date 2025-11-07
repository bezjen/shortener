package service_test

import (
	"context"
	"errors"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/mocks"
	"github.com/bezjen/shortener/internal/model"
	"github.com/bezjen/shortener/internal/repository"
	"github.com/bezjen/shortener/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestGenerateShortURLPart(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	mockPositiveRepo := new(mocks.Repository)
	mockPositiveRepo.On("Save", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	mockCollisionRepo := new(mocks.Repository)
	mockCollisionRepo.On("Save", mock.Anything, mock.Anything, mock.Anything).
		Return(repository.ErrShortURLConflict)
	tests := []struct {
		name    string
		storage repository.Repository
		url     string
		want    string
		wantErr error
	}{
		{
			name:    "Simple positive case",
			storage: mockPositiveRepo,
			url:     "https://practicum.yandex.ru/",
			wantErr: nil,
		},
		{
			name:    "Too many collisions case",
			storage: mockCollisionRepo,
			url:     "https://practicum.yandex.ru/",
			wantErr: service.ErrGenerate,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := service.NewURLShortener(tt.storage, testLogger)
			userID, err := uuid.NewUUID()
			if err != nil {
				t.Fatalf("Failed to generate uuid: %v", err)
			}
			shortURL, err := u.GenerateShortURLPart(context.TODO(), userID.String(), tt.url)
			if err != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GenerateShortURLPart() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			assert.Equal(t, 8, len(shortURL))
		})
	}
}

func TestGenerateShortURLPartBatch(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	mockPositiveRepo := new(mocks.Repository)
	mockPositiveRepo.On("SaveBatch", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	mockCollisionRepo := new(mocks.Repository)
	mockCollisionRepo.On("SaveBatch", mock.Anything, mock.Anything, mock.Anything).
		Return(repository.ErrShortURLConflict)
	tests := []struct {
		name    string
		storage repository.Repository
		urls    []model.ShortenBatchRequestItem
		wantErr error
	}{
		{
			name:    "Simple positive case",
			storage: mockPositiveRepo,
			urls: []model.ShortenBatchRequestItem{
				*model.NewShortenBatchRequestItem("123", "https://practicum.yandex.ru/"),
				*model.NewShortenBatchRequestItem("456", "https://practicum.yandex.ru/"),
			},
			wantErr: nil,
		},
		{
			name:    "Too many collisions case",
			storage: mockCollisionRepo,
			urls: []model.ShortenBatchRequestItem{
				*model.NewShortenBatchRequestItem("123", "https://practicum.yandex.ru/"),
			},
			wantErr: service.ErrGenerate,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := service.NewURLShortener(tt.storage, testLogger)
			userID, err := uuid.NewUUID()
			if err != nil {
				t.Fatalf("Failed to generate uuid: %v", err)
			}
			shortURL, err := u.GenerateShortURLPartBatch(context.TODO(), userID.String(), tt.urls)
			if err != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GenerateShortURLPart() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			assert.Equal(t, 2, len(shortURL))
		})
	}
}

func TestGetURLByShortURLPart(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	mockRepoPositive := new(mocks.Repository)
	mockRepoPositive.On("GetByShortURL", mock.Anything, "qwerty12").
		Return(model.NewURL("qwerty12", "https://practicum.yandex.ru/"), nil)
	mockRepoNotFound := new(mocks.Repository)
	mockRepoNotFound.On("GetByShortURL", mock.Anything, "qwerty12").
		Return(nil, repository.ErrNotFound)

	tests := []struct {
		name         string
		storage      repository.Repository
		shortURLPart string
		want         *model.URL
		wantErr      error
	}{
		{
			name:         "Simple positive case",
			storage:      mockRepoPositive,
			shortURLPart: "qwerty12",
			want:         model.NewURL("qwerty12", "https://practicum.yandex.ru/"),
			wantErr:      nil,
		},
		{
			name:         "URL not found case",
			storage:      mockRepoNotFound,
			shortURLPart: "qwerty12",
			want:         nil,
			wantErr:      repository.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := service.NewURLShortener(tt.storage, testLogger)
			got, err := u.GetURLByShortURLPart(context.TODO(), tt.shortURLPart)
			if err != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GenerateShortURLPart() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDeleteUserShortURLsBatch(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	mockRepo := new(mocks.Repository)
	shortener := service.NewURLShortener(mockRepo, testLogger)
	defer shortener.Close() // Важно закрыть в конце

	userID := "test-user"
	shortURLs := []string{"abc123", "def456"}

	// Настраиваем ожидание вызова DeleteBatch
	mockRepo.On("DeleteBatch", mock.Anything, userID, shortURLs).Return(nil)

	err := shortener.DeleteUserShortURLsBatch(context.Background(), userID, shortURLs)
	assert.NoError(t, err)

	// Даем время воркеру обработать задачу
	time.Sleep(100 * time.Millisecond)

	// Проверяем что DeleteBatch был вызван
	mockRepo.AssertCalled(t, "DeleteBatch", mock.Anything, userID, shortURLs)
}

func TestGetURLsByUserID(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	mockRepo := new(mocks.Repository)
	shortener := service.NewURLShortener(mockRepo, testLogger)

	userID := "test-user"
	expectedURLs := []model.URL{
		*model.NewURL("abc123", "https://example.com/1"),
		*model.NewURL("def456", "https://example.com/2"),
	}

	mockRepo.On("GetByUserID", mock.Anything, userID).Return(expectedURLs, nil)

	urls, err := shortener.GetURLsByUserID(context.Background(), userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedURLs, urls)
}

func TestGetURLsByUserID_Error(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	mockRepo := new(mocks.Repository)
	shortener := service.NewURLShortener(mockRepo, testLogger)

	userID := "test-user"
	expectedError := errors.New("database error")

	mockRepo.On("GetByUserID", mock.Anything, userID).Return([]model.URL{}, expectedError)

	urls, err := shortener.GetURLsByUserID(context.Background(), userID)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, urls)
}

func TestPingRepository(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	mockRepo := new(mocks.Repository)
	shortener := service.NewURLShortener(mockRepo, testLogger)

	mockRepo.On("Ping", mock.Anything).Return(nil)

	err := shortener.PingRepository(context.Background())
	assert.NoError(t, err)
}

func TestPingRepository_Error(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	mockRepo := new(mocks.Repository)
	shortener := service.NewURLShortener(mockRepo, testLogger)

	expectedError := errors.New("connection failed")
	mockRepo.On("Ping", mock.Anything).Return(expectedError)

	err := shortener.PingRepository(context.Background())
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

func TestURLShortener_Close(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	mockRepo := new(mocks.Repository)
	shortener := service.NewURLShortener(mockRepo, testLogger)

	// Настраиваем ожидание для вызовов DeleteBatch
	mockRepo.On("DeleteBatch", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Добавляем несколько задач в очередь
	for i := 0; i < 3; i++ {
		err := shortener.DeleteUserShortURLsBatch(context.Background(), "user", []string{string(rune('a' + i))})
		assert.NoError(t, err)
	}

	// Закрываем и проверяем что нет паники
	shortener.Close()

	// Проверяем что все задачи были обработаны
	mockRepo.AssertNumberOfCalls(t, "DeleteBatch", 3)
}

func TestGenerateShortURLPart_StorageError(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	mockRepo := new(mocks.Repository)
	shortener := service.NewURLShortener(mockRepo, testLogger)

	storageError := errors.New("storage error")
	mockRepo.On("Save", mock.Anything, mock.Anything, mock.Anything).Return(storageError)

	userID := "test-user"
	url := "https://example.com"

	shortURL, err := shortener.GenerateShortURLPart(context.Background(), userID, url)
	assert.Error(t, err)
	assert.Equal(t, storageError, err)
	assert.Empty(t, shortURL)
}

func TestGenerateShortURLPartBatch_StorageError(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	mockRepo := new(mocks.Repository)
	shortener := service.NewURLShortener(mockRepo, testLogger)

	storageError := errors.New("storage error")
	mockRepo.On("SaveBatch", mock.Anything, mock.Anything, mock.Anything).Return(storageError)

	userID := "test-user"
	urls := []model.ShortenBatchRequestItem{
		*model.NewShortenBatchRequestItem("1", "https://example.com/1"),
	}

	result, err := shortener.GenerateShortURLPartBatch(context.Background(), userID, urls)
	assert.Error(t, err)
	assert.Equal(t, storageError, err)
	assert.Nil(t, result)
}
