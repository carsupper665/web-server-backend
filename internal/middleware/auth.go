package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware is a simple placeholder for authentication.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		// TODO: validate token
		c.Next()
	}
}
