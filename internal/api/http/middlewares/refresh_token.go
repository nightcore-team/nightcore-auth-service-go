package http_middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func EnsureRefreshTokenExists(c *gin.Context) {
	refresh_token, err := c.Cookie("refresh_token")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "Refresh token not provided in request")
		return
	}

	c.Keys["refresh_token"] = refresh_token
	c.Next()
}