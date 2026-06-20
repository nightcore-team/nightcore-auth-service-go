package http_handlers

import (
	"context"
	"net/http"
	"net/url"
	"nightcore-team/nightcore-auth-service-go/config"
	"nightcore-team/nightcore-auth-service-go/internal/domain"
	"nightcore-team/nightcore-auth-service-go/internal/domain/entity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OICService interface {
	Login(ctx context.Context, code, ipAddress string) (*entity.TokenExchangeResult, *domain.AppError)
	Refresh(ctx context.Context, refreshToken, ipAddress string) (*entity.TokenExchangeResult, *domain.AppError)
	Logout(ctx context.Context, refreshToken string) *domain.AppError
}

type AuthHandler struct {
	oicService OICService
}

func NewAuthHandler(oicService OICService) *AuthHandler {
	return &AuthHandler{
		oicService: oicService,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	state := uuid.New().String()
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("oauth_state", state, 300, "", "", true, true)

	params := url.Values{}
	params.Add("client_id", config.OAuth.DISCORD_AUTH_CLIENT_ID)
	params.Add("redirect_uri", config.OAuth.DISCORD_AUTH_REDIRECT_URI)
	params.Add("response_type", "code")
	params.Add("scope", "identify")
	params.Add("state", state)

	location := "https://discord.com/oauth2/authorize?" + params.Encode()

	c.Redirect(http.StatusTemporaryRedirect, location)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	ctx := c.Request.Context()
	refresh_token := c.GetString("refresh_token")

	appErr := h.oicService.Logout(ctx, refresh_token)
	if appErr != nil {
		c.JSON(appErr.Status, gin.H{"error": appErr.Message})
		return
	}

	c.SetCookie("refresh_token", "", -1, "", "", true, true)
	c.Status(http.StatusOK)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	ctx := c.Request.Context()
	refresh_token := c.GetString("refresh_token")
	ipAddress := c.GetString("ipAddress")

	tokenResponse, appErr := h.oicService.Refresh(ctx, refresh_token, ipAddress)
	if appErr != nil {
		c.JSON(appErr.Status, gin.H{"error": appErr.Message})
		return
	}

	maxAge := int(config.JWT.RefreshTokenDaysTTL) * 24 * 60 * 60
	c.SetCookie("refresh_token", tokenResponse.RefreshToken, maxAge, "", "", true, true)

	c.JSON(http.StatusOK, gin.H{"access_token": tokenResponse.AccessToken})
}

func (h *AuthHandler) DiscordCallback(c *gin.Context) {
	ctx := c.Request.Context()
	ipAddress := c.GetString("ipAddress")

	discordError := c.Query("error")
	if discordError != "" {
		c.Redirect(http.StatusTemporaryRedirect, config.API.DASHBOARD_FRONTEND_URI+"?error="+discordError)
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code not found"})
		return
	}

	state := c.Query("state")
	cookieState, err := c.Cookie("oauth_state")
	if err != nil || state == "" || state != cookieState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid oauth state"})
		return
	}
	c.SetCookie("oauth_state", "", -1, "", "", true, true)

	tokenResponse, appErr := h.oicService.Login(ctx, code, ipAddress)
	if appErr != nil {
		c.JSON(appErr.Status, gin.H{"error": appErr.Message})
		return
	}

	maxAge := int(config.JWT.RefreshTokenDaysTTL) * 24 * 60 * 60
	c.SetCookie("refresh_token", tokenResponse.RefreshToken, maxAge, "", "", true, true)

	c.Redirect(http.StatusFound, config.API.DASHBOARD_FRONTEND_URI)
}