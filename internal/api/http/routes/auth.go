package http_routes

import (
	http_middlewares "nightcore-team/nightcore-auth-service-go/internal/api/http/middlewares"

	"github.com/gin-gonic/gin"
)

type AuthHandler interface {
	Login(*gin.Context)
	Logout(*gin.Context)
	DiscordCallback(*gin.Context)
	Refresh(*gin.Context)
}

func AddAuthRoutes(rg *gin.RouterGroup, handler AuthHandler) {
	rg.Group("/auth")

	rg.POST("/refresh", http_middlewares.IpAddress, http_middlewares.EnsureRefreshTokenExists,  handler.Refresh)
	rg.POST("/logout", http_middlewares.EnsureRefreshTokenExists, handler.Logout)
	rg.GET("/login", handler.Login)
	rg.GET("/discord/callback", handler.DiscordCallback)
}