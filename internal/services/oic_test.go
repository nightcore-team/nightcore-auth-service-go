package services

import (
	"context"
	"testing"
	"time"

	"nightcore-team/nightcore-auth-service-go/config"
	"nightcore-team/nightcore-auth-service-go/internal/domain"
	"nightcore-team/nightcore-auth-service-go/internal/domain/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockSessionRepository struct{ mock.Mock }

func (m *MockSessionRepository) Create(ctx context.Context, ttl time.Duration, ipAddress, refreshToken string, userID int64) (*entity.Session, *domain.AppError) {
	args := m.Called(ctx, ttl, ipAddress, refreshToken, userID)
	return sessionPtr(args, 0), appErr(args, 1)
}
func (m *MockSessionRepository) Delete(ctx context.Context, refreshToken string, userID int64) (int64, *domain.AppError) {
	args := m.Called(ctx, refreshToken, userID)
	return args.Get(0).(int64), appErr(args, 1)
}
func (m *MockSessionRepository) Get(ctx context.Context, refreshToken string) (*entity.Session, *domain.AppError) {
	args := m.Called(ctx, refreshToken)
	return sessionPtr(args, 0), appErr(args, 1)
}
func (m *MockSessionRepository) GetDel(ctx context.Context, refreshToken string) (*entity.Session, *domain.AppError) {
	args := m.Called(ctx, refreshToken)
	return sessionPtr(args, 0), appErr(args, 1)
}

type MockOauthProvider struct{ mock.Mock }

func (m *MockOauthProvider) ExchangeCode(ctx context.Context, code string) (*entity.DiscordTokenData, *domain.AppError) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, appErr(args, 1)
	}
	return args.Get(0).(*entity.DiscordTokenData), appErr(args, 1)
}
func (m *MockOauthProvider) GetUserInfo(ctx context.Context, accessToken string) (*entity.DiscordUserData, *domain.AppError) {
	args := m.Called(ctx, accessToken)
	if args.Get(0) == nil {
		return nil, appErr(args, 1)
	}
	return args.Get(0).(*entity.DiscordUserData), appErr(args, 1)
}

type MockTokenService struct{ mock.Mock }

func (m *MockTokenService) CreateAccessToken(userID int64) (string, *domain.AppError) {
	args := m.Called(userID)
	return args.String(0), appErr(args, 1)
}
func (m *MockTokenService) CreateRefreshToken() string {
	args := m.Called()
	return args.String(0)
}

func appErr(args mock.Arguments, index int) *domain.AppError {
	if args.Get(index) == nil {
		return nil
	}
	return args.Get(index).(*domain.AppError)
}
func sessionPtr(args mock.Arguments, index int) *entity.Session {
	if args.Get(index) == nil {
		return nil
	}
	return args.Get(index).(*entity.Session)
}

func setupOIC() (*OICService, *MockSessionRepository, *MockOauthProvider, *MockTokenService) {
	config.JWT.RefreshTokenDaysTTL = 7

	repo := new(MockSessionRepository)
	oauth := new(MockOauthProvider)
	token := new(MockTokenService)

	return NewOICService(repo, oauth, token), repo, oauth, token
}

