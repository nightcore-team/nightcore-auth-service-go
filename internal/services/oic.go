package services

import (
	"nightcore-team/nightcore-auth-service-go/config"
	"nightcore-team/nightcore-auth-service-go/internal/domain/entity"
)

type TokenService interface {
	CreateAccessToken(userID int64) (string, error)
	CreateRefreshToken() string
}

type SessionRepository interface {
	Create(ttl int64, ipAddress, refreshToken string, userID int64) *entity.Session
	Delete(refreshToken string, userID int64) int64
	DeleteAll(userID int64)
	Get(refreshToken string) *entity.Session
	GetAll(userID int64) []string
}

type OauthProvider interface {
	ExchangeCode(code string) *entity.DiscordTokenData
	GetUserInfo(accessToken string) *entity.DiscordUserData
}

type OICService struct {
	sessionRepo   SessionRepository
	oauthProvider OauthProvider
	tokenService  TokenService
}

func NewOICService(sessionRepo SessionRepository, oauthProvider OauthProvider, tokenService TokenService) *OICService {
	return &OICService{
		sessionRepo:   sessionRepo,
		oauthProvider: oauthProvider,
		tokenService:  tokenService,
	}
}

func (s *OICService) Login(code, ipAddress string) *entity.TokenExchangeResult {
	exchangeResult := s.oauthProvider.ExchangeCode(code)
	userInfo := s.oauthProvider.GetUserInfo(exchangeResult.Access_token)

	refreshToken := s.tokenService.CreateRefreshToken()

	userSessions := s.sessionRepo.GetAll(userInfo.ID)
	if len(userSessions) >= 2 {
		s.sessionRepo.DeleteAll(userInfo.ID)
	}

	s.sessionRepo.Create(config.JWT.RefreshTokenDaysTTL, ipAddress, refreshToken, userInfo.ID)

	accessToken, err := s.tokenService.CreateAccessToken(userInfo.ID)
	if err != nil {
		panic("")
	}

	return &entity.TokenExchangeResult{
		AccessToken: accessToken,
		RefreshToken: refreshToken,
	}
}

func (s *OICService) Refresh(refreshToken, ipAddress string) *entity.TokenExchangeResult {
	session := s.sessionRepo.Get(refreshToken)

	if session == nil {
		panic("")
	}

	if session.IpAddress!= ipAddress {
		panic("")
	}

	keysCount := s.sessionRepo.Delete(refreshToken, session.UserID)

	if keysCount != 1 {
		panic("")
	}

	refreshToken = s.tokenService.CreateRefreshToken()

	session = s.sessionRepo.Create(config.JWT.RefreshTokenDaysTTL, ipAddress, refreshToken, session.UserID)

	accessToken, err := s.tokenService.CreateAccessToken(session.UserID)
	if err != nil {
		panic("")
	}

	return &entity.TokenExchangeResult{
		AccessToken: accessToken,
		RefreshToken: refreshToken,
	}
}

func (s *OICService) Logout(refreshToken string) {
	session := s.sessionRepo.Get(refreshToken)
	if session == nil {
		return
	}

	s.sessionRepo.Delete(refreshToken, session.UserID)
}