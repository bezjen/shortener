package middleware

import (
	"errors"
	"github.com/bezjen/shortener/internal/logger"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockAuthorizer struct {
	validTokens   map[string]string
	shouldFail    bool
	createFail    bool
	validateError bool
}

func (m *mockAuthorizer) CreateToken(userID string) (string, error) {
	if m.createFail {
		return "", errors.New("token creation failed")
	}
	return "token_" + userID, nil
}

func (m *mockAuthorizer) ValidateToken(token string) (string, error) {
	if m.validateError {
		return "", errors.New("validation error")
	}
	if userID, exists := m.validTokens[token]; exists {
		return userID, nil
	}
	return "", errors.New("invalid token")
}

func TestNewAuthMiddleware(t *testing.T) {
	logger, _ := logger.NewLogger("debug")
	authorizer := &mockAuthorizer{}
	middleware := NewAuthMiddleware(authorizer, logger)

	if middleware == nil || middleware.authorizer != authorizer {
		t.Error("Expected authorizer to be set")
	}
}

func TestWithAuth_NoCredentials(t *testing.T) {
	logger, _ := logger.NewLogger("debug")
	authorizer := &mockAuthorizer{}
	middleware := NewAuthMiddleware(authorizer, logger)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler := middleware.WithAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(UserIDKey)
		if userID == nil {
			t.Error("Expected user ID in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	cookie := rr.Header().Get("Set-Cookie")
	if cookie == "" {
		t.Error("Expected cookie to be set")
	}
}

func TestWithAuth_ValidHeader(t *testing.T) {
	logger, _ := logger.NewLogger("debug")
	authorizer := &mockAuthorizer{
		validTokens: map[string]string{"valid_token": "user123"},
	}
	middleware := NewAuthMiddleware(authorizer, logger)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "valid_token")
	rr := httptest.NewRecorder()

	var capturedUserID string
	handler := middleware.WithAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = r.Context().Value(UserIDKey).(string)
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if capturedUserID != "user123" {
		t.Errorf("Expected user123, got %s", capturedUserID)
	}
}

func TestWithAuth_ValidCookie(t *testing.T) {
	logger, _ := logger.NewLogger("debug")
	authorizer := &mockAuthorizer{
		validTokens: map[string]string{"cookie_token": "user456"},
	}
	middleware := NewAuthMiddleware(authorizer, logger)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: "cookie_token"})
	rr := httptest.NewRecorder()

	var capturedUserID string
	handler := middleware.WithAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = r.Context().Value(UserIDKey).(string)
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if capturedUserID != "user456" {
		t.Errorf("Expected user456, got %s", capturedUserID)
	}
}

func TestWithAuth_InvalidHeader(t *testing.T) {
	logger, _ := logger.NewLogger("debug")
	authorizer := &mockAuthorizer{validateError: true}
	middleware := NewAuthMiddleware(authorizer, logger)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "invalid_token")
	rr := httptest.NewRecorder()

	handler := middleware.WithAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}
}

func TestWithAuth_TokenCreationError(t *testing.T) {
	logger, _ := logger.NewLogger("debug")
	authorizer := &mockAuthorizer{createFail: true}
	middleware := NewAuthMiddleware(authorizer, logger)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler := middleware.WithAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rr.Code)
	}
}

func TestUpdateCookie(t *testing.T) {
	rr := httptest.NewRecorder()
	updateCookie(rr, "test_token")

	cookieHeader := rr.Header().Get("Set-Cookie")
	if cookieHeader == "" {
		t.Error("Expected cookie to be set")
	}
}

func TestGenerateNewUserID(t *testing.T) {
	userID, err := generateNewUserID()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if userID == "" {
		t.Error("Expected non-empty user ID")
	}
}
