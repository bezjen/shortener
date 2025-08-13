package service

import (
	"errors"
	"github.com/bezjen/shortener/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Save(id string, url string) error {
	args := m.Called(id, url)
	return args.Error(0)
}

func (m *MockRepository) GetByShortURL(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func (m *MockRepository) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestGenerateShortURLPart(t *testing.T) {
	mockPositiveRepo := new(MockRepository)
	mockPositiveRepo.On("Save", mock.Anything, "https://practicum.yandex.ru/").
		Return(nil)
	mockCollisionRepo := new(MockRepository)
	mockCollisionRepo.On("Save", mock.Anything, "https://practicum.yandex.ru/").
		Return(repository.ErrConflict)
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
			want:    "result12",
			wantErr: nil,
		},
		{
			name:    "Too many collisions case",
			storage: mockCollisionRepo,
			url:     "https://practicum.yandex.ru/",
			want:    "",
			wantErr: ErrGenerate,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &URLShortener{
				storage: tt.storage,
			}
			shortURL, err := u.GenerateShortURLPart(tt.url)
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

func TestGetURLByShortURLPart(t *testing.T) {
	mockRepoPositive := new(MockRepository)
	mockRepoPositive.On("GetByShortURL", "qwerty12").
		Return("https://practicum.yandex.ru/", nil)
	mockRepoNotFound := new(MockRepository)
	mockRepoNotFound.On("GetByShortURL", "qwerty12").
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
			got, err := u.GetURLByShortURLPart(tt.shortURLPart)
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
