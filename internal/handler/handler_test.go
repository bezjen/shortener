package handler

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockShortener struct {
	mock.Mock
}

func (m *MockShortener) GenerateShortURLPart(url string) (string, error) {
	args := m.Called(url)
	return args.String(0), args.Error(1)
}

func (m *MockShortener) GetURLByShortURLPart(id string) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func TestHandleGetShortURL(t *testing.T) {
	mockShortener := new(MockShortener)
	mockShortener.On("GetURLByShortURLPart", "qwerty12").
		Return("https://practicum.yandex.ru/", nil)
	h := NewShortenerHandler(mockShortener)

	tests := []struct {
		name             string
		path             string
		expectedCode     int
		expectedLocation string
	}{
		{
			name:             "Simple positive case",
			path:             "/qwerty12",
			expectedCode:     http.StatusTemporaryRedirect,
			expectedLocation: "https://practicum.yandex.ru/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/qwerty12", nil)
			rr := httptest.NewRecorder()

			h.HandleMainPage()(rr, req)
			res := rr.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.expectedCode, res.StatusCode, "Response code didn't match expected")
			contentType := res.Header.Get("Content-Type")
			assert.Equal(t, "text/plain", contentType, "Content-Type didn't match expected")
			location := res.Header.Get("Location")
			assert.Equal(t, tt.expectedLocation, location, "Location didn't match expected")
		})
	}
}
