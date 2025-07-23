package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/odeke-em/go-uuid"
	"go.uber.org/zap"
)

func TraceMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		// 从请求头中获取当前用户
		user := c.GetHeader("X-User")

		// 将当前用户添加到请求上下文中
		// 记录日志，包括当前用户

		reqID := c.Request.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New()
			c.Request.Header.Set("X-Request-ID", reqID)
		}
		logger.Info("request",
			zap.String("X-User", user),
			zap.String("request_id", reqID),
			zap.String("url", c.Request.URL.String()),
			zap.String("method", c.Request.Method),
			zap.String("source", c.Request.RemoteAddr),
		)
		c.Set("requestID", reqID)

		c.Next()
	}
}
