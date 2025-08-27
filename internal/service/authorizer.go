//go:generate mockery --name=Authorizer --output=../mocks --case=underscore
package service

import (
	"errors"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"time"
)

var ErrSign = errors.New("failed to generate token")
var ErrValidate = errors.New("failed to validate token")

type Authorizer interface {
	CreateToken(userID string) (string, error)
	ValidateToken(tokenString string) (string, error)
}

type JWTAuthorizer struct {
	secretKey []byte
	logger    *logger.Logger
}

func NewAuthorizer(secretKey []byte, logger *logger.Logger) *JWTAuthorizer {
	return &JWTAuthorizer{
		secretKey: secretKey,
		logger:    logger,
	}
}

func (a *JWTAuthorizer) CreateToken(userID string) (string, error) {
	claims := shortenerClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "url-shortener",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(a.secretKey)
	if err != nil {
		a.logger.Error("Failed to sign token", zap.Error(err))
		return "", ErrSign
	}

	return tokenString, nil
}

func (a *JWTAuthorizer) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &shortenerClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			a.logger.Error("Unexpected signing method")
			return nil, ErrValidate
		}
		return a.secretKey, nil
	})

	if err != nil {
		a.logger.Error("Failed to parse token", zap.Error(err), zap.String("token", tokenString))
		return "", ErrValidate
	}

	if claims, ok := token.Claims.(*shortenerClaims); ok && token.Valid {
		return claims.UserID, nil
	}

	a.logger.Error("Invalid token", zap.String("token", tokenString))
	return "", ErrValidate
}

type shortenerClaims struct {
	UserID string `json:"userID"`
	jwt.RegisteredClaims
}
