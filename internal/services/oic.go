package services

import (
	"context"
	"log/slog"
	"nightcore-team/nightcore-auth-service-go/config"
	"nightcore-team/nightcore-auth-service-go/internal/domain"
	"nightcore-team/nightcore-auth-service-go/internal/domain/entity"
	"time"
)

type TokenService interface {
	CreateAccessToken( userID int64) (string, *domain.AppError)
	CreateRefreshToken() string
}

type SessionRepository interface {
	Create(ctx context.Context, ttl time.Duration, ipAddress, refreshToken string, userID int64) (*entity.Session, *domain.AppError)
	Delete(ctx context.Context, refreshToken string, userID int64) (int64, *domain.AppError)
	DeleteAll(ctx context.Context, userID int64) *domain.AppError
	Get(ctx context.Context, refreshToken string) ( *entity.Session, *domain.AppError)
	GetDel(ctx context.Context, refreshToken string) ( *entity.Session, *domain.AppError)
	GetAll(ctx context.Context, userID int64) ([]string, *domain.AppError)
}

type OauthProvider interface {
	ExchangeCode(ctx context.Context, code string) (*entity.DiscordTokenData, *domain.AppError)
	GetUserInfo(ctx context.Context, accessToken string) (*entity.DiscordUserData, *domain.AppError)
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

func (s *OICService) Login(ctx context.Context, code, ipAddress string) (*entity.TokenExchangeResult, *domain.AppError) {
	exchangeResult, err := s.oauthProvider.ExchangeCode(ctx, code)
	if err != nil {
		slog.Error("failed to exchange discord code", "code", code, "error", err)
		return nil, err
	}
	userInfo, err := s.oauthProvider.GetUserInfo(ctx, exchangeResult.Access_token)
	if err != nil {
		slog.Error("failed to get discord user info", "error", err)
		return nil, err
	}

	refreshToken := s.tokenService.CreateRefreshToken()

	userSessions, err := s.sessionRepo.GetAll(ctx, userInfo.ID)
	if err != nil {
		slog.Error("failed to get user sessions", "user_id", userInfo.ID, "error", err)
		return nil, err
	}
	
	if len(userSessions) >= 2 {
		err := s.sessionRepo.DeleteAll(ctx, userInfo.ID)
		if err != nil {
			slog.Error("failed to delete all user sessions", "user_id", userInfo.ID, "error", err)
			return nil, err
		}
	}

	ttl := time.Hour * 24 * time.Duration(config.JWT.RefreshTokenDaysTTL)
	_, err = s.sessionRepo.Create(ctx, ttl, ipAddress, refreshToken, userInfo.ID)
	if err != nil {
		slog.Error("failed to create session", "user_id", userInfo.ID, "ip", ipAddress, "error", err)
		return nil, err
	}

	accessToken, err := s.tokenService.CreateAccessToken(userInfo.ID)
	if err != nil {
		slog.Error("failed to create access token", "user_id", userInfo.ID, "error", err)
		return nil, err
	}

	return &entity.TokenExchangeResult{
		AccessToken: accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *OICService) Refresh(ctx context.Context, refreshToken, ipAddress string) (*entity.TokenExchangeResult, *domain.AppError) {
	session, err := s.sessionRepo.GetDel(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	if session == nil {
		slog.Warn("session not found on refresh", "refresh_token", refreshToken)
		return nil, domain.ErrSessionNotFound
	}

	if session.IpAddress!= ipAddress {
		slog.Warn("ip mismatch on refresh", "session_ip", session.IpAddress, "request_ip", ipAddress, "user_id", session.UserID)
		return nil, domain.ErrSessionIPMismatch
	}

	refreshToken = s.tokenService.CreateRefreshToken()

	ttl := time.Hour * 24 * time.Duration(config.JWT.RefreshTokenDaysTTL)
	session, err = s.sessionRepo.Create(ctx, ttl, ipAddress, refreshToken, session.UserID)
	if err != nil {
		slog.Error("failed to create session on refresh", "user_id", session.UserID, "error", err)
		return nil, err
	}

	accessToken, err := s.tokenService.CreateAccessToken(session.UserID)
	if err != nil {
		slog.Error("failed to create access token on refresh", "user_id", session.UserID, "error", err)
		return nil, err
	}

	return &entity.TokenExchangeResult{
		AccessToken: accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *OICService) Logout(ctx context.Context, refreshToken string) *domain.AppError {
	session, err := s.sessionRepo.Get(ctx, refreshToken)
	if err != nil {
		slog.Error("failed to get session on logout", "refresh_token", refreshToken, "error", err)
		return err
	}

	if session == nil {
		slog.Warn("session not found on logout", "refresh_token", refreshToken)
		return domain.ErrSessionNotFound
	}

	_, err = s.sessionRepo.Delete(ctx, refreshToken, session.UserID)
	if err != nil {
		slog.Error("failed to delete session on logout", "user_id", session.UserID, "refresh_token", refreshToken, "error", err)
		return err
	}

	return nil
}