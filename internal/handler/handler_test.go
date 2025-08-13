package handler

import (
	"bytes"
	"context"
	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type MockShortener struct {
	mock.Mock
}

func (m *MockShortener) GenerateShortURLPart(ctx context.Context, url string) (string, error) {
	args := m.Called(ctx, url)
	return args.String(0), args.Error(1)
}

func (m *MockShortener) GetURLByShortURLPart(ctx context.Context, id string) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

func (m *MockShortener) PingRepository(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func testConfig() config.Config {
	return config.Config{
		ServerAddr:      "localhost:8080",
		BaseURL:         "http://localhost:8080",
		LogLevel:        "info",
		FileStoragePath: "./storage.json",
	}
}

func TestHandleGetShortURLRedirect(t *testing.T) {
	testCfg := testConfig()
	testLogger, _ := logger.NewLogger("debug")
	mockShortener := new(MockShortener)
	mockShortener.On("GetURLByShortURLPart", mock.Anything, "qwerty12").
		Return("https://practicum.yandex.ru/", nil)
	h := NewShortenerHandler(testCfg, testLogger, mockShortener)

	tests := []struct {
		name                string
		path                string
		expectedContentType string
		expectedCode        int
		expectedBody        string
		expectedLocation    string
	}{
		{
			name:                "Simple positive case",
			path:                "qwerty12",
			expectedCode:        http.StatusTemporaryRedirect,
			expectedContentType: "text/plain",
			expectedBody:        "",
			expectedLocation:    "https://practicum.yandex.ru/",
		},
		{
			name:                "Empty short url",
			path:                "",
			expectedCode:        http.StatusBadRequest,
			expectedContentType: "",
			expectedBody:        "short url is empty\n",
			expectedLocation:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+tt.path, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("shortURL", tt.path)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			h.HandleGetShortURLRedirect(rr, req)
			res := rr.Result()
			defer res.Body.Close()
			resBody, _ := io.ReadAll(res.Body)
			assert.Equal(t, tt.expectedCode, res.StatusCode, "Response code didn't match expected")
			if tt.expectedContentType != "" {
				contentType := res.Header.Get("Content-Type")
				assert.True(t, strings.HasPrefix(contentType, "text/plain"), "Content-Type didn't match expected")
			}
			assert.Equal(t, tt.expectedBody, string(resBody), "Body didn't match expected")
			if tt.expectedLocation != "" {
				location := res.Header.Get("Location")
				assert.Equal(t, tt.expectedLocation, location, "Location didn't match expected")
			}
		})
	}
}

func TestHandlePostShortURLTextPlain(t *testing.T) {
	testCfg := testConfig()
	testLogger, _ := logger.NewLogger("debug")
	mockShortener := new(MockShortener)
	mockShortener.On("GenerateShortURLPart", mock.Anything, "https://practicum.yandex.ru/").
		Return("qwerty12", nil)
	h := NewShortenerHandler(testCfg, testLogger, mockShortener)

	tests := []struct {
		name         string
		contentType  string
		body         string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Simple positive case",
			contentType:  "text/plain",
			body:         "https://practicum.yandex.ru/",
			expectedCode: http.StatusCreated,
			expectedBody: "http://localhost:8080/qwerty12",
		},
		{
			name:         "Incorrect OriginalURL",
			contentType:  "text/plain",
			body:         "incorrect_URL",
			expectedCode: http.StatusBadRequest,
			expectedBody: "incorrect url\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/qwerty12", bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rr := httptest.NewRecorder()

			h.HandlePostShortURLTextPlain(rr, req)
			res := rr.Result()
			defer res.Body.Close()
			resBody, _ := io.ReadAll(res.Body)
			assert.Equal(t, tt.expectedCode, res.StatusCode, "Response code didn't match expected")
			contentType := res.Header.Get("Content-Type")
			assert.True(t, strings.HasPrefix(contentType, "text/plain"), "Content-Type didn't match expected")
			assert.Equal(t, tt.expectedBody, string(resBody), "Body didn't match expected")
		})
	}
}

func TestHandlePostShortURLJSON(t *testing.T) {
	testCfg := testConfig()
	testLogger, _ := logger.NewLogger("debug")
	mockShortener := new(MockShortener)
	mockShortener.On("GenerateShortURLPart", mock.Anything, "https://practicum.yandex.ru/").
		Return("qwerty12", nil)
	h := NewShortenerHandler(testCfg, testLogger, mockShortener)

	tests := []struct {
		name         string
		contentType  string
		body         string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Simple positive case",
			contentType:  "application/json",
			body:         `{"url":"https://practicum.yandex.ru/"}`,
			expectedCode: http.StatusCreated,
			expectedBody: `{"result":"http://localhost:8080/qwerty12"}` + "\n",
		},
		{
			name:         "Incorrect OriginalURL",
			contentType:  "application/json",
			body:         `incorrect_JSON`,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"incorrect json"}` + "\n",
		},
		{
			name:         "Incorrect OriginalURL",
			contentType:  "application/json",
			body:         `{"url":"incorrect_URL"}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"incorrect url"}` + "\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/qwerty12", bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rr := httptest.NewRecorder()

			h.HandlePostShortURLJSON(rr, req)
			res := rr.Result()
			defer res.Body.Close()
			resBody, _ := io.ReadAll(res.Body)
			assert.Equal(t, tt.expectedCode, res.StatusCode, "Response code didn't match expected")
			contentType := res.Header.Get("Content-Type")
			assert.True(t, strings.HasPrefix(contentType, "application/json"), "Content-Type didn't match expected")
			assert.Equal(t, tt.expectedBody, string(resBody), "Body didn't match expected")
		})
	}
}

func TestHandlePingRepository(t *testing.T) {
	testCfg := testConfig()
	testLogger, _ := logger.NewLogger("debug")
	mockShortener := new(MockShortener)
	mockShortener.On("PingRepository", mock.Anything).Return(nil)
	h := NewShortenerHandler(testCfg, testLogger, mockShortener)

	tests := []struct {
		name         string
		expectedCode int
	}{
		{
			name:         "Simple positive case",
			expectedCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			rr := httptest.NewRecorder()

			h.HandlePingRepository(rr, req)
			res := rr.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.expectedCode, res.StatusCode, "Response code didn't match expected")
		})
	}
}
