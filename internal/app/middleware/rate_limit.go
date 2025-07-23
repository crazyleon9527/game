package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateData struct {
	Count     int
	Timestamp time.Time
}

var rateLimiter sync.Map

// RateLimitMiddleware 限制每个 IP 对每个路由的请求频率
// 每个 IP 每个接口 每分钟不得超过 limit 次
func RateLimitMiddleware(limit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		endpoint := c.FullPath()
		if endpoint == "" {
			// fallback路径兼容：如静态文件或未匹配路由
			endpoint = c.Request.URL.Path
		}

		key := clientIP + "|" + endpoint
		now := time.Now()

		v, _ := rateLimiter.LoadOrStore(key, &rateData{
			Count:     1,
			Timestamp: now,
		})
		data := v.(*rateData)

		// 重置时间窗
		if now.Sub(data.Timestamp) > time.Minute {
			data.Count = 1
			data.Timestamp = now
		} else {
			data.Count++
		}

		if data.Count > limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "请求过于频繁，请稍后再试",
				"message": "Rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}
