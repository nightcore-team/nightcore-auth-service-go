package http_middlewares

import (
	"github.com/gin-gonic/gin"
)

func IpAddress(c *gin.Context) {
	ipAddress := c.GetHeader("CF-Connecting-IP")

	c.Keys["ipAddress"] = ipAddress
	c.Next()
}