package jwt

import (
	"nightcore-team/nightcore-auth-service-go/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenService struct{}

func NewTokenService() *TokenService {
	return &TokenService{}
}

func (s *TokenService) CreateRefreshToken() string {
	return uuid.New().String()
}

func (s *TokenService) CreateAccessToken(userID int64) (string, error) {
	return s.sign(userID)
}

func (s *TokenService) sign(userID int64) (string, error) {

	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Minute * time.Duration(config.JWT.AccessTokenMinutesTTL)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	
	return token.SignedString(config.JWT.JWT_PRIVATE_KEY)
}