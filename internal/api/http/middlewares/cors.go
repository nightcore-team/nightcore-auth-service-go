package http_middlewares

import (
	"net/http"
	"nightcore-team/nightcore-auth-service-go/config"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow requests from any origin.
		c.Writer.Header().Set("Access-Control-Allow-Origin", config.API.DASHBOARD_FRONTEND_URI)

		// Allowed HTTP methods.
		c.Writer.Header().
			Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")

		// Allowed headers.
		c.Writer.Header().
			Set("Access-Control-Allow-Headers", "*")

		// Allow credentials (cookies, auth headers, etc.).
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight OPTIONS request.
		if c.Request.Method == "OPTIONS" {
			// Cache preflight response for 24 hours.
			c.Writer.Header().Set("Access-Control-Max-Age", "86400")

			// Return 204 No Content for preflight.
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// Pass control to the next middleware/handler.
		c.Next()
	}
}