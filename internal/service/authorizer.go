// Package service provides business logic for URL shortening service.
//
//go:generate mockery --name=Authorizer --output=../mocks --case=underscore
package service

import (
	"errors"
	"github.com/bezjen/shortener/internal/logger"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"time"
)

// ErrSign is returned when JWT token creation fails.
var ErrSign = errors.New("failed to generate token")

// ErrValidate is returned when JWT token validation fails.
var ErrValidate = errors.New("failed to validate token")

// Authorizer defines the interface for JWT token creation and validation.
type Authorizer interface {
	// CreateToken generates a JWT token for the given user ID.
	//
	// Parameters:
	//   - userID: User identifier to include in the token
	//
	// Returns:
	//   - string: Signed JWT token
	//   - error: Error if token generation fails
	CreateToken(userID string) (string, error)

	// ValidateToken verifies a JWT token and extracts the user ID.
	//
	// Parameters:
	//   - tokenString: JWT token string to validate
	//
	// Returns:
	//   - string: User ID extracted from the token
	//   - error: Error if token validation fails
	ValidateToken(tokenString string) (string, error)
}

// JWTAuthorizer implements Authorizer using JWT tokens with HMAC signing.
type JWTAuthorizer struct {
	secretKey []byte
	logger    *logger.Logger
}

// ShortenerClaims defines the JWT claims structure for URL shortening service.
type ShortenerClaims struct {
	UserID string `json:"userID"`
	jwt.RegisteredClaims
}

// NewAuthorizer creates a new JWTAuthorizer instance with the given secret key.
//
// Parameters:
//   - secretKey: HMAC secret key for signing and verifying tokens
//   - logger: Logger instance for error logging
//
// Returns:
//   - *JWTAuthorizer: Initialized JWT authorizer
func NewAuthorizer(secretKey []byte, logger *logger.Logger) *JWTAuthorizer {
	return &JWTAuthorizer{
		secretKey: secretKey,
		logger:    logger,
	}
}

// CreateToken generates a JWT token for the given user ID with 30-day expiration.
// Uses HS256 signing method for token security.
//
// Parameters:
//   - userID: User identifier to include in the token
//
// Returns:
//   - string: Signed JWT token
//   - error: Error if token signing fails
func (a *JWTAuthorizer) CreateToken(userID string) (string, error) {
	claims := ShortenerClaims{
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

// ValidateToken verifies a JWT token's signature and extracts the user ID.
// Returns the user ID if the token is valid and properly signed.
//
// Parameters:
//   - tokenString: JWT token string to validate
//
// Returns:
//   - string: User ID extracted from the token
//   - error: Error if token is invalid, expired, or signature verification fails
func (a *JWTAuthorizer) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ShortenerClaims{}, func(token *jwt.Token) (interface{}, error) {
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

	if claims, ok := token.Claims.(*ShortenerClaims); ok && token.Valid {
		return claims.UserID, nil
	}

	a.logger.Error("Invalid token", zap.String("token", tokenString))
	return "", ErrValidate
}
