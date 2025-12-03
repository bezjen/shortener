package router

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/handler"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/mocks"
	"github.com/bezjen/shortener/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewRouter(t *testing.T) {
	testLogger, _ := logger.NewLogger("error") // Используем error level чтобы уменьшить логи

	tests := []struct {
		name         string
		method       string
		path         string
		body         []byte
		setupMocks   func(*mocks.Authorizer, *mocks.Shortener, *mocks.AuditService)
		setupRequest func(*http.Request)
		expectedCode int
	}{
		{
			name:   "POST / with valid URL",
			method: "POST",
			path:   "/",
			body:   []byte("https://example.com"),
			setupMocks: func(a *mocks.Authorizer, s *mocks.Shortener, audit *mocks.AuditService) {
				// Настраиваем создание токена для нового пользователя
				a.On("CreateToken", mock.AnythingOfType("string")).Return("test-token", nil)
				s.On("GenerateShortURLPart", mock.Anything, mock.Anything, "https://example.com").Return("abc123", nil)
				audit.On("NotifyAll", mock.Anything).Return()
			},
			expectedCode: 201,
		},
		{
			name:   "POST / with invalid URL",
			method: "POST",
			path:   "/",
			body:   []byte("invalid-url"),
			setupMocks: func(a *mocks.Authorizer, s *mocks.Shortener, audit *mocks.AuditService) {
				a.On("CreateToken", mock.AnythingOfType("string")).Return("test-token", nil)
				// validateURL вернет ошибку, поэтому GenerateShortURLPart не будет вызван
			},
			expectedCode: 400,
		},
		{
			name:   "GET /ping",
			method: "GET",
			path:   "/ping",
			body:   nil,
			setupMocks: func(a *mocks.Authorizer, s *mocks.Shortener, audit *mocks.AuditService) {
				a.On("CreateToken", mock.AnythingOfType("string")).Return("test-token", nil)
				s.On("PingRepository", mock.Anything).Return(nil)
			},
			expectedCode: 200,
		},
		{
			name:   "GET /ping with error",
			method: "GET",
			path:   "/ping",
			body:   nil,
			setupMocks: func(a *mocks.Authorizer, s *mocks.Shortener, audit *mocks.AuditService) {
				a.On("CreateToken", mock.AnythingOfType("string")).Return("test-token", nil)
				s.On("PingRepository", mock.Anything).Return(errors.New("db error"))
			},
			expectedCode: 500,
		},
		{
			name:   "GET existing short URL",
			method: "GET",
			path:   "/abc123",
			body:   nil,
			setupMocks: func(a *mocks.Authorizer, s *mocks.Shortener, audit *mocks.AuditService) {
				a.On("CreateToken", mock.AnythingOfType("string")).Return("test-token", nil)
				url := model.NewURL("abc123", "https://example.com")
				s.On("GetURLByShortURLPart", mock.Anything, "abc123").Return(url, nil)
				audit.On("NotifyAll", mock.Anything).Return()
			},
			expectedCode: 307,
		},
		{
			name:   "GET non-existent short URL",
			method: "GET",
			path:   "/nonexistent",
			body:   nil,
			setupMocks: func(a *mocks.Authorizer, s *mocks.Shortener, audit *mocks.AuditService) {
				a.On("CreateToken", mock.AnythingOfType("string")).Return("test-token", nil)
				s.On("GetURLByShortURLPart", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))
			},
			expectedCode: 500,
		},
		{
			name:   "GET deleted short URL",
			method: "GET",
			path:   "/deleted123",
			body:   nil,
			setupMocks: func(a *mocks.Authorizer, s *mocks.Shortener, audit *mocks.AuditService) {
				a.On("CreateToken", mock.AnythingOfType("string")).Return("test-token", nil)
				url := model.NewURL("deleted123", "https://example.com")
				url.IsDeleted = true
				s.On("GetURLByShortURLPart", mock.Anything, "deleted123").Return(url, nil)
			},
			expectedCode: 410,
		},
		{
			name:   "POST /api/shorten with valid JSON",
			method: "POST",
			path:   "/api/shorten",
			body:   []byte(`{"url":"https://example.com"}`),
			setupMocks: func(a *mocks.Authorizer, s *mocks.Shortener, audit *mocks.AuditService) {
				a.On("CreateToken", mock.AnythingOfType("string")).Return("test-token", nil)
				s.On("GenerateShortURLPart", mock.Anything, mock.Anything, "https://example.com").Return("abc123", nil)
				audit.On("NotifyAll", mock.Anything).Return()
			},
			expectedCode: 201,
		},
		{
			name:   "POST /api/shorten with invalid JSON",
			method: "POST",
			path:   "/api/shorten",
			body:   []byte(`invalid json`),
			setupMocks: func(a *mocks.Authorizer, s *mocks.Shortener, audit *mocks.AuditService) {
				a.On("CreateToken", mock.AnythingOfType("string")).Return("test-token", nil)
			},
			expectedCode: 400,
		},
		{
			name:   "POST /api/shorten with invalid URL in JSON",
			method: "POST",
			path:   "/api/shorten",
			body:   []byte(`{"url":"invalid-url"}`),
			setupMocks: func(a *mocks.Authorizer, s *mocks.Shortener, audit *mocks.AuditService) {
				a.On("CreateToken", mock.AnythingOfType("string")).Return("test-token", nil)
			},
			expectedCode: 400,
		},
		{
			name:   "POST /api/shorten/batch with valid JSON",
			method: "POST",
			path:   "/api/shorten/batch",
			body:   []byte(`[{"correlation_id":"1","original_url":"https://example.com"}]`),
			setupMocks: func(a *mocks.Authorizer, s *mocks.Shortener, audit *mocks.AuditService) {
				a.On("CreateToken", mock.AnythingOfType("string")).Return("test-token", nil)
				requestItems := []model.ShortenBatchRequestItem{
					*model.NewShortenBatchRequestItem("1", "https://example.com"),
				}
				responseItems := []model.ShortenBatchResponseItem{
					*model.NewShortenBatchResponseItem("1", "abc123"),
				}
				s.On("GenerateShortURLPartBatch", mock.Anything, mock.Anything, requestItems).Return(responseItems, nil)
			},
			expectedCode: 201,
		},
		{
			name:   "GET /api/user/urls with no URLs",
			method: "GET",
			path:   "/api/user/urls",
			body:   nil,
			setupMocks: func(a *mocks.Authorizer, s *mocks.Shortener, audit *mocks.AuditService) {
				a.On("CreateToken", mock.AnythingOfType("string")).Return("test-token", nil)
				s.On("GetURLsByUserID", mock.Anything, mock.Anything).Return([]model.URL{}, nil)
			},
			expectedCode: 204,
		},
		{
			name:   "DELETE /api/user/urls with valid JSON",
			method: "DELETE",
			path:   "/api/user/urls",
			body:   []byte(`["abc123","def456"]`),
			setupMocks: func(a *mocks.Authorizer, s *mocks.Shortener, audit *mocks.AuditService) {
				a.On("CreateToken", mock.AnythingOfType("string")).Return("test-token", nil)
				s.On("DeleteUserShortURLsBatch", mock.Anything, mock.Anything, []string{"abc123", "def456"}).Return(nil)
			},
			expectedCode: 202,
		},
		{
			name:   "DELETE /api/user/urls with invalid JSON",
			method: "DELETE",
			path:   "/api/user/urls",
			body:   []byte(`invalid json`),
			setupMocks: func(a *mocks.Authorizer, s *mocks.Shortener, audit *mocks.AuditService) {
				a.On("CreateToken", mock.AnythingOfType("string")).Return("test-token", nil)
			},
			expectedCode: 400,
		},
		{
			name:   "DELETE /api/user/urls with queue full",
			method: "DELETE",
			path:   "/api/user/urls",
			body:   []byte(`["abc123","def456"]`),
			setupMocks: func(a *mocks.Authorizer, s *mocks.Shortener, audit *mocks.AuditService) {
				a.On("CreateToken", mock.AnythingOfType("string")).Return("test-token", nil)
				s.On("DeleteUserShortURLsBatch", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("queue full"))
			},
			expectedCode: 429,
		},
		{
			name:   "GET /api/internal/stats with trusted IP",
			method: "GET",
			path:   "/api/internal/stats",
			body:   nil,
			setupMocks: func(a *mocks.Authorizer, s *mocks.Shortener, audit *mocks.AuditService) {
				a.On("CreateToken", mock.AnythingOfType("string")).Return("test-token", nil)
				s.On("GetStats", mock.Anything).Return(150, 25, nil)
			},
			setupRequest: func(req *http.Request) {
				req.Header.Set("X-Real-IP", "192.168.1.100")
			},
			expectedCode: 200,
		},
		{
			name:   "GET /api/internal/stats without trusted subnet",
			method: "GET",
			path:   "/api/internal/stats",
			body:   nil,
			setupMocks: func(a *mocks.Authorizer, s *mocks.Shortener, audit *mocks.AuditService) {
				a.On("CreateToken", mock.AnythingOfType("string")).Return("test-token", nil)
			},
			expectedCode: 403,
		},
		{
			name:   "GET /debug/pprof/",
			method: "GET",
			path:   "/debug/pprof/",
			body:   nil,
			setupMocks: func(a *mocks.Authorizer, s *mocks.Shortener, audit *mocks.AuditService) {
				a.On("CreateToken", mock.AnythingOfType("string")).Return("test-token", nil)
			},
			expectedCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthorizer := new(mocks.Authorizer)
			mockShortener := new(mocks.Shortener)
			mockAudit := new(mocks.AuditService)

			// Настраиваем моки согласно тестовому сценарию
			tt.setupMocks(mockAuthorizer, mockShortener, mockAudit)

			shortenerHandler := handler.NewShortenerHandler(
				config.Config{BaseURL: "http://test.com", TrustedSubnet: "192.168.1.0/24"},
				testLogger,
				mockShortener,
				mockAudit,
			)

			router := NewRouter(testLogger, mockAuthorizer, *shortenerHandler)

			// Создаем запрос
			var req *http.Request
			if tt.body != nil {
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewBuffer(tt.body))
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			// Применяем setupRequest если есть
			if tt.setupRequest != nil {
				tt.setupRequest(req)
			}

			// Для теста с аутентифицированным пользователем добавляем cookie
			if tt.name == "GET /api/user/urls with authenticated user" {
				req.AddCookie(&http.Cookie{
					Name:  "user_token",
					Value: "valid-token",
				})
			}

			// Устанавливаем Content-Type для POST/PUT/DELETE запросов
			if tt.method == "POST" || tt.method == "PUT" || tt.method == "DELETE" {
				if tt.path == "/" {
					req.Header.Set("Content-Type", "text/plain")
				} else {
					req.Header.Set("Content-Type", "application/json")
				}
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code,
				"Route %s %s returned wrong status code: got %d, want %d",
				tt.method, tt.path, rr.Code, tt.expectedCode)

			// Проверяем, что все ожидаемые вызовы моков были выполнены
			mockAuthorizer.AssertExpectations(t)
			mockShortener.AssertExpectations(t)
			mockAudit.AssertExpectations(t)
		})
	}
}
