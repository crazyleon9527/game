package middleware

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TraceIDMiddleware 注入唯一 TraceID 到每个请求
func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := uuid.New().String()
		c.Set("TraceID", traceID)
		c.Writer.Header().Set("X-Trace-ID", traceID)
		c.Next()
	}
}

// RecoverMiddleware 捕获所有 panic 并写入 TraceID 日志
func RecoverMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				traceID, _ := c.Get("TraceID")
				logger.Error("panic recovered",
					zap.Any("error", r),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
					zap.String("trace_id", traceID.(string)),
					zap.ByteString("stack", debug.Stack()),
					zap.Time("time", time.Now()),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":    "服务异常，请稍后重试",
					"message":  "panic recovered",
					"trace_id": traceID,
				})
			}
		}()

		c.Next()
	}
}
