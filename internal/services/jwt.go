package services

import (
	"nightcore-team/nightcore-auth-service-go/config"
	"nightcore-team/nightcore-auth-service-go/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTTokenService struct{}

func NewTokenService() *JWTTokenService {
	return &JWTTokenService{}
}

func (s *JWTTokenService) CreateRefreshToken() string {
	return uuid.New().String()
}

func (s *JWTTokenService) CreateAccessToken(userID int64) (string, *domain.AppError) {
	return s.sign(userID)
}

func (s *JWTTokenService) sign(userID int64) (string, *domain.AppError) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(config.JWT.AccessTokenMinutesTTL))),
	}
 
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	signed, err := token.SignedString(config.JWT.JWT_PRIVATE_KEY)
	if err != nil {
		return "", domain.ErrTokenSigningFailed.WithCause(err)
	}

	return signed, nil
}