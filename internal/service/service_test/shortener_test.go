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
