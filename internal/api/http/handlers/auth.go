package http_handlers

import (
	"fmt"
	"net/http"
	"nightcore-team/nightcore-auth-service-go/config"
	"nightcore-team/nightcore-auth-service-go/internal/domain/entity"

	"github.com/gin-gonic/gin"
)

type IOICService interface {
	Login(code, ipAddress string) *entity.TokenExchangeResult
	Refresh(refreshToken, ipAddress string) *entity.TokenExchangeResult
	Logout(refreshToken string)
}

type AuthHandler struct {
	oicService IOICService
}

func NewAuthHandler(oicService IOICService) *AuthHandler {
	return &AuthHandler{
		oicService: oicService,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	location := fmt.Sprintf("https://discord.com/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=identify", config.OAuth.DISCORD_AUTH_CLIENT_ID, config.OAuth.DISCORD_AUTH_REDIRECT_URI)

	c.Redirect(http.StatusTemporaryRedirect, location)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	refresh_token := c.GetString("refresh_token")

	h.oicService.Logout(refresh_token)

	c.SetCookie("refresh_token", "", -1, "", "", true, true)
	c.Status(http.StatusOK)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refresh_token := c.GetString("refresh_token")
	ipAddress := c.GetString("ipAddress")

	tokenResponse := h.oicService.Refresh(refresh_token, ipAddress)

	c.SetCookie("refresh_token", tokenResponse.RefreshToken, int(config.JWT.RefreshTokenDaysTTL), "", "", true, true)
}

func (h *AuthHandler) DiscordCallback(c *gin.Context) {
	ipAddress := c.GetString("ipAddress")

	discordError := c.Query("error")
	if discordError != "" {
	    c.JSON(http.StatusInternalServerError, discordError)
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, "Code not found")
		return
	}

	tokenResponse := h.oicService.Login(code, ipAddress)

	c.SetCookie("refresh_token", tokenResponse.AccessToken, int(config.JWT.RefreshTokenDaysTTL), "", "", true, true)
	c.Redirect(http.StatusOK, config.API.DASHBOARD_FRONTEND_URI)
}