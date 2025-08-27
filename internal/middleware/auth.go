package middleware

import (
	"context"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/service"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
)

const (
	CookieName = "user_token"
	UserIDKey  = "userID"
)

type AuthMiddleware struct {
	authorizer service.Authorizer
	logger     *logger.Logger
}

func NewAuthMiddleware(authorizer service.Authorizer, logger *logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authorizer: authorizer,
		logger:     logger,
	}
}

func (m *AuthMiddleware) WithAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string

		cookie, err := r.Cookie(CookieName)
		if err != nil {
			userID, err = generateNewUserID()
			if err != nil {
				m.logger.Error("Failed to generate new user ID", zap.Error(err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			m.logger.Infoln("New user created", zap.String("userID", userID))
		} else {
			userID, err = m.authorizer.ValidateToken(cookie.Value)
			if err != nil {
				http.Error(w, "Invalid cookie", http.StatusUnauthorized)
				return
			}
		}

		newToken, err := m.authorizer.CreateToken(userID)
		if err != nil {
			m.logger.Error("Failed to create user token", zap.Error(err), zap.String("userID", userID))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		updateCookie(w, newToken)

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func updateCookie(w http.ResponseWriter, newToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    newToken,
		Path:     "/",
		MaxAge:   3600 * 24 * 30,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func generateNewUserID() (string, error) {
	userID, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return userID.String(), nil
}