func TestOICService_Login(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, repo, oauth, token := setupOIC()
		ctx := context.Background()

		oauth.On("ExchangeCode", ctx, "discord-code").Return(&entity.DiscordTokenData{Access_token: "discord-access"}, nil)
		oauth.On("GetUserInfo", ctx, "discord-access").Return(&entity.DiscordUserData{ID: 123}, nil)
		token.On("CreateRefreshToken").Return("new-refresh-token")

		expectedTTL := time.Hour * 24 * time.Duration(config.JWT.RefreshTokenDaysTTL)
		repo.On("Create", ctx, expectedTTL, "127.0.0.1", "new-refresh-token", int64(123)).Return(&entity.Session{UserID: 123}, nil)
		token.On("CreateAccessToken", int64(123)).Return("new-access-token", nil)

		result, err := svc.Login(ctx, "discord-code", "127.0.0.1")

		require.Nil(t, err)
		assert.Equal(t, "new-access-token", result.AccessToken)
		assert.Equal(t, "new-refresh-token", result.RefreshToken)

		oauth.AssertExpectations(t)
		repo.AssertExpectations(t)
		token.AssertExpectations(t)
	})

	t.Run("error on ExchangeCode", func(t *testing.T) {
		svc, _, oauth, _ := setupOIC()
		ctx := context.Background()

		oauth.On("ExchangeCode", ctx, "bad-code").Return(nil, domain.ErrOauthExchangeFailed)

		result, err := svc.Login(ctx, "bad-code", "127.0.0.1")

		assert.Nil(t, result)
		assert.Equal(t, domain.ErrOauthExchangeFailed, err)
	})

	t.Run("error on GetUserInfo", func(t *testing.T) {
		svc, _, oauth, _ := setupOIC()
		ctx := context.Background()

		oauth.On("ExchangeCode", ctx, "code").Return(&entity.DiscordTokenData{Access_token: "acc"}, nil)
		oauth.On("GetUserInfo", ctx, "acc").Return(nil, domain.ErrOauthExchangeFailed)

		result, err := svc.Login(ctx, "code", "127.0.0.1")

		assert.Nil(t, result)
		assert.Equal(t, domain.ErrOauthExchangeFailed, err)
	})

	t.Run("error on Create session", func(t *testing.T) {
		svc, repo, oauth, token := setupOIC()
		ctx := context.Background()

		oauth.On("ExchangeCode", ctx, "code").Return(&entity.DiscordTokenData{Access_token: "acc"}, nil)
		oauth.On("GetUserInfo", ctx, "acc").Return(&entity.DiscordUserData{ID: 123}, nil)
		token.On("CreateRefreshToken").Return("new-refresh")

		expectedTTL := time.Hour * 24 * time.Duration(config.JWT.RefreshTokenDaysTTL)
		repo.On("Create", ctx, expectedTTL, "127.0.0.1", "new-refresh", int64(123)).Return(nil, domain.ErrUnknownRedis)

		result, err := svc.Login(ctx, "code", "127.0.0.1")

		assert.Nil(t, result)
		assert.Equal(t, domain.ErrUnknownRedis, err)
	})

	t.Run("error on CreateAccessToken", func(t *testing.T) {
		svc, repo, oauth, token := setupOIC()
		ctx := context.Background()

		oauth.On("ExchangeCode", ctx, "code").Return(&entity.DiscordTokenData{Access_token: "acc"}, nil)
		oauth.On("GetUserInfo", ctx, "acc").Return(&entity.DiscordUserData{ID: 123}, nil)
		token.On("CreateRefreshToken").Return("new-refresh")

		expectedTTL := time.Hour * 24 * time.Duration(config.JWT.RefreshTokenDaysTTL)
		repo.On("Create", ctx, expectedTTL, "127.0.0.1", "new-refresh", int64(123)).Return(&entity.Session{UserID: 123}, nil)
		token.On("CreateAccessToken", int64(123)).Return("", domain.ErrTokenSigningFailed)

		result, err := svc.Login(ctx, "code", "127.0.0.1")

		assert.Nil(t, result)
		assert.Equal(t, domain.ErrTokenSigningFailed, err)
	})
}

