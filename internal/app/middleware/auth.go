package middleware

import (
	"net/http"
	"rk-api/internal/app/constant"

	"github.com/gin-gonic/gin"
)

func R88AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求中获取 Authorization 头部
		authHeader := c.GetHeader("Authorization")
		// logger.Error("-----------------------------", authHeader)
		if authHeader == "" {
			// 如果 Authorization 头部不存在，返回错误
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is missing",
			})
			return
		}
		// 例如, 与预设的session id做对比或是调用验证服务
		if authHeader != constant.AuthorizationFixKey {
			// 如果认证失败，返回错误
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid Authorization token",
			})
			return
		}
		// 认证通过，请求会继续执行
		c.Next()
	}
}
