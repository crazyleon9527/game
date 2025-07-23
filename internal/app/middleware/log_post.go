package middleware

import (
	"fmt"
	"rk-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LogPostParamsMiddleware 创建一个中间件来打印POST请求中的JSON参数
func LogPostParamsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否是POST请求
		if c.Request.Method == "POST" {
			// 检查传入的内容类型是否为JSON
			contentType := c.GetHeader("Content-Type")
			if contentType == "application/json" {
				var jsonMap map[string]interface{}
				// 绑定JSON到map
				if err := c.BindJSON(&jsonMap); err == nil {
					// 遍历并打印JSON字段和值

					logger.ZInfo("post request", zap.Any("param", jsonMap))
					// for key, value := range jsonMap {
					// 	fmt.Printf("key: %s, value: %v\n", key, value)
					// }
				} else {
					// 如果无法绑定JSON，打印错误
					fmt.Printf("Error binding JSON: %s\n", err)
					logger.Error(fmt.Sprintf("Error binding JSON: %s", err))
				}
			}
		}

		c.Next() // 处理下一个中间件或路由处理器
	}
}
