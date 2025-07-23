package middleware

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"rk-api/internal/app/config"
	"rk-api/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
)

// GenerateMD5Token 根据 userID, appKey 和当前日期生成MD5
func GenerateMD5Token(userID string, timezone string, appKey string) string {

	// 加载时区，例如东京时间
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return ""
	}
	// 使用加载的时区创建一个时间对象
	currentDate := time.Now().In(location).Format("2006-01-02")
	data := userID + appKey + currentDate
	logger.Info("----------------GenerateMD5Token-----------------" + data)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ValidateMD5Token 核对提供的MD5值是否和生成的相同  1julier@landing20232024-02-19
func ValidateMD5Token(userID string, timezone, appKey string, token string) bool {
	// 生成 token
	// logger.Info("--------ValidateMD5Token---------------", userID, appKey, token)
	generatedToken := GenerateMD5Token(userID, timezone, appKey)
	// 比较提供的 token 与生成的 token

	return generatedToken == token
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		timezone := c.Query("timezone")
		userID := c.Query("uid")
		token := c.Query("token")
		if !ValidateMD5Token(userID, timezone, config.Get().ServiceSettings.JwtSignKey, token) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid  token"})
			c.Abort()
			return
		}
		// Extract userID from JWT claims
		c.Set("userID", userID)
		c.Next()
	}
}
