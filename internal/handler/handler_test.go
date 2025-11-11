package handler

import (
	"bytes"
	"context"
	"errors"
	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/middleware"
	"github.com/bezjen/shortener/internal/mocks"
	"github.com/bezjen/shortener/internal/model"
	"github.com/bezjen/shortener/internal/repository"
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
	deletedURL := model.NewURL("qwerty13", "https://practicum.yandex1.ru/")
	deletedURL.IsDeleted = true
	mockShortener.On("GetURLByShortURLPart", mock.Anything, "qwerty13").Return(deletedURL, nil)
	mockAudit := new(mocks.AuditService)
	mockAudit.On("NotifyAll", mock.Anything).Return(nil)
	h := NewShortenerHandler(testCfg, testLogger, mockShortener, mockAudit)

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
	mockAudit := new(mocks.AuditService)
	mockAudit.On("NotifyAll", mock.Anything).Return(nil)
	h := NewShortenerHandler(testCfg, testLogger, mockShortener, mockAudit)

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
	mockAudit := new(mocks.AuditService)
	mockAudit.On("NotifyAll", mock.Anything).Return(nil)
	h := NewShortenerHandler(testCfg, testLogger, mockShortener, mockAudit)

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
	mockAudit := new(mocks.AuditService)
	mockAudit.On("NotifyAll", mock.Anything).Return(nil)
	h := NewShortenerHandler(testCfg, testLogger, mockShortener, mockAudit)

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

			mockAudit := new(mocks.AuditService)
			h := NewShortenerHandler(testCfg, testLogger, mockShortener, mockAudit)

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

	tests := []struct {
		name         string
		mockSetup    func(*mocks.Shortener)
		expectedCode int
	}{
		{
			name: "Successful ping",
			mockSetup: func(m *mocks.Shortener) {
				m.On("PingRepository", mock.Anything).Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Ping failure",
			mockSetup: func(m *mocks.Shortener) {
				m.On("PingRepository", mock.Anything).Return(errors.New("connection failed"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockShortener := new(mocks.Shortener)
			tt.mockSetup(mockShortener)
			mockAudit := new(mocks.AuditService)
			h := NewShortenerHandler(testCfg, testLogger, mockShortener, mockAudit)

			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			rr := httptest.NewRecorder()

			h.HandlePingRepository(rr, req)

			res := rr.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.expectedCode, res.StatusCode)
		})
	}
}

// BenchmarkHandlePostShortURLTextPlain измеряет производительность текстового POST
func BenchmarkHandlePostShortURLTextPlain(b *testing.B) {
	testCfg := testConfig()
	testLogger, _ := logger.NewLogger("info") // Используем info для уменьшения логов

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer() // Останавливаем таймер для настройки моков

		mockShortener := new(mocks.Shortener)
		mockAudit := new(mocks.AuditService)

		mockShortener.On("GenerateShortURLPart", mock.Anything, mock.Anything, "https://practicum.yandex.ru/").
			Return("qwerty12", nil)
		mockAudit.On("NotifyAll", mock.Anything).Return()

		h := NewShortenerHandler(testCfg, testLogger, mockShortener, mockAudit)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://practicum.yandex.ru/"))
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, "bench-user")
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		b.StartTimer() // Запускаем таймер для измерения
		h.HandlePostShortURLTextPlain(rr, req)
		b.StopTimer()
	}
}

// BenchmarkHandleGetShortURLRedirect измеряет производительность редиректа
func BenchmarkHandleGetShortURLRedirect(b *testing.B) {
	testCfg := testConfig()
	testLogger, _ := logger.NewLogger("info")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		mockShortener := new(mocks.Shortener)
		mockAudit := new(mocks.AuditService)

		url := model.NewURL("qwerty12", "https://practicum.yandex.ru/")
		mockShortener.On("GetURLByShortURLPart", mock.Anything, "qwerty12").Return(url, nil)
		mockAudit.On("NotifyAll", mock.Anything).Return()

		h := NewShortenerHandler(testCfg, testLogger, mockShortener, mockAudit)

		req := httptest.NewRequest(http.MethodGet, "/qwerty12", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("shortURL", "qwerty12")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rr := httptest.NewRecorder()

		b.StartTimer()
		h.HandleGetShortURLRedirect(rr, req)
		b.StopTimer()
	}
}

