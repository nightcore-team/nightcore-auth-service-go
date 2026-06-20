package services

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"testing"
	"time"

	"nightcore-team/nightcore-auth-service-go/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestRSAKey(t *testing.T) *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return privateKey
}

func TestJWTTokenService_CreateRefreshToken(t *testing.T) {
	service := NewTokenService()

	t.Run("valid uuid format", func(t *testing.T) {
		token := service.CreateRefreshToken()

		_, err := uuid.Parse(token)
		assert.NoError(t, err, "Refresh token should be a valid UUID")
	})

	t.Run("uniqueness", func(t *testing.T) {
		tokens := make(map[string]bool)
		for range 100 {
			token := service.CreateRefreshToken()
			assert.False(t, tokens[token], "Refresh token should be unique")
			tokens[token] = true
		}
	})
}

func TestJWTTokenService_CreateAccessToken(t *testing.T) {
	privateKey := generateTestRSAKey(t)
	
	config.JWT.JWT_PRIVATE_KEY = privateKey
	config.JWT.AccessTokenMinutesTTL = 15

	service := NewTokenService()
	userID := int64(42)

	t.Run("success and valid claims", func(t *testing.T) {
		tokenString, appErr := service.CreateAccessToken(userID)
		
		require.Nil(t, appErr)
		assert.NotEmpty(t, tokenString)

		parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return &privateKey.PublicKey, nil
		})

		require.NoError(t, err)
		require.True(t, parsedToken.Valid, "Token signature must be valid")

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		require.True(t, ok)

		sub, ok := claims["sub"].(float64)
		require.True(t, ok, "Claim 'sub' must be a number")
		assert.Equal(t, float64(userID), sub)

		exp, ok := claims["exp"].(float64)
		require.True(t, ok, "Claim 'exp' must be a number")
		
		expectedExp := float64(time.Now().Add(time.Minute * time.Duration(config.JWT.AccessTokenMinutesTTL)).Unix())
		
		assert.InDelta(t, expectedExp, exp, 2.0, "Expiration time should be within 2 seconds of expected")
	})
}