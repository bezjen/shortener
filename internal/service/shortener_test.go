package service

import (
	"context"
	"errors"
	"github.com/bezjen/shortener/internal/model"
	"github.com/bezjen/shortener/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Save(ctx context.Context, url model.URL) error {
	args := m.Called(ctx, url)
	return args.Error(0)
}

func (m *MockRepository) SaveBatch(ctx context.Context, url []model.URL) error {
	args := m.Called(ctx, url)
	return args.Error(0)
}

func (m *MockRepository) GetByShortURL(ctx context.Context, id string) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

func (m *MockRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestGenerateShortURLPart(t *testing.T) {
	mockPositiveRepo := new(MockRepository)
	mockPositiveRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
	mockCollisionRepo := new(MockRepository)
	mockCollisionRepo.On("Save", mock.Anything, mock.Anything).Return(repository.ErrShortURLConflict)
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
			wantErr: ErrGenerate,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &URLShortener{
				storage: tt.storage,
			}
			shortURL, err := u.GenerateShortURLPart(context.TODO(), tt.url)
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
	mockPositiveRepo := new(MockRepository)
	mockPositiveRepo.On("SaveBatch", mock.Anything, mock.Anything).Return(nil)
	mockCollisionRepo := new(MockRepository)
	mockCollisionRepo.On("SaveBatch", mock.Anything, mock.Anything).Return(repository.ErrShortURLConflict)
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
			wantErr: ErrGenerate,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &URLShortener{
				storage: tt.storage,
			}
			shortURL, err := u.GenerateShortURLPartBatch(context.TODO(), tt.urls)
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
	mockRepoPositive := new(MockRepository)
	mockRepoPositive.On("GetByShortURL", mock.Anything, "qwerty12").
		Return("https://practicum.yandex.ru/", nil)
	mockRepoNotFound := new(MockRepository)
	mockRepoNotFound.On("GetByShortURL", mock.Anything, "qwerty12").
		Return("", repository.ErrNotFound)

	tests := []struct {
		name         string
		storage      repository.Repository
		shortURLPart string
		want         string
		wantErr      error
	}{
		{
			name:         "Simple positive case",
			storage:      mockRepoPositive,
			shortURLPart: "qwerty12",
			want:         "https://practicum.yandex.ru/",
			wantErr:      nil,
		},
		{
			name:         "Url not found case",
			storage:      mockRepoNotFound,
			shortURLPart: "qwerty12",
			want:         "",
			wantErr:      repository.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &URLShortener{
				storage: tt.storage,
			}
			got, err := u.GetURLByShortURLPart(context.TODO(), tt.shortURLPart)
			if err != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GenerateShortURLPart() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if got != tt.want {
				t.Errorf("GetURLByShortURLPart() got = %v, want %v", got, tt.want)
			}
		})
	}
}
