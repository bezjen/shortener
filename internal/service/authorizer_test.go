package service

import (
	"errors"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"testing"
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
			authorizer := NewAuthorizer(tt.secretKey, testLogger)

			token, err := authorizer.CreateToken(tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				parsedToken, err := jwt.ParseWithClaims(token, &shortenerClaims{}, func(token *jwt.Token) (interface{}, error) {
					return tt.secretKey, nil
				})

				assert.NoError(t, err)
				assert.True(t, parsedToken.Valid)

				if claims, ok := parsedToken.Claims.(*shortenerClaims); ok {
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
	authorizer := NewAuthorizer(secretKey, testLogger)

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
	authorizer := NewAuthorizer(secretKey, testLogger)

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
			assert.True(t, errors.Is(err, ErrValidate))
			assert.Empty(t, userID)
		})
	}
}
