package http_handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nightcore-team/nightcore-auth-service-go/config"
	"nightcore-team/nightcore-auth-service-go/internal/domain"
	"nightcore-team/nightcore-auth-service-go/internal/domain/entity"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockOICService struct {
	mock.Mock
}

func (m *MockOICService) Login(ctx context.Context, code, ipAddress string) (*entity.TokenExchangeResult, *domain.AppError) {
	args := m.Called(ctx, code, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*domain.AppError)
	}
	return args.Get(0).(*entity.TokenExchangeResult), nil
}

func (m *MockOICService) Refresh(ctx context.Context, refreshToken, ipAddress string) (*entity.TokenExchangeResult, *domain.AppError) {
	args := m.Called(ctx, refreshToken, ipAddress)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*domain.AppError)
	}
	return args.Get(0).(*entity.TokenExchangeResult), nil
}

func (m *MockOICService) Logout(ctx context.Context, refreshToken string) *domain.AppError {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.AppError)
}

func setupTest() (*AuthHandler, *MockOICService) {
	gin.SetMode(gin.TestMode)

	config.OAuth.DISCORD_AUTH_CLIENT_ID = "test-client-id"
	config.OAuth.DISCORD_AUTH_REDIRECT_URI = "http://localhost/callback"
	config.API.DASHBOARD_FRONTEND_URI = "http://localhost:3000"
	config.JWT.RefreshTokenDaysTTL = 7

	mockService := new(MockOICService)
	handler := NewAuthHandler(mockService)

	return handler, mockService
}

func getCookie(w *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range w.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func TestAuthHandler_Login(t *testing.T) {
	handler, _ := setupTest()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/login", nil)

	handler.Login(c)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	location := w.Header().Get("Location")
	assert.True(t, strings.HasPrefix(location, "https://discord.com/oauth2/authorize?"))
	assert.Contains(t, location, "client_id=test-client-id")
	assert.Contains(t, location, "redirect_uri=http%3A%2F%2Flocalhost%2Fcallback")
	assert.Contains(t, location, "state=")

	cookie := getCookie(w, "oauth_state")
	require.NotNil(t, cookie)
	assert.NotEmpty(t, cookie.Value)
	assert.Equal(t, 300, cookie.MaxAge)
	assert.True(t, cookie.HttpOnly)
	assert.True(t, cookie.Secure)
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, mockService := setupTest()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/logout", nil)
		c.Set("refresh_token", "valid-token")

		mockService.On("Logout", mock.Anything, "valid-token").Return(nil)

		handler.Logout(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		cookie := getCookie(w, "refresh_token")
		require.NotNil(t, cookie)
		assert.Equal(t, "", cookie.Value)
		assert.Equal(t, -1, cookie.MaxAge)
	})

	t.Run("service error", func(t *testing.T) {
		handler, mockService := setupTest()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/logout", nil)
		c.Set("refresh_token", "bad-token")

		appErr := &domain.AppError{Status: http.StatusUnauthorized, Message: "unauthorized"}
		mockService.On("Logout", mock.Anything, "bad-token").Return(appErr)

		handler.Logout(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "unauthorized")
		mockService.AssertExpectations(t)
	})
}

func TestAuthHandler_Refresh(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, mockService := setupTest()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/refresh", nil)
		c.Set("refresh_token", "old-token")
		c.Set("ipAddress", "127.0.0.1")

		tokenResp := &entity.TokenExchangeResult{AccessToken: "new-access", RefreshToken: "new-refresh"}
		mockService.On("Refresh", mock.Anything, "old-token", "127.0.0.1").Return(tokenResp, nil)

		handler.Refresh(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "new-access")
		mockService.AssertExpectations(t)

		cookie := getCookie(w, "refresh_token")
		require.NotNil(t, cookie)
		assert.Equal(t, "new-refresh", cookie.Value)
		assert.Equal(t, 7*24*60*60, cookie.MaxAge)
	})

	t.Run("service error", func(t *testing.T) {
		handler, mockService := setupTest()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/refresh", nil)
		c.Set("refresh_token", "bad-token")
		c.Set("ipAddress", "127.0.0.1")

		appErr := &domain.AppError{Status: http.StatusForbidden, Message: "forbidden"}
		mockService.On("Refresh", mock.Anything, "bad-token", "127.0.0.1").Return(nil, appErr)

		handler.Refresh(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "forbidden")
		mockService.AssertExpectations(t)
	})
}

func TestAuthHandler_DiscordCallback(t *testing.T) {
	t.Run("discord error passed", func(t *testing.T) {
		handler, _ := setupTest()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/callback?error=access_denied", nil)

		handler.DiscordCallback(c)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Contains(t, w.Header().Get("Location"), "error=access_denied")
	})

	t.Run("missing code", func(t *testing.T) {
		handler, _ := setupTest()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/callback", nil)

		handler.DiscordCallback(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Code not found")
	})

	t.Run("invalid state mismatch", func(t *testing.T) {
		handler, _ := setupTest()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/callback?code=123&state=wrong", nil)
		req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "correct"})
		c.Request = req

		handler.DiscordCallback(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid oauth state")
	})

	t.Run("missing state cookie", func(t *testing.T) {
		handler, _ := setupTest()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/callback?code=123&state=123", nil)

		handler.DiscordCallback(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid oauth state")
	})

	t.Run("service error", func(t *testing.T) {
		handler, mockService := setupTest()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/callback?code=123&state=valid", nil)
		req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "valid"})
		c.Request = req
		c.Set("ipAddress", "127.0.0.1")

		appErr := &domain.AppError{Status: http.StatusInternalServerError, Message: "login failed"}
		mockService.On("Login", mock.Anything, "123", "127.0.0.1").Return(nil, appErr)

		handler.DiscordCallback(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "login failed")
		mockService.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		handler, mockService := setupTest()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/callback?code=123&state=valid", nil)
		req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "valid"})
		c.Request = req
		c.Set("ipAddress", "127.0.0.1")

		tokenResp := &entity.TokenExchangeResult{AccessToken: "acc", RefreshToken: "ref"}
		mockService.On("Login", mock.Anything, "123", "127.0.0.1").Return(tokenResp, nil)

		handler.DiscordCallback(c)

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Location"))
		mockService.AssertExpectations(t)

		refreshCookie := getCookie(w, "refresh_token")
		require.NotNil(t, refreshCookie)
		assert.Equal(t, "ref", refreshCookie.Value)
		assert.Equal(t, 7*24*60*60, refreshCookie.MaxAge)

		stateCookie := getCookie(w, "oauth_state")
		require.NotNil(t, stateCookie)
		assert.Equal(t, "", stateCookie.Value)
		assert.Equal(t, -1, stateCookie.MaxAge)
	})
}