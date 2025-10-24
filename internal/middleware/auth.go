// Package middleware provides HTTP middleware components for the URL shortening service.
package middleware

import (
	"context"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/service"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
)

// userIDKey is the context key type for storing user ID in request context.
type userIDKey string

// Auth constants for middleware.
const (
	// CookieName is the name of the authentication cookie.
	CookieName = "user_token"

	// UserIDKey is the context key for storing and retrieving user ID.
	UserIDKey userIDKey = "userID"
)

// AuthMiddleware provides JWT-based authentication for HTTP requests.
// It handles both Authorization header and cookie-based authentication.
type AuthMiddleware struct {
	authorizer service.Authorizer
	logger     *logger.Logger
}

// NewAuthMiddleware creates a new AuthMiddleware instance.
//
// Parameters:
//   - authorizer: JWT authorizer service for token validation and creation
//   - logger: logger instance for authentication events
//
// Returns:
//   - *AuthMiddleware: initialized authentication middleware
func NewAuthMiddleware(authorizer service.Authorizer, logger *logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authorizer: authorizer,
		logger:     logger,
	}
}

// WithAuth wraps an HTTP handler with JWT authentication.
// Supports both Authorization header and cookie-based authentication.
// Creates new user ID and token for unauthenticated requests.
//
// Parameters:
//   - h: HTTP handler to wrap
//
// Returns:
//   - http.Handler: wrapped handler that requires authentication
func (m *AuthMiddleware) WithAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string
		var err error

		// Check Authorization header first
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			userID, err = m.authorizer.ValidateToken(authHeader)
			if err != nil {
				http.Error(w, "Invalid auth header", http.StatusUnauthorized)
				return
			}
		}

		// If no header, check cookie
		if userID == "" {
			cookie, cookieErr := r.Cookie(CookieName)
			if cookieErr == nil {
				userID, cookieErr = m.authorizer.ValidateToken(cookie.Value)
				if cookieErr != nil {
					http.Error(w, "Invalid cookie", http.StatusUnauthorized)
					return
				}
			}
		}

		// If still no user ID, create new user
		if userID == "" {
			userID, err = generateNewUserID()
			if err != nil {
				m.logger.Error("Failed to generate new user ID", zap.Error(err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			m.logger.Infoln("New user created", zap.String("userID", userID))
		}

		// Create new token and update cookie
		newToken, err := m.authorizer.CreateToken(userID)
		if err != nil {
			m.logger.Error("Failed to create user token", zap.Error(err), zap.String("userID", userID))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		updateCookie(w, newToken)

		// Set Authorization header and user ID in context
		w.Header().Set("Authorization", newToken)
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// updateCookie sets or updates the authentication cookie with the new token.
//
// Parameters:
//   - w: HTTP response writer
//   - newToken: JWT token to set in the cookie
func updateCookie(w http.ResponseWriter, newToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    newToken,
		Path:     "/",
		MaxAge:   3600 * 24 * 30, // 30 days
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// generateNewUserID creates a new unique user identifier using UUID.
//
// Returns:
//   - string: generated user ID
//   - error: error if UUID generation fails
func generateNewUserID() (string, error) {
	userID, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return userID.String(), nil
}