// BenchmarkHandlePostShortURLJSON измеряет производительность JSON POST
func BenchmarkHandlePostShortURLJSON(b *testing.B) {
	testCfg := testConfig()
	testLogger, _ := logger.NewLogger("info")

	jsonBody := `{"url":"https://practicum.yandex.ru/"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		mockShortener := new(mocks.Shortener)
		mockAudit := new(mocks.AuditService)

		mockShortener.On("GenerateShortURLPart", mock.Anything, mock.Anything, "https://practicum.yandex.ru/").
			Return("qwerty12", nil)
		mockAudit.On("NotifyAll", mock.Anything).Return()

		h := NewShortenerHandler(testCfg, testLogger, mockShortener, mockAudit)

		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, "bench-user")
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		b.StartTimer()
		h.HandlePostShortURLJSON(rr, req)
		b.StopTimer()
	}
}

// BenchmarkHandlePostShortURLBatchJSON измеряет производительность batch JSON POST
func BenchmarkHandlePostShortURLBatchJSON(b *testing.B) {
	testCfg := testConfig()
	testLogger, _ := logger.NewLogger("info")

	jsonBody := `[{"correlation_id":"123","original_url":"https://practicum.yandex.ru/"}]`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		mockShortener := new(mocks.Shortener)
		mockAudit := new(mocks.AuditService)

		mockShortener.On("GenerateShortURLPartBatch", mock.Anything, mock.Anything,
			[]model.ShortenBatchRequestItem{*model.NewShortenBatchRequestItem(
				"123", "https://practicum.yandex.ru/"),
			}).Return(
			[]model.ShortenBatchResponseItem{
				*model.NewShortenBatchResponseItem("123", "qwerty12"),
			}, nil)
		mockAudit.On("NotifyAll", mock.Anything).Return()

		h := NewShortenerHandler(testCfg, testLogger, mockShortener, mockAudit)

		req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBufferString(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, "bench-user")
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		b.StartTimer()
		h.HandlePostShortURLBatchJSON(rr, req)
		b.StopTimer()
	}
}

// BenchmarkHandleGetUserURLsJSON измеряет производительность получения пользовательских URL
func BenchmarkHandleGetUserURLsJSON(b *testing.B) {
	testCfg := testConfig()
	testLogger, _ := logger.NewLogger("info")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		mockShortener := new(mocks.Shortener)
		mockAudit := new(mocks.AuditService)

		mockShortener.On("GetURLsByUserID", mock.Anything, "user123").Return(
			[]model.URL{
				*model.NewURL("qwerty12", "https://example.com/page1"),
				*model.NewURL("qwerty34", "https://example.com/page2"),
			}, nil)

		h := NewShortenerHandler(testCfg, testLogger, mockShortener, mockAudit)

		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user123")
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		b.StartTimer()
		h.HandleGetUserURLsJSON(rr, req)
		b.StopTimer()
	}
}

func TestHandlePostShortURLTextPlain_URLConflict(t *testing.T) {
	testCfg := testConfig()
	testLogger, _ := logger.NewLogger("debug")
	mockShortener := new(mocks.Shortener)

	// Мокируем конфликт URL
	conflictErr := &repository.ErrURLConflict{
		ShortURL: "existing123",
	}
	mockShortener.On("GenerateShortURLPart", mock.Anything, mock.Anything, "https://conflict.example.com").
		Return("", conflictErr)

	mockAudit := new(mocks.AuditService)
	h := NewShortenerHandler(testCfg, testLogger, mockShortener, mockAudit)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://conflict.example.com"))
	rr := httptest.NewRecorder()

	h.HandlePostShortURLTextPlain(rr, req)

	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusConflict, res.StatusCode)
	assert.Equal(t, "http://localhost:8080/existing123", string(resBody))
	assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))
}

func TestHandlePostShortURLTextPlain_GenerationError(t *testing.T) {
	testCfg := testConfig()
	testLogger, _ := logger.NewLogger("debug")
	mockShortener := new(mocks.Shortener)

	// Мокируем общую ошибку генерации
	mockShortener.On("GenerateShortURLPart", mock.Anything, mock.Anything, "https://error.example.com").
		Return("", errors.New("database error"))

	mockAudit := new(mocks.AuditService)
	h := NewShortenerHandler(testCfg, testLogger, mockShortener, mockAudit)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://error.example.com"))
	rr := httptest.NewRecorder()

	h.HandlePostShortURLTextPlain(rr, req)

	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Equal(t, "Internal Server Error\n", string(resBody))
}

func TestHandlePostShortURLJSON_URLConflict(t *testing.T) {
	testCfg := testConfig()
	testLogger, _ := logger.NewLogger("debug")
	mockShortener := new(mocks.Shortener)

	conflictErr := &repository.ErrURLConflict{
		ShortURL: "existing456",
	}
	mockShortener.On("GenerateShortURLPart", mock.Anything, mock.Anything, "https://conflict.example.com").
		Return("", conflictErr)

	mockAudit := new(mocks.AuditService)
	h := NewShortenerHandler(testCfg, testLogger, mockShortener, mockAudit)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten",
		bytes.NewBufferString(`{"url":"https://conflict.example.com"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.HandlePostShortURLJSON(rr, req)

	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusConflict, res.StatusCode)
	assert.Equal(t, `{"result":"http://localhost:8080/existing456"}`+"\n", string(resBody))
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestHandlePostShortURLJSON_GenerationError(t *testing.T) {
	testCfg := testConfig()
	testLogger, _ := logger.NewLogger("debug")
	mockShortener := new(mocks.Shortener)

	mockShortener.On("GenerateShortURLPart", mock.Anything, mock.Anything, "https://error.example.com").
		Return("", errors.New("storage unavailable"))

	mockAudit := new(mocks.AuditService)
	h := NewShortenerHandler(testCfg, testLogger, mockShortener, mockAudit)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten",
		bytes.NewBufferString(`{"url":"https://error.example.com"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.HandlePostShortURLJSON(rr, req)

	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Equal(t, `{"error":"Internal Server Error"}`+"\n", string(resBody))
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestHandlePostShortURLTextPlain_URLConflict_BuildFullURLError(t *testing.T) {
	// Создаем конфиг с невалидным BaseURL для тестирования ошибки построения полного URL
	invalidCfg := config.Config{
		ServerAddr:      "localhost:8080",
		BaseURL:         "://invalid-url", // Невалидный URL
		LogLevel:        "info",
		FileStoragePath: "./storage.json",
	}

	testLogger, _ := logger.NewLogger("debug")
	mockShortener := new(mocks.Shortener)

	// Мокируем конфликт URL
	conflictErr := &repository.ErrURLConflict{
		ShortURL: "existing123",
	}
	mockShortener.On("GenerateShortURLPart", mock.Anything, mock.Anything, "https://conflict.example.com").
		Return("", conflictErr)

	mockAudit := new(mocks.AuditService)
	h := NewShortenerHandler(invalidCfg, testLogger, mockShortener, mockAudit)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("https://conflict.example.com"))
	rr := httptest.NewRecorder()

	h.HandlePostShortURLTextPlain(rr, req)

	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	// Должны получить 500 из-за ошибки построения полного URL
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Equal(t, "Internal Server Error\n", string(resBody))
}

func TestHandlePostShortURLJSON_URLConflict_BuildFullURLError(t *testing.T) {
	// Создаем конфиг с невалидным BaseURL
	invalidCfg := config.Config{
		ServerAddr:      "localhost:8080",
		BaseURL:         "://invalid-url", // Невалидный URL
		LogLevel:        "info",
		FileStoragePath: "./storage.json",
	}

	testLogger, _ := logger.NewLogger("debug")
	mockShortener := new(mocks.Shortener)

	// Мокируем конфликт URL
	conflictErr := &repository.ErrURLConflict{
		ShortURL: "existing456",
	}
	mockShortener.On("GenerateShortURLPart", mock.Anything, mock.Anything, "https://conflict.example.com").
		Return("", conflictErr)

	mockAudit := new(mocks.AuditService)
	h := NewShortenerHandler(invalidCfg, testLogger, mockShortener, mockAudit)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten",
		bytes.NewBufferString(`{"url":"https://conflict.example.com"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.HandlePostShortURLJSON(rr, req)

	res := rr.Result()
	defer res.Body.Close()
	resBody, _ := io.ReadAll(res.Body)

	// Должны получить 500 из-за ошибки построения полного URL
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Equal(t, `{"error":"Internal Server Error"}`+"\n", string(resBody))
}

func TestHandleDeleteShortURLsBatchJSON(t *testing.T) {
	testCfg := testConfig()
	testLogger, _ := logger.NewLogger("debug")

	tests := []struct {
		name         string
		body         string
		userID       string
		mockSetup    func(*mocks.Shortener)
		expectedCode int
		expectedBody string
	}{
		{
			name:   "Successful deletion request",
			body:   `["abc123", "def456"]`,
			userID: "user123",
			mockSetup: func(m *mocks.Shortener) {
				m.On("DeleteUserShortURLsBatch", mock.Anything, "user123", []string{"abc123", "def456"}).
					Return(nil)
			},
			expectedCode: http.StatusAccepted,
			expectedBody: "",
		},
		{
			name:         "Invalid JSON",
			body:         `invalid json`,
			userID:       "user123",
			mockSetup:    func(m *mocks.Shortener) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"incorrect json"}` + "\n",
		},
		{
			name:   "Deletion queue full",
			body:   `["abc123", "def456"]`,
			userID: "user123",
			mockSetup: func(m *mocks.Shortener) {
				m.On("DeleteUserShortURLsBatch", mock.Anything, "user123", []string{"abc123", "def456"}).
					Return(errors.New("queue full"))
			},
			expectedCode: http.StatusTooManyRequests,
			expectedBody: `{"error":"Too Many Requests"}` + "\n",
		},
		{
			name:         "Empty request body",
			body:         "",
			userID:       "user123",
			mockSetup:    func(m *mocks.Shortener) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"incorrect json"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockShortener := new(mocks.Shortener)
			tt.mockSetup(mockShortener)
			mockAudit := new(mocks.AuditService)
			h := NewShortenerHandler(testCfg, testLogger, mockShortener, mockAudit)

			req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			h.HandleDeleteShortURLsBatchJSON(rr, req)

			res := rr.Result()
			defer res.Body.Close()
			resBody, _ := io.ReadAll(res.Body)

			assert.Equal(t, tt.expectedCode, res.StatusCode)
			assert.Equal(t, tt.expectedBody, string(resBody))
		})
	}
}

func TestReadBody(t *testing.T) {
	testCfg := testConfig()
	testLogger, _ := logger.NewLogger("debug")
	h := NewShortenerHandler(testCfg, testLogger, nil, nil)

	tests := []struct {
		name        string
		body        string
		want        string
		expectError bool
	}{
		{
			name:        "Valid body",
			body:        "test content",
			want:        "test content",
			expectError: false,
		},
		{
			name:        "Empty body",
			body:        "",
			want:        "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			result, err := h.readBody(req)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	testCfg := testConfig()
	testLogger, _ := logger.NewLogger("debug")
	h := NewShortenerHandler(testCfg, testLogger, nil, nil)

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "Valid HTTP URL",
			url:     "http://example.com",
			wantErr: false,
		},
		{
			name:    "Valid HTTPS URL",
			url:     "https://example.com",
			wantErr: false,
		},
		{
			name:    "Invalid URL",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "Empty URL",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.validateURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildFullURL(t *testing.T) {
	testCfg := testConfig()
	testLogger, _ := logger.NewLogger("debug")
	h := NewShortenerHandler(testCfg, testLogger, nil, nil)

	tests := []struct {
		name        string
		baseURL     string
		shortURL    string
		want        string
		expectError bool
	}{
		{
			name:        "Valid URL combination",
			baseURL:     "http://localhost:8080",
			shortURL:    "abc123",
			want:        "http://localhost:8080/abc123",
			expectError: false,
		},
		{
			name:        "Empty short URL",
			baseURL:     "http://localhost:8080",
			shortURL:    "",
			want:        "http://localhost:8080",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h.cfg.BaseURL = tt.baseURL
			result, err := h.buildFullURL(tt.shortURL)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}
