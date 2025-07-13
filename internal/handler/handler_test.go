package handler

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
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

func TestHandlePostShortURL(t *testing.T) {
	mockShortener := new(MockShortener)
	mockShortener.On("GenerateShortURLPart", "https://practicum.yandex.ru/").
		Return("qwerty12", nil)
	h := NewShortenerHandler(mockShortener)

	tests := []struct {
		name         string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Simple positive case",
			expectedCode: http.StatusCreated,
			expectedBody: "http://localhost:8080/qwerty12",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/qwerty12",
				bytes.NewBufferString("https://practicum.yandex.ru/"))
			req.Header.Set("Content-Type", "text/plain")
			rr := httptest.NewRecorder()

			h.HandleMainPage()(rr, req)
			res := rr.Result()
			defer res.Body.Close()
			resBody, _ := io.ReadAll(res.Body)
			assert.Equal(t, tt.expectedCode, res.StatusCode, "Response code didn't match expected")
			contentType := res.Header.Get("Content-Type")
			assert.Equal(t, "text/plain", contentType, "Content-Type didn't match expected")
			assert.Equal(t, tt.expectedBody, string(resBody), "Location didn't match expected")
		})
	}
}