func TestOICService_Refresh(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, repo, _, token := setupOIC()
		ctx := context.Background()

		session := &entity.Session{UserID: 42, IpAddress: "127.0.0.1"}
		repo.On("GetDel", ctx, "old-refresh").Return(session, nil)
		token.On("CreateRefreshToken").Return("new-refresh")

		expectedTTL := time.Hour * 24 * time.Duration(config.JWT.RefreshTokenDaysTTL)
		repo.On("Create", ctx, expectedTTL, "127.0.0.1", "new-refresh", int64(42)).Return(&entity.Session{UserID: 42}, nil)
		token.On("CreateAccessToken", int64(42)).Return("new-access", nil)

		result, err := svc.Refresh(ctx, "old-refresh", "127.0.0.1")

		require.Nil(t, err)
		assert.Equal(t, "new-access", result.AccessToken)
		assert.Equal(t, "new-refresh", result.RefreshToken)
	})

	t.Run("session not found", func(t *testing.T) {
		svc, repo, _, _ := setupOIC()
		ctx := context.Background()

		repo.On("GetDel", ctx, "missing-token").Return(nil, nil)

		result, err := svc.Refresh(ctx, "missing-token", "127.0.0.1")

		assert.Nil(t, result)
		assert.Equal(t, domain.ErrSessionNotFound, err)
	})

	t.Run("ip mismatch", func(t *testing.T) {
		svc, repo, _, _ := setupOIC()
		ctx := context.Background()

		session := &entity.Session{UserID: 42, IpAddress: "1.1.1.1"}
		repo.On("GetDel", ctx, "old-refresh").Return(session, nil)

		result, err := svc.Refresh(ctx, "old-refresh", "2.2.2.2")

		assert.Nil(t, result)
		assert.Equal(t, domain.ErrSessionIPMismatch, err)
		repo.AssertNotCalled(t, "Create")
	})

	t.Run("error on GetDel", func(t *testing.T) {
		svc, repo, _, _ := setupOIC()
		ctx := context.Background()

		repo.On("GetDel", ctx, "bad-token").Return(nil, domain.ErrUnknownRedis)

		result, err := svc.Refresh(ctx, "bad-token", "127.0.0.1")

		assert.Nil(t, result)
		assert.Equal(t, domain.ErrUnknownRedis, err)
	})

	t.Run("error on Create session", func(t *testing.T) {
		svc, repo, _, token := setupOIC()
		ctx := context.Background()

		session := &entity.Session{UserID: 42, IpAddress: "127.0.0.1"}
		repo.On("GetDel", ctx, "old-refresh").Return(session, nil)
		token.On("CreateRefreshToken").Return("new-refresh")

		expectedTTL := time.Hour * 24 * time.Duration(config.JWT.RefreshTokenDaysTTL)
		repo.On("Create", ctx, expectedTTL, "127.0.0.1", "new-refresh", int64(42)).Return(nil, domain.ErrUnknownRedis)

		result, err := svc.Refresh(ctx, "old-refresh", "127.0.0.1")

		assert.Nil(t, result)
		assert.Equal(t, domain.ErrUnknownRedis, err)
	})

	t.Run("error on CreateAccessToken", func(t *testing.T) {
		svc, repo, _, token := setupOIC()
		ctx := context.Background()

		session := &entity.Session{UserID: 42, IpAddress: "127.0.0.1"}
		repo.On("GetDel", ctx, "old-refresh").Return(session, nil)
		token.On("CreateRefreshToken").Return("new-refresh")

		expectedTTL := time.Hour * 24 * time.Duration(config.JWT.RefreshTokenDaysTTL)
		repo.On("Create", ctx, expectedTTL, "127.0.0.1", "new-refresh", int64(42)).Return(&entity.Session{UserID: 42}, nil)
		token.On("CreateAccessToken", int64(42)).Return("", domain.ErrTokenSigningFailed)

		result, err := svc.Refresh(ctx, "old-refresh", "127.0.0.1")

		assert.Nil(t, result)
		assert.Equal(t, domain.ErrTokenSigningFailed, err)
	})
}

func TestOICService_Logout(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, repo, _, _ := setupOIC()
		ctx := context.Background()

		session := &entity.Session{UserID: 99}
		repo.On("Get", ctx, "valid-refresh").Return(session, nil)
		repo.On("Delete", ctx, "valid-refresh", int64(99)).Return(int64(1), nil)

		err := svc.Logout(ctx, "valid-refresh")

		assert.Nil(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("session not found", func(t *testing.T) {
		svc, repo, _, _ := setupOIC()
		ctx := context.Background()

		repo.On("Get", ctx, "missing-token").Return(nil, nil)

		err := svc.Logout(ctx, "missing-token")

		assert.Equal(t, domain.ErrSessionNotFound, err)
	})

	t.Run("error on Get", func(t *testing.T) {
		svc, repo, _, _ := setupOIC()
		ctx := context.Background()

		repo.On("Get", ctx, "token").Return(nil, domain.ErrUnknownRedis)

		err := svc.Logout(ctx, "token")

		assert.Equal(t, domain.ErrUnknownRedis, err)
	})

	t.Run("error on Delete", func(t *testing.T) {
		svc, repo, _, _ := setupOIC()
		ctx := context.Background()

		session := &entity.Session{UserID: 99}
		repo.On("Get", ctx, "valid-refresh").Return(session, nil)
		repo.On("Delete", ctx, "valid-refresh", int64(99)).Return(int64(0), domain.ErrUnknownRedis)

		err := svc.Logout(ctx, "valid-refresh")

		assert.Equal(t, domain.ErrUnknownRedis, err)
	})
}