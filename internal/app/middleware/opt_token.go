package middleware

import (
	"net/http"
	"rk-api/internal/app/config"

	"github.com/gin-gonic/gin"
)

// ValidateMD5Token 核对提供的MD5值是否和生成的相同  1julier@landing20232024-02-19
func ValidateOptToken(appKey string, token string) bool {

	return appKey == token
}

func OptMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if !ValidateOptToken(config.Get().ServiceSettings.JwtSignKey, token) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid  token"})
			c.Abort()
			return
		}
		// Extract userID from JWT claims
		c.Set("userID", 1)
		c.Next()
	}
}
