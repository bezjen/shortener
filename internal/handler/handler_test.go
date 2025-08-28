package handler

import (
	"bytes"
	"context"
	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/middleware"
	"github.com/bezjen/shortener/internal/mocks"
	"github.com/bezjen/shortener/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
	mockShortener := new(mocks.Shortener)
	mockShortener.On("GetURLByShortURLPart", mock.Anything, "qwerty12").
		Return(model.NewURL("qwerty12", "https://practicum.yandex.ru/"), nil)
	deletedUrl := model.NewURL("qwerty13", "https://practicum.yandex1.ru/")
	deletedUrl.IsDeleted = true
	mockShortener.On("GetURLByShortURLPart", mock.Anything, "qwerty13").Return(deletedUrl, nil)
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
		{
			name:                "Deleted url",
			path:                "qwerty13",
			expectedCode:        http.StatusGone,
			expectedContentType: "",
			expectedBody:        "",
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
	mockShortener := new(mocks.Shortener)
	mockShortener.On("GenerateShortURLPart", mock.Anything, mock.Anything, "https://practicum.yandex.ru/").
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
	mockShortener := new(mocks.Shortener)
	mockShortener.On("GenerateShortURLPart", mock.Anything, mock.Anything, "https://practicum.yandex.ru/").
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

func TestHandlePostShortURLBatchJSON(t *testing.T) {
	testCfg := testConfig()
	testLogger, _ := logger.NewLogger("debug")
	mockShortener := new(mocks.Shortener)
	mockShortener.On("GenerateShortURLPartBatch", mock.Anything, mock.Anything,
		[]model.ShortenBatchRequestItem{*model.NewShortenBatchRequestItem(
			"123", "https://practicum.yandex.ru/"),
		}).Return(
		[]model.ShortenBatchResponseItem{
			*model.NewShortenBatchResponseItem("123", "qwerty12"),
		}, nil)
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
			body:         `[{"correlation_id":"123","original_url":"https://practicum.yandex.ru/"}]`,
			expectedCode: http.StatusCreated,
			expectedBody: `[{"correlation_id":"123","short_url":"http://localhost:8080/qwerty12"}]` + "\n",
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
			body:         `[{"correlation_id":"123","original_url":"incorrect_URL"}]`,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"incorrect url incorrect_URL"}` + "\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/qwerty12", bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rr := httptest.NewRecorder()

			h.HandlePostShortURLBatchJSON(rr, req)
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

func TestHandleGetUserURLsJSON(t *testing.T) {
	testCfg := testConfig()
	testLogger, _ := logger.NewLogger("debug")

	tests := []struct {
		name         string
		mockSetup    func(*mocks.Shortener)
		setupRequest func(*http.Request)
		expectedCode int
		expectedBody string
		expectJSON   bool
	}{
		{
			name: "Simple positive case",
			mockSetup: func(m *mocks.Shortener) {
				m.On("GetURLsByUserID", mock.Anything, "user123").Return(
					[]model.URL{
						*model.NewURL("qwerty12", "https://example.com/page1"),
						*model.NewURL("qwerty34", "https://example.com/page2"),
					}, nil)
			},
			setupRequest: func(req *http.Request) {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user123")
				*req = *req.WithContext(ctx)
			},
			expectedCode: http.StatusOK,
			expectedBody: `[{"short_url":"http://localhost:8080/qwerty12","original_url":"https://example.com/page1"},{"short_url":"http://localhost:8080/qwerty34","original_url":"https://example.com/page2"}]` + "\n",
			expectJSON:   true,
		},
		{
			name: "No URLs for user - 204 No Content",
			mockSetup: func(m *mocks.Shortener) {
				m.On("GetURLsByUserID", mock.Anything, "user123").Return([]model.URL{}, nil)
			},
			setupRequest: func(req *http.Request) {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user123")
				*req = *req.WithContext(ctx)
			},
			expectedCode: http.StatusNoContent,
			expectedBody: "",
			expectJSON:   false,
		},
		{
			name: "Unauthorized - no cookie",
			mockSetup: func(m *mocks.Shortener) {
			},
			setupRequest: func(req *http.Request) {
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"error":"Unauthorized"}` + "\n",
			expectJSON:   true,
		},
		{
			name: "Unauthorized - invalid cookie",
			mockSetup: func(m *mocks.Shortener) {
			},
			setupRequest: func(req *http.Request) {
				req.AddCookie(&http.Cookie{Name: "user_token", Value: ""})
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"error":"Unauthorized"}` + "\n",
			expectJSON:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockShortener := new(mocks.Shortener)
			tt.mockSetup(mockShortener)

			h := NewShortenerHandler(testCfg, testLogger, mockShortener)

			req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			if tt.setupRequest != nil {
				tt.setupRequest(req)
			}

			rr := httptest.NewRecorder()

			h.HandleGetUserURLsJSON(rr, req)

			res := rr.Result()
			defer res.Body.Close()

			resBody, _ := io.ReadAll(res.Body)

			assert.Equal(t, tt.expectedCode, res.StatusCode, "Response code didn't match expected")

			if tt.expectJSON {
				contentType := res.Header.Get("Content-Type")
				assert.True(t, strings.HasPrefix(contentType, "application/json"),
					"Content-Type didn't match expected, got: %s", contentType)
			}

			assert.Equal(t, tt.expectedBody, string(resBody), "Body didn't match expected")
		})
	}
}

func TestHandlePingRepository(t *testing.T) {
	testCfg := testConfig()
	testLogger, _ := logger.NewLogger("debug")
	mockShortener := new(mocks.Shortener)
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
