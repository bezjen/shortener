package grpc

import (
	"context"
	"errors"
	"github.com/bezjen/shortener/internal/mocks"
	"testing"

	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/model"
	"github.com/bezjen/shortener/internal/repository"
	pb "github.com/bezjen/shortener/pkg/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// setupTest подготавливает окружение для тестов
func setupTest() (*Handler, *mocks.Authorizer, *mocks.Shortener, *mocks.AuditService) {
	mockAuthorizer := new(mocks.Authorizer)
	mockShortener := new(mocks.Shortener)
	mockAudit := new(mocks.AuditService)
	cfg := config.Config{BaseURL: "http://localhost:8080"}
	log, _ := logger.NewLogger("info")

	handler := NewHandler(cfg, log, mockAuthorizer, mockShortener, mockAudit)
	return handler, mockAuthorizer, mockShortener, mockAudit
}

// createAuthContext создает контекст с токеном
func createAuthContext(token string) context.Context {
	md := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs("authorization", "Bearer "+token),
	)
	return md
}

func TestGRPCHandler_ShortenURL(t *testing.T) {
	handler, auth, shortener, audit := setupTest()

	tests := []struct {
		name        string
		req         *pb.URLShortenRequest
		token       string
		setupMocks  func()
		expectError bool
		expectCode  codes.Code
		expectResp  string
	}{
		{
			name:  "Success",
			req:   &pb.URLShortenRequest{Url: "http://example.com"},
			token: "valid_token",
			setupMocks: func() {
				auth.On("ValidateToken", "valid_token").Return("user1", nil)
				shortener.On("GenerateShortURLPart", mock.Anything, "user1", "http://example.com").Return("short123", nil)
				audit.On("NotifyAll", mock.MatchedBy(func(e model.AuditEvent) bool {
					return e.Action == model.ActionShorten && e.UserID == "user1"
				})).Return()
			},
			expectError: false,
			expectResp:  "http://localhost:8080/short123",
		},
		{
			name:        "Unauthenticated - No Token",
			req:         &pb.URLShortenRequest{Url: "http://example.com"},
			token:       "",
			setupMocks:  func() {},
			expectError: true,
			expectCode:  codes.Unauthenticated,
		},
		{
			name:  "Invalid Argument - Bad URL",
			req:   &pb.URLShortenRequest{Url: "ftp://bad-scheme.com"},
			token: "valid_token",
			setupMocks: func() {
				auth.On("ValidateToken", "valid_token").Return("user1", nil)
			},
			expectError: true,
			expectCode:  codes.InvalidArgument,
		},
		{
			name:  "Conflict - URL Already Exists",
			req:   &pb.URLShortenRequest{Url: "http://exists.com"},
			token: "valid_token",
			setupMocks: func() {
				auth.On("ValidateToken", "valid_token").Return("user1", nil)
				// Эмулируем ошибку конфликта
				conflictErr := &repository.ErrURLConflict{ShortURL: "existing123"}
				shortener.On("GenerateShortURLPart", mock.Anything, "user1", "http://exists.com").Return("", conflictErr)
			},
			expectError: false,
			expectResp:  "http://localhost:8080/existing123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			var ctx context.Context
			if tt.token != "" {
				ctx = createAuthContext(tt.token)
			} else {
				ctx = context.Background()
			}

			resp, err := handler.ShortenURL(ctx, tt.req)

			if tt.expectError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectResp, resp.Result)
			}
		})
	}
}

func TestGRPCHandler_ExpandURL(t *testing.T) {
	handler, _, shortener, audit := setupTest()

	tests := []struct {
		name        string
		req         *pb.URLExpandRequest
		setupMocks  func()
		expectError bool
		expectCode  codes.Code
		expectResp  string
	}{
		{
			name: "Success",
			req:  &pb.URLExpandRequest{Id: "short123"},
			setupMocks: func() {
				urlModel := &model.URL{OriginalURL: "http://original.com", IsDeleted: false}
				shortener.On("GetURLByShortURLPart", mock.Anything, "short123").Return(urlModel, nil)
				audit.On("NotifyAll", mock.MatchedBy(func(e model.AuditEvent) bool {
					return e.Action == model.ActionFollow
				})).Return()
			},
			expectError: false,
			expectResp:  "http://original.com",
		},
		{
			name: "Not Found",
			req:  &pb.URLExpandRequest{Id: "unknown"},
			setupMocks: func() {
				shortener.On("GetURLByShortURLPart", mock.Anything, "unknown").Return(nil, errors.New("not found"))
			},
			expectError: true,
			expectCode:  codes.NotFound,
		},
		{
			name: "Deleted URL",
			req:  &pb.URLExpandRequest{Id: "deleted"},
			setupMocks: func() {
				urlModel := &model.URL{OriginalURL: "http://deleted.com", IsDeleted: true}
				shortener.On("GetURLByShortURLPart", mock.Anything, "deleted").Return(urlModel, nil)
			},
			expectError: true,
			expectCode:  codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			// Аутентификация для ExpandURL опциональна, но метод пытается извлечь ID
			// Передадим пустой контекст или с невалидной метадатой - это не должно ломать логику
			resp, err := handler.ExpandURL(context.Background(), tt.req)

			if tt.expectError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectResp, resp.Result)
			}
		})
	}
}

func TestGRPCHandler_ListUserURLs(t *testing.T) {
	handler, auth, shortener, _ := setupTest()

	tests := []struct {
		name        string
		token       string
		setupMocks  func()
		expectError bool
		expectCode  codes.Code
		expectLen   int
	}{
		{
			name:  "Success",
			token: "valid_token",
			setupMocks: func() {
				auth.On("ValidateToken", "valid_token").Return("user1", nil)
				urls := []model.URL{
					{ShortURL: "s1", OriginalURL: "http://o1.com"},
					{ShortURL: "s2", OriginalURL: "http://o2.com"},
				}
				shortener.On("GetURLsByUserID", mock.Anything, "user1").Return(urls, nil)
			},
			expectError: false,
			expectLen:   2,
		},
		{
			name:  "Unauthenticated",
			token: "invalid",
			setupMocks: func() {
				auth.On("ValidateToken", "invalid").Return("", errors.New("invalid token"))
			},
			expectError: true,
			expectCode:  codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			ctx := createAuthContext(tt.token)
			resp, err := handler.ListUserURLs(ctx, &pb.Empty{})

			if tt.expectError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.Len(t, resp.Url, tt.expectLen)
			}
		})
	}
}
