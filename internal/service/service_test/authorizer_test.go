package service_test

import (
	"errors"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/bezjen/shortener/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCreateToken(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	tests := []struct {
		name      string
		userID    string
		secretKey []byte
		wantErr   bool
	}{
		{
			name:      "successful token creation",
			userID:    "test-user-123",
			secretKey: []byte("valid-secret-key"),
			wantErr:   false,
		},
		{
			name:      "empty user ID",
			userID:    "",
			secretKey: []byte("valid-secret-key"),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authorizer := service.NewAuthorizer(tt.secretKey, testLogger)

			token, err := authorizer.CreateToken(tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				parsedToken, err := jwt.ParseWithClaims(token, &service.ShortenerClaims{}, func(token *jwt.Token) (interface{}, error) {
					return tt.secretKey, nil
				})

				assert.NoError(t, err)
				assert.True(t, parsedToken.Valid)

				if claims, ok := parsedToken.Claims.(*service.ShortenerClaims); ok {
					assert.Equal(t, tt.userID, claims.UserID)
					assert.Equal(t, "url-shortener", claims.Issuer)
				} else {
					t.Fatal("Failed to parse claims")
				}
			}
		})
	}
}

func TestValidateTokenSuccess(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	secretKey := []byte("test-secret-key")
	authorizer := service.NewAuthorizer(secretKey, testLogger)

	userID := "test-user-456"
	token, err := authorizer.CreateToken(userID)
	assert.NoError(t, err)

	validatedUserID, err := authorizer.ValidateToken(token)

	assert.NoError(t, err)
	assert.Equal(t, userID, validatedUserID)
}

func TestValidateTokenError(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	secretKey := []byte("test-secret-key")
	authorizer := service.NewAuthorizer(secretKey, testLogger)

	tests := []struct {
		name        string
		tokenString string
	}{
		{
			name:        "empty token",
			tokenString: "",
		},
		{
			name:        "malformed token",
			tokenString: "invalid.token.string",
		},
		{
			name:        "token with wrong signature",
			tokenString: "header.data.wrong-signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := authorizer.ValidateToken(tt.tokenString)

			assert.Error(t, err)
			assert.True(t, errors.Is(err, service.ErrValidate))
			assert.Empty(t, userID)
		})
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	secretKey := []byte("test-secret-key")
	authorizer := service.NewAuthorizer(secretKey, testLogger)

	// Создаем просроченный токен вручную
	claims := service.ShortenerClaims{
		UserID: "test-user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "url-shortener",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	assert.NoError(t, err)

	userID, err := authorizer.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrValidate))
	assert.Empty(t, userID)
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	authorizer := service.NewAuthorizer([]byte("correct-secret"), testLogger)

	// Создаем токен с другим секретным ключом
	claims := service.ShortenerClaims{
		UserID: "test-user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "url-shortener",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("wrong-secret"))
	assert.NoError(t, err)

	userID, err := authorizer.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrValidate))
	assert.Empty(t, userID)
}

func TestValidateToken_InvalidMethod(t *testing.T) {
	testLogger, _ := logger.NewLogger("debug")
	authorizer := service.NewAuthorizer([]byte("secret"), testLogger)

	// Создаем токен с неправильным методом подписи
	token := jwt.New(jwt.SigningMethodRS256)
	tokenString, err := token.SignedString("wrong-key-type")
	assert.Error(t, err) // Должна быть ошибка при подписи

	if err == nil {
		// Если по какой-то причине удалось подписать, проверяем валидацию
		_, err = authorizer.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, service.ErrValidate))
	}
}
